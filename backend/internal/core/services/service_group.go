package services

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

type GroupService struct {
	repos *repo.AllRepos
}

// requirePermission asserts the acting viewer holds perm in the active
// collection. Route middleware (mwPermission) is the first line of defense and
// the ent privacy layer the source of truth; this is the service-level
// backstop, so an admin action wired onto ordinary member middleware can't
// silently reintroduce a bypass.
func requirePermission(ctx Context, perm authz.Permission) error {
	if !ctx.Viewer.Has(perm) {
		return validate.NewRequestError(errors.New("missing required permission: "+string(perm)), http.StatusForbidden)
	}
	return nil
}

func (svc *GroupService) UpdateGroup(ctx Context, data repo.GroupUpdate) (repo.Group, error) {
	if err := requirePermission(ctx, authz.PermSettingsManage); err != nil {
		return repo.Group{}, err
	}

	if data.Name == "" {
		return repo.Group{}, errors.New("group name cannot be empty")
	}

	if data.Currency == "" {
		return repo.Group{}, errors.New("currency cannot be empty")
	}

	return svc.repos.Groups.GroupUpdate(ctx.Context, ctx.GID, data)
}

func (svc *GroupService) CreateGroup(ctx Context, name string) (repo.Group, error) {
	if name == "" {
		return repo.Group{}, errors.New("group name cannot be empty")
	}

	if ctx.UID == uuid.Nil {
		return repo.Group{}, errors.New("user ID cannot be empty when creating a group")
	}

	// Creating a tenant also creates the owner membership in the new tenant,
	// where the caller has no viewer yet.
	sysCtx := authz.NewSystemContext(ctx.Context)
	group, err := svc.repos.Groups.GroupCreate(sysCtx, name, ctx.UID)
	if err != nil {
		return repo.Group{}, err
	}

	// Unlike registration, this path doesn't seed default locations/tags, so
	// nothing would lazily create the entity types — leaving the collection
	// unable to create items or locations until a type was added by hand.
	if err := ensureDefaultEntityTypes(sysCtx, svc.repos, group.ID); err != nil {
		return repo.Group{}, err
	}

	return group, nil
}

func (svc *GroupService) DeleteGroup(ctx Context) error {
	if err := requirePermission(ctx, authz.PermCollectionDelete); err != nil {
		return err
	}

	return svc.repos.Groups.GroupDelete(ctx.Context, ctx.GID)
}

// NewInvitation creates an invitation token. permissions selects the
// tenant-wide permissions applied to the membership created on acceptance;
// nil/empty means the default member list (authz.DefaultMember), never admin
// access. The ORM layer
// additionally rejects invitations carrying permissions the inviter does not
// hold (unless the inviter has permissions:manage).
func (svc *GroupService) NewInvitation(ctx Context, uses int, expiresAt time.Time, permissions []string) (repo.GroupInvitation, string, error) {
	if err := requirePermission(ctx, authz.PermMembersManage); err != nil {
		return repo.GroupInvitation{}, "", err
	}
	if invalid := authz.ValidateStrings(permissions); len(invalid) > 0 {
		return repo.GroupInvitation{}, "", validate.NewFieldErrors(validate.FieldError{
			Field: "permissions",
			Error: "unknown permissions: " + strings.Join(invalid, ", "),
		})
	}
	if len(permissions) == 0 {
		permissions = authz.DefaultMember()
	}

	token := hasher.GenerateToken()

	invitation, err := svc.repos.Groups.InvitationCreate(ctx, ctx.GID, repo.GroupInvitationCreate{
		Token:       token.Hash,
		Uses:        uses,
		ExpiresAt:   expiresAt,
		Permissions: permissions,
	})
	if err != nil {
		return repo.GroupInvitation{}, "", err
	}

	return invitation, token.Raw, nil
}

func (svc *GroupService) RemoveMember(ctx Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	// Leaving a collection is always allowed; removing someone else is not.
	if userID != ctx.UID {
		if err := requirePermission(ctx, authz.PermMembersManage); err != nil {
			return err
		}
	}

	removeCtx := ctx.Context
	if userID == ctx.UID {
		// Self-leave: any member may leave a tenant without members:manage.
		// The privacy layer would deny the membership delete, so this runs
		// under a system context after the identity check above.
		removeCtx = authz.NewSystemContext(removeCtx)
	}

	err := svc.repos.Groups.RemoveMember(removeCtx, ctx.GID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrLastAdmin) {
			return validate.NewRequestError(err, http.StatusConflict)
		}
		return err
	}

	// Post-removal bookkeeping runs under a system context: the removed user
	// may no longer share a tenant with the caller and would be invisible.
	sysCtx := authz.NewSystemContext(ctx.Context)

	// If the removed group was the user's default group, reassign to another group
	removedUser, err := svc.repos.Users.GetOneID(sysCtx, userID)
	if err != nil {
		return err
	}

	if removedUser.DefaultGroupID == ctx.GID {
		// Find another group the user is still a member of
		var newDefaultGroupID uuid.UUID
		for _, gid := range removedUser.GroupIDs {
			if gid != ctx.GID {
				newDefaultGroupID = gid
				break
			}
		}
		// Update to another group, or uuid.Nil if the user has no remaining groups
		if err := svc.repos.Users.UpdateDefaultGroup(sysCtx, userID, newDefaultGroupID); err != nil {
			return err
		}
	}

	return nil
}

func (svc *GroupService) DeleteInvitation(ctx Context, id uuid.UUID) error {
	if err := requirePermission(ctx, authz.PermMembersManage); err != nil {
		return err
	}

	return svc.repos.Groups.InvitationDelete(ctx.Context, ctx.GID, id)
}

func (svc *GroupService) AcceptInvitation(ctx Context, token string) (repo.Group, error) {
	hashedToken := hasher.HashToken(token)
	// The invitation token is the authorization: it creates a membership in a
	// tenant the caller is not yet a member of.
	return svc.repos.Groups.InvitationAccept(authz.NewSystemContext(ctx.Context), hashedToken, ctx.UID)
}
