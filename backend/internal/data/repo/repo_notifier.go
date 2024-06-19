package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/notifier"
)

type NotifierRepository struct {
	db     *ent.Client
	mapper MapFunc[*ent.Notifier, NotifierOut]
}

func NewNotifierRepository(db *ent.Client) *NotifierRepository {
	return &NotifierRepository{
		db: db,
		mapper: func(n *ent.Notifier) NotifierOut {
			return NotifierOut{
				ID:        n.ID,
				UserID:    n.UserID,
				GroupID:   n.GroupID,
				CreatedAt: n.CreatedAt,
				UpdatedAt: n.UpdatedAt,

				Name:     n.Name,
				IsActive: n.IsActive,
				URL:      n.URL,
			}
		},
	}
}

type (
	NotifierCreate struct {
		Name     string `json:"name"     validate:"required,min=1,max=255"`
		IsActive bool   `json:"isActive"`
		URL      string `json:"url"      validate:"required,shoutrrr"`
	}

	NotifierUpdate struct {
		Name     string  `json:"name"     validate:"required,min=1,max=255"`
		IsActive bool    `json:"isActive"`
		URL      *string `json:"url"      validate:"omitempty,shoutrrr"     extensions:"x-nullable"`
	}

	NotifierOut struct {
		ID        uuid.UUID `json:"id"`
		UserID    uuid.UUID `json:"userId"`
		GroupID   uuid.UUID `json:"groupId"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`

		Name     string `json:"name"`
		IsActive bool   `json:"isActive"`
		URL      string `json:"-"` // URL field is not exposed to the client
	}
)

func (r *NotifierRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]NotifierOut, error) {
	notifier, err := r.db.Notifier.Query().
		Where(notifier.UserID(userID)).
		Order(ent.Asc(notifier.FieldName)).
		All(ctx)

	return r.mapper.MapEachErr(notifier, err)
}

func (r *NotifierRepository) GetByGroup(ctx context.Context, groupID uuid.UUID) ([]NotifierOut, error) {
	notifier, err := r.db.Notifier.Query().
		Where(notifier.GroupID(groupID)).
		Order(ent.Asc(notifier.FieldName)).
		All(ctx)

	return r.mapper.MapEachErr(notifier, err)
}

func (r *NotifierRepository) GetActiveByGroup(ctx context.Context, groupID uuid.UUID) ([]NotifierOut, error) {
	notifier, err := r.db.Notifier.Query().
		Where(notifier.GroupID(groupID), notifier.IsActive(true)).
		Order(ent.Asc(notifier.FieldName)).
		All(ctx)

	return r.mapper.MapEachErr(notifier, err)
}

func (r *NotifierRepository) Create(ctx context.Context, groupID, userID uuid.UUID, input NotifierCreate) (NotifierOut, error) {
	notifier, err := r.db.Notifier.
		Create().
		SetGroupID(groupID).
		SetUserID(userID).
		SetName(input.Name).
		SetIsActive(input.IsActive).
		SetURL(input.URL).
		Save(ctx)

	return r.mapper.MapErr(notifier, err)
}

func (r *NotifierRepository) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input NotifierUpdate) (NotifierOut, error) {
	q := r.db.Notifier.
		UpdateOneID(id).
		SetName(input.Name).
		SetIsActive(input.IsActive)

	if input.URL != nil {
		q.SetURL(*input.URL)
	}

	notifier, err := q.Save(ctx)

	return r.mapper.MapErr(notifier, err)
}

func (r *NotifierRepository) Delete(ctx context.Context, userID uuid.UUID, ID uuid.UUID) error {
	_, err := r.db.Notifier.Delete().Where(notifier.UserID(userID), notifier.ID(ID)).Exec(ctx)
	return err
}
