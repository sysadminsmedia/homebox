package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authroles"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

var (
	oneWeek              = time.Hour * 24 * 7
	ErrorInvalidLogin    = errors.New("invalid username or password")
	ErrorInvalidToken    = errors.New("invalid token")
	ErrorTokenIDMismatch = errors.New("token id mismatch")
)

type UserService struct {
	repos *repo.AllRepos
}

type (
	UserRegistration struct {
		GroupToken string `json:"token"`
		Name       string `json:"name"`
		Email      string `json:"email"`
		Password   string `json:"password"`
	}
	UserAuthTokenDetail struct {
		Raw             string    `json:"raw"`
		AttachmentToken string    `json:"attachmentToken"`
		ExpiresAt       time.Time `json:"expiresAt"`
	}
	LoginForm struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
)

// RegisterUser creates a new user and group in the data with the provided data. It also bootstraps the user's group
// with default Labels and Locations.
func (svc *UserService) RegisterUser(ctx context.Context, data UserRegistration) (repo.UserOut, error) {
	log.Debug().
		Str("name", data.Name).
		Str("email", data.Email).
		Str("groupToken", data.GroupToken).
		Msg("Registering new user")

	var (
		err   error
		group repo.Group
		token repo.GroupInvitation

		// creatingGroup is true if the user is creating a new group.
		creatingGroup = false
	)

	switch data.GroupToken {
	case "":
		log.Debug().Msg("creating new group")
		creatingGroup = true
		group, err = svc.repos.Groups.GroupCreate(ctx, "Home")
		if err != nil {
			log.Err(err).Msg("Failed to create group")
			return repo.UserOut{}, err
		}
	default:
		log.Debug().Msg("joining existing group")
		token, err = svc.repos.Groups.InvitationGet(ctx, hasher.HashToken(data.GroupToken))
		if err != nil {
			log.Err(err).Msg("Failed to get invitation token")
			return repo.UserOut{}, err
		}
		group = token.Group
	}

	hashed, _ := hasher.HashPassword(data.Password)
	usrCreate := repo.UserCreate{
		Name:        data.Name,
		Email:       data.Email,
		Password:    hashed,
		IsSuperuser: false,
		GroupID:     group.ID,
		IsOwner:     creatingGroup,
	}

	usr, err := svc.repos.Users.Create(ctx, usrCreate)
	if err != nil {
		return repo.UserOut{}, err
	}
	log.Debug().Msg("user created")

	// Create the default labels and locations for the group.
	if creatingGroup {
		log.Debug().Msg("creating default labels")
		for _, label := range defaultLabels() {
			_, err := svc.repos.Labels.Create(ctx, usr.GroupID, label)
			if err != nil {
				return repo.UserOut{}, err
			}
		}

		log.Debug().Msg("creating default locations")
		for _, location := range defaultLocations() {
			_, err := svc.repos.Locations.Create(ctx, usr.GroupID, location)
			if err != nil {
				return repo.UserOut{}, err
			}
		}
	}

	// Decrement the invitation token if it was used.
	if token.ID != uuid.Nil {
		log.Debug().Msg("decrementing invitation token")
		err = svc.repos.Groups.InvitationUpdate(ctx, token.ID, token.Uses-1)
		if err != nil {
			log.Err(err).Msg("Failed to update invitation token")
			return repo.UserOut{}, err
		}
	}

	return usr, nil
}

// GetSelf returns the user that is currently logged in based of the token provided within
func (svc *UserService) GetSelf(ctx context.Context, requestToken string) (repo.UserOut, error) {
	hash := hasher.HashToken(requestToken)
	return svc.repos.AuthTokens.GetUserFromToken(ctx, hash)
}

func (svc *UserService) UpdateSelf(ctx context.Context, id uuid.UUID, data repo.UserUpdate) (repo.UserOut, error) {
	err := svc.repos.Users.Update(ctx, id, data)
	if err != nil {
		return repo.UserOut{}, err
	}

	return svc.repos.Users.GetOneID(ctx, id)
}

// ============================================================================
// User Authentication

func (svc *UserService) createSessionToken(ctx context.Context, userID uuid.UUID, extendedSession bool) (UserAuthTokenDetail, error) {
	attachmentToken := hasher.GenerateToken()

	expiresAt := time.Now().Add(oneWeek)
	if extendedSession {
		expiresAt = time.Now().Add(oneWeek * 4)
	}

	attachmentData := repo.UserAuthTokenCreate{
		UserID:    userID,
		TokenHash: attachmentToken.Hash,
		ExpiresAt: expiresAt,
	}

	_, err := svc.repos.AuthTokens.CreateToken(ctx, attachmentData, authroles.RoleAttachments)
	if err != nil {
		return UserAuthTokenDetail{}, err
	}

	userToken := hasher.GenerateToken()
	data := repo.UserAuthTokenCreate{
		UserID:    userID,
		TokenHash: userToken.Hash,
		ExpiresAt: expiresAt,
	}

	created, err := svc.repos.AuthTokens.CreateToken(ctx, data, authroles.RoleUser)
	if err != nil {
		return UserAuthTokenDetail{}, err
	}

	return UserAuthTokenDetail{
		Raw:             userToken.Raw,
		ExpiresAt:       created.ExpiresAt,
		AttachmentToken: attachmentToken.Raw,
	}, nil
}

func (svc *UserService) Login(ctx context.Context, username, password string, extendedSession bool) (UserAuthTokenDetail, error) {
	usr, err := svc.repos.Users.GetOneEmail(ctx, username)
	if err != nil {
		// SECURITY: Perform hash to ensure response times are the same
		hasher.CheckPasswordHash("not-a-real-password", "not-a-real-password")
		return UserAuthTokenDetail{}, ErrorInvalidLogin
	}

	check, rehash := hasher.CheckPasswordHash(password, usr.PasswordHash)

	if !check {
		return UserAuthTokenDetail{}, ErrorInvalidLogin
	}

	if rehash {
		hash, err := hasher.HashPassword(password)
		if err != nil {
			log.Err(err).Msg("Failed to hash password")
			return UserAuthTokenDetail{}, err
		}
		err = svc.repos.Users.ChangePassword(ctx, usr.ID, hash)
		if err != nil {
			return UserAuthTokenDetail{}, err
		}
	}
	return svc.createSessionToken(ctx, usr.ID, extendedSession)
}

func (svc *UserService) Logout(ctx context.Context, token string) error {
	hash := hasher.HashToken(token)
	err := svc.repos.AuthTokens.DeleteToken(ctx, hash)
	return err
}

func (svc *UserService) RenewToken(ctx context.Context, token string) (UserAuthTokenDetail, error) {
	hash := hasher.HashToken(token)

	dbToken, err := svc.repos.AuthTokens.GetUserFromToken(ctx, hash)
	if err != nil {
		return UserAuthTokenDetail{}, ErrorInvalidToken
	}

	return svc.createSessionToken(ctx, dbToken.ID, false)
}

// DeleteSelf deletes the user that is currently logged based of the provided UUID
// There is _NO_ protection against deleting the wrong user, as such this should only
// be used when the identify of the user has been confirmed.
func (svc *UserService) DeleteSelf(ctx context.Context, id uuid.UUID) error {
	return svc.repos.Users.Delete(ctx, id)
}

func (svc *UserService) ChangePassword(ctx Context, current string, new string) (ok bool) {
	usr, err := svc.repos.Users.GetOneID(ctx, ctx.UID)
	if err != nil {
		return false
	}

	match, _ := hasher.CheckPasswordHash(current, usr.PasswordHash)
	if !match {
		log.Err(errors.New("current password is incorrect")).Msg("Failed to change password")
		return false
	}

	hashed, err := hasher.HashPassword(new)
	if err != nil {
		log.Err(err).Msg("Failed to hash password")
		return false
	}

	err = svc.repos.Users.ChangePassword(ctx.Context, ctx.UID, hashed)
	if err != nil {
		log.Err(err).Msg("Failed to change password")
		return false
	}

	return true
}

func (svc *UserService) GetSettings(ctx context.Context, uid uuid.UUID) (schema.UserSettings, error) {
	return svc.repos.Users.GetSettings(ctx, uid)
}

func (svc *UserService) SetSettings(ctx context.Context, uid uuid.UUID, settings schema.UserSettings) error {
	return svc.repos.Users.SetSettings(ctx, uid, settings)
}
