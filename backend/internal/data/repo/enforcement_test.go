package repo

// Enforcement suite for the ent privacy layer. The business-logic test suite
// runs under a privacy-bypassing system context (testCtx); these tests build
// real viewers and assert the ORM-level authorization behavior:
//
//   - reads are filtered to the viewer's tenant and permissions
//   - inaccessible single rows surface as NotFoundError (no existence leak)
//   - row-level grants open exactly the granted entity (and its children)
//   - mutations are pinned to writable rows and denied without permissions
//   - the last-admin invariant holds
//   - instance superusers bypass permission checks but stay tenant-scoped

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/privacy"
)

// authzTenant bundles a fresh tenant with its admin (full catalog) user.
type authzTenant struct {
	group Group
	admin UserOut
}

func newAuthzTenant(t *testing.T) authzTenant {
	t.Helper()

	g, err := tRepos.Groups.GroupCreate(testCtx(), "authz-"+fk.Str(8), uuid.Nil)
	require.NoError(t, err)

	admin := newAuthzMember(t, g.ID, nil, true)

	t.Cleanup(func() {
		_ = tRepos.Groups.GroupDelete(testCtx(), g.ID)
	})

	return authzTenant{group: g, admin: admin}
}

// newAuthzMember creates a member of gid with the given direct permissions.
// nil perms means the full catalog; pass an explicit empty slice via
// noPerms() for a member with no permissions at all.
func newAuthzMember(t *testing.T, gid uuid.UUID, perms []string, owner bool) UserOut {
	t.Helper()

	u, err := tRepos.Users.Create(testCtx(), UserCreate{
		Name:           fk.Str(10),
		Email:          fk.Email(),
		DefaultGroupID: gid,
		IsOwner:        owner,
		Permissions:    perms,
	})
	require.NoError(t, err)

	// UserCreate treats an empty permission list as "full catalog" for
	// back-compat, so explicitly restricted members are set after creation.
	if perms != nil {
		require.NoError(t, tRepos.Permissions.MemberPermissionsSet(testCtx(), gid, u.ID, perms))
	}

	t.Cleanup(func() {
		_ = tRepos.Users.Delete(testCtx(), u.ID)
	})
	return u
}

// viewerCtx resolves a real viewer for (uid, gid) and returns a context
// carrying it, exactly like the viewer middleware does.
func viewerCtx(t *testing.T, uid, gid uuid.UUID, superuser bool) context.Context {
	t.Helper()
	v, err := tRepos.Permissions.ResolveViewer(context.Background(), uid, gid, superuser)
	require.NoError(t, err)
	return authz.NewContext(context.Background(), v)
}

// newAuthzEntity creates an entity in gid (system context).
func newAuthzEntity(t *testing.T, gid uuid.UUID) EntityOut {
	t.Helper()
	et, err := tRepos.EntityTypes.GetDefault(testCtx(), gid, false)
	require.NoError(t, err)

	e, err := tRepos.Entities.Create(testCtx(), gid, EntityCreate{
		Name:         fk.Str(10),
		Description:  fk.Str(20),
		EntityTypeID: et.ID,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tRepos.Entities.Delete(testCtx(), e.ID)
	})
	return e
}

func grantEntity(t *testing.T, gid, entityID uuid.UUID, target AccessGrantCreate, actions authz.GrantActions) AccessGrantOut {
	t.Helper()
	g, err := tRepos.Permissions.GrantCreate(testCtx(), gid, entityID, target, actions)
	require.NoError(t, err)
	return g
}

func TestEnforcement_ReadFiltering(t *testing.T) {
	a := newAuthzTenant(t)
	b := newAuthzTenant(t)

	e1 := newAuthzEntity(t, a.group.ID)
	e2 := newAuthzEntity(t, a.group.ID)
	_ = newAuthzEntity(t, b.group.ID)

	reader := newAuthzMember(t, a.group.ID, []string{"entity:read"}, false)
	noEntity := newAuthzMember(t, a.group.ID, []string{"notifier:manage"}, false)

	// Reader sees exactly tenant A's entities.
	ids, err := tClient.Entity.Query().IDs(viewerCtx(t, reader.ID, a.group.ID, false))
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{e1.ID, e2.ID}, ids)

	// A member without entity:read sees nothing.
	ids, err = tClient.Entity.Query().IDs(viewerCtx(t, noEntity.ID, a.group.ID, false))
	require.NoError(t, err)
	require.Empty(t, ids)

	// Tenant B's admin sees none of A's entities, and a direct lookup of an
	// A entity is a NotFound — existence is not leaked.
	bCtx := viewerCtx(t, b.admin.ID, b.group.ID, false)
	_, err = tClient.Entity.Get(bCtx, e1.ID)
	require.True(t, ent.IsNotFound(err), "expected NotFound, got %v", err)
}

func TestEnforcement_GrantOnlyAccess(t *testing.T) {
	tn := newAuthzTenant(t)

	target := newAuthzEntity(t, tn.group.ID)
	other := newAuthzEntity(t, tn.group.ID)

	// Child rows on the granted entity.
	att, err := tClient.Attachment.Create().
		SetEntityID(target.ID).
		SetTitle("manual").
		Save(testCtx())
	require.NoError(t, err)
	maint, err := tClient.MaintenanceEntry.Create().
		SetEntityID(target.ID).
		SetName("oil change").
		SetDate(time.Now()).
		Save(testCtx())
	require.NoError(t, err)

	user := newAuthzMember(t, tn.group.ID, []string{"notifier:manage"}, false)

	grantEntity(t, tn.group.ID, target.ID, AccessGrantCreate{
		TargetType: AccessGrantTargetUser,
		TargetID:   user.ID,
		Actions:    []string{"read"},
	}, authz.GrantActions{Read: true})

	ctx := viewerCtx(t, user.ID, tn.group.ID, false)

	// Sees exactly the granted entity.
	ids, err := tClient.Entity.Query().IDs(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{target.ID}, ids)

	// And its child rows.
	gotAtt, err := tClient.Attachment.Query().IDs(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{att.ID}, gotAtt)

	gotMaint, err := tClient.MaintenanceEntry.Query().IDs(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{maint.ID}, gotMaint)

	// The other entity stays invisible.
	_, err = tClient.Entity.Get(ctx, other.ID)
	require.True(t, ent.IsNotFound(err))
}

func TestEnforcement_GrantActionsAreExact(t *testing.T) {
	tn := newAuthzTenant(t)
	target := newAuthzEntity(t, tn.group.ID)

	user := newAuthzMember(t, tn.group.ID, []string{"notifier:manage"}, false)

	grant := grantEntity(t, tn.group.ID, target.ID, AccessGrantCreate{
		TargetType: AccessGrantTargetUser,
		TargetID:   user.ID,
		Actions:    []string{"update"},
	}, authz.GrantActions{Read: true, Update: true})

	ctx := viewerCtx(t, user.ID, tn.group.ID, false)

	// can_update allows updating the granted row...
	err := tClient.Entity.UpdateOneID(target.ID).SetNotes("updated by grantee").Exec(ctx)
	require.NoError(t, err)

	// ...but not deleting it (no can_delete: the delete pin matches nothing).
	err = tClient.Entity.DeleteOneID(target.ID).Exec(ctx)
	require.Error(t, err)
	require.True(t, ent.IsNotFound(err), "expected NotFound, got %v", err)

	// And another, ungranted entity cannot be updated.
	otherEntity := newAuthzEntity(t, tn.group.ID)
	err = tClient.Entity.UpdateOneID(otherEntity.ID).SetNotes("nope").Exec(ctx)
	require.True(t, ent.IsNotFound(err), "expected NotFound, got %v", err)

	// Attachments: update grant covers attachment writes on the entity.
	_, err = tClient.Attachment.Create().SetEntityID(target.ID).SetTitle("receipt").Save(ctx)
	require.NoError(t, err)

	// Revoke the grant: access disappears.
	require.NoError(t, tRepos.Permissions.GrantDelete(testCtx(), tn.group.ID, target.ID, grant.ID))
	refreshed := viewerCtx(t, user.ID, tn.group.ID, false)
	_, err = tClient.Entity.Get(refreshed, target.ID)
	require.True(t, ent.IsNotFound(err))
}

func TestEnforcement_PermissionGroup(t *testing.T) {
	tn := newAuthzTenant(t)
	e := newAuthzEntity(t, tn.group.ID)

	user := newAuthzMember(t, tn.group.ID, []string{"notifier:manage"}, false)

	// Permission group conveys tenant-wide entity:read.
	pg, err := tRepos.Permissions.PermissionGroupCreate(testCtx(), tn.group.ID, PermissionGroupCreate{
		Name:        "readers-" + fk.Str(6),
		Permissions: []string{"entity:read"},
	})
	require.NoError(t, err)
	_, err = tRepos.Permissions.PermissionGroupSetMembers(testCtx(), tn.group.ID, pg.ID, []uuid.UUID{user.ID})
	require.NoError(t, err)

	ctx := viewerCtx(t, user.ID, tn.group.ID, false)
	got, err := tClient.Entity.Get(ctx, e.ID)
	require.NoError(t, err)
	require.Equal(t, e.ID, got.ID)

	// Row grants can target permission groups too.
	tn2 := newAuthzTenant(t)
	_ = tn2 // separate tenant sanity: pg from tenant A cannot be used in B
	e2 := newAuthzEntity(t, tn.group.ID)
	pg2, err := tRepos.Permissions.PermissionGroupCreate(testCtx(), tn.group.ID, PermissionGroupCreate{
		Name:        "editors-" + fk.Str(6),
		Permissions: []string{}, // no tenant-wide perms; row grant only
	})
	require.NoError(t, err)
	_, err = tRepos.Permissions.PermissionGroupSetMembers(testCtx(), tn.group.ID, pg2.ID, []uuid.UUID{user.ID})
	require.NoError(t, err)

	grantEntity(t, tn.group.ID, e2.ID, AccessGrantCreate{
		TargetType: AccessGrantTargetPermissionGroup,
		TargetID:   pg2.ID,
		Actions:    []string{"update"},
	}, authz.GrantActions{Read: true, Update: true})

	ctx = viewerCtx(t, user.ID, tn.group.ID, false)
	require.NoError(t, tClient.Entity.UpdateOneID(e2.ID).SetNotes("via pgroup grant").Exec(ctx))
}

func TestEnforcement_ManageGates(t *testing.T) {
	tn := newAuthzTenant(t)

	user := newAuthzMember(t, tn.group.ID, []string{"entity:read", "entity:create", "entity:update"}, false)
	ctx := viewerCtx(t, user.ID, tn.group.ID, false)

	// No tag:manage: creating a tag is denied by the privacy layer.
	_, err := tClient.Tag.Create().
		SetName("nope").
		SetGroupID(tn.group.ID).
		Save(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, privacy.Deny, "expected privacy deny, got %v", err)

	// tag:manage in tenant A cannot touch tenant B's tags (write pin).
	b := newAuthzTenant(t)
	bTag, err := tRepos.Tags.Create(testCtx(), b.group.ID, TagCreate{Name: "b-tag-" + fk.Str(6)})
	require.NoError(t, err)

	manager := newAuthzMember(t, tn.group.ID, []string{"tag:manage"}, false)
	mCtx := viewerCtx(t, manager.ID, tn.group.ID, false)
	err = tClient.Tag.UpdateOneID(bTag.ID).SetName("hijacked").Exec(mCtx)
	require.True(t, ent.IsNotFound(err), "expected NotFound, got %v", err)
}

func TestEnforcement_LastAdmin(t *testing.T) {
	tn := newAuthzTenant(t)
	member := newAuthzMember(t, tn.group.ID, []string{"entity:read"}, false)

	adminCtx := viewerCtx(t, tn.admin.ID, tn.group.ID, false)

	// Demoting the only admin is rejected.
	err := tRepos.Permissions.MemberPermissionsSet(adminCtx, tn.group.ID, tn.admin.ID, []string{"entity:read"})
	require.ErrorIs(t, err, ErrLastAdmin)

	// Removing the only admin is rejected while other members remain.
	err = tRepos.Groups.RemoveMember(adminCtx, tn.group.ID, tn.admin.ID)
	require.ErrorIs(t, err, ErrLastAdmin)

	// Promote the member, then the original admin can step down.
	require.NoError(t, tRepos.Permissions.MemberPermissionsSet(adminCtx, tn.group.ID, member.ID, authz.AllStrings()))
	require.NoError(t, tRepos.Permissions.MemberPermissionsSet(adminCtx, tn.group.ID, tn.admin.ID, []string{"entity:read"}))
}

func TestEnforcement_SuperuserBypassesPermissionsNotTenancy(t *testing.T) {
	a := newAuthzTenant(t)
	b := newAuthzTenant(t)

	eA := newAuthzEntity(t, a.group.ID)
	eB := newAuthzEntity(t, b.group.ID)

	// A superuser member of A with no permissions at all.
	su := newAuthzMember(t, a.group.ID, []string{"notifier:manage"}, false)
	ctx := viewerCtx(t, su.ID, a.group.ID, true)

	// Permission checks pass...
	got, err := tClient.Entity.Get(ctx, eA.ID)
	require.NoError(t, err)
	require.Equal(t, eA.ID, got.ID)
	require.NoError(t, tClient.Entity.UpdateOneID(eA.ID).SetNotes("superuser was here").Exec(ctx))

	// ...but reads stay scoped to the active tenant.
	_, err = tClient.Entity.Get(ctx, eB.ID)
	require.True(t, ent.IsNotFound(err))
}

func TestEnforcement_InvitationEscalationDenied(t *testing.T) {
	tn := newAuthzTenant(t)

	inviter := newAuthzMember(t, tn.group.ID, []string{"members:manage", "entity:read"}, false)
	ctx := viewerCtx(t, inviter.ID, tn.group.ID, false)

	// Inviting with permissions the inviter does not hold is denied.
	_, err := tRepos.Groups.InvitationCreate(ctx, tn.group.ID, GroupInvitationCreate{
		Token:       []byte(fk.Str(32)),
		ExpiresAt:   time.Now().Add(time.Hour),
		Uses:        1,
		Permissions: []string{"settings:manage"},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, privacy.Deny, "expected privacy deny, got %v", err)

	// Inviting within the inviter's own permissions works.
	_, err = tRepos.Groups.InvitationCreate(ctx, tn.group.ID, GroupInvitationCreate{
		Token:       []byte(fk.Str(32)),
		ExpiresAt:   time.Now().Add(time.Hour),
		Uses:        1,
		Permissions: []string{"entity:read"},
	})
	require.NoError(t, err)
}

func TestEnforcement_PermissionManagementGate(t *testing.T) {
	tn := newAuthzTenant(t)
	user := newAuthzMember(t, tn.group.ID, []string{"entity:read"}, false)
	ctx := viewerCtx(t, user.ID, tn.group.ID, false)

	// Without permissions:manage, creating a permission group is denied.
	_, err := tRepos.Permissions.PermissionGroupCreate(ctx, tn.group.ID, PermissionGroupCreate{
		Name: "sneaky-" + fk.Str(6),
	})
	require.Error(t, err)
	require.ErrorIs(t, err, privacy.Deny, "expected privacy deny, got %v", err)

	// Same for self-escalation of direct permissions.
	err = tRepos.Permissions.MemberPermissionsSet(ctx, tn.group.ID, user.ID, authz.AllStrings())
	require.Error(t, err)
	require.ErrorIs(t, err, privacy.Deny, "expected privacy deny, got %v", err)
}

func TestEnforcement_RepoMutationsSurfaceNotFound(t *testing.T) {
	tn := newAuthzTenant(t)
	e := newAuthzEntity(t, tn.group.ID)

	reader := newAuthzMember(t, tn.group.ID, []string{"entity:read"}, false)
	ctx := viewerCtx(t, reader.ID, tn.group.ID, false)

	// Repo-level update by a read-only viewer surfaces NotFound (not a
	// silent no-op) and leaves the row untouched.
	_, err := tRepos.Entities.UpdateByGroup(ctx, tn.group.ID, EntityUpdate{
		ID:       e.ID,
		Name:     "should not stick",
		Quantity: 1,
	})
	require.True(t, ent.IsNotFound(err), "expected NotFound, got %v", err)

	// Delete: same, and no side effects ran.
	err = tRepos.Entities.DeleteByGroup(ctx, tn.group.ID, e.ID)
	require.True(t, ent.IsNotFound(err), "expected NotFound, got %v", err)

	got, err := tRepos.Entities.GetOne(testCtx(), e.ID)
	require.NoError(t, err)
	require.Equal(t, e.Name, got.Name, "row must be unchanged")
}

func TestEnforcement_WildcardMemberships(t *testing.T) {
	tn := newAuthzTenant(t)
	e := newAuthzEntity(t, tn.group.ID)

	// The tenant admin's membership is stored as the wildcard, not an
	// enumerated snapshot — this is what keeps full-access members covered
	// when new permissions are added to the catalog.
	mp, err := tRepos.Permissions.MemberPermissionsGet(testCtx(), tn.group.ID, tn.admin.ID)
	require.NoError(t, err)
	require.Equal(t, []string{authz.Wildcard}, mp.Direct, "full access must be stored as the wildcard")
	require.Len(t, mp.Effective, len(authz.All()), "effective set expands to the full catalog")

	// A wildcard holder counts as a permissions administrator.
	holders, err := tRepos.Permissions.AdminHolders(testCtx(), tn.group.ID)
	require.NoError(t, err)
	require.Equal(t, 1, holders)

	// Resource wildcards work end to end: entity:* allows entity writes but
	// not tag management.
	user := newAuthzMember(t, tn.group.ID, []string{"entity:*"}, false)
	ctx := viewerCtx(t, user.ID, tn.group.ID, false)
	require.NoError(t, tClient.Entity.UpdateOneID(e.ID).SetNotes("entity:* holder").Exec(ctx))
	_, err = tClient.Tag.Create().SetName("nope").SetGroupID(tn.group.ID).Save(ctx)
	require.True(t, errors.Is(err, privacy.Deny), "expected privacy deny, got %v", err)
}
