// Package repo provides the data access layer for the application.
package repo

import (
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// AllRepos is a container for all the repository interfaces
type AllRepos struct {
	Users       *UserRepository
	AuthTokens  *TokenRepository
	Groups      *GroupRepository
	Locations   *LocationRepository
	Tags      	*TagRepository
	Items       *ItemsRepository
	Attachments *AttachmentRepo
	MaintEntry  *MaintenanceEntryRepository
	Notifiers   *NotifierRepository
	EntityType  *EntityTypeRepository
	Entities    *EntitiesRepository
	ItemTemplates *ItemTemplatesRepository
}

func New(db *ent.Client, bus *eventbus.EventBus, storage config.Storage, pubSubConn string, thumbnail config.Thumbnail) *AllRepos {
	attachments := &AttachmentRepo{db, storage, pubSubConn, thumbnail}
	return &AllRepos{
		Users:       &UserRepository{db},
		AuthTokens:  &TokenRepository{db},
		Groups:      NewGroupRepository(db),
		Locations:   &LocationRepository{db, bus},
		Tags:      	 &TagRepository{db, bus},
		Items:       &ItemsRepository{db, bus},
		Attachments: attachments,
		MaintEntry:  &MaintenanceEntryRepository{db},
		Notifiers:   NewNotifierRepository(db),
		EntityType:  &EntityTypeRepository{db, bus},
		Entities:    &EntitiesRepository{db, bus},
		ItemTemplates: &ItemTemplatesRepository{db, bus},
	}
}
