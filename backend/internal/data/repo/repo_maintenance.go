package repo

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
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
		ItemName:         entry.Edges.Entity.Name,
		ItemID:           entry.EntityID,
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
		maintenanceentry.HasEntityWith(
			entity.HasGroupWith(group.IDEQ(groupID)),
		),
	)

	switch filters.Status {
	case MaintenanceFilterStatusScheduled:
		query = query.Where(maintenanceentry.Or(
			maintenanceentry.DateIsNil(),
			maintenanceentry.DateEQ(time.Time{}),
			maintenanceentry.DateGT(time.Now()),
		))
		query = query.Order(maintenanceentry.ByScheduledDate(sql.OrderAsc()))
	case MaintenanceFilterStatusCompleted:
		query = query.Where(
			maintenanceentry.Not(maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{}),
				maintenanceentry.DateGT(time.Now())),
			))
		query = query.Order(maintenanceentry.ByDate(sql.OrderDesc()))
	case MaintenanceFilterStatusBoth, "":
		query = query.Order(
			maintenanceentry.ByScheduledDate(sql.OrderDesc()),
			maintenanceentry.ByDate(sql.OrderDesc()),
		)
	default:
		return nil, fmt.Errorf("unknown status %s", filters.Status)
	}
	entries, err := query.WithEntity().All(ctx)

	if err != nil {
		return nil, err
	}

	return mapEachMaintenanceEntryWithDetails(entries), nil
}
