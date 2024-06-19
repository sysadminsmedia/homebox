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
	item := useItems(t, 1)[0]

	// Create 10 maintenance entries for the item
	created := make([]MaintenanceEntryCreate, 10)

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

	for _, entry := range created {
		_, err := tRepos.MaintEntry.Create(context.Background(), item.ID, entry)
		if err != nil {
			t.Fatalf("failed to create maintenance entry: %v", err)
		}
	}

	// Get the log for the item
	log, err := tRepos.MaintEntry.GetLog(context.Background(), tGroup.ID, item.ID, MaintenanceLogQuery{
		Completed: true,
	})
	if err != nil {
		t.Fatalf("failed to get maintenance log: %v", err)
	}

	assert.Equal(t, item.ID, log.ItemID)
	assert.Len(t, log.Entries, 10)

	// Calculate the average cost
	var total float64

	for _, entry := range log.Entries {
		total += entry.Cost
	}

	assert.InDelta(t, total, log.CostTotal, .001, "total cost should be equal to the sum of all entries")
	assert.InDelta(t, total/2, log.CostAverage, 001, "average cost should be the average of the two months")

	for _, entry := range log.Entries {
		err := tRepos.MaintEntry.Delete(context.Background(), entry.ID)
		require.NoError(t, err)
	}
}
