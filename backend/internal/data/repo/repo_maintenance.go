package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/item"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
)

type (
	MaintenanceEntryWithDetails struct {
		MaintenanceEntry
		ItemName string    `json:"itemName"`
		ItemID   uuid.UUID `json:"itemID"`
	}
)

var (
	mapEachMaintenanceEntryWithDetails = mapTEachFunc(mapMaintenanceEntryWithDetails)
)

func mapMaintenanceEntryWithDetails(entry *ent.MaintenanceEntry) MaintenanceEntryWithDetails {
	return MaintenanceEntryWithDetails{
		MaintenanceEntry: mapMaintenanceEntry(entry),
		ItemName:         entry.Edges.Item.Name,
		ItemID:           entry.ItemID,
	}
}

type MaintenanceFilterStatus string

const (
	MaintenanceFilterStatusScheduled MaintenanceFilterStatus = "scheduled"
	MaintenanceFilterStatusCompleted MaintenanceFilterStatus = "completed"
	MaintenanceFilterStatusBoth      MaintenanceFilterStatus = "both"
)

type MaintenanceFilters struct {
	Status MaintenanceFilterStatus `json:"status" schema:"status"`
}

func (r *MaintenanceEntryRepository) GetAllMaintenance(ctx context.Context, groupID uuid.UUID, filters MaintenanceFilters) ([]MaintenanceEntryWithDetails, error) {
	query := r.db.MaintenanceEntry.Query().Where(
		maintenanceentry.HasItemWith(
			item.HasGroupWith(group.IDEQ(groupID)),
		),
	)

	switch filters.Status {
	case MaintenanceFilterStatusScheduled:
		query = query.Where(maintenanceentry.Or(
			maintenanceentry.DateIsNil(),
			maintenanceentry.DateEQ(time.Time{}),
		))
	case MaintenanceFilterStatusCompleted:
		query = query.Where(
			maintenanceentry.Not(maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{})),
			))
	case MaintenanceFilterStatusBoth:
		// No additional filters needed
	}
	entries, err := query.WithItem().Order(maintenanceentry.ByScheduledDate()).All(ctx)

	if err != nil {
		return nil, err
	}

	return mapEachMaintenanceEntryWithDetails(entries), nil
}
