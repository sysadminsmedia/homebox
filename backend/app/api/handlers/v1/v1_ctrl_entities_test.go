package v1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

func TestEntitiesTotalPrice(t *testing.T) {
	items := []repo.EntitySummary{
		{PurchasePrice: 1, Quantity: 2},
		{PurchasePrice: 0.1, Quantity: 3},
		{PurchasePrice: 10, Quantity: 1.5},
		{PurchasePrice: 100, Quantity: 4, SoldDate: types.DateFromTime(time.Now())},
	}

	// 1*2 + 0.1*3 + 10*1.5 = 17.30; the sold item is excluded.
	assert.InDelta(t, 17.30, entitiesTotalPrice(items), 0.000001)
	assert.InDelta(t, 0.0, entitiesTotalPrice(nil), 0.000001)
}
