package repo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

// TestWipeInventory_Integration tests the complete wipe inventory flow
func TestWipeInventory_Integration(t *testing.T) {
	// Create test data: locations, tags, items with maintenance

	// 1. Create locations
	loc1, err := tRepos.Locations.Create(context.Background(), tGroup.ID, LocationCreate{
		Name:        "Test Garage",
		Description: "Garage location",
	})
	require.NoError(t, err)

	loc2, err := tRepos.Locations.Create(context.Background(), tGroup.ID, LocationCreate{
		Name:        "Test Basement",
		Description: "Basement location",
	})
	require.NoError(t, err)

	// 2. Create tags
	tag1, err := tRepos.Tags.Create(context.Background(), tGroup.ID, TagCreate{
		Name:        "Test Electronics",
		Description: "Electronics tag",
	})
	require.NoError(t, err)

	tag2, err := tRepos.Tags.Create(context.Background(), tGroup.ID, TagCreate{
		Name:        "Test Tools",
		Description: "Tools tag",
	})
	require.NoError(t, err)

	// 3. Create items
	item1, err := tRepos.Items.Create(context.Background(), tGroup.ID, ItemCreate{
		Name:        "Test Laptop",
		Description: "Work laptop",
		LocationID:  loc1.ID,
		TagIDs:      []uuid.UUID{tag1.ID},
	})
	require.NoError(t, err)

	item2, err := tRepos.Items.Create(context.Background(), tGroup.ID, ItemCreate{
		Name:        "Test Drill",
		Description: "Power drill",
		LocationID:  loc2.ID,
		TagIDs:      []uuid.UUID{tag2.ID},
	})
	require.NoError(t, err)

	item3, err := tRepos.Items.Create(context.Background(), tGroup.ID, ItemCreate{
		Name:        "Test Monitor",
		Description: "Computer monitor",
		LocationID:  loc1.ID,
		TagIDs:      []uuid.UUID{tag1.ID},
	})
	require.NoError(t, err)

	// 4. Create maintenance entries
	_, err = tRepos.MaintEntry.Create(context.Background(), item1.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Laptop cleaning",
		Description:   "Cleaned keyboard and screen",
		Cost:          0,
	})
	require.NoError(t, err)

	_, err = tRepos.MaintEntry.Create(context.Background(), item2.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Drill maintenance",
		Description:   "Oiled motor",
		Cost:          5.00,
	})
	require.NoError(t, err)

	_, err = tRepos.MaintEntry.Create(context.Background(), item3.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Monitor calibration",
		Description:   "Color calibration",
		Cost:          0,
	})
	require.NoError(t, err)

	// 5. Verify items exist
	allItems, err := tRepos.Items.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(allItems), 3, "Should have at least 3 items")

	// 6. Verify maintenance entries exist
	maint1List, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, item1.ID, MaintenanceFilters{})
	require.NoError(t, err)
	assert.NotEmpty(t, maint1List, "Item 1 should have maintenance records")

	maint2List, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, item2.ID, MaintenanceFilters{})
	require.NoError(t, err)
	assert.NotEmpty(t, maint2List, "Item 2 should have maintenance records")

	// 7. Test wipe inventory with all options enabled
	deleted, err := tRepos.Items.WipeInventory(context.Background(), tGroup.ID, true, true, true)
	require.NoError(t, err)
	assert.Positive(t, deleted, "Should have deleted entities")

	// 8. Verify all items are deleted
	allItemsAfter, err := tRepos.Items.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)
	assert.Empty(t, allItemsAfter, "All items should be deleted")

	// 9. Verify maintenance entries are deleted
	maint1After, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, item1.ID, MaintenanceFilters{})
	require.NoError(t, err)
	assert.Empty(t, maint1After, "Item 1 maintenance records should be deleted")

	// 10. Verify tags are deleted
	_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tag1.ID)
	require.Error(t, err, "Tag 1 should be deleted")

	_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tag2.ID)
	require.Error(t, err, "Tag 2 should be deleted")
	// 11. Verify locations are deleted
	_, err = tRepos.Locations.Get(context.Background(), loc1.ID)
	require.Error(t, err, "Location 1 should be deleted")

	_, err = tRepos.Locations.Get(context.Background(), loc2.ID)
	require.Error(t, err, "Location 2 should be deleted")
}

// TestWipeInventory_SelectiveWipe tests wiping only certain entity types
func TestWipeInventory_SelectiveWipe(t *testing.T) {
	// Create test data
	loc, err := tRepos.Locations.Create(context.Background(), tGroup.ID, LocationCreate{
		Name:        "Test Office",
		Description: "Office location",
	})
	require.NoError(t, err)

	tag, err := tRepos.Tags.Create(context.Background(), tGroup.ID, TagCreate{
		Name:        "Test Important",
		Description: "Important tag",
	})
	require.NoError(t, err)

	item, err := tRepos.Items.Create(context.Background(), tGroup.ID, ItemCreate{
		Name:        "Test Computer",
		Description: "Desktop computer",
		LocationID:  loc.ID,
		TagIDs:      []uuid.UUID{tag.ID},
	})
	require.NoError(t, err)

	_, err = tRepos.MaintEntry.Create(context.Background(), item.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "System update",
		Description:   "OS update",
		Cost:          0,
	})
	require.NoError(t, err)

	// Test: Wipe only items (keep tags and locations)
	deleted, err := tRepos.Items.WipeInventory(context.Background(), tGroup.ID, false, false, false)
	require.NoError(t, err)
	assert.Positive(t, deleted, "Should have deleted at least items")

	// Verify item is deleted
	_, err = tRepos.Items.GetOneByGroup(context.Background(), tGroup.ID, item.ID)
	require.Error(t, err, "Item should be deleted")

	// Verify maintenance is cascade deleted
	maintList, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, item.ID, MaintenanceFilters{})
	require.NoError(t, err)
	assert.Empty(t, maintList, "Maintenance should be cascade deleted")

	// Verify tag still exists
	_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tag.ID)
	require.NoError(t, err, "Tag should still exist")
	// Verify location still exists
	_, err = tRepos.Locations.Get(context.Background(), loc.ID)
	require.NoError(t, err, "Location should still exist")

	// Cleanup
	_ = tRepos.Tags.DeleteByGroup(context.Background(), tGroup.ID, tag.ID)
	_ = tRepos.Locations.delete(context.Background(), loc.ID)
}
