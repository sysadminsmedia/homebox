package services

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// PermissionService exposes permission group management, direct member
// permissions, and row-level access grants. Authorization is enforced by the
// ent privacy layer; this service adds input validation and friendly error
// translation (last-admin -> 409, duplicates -> 409, bad keys -> 422).
type PermissionService struct {
	repos *repo.AllRepos
}

// PermissionDefinition is one catalog entry, used by the UI to render the
// resource x action matrix.
type PermissionDefinition struct {
	Key      string `json:"key"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// EffectivePermissionsOut is the caller's resolved permission picture for the
// active tenant.
type EffectivePermissionsOut struct {
	GroupID     uuid.UUID `json:"groupId"`
	Permissions []string  `json:"permissions"`
	IsOwner     bool      `json:"isOwner"`
	IsSuperuser bool      `json:"isSuperuser"`
}

// Catalog returns the full permission catalog.
func (svc *PermissionService) Catalog() []PermissionDefinition {
	all := authz.All()
	out := make([]PermissionDefinition, 0, len(all))
	for _, p := range all {
		resource, action, _ := strings.Cut(string(p), ":")
		out = append(out, PermissionDefinition{
			Key:      string(p),
			Resource: resource,
			Action:   action,
		})
	}
	return out
}

// Self returns the caller's effective permissions for the active tenant.
func (svc *PermissionService) Self(ctx Context) (EffectivePermissionsOut, error) {
	if ctx.Viewer == nil {
		return EffectivePermissionsOut{}, errors.New("no viewer resolved for request")
	}
	isOwner, err := svc.repos.Groups.IsOwnerOf(ctx, ctx.UID, ctx.GID)
	if err != nil {
		return EffectivePermissionsOut{}, err
	}
	return EffectivePermissionsOut{
		GroupID:     ctx.GID,
		Permissions: ctx.Viewer.PermStrings(),
		IsOwner:     isOwner,
		IsSuperuser: ctx.Viewer.Superuser,
	}, nil
}

func validatePermissionKeys(perms []string) error {
	if invalid := authz.ValidateStrings(perms); len(invalid) > 0 {
		return validate.NewFieldErrors(validate.FieldError{
			Field: "permissions",
			Error: "unknown permissions: " + strings.Join(invalid, ", "),
		})
	}
	return nil
}

// translateErr maps repository errors onto HTTP-aware errors.
func translatePermissionErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repo.ErrLastAdmin):
		return validate.NewRequestError(err, http.StatusConflict)
	case ent.IsConstraintError(err):
		return validate.NewRequestError(errors.New("a conflicting entry already exists"), http.StatusConflict)
	default:
		return err
	}
}

// --- Permission groups ------------------------------------------------------

func (svc *PermissionService) GetPermissionGroups(ctx Context) ([]repo.PermissionGroupOut, error) {
	return svc.repos.Permissions.PermissionGroupGetAll(ctx, ctx.GID)
}

func (svc *PermissionService) GetPermissionGroup(ctx Context, id uuid.UUID) (repo.PermissionGroupOut, error) {
	return svc.repos.Permissions.PermissionGroupGetOne(ctx, ctx.GID, id)
}

func (svc *PermissionService) CreatePermissionGroup(ctx Context, data repo.PermissionGroupCreate) (repo.PermissionGroupOut, error) {
	if err := validatePermissionKeys(data.Permissions); err != nil {
		return repo.PermissionGroupOut{}, err
	}
	out, err := svc.repos.Permissions.PermissionGroupCreate(ctx, ctx.GID, data)
	return out, translatePermissionErr(err)
}

func (svc *PermissionService) UpdatePermissionGroup(ctx Context, id uuid.UUID, data repo.PermissionGroupUpdate) (repo.PermissionGroupOut, error) {
	if err := validatePermissionKeys(data.Permissions); err != nil {
		return repo.PermissionGroupOut{}, err
	}
	out, err := svc.repos.Permissions.PermissionGroupUpdate(ctx, ctx.GID, id, data)
	return out, translatePermissionErr(err)
}

func (svc *PermissionService) DeletePermissionGroup(ctx Context, id uuid.UUID) error {
	return translatePermissionErr(svc.repos.Permissions.PermissionGroupDelete(ctx, ctx.GID, id))
}

func (svc *PermissionService) SetPermissionGroupMembers(ctx Context, id uuid.UUID, userIDs []uuid.UUID) (repo.PermissionGroupOut, error) {
	out, err := svc.repos.Permissions.PermissionGroupSetMembers(ctx, ctx.GID, id, userIDs)
	return out, translatePermissionErr(err)
}

// --- Direct member permissions ----------------------------------------------

func (svc *PermissionService) GetMemberPermissions(ctx Context, userID uuid.UUID) (repo.MemberPermissions, error) {
	return svc.repos.Permissions.MemberPermissionsGet(ctx, ctx.GID, userID)
}

func (svc *PermissionService) SetMemberPermissions(ctx Context, userID uuid.UUID, perms []string) (repo.MemberPermissions, error) {
	if err := validatePermissionKeys(perms); err != nil {
		return repo.MemberPermissions{}, err
	}
	if err := translatePermissionErr(svc.repos.Permissions.MemberPermissionsSet(ctx, ctx.GID, userID, perms)); err != nil {
		return repo.MemberPermissions{}, err
	}
	return svc.repos.Permissions.MemberPermissionsGet(ctx, ctx.GID, userID)
}

// --- Row-level access grants -------------------------------------------------

func (svc *PermissionService) GetEntityGrants(ctx Context, entityID uuid.UUID) ([]repo.AccessGrantOut, error) {
	return svc.repos.Permissions.GrantsByEntity(ctx, ctx.GID, entityID)
}

func (svc *PermissionService) CreateEntityGrant(ctx Context, entityID uuid.UUID, data repo.AccessGrantCreate) (repo.AccessGrantOut, error) {
	actions, err := parseGrantActions(data.Actions)
	if err != nil {
		return repo.AccessGrantOut{}, err
	}
	out, err := svc.repos.Permissions.GrantCreate(ctx, ctx.GID, entityID, data, actions)
	return out, translatePermissionErr(err)
}

func (svc *PermissionService) UpdateEntityGrant(ctx Context, entityID, grantID uuid.UUID, actionStrs []string) (repo.AccessGrantOut, error) {
	actions, err := parseGrantActions(actionStrs)
	if err != nil {
		return repo.AccessGrantOut{}, err
	}
	out, err := svc.repos.Permissions.GrantUpdate(ctx, ctx.GID, entityID, grantID, actions)
	return out, translatePermissionErr(err)
}

func (svc *PermissionService) DeleteEntityGrant(ctx Context, entityID, grantID uuid.UUID) error {
	return translatePermissionErr(svc.repos.Permissions.GrantDelete(ctx, ctx.GID, entityID, grantID))
}

func parseGrantActions(actionStrs []string) (authz.GrantActions, error) {
	actions, invalid := authz.GrantActionsFromStrings(actionStrs)
	if len(invalid) > 0 {
		return authz.GrantActions{}, validate.NewFieldErrors(validate.FieldError{
			Field: "actions",
			Error: "unknown actions: " + strings.Join(invalid, ", "),
		})
	}
	if !actions.Any() {
		return authz.GrantActions{}, validate.NewFieldErrors(validate.FieldError{
			Field: "actions",
			Error: "at least one action is required",
		})
	}
	return actions, nil
}
