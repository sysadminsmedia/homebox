package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
)

type TagRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	TagCreate struct {
		Name        string `json:"name"        validate:"required,min=1,max=255"`
		Description string `json:"description" validate:"max=1000"`
		Color       string `json:"color"`
	}

	TagUpdate struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"        validate:"required,min=1,max=255"`
		Description string    `json:"description" validate:"max=1000"`
		Color       string    `json:"color"`
	}

	TagSummary struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Color       string    `json:"color"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	TagOut struct {
		TagSummary
	}
)

func mapTagSummary(tag *ent.Tag) TagSummary {
	return TagSummary{
		ID:          tag.ID,
		Name:        tag.Name,
		Description: tag.Description,
		Color:       tag.Color,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
}

var (
	mapTagOutErr = mapTErrFunc(mapTagOut)
	mapTagsOut   = mapTEachErrFunc(mapTagSummary)
)

func mapTagOut(tag *ent.Tag) TagOut {
	return TagOut{
		TagSummary: mapTagSummary(tag),
	}
}

func (r *TagRepository) publishMutationEvent(gid uuid.UUID) {
	if r.bus != nil {
		r.bus.Publish(eventbus.EventTagMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

func (r *TagRepository) getOne(ctx context.Context, where ...predicate.Tag) (TagOut, error) {
	return mapTagOutErr(r.db.Tag.Query().
		Where(where...).
		WithGroup().
		Only(ctx),
	)
}

func (r *TagRepository) GetOne(ctx context.Context, id uuid.UUID) (TagOut, error) {
	return r.getOne(ctx, tag.ID(id))
}

func (r *TagRepository) GetOneByGroup(ctx context.Context, gid, ld uuid.UUID) (TagOut, error) {
	return r.getOne(ctx, tag.ID(ld), tag.HasGroupWith(group.ID(gid)))
}

func (r *TagRepository) GetAll(ctx context.Context, groupID uuid.UUID) ([]TagSummary, error) {
	return mapTagsOut(r.db.Tag.Query().
		Where(tag.HasGroupWith(group.ID(groupID))).
		Order(ent.Asc(tag.FieldName)).
		WithGroup().
		All(ctx),
	)
}

func (r *TagRepository) Create(ctx context.Context, groupID uuid.UUID, data TagCreate) (TagOut, error) {
	tag, err := r.db.Tag.Create().
		SetName(data.Name).
		SetDescription(data.Description).
		SetColor(data.Color).
		SetGroupID(groupID).
		Save(ctx)
	if err != nil {
		return TagOut{}, err
	}

	tag.Edges.Group = &ent.Group{ID: groupID} // bootstrap group ID
	r.publishMutationEvent(groupID)
	return mapTagOut(tag), err
}

func (r *TagRepository) update(ctx context.Context, data TagUpdate, where ...predicate.Tag) (int, error) {
	if len(where) == 0 {
		panic("empty where not supported empty")
	}

	return r.db.Tag.Update().
		Where(where...).
		SetName(data.Name).
		SetDescription(data.Description).
		SetColor(data.Color).
		Save(ctx)
}

func (r *TagRepository) UpdateByGroup(ctx context.Context, gid uuid.UUID, data TagUpdate) (TagOut, error) {
	_, err := r.update(ctx, data, tag.ID(data.ID), tag.HasGroupWith(group.ID(gid)))
	if err != nil {
		return TagOut{}, err
	}

	r.publishMutationEvent(gid)
	return r.GetOne(ctx, data.ID)
}

// delete removes the tag from the database. This should only be used when
// the tag's ownership is already confirmed/validated.
func (r *TagRepository) delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Tag.DeleteOneID(id).Exec(ctx)
}

func (r *TagRepository) DeleteByGroup(ctx context.Context, gid, id uuid.UUID) error {
	_, err := r.db.Tag.Delete().
		Where(
			tag.ID(id),
			tag.HasGroupWith(group.ID(gid)),
		).Exec(ctx)
	if err != nil {
		return err
	}

	r.publishMutationEvent(gid)

	return nil
}
