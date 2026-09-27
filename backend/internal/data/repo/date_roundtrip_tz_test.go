package repo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

// unmarshalDate parses a date the way the API does when the browser PUTs it.
func unmarshalDate(t *testing.T, yyyymmdd string) types.Date {
	t.Helper()

	var body struct {
		Date types.Date `json:"date"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"date":"`+yyyymmdd+`"}`), &body))
	return body.Date
}

// Dates entered in the UI must survive the round-trip unchanged, and must keep
// surviving it when the item is saved again without touching the picker. Run
// under TZ=America/Chicago (or any zone west of Greenwich) to cover the
// timezone shift that made purchase dates drift a day earlier per save.
func TestEntityDatesRoundTripUnchanged(t *testing.T) {
	ctx := context.Background()
	itemET := useItemEntityType(t)

	cf := containerFactory()
	cf.EntityTypeID = useContainerEntityType(t).ID
	container, err := tRepos.Entities.Create(ctx, tGroup.ID, cf)
	require.NoError(t, err)

	ef := entityFactory()
	ef.EntityTypeID = itemET.ID
	ef.ParentID = container.ID
	e, err := tRepos.Entities.Create(ctx, tGroup.ID, ef)
	require.NoError(t, err)

	const (
		purchase = "2026-08-27"
		warranty = "2030-01-01"
		sold     = "2026-12-31"
	)

	upd := EntityUpdate{
		ID:              e.ID,
		Name:            e.Name,
		EntityTypeID:    itemET.ID,
		ParentID:        container.ID,
		Quantity:        1,
		PurchaseDate:    unmarshalDate(t, purchase),
		WarrantyExpires: unmarshalDate(t, warranty),
		SoldDate:        unmarshalDate(t, sold),
	}

	// Three saves: the first is the user's edit, the next two echo back
	// exactly what the API served, as the edit form does for untouched fields.
	for save := 1; save <= 3; save++ {
		out, err := tRepos.Entities.UpdateByGroup(ctx, tGroup.ID, upd)
		require.NoError(t, err)

		require.Equal(t, purchase, out.PurchaseDate.String(), "purchase date after save %d", save)
		require.Equal(t, warranty, out.WarrantyExpires.String(), "warranty expiry after save %d", save)
		require.Equal(t, sold, out.SoldDate.String(), "sold date after save %d", save)

		upd.PurchaseDate = out.PurchaseDate
		upd.WarrantyExpires = out.WarrantyExpires
		upd.SoldDate = out.SoldDate
	}

	got, err := tRepos.Entities.GetOne(ctx, e.ID)
	require.NoError(t, err)
	require.Equal(t, purchase, got.PurchaseDate.String())
	require.Equal(t, warranty, got.WarrantyExpires.String())
	require.Equal(t, sold, got.SoldDate.String())
}

// ZeroOutTimeFields rewrites every date column in the group. It must not move
// the day while it strips the time.
func TestZeroOutTimeFieldsKeepsTheDay(t *testing.T) {
	ctx := context.Background()
	itemET := useItemEntityType(t)

	cf := containerFactory()
	cf.EntityTypeID = useContainerEntityType(t).ID
	container, err := tRepos.Entities.Create(ctx, tGroup.ID, cf)
	require.NoError(t, err)

	ef := entityFactory()
	ef.EntityTypeID = itemET.ID
	ef.ParentID = container.ID
	e, err := tRepos.Entities.Create(ctx, tGroup.ID, ef)
	require.NoError(t, err)

	const purchase = "2026-08-27"

	_, err = tRepos.Entities.UpdateByGroup(ctx, tGroup.ID, EntityUpdate{
		ID:           e.ID,
		Name:         e.Name,
		EntityTypeID: itemET.ID,
		ParentID:     container.ID,
		Quantity:     1,
		PurchaseDate: unmarshalDate(t, purchase),
	})
	require.NoError(t, err)

	_, err = tRepos.Entities.ZeroOutTimeFields(ctx, tGroup.ID)
	require.NoError(t, err)

	got, err := tRepos.Entities.GetOne(ctx, e.ID)
	require.NoError(t, err)
	require.Equal(t, purchase, got.PurchaseDate.String())
}

// Maintenance entries map their dates with a plain type conversion, which is
// exactly where a driver-supplied local zone used to leak through.
func TestMaintenanceEntryDatesRoundTripUnchanged(t *testing.T) {
	ctx := context.Background()
	item := useEntities(t, 1)[0]

	const (
		completed = "2026-08-27"
		scheduled = "2026-09-15"
	)

	created, err := tRepos.MaintEntry.Create(ctx, tGroup.ID, item.ID, MaintenanceEntryCreate{
		CompletedDate: unmarshalDate(t, completed),
		ScheduledDate: unmarshalDate(t, scheduled),
		Name:          "Oil change",
		Description:   "Maintenance description",
		Cost:          10,
	})
	require.NoError(t, err)
	require.Equal(t, completed, created.CompletedDate.String())
	require.Equal(t, scheduled, created.ScheduledDate.String())

	entries, err := tRepos.MaintEntry.GetMaintenanceByItemID(ctx, tGroup.ID, item.ID, MaintenanceFilters{})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, completed, entries[0].CompletedDate.String())
	require.Equal(t, scheduled, entries[0].ScheduledDate.String())
}
