package repo

import (
	"context"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/oauth"
	"time"
)

type OAuthRepository struct {
	db *ent.Client
}

type (
	OAuthCreate struct {
		Provider string    `json:"provider"`
		Subject  string    `json:"sub"`
		UserId   uuid.UUID `json:"userId"`
	}

	OAuth struct {
		OAuthCreate
		CreatedAt time.Time `json:"createdAt"`
	}
)

func (r *OAuthRepository) GetUserFromToken(ctx context.Context, provider string, sub string) (UserOut, error) {
	user, err := r.db.OAuth.Query().
		Where(oauth.Provider(provider)).
		Where(oauth.Sub(sub)).
		WithUser().
		QueryUser().
		WithGroup().
		Only(ctx)
	if err != nil {
		return UserOut{}, err
	}

	return mapUserOut(user), nil
}

func (r *OAuthRepository) Create(ctx context.Context, create OAuthCreate) (OAuth, error) {
	dbOauth, err := r.db.OAuth.Create().
		SetProvider(create.Provider).
		SetSub(create.Subject).
		SetUserID(create.UserId).
		Save(ctx)
	if err != nil {
		return OAuth{}, err
	}

	return OAuth{
		OAuthCreate: OAuthCreate{
			Provider: dbOauth.Provider,
			Subject:  dbOauth.Sub,
			UserId:   create.UserId,
		},
		CreatedAt: dbOauth.CreatedAt,
	}, nil
}

// TODO: delete connection, checking if password or other connections exists or delete user
