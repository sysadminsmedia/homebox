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
	user, err := r.db.AuthTokens.Query().
		Where(authtokens.Token(token)).
		Where(authtokens.ExpiresAtGTE(time.Now())).
		WithUser().
		QueryUser().
		WithGroups().
		Only(ctx)
	if err != nil {
		return UserOut{}, err
	}

	return mapUserOut(user), nil
}

func (r *TokenRepository) GetRoles(ctx context.Context, token string) (*set.Set[string], error) {
	tokenHash := hasher.HashToken(token)

	roles, err := r.db.AuthRoles.
		Query().
		Where(authroles.HasTokenWith(
			authtokens.Token(tokenHash),
		)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	roleStrings := lo.Map(roles, func(role *ent.AuthRoles, _ int) string {
		return role.Role.String()
	})

	roleSet := set.New(roleStrings...)
	return &roleSet, nil
}

// CreateToken Creates a token for a user
func (r *TokenRepository) CreateToken(ctx context.Context, createToken UserAuthTokenCreate, roles ...authroles.Role) (UserAuthToken, error) {
	dbToken, err := r.db.AuthTokens.Create().
		SetToken(createToken.TokenHash).
		SetUserID(createToken.UserID).
		SetExpiresAt(createToken.ExpiresAt).
		Save(ctx)
	if err != nil {
		return UserAuthToken{}, err
	}

	for _, role := range roles {
		_, err := r.db.AuthRoles.Create().
			SetRole(role).
			SetToken(dbToken).
			Save(ctx)
		if err != nil {
			return UserAuthToken{}, err
		}
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
	_, err := r.db.AuthTokens.Delete().Where(authtokens.Token(token)).Exec(ctx)
	return err
}

// PurgeExpiredTokens removes all expired tokens from the database
func (r *TokenRepository) PurgeExpiredTokens(ctx context.Context) (int, error) {
	tokensDeleted, err := r.db.AuthTokens.Delete().Where(authtokens.ExpiresAtLTE(time.Now())).Exec(ctx)
	if err != nil {
		return 0, err
	}

	return tokensDeleted, nil
}

func (r *TokenRepository) DeleteAll(ctx context.Context) (int, error) {
	amount, err := r.db.AuthTokens.Delete().Exec(ctx)
	return amount, err
}
