// Package repo provides the data access layer for the application.
package repo

import (
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
)

// AllRepos is a container for all the repository interfaces
type AllRepos struct {
	Users       *UserRepository
	AuthTokens  *TokenRepository
	Groups      *GroupRepository
	Locations   *LocationRepository
	Labels      *LabelRepository
	Items       *ItemsRepository
	Attachments *AttachmentRepo
	MaintEntry  *MaintenanceEntryRepository
	Notifiers   *NotifierRepository
}

func New(db *ent.Client, bus *eventbus.EventBus, root string) *AllRepos {
	return &AllRepos{
		Users:       &UserRepository{db},
		AuthTokens:  &TokenRepository{db},
		Groups:      NewGroupRepository(db),
		Locations:   &LocationRepository{db, bus},
		Labels:      &LabelRepository{db, bus},
		Items:       &ItemsRepository{db, bus},
		Attachments: &AttachmentRepo{db, root},
		MaintEntry:  &MaintenanceEntryRepository{db},
		Notifiers:   NewNotifierRepository(db),
	}
}
