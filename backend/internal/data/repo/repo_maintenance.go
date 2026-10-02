package repo

import (
	"context"
	"fmt"
	"time"

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
		// DateGT(now) belongs here, and its absence is what made the two maintenance
		// endpoints disagree: a completedDate in the FUTURE is a date that has not
		// happened yet, so the entry is scheduled rather than completed. The per-entity
		// query (GetMaintenanceByItemID) has always applied this clause; this one did
		// not, so the same entry read as scheduled from one endpoint and completed from
		// the other, with the same status filter.
		query = query.Where(maintenanceentry.Or(
			maintenanceentry.DateIsNil(),
			maintenanceentry.DateEQ(time.Time{}),
			maintenanceentry.DateGT(time.Now()),
		))
	case MaintenanceFilterStatusCompleted:
		query = query.Where(
			maintenanceentry.Not(maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{}),
				maintenanceentry.DateGT(time.Now())),
			))
	case MaintenanceFilterStatusBoth:
		// No additional filters needed
	default:
		return nil, fmt.Errorf("unknown status %s", filters.Status)
	}
	entries, err := query.WithEntity().Order(maintenanceentry.ByScheduledDate()).All(ctx)

	if err != nil {
		return nil, err
	}

	return mapEachMaintenanceEntryWithDetails(entries), nil
}
