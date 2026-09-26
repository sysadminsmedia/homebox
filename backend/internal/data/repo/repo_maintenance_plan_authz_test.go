package repo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/maintenanceplan"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

func useMaintenancePlan(t *testing.T, itemID uuid.UUID, start time.Time) MaintenancePlan {
	t.Helper()

	plan, err := tRepos.MaintEntry.CreatePlan(context.Background(), tGroup.ID, itemID, MaintenancePlanCreate{
		Name:          "Furnace filter",
		IntervalValue: 90,
		IntervalUnit:  MaintenancePlanIntervalUnitDay,
		StartDate:     types.DateFromTime(start),
		Active:        true,
	})
	require.NoError(t, err)

	return plan
}

func useForeignGroup(t *testing.T) Group {
	t.Helper()

	g, err := tRepos.Groups.GroupCreate(context.Background(), "maint-foreign-"+uuid.NewString(), uuid.Nil)
	require.NoError(t, err)

	return g
}

func TestMaintenancePlan_CreateRejectsForeignGroup(t *testing.T) {
	item := useEntities(t, 1)[0]
	foreign := useForeignGroup(t)

	_, err := tRepos.MaintEntry.CreatePlan(context.Background(), foreign.ID, item.ID, MaintenancePlanCreate{
		Name:          "Injected",
		IntervalValue: 1,
		IntervalUnit:  MaintenancePlanIntervalUnitDay,
		Active:        true,
	})
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	count, err := tRepos.MaintEntry.db.MaintenancePlan.Query().
		Where(maintenanceplan.EntityIDEQ(item.ID)).
		Count(context.Background())
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestMaintenancePlan_UpdateRejectsForeignGroup(t *testing.T) {
	item := useEntities(t, 1)[0]
	plan := useMaintenancePlan(t, item.ID, time.Now().UTC())
	foreign := useForeignGroup(t)

	_, err := tRepos.MaintEntry.UpdatePlan(context.Background(), foreign.ID, plan.ID, MaintenancePlanUpdate{
		Name:          "Hijacked",
		IntervalValue: 1,
		IntervalUnit:  MaintenancePlanIntervalUnitDay,
	})
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	stored, err := tRepos.MaintEntry.db.MaintenancePlan.Get(context.Background(), plan.ID)
	require.NoError(t, err)
	assert.Equal(t, "Furnace filter", stored.Name)
	assert.True(t, stored.Active)
}

func TestMaintenancePlan_DeleteRejectsForeignGroup(t *testing.T) {
	item := useEntities(t, 1)[0]
	plan := useMaintenancePlan(t, item.ID, time.Now().UTC())
	foreign := useForeignGroup(t)

	err := tRepos.MaintEntry.DeletePlan(context.Background(), foreign.ID, plan.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	exists, err := tRepos.MaintEntry.db.MaintenancePlan.Query().
		Where(maintenanceplan.ID(plan.ID)).
		Exist(context.Background())
	require.NoError(t, err)
	assert.True(t, exists)

	require.NoError(t, tRepos.MaintEntry.DeletePlan(context.Background(), tGroup.ID, plan.ID))
}

func TestMaintenanceEntry_CannotAttachPlanFromAnotherEntity(t *testing.T) {
	items := useEntities(t, 2)
	plan := useMaintenancePlan(t, items[0].ID, time.Now().UTC())

	_, err := tRepos.MaintEntry.Create(context.Background(), tGroup.ID, items[1].ID, MaintenanceEntryCreate{
		Name:          "Borrowed plan",
		CompletedDate: types.DateFromTime(time.Now().UTC()),
		PlanID:        plan.ID,
	})
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	entry, err := tRepos.MaintEntry.Create(context.Background(), tGroup.ID, items[1].ID, MaintenanceEntryCreate{
		Name:          "Standalone",
		ScheduledDate: types.DateFromTime(time.Now().UTC().AddDate(0, 0, 1)),
	})
	require.NoError(t, err)

	// Completing an entry with a foreign plan must not roll that plan forward.
	_, err = tRepos.MaintEntry.Update(context.Background(), tGroup.ID, entry.ID, MaintenanceEntryUpdate{
		Name:          entry.Name,
		ScheduledDate: entry.ScheduledDate,
		CompletedDate: types.DateFromTime(time.Now().UTC()),
		PlanID:        plan.ID,
	})
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	stored, err := tRepos.MaintEntry.db.MaintenancePlan.Get(context.Background(), plan.ID)
	require.NoError(t, err)
	assert.Nil(t, stored.LastCompletedAt)
}

func TestMaintenanceEntry_GetOverdueRecurring(t *testing.T) {
	item := useEntities(t, 1)[0]
	today := types.DateFromTime(time.Now().UTC())

	overduePlan := useMaintenancePlan(t, item.ID, today.Time().AddDate(0, 0, -5))
	_ = useMaintenancePlan(t, item.ID, today.Time().AddDate(0, 0, 5))

	// One-off overdue entry: must not be included.
	_, err := tRepos.MaintEntry.Create(context.Background(), tGroup.ID, item.ID, MaintenanceEntryCreate{
		Name:          "One-off",
		ScheduledDate: types.DateFromTime(today.Time().AddDate(0, 0, -3)),
	})
	require.NoError(t, err)

	overdue, err := tRepos.MaintEntry.GetOverdueRecurring(context.Background(), tGroup.ID, today)
	require.NoError(t, err)

	var mine []MaintenanceEntry
	for _, e := range overdue {
		if e.PlanID == overduePlan.ID {
			mine = append(mine, e)
		}
		assert.NotEqual(t, "One-off", e.Name)
	}
	require.Len(t, mine, 1)

	// Paused plans stop nagging.
	_, err = tRepos.MaintEntry.UpdatePlan(context.Background(), tGroup.ID, overduePlan.ID, MaintenancePlanUpdate{
		Name:          overduePlan.Name,
		IntervalValue: overduePlan.IntervalValue,
		IntervalUnit:  overduePlan.IntervalUnit,
		Active:        false,
	})
	require.NoError(t, err)

	overdue, err = tRepos.MaintEntry.GetOverdueRecurring(context.Background(), tGroup.ID, today)
	require.NoError(t, err)
	for _, e := range overdue {
		assert.NotEqual(t, overduePlan.ID, e.PlanID)
	}

	// Other groups see nothing.
	foreign := useForeignGroup(t)
	overdue, err = tRepos.MaintEntry.GetOverdueRecurring(context.Background(), foreign.ID, today)
	require.NoError(t, err)
	assert.Empty(t, overdue)
}
