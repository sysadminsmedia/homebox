package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceplan"
)

type MaintenancePlanIntervalUnit string

const (
	MaintenancePlanIntervalUnitHour  MaintenancePlanIntervalUnit = "hour"
	MaintenancePlanIntervalUnitDay   MaintenancePlanIntervalUnit = "day"
	MaintenancePlanIntervalUnitWeek  MaintenancePlanIntervalUnit = "week"
	MaintenancePlanIntervalUnitMonth MaintenancePlanIntervalUnit = "month"
	MaintenancePlanIntervalUnitYear  MaintenancePlanIntervalUnit = "year"
)

type MaintenancePlanCreate struct {
	Name          string                      `json:"name" validate:"required"`
	Description   string                      `json:"description"`
	IntervalValue int                         `json:"intervalValue" validate:"required,min=1"`
	IntervalUnit  MaintenancePlanIntervalUnit `json:"intervalUnit" validate:"required"`
	StartDate     time.Time                   `json:"startDate"`
	Active        bool                        `json:"active"`
}

type MaintenancePlanUpdate struct {
	Name          string                      `json:"name"`
	Description   string                      `json:"description"`
	IntervalValue int                         `json:"intervalValue"`
	IntervalUnit  MaintenancePlanIntervalUnit `json:"intervalUnit"`
	NextDueAt     time.Time                   `json:"nextDueAt"`
	Active        bool                        `json:"active"`
}

type MaintenancePlan struct {
	ID              uuid.UUID                   `json:"id"`
	ItemID          uuid.UUID                   `json:"itemID"`
	Name            string                      `json:"name"`
	Description     string                      `json:"description"`
	IntervalValue   int                         `json:"intervalValue"`
	IntervalUnit    MaintenancePlanIntervalUnit `json:"intervalUnit"`
	Active          bool                        `json:"active"`
	LastCompletedAt time.Time                   `json:"lastCompletedAt"`
	NextDueAt       time.Time                   `json:"nextDueAt"`
}

func (mc MaintenancePlanCreate) Validate() error {
	if mc.IntervalValue < 1 {
		return errors.New("intervalValue must be greater than 0")
	}

	return validateMaintenancePlanUnit(mc.IntervalUnit)
}

func (mu MaintenancePlanUpdate) Validate() error {
	if mu.IntervalValue < 1 {
		return errors.New("intervalValue must be greater than 0")
	}

	return validateMaintenancePlanUnit(mu.IntervalUnit)
}

func validateMaintenancePlanUnit(unit MaintenancePlanIntervalUnit) error {
	switch unit {
	case MaintenancePlanIntervalUnitHour,
		MaintenancePlanIntervalUnitDay,
		MaintenancePlanIntervalUnitWeek,
		MaintenancePlanIntervalUnitMonth,
		MaintenancePlanIntervalUnitYear:
		return nil
	default:
		return errors.New("invalid intervalUnit")
	}
}

func mapMaintenancePlan(entry *ent.MaintenancePlan) MaintenancePlan {
	last := time.Time{}
	next := time.Time{}
	if entry.LastCompletedAt != nil {
		last = *entry.LastCompletedAt
	}
	if entry.NextDueAt != nil {
		next = *entry.NextDueAt
	}

	return MaintenancePlan{
		ID:              entry.ID,
		ItemID:          entry.EntityID,
		Name:            entry.Name,
		Description:     entry.Description,
		IntervalValue:   entry.IntervalValue,
		IntervalUnit:    MaintenancePlanIntervalUnit(entry.IntervalUnit),
		Active:          entry.Active,
		LastCompletedAt: last,
		NextDueAt:       next,
	}
}

func (r *MaintenanceEntryRepository) ListPlansByItemID(ctx context.Context, groupID, itemID uuid.UUID) ([]MaintenancePlan, error) {
	items, err := r.db.MaintenancePlan.Query().
		Where(
			maintenanceplan.HasEntityWith(
				entity.IDEQ(itemID),
				entity.HasGroupWith(group.IDEQ(groupID)),
			),
		).
		Order(ent.Asc(maintenanceplan.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return mapEach(items, mapMaintenancePlan), nil
}

func (r *MaintenanceEntryRepository) CreatePlan(ctx context.Context, itemID uuid.UUID, input MaintenancePlanCreate) (MaintenancePlan, error) {
	base := input.StartDate
	if base.IsZero() {
		base = time.Now().UTC()
	}
	firstDue := base.UTC()
	item, err := r.db.MaintenancePlan.Create().
		SetEntityID(itemID).
		SetName(input.Name).
		SetDescription(input.Description).
		SetIntervalValue(input.IntervalValue).
		SetIntervalUnit(maintenanceplan.IntervalUnit(input.IntervalUnit)).
		SetActive(input.Active).
		SetNextDueAt(firstDue).
		Save(ctx)
	if err != nil {
		return MaintenancePlan{}, err
	}

	_, err = r.db.MaintenanceEntry.Create().
		SetEntityID(itemID).
		SetPlanID(item.ID).
		SetName(item.Name).
		SetDescription(item.Description).
		SetScheduledDate(firstDue).
		SetDate(time.Time{}).
		Save(ctx)
	if err != nil {
		return MaintenancePlan{}, err
	}

	return mapMaintenancePlan(item), nil
}

func (r *MaintenanceEntryRepository) UpdatePlan(ctx context.Context, planID uuid.UUID, input MaintenancePlanUpdate) (MaintenancePlan, error) {
	up := r.db.MaintenancePlan.UpdateOneID(planID).
		SetName(input.Name).
		SetDescription(input.Description).
		SetIntervalValue(input.IntervalValue).
		SetIntervalUnit(maintenanceplan.IntervalUnit(input.IntervalUnit)).
		SetActive(input.Active)
	if input.NextDueAt.IsZero() {
		up = up.ClearNextDueAt()
	} else {
		up = up.SetNextDueAt(input.NextDueAt)
	}

	item, err := up.Save(ctx)
	if err != nil {
		return MaintenancePlan{}, err
	}

	return mapMaintenancePlan(item), nil
}

func (r *MaintenanceEntryRepository) DeletePlan(ctx context.Context, id uuid.UUID) error {
	return r.db.MaintenancePlan.DeleteOneID(id).Exec(ctx)
}

func (r *MaintenanceEntryRepository) rollPlanFromCompletion(ctx context.Context, planID uuid.UUID, completedAt time.Time, itemID uuid.UUID) (MaintenancePlan, error) {
	plan, err := r.db.MaintenancePlan.Query().Where(maintenanceplan.IDEQ(planID)).Only(ctx)
	if err != nil {
		return MaintenancePlan{}, err
	}

	nextDue := computeNextDue(completedAt, plan.IntervalValue, MaintenancePlanIntervalUnit(plan.IntervalUnit))
	updated, err := r.db.MaintenancePlan.UpdateOneID(planID).
		SetLastCompletedAt(completedAt).
		SetNextDueAt(nextDue).
		Save(ctx)
	if err != nil {
		return MaintenancePlan{}, err
	}

	openCount, err := r.db.MaintenanceEntry.Query().
		Where(
			maintenanceentry.PlanIDEQ(planID),
			maintenanceentry.ScheduledDateEQ(nextDue),
			maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{}),
			),
		).
		Count(ctx)
	if err != nil {
		return MaintenancePlan{}, err
	}

	if openCount == 0 {
		if _, err := r.db.MaintenanceEntry.Create().
			SetEntityID(itemID).
			SetPlanID(planID).
			SetName(updated.Name).
			SetDescription(updated.Description).
			SetScheduledDate(nextDue).
			SetDate(time.Time{}).
			Save(ctx); err != nil {
			return MaintenancePlan{}, err
		}
	}

	return mapMaintenancePlan(updated), nil
}

func computeNextDue(base time.Time, intervalValue int, intervalUnit MaintenancePlanIntervalUnit) time.Time {
	switch intervalUnit {
	case MaintenancePlanIntervalUnitHour:
		return base.Add(time.Duration(intervalValue) * time.Hour)
	case MaintenancePlanIntervalUnitDay:
		return base.AddDate(0, 0, intervalValue)
	case MaintenancePlanIntervalUnitWeek:
		return base.AddDate(0, 0, 7*intervalValue)
	case MaintenancePlanIntervalUnitMonth:
		return base.AddDate(0, intervalValue, 0)
	case MaintenancePlanIntervalUnitYear:
		return base.AddDate(intervalValue, 0, 0)
	default:
		return base
	}
}
