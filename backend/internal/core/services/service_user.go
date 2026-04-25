package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authroles"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
// with default Tags and Locations.
func (svc *UserService) RegisterUser(ctx context.Context, data UserRegistration) (repo.UserOut, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.RegisterUser",
		trace.WithAttributes(
			attribute.Int("user.name.length", len(data.Name)),
			attribute.Int("user.email.length", len(data.Email)),
			attribute.Int("user.password.length", len(data.Password)),
			attribute.Bool("registration.has_group_token", data.GroupToken != ""),
		))
	defer span.End()

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
		groupCtx, groupSpan := entityServiceTracer().Start(ctx, "service.UserService.RegisterUser.createGroup")
		log.Debug().Msg("creating new group")
		creatingGroup = true
		group, err = svc.repos.Groups.GroupCreate(groupCtx, fmt.Sprintf("%s's Home", data.Name), uuid.Nil)
		if err != nil {
			recordServiceSpanError(groupSpan, err)
			groupSpan.End()
			recordServiceSpanError(span, err)
			log.Err(err).Msg("Failed to create group")
			return repo.UserOut{}, err
		}
		groupSpan.SetAttributes(attribute.String("group.id", group.ID.String()))
		groupSpan.End()
	default:
		joinCtx, joinSpan := entityServiceTracer().Start(ctx, "service.UserService.RegisterUser.joinGroup")
		log.Debug().Msg("joining existing group")
		token, err = svc.repos.Groups.InvitationGet(joinCtx, hasher.HashToken(data.GroupToken))
		if err != nil {
			recordServiceSpanError(joinSpan, err)
			joinSpan.End()
			recordServiceSpanError(span, err)
			log.Err(err).Msg("Failed to get invitation token")
			return repo.UserOut{}, err
		}

		if token.ExpiresAt.Before(time.Now()) {
			err := errors.New("invitation expired")
			joinSpan.SetAttributes(attribute.String("invitation.error", "expired"))
			recordServiceSpanError(joinSpan, err)
			joinSpan.End()
			recordServiceSpanError(span, err)
			return repo.UserOut{}, err
		}
		if token.Uses <= 0 {
			err := errors.New("invitation used up")
			joinSpan.SetAttributes(attribute.String("invitation.error", "used_up"))
			recordServiceSpanError(joinSpan, err)
			joinSpan.End()
			recordServiceSpanError(span, err)
			return repo.UserOut{}, err
		}

		group = token.Group
		joinSpan.SetAttributes(
			attribute.String("group.id", group.ID.String()),
			attribute.Int("invitation.uses_remaining", token.Uses),
		)
		joinSpan.End()
	}

	span.SetAttributes(
		attribute.String("group.id", group.ID.String()),
		attribute.Bool("registration.creating_group", creatingGroup),
	)

	hashed, err := hasher.HashPasswordCtx(ctx, data.Password)
	if err != nil {
		recordServiceSpanError(span, err)
		log.Err(err).Msg("Failed to hash password")
		return repo.UserOut{}, err
	}
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
		recordServiceSpanError(span, err)
		return repo.UserOut{}, err
	}
	span.SetAttributes(attribute.String("user.id", usr.ID.String()))
	log.Debug().Msg("user created")

	// Create the default tags and locations for the group.
	if creatingGroup {
		bootstrapCtx, bootstrapSpan := entityServiceTracer().Start(ctx, "service.UserService.RegisterUser.bootstrap")
		log.Debug().Msg("creating default tags")
		tagsCreated := 0
		for _, tag := range defaultTags() {
			_, err := svc.repos.Tags.Create(bootstrapCtx, usr.DefaultGroupID, tag)
			if err != nil {
				recordServiceSpanError(bootstrapSpan, err)
				bootstrapSpan.End()
				recordServiceSpanError(span, err)
				return repo.UserOut{}, err
			}
			tagsCreated++
		}

		log.Debug().Msg("creating default locations")
		locsCreated := 0
		for _, loc := range defaultLocations() {
			_, err := svc.repos.Entities.CreateContainer(bootstrapCtx, usr.DefaultGroupID, loc)
			if err != nil {
				recordServiceSpanError(bootstrapSpan, err)
				bootstrapSpan.End()
				recordServiceSpanError(span, err)
				return repo.UserOut{}, err
			}
			locsCreated++
		}
		bootstrapSpan.SetAttributes(
			attribute.Int("tags.created.count", tagsCreated),
			attribute.Int("locations.created.count", locsCreated),
		)
		bootstrapSpan.End()
	}

	// Decrement the invitation token if it was used.
	if token.ID != uuid.Nil {
		decCtx, decSpan := entityServiceTracer().Start(ctx, "service.UserService.RegisterUser.decrementInvitation")
		log.Debug().Msg("decrementing invitation token")
		err = svc.repos.Groups.InvitationDecrement(decCtx, token.ID)
		if err != nil {
			recordServiceSpanError(decSpan, err)
			decSpan.End()
			recordServiceSpanError(span, err)
			log.Err(err).Msg("Failed to update invitation token")
			return repo.UserOut{}, err
		}
		decSpan.End()
	}

	return usr, nil
}

// GetSelf returns the user that is currently logged in based of the token provided within
func (svc *UserService) GetSelf(ctx context.Context, requestToken string) (repo.UserOut, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.GetSelf",
		trace.WithAttributes(attribute.Int("token.length", len(requestToken))))
	defer span.End()

	hash := hasher.HashToken(requestToken)
	out, err := svc.repos.AuthTokens.GetUserFromToken(ctx, hash)
	if err != nil {
		span.SetAttributes(attribute.Bool("user.found", false))
		if !ent.IsNotFound(err) {
			recordServiceSpanError(span, err)
		}
		return out, err
	}
	span.SetAttributes(
		attribute.Bool("user.found", true),
		attribute.String("user.id", out.ID.String()),
	)
	return out, nil
}

func (svc *UserService) UpdateSelf(ctx context.Context, id uuid.UUID, data repo.UserUpdate) (repo.UserOut, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.UpdateSelf",
		trace.WithAttributes(attribute.String("user.id", id.String())))
	defer span.End()

	err := svc.repos.Users.Update(ctx, id, data)
	if err != nil {
		recordServiceSpanError(span, err)
		return repo.UserOut{}, err
	}

	out, err := svc.repos.Users.GetOneID(ctx, id)
	if err != nil {
		recordServiceSpanError(span, err)
	}
	return out, err
}

// ============================================================================
// User Authentication

func (svc *UserService) createSessionToken(ctx context.Context, userID uuid.UUID, extendedSession bool) (UserAuthTokenDetail, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.createSessionToken",
		trace.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.Bool("session.extended", extendedSession),
		))
	defer span.End()

	attachmentToken := hasher.GenerateTokenCtx(ctx)

	expiresAt := time.Now().Add(oneWeek)
	if extendedSession {
		expiresAt = time.Now().Add(oneWeek * 4)
	}
	span.SetAttributes(attribute.String("session.expires_at", expiresAt.Format(time.RFC3339)))

	attachmentData := repo.UserAuthTokenCreate{
		UserID:    userID,
		TokenHash: attachmentToken.Hash,
		ExpiresAt: expiresAt,
	}

	attCtx, attSpan := entityServiceTracer().Start(ctx, "service.UserService.createSessionToken.attachmentToken")
	_, err := svc.repos.AuthTokens.CreateToken(attCtx, attachmentData, authroles.RoleAttachments)
	if err != nil {
		recordServiceSpanError(attSpan, err)
		attSpan.End()
		recordServiceSpanError(span, err)
		return UserAuthTokenDetail{}, err
	}
	attSpan.End()

	userToken := hasher.GenerateTokenCtx(ctx)
	data := repo.UserAuthTokenCreate{
		UserID:    userID,
		TokenHash: userToken.Hash,
		ExpiresAt: expiresAt,
	}

	userCtx, userSpan := entityServiceTracer().Start(ctx, "service.UserService.createSessionToken.userToken")
	created, err := svc.repos.AuthTokens.CreateToken(userCtx, data, authroles.RoleUser)
	if err != nil {
		recordServiceSpanError(userSpan, err)
		userSpan.End()
		recordServiceSpanError(span, err)
		return UserAuthTokenDetail{}, err
	}
	userSpan.End()

	return UserAuthTokenDetail{
		Raw:             userToken.Raw,
		ExpiresAt:       created.ExpiresAt,
		AttachmentToken: attachmentToken.Raw,
	}, nil
}

// Login is the main local-credential login path. The span and its sub-spans capture
// every branch (user-not-found, OIDC-only user, password mismatch, password rehash)
// so an intermittent password rejection trace points directly at the failing step.
func (svc *UserService) Login(ctx context.Context, username, password string, extendedSession bool) (UserAuthTokenDetail, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.Login",
		trace.WithAttributes(
			attribute.Int("user.email.length", len(username)),
			attribute.Int("password.length", len(password)),
			attribute.Bool("session.extended", extendedSession),
		))
	defer span.End()

	usr, err := svc.repos.Users.GetOneEmail(ctx, username)
	if err != nil {
		span.SetAttributes(
			attribute.Bool("user.found", false),
			attribute.String("login.outcome", "user_not_found"),
		)
		// SECURITY: Perform hash to ensure response times are the same
		_, dummySpan := entityServiceTracer().Start(ctx, "service.UserService.Login.timingDummy",
			trace.WithAttributes(attribute.String("reason", "user_not_found")))
		hasher.CheckPasswordHashCtx(ctx, "not-a-real-password", "not-a-real-password")
		dummySpan.End()
		return UserAuthTokenDetail{}, ErrorInvalidLogin
	}
	span.SetAttributes(
		attribute.Bool("user.found", true),
		attribute.String("user.id", usr.ID.String()),
		attribute.Bool("user.has_password_hash", usr.PasswordHash != ""),
		attribute.Bool("user.has_oidc", usr.OidcIssuer != nil && usr.OidcSubject != nil),
	)

	// SECURITY: Deny login for users with null or empty password (OIDC users)
	if usr.PasswordHash == "" {
		log.Warn().Str("email", username).Msg("Login attempt blocked for user with null password (likely OIDC user)")
		span.SetAttributes(attribute.String("login.outcome", "blocked_no_password_hash"))
		_, dummySpan := entityServiceTracer().Start(ctx, "service.UserService.Login.timingDummy",
			trace.WithAttributes(attribute.String("reason", "no_password_hash")))
		hasher.CheckPasswordHashCtx(ctx, "not-a-real-password", "not-a-real-password")
		dummySpan.End()
		return UserAuthTokenDetail{}, ErrorInvalidLogin
	}

	check, rehash := hasher.CheckPasswordHashCtx(ctx, password, usr.PasswordHash)
	span.SetAttributes(
		attribute.Bool("password.match", check),
		attribute.Bool("password.rehash_needed", rehash),
	)

	if !check {
		span.SetAttributes(attribute.String("login.outcome", "password_mismatch"))
		return UserAuthTokenDetail{}, ErrorInvalidLogin
	}

	if rehash {
		rehashCtx, rehashSpan := entityServiceTracer().Start(ctx, "service.UserService.Login.rehash",
			trace.WithAttributes(attribute.String("user.id", usr.ID.String())))
		hash, err := hasher.HashPasswordCtx(rehashCtx, password)
		if err != nil {
			recordServiceSpanError(rehashSpan, err)
			rehashSpan.End()
			recordServiceSpanError(span, err)
			log.Err(err).Msg("Failed to hash password")
			return UserAuthTokenDetail{}, err
		}
		err = svc.repos.Users.ChangePassword(rehashCtx, usr.ID, hash)
		if err != nil {
			recordServiceSpanError(rehashSpan, err)
			rehashSpan.End()
			recordServiceSpanError(span, err)
			return UserAuthTokenDetail{}, err
		}
		rehashSpan.End()
	}

	span.SetAttributes(attribute.String("login.outcome", "success"))
	out, err := svc.createSessionToken(ctx, usr.ID, extendedSession)
	if err != nil {
		recordServiceSpanError(span, err)
	}
	return out, err
}

// LoginOIDC creates a session token for a user authenticated via OIDC.
// It now uses issuer + subject for identity association (OIDC spec compliance).
// If the user doesn't exist, it will create one.
func (svc *UserService) LoginOIDC(ctx context.Context, issuer, subject, email, name string) (UserAuthTokenDetail, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.LoginOIDC",
		trace.WithAttributes(
			attribute.String("oidc.issuer", issuer),
			attribute.Int("oidc.subject.length", len(subject)),
			attribute.Int("oidc.email.length", len(email)),
			attribute.Int("oidc.name.length", len(name)),
		))
	defer span.End()

	issuer = strings.TrimSpace(issuer)
	subject = strings.TrimSpace(subject)
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	if issuer == "" || subject == "" {
		log.Warn().Str("issuer", issuer).Str("subject", subject).Msg("OIDC login missing issuer or subject")
		span.SetAttributes(attribute.String("oidc.outcome", "missing_issuer_or_subject"))
		return UserAuthTokenDetail{}, ErrorInvalidLogin
	}

	// Try to get existing user by OIDC identity
	usr, err := svc.repos.Users.GetOneOIDC(ctx, issuer, subject)
	if err != nil {
		if !ent.IsNotFound(err) {
			recordServiceSpanError(span, err)
			log.Err(err).Str("issuer", issuer).Str("subject", subject).Msg("failed to lookup user by OIDC identity")
			return UserAuthTokenDetail{}, err
		}
		// Not found: attempt migration path by email (legacy) if email provided
		if email != "" {
			migrationCtx, migrationSpan := entityServiceTracer().Start(ctx, "service.UserService.LoginOIDC.legacyEmailMigration")
			legacyUsr, lerr := svc.repos.Users.GetOneEmail(migrationCtx, email)
			if lerr == nil {
				log.Info().Str("email", email).Str("issuer", issuer).Str("subject", subject).Msg("migrating legacy email-based OIDC user to issuer+subject")
				if uerr := svc.repos.Users.SetOIDCIdentity(migrationCtx, legacyUsr.ID, issuer, subject); uerr == nil {
					usr = legacyUsr
					migrationSpan.SetAttributes(
						attribute.String("oidc.migration.outcome", "migrated"),
						attribute.String("user.id", legacyUsr.ID.String()),
					)
				} else {
					migrationSpan.SetAttributes(attribute.String("oidc.migration.outcome", "set_identity_failed"))
					recordServiceSpanError(migrationSpan, uerr)
					log.Err(uerr).Str("email", email).Msg("failed to set OIDC identity on legacy user")
				}
			} else {
				migrationSpan.SetAttributes(attribute.String("oidc.migration.outcome", "no_legacy_user"))
			}
			migrationSpan.End()
		}
	}

	// Create user if still not resolved
	if usr.ID == uuid.Nil {
		log.Debug().Str("issuer", issuer).Str("subject", subject).Msg("OIDC user not found, creating new user")
		span.SetAttributes(attribute.String("oidc.outcome", "creating_user"))
		usr, err = svc.registerOIDCUser(ctx, issuer, subject, email, name)
		if err != nil {
			if ent.IsConstraintError(err) {
				if usr2, gerr := svc.repos.Users.GetOneOIDC(ctx, issuer, subject); gerr == nil {
					log.Info().Str("issuer", issuer).Str("subject", subject).Msg("OIDC user created concurrently; proceeding")
					usr = usr2
					span.SetAttributes(attribute.String("oidc.outcome", "concurrent_create_resolved"))
				} else {
					recordServiceSpanError(span, gerr)
					log.Err(gerr).Str("issuer", issuer).Str("subject", subject).Msg("failed to fetch user after constraint error")
					return UserAuthTokenDetail{}, gerr
				}
			} else {
				recordServiceSpanError(span, err)
				log.Err(err).Str("issuer", issuer).Str("subject", subject).Msg("failed to create OIDC user")
				return UserAuthTokenDetail{}, err
			}
		}
	} else {
		span.SetAttributes(attribute.String("oidc.outcome", "existing_user"))
	}

	span.SetAttributes(attribute.String("user.id", usr.ID.String()))
	out, err := svc.createSessionToken(ctx, usr.ID, true)
	if err != nil {
		recordServiceSpanError(span, err)
	}
	return out, err
}

// registerOIDCUser creates a new user for OIDC authentication with issuer+subject identity.
func (svc *UserService) registerOIDCUser(ctx context.Context, issuer, subject, email, name string) (repo.UserOut, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.registerOIDCUser",
		trace.WithAttributes(
			attribute.String("oidc.issuer", issuer),
			attribute.Int("oidc.subject.length", len(subject)),
		))
	defer span.End()

	group, err := svc.repos.Groups.GroupCreate(ctx, "Home", uuid.Nil)
	if err != nil {
		recordServiceSpanError(span, err)
		log.Err(err).Msg("Failed to create group for OIDC user")
		return repo.UserOut{}, err
	}
	span.SetAttributes(attribute.String("group.id", group.ID.String()))

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
		recordServiceSpanError(span, err)
		return repo.UserOut{}, err
	}
	span.SetAttributes(attribute.String("user.id", entUser.ID.String()))

	bootstrapCtx, bootstrapSpan := entityServiceTracer().Start(ctx, "service.UserService.registerOIDCUser.bootstrap")
	log.Debug().Str("issuer", issuer).Str("subject", subject).Msg("creating default tags for OIDC user")
	tagsCreated := 0
	for _, tag := range defaultTags() {
		_, err := svc.repos.Tags.Create(bootstrapCtx, group.ID, tag)
		if err != nil {
			recordServiceSpanError(bootstrapSpan, err)
			log.Err(err).Msg("Failed to create default tag")
			continue
		}
		tagsCreated++
	}

	log.Debug().Str("issuer", issuer).Str("subject", subject).Msg("creating default locations for OIDC user")
	locsCreated := 0
	for _, loc := range defaultLocations() {
		_, err := svc.repos.Entities.CreateContainer(bootstrapCtx, group.ID, loc)
		if err != nil {
			recordServiceSpanError(bootstrapSpan, err)
			log.Err(err).Msg("Failed to create default location")
			continue
		}
		locsCreated++
	}
	bootstrapSpan.SetAttributes(
		attribute.Int("tags.created.count", tagsCreated),
		attribute.Int("locations.created.count", locsCreated),
	)
	bootstrapSpan.End()

	return entUser, nil
}

func (svc *UserService) Logout(ctx context.Context, token string) error {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.Logout",
		trace.WithAttributes(attribute.Int("token.length", len(token))))
	defer span.End()

	hash := hasher.HashToken(token)
	err := svc.repos.AuthTokens.DeleteToken(ctx, hash)
	recordServiceSpanError(span, err)
	return err
}

func (svc *UserService) RenewToken(ctx context.Context, token string) (UserAuthTokenDetail, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.RenewToken",
		trace.WithAttributes(attribute.Int("token.length", len(token))))
	defer span.End()

	hash := hasher.HashToken(token)

	dbToken, err := svc.repos.AuthTokens.GetUserFromToken(ctx, hash)
	if err != nil {
		span.SetAttributes(
			attribute.Bool("user.found", false),
			attribute.String("renew.outcome", "invalid_token"),
		)
		if !ent.IsNotFound(err) {
			recordServiceSpanError(span, err)
		}
		return UserAuthTokenDetail{}, ErrorInvalidToken
	}
	span.SetAttributes(
		attribute.Bool("user.found", true),
		attribute.String("user.id", dbToken.ID.String()),
	)

	out, err := svc.createSessionToken(ctx, dbToken.ID, false)
	if err != nil {
		recordServiceSpanError(span, err)
	}
	return out, err
}

// DeleteSelf deletes the user that is currently logged based of the provided UUID
// There is _NO_ protection against deleting the wrong user, as such this should only
// be used when the identify of the user has been confirmed.
func (svc *UserService) DeleteSelf(ctx context.Context, id uuid.UUID) error {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.DeleteSelf",
		trace.WithAttributes(attribute.String("user.id", id.String())))
	defer span.End()

	err := svc.repos.Users.Delete(ctx, id)
	recordServiceSpanError(span, err)
	return err
}

func (svc *UserService) ChangePassword(ctx Context, current string, new string) (ok bool) {
	spanCtx, span := entityServiceTracer().Start(ctx.Context, "service.UserService.ChangePassword",
		trace.WithAttributes(
			attribute.String("user.id", ctx.UID.String()),
			attribute.Int("password.current.length", len(current)),
			attribute.Int("password.new.length", len(new)),
		))
	defer span.End()
	ctx.Context = spanCtx

	usr, err := svc.repos.Users.GetOneID(ctx, ctx.UID)
	if err != nil {
		recordServiceSpanError(span, err)
		span.SetAttributes(attribute.String("change_password.outcome", "user_not_found"))
		return false
	}

	match, _ := hasher.CheckPasswordHashCtx(spanCtx, current, usr.PasswordHash)
	span.SetAttributes(attribute.Bool("password.current.match", match))
	if !match {
		span.SetAttributes(attribute.String("change_password.outcome", "current_password_incorrect"))
		log.Err(errors.New("current password is incorrect")).Msg("Failed to change password")
		return false
	}

	hashed, err := hasher.HashPasswordCtx(spanCtx, new)
	if err != nil {
		recordServiceSpanError(span, err)
		span.SetAttributes(attribute.String("change_password.outcome", "hash_failed"))
		log.Err(err).Msg("Failed to hash password")
		return false
	}

	err = svc.repos.Users.ChangePassword(ctx.Context, ctx.UID, hashed)
	if err != nil {
		recordServiceSpanError(span, err)
		span.SetAttributes(attribute.String("change_password.outcome", "persist_failed"))
		log.Err(err).Msg("Failed to change password")
		return false
	}

	span.SetAttributes(attribute.String("change_password.outcome", "success"))
	return true
}

func (svc *UserService) GetSettings(ctx context.Context, uid uuid.UUID) (map[string]interface{}, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.GetSettings",
		trace.WithAttributes(attribute.String("user.id", uid.String())))
	defer span.End()

	out, err := svc.repos.Users.GetSettings(ctx, uid)
	if err != nil {
		recordServiceSpanError(span, err)
		return out, err
	}
	span.SetAttributes(attribute.Int("settings.keys.count", len(out)))
	return out, nil
}

func (svc *UserService) SetSettings(ctx context.Context, uid uuid.UUID, settings map[string]interface{}) error {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.SetSettings",
		trace.WithAttributes(
			attribute.String("user.id", uid.String()),
			attribute.Int("settings.keys.count", len(settings)),
		))
	defer span.End()

	err := svc.repos.Users.SetSettings(ctx, uid, settings)
	recordServiceSpanError(span, err)
	return err
}

// EnsureUserPassword ensures that the user with the given email has the specified password. If the password does not match, it updates the user's password to the new value.
// WARNING: This method bypasses normal checks, it should only be used for demos and/or superuser level administrative processes.
func (svc *UserService) EnsureUserPassword(ctx context.Context, email, password string) error {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.EnsureUserPassword",
		trace.WithAttributes(
			attribute.Int("user.email.length", len(email)),
			attribute.Int("password.length", len(password)),
		))
	defer span.End()

	usr, err := svc.repos.Users.GetOneEmailNoEdges(ctx, email)
	if err != nil {
		if !ent.IsNotFound(err) {
			recordServiceSpanError(span, err)
		}
		span.SetAttributes(attribute.String("ensure.outcome", "user_not_found"))
		return err
	}
	span.SetAttributes(attribute.String("user.id", usr.ID.String()))

	match := false
	if usr.PasswordHash != "" {
		match, _ = hasher.CheckPasswordHashCtx(ctx, password, usr.PasswordHash)
	}
	span.SetAttributes(attribute.Bool("password.match", match))
	if !match {
		hash, herr := hasher.HashPasswordCtx(ctx, password)
		if herr != nil {
			recordServiceSpanError(span, herr)
			span.SetAttributes(attribute.String("ensure.outcome", "hash_failed"))
			return herr
		}
		if cerr := svc.repos.Users.ChangePassword(ctx, usr.ID, hash); cerr != nil {
			recordServiceSpanError(span, cerr)
			span.SetAttributes(attribute.String("ensure.outcome", "persist_failed"))
			return cerr
		}
		span.SetAttributes(attribute.String("ensure.outcome", "rehashed"))
		return nil
	}
	span.SetAttributes(attribute.String("ensure.outcome", "noop_match"))
	return nil
}

// ExistsByEmail returns true if a user with the given email exists.
func (svc *UserService) ExistsByEmail(ctx context.Context, email string) bool {
	ctx, span := entityServiceTracer().Start(ctx, "service.UserService.ExistsByEmail",
		trace.WithAttributes(attribute.Int("user.email.length", len(email))))
	defer span.End()

	_, err := svc.repos.Users.GetOneEmailNoEdges(ctx, email)
	exists := err == nil
	span.SetAttributes(attribute.Bool("user.exists", exists))
	return exists
}
