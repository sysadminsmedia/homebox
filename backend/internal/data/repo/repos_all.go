// Package repo provides the data access layer for the application.
package repo

import (
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// AllRepos is a container for all the repository interfaces
type AllRepos struct {
	Users               *UserRepository
	AuthTokens          *TokenRepository
	PasswordResetTokens *PasswordResetTokenRepository
	APIKeys             *APIKeyRepository
	Groups              *GroupRepository
	Entities            *EntityRepository
	EntityTypes         *EntityTypeRepository
	EntityTemplates     *EntityTemplatesRepository
	Tags                *TagRepository
	Attachments         *AttachmentRepo
	MaintEntry          *MaintenanceEntryRepository
	Notifiers           *NotifierRepository
	Exports             *ExportRepository
	Permissions         *PermissionsRepository
}

func New(db *ent.Client, bus *eventbus.EventBus, storage config.Storage, pubSubConn string, thumbnail config.Thumbnail) *AllRepos {
	attachments := &AttachmentRepo{db, storage, pubSubConn, thumbnail}
	return &AllRepos{
		Users:               &UserRepository{db},
		AuthTokens:          &TokenRepository{db},
		PasswordResetTokens: &PasswordResetTokenRepository{db},
		APIKeys:             NewAPIKeyRepository(db),
		Groups:              NewGroupRepository(db),
		Entities:            &EntityRepository{db, bus, attachments},
		EntityTypes:         &EntityTypeRepository{db, bus},
		EntityTemplates:     &EntityTemplatesRepository{db, bus},
		Tags:                &TagRepository{db, bus},
		Attachments:         attachments,
		MaintEntry:          &MaintenanceEntryRepository{db},
		Notifiers:           NewNotifierRepository(db),
		Exports:             &ExportRepository{db},
		Permissions:         &PermissionsRepository{db},
	}
}
