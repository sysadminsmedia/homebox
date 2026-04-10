package services

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

type GroupService struct {
	repos *repo.AllRepos
}

// validateCanLeaveGroup Validate whether a user can leave the current group
// Returns the current user when there is no error
func (svc *GroupService) validateCanLeaveGroup(ctx Context) (repo.UserOut, error) {
	currentUser, err := svc.repos.Users.GetOneID(ctx, ctx.UID)
	if err != nil {
		return repo.UserOut{}, err
	}

	// Validate not leaving only group
	if len(currentUser.GroupIDs) <= 1 {
		return repo.UserOut{}, validate.NewRequestError(errors.New("cannot leave the only group you are a member of"), http.StatusBadRequest)
	}

	members, err := svc.repos.Users.GetUsersByGroupID(ctx, ctx.GID)
	if err != nil {
		return repo.UserOut{}, err
	}

	// Validate not last member
	if len(members) <= 1 {
		return repo.UserOut{}, validate.NewRequestError(errors.New(repo.ErrCannotLeaveLastMember.Error()), http.StatusBadRequest)
	}

	return currentUser, nil
}

func (svc *GroupService) getNewDefaultGroupID(currentUser repo.UserOut, leavingGroupID uuid.UUID) uuid.UUID {
	if currentUser.DefaultGroupID == leavingGroupID {
		newDefaultGroupID, _ := lo.Find(currentUser.GroupIDs, func(gid uuid.UUID) bool {
			return gid != leavingGroupID
		})
		return newDefaultGroupID
	}
	return uuid.Nil
}

func (svc *GroupService) UpdateGroup(ctx Context, data repo.GroupUpdate) (repo.Group, error) {
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

	return svc.repos.Groups.GroupCreate(ctx.Context, name, ctx.UID)
}

func (svc *GroupService) DeleteGroup(ctx Context) error {
	return svc.repos.Groups.GroupDelete(ctx.Context, ctx.GID)
}

func (svc *GroupService) NewInvitation(ctx Context, uses int, expiresAt time.Time) (repo.GroupInvitation, string, error) {
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

func (svc *GroupService) AddMember(ctx Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	return svc.repos.Groups.AddMember(ctx.Context, ctx.GID, userID)
}

func (svc *GroupService) RemoveMember(ctx Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	return svc.repos.Groups.RemoveMember(ctx.Context, ctx.GID, userID)
}

func (svc *GroupService) DeleteInvitation(ctx Context, id uuid.UUID) error {
	return svc.repos.Groups.InvitationDelete(ctx.Context, ctx.GID, id)
}

func (svc *GroupService) AcceptInvitation(ctx Context, token string) (repo.Group, error) {
	hashedToken := hasher.HashToken(token)
	return svc.repos.Groups.InvitationAccept(ctx.Context, hashedToken, ctx.UID)
}

func (svc *GroupService) LeaveGroup(ctx Context) error {
	currentUser, err := svc.validateCanLeaveGroup(ctx)
	if err != nil {
		return err
	}

	newDefaultGroupID := svc.getNewDefaultGroupID(currentUser, ctx.GID)
	err = svc.repos.Groups.GroupLeave(ctx, ctx.GID, ctx.UID, newDefaultGroupID)
	if err != nil {
		if errors.Is(err, repo.ErrCannotLeaveLastMember) {
			return validate.NewRequestError(errors.New(repo.ErrCannotLeaveLastMember.Error()), http.StatusBadRequest)
		}
		// Map any unexpected not-found (e.g. stale membership read) to a 400
		var entNotFound *ent.NotFoundError
		if errors.As(err, &entNotFound) {
			return validate.NewRequestError(errors.New("group membership not found; try again"), http.StatusBadRequest)
		}
		return err
	}
	return nil
}
