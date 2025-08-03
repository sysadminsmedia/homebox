package repo

import (
	"context"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
)

type EntityTypeRepository struct {
	db  *ent.Client
	bus *eventbus.EventBus
}

func (e *EntityTypeRepository) CreateDefaultEntities(ctx context.Context, gid uuid.UUID) error {
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
