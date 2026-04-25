package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authroles"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authtokens"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
	"github.com/sysadminsmedia/homebox/backend/pkgs/set"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type TokenRepository struct {
	db *ent.Client
}

type (
	UserAuthTokenCreate struct {
		TokenHash []byte    `json:"token"`
		UserID    uuid.UUID `json:"userId"`
		ExpiresAt time.Time `json:"expiresAt"`
	}

	UserAuthToken struct {
		UserAuthTokenCreate
		CreatedAt time.Time `json:"createdAt"`
	}
)

func (u UserAuthToken) IsExpired() bool {
	return u.ExpiresAt.Before(time.Now())
}

// GetUserFromToken get's a user from a token
func (r *TokenRepository) GetUserFromToken(ctx context.Context, token []byte) (UserOut, error) {
	ctx, span := entityTracer().Start(ctx, "repo.TokenRepository.GetUserFromToken",
		trace.WithAttributes(attribute.Int("token.hash.length", len(token))))
	defer span.End()

	user, err := r.db.AuthTokens.Query().
		Where(authtokens.Token(token)).
		Where(authtokens.ExpiresAtGTE(time.Now())).
		WithUser().
		QueryUser().
		WithGroups().
		Only(ctx)
	if err != nil {
		span.SetAttributes(
			attribute.Bool("user.found", false),
			attribute.Bool("token.lookup.not_found", ent.IsNotFound(err)),
		)
		if !ent.IsNotFound(err) {
			recordSpanError(span, err)
		}
		return UserOut{}, err
	}

	out := mapUserOut(user)
	span.SetAttributes(attribute.Bool("user.found", true))
	span.SetAttributes(userSpanAttrs(out)...)
	return out, nil
}

func (r *TokenRepository) GetRoles(ctx context.Context, token string) (*set.Set[string], error) {
	ctx, span := entityTracer().Start(ctx, "repo.TokenRepository.GetRoles",
		trace.WithAttributes(attribute.Int("token.length", len(token))))
	defer span.End()

	tokenHash := hasher.HashToken(token)

	roles, err := r.db.AuthRoles.
		Query().
		Where(authroles.HasTokenWith(
			authtokens.Token(tokenHash),
		)).
		All(ctx)
	if err != nil {
		recordSpanError(span, err)
		return nil, err
	}

	roleStrings := lo.Map(roles, func(role *ent.AuthRoles, _ int) string {
		return role.Role.String()
	})

	span.SetAttributes(attribute.Int("roles.count", len(roleStrings)))
	return new(set.New(roleStrings...)), nil
}

// CreateToken Creates a token for a user
func (r *TokenRepository) CreateToken(ctx context.Context, createToken UserAuthTokenCreate, roles ...authroles.Role) (UserAuthToken, error) {
	ctx, span := entityTracer().Start(ctx, "repo.TokenRepository.CreateToken",
		trace.WithAttributes(
			attribute.String("user.id", createToken.UserID.String()),
			attribute.String("token.expires_at", createToken.ExpiresAt.Format(time.RFC3339)),
			attribute.Int("token.roles.count", len(roles)),
		))
	defer span.End()

	tokenCtx, tokenSpan := entityTracer().Start(ctx, "repo.TokenRepository.CreateToken.token")
	dbToken, err := r.db.AuthTokens.Create().
		SetToken(createToken.TokenHash).
		SetUserID(createToken.UserID).
		SetExpiresAt(createToken.ExpiresAt).
		Save(tokenCtx)
	if err != nil {
		recordSpanError(tokenSpan, err)
		tokenSpan.End()
		recordSpanError(span, err)
		return UserAuthToken{}, err
	}
	tokenSpan.End()

	if len(roles) > 0 {
		rolesCtx, rolesSpan := entityTracer().Start(ctx, "repo.TokenRepository.CreateToken.roles",
			trace.WithAttributes(attribute.Int("roles.count", len(roles))))
		for _, role := range roles {
			_, err := r.db.AuthRoles.Create().
				SetRole(role).
				SetToken(dbToken).
				Save(rolesCtx)
			if err != nil {
				recordSpanError(rolesSpan, err)
				rolesSpan.End()
				recordSpanError(span, err)
				return UserAuthToken{}, err
			}
		}
		rolesSpan.End()
	}

	return UserAuthToken{
		UserAuthTokenCreate: UserAuthTokenCreate{
			TokenHash: dbToken.Token,
			UserID:    createToken.UserID,
			ExpiresAt: dbToken.ExpiresAt,
		},
		CreatedAt: dbToken.CreatedAt,
	}, nil
}

// DeleteToken remove a single token from the database - equivalent to revoke or logout
func (r *TokenRepository) DeleteToken(ctx context.Context, token []byte) error {
	ctx, span := entityTracer().Start(ctx, "repo.TokenRepository.DeleteToken",
		trace.WithAttributes(attribute.Int("token.hash.length", len(token))))
	defer span.End()

	deleted, err := r.db.AuthTokens.Delete().Where(authtokens.Token(token)).Exec(ctx)
	if err != nil {
		recordSpanError(span, err)
		return err
	}
	span.SetAttributes(attribute.Int("tokens.deleted.count", deleted))
	return nil
}

// PurgeExpiredTokens removes all expired tokens from the database
func (r *TokenRepository) PurgeExpiredTokens(ctx context.Context) (int, error) {
	ctx, span := entityTracer().Start(ctx, "repo.TokenRepository.PurgeExpiredTokens")
	defer span.End()

	tokensDeleted, err := r.db.AuthTokens.Delete().Where(authtokens.ExpiresAtLTE(time.Now())).Exec(ctx)
	if err != nil {
		recordSpanError(span, err)
		return 0, err
	}

	span.SetAttributes(attribute.Int("tokens.deleted.count", tokensDeleted))
	return tokensDeleted, nil
}

func (r *TokenRepository) DeleteAll(ctx context.Context) (int, error) {
	ctx, span := entityTracer().Start(ctx, "repo.TokenRepository.DeleteAll")
	defer span.End()

	amount, err := r.db.AuthTokens.Delete().Exec(ctx)
	if err != nil {
		recordSpanError(span, err)
		return 0, err
	}
	span.SetAttributes(attribute.Int("tokens.deleted.count", amount))
	return amount, nil
}
