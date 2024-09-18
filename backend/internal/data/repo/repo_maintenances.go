package repo

import (
	"context"
	"time"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
)

type MaintenancesFilter string

const (
	MaintenancesFilterScheduled MaintenancesFilter = "scheduled"
	MaintenancesFilterCompleted MaintenancesFilter = "completed"
	MaintenancesFilterBoth      MaintenancesFilter = "both"
)

func (r *MaintenanceEntryRepository) GetAllMaintenances(ctx context.Context, filters MaintenancesFilter) ([]MaintenanceEntry, error) {
	query := r.db.MaintenanceEntry.Query()
	if filters == MaintenancesFilterCompleted {
		query = query.Where(maintenanceentry.Or(
			maintenanceentry.DateIsNil(),
			maintenanceentry.DateEQ(time.Time{}),
		))
	} else if filters == MaintenancesFilterScheduled {
		query = query.Where(
			maintenanceentry.Not(maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{})),
			))
	}
	entries, err := query.Order(maintenanceentry.ByScheduledDate()).All(ctx)

	if err != nil {
		return nil, err
	}

	return mapEachMaintenanceEntry(entries), nil
}
