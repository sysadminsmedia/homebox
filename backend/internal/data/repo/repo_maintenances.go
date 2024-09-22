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

type MaintenancesFilterStatus string

const (
	MaintenancesFilterStatusScheduled MaintenancesFilterStatus = "scheduled"
	MaintenancesFilterStatusCompleted MaintenancesFilterStatus = "completed"
	MaintenancesFilterStatusBoth      MaintenancesFilterStatus = "both"
)

type MaintenancesFilters struct {
	Status MaintenancesFilterStatus `json:"status" schema:"status"`
}

func (r *MaintenanceEntryRepository) GetAllMaintenances(ctx context.Context, groupID uuid.UUID, filters MaintenancesFilters) ([]MaintenanceEntryWithDetails, error) {
	query := r.db.MaintenanceEntry.Query().Where(
		maintenanceentry.HasItemWith(
			item.HasGroupWith(group.IDEQ(groupID)),
		),
	)

	if filters.Status == MaintenancesFilterStatusScheduled {
		query = query.Where(maintenanceentry.Or(
			maintenanceentry.DateIsNil(),
			maintenanceentry.DateEQ(time.Time{}),
		))
	} else if filters.Status == MaintenancesFilterStatusCompleted {
		query = query.Where(
			maintenanceentry.Not(maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{})),
			))
	}
	entries, err := query.WithItem().Order(maintenanceentry.ByScheduledDate()).All(ctx)

	if err != nil {
		return nil, err
	}

	return mapEachMaintenanceEntryWithDetails(entries), nil
}
