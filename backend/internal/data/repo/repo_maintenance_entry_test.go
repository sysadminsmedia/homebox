package repo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceentry"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceplan"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

// get the previous month from the current month, accounts for errors when run
// near the beginning or end of the month/year
func getPrevMonth(now time.Time) time.Time {
	t := now.AddDate(0, -1, 0)

	// avoid infinite loop
	max := 15
	for t.Month() == now.Month() {
		t = t.AddDate(0, 0, -1)

		max--
		if max == 0 {
			panic("max exceeded")
		}
	}

	return t
}

func TestMaintenanceEntryRepository_GetLog(t *testing.T) {
	item := useEntities(t, 1)[0]

	// Create 11 maintenance entries for the item
	created := make([]MaintenanceEntryCreate, 11)

	thisMonth := time.Now()
	lastMonth := getPrevMonth(thisMonth)

	for i := 0; i < 10; i++ {
		dt := lastMonth
		if i%2 == 0 {
			dt = thisMonth
		}

		created[i] = MaintenanceEntryCreate{
			CompletedDate: types.DateFromTime(dt),
			Name:          "Maintenance",
			Description:   "Maintenance description",
			Cost:          10,
		}
	}

	// Add an entry completed in the future
	created[10] = MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now().AddDate(0, 0, 1)),
		Name:          "Maintenance",
		Description:   "Maintenance description",
		Cost:          10,
	}

	for _, entry := range created {
		_, err := tRepos.MaintEntry.Create(context.Background(), item.ID, entry)
		if err != nil {
			t.Fatalf("failed to create maintenance entry: %v", err)
		}
	}

	// Get the log for the item
	log, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, item.ID, MaintenanceFilters{Status: MaintenanceFilterStatusCompleted})
	if err != nil {
		t.Fatalf("failed to get maintenance log: %v", err)
	}

	assert.Len(t, log, 10)

	for _, entry := range log {
		err := tRepos.MaintEntry.Delete(context.Background(), entry.ID)
		require.NoError(t, err)
	}
}

func TestMaintenanceEntryRepository_GetLog_Overdue(t *testing.T) {
	item := useEntities(t, 1)[0]

	_, err := tRepos.MaintEntry.Create(context.Background(), item.ID, MaintenanceEntryCreate{
		ScheduledDate: types.DateFromTime(time.Now().AddDate(0, 0, -2)),
		Name:          "Filter replacement",
		Description:   "Overdue task",
		Cost:          0,
	})
	require.NoError(t, err)

	log, err := tRepos.MaintEntry.GetMaintenanceByItemID(
		context.Background(),
		tGroup.ID,
		item.ID,
		MaintenanceFilters{Status: MaintenanceFilterStatusOverdue},
	)
	require.NoError(t, err)
	require.Len(t, log, 1)
	assert.True(t, log[0].IsOverdue)
}

func TestMaintenanceEntryRepository_Update_ScheduledCanBeCompleted(t *testing.T) {
	item := useEntities(t, 1)[0]
	scheduledDate := time.Now().UTC().AddDate(0, 0, 2)

	created, err := tRepos.MaintEntry.Create(context.Background(), item.ID, MaintenanceEntryCreate{
		ScheduledDate: types.DateFromTime(scheduledDate),
		Name:          "Oil change",
		Description:   "Scheduled maintenance",
		Cost:          12.5,
	})
	require.NoError(t, err)

	completedDate := types.DateFromTime(time.Now().UTC())
	updated, err := tRepos.MaintEntry.Update(context.Background(), created.ID, MaintenanceEntryUpdate{
		Name:          created.Name,
		Description:   created.Description,
		Cost:          created.Cost,
		ScheduledDate: created.ScheduledDate,
		CompletedDate: completedDate,
	})
	require.NoError(t, err)
	assert.Equal(t, completedDate.Time(), updated.CompletedDate.Time())
	assert.Equal(t, created.ScheduledDate.Time(), updated.ScheduledDate.Time())
}

func TestMaintenanceEntryRepository_Update_RecurringCompletionCreatesNextEntry(t *testing.T) {
	item := useEntities(t, 1)[0]
	startDate := time.Now().UTC().AddDate(0, 0, -1)

	plan, err := tRepos.MaintEntry.CreatePlan(context.Background(), item.ID, MaintenancePlanCreate{
		Name:          "Filter replacement",
		Description:   "Recurring filter task",
		IntervalValue: 1,
		IntervalUnit:  MaintenancePlanIntervalUnitMonth,
		StartDate:     startDate,
		Active:        true,
	})
	require.NoError(t, err)

	initialEntry, err := tRepos.MaintEntry.db.MaintenanceEntry.Query().
		Where(
			maintenanceentry.EntityID(item.ID),
			maintenanceentry.PlanIDEQ(plan.ID),
		).
		Only(context.Background())
	require.NoError(t, err)

	completedAt := types.DateFromTime(time.Now().UTC())
	_, err = tRepos.MaintEntry.Update(context.Background(), initialEntry.ID, MaintenanceEntryUpdate{
		Name:          initialEntry.Name,
		Description:   initialEntry.Description,
		Cost:          initialEntry.Cost,
		ScheduledDate: types.DateFromTime(initialEntry.ScheduledDate),
		CompletedDate: completedAt,
		PlanID:        plan.ID,
	})
	require.NoError(t, err)

	expectedNextDue := computeNextDue(completedAt.Time(), plan.IntervalValue, plan.IntervalUnit)

	refreshedPlan, err := tRepos.MaintEntry.db.MaintenancePlan.Query().
		Where(maintenanceplan.IDEQ(plan.ID)).
		Only(context.Background())
	require.NoError(t, err)
	require.NotNil(t, refreshedPlan.NextDueAt)
	require.NotNil(t, refreshedPlan.LastCompletedAt)
	assert.True(t, expectedNextDue.Equal(*refreshedPlan.NextDueAt))
	assert.True(t, completedAt.Time().Equal(*refreshedPlan.LastCompletedAt))

	openEntries, err := tRepos.MaintEntry.db.MaintenanceEntry.Query().
		Where(
			maintenanceentry.EntityID(item.ID),
			maintenanceentry.PlanIDEQ(plan.ID),
			maintenanceentry.ScheduledDateEQ(expectedNextDue),
			maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{}),
			),
		).
		All(context.Background())
	require.NoError(t, err)
	assert.Len(t, openEntries, 1)

	assert.NotEqual(t, uuid.Nil, openEntries[0].ID)
	assert.NotEqual(t, initialEntry.ID, openEntries[0].ID)
}

func TestMaintenanceEntryRepository_CreatePlan_UsesStartDateAsFirstDueDate(t *testing.T) {
	item := useEntities(t, 1)[0]
	startDate := time.Date(2026, time.March, 10, 9, 30, 0, 0, time.UTC)

	plan, err := tRepos.MaintEntry.CreatePlan(context.Background(), item.ID, MaintenancePlanCreate{
		Name:          "Weekly maintenance",
		Description:   "Recurring task with explicit scheduled date",
		IntervalValue: 1,
		IntervalUnit:  MaintenancePlanIntervalUnitWeek,
		StartDate:     startDate,
		Active:        true,
	})
	require.NoError(t, err)

	assert.Equal(t, startDate, plan.NextDueAt)

	openEntries, err := tRepos.MaintEntry.db.MaintenanceEntry.Query().
		Where(
			maintenanceentry.EntityID(item.ID),
			maintenanceentry.PlanIDEQ(plan.ID),
			maintenanceentry.ScheduledDateEQ(startDate),
			maintenanceentry.Or(
				maintenanceentry.DateIsNil(),
				maintenanceentry.DateEQ(time.Time{}),
			),
		).
		All(context.Background())
	require.NoError(t, err)
	require.Len(t, openEntries, 1)
}
