package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entitytype"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
)

type EntityTypeRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

type (
	EntityType struct {
		Name        string `json:"name"`
		IsLocation  bool   `json:"isLocation"`
		Description string `json:"description" extension:"x-nullable"`
		Icon        string `json:"icon"        extension:"x-nullable"`
		Color       string `json:"color"       extension:"x-nullable"`
	}

	EntityTypeCreate struct {
		Name        string `json:"name"        validate:"required"`
		IsLocation  bool   `json:"isLocation"  validate:"required"`
		Description string `json:"description" extension:"x-nullable"`
		Icon        string `json:"icon"        extension:"x-nullable"`
		Color       string `json:"color"       extension:"x-nullable"`
	}

	EntityTypeUpdate struct {
		Name        string `json:"name"        validate:"omitempty,min=1"`
		Description string `json:"description" extension:"x-nullable"`
		Icon        string `json:"icon"        extension:"x-nullable"`
		Color       string `json:"color"       extension:"x-nullable"`
	}
)

func (e *EntityTypeRepository) CreateDefaultEntityTypes(ctx context.Context, gid uuid.UUID) error {
	_, err := e.db.EntityType.Create().SetIsLocation(true).SetName("Location").SetGroupID(gid).Save(ctx)
	if err != nil {
		return err
	}
	_, err = e.db.EntityType.Create().SetIsLocation(false).SetName("Item").SetGroupID(gid).Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (e *EntityTypeRepository) GetOneByGroup(ctx context.Context, gid uuid.UUID, id uuid.UUID) (EntityType, error) {
	entityType, err := e.db.EntityType.Query().Where(entitytype.HasGroupWith(group.ID(gid)), entitytype.ID(id)).Only(ctx)
	if err != nil {
		return EntityType{}, err
	}
	return EntityType{
		Name:        entityType.Name,
		IsLocation:  entityType.IsLocation,
		Description: entityType.Description,
		Icon:        entityType.Icon,
		Color:       entityType.Color,
	}, nil
}

func (e *EntityTypeRepository) GetEntityTypesByGroupID(ctx context.Context, gid uuid.UUID) ([]EntityType, error) {
	entityTypes, err := e.db.EntityType.Query().Where(entitytype.HasGroupWith(group.ID(gid))).All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]EntityType, 0, len(entityTypes))
	for _, et := range entityTypes {
		result = append(result, EntityType{
			Name:        et.Name,
			IsLocation:  et.IsLocation,
			Description: et.Description,
			Icon:        et.Icon,
			Color:       et.Color,
		})
	}
	return result, nil
}

func (e *EntityTypeRepository) CreateEntityType(ctx context.Context, gid uuid.UUID, data EntityTypeCreate) (EntityType, error) {
	entityType, err := e.db.EntityType.Create().
		SetGroupID(gid).
		SetName(data.Name).
		SetIsLocation(data.IsLocation).
		SetDescription(data.Description).
		SetIcon(data.Icon).
		SetColor(data.Color).
		Save(ctx)
	if err != nil {
		return EntityType{}, err
	}
	return EntityType{
		Name:        entityType.Name,
		IsLocation:  entityType.IsLocation,
		Description: entityType.Description,
		Icon:        entityType.Icon,
		Color:       entityType.Color,
	}, nil
}

func (e *EntityTypeRepository) UpdateEntityType(ctx context.Context, gid uuid.UUID, id uuid.UUID, data EntityTypeUpdate) (EntityType, error) {
	et, err := e.GetOneByGroup(ctx, gid, id)
	if err != nil {
		return EntityType{}, err
	}
	update := e.db.EntityType.Update().Where(entitytype.ID(id))
	if data.Name != "" {
		update = update.SetName(data.Name)
		et.Name = data.Name
	}
	if data.Description != "" {
		update = update.SetDescription(data.Description)
		et.Description = data.Description
	}
	if data.Icon != "" {
		update = update.SetIcon(data.Icon)
		et.Icon = data.Icon
	}
	if data.Color != "" {
		update = update.SetColor(data.Color)
		et.Color = data.Color
	}
	_, err = update.Save(ctx)
	if err != nil {
		return EntityType{}, err
	}
	return et, nil
}
