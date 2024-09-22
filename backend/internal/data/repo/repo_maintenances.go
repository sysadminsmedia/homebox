package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/item"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

type (
	MaintenanceEntryWithDetails struct {
		ID            uuid.UUID  `json:"id"`
		CompletedDate types.Date `json:"completedDate"`
		ScheduledDate types.Date `json:"scheduledDate"`
		Name          string     `json:"name"`
		Description   string     `json:"description"`
		Cost          float64    `json:"cost,string"`
		ItemName      string     `json:"itemName"`
		ItemID        uuid.UUID  `json:"itemID"`
	}
)

var (
	mapEachMaintenanceEntryWithDetails = mapTEachFunc(mapMaintenanceEntryWithDetails)
)

func mapMaintenanceEntryWithDetails(entry *ent.MaintenanceEntry) MaintenanceEntryWithDetails {
	return MaintenanceEntryWithDetails{
		ID:            entry.ID,
		CompletedDate: types.Date(entry.Date),
		ScheduledDate: types.Date(entry.ScheduledDate),
		Name:          entry.Name,
		Description:   entry.Description,
		Cost:          entry.Cost,
		ItemName:      entry.Edges.Item.Name,
		ItemID:        entry.ItemID,
	}
}

type MaintenancesFilter string

const (
	MaintenancesFilterScheduled MaintenancesFilter = "scheduled"
	MaintenancesFilterCompleted MaintenancesFilter = "completed"
	MaintenancesFilterBoth      MaintenancesFilter = "both"
)

func (r *MaintenanceEntryRepository) GetAllMaintenances(ctx context.Context, groupID uuid.UUID, filters MaintenancesFilter) ([]MaintenanceEntryWithDetails, error) {
	query := r.db.MaintenanceEntry.Query().Where(
		maintenanceentry.HasItemWith(
			item.HasGroupWith(group.IDEQ(groupID)),
		),
	)

	if filters == MaintenancesFilterScheduled {
		query = query.Where(maintenanceentry.Or(
			maintenanceentry.DateIsNil(),
			maintenanceentry.DateEQ(time.Time{}),
		))
	} else if filters == MaintenancesFilterCompleted {
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
