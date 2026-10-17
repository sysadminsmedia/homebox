package services

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// adminFixture builds a collection with a full-access owner plus a member
// whose permissions are restricted to entity CRUD, mirroring the shape
// produced by accepting a restricted invitation. Contexts carry real viewers
// resolved from the stored memberships.
type adminFixture struct {
	group     repo.Group
	ownerCtx  Context
	memberCtx Context
}

// newViewerContext resolves the viewer for (uid, gid) from the database and
// returns a service context carrying it, as the viewer middleware would.
func newViewerContext(t *testing.T, uid, gid uuid.UUID) Context {
	t.Helper()

	v, err := tRepos.Permissions.ResolveViewer(testCtx(), uid, gid, false)
	require.NoError(t, err)

	return Context{Context: authz.NewContext(t.Context(), v), GID: gid, UID: uid, Viewer: v}
}

func newMemberWithPermissions(t *testing.T, gid uuid.UUID, perms ...authz.Permission) Context {
	t.Helper()

	stored := make([]string, len(perms))
	for i, p := range perms {
		stored[i] = string(p)
	}

	member, err := tRepos.Users.Create(testCtx(), repo.UserCreate{
		Name:           fk.Str(10),
		Email:          fk.Email(),
		Password:       new(fk.Str(10)),
		DefaultGroupID: gid,
		Permissions:    stored,
	})
	require.NoError(t, err)

	return newViewerContext(t, member.ID, gid)
}

func newAdminFixture(t *testing.T) adminFixture {
	t.Helper()

	group, err := tRepos.Groups.GroupCreate(testCtx(), fk.Str(10), uuid.Nil)
	require.NoError(t, err)

	owner, err := tRepos.Users.Create(testCtx(), repo.UserCreate{
		Name:           fk.Str(10),
		Email:          fk.Email(),
		Password:       new(fk.Str(10)),
		DefaultGroupID: group.ID,
		IsOwner:        true,
	})
	require.NoError(t, err)

	return adminFixture{
		group:    group,
		ownerCtx: newViewerContext(t, owner.ID, group.ID),
		memberCtx: newMemberWithPermissions(t, group.ID,
			authz.PermEntityRead, authz.PermEntityCreate, authz.PermEntityUpdate, authz.PermEntityDelete),
	}
}

// assertForbidden asserts the error is a 403 rather than a generic failure, so
// a member is told they lack permission instead of the action half-succeeding.
func assertForbidden(t *testing.T, err error, perm authz.Permission) {
	t.Helper()

	require.Error(t, err)

	var reqErr *validate.RequestError
	require.ErrorAs(t, err, &reqErr, "expected a validate.RequestError, got %T: %v", err, err)
	assert.Equal(t, http.StatusForbidden, reqErr.Status)
	assert.Contains(t, reqErr.Error(), string(perm))
}

func TestGroupService_UpdateGroup_MemberForbidden(t *testing.T) {
	f := newAdminFixture(t)

	_, err := tSvc.Group.UpdateGroup(f.memberCtx, repo.GroupUpdate{
		Name:     "member-controlled",
		Currency: "eur",
	})
	assertForbidden(t, err, authz.PermSettingsManage)

	// The collection must be untouched by the rejected request.
	after, err := tRepos.Groups.GroupByID(testCtx(), f.group.ID)
	require.NoError(t, err)
	assert.Equal(t, f.group.Name, after.Name)
	assert.Equal(t, f.group.Currency, after.Currency)
}

func TestGroupService_UpdateGroup_OwnerAllowed(t *testing.T) {
	f := newAdminFixture(t)

	updated, err := tSvc.Group.UpdateGroup(f.ownerCtx, repo.GroupUpdate{
		Name:     "owner-renamed",
		Currency: "eur",
	})
	require.NoError(t, err)
	assert.Equal(t, "owner-renamed", updated.Name)
	assert.Equal(t, "EUR", updated.Currency)
}

func TestGroupService_UpdateGroup_DelegatedSettingsAllowed(t *testing.T) {
	f := newAdminFixture(t)
	settingsCtx := newMemberWithPermissions(t, f.group.ID, authz.PermSettingsManage)

	updated, err := tSvc.Group.UpdateGroup(settingsCtx, repo.GroupUpdate{
		Name:     "delegated-rename",
		Currency: "usd",
	})
	require.NoError(t, err)
	assert.Equal(t, "delegated-rename", updated.Name)
}

func TestGroupService_NewInvitation_MemberForbidden(t *testing.T) {
	f := newAdminFixture(t)

	_, token, err := tSvc.Group.NewInvitation(f.memberCtx, 1, time.Now().Add(time.Hour), nil)
	assertForbidden(t, err, authz.PermMembersManage)
	assert.Empty(t, token, "no invitation token may be minted without members:manage")

	invitations, err := tRepos.Groups.InvitationGetAll(testCtx(), f.group.ID)
	require.NoError(t, err)
	assert.Empty(t, invitations)
}

func TestGroupService_NewInvitation_OwnerAllowed(t *testing.T) {
	f := newAdminFixture(t)

	invitation, token, err := tSvc.Group.NewInvitation(f.ownerCtx, 1, time.Now().Add(time.Hour), nil)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, 1, invitation.Uses)
}

// An invitation created without explicit permissions grants the ordinary
// member list, never collection administration.
func TestGroupService_NewInvitation_DefaultsToMemberPermissions(t *testing.T) {
	f := newAdminFixture(t)

	invitation, _, err := tSvc.Group.NewInvitation(f.ownerCtx, 1, time.Now().Add(time.Hour), nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, authz.DefaultMember(), invitation.Permissions)
	assert.False(t, authz.SetHas(invitation.Permissions, authz.PermMembersManage))
	assert.False(t, authz.SetHas(invitation.Permissions, authz.PermSettingsManage))
}

func TestGroupService_DeleteInvitation_MemberForbidden(t *testing.T) {
	f := newAdminFixture(t)

	invitation, _, err := tSvc.Group.NewInvitation(f.ownerCtx, 1, time.Now().Add(time.Hour), nil)
	require.NoError(t, err)

	err = tSvc.Group.DeleteInvitation(f.memberCtx, invitation.ID)
	assertForbidden(t, err, authz.PermMembersManage)

	// The owner's invitation must survive the revoke attempt.
	invitations, err := tRepos.Groups.InvitationGetAll(testCtx(), f.group.ID)
	require.NoError(t, err)
	require.Len(t, invitations, 1)
	assert.Equal(t, invitation.ID, invitations[0].ID)
}

func TestGroupService_DeleteInvitation_OwnerAllowed(t *testing.T) {
	f := newAdminFixture(t)

	invitation, _, err := tSvc.Group.NewInvitation(f.ownerCtx, 1, time.Now().Add(time.Hour), nil)
	require.NoError(t, err)

	require.NoError(t, tSvc.Group.DeleteInvitation(f.ownerCtx, invitation.ID))

	invitations, err := tRepos.Groups.InvitationGetAll(testCtx(), f.group.ID)
	require.NoError(t, err)
	assert.Empty(t, invitations)
}

func TestGroupService_RemoveMember_MemberForbidden(t *testing.T) {
	f := newAdminFixture(t)

	err := tSvc.Group.RemoveMember(f.memberCtx, f.ownerCtx.UID)
	assertForbidden(t, err, authz.PermMembersManage)

	// The owner must still be a member, and still the owner.
	isOwner, err := tRepos.Groups.IsOwnerOf(testCtx(), f.ownerCtx.UID, f.group.ID)
	require.NoError(t, err)
	assert.True(t, isOwner)
}

func TestGroupService_RemoveMember_SelfLeaveAllowed(t *testing.T) {
	f := newAdminFixture(t)

	require.NoError(t, tSvc.Group.RemoveMember(f.memberCtx, f.memberCtx.UID))

	isMember, err := tRepos.Groups.IsMember(testCtx(), f.group.ID, f.memberCtx.UID)
	require.NoError(t, err)
	assert.False(t, isMember, "member should have left the collection")
}

func TestGroupService_DeleteGroup_MemberForbidden(t *testing.T) {
	f := newAdminFixture(t)

	err := tSvc.Group.DeleteGroup(f.memberCtx)
	assertForbidden(t, err, authz.PermCollectionDelete)

	// The collection must still exist.
	after, err := tRepos.Groups.GroupByID(testCtx(), f.group.ID)
	require.NoError(t, err)
	assert.Equal(t, f.group.ID, after.ID)
}

// settings:manage covers renaming and currency, not destroying the collection.
func TestGroupService_DeleteGroup_SettingsManagerForbidden(t *testing.T) {
	f := newAdminFixture(t)
	settingsCtx := newMemberWithPermissions(t, f.group.ID, authz.PermSettingsManage)

	err := tSvc.Group.DeleteGroup(settingsCtx)
	assertForbidden(t, err, authz.PermCollectionDelete)

	after, err := tRepos.Groups.GroupByID(testCtx(), f.group.ID)
	require.NoError(t, err)
	assert.Equal(t, f.group.ID, after.ID)
}

func TestGroupService_DeleteGroup_OwnerAllowed(t *testing.T) {
	f := newAdminFixture(t)

	require.NoError(t, tSvc.Group.DeleteGroup(f.ownerCtx))

	_, err := tRepos.Groups.GroupByID(testCtx(), f.group.ID)
	require.Error(t, err)
}
