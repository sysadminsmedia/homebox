package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

type GroupService struct {
	repos *repo.AllRepos
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

func (svc *GroupService) NewInvitation(ctx Context, uses int, expiresAt time.Time) (string, error) {
	token := hasher.GenerateToken()

	_, err := svc.repos.Groups.InvitationCreate(ctx, ctx.GID, repo.GroupInvitationCreate{
		Token:     token.Hash,
		Uses:      uses,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}

	return token.Raw, nil
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
