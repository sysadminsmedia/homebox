package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authroles"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

var (
	oneWeek           = time.Hour * 24 * 7
	ErrorInvalidLogin = errors.New("invalid username or password")
	ErrorInvalidToken = errors.New("invalid token")
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
		group, err = svc.repos.Groups.GroupCreate(ctx, "Home", uuid.Nil)
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
		Name:           data.Name,
		Email:          data.Email,
		Password:       &hashed,
		IsSuperuser:    false,
		DefaultGroupID: group.ID,
		IsOwner:        creatingGroup,
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
			_, err := svc.repos.Labels.Create(ctx, usr.DefaultGroupID, label)
			if err != nil {
				return repo.UserOut{}, err
			}
		}

		log.Debug().Msg("creating default locations")
		for _, location := range defaultLocations() {
			_, err := svc.repos.Locations.Create(ctx, usr.DefaultGroupID, location)
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

	// SECURITY: Deny login for users with null or empty password (OIDC users)
	if usr.PasswordHash == "" {
		log.Warn().Str("email", username).Msg("Login attempt blocked for user with null password (likely OIDC user)")
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

// LoginOIDC creates a session token for a user authenticated via OIDC.
// It now uses issuer + subject for identity association (OIDC spec compliance).
// If the user doesn't exist, it will create one.
func (svc *UserService) LoginOIDC(ctx context.Context, issuer, subject, email, name string) (UserAuthTokenDetail, error) {
	issuer = strings.TrimSpace(issuer)
	subject = strings.TrimSpace(subject)
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	if issuer == "" || subject == "" {
		log.Warn().Str("issuer", issuer).Str("subject", subject).Msg("OIDC login missing issuer or subject")
		return UserAuthTokenDetail{}, ErrorInvalidLogin
	}

	// Try to get existing user by OIDC identity
	usr, err := svc.repos.Users.GetOneOIDC(ctx, issuer, subject)
	if err != nil {
		if !ent.IsNotFound(err) {
			log.Err(err).Str("issuer", issuer).Str("subject", subject).Msg("failed to lookup user by OIDC identity")
			return UserAuthTokenDetail{}, err
		}
		// Not found: attempt migration path by email (legacy) if email provided
		if email != "" {
			legacyUsr, lerr := svc.repos.Users.GetOneEmail(ctx, email)
			if lerr == nil {
				log.Info().Str("email", email).Str("issuer", issuer).Str("subject", subject).Msg("migrating legacy email-based OIDC user to issuer+subject")
				// Update user with OIDC identity fields
				if uerr := svc.repos.Users.SetOIDCIdentity(ctx, legacyUsr.ID, issuer, subject); uerr == nil {
					usr = legacyUsr
				} else {
					log.Err(uerr).Str("email", email).Msg("failed to set OIDC identity on legacy user")
				}
			}
		}
	}

	// Create user if still not resolved
	if usr.ID == uuid.Nil {
		log.Debug().Str("issuer", issuer).Str("subject", subject).Msg("OIDC user not found, creating new user")
		usr, err = svc.registerOIDCUser(ctx, issuer, subject, email, name)
		if err != nil {
			if ent.IsConstraintError(err) {
				if usr2, gerr := svc.repos.Users.GetOneOIDC(ctx, issuer, subject); gerr == nil {
					log.Info().Str("issuer", issuer).Str("subject", subject).Msg("OIDC user created concurrently; proceeding")
					usr = usr2
				} else {
					log.Err(gerr).Str("issuer", issuer).Str("subject", subject).Msg("failed to fetch user after constraint error")
					return UserAuthTokenDetail{}, gerr
				}
			} else {
				log.Err(err).Str("issuer", issuer).Str("subject", subject).Msg("failed to create OIDC user")
				return UserAuthTokenDetail{}, err
			}
		}
	}

	return svc.createSessionToken(ctx, usr.ID, true)
}

// registerOIDCUser creates a new user for OIDC authentication with issuer+subject identity.
func (svc *UserService) registerOIDCUser(ctx context.Context, issuer, subject, email, name string) (repo.UserOut, error) {
	group, err := svc.repos.Groups.GroupCreate(ctx, "Home", uuid.Nil)
	if err != nil {
		log.Err(err).Msg("Failed to create group for OIDC user")
		return repo.UserOut{}, err
	}

	usrCreate := repo.UserCreate{
		Name:           name,
		Email:          email,
		Password:       nil,
		IsSuperuser:    false,
		DefaultGroupID: group.ID,
		IsOwner:        true,
	}

	entUser, err := svc.repos.Users.CreateWithOIDC(ctx, usrCreate, issuer, subject)
	if err != nil {
		return repo.UserOut{}, err
	}

	log.Debug().Str("issuer", issuer).Str("subject", subject).Msg("creating default labels for OIDC user")
	for _, label := range defaultLabels() {
		_, err := svc.repos.Labels.Create(ctx, group.ID, label)
		if err != nil {
			log.Err(err).Msg("Failed to create default label")
		}
	}

	log.Debug().Str("issuer", issuer).Str("subject", subject).Msg("creating default locations for OIDC user")
	for _, location := range defaultLocations() {
		_, err := svc.repos.Locations.Create(ctx, group.ID, location)
		if err != nil {
			log.Err(err).Msg("Failed to create default location")
		}
	}

	return entUser, nil
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

func (svc *UserService) EnsureUserPassword(ctx context.Context, email, password string) error {
	usr, err := svc.repos.Users.GetOneEmailNoEdges(ctx, email)
	if err != nil {
		return err
	}
	match := false
	if usr.PasswordHash != "" {
		match, _ = hasher.CheckPasswordHash(password, usr.PasswordHash)
	}
	if !match {
		hash, herr := hasher.HashPassword(password)
		if herr != nil {
			return herr
		}
		if cerr := svc.repos.Users.ChangePassword(ctx, usr.ID, hash); cerr != nil {
			return cerr
		}
	}
	return nil
}

// ExistsByEmail returns true if a user with the given email exists.
func (svc *UserService) ExistsByEmail(ctx context.Context, email string) bool {
	_, err := svc.repos.Users.GetOneEmailNoEdges(ctx, email)
	return err == nil
}
