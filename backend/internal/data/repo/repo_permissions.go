package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authzrules"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/accessgrant"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/permissiongroup"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/user"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/usergroup"
)

// ErrLastAdmin is returned when a change would leave a tenant without any
// member holding permissions:manage. Callers translate it to HTTP 409.
var ErrLastAdmin = errors.New("cannot remove the last administrator of this group")

// AccessGrantTarget* are the API values for AccessGrantCreate.TargetType.
const (
	AccessGrantTargetUser            = "user"
	AccessGrantTargetPermissionGroup = "permissionGroup"
)

type (
	// PermissionsRepository manages permission groups, direct membership
	// permissions, and row-level access grants. Authorization for these
	// operations is enforced by the ent privacy layer (permissions:manage);
	// the repository additionally enforces the last-admin invariant inside
	// transactions.
	PermissionsRepository struct {
		db *ent.Client
	}

	PermissionGroupCreate struct {
		Name        string   `json:"name"        validate:"required,max=255"`
		Description string   `json:"description" validate:"max=1000"`
		Permissions []string `json:"permissions"`
	}

	PermissionGroupUpdate struct {
		Name        string   `json:"name"        validate:"required,max=255"`
		Description string   `json:"description" validate:"max=1000"`
		Permissions []string `json:"permissions"`
	}

	PermissionGroupSummary struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Permissions []string  `json:"permissions"`
	}

	PermissionGroupOut struct {
		ID          uuid.UUID     `json:"id"`
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Permissions []string      `json:"permissions"`
		Members     []UserSummary `json:"members"`
		CreatedAt   time.Time     `json:"createdAt"`
		UpdatedAt   time.Time     `json:"updatedAt"`
	}

	// MemberPermissions is the full permission picture for one tenant member.
	MemberPermissions struct {
		UserID           uuid.UUID                `json:"userId"`
		Role             string                   `json:"role"`
		Direct           []string                 `json:"direct"`
		PermissionGroups []PermissionGroupSummary `json:"permissionGroups"`
		Effective        []string                 `json:"effective"`
	}

	AccessGrantCreate struct {
		TargetType string    `json:"targetType" validate:"required,oneof=user permissionGroup" enums:"user,permissionGroup"`
		TargetID   uuid.UUID `json:"targetId"   validate:"required"`
		Actions    []string  `json:"actions"    validate:"required,min=1"`
	}

	AccessGrantOut struct {
		ID         uuid.UUID `json:"id"`
		EntityID   uuid.UUID `json:"entityId"`
		TargetType string    `json:"targetType" enums:"user,permissionGroup"`
		TargetID   uuid.UUID `json:"targetId"`
		TargetName string    `json:"targetName"`
		Actions    []string  `json:"actions"`
		CreatedAt  time.Time `json:"createdAt"`
	}
)

func mapPermissionGroupSummary(pg *ent.PermissionGroup) PermissionGroupSummary {
	return PermissionGroupSummary{
		ID:          pg.ID,
		Name:        pg.Name,
		Permissions: pg.Permissions,
	}
}

func mapPermissionGroupOut(pg *ent.PermissionGroup) PermissionGroupOut {
	members := make([]UserSummary, 0, len(pg.Edges.Users))
	for _, u := range pg.Edges.Users {
		members = append(members, mapUserSummary(u))
	}
	return PermissionGroupOut{
		ID:          pg.ID,
		Name:        pg.Name,
		Description: pg.Description,
		Permissions: pg.Permissions,
		Members:     members,
		CreatedAt:   pg.CreatedAt,
		UpdatedAt:   pg.UpdatedAt,
	}
}

func mapAccessGrantOut(g *ent.AccessGrant) AccessGrantOut {
	out := AccessGrantOut{
		ID:       g.ID,
		EntityID: g.EntityID,
		Actions: authz.GrantActions{
			Read:        g.CanRead,
			Update:      g.CanUpdate,
			Delete:      g.CanDelete,
			Attachments: g.CanAttachments,
		}.Strings(),
		CreatedAt: g.CreatedAt,
	}
	switch {
	case g.UserID != nil:
		out.TargetType = AccessGrantTargetUser
		out.TargetID = *g.UserID
		if g.Edges.User != nil {
			out.TargetName = g.Edges.User.Name
		}
	case g.PermissionGroupID != nil:
		out.TargetType = AccessGrantTargetPermissionGroup
		out.TargetID = *g.PermissionGroupID
		if g.Edges.PermissionGroup != nil {
			out.TargetName = g.Edges.PermissionGroup.Name
		}
	}
	return out
}

// ResolveViewer builds the authorization viewer for one (user, tenant) pair:
// the membership's direct permissions unioned with the permissions of every
// permission group the user belongs to in that tenant. It runs under a system
// context because it executes before any viewer exists.
func (r *PermissionsRepository) ResolveViewer(ctx context.Context, uid, gid uuid.UUID, superuser bool) (*authz.Viewer, error) {
	sysCtx := authz.NewSystemContext(ctx)

	membership, err := r.db.UserGroup.Query().
		Where(usergroup.UserID(uid), usergroup.GroupID(gid)).
		Only(sysCtx)
	if err != nil {
		return nil, err // NotFound: not a member of the tenant
	}

	pgroups, err := r.db.PermissionGroup.Query().
		Where(
			permissiongroup.GroupID(gid),
			permissiongroup.HasUsersWith(user.ID(uid)),
		).
		All(sysCtx)
	if err != nil {
		return nil, err
	}

	pgIDs := make([]uuid.UUID, 0, len(pgroups))
	v := authz.NewViewer(uid, gid, superuser, membership.Permissions, nil)
	for _, pg := range pgroups {
		pgIDs = append(pgIDs, pg.ID)
		v.AddPerms(pg.Permissions)
	}
	v.PermGroupIDs = pgIDs
	return v, nil
}

// CountAdminHolders returns how many distinct tenant members hold
// permissions:manage, directly or via a permission group. Members listed in
// excluding are ignored (used to evaluate "what if this user were removed").
// When called from a mutation, pass a client bound to its transaction.
func CountAdminHolders(ctx context.Context, client *ent.Client, gid uuid.UUID, excluding ...uuid.UUID) (int, error) {
	sysCtx := authz.NewSystemContext(ctx)
	holders := map[uuid.UUID]struct{}{}
	excluded := make(map[uuid.UUID]struct{}, len(excluding))
	for _, id := range excluding {
		excluded[id] = struct{}{}
	}

	memberships, err := client.UserGroup.Query().
		Where(usergroup.GroupID(gid)).
		All(sysCtx)
	if err != nil {
		return 0, err
	}
	for _, m := range memberships {
		if _, skip := excluded[m.UserID]; skip {
			continue
		}
		if authz.SetHas(m.Permissions, authz.PermPermissionsManage) {
			holders[m.UserID] = struct{}{}
		}
	}

	pgroups, err := client.PermissionGroup.Query().
		Where(permissiongroup.GroupID(gid)).
		WithUsers().
		All(sysCtx)
	if err != nil {
		return 0, err
	}
	for _, pg := range pgroups {
		if !authz.SetHas(pg.Permissions, authz.PermPermissionsManage) {
			continue
		}
		for _, u := range pg.Edges.Users {
			if _, skip := excluded[u.ID]; skip {
				continue
			}
			holders[u.ID] = struct{}{}
		}
	}

	return len(holders), nil
}

// AdminHolders is CountAdminHolders bound to the repository's client.
func (r *PermissionsRepository) AdminHolders(ctx context.Context, gid uuid.UUID, excluding ...uuid.UUID) (int, error) {
	return CountAdminHolders(ctx, r.db, gid, excluding...)
}

// MemberCount returns the number of members in a tenant.
func (r *PermissionsRepository) MemberCount(ctx context.Context, gid uuid.UUID) (int, error) {
	return r.db.UserGroup.Query().
		Where(usergroup.GroupID(gid)).
		Count(authz.NewSystemContext(ctx))
}

// withTx runs fn inside a transaction, enforcing the last-admin invariant
// after fn's changes: if the tenant would be left without a permissions
// administrator the transaction is rolled back with ErrLastAdmin.
func (r *PermissionsRepository) withTxGuard(ctx context.Context, gid uuid.UUID, fn func(tx *ent.Tx) error) error {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}
	rollback := func(err error) error {
		_ = tx.Rollback()
		return err
	}

	if err := fn(tx); err != nil {
		return rollback(err)
	}

	holders, err := CountAdminHolders(ctx, tx.Client(), gid)
	if err != nil {
		return rollback(err)
	}
	if holders == 0 {
		return rollback(ErrLastAdmin)
	}

	return tx.Commit()
}

// --- Permission groups ------------------------------------------------------

func (r *PermissionsRepository) PermissionGroupGetAll(ctx context.Context, gid uuid.UUID) ([]PermissionGroupOut, error) {
	pgs, err := r.db.PermissionGroup.Query().
		Where(permissiongroup.GroupID(gid)).
		WithUsers().
		Order(ent.Asc(permissiongroup.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PermissionGroupOut, 0, len(pgs))
	for _, pg := range pgs {
		out = append(out, mapPermissionGroupOut(pg))
	}
	return out, nil
}

func (r *PermissionsRepository) PermissionGroupGetOne(ctx context.Context, gid, id uuid.UUID) (PermissionGroupOut, error) {
	pg, err := r.db.PermissionGroup.Query().
		Where(permissiongroup.ID(id), permissiongroup.GroupID(gid)).
		WithUsers().
		Only(ctx)
	if err != nil {
		return PermissionGroupOut{}, err
	}
	return mapPermissionGroupOut(pg), nil
}

func (r *PermissionsRepository) PermissionGroupCreate(ctx context.Context, gid uuid.UUID, data PermissionGroupCreate) (PermissionGroupOut, error) {
	pg, err := r.db.PermissionGroup.Create().
		SetGroupID(gid).
		SetName(data.Name).
		SetDescription(data.Description).
		SetPermissions(data.Permissions).
		Save(ctx)
	if err != nil {
		return PermissionGroupOut{}, err
	}
	return r.PermissionGroupGetOne(ctx, gid, pg.ID)
}

func (r *PermissionsRepository) PermissionGroupUpdate(ctx context.Context, gid, id uuid.UUID, data PermissionGroupUpdate) (PermissionGroupOut, error) {
	err := r.withTxGuard(ctx, gid, func(tx *ent.Tx) error {
		n, err := tx.PermissionGroup.Update().
			Where(permissiongroup.ID(id), permissiongroup.GroupID(gid)).
			SetName(data.Name).
			SetDescription(data.Description).
			SetPermissions(data.Permissions).
			Save(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			return &ent.NotFoundError{}
		}
		return nil
	})
	if err != nil {
		return PermissionGroupOut{}, err
	}
	return r.PermissionGroupGetOne(ctx, gid, id)
}

func (r *PermissionsRepository) PermissionGroupDelete(ctx context.Context, gid, id uuid.UUID) error {
	return r.withTxGuard(ctx, gid, func(tx *ent.Tx) error {
		n, err := tx.PermissionGroup.Delete().
			Where(permissiongroup.ID(id), permissiongroup.GroupID(gid)).
			Exec(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			return &ent.NotFoundError{}
		}
		return nil
	})
}

// PermissionGroupSetMembers replaces the member list of a permission group.
// Every member must already belong to the tenant.
func (r *PermissionsRepository) PermissionGroupSetMembers(ctx context.Context, gid, id uuid.UUID, userIDs []uuid.UUID) (PermissionGroupOut, error) {
	err := r.withTxGuard(ctx, gid, func(tx *ent.Tx) error {
		if _, err := tx.PermissionGroup.Query().
			Where(permissiongroup.ID(id), permissiongroup.GroupID(gid)).
			Only(ctx); err != nil {
			return err
		}

		if len(userIDs) > 0 {
			n, err := tx.User.Query().
				Where(user.IDIn(userIDs...), user.HasGroupsWith(group.ID(gid))).
				Count(authz.NewSystemContext(ctx))
			if err != nil {
				return err
			}
			if n != len(userIDs) {
				return &ent.NotFoundError{} // some user is not a tenant member
			}
		}

		return tx.PermissionGroup.UpdateOneID(id).
			ClearUsers().
			AddUserIDs(userIDs...).
			Exec(ctx)
	})
	if err != nil {
		return PermissionGroupOut{}, err
	}
	return r.PermissionGroupGetOne(ctx, gid, id)
}

// --- Direct member permissions ----------------------------------------------

func (r *PermissionsRepository) MemberPermissionsGet(ctx context.Context, gid, userID uuid.UUID) (MemberPermissions, error) {
	membership, err := r.db.UserGroup.Query().
		Where(usergroup.UserID(userID), usergroup.GroupID(gid)).
		Only(ctx)
	if err != nil {
		return MemberPermissions{}, err
	}

	pgroups, err := r.db.PermissionGroup.Query().
		Where(
			permissiongroup.GroupID(gid),
			permissiongroup.HasUsersWith(user.ID(userID)),
		).
		Order(ent.Asc(permissiongroup.FieldName)).
		All(ctx)
	if err != nil {
		return MemberPermissions{}, err
	}

	out := MemberPermissions{
		UserID: userID,
		Role:   membership.Role.String(),
		Direct: membership.Permissions,
	}

	effective := authz.NewViewer(userID, gid, false, membership.Permissions, nil)
	for _, pg := range pgroups {
		out.PermissionGroups = append(out.PermissionGroups, mapPermissionGroupSummary(pg))
		effective.AddPerms(pg.Permissions)
	}
	out.Effective = effective.PermStrings()
	return out, nil
}

func (r *PermissionsRepository) MemberPermissionsSet(ctx context.Context, gid, userID uuid.UUID, perms []string) error {
	return r.withTxGuard(ctx, gid, func(tx *ent.Tx) error {
		n, err := tx.UserGroup.Update().
			Where(usergroup.UserID(userID), usergroup.GroupID(gid)).
			SetPermissions(perms).
			Save(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			return &ent.NotFoundError{}
		}
		return nil
	})
}

// --- Row-level access grants -------------------------------------------------

func (r *PermissionsRepository) GrantsByEntity(ctx context.Context, gid, entityID uuid.UUID) ([]AccessGrantOut, error) {
	if err := assertEntityInGroup(ctx, r.db.Entity, gid, entityID); err != nil {
		return nil, err
	}
	grants, err := r.db.AccessGrant.Query().
		Where(accessgrant.EntityID(entityID), accessgrant.GroupID(gid)).
		WithUser().
		WithPermissionGroup().
		Order(ent.Asc(accessgrant.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AccessGrantOut, 0, len(grants))
	for _, g := range grants {
		out = append(out, mapAccessGrantOut(g))
	}
	return out, nil
}

func (r *PermissionsRepository) GrantCreate(ctx context.Context, gid, entityID uuid.UUID, data AccessGrantCreate, actions authz.GrantActions) (AccessGrantOut, error) {
	if err := assertEntityInGroup(ctx, r.db.Entity, gid, entityID); err != nil {
		return AccessGrantOut{}, err
	}

	create := r.db.AccessGrant.Create().
		SetGroupID(gid).
		SetEntityID(entityID).
		SetCanRead(actions.Read).
		SetCanUpdate(actions.Update).
		SetCanDelete(actions.Delete).
		SetCanAttachments(actions.Attachments)

	switch data.TargetType {
	case AccessGrantTargetUser:
		// The target must be a member of the tenant.
		ok, err := r.db.User.Query().
			Where(user.ID(data.TargetID), user.HasGroupsWith(group.ID(gid))).
			Exist(authz.NewSystemContext(ctx))
		if err != nil {
			return AccessGrantOut{}, err
		}
		if !ok {
			return AccessGrantOut{}, &ent.NotFoundError{}
		}
		create.SetUserID(data.TargetID)
	case AccessGrantTargetPermissionGroup:
		ok, err := r.db.PermissionGroup.Query().
			Where(permissiongroup.ID(data.TargetID), permissiongroup.GroupID(gid)).
			Exist(authz.NewSystemContext(ctx))
		if err != nil {
			return AccessGrantOut{}, err
		}
		if !ok {
			return AccessGrantOut{}, &ent.NotFoundError{}
		}
		create.SetPermissionGroupID(data.TargetID)
	default:
		return AccessGrantOut{}, errors.New("invalid grant target type")
	}

	g, err := create.Save(ctx)
	if err != nil {
		return AccessGrantOut{}, err
	}
	return r.grantGetOne(ctx, gid, g.ID)
}

func (r *PermissionsRepository) GrantUpdate(ctx context.Context, gid, entityID, grantID uuid.UUID, actions authz.GrantActions) (AccessGrantOut, error) {
	n, err := r.db.AccessGrant.Update().
		Where(
			accessgrant.ID(grantID),
			accessgrant.EntityID(entityID),
			accessgrant.GroupID(gid),
		).
		SetCanRead(actions.Read).
		SetCanUpdate(actions.Update).
		SetCanDelete(actions.Delete).
		SetCanAttachments(actions.Attachments).
		Save(ctx)
	if err != nil {
		return AccessGrantOut{}, err
	}
	if n == 0 {
		return AccessGrantOut{}, &ent.NotFoundError{}
	}
	return r.grantGetOne(ctx, gid, grantID)
}

func (r *PermissionsRepository) GrantDelete(ctx context.Context, gid, entityID, grantID uuid.UUID) error {
	n, err := r.db.AccessGrant.Delete().
		Where(
			accessgrant.ID(grantID),
			accessgrant.EntityID(entityID),
			accessgrant.GroupID(gid),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

func (r *PermissionsRepository) grantGetOne(ctx context.Context, gid, id uuid.UUID) (AccessGrantOut, error) {
	g, err := r.db.AccessGrant.Query().
		Where(accessgrant.ID(id), accessgrant.GroupID(gid)).
		WithUser().
		WithPermissionGroup().
		Only(ctx)
	if err != nil {
		return AccessGrantOut{}, err
	}
	return mapAccessGrantOut(g), nil
}

// EntityCapabilities computes the actions the viewer may perform on one
// entity the viewer can already read: read, update, delete, attachments, and
// permissions (managing the entity's grants). Used to populate
// EntityOut.Capabilities on single-entity responses.
func (r *PermissionsRepository) EntityCapabilities(ctx context.Context, v *authz.Viewer, entityID uuid.UUID) ([]string, error) {
	if v == nil {
		return nil, nil
	}

	var ga authz.GrantActions
	grants, err := r.db.AccessGrant.Query().
		Where(accessgrant.EntityID(entityID), authzrules.GrantTarget(v)).
		All(authz.NewSystemContext(ctx))
	if err != nil {
		return nil, err
	}
	for _, g := range grants {
		ga.Read = ga.Read || g.CanRead
		ga.Update = ga.Update || g.CanUpdate
		ga.Delete = ga.Delete || g.CanDelete
		ga.Attachments = ga.Attachments || g.CanAttachments
	}

	caps := []string{"read"} // the entity was visible to the caller
	if v.Has(authz.PermEntityUpdate) || ga.Update {
		caps = append(caps, "update")
	}
	if v.Has(authz.PermEntityDelete) || ga.Delete {
		caps = append(caps, "delete")
	}
	if v.Has(authz.PermEntityUpdate) || ga.Update || ga.Attachments {
		caps = append(caps, "attachments")
	}
	if v.Has(authz.PermPermissionsManage) {
		caps = append(caps, "permissions")
	}
	return caps, nil
}
