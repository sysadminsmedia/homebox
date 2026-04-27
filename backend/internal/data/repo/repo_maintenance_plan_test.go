package repo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComputeNextDue(t *testing.T) {
	base := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)

	assert.Equal(t, base.Add(2*time.Hour), computeNextDue(base, 2, MaintenancePlanIntervalUnitHour))
	assert.Equal(t, base.AddDate(0, 0, 7), computeNextDue(base, 7, MaintenancePlanIntervalUnitDay))
	assert.Equal(t, base.AddDate(0, 0, 14), computeNextDue(base, 2, MaintenancePlanIntervalUnitWeek))
	assert.Equal(t, base.AddDate(0, 1, 0), computeNextDue(base, 1, MaintenancePlanIntervalUnitMonth))
	assert.Equal(t, base.AddDate(1, 0, 0), computeNextDue(base, 1, MaintenancePlanIntervalUnitYear))
}
