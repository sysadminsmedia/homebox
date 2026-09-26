package services

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

// ErrNotGroupOwner is returned when a member of a collection attempts an action
// reserved for the collection's owner.
var ErrNotGroupOwner = errors.New("only the owner of this collection can perform this action")

type GroupService struct {
	repos *repo.AllRepos
}

// requireOwner asserts the acting user owns the collection identified by
// ctx.GID. The mwGroupOwner route middleware is the first line of defense; this
// is the backstop, so an admin action wired onto ordinary member middleware
// can't silently reintroduce the bypass.
func (svc *GroupService) requireOwner(ctx Context) error {
	isOwner, err := svc.repos.Groups.IsOwnerOf(ctx.Context, ctx.UID, ctx.GID)
	if err != nil {
		return err
	}

	if !isOwner {
		return validate.NewRequestError(ErrNotGroupOwner, http.StatusForbidden)
	}

	return nil
}

func (svc *GroupService) UpdateGroup(ctx Context, data repo.GroupUpdate) (repo.Group, error) {
	if err := svc.requireOwner(ctx); err != nil {
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

	group, err := svc.repos.Groups.GroupCreate(ctx.Context, name, ctx.UID)
	if err != nil {
		return repo.Group{}, err
	}

	// Unlike registration, this path doesn't seed default locations/tags, so
	// nothing would lazily create the entity types — leaving the collection
	// unable to create items or locations until a type was added by hand.
	if err := ensureDefaultEntityTypes(ctx.Context, svc.repos, group.ID); err != nil {
		return repo.Group{}, err
	}

	return group, nil
}

func (svc *GroupService) DeleteGroup(ctx Context) error {
	if err := svc.requireOwner(ctx); err != nil {
		return err
	}

	return svc.repos.Groups.GroupDelete(ctx.Context, ctx.GID)
}

func (svc *GroupService) NewInvitation(ctx Context, uses int, expiresAt time.Time) (repo.GroupInvitation, string, error) {
	if err := svc.requireOwner(ctx); err != nil {
		return repo.GroupInvitation{}, "", err
	}

	token := hasher.GenerateToken()

	invitation, err := svc.repos.Groups.InvitationCreate(ctx, ctx.GID, repo.GroupInvitationCreate{
		Token:     token.Hash,
		Uses:      uses,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return repo.GroupInvitation{}, "", err
	}

	return invitation, token.Raw, nil
}

func (svc *GroupService) RemoveMember(ctx Context, userID uuid.UUID) error {
	if err := svc.requireOwner(ctx); err != nil {
		return err
	}

	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	err := svc.repos.Groups.RemoveMember(ctx.Context, ctx.GID, userID)
	if err != nil {
		return err
	}

	// If the removed group was the user's default group, reassign to another group
	removedUser, err := svc.repos.Users.GetOneID(ctx.Context, userID)
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
		if err := svc.repos.Users.UpdateDefaultGroup(ctx.Context, userID, newDefaultGroupID); err != nil {
			return err
		}
	}

	return nil
}

func (svc *GroupService) DeleteInvitation(ctx Context, id uuid.UUID) error {
	if err := svc.requireOwner(ctx); err != nil {
		return err
	}

	return svc.repos.Groups.InvitationDelete(ctx.Context, ctx.GID, id)
}

func (svc *GroupService) AcceptInvitation(ctx Context, token string) (repo.Group, error) {
	hashedToken := hasher.HashToken(token)
	return svc.repos.Groups.InvitationAccept(ctx.Context, hashedToken, ctx.UID)
}
