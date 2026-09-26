package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceplan"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
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
	Name                string                      `json:"name"                          validate:"required"`
	Description         string                      `json:"description"`
	IntervalValue       int                         `json:"intervalValue"                 validate:"required,min=1"`
	IntervalUnit        MaintenancePlanIntervalUnit `json:"intervalUnit"                  validate:"required"`
	StartDate           types.Date                  `json:"startDate"`
	Active              bool                        `json:"active"`
	LinkExistingEntryID *uuid.UUID                  `json:"linkExistingEntryID,omitempty"`
}

type MaintenancePlanUpdate struct {
	Name          string                      `json:"name"`
	Description   string                      `json:"description"`
	IntervalValue int                         `json:"intervalValue"`
	IntervalUnit  MaintenancePlanIntervalUnit `json:"intervalUnit"`
	NextDueAt     *types.Date                 `json:"nextDueAt,omitempty"`
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

// ownedEntity reports whether itemID belongs to the given group.
func ownedEntity(ctx context.Context, db *ent.Client, gid, itemID uuid.UUID) (bool, error) {
	return db.Entity.Query().
		Where(entity.ID(itemID), entity.HasGroupWith(group.ID(gid))).
		Exist(ctx)
}

// ownedPlan returns the plan only if it belongs to an entity in the given group.
func ownedPlan(ctx context.Context, db *ent.Client, gid, planID uuid.UUID) (*ent.MaintenancePlan, error) {
	return db.MaintenancePlan.Query().
		Where(
			maintenanceplan.ID(planID),
			maintenanceplan.HasEntityWith(entity.HasGroupWith(group.ID(gid))),
		).
		Only(ctx)
}

// withTx runs fn inside a transaction, committing on success and rolling
// back on error.
func (r *MaintenanceEntryRepository) withTx(ctx context.Context, fn func(db *ent.Client) error) error {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx.Client()); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w (rollback failed: %w)", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func (r *MaintenanceEntryRepository) CreatePlan(ctx context.Context, gid, itemID uuid.UUID, input MaintenancePlanCreate) (MaintenancePlan, error) {
	owned, err := ownedEntity(ctx, r.db, gid, itemID)
	if err != nil {
		return MaintenancePlan{}, err
	}
	if !owned {
		return MaintenancePlan{}, &ent.NotFoundError{}
	}

	base := input.StartDate.Time()
	if base.IsZero() {
		base = time.Now().UTC()
	}
	firstDue := types.DateFromTime(base.UTC()).Time()

	var plan *ent.MaintenancePlan
	err = r.withTx(ctx, func(db *ent.Client) error {
		var err error
		plan, err = db.MaintenancePlan.Create().
			SetEntityID(itemID).
			SetName(input.Name).
			SetDescription(input.Description).
			SetIntervalValue(input.IntervalValue).
			SetIntervalUnit(maintenanceplan.IntervalUnit(input.IntervalUnit)).
			SetActive(input.Active).
			SetNextDueAt(firstDue).
			Save(ctx)
		if err != nil {
			return err
		}

		if input.LinkExistingEntryID != nil && *input.LinkExistingEntryID != uuid.Nil {
			// Only entries on the same entity can be linked; the entity is
			// already known to belong to the caller's group.
			existing, err := db.MaintenanceEntry.Query().
				Where(
					maintenanceentry.IDEQ(*input.LinkExistingEntryID),
					maintenanceentry.EntityIDEQ(itemID),
				).
				Only(ctx)
			if err != nil {
				return fmt.Errorf("link existing maintenance entry: %w", err)
			}
			_, err = db.MaintenanceEntry.UpdateOneID(existing.ID).
				SetPlanID(plan.ID).
				SetScheduledDate(firstDue).
				Save(ctx)
			return err
		}

		_, err = db.MaintenanceEntry.Create().
			SetEntityID(itemID).
			SetPlanID(plan.ID).
			SetName(plan.Name).
			SetDescription(plan.Description).
			SetScheduledDate(firstDue).
			SetDate(time.Time{}).
			Save(ctx)
		return err
	})
	if err != nil {
		return MaintenancePlan{}, err
	}

	return mapMaintenancePlan(plan), nil
}

func (r *MaintenanceEntryRepository) UpdatePlan(ctx context.Context, gid, planID uuid.UUID, input MaintenancePlanUpdate) (MaintenancePlan, error) {
	if _, err := ownedPlan(ctx, r.db, gid, planID); err != nil {
		return MaintenancePlan{}, err
	}

	up := r.db.MaintenancePlan.UpdateOneID(planID).
		SetName(input.Name).
		SetDescription(input.Description).
		SetIntervalValue(input.IntervalValue).
		SetIntervalUnit(maintenanceplan.IntervalUnit(input.IntervalUnit)).
		SetActive(input.Active)
	if input.NextDueAt != nil {
		t := input.NextDueAt.Time()
		if t.IsZero() {
			up = up.ClearNextDueAt()
		} else {
			up = up.SetNextDueAt(t.UTC())
		}
	}

	plan, err := up.Save(ctx)
	if err != nil {
		return MaintenancePlan{}, err
	}

	return mapMaintenancePlan(plan), nil
}

func (r *MaintenanceEntryRepository) DeletePlan(ctx context.Context, gid, planID uuid.UUID) error {
	n, err := r.db.MaintenancePlan.Delete().
		Where(
			maintenanceplan.ID(planID),
			maintenanceplan.HasEntityWith(entity.HasGroupWith(group.ID(gid))),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return &ent.NotFoundError{}
	}

	return nil
}

// validatePlanForEntity ensures planID refers to a plan on the given entity.
// Entries may only be attached to plans on their own entity; this also keeps
// plans from other groups out of reach.
func validatePlanForEntity(ctx context.Context, db *ent.Client, planID, entityID uuid.UUID) error {
	exists, err := db.MaintenancePlan.Query().
		Where(maintenanceplan.ID(planID), maintenanceplan.EntityIDEQ(entityID)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return &ent.NotFoundError{}
	}

	return nil
}

// rollPlanFromCompletion advances a plan after one of its entries is completed
// and creates the next open entry. The caller must have verified that the plan
// belongs to entityID.
func rollPlanFromCompletion(ctx context.Context, db *ent.Client, planID uuid.UUID, completedAt time.Time, entityID uuid.UUID) error {
	plan, err := db.MaintenancePlan.Query().
		Where(maintenanceplan.ID(planID), maintenanceplan.EntityIDEQ(entityID)).
		Only(ctx)
	if err != nil {
		return err
	}

	if !plan.Active {
		// Paused plans record the completion but do not schedule a follow-up.
		return db.MaintenancePlan.UpdateOneID(planID).
			SetLastCompletedAt(completedAt).
			Exec(ctx)
	}

	nextDue := computeNextDue(completedAt, plan.IntervalValue, MaintenancePlanIntervalUnit(plan.IntervalUnit))
	updated, err := db.MaintenancePlan.UpdateOneID(planID).
		SetLastCompletedAt(completedAt).
		SetNextDueAt(nextDue).
		Save(ctx)
	if err != nil {
		return err
	}

	openCount, err := db.MaintenanceEntry.Query().
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
		return err
	}

	if openCount > 0 {
		return nil
	}

	return db.MaintenanceEntry.Create().
		SetEntityID(entityID).
		SetPlanID(planID).
		SetName(updated.Name).
		SetDescription(updated.Description).
		SetScheduledDate(nextDue).
		SetDate(time.Time{}).
		Exec(ctx)
}

// addMonthsClamped adds n months to t, clamping the day to the last day of the
// target month so that e.g. Jan 31 + 1 month is Feb 28/29 rather than Mar 3.
func addMonthsClamped(t time.Time, n int) time.Time {
	y, m, d := t.Date()
	firstOfTarget := time.Date(y, m+time.Month(n), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	lastDay := firstOfTarget.AddDate(0, 1, -1).Day()
	if d > lastDay {
		d = lastDay
	}

	return time.Date(firstOfTarget.Year(), firstOfTarget.Month(), d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
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
		return addMonthsClamped(base, intervalValue)
	case MaintenancePlanIntervalUnitYear:
		return addMonthsClamped(base, 12*intervalValue)
	default:
		return base
	}
}
