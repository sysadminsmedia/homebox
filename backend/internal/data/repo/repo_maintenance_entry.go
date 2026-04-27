package repo

import (
	"context"
	"errors"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
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
	PlanID        uuid.UUID  `json:"planID"`
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
	PlanID        uuid.UUID  `json:"planID"`
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
		PlanID        uuid.UUID  `json:"planID,omitempty"`
		IsOverdue     bool       `json:"isOverdue"`
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
	planID := uuid.Nil
	if entry.PlanID != nil {
		planID = *entry.PlanID
	}
	return MaintenanceEntry{
		ID:            entry.ID,
		CompletedDate: types.Date(entry.Date),
		ScheduledDate: types.Date(entry.ScheduledDate),
		PlanID:        planID,
		IsOverdue:     isEntryOverdue(entry.Date, entry.ScheduledDate),
		Name:          entry.Name,
		Description:   entry.Description,
		Cost:          entry.Cost,
	}
}

func (r *MaintenanceEntryRepository) GetScheduled(ctx context.Context, gid uuid.UUID, dt types.Date) ([]MaintenanceEntry, error) {
	entries, err := r.db.MaintenanceEntry.Query().
		Where(
			maintenanceentry.HasEntityWith(
				entity.HasGroupWith(group.ID(gid)),
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
		SetEntityID(itemID).
		SetDate(input.CompletedDate.Time()).
		SetScheduledDate(input.ScheduledDate.Time()).
		SetName(input.Name).
		SetDescription(input.Description).
		SetCost(input.Cost).
		Save(ctx)
	if err != nil {
		return mapMaintenanceEntryErr(item, err)
	}

	if input.PlanID != uuid.Nil {
		item, err = r.db.MaintenanceEntry.UpdateOneID(item.ID).
			SetPlanID(input.PlanID).
			Save(ctx)
	}

	return mapMaintenanceEntryErr(item, err)
}

func (r *MaintenanceEntryRepository) Update(ctx context.Context, id uuid.UUID, input MaintenanceEntryUpdate) (MaintenanceEntry, error) {
	current, err := r.db.MaintenanceEntry.Query().Where(maintenanceentry.IDEQ(id)).Only(ctx)
	if err != nil {
		return MaintenanceEntry{}, err
	}

	completedDate := input.CompletedDate.Time()
	if completedDate.IsZero() {
		completedDate = current.Date
	}

	scheduledDate := input.ScheduledDate.Time()
	if scheduledDate.IsZero() {
		scheduledDate = current.ScheduledDate
	}

	name := input.Name
	if name == "" {
		name = current.Name
	}

	description := input.Description
	if description == "" {
		description = current.Description
	}

	cost := input.Cost
	if input.Cost == 0 && current.Cost != 0 {
		cost = current.Cost
	}

	updater := r.db.MaintenanceEntry.UpdateOneID(id).
		SetDate(completedDate).
		SetScheduledDate(scheduledDate).
		SetName(name).
		SetDescription(description).
		SetCost(cost)

	if input.PlanID != uuid.Nil {
		updater = updater.SetPlanID(input.PlanID)
	} else {
		updater = updater.ClearPlanID()
	}

	item, err := updater.Save(ctx)
	if err != nil {
		return mapMaintenanceEntryErr(item, err)
	}

	if current.Date.IsZero() && !completedDate.IsZero() && item.PlanID != nil && *item.PlanID != uuid.Nil {
		if _, err = r.rollPlanFromCompletion(ctx, *item.PlanID, completedDate, item.EntityID); err != nil {
			return MaintenanceEntry{}, err
		}
	}

	return mapMaintenanceEntryErr(item, err)
}

func (r *MaintenanceEntryRepository) GetMaintenanceByItemID(ctx context.Context, groupID, itemID uuid.UUID, filters MaintenanceFilters) ([]MaintenanceEntryWithDetails, error) {
	query := r.db.MaintenanceEntry.Query().Where(
		maintenanceentry.EntityID(itemID),
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
		// Sort scheduled entries by ascending scheduled date
		query = query.Order(
			maintenanceentry.ByScheduledDate(sql.OrderAsc()),
		)
	case MaintenanceFilterStatusCompleted:
		query = query.Where(
			maintenanceentry.Not(maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{}),
				maintenanceentry.DateGT(time.Now()),
			)))
		// Sort completed entries by descending completion date
		query = query.Order(
			maintenanceentry.ByDate(sql.OrderDesc()),
		)
	case MaintenanceFilterStatusOverdue:
		query = query.Where(
			maintenanceentry.ScheduledDateLT(time.Now()),
			maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{}),
			),
		)
		query = query.Order(
			maintenanceentry.ByScheduledDate(sql.OrderAsc()),
		)
	default:
		// Sort entries by default by scheduled and maintenance date in descending order
		query = query.Order(
			maintenanceentry.ByScheduledDate(sql.OrderDesc()),
			maintenanceentry.ByDate(sql.OrderDesc()),
		)
	}
	entries, err := query.WithEntity().All(ctx)

	if err != nil {
		return []MaintenanceEntryWithDetails{}, err
	}

	return mapEachMaintenanceEntryWithDetails(entries), nil
}

func isEntryOverdue(completedDate, scheduledDate time.Time) bool {
	if scheduledDate.IsZero() {
		return false
	}

	if completedDate.IsZero() {
		return scheduledDate.Before(time.Now())
	}

	return false
}

func (r *MaintenanceEntryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.MaintenanceEntry.DeleteOneID(id).Exec(ctx)
}
