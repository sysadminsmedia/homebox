package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		_, err := tRepos.MaintEntry.Create(context.Background(), tGroup.ID, item.ID, entry)
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
		err := tRepos.MaintEntry.Delete(context.Background(), tGroup.ID, entry.ID)
		require.NoError(t, err)
	}
}

// The two maintenance endpoints must agree about a future-dated completion.
//
// `GetAllMaintenance` omitted the `DateGT(now)` clause that
// `GetMaintenanceByItemID` applies, so the same entry read as SCHEDULED from
// /v1/entities/{id}/maintenance and COMPLETED from /v1/maintenance, with the
// same status filter. A completedDate in the future is a date that has not
// happened yet.
func TestGetAllMaintenance_FutureCompletionIsScheduledNotCompleted(t *testing.T) {
	item := useEntities(t, 1)[0]

	past := MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(getPrevMonth(time.Now())),
		Name:          "Done last month",
		Description:   "Maintenance description",
		Cost:          10,
	}
	future := MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now().AddDate(0, 0, 7)),
		Name:          "Booked for next week",
		Description:   "Maintenance description",
		Cost:          10,
	}

	for _, entry := range []MaintenanceEntryCreate{past, future} {
		_, err := tRepos.MaintEntry.Create(context.Background(), tGroup.ID, item.ID, entry)
		require.NoError(t, err)
	}

	ctx := context.Background()
	completedAll, err := tRepos.MaintEntry.GetAllMaintenance(ctx, tGroup.ID, MaintenanceFilters{Status: MaintenanceFilterStatusCompleted})
	require.NoError(t, err)
	scheduledAll, err := tRepos.MaintEntry.GetAllMaintenance(ctx, tGroup.ID, MaintenanceFilters{Status: MaintenanceFilterStatusScheduled})
	require.NoError(t, err)

	completedNames := make([]string, 0, len(completedAll))
	for _, e := range completedAll {
		completedNames = append(completedNames, e.Name)
	}
	scheduledNames := make([]string, 0, len(scheduledAll))
	for _, e := range scheduledAll {
		scheduledNames = append(scheduledNames, e.Name)
	}

	assert.Contains(t, completedNames, "Done last month")
	assert.NotContains(t, completedNames, "Booked for next week",
		"a completedDate in the future is not a completion")
	assert.Contains(t, scheduledNames, "Booked for next week")

	// The property the issue is actually about: the global and per-entity endpoints
	// classify the same entry the same way. Asserting only the global side would let
	// the two drift apart again in the other direction.
	completedItem, err := tRepos.MaintEntry.GetMaintenanceByItemID(ctx, tGroup.ID, item.ID, MaintenanceFilters{Status: MaintenanceFilterStatusCompleted})
	require.NoError(t, err)
	assert.Len(t, completedItem, len(completedAll))

	for _, entry := range append(completedAll, scheduledAll...) {
		require.NoError(t, tRepos.MaintEntry.Delete(ctx, tGroup.ID, entry.ID))
	}
}
