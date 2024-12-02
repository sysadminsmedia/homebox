package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/item"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

// MaintenanceEntryRepository is a repository for maintenance entries that are
// associated with an item in the database. An entry represents a maintenance event
// that has been performed on an item.
type MaintenanceEntryRepository struct {
	db *ent.Client
}

type MaintenanceEntryCreate struct {
	CompletedDate types.Date `json:"completedDate"`
	ScheduledDate types.Date `json:"scheduledDate"`
	Name          string     `json:"name"          validate:"required"`
	Description   string     `json:"description"`
	Cost          float64    `json:"cost,string"`
}

func (mc MaintenanceEntryCreate) Validate() error {
	if mc.CompletedDate.Time().IsZero() && mc.ScheduledDate.Time().IsZero() {
		return errors.New("either completedDate or scheduledDate must be set")
	}
	return nil
}

type MaintenanceEntryUpdate struct {
	CompletedDate types.Date `json:"completedDate"`
	ScheduledDate types.Date `json:"scheduledDate"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Cost          float64    `json:"cost,string"`
}

func (mu MaintenanceEntryUpdate) Validate() error {
	if mu.CompletedDate.Time().IsZero() && mu.ScheduledDate.Time().IsZero() {
		return errors.New("either completedDate or scheduledDate must be set")
	}
	return nil
}

type (
	MaintenanceEntry struct {
		ID            uuid.UUID  `json:"id"`
		CompletedDate types.Date `json:"completedDate"`
		ScheduledDate types.Date `json:"scheduledDate"`
		Name          string     `json:"name"`
		Description   string     `json:"description"`
		Cost          float64    `json:"cost,string"`
	}
)

var (
	mapMaintenanceEntryErr  = mapTErrFunc(mapMaintenanceEntry)
	mapEachMaintenanceEntry = mapTEachFunc(mapMaintenanceEntry)
)

func mapMaintenanceEntry(entry *ent.MaintenanceEntry) MaintenanceEntry {
	return MaintenanceEntry{
		ID:            entry.ID,
		CompletedDate: types.Date(entry.Date),
		ScheduledDate: types.Date(entry.ScheduledDate),
		Name:          entry.Name,
		Description:   entry.Description,
		Cost:          entry.Cost,
	}
}

func (r *MaintenanceEntryRepository) GetScheduled(ctx context.Context, gid uuid.UUID, dt types.Date) ([]MaintenanceEntry, error) {
	entries, err := r.db.MaintenanceEntry.Query().
		Where(
			maintenanceentry.HasItemWith(
				item.HasGroupWith(group.ID(gid)),
			),
			maintenanceentry.ScheduledDate(dt.Time()),
			maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{}),
			),
		).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return mapEachMaintenanceEntry(entries), nil
}

func (r *MaintenanceEntryRepository) Create(ctx context.Context, itemID uuid.UUID, input MaintenanceEntryCreate) (MaintenanceEntry, error) {
	item, err := r.db.MaintenanceEntry.Create().
		SetItemID(itemID).
		SetDate(input.CompletedDate.Time()).
		SetScheduledDate(input.ScheduledDate.Time()).
		SetName(input.Name).
		SetDescription(input.Description).
		SetCost(input.Cost).
		Save(ctx)

	return mapMaintenanceEntryErr(item, err)
}

func (r *MaintenanceEntryRepository) Update(ctx context.Context, id uuid.UUID, input MaintenanceEntryUpdate) (MaintenanceEntry, error) {
	item, err := r.db.MaintenanceEntry.UpdateOneID(id).
		SetDate(input.CompletedDate.Time()).
		SetScheduledDate(input.ScheduledDate.Time()).
		SetName(input.Name).
		SetDescription(input.Description).
		SetCost(input.Cost).
		Save(ctx)

	return mapMaintenanceEntryErr(item, err)
}

func (r *MaintenanceEntryRepository) GetMaintenanceByItemID(ctx context.Context, groupID, itemID uuid.UUID, filters MaintenanceFilters) ([]MaintenanceEntryWithDetails, error) {
	query := r.db.MaintenanceEntry.Query().Where(
		maintenanceentry.ItemID(itemID),
		maintenanceentry.HasItemWith(
			item.HasGroupWith(group.IDEQ(groupID)),
		),
	)
	if filters.Status == MaintenanceFilterStatusScheduled {
		query = query.Where(maintenanceentry.Or(
			maintenanceentry.DateIsNil(),
			maintenanceentry.DateEQ(time.Time{}),
			maintenanceentry.DateGT(time.Now()),
		))
	} else if filters.Status == MaintenanceFilterStatusCompleted {
		query = query.Where(
			maintenanceentry.Not(maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{}),
				maintenanceentry.DateGT(time.Now()),
			)))
	}
	entries, err := query.WithItem().Order(maintenanceentry.ByScheduledDate()).All(ctx)

	if err != nil {
		return []MaintenanceEntryWithDetails{}, err
	}

	return mapEachMaintenanceEntryWithDetails(entries), nil
}

func (r *MaintenanceEntryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.MaintenanceEntry.DeleteOneID(id).Exec(ctx)
}
