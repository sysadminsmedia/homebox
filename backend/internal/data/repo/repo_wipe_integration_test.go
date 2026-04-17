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
	containerET := useContainerEntityType(t)
	itemET := useItemEntityType(t)

	// 1. Create containers
	c1f := EntityCreate{
		Name:         "Test Garage",
		Description:  "Garage container",
		EntityTypeID: containerET.ID,
	}
	container1, err := tRepos.Entities.Create(context.Background(), tGroup.ID, c1f)
	require.NoError(t, err)

	c2f := EntityCreate{
		Name:         "Test Basement",
		Description:  "Basement container",
		EntityTypeID: containerET.ID,
	}
	container2, err := tRepos.Entities.Create(context.Background(), tGroup.ID, c2f)
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
	entity1, err := tRepos.Entities.Create(context.Background(), tGroup.ID, EntityCreate{
		Name:         "Test Laptop",
		Description:  "Work laptop",
		ParentID:     container1.ID,
		EntityTypeID: itemET.ID,
		TagIDs:       []uuid.UUID{tag1.ID},
	})
	require.NoError(t, err)

	entity2, err := tRepos.Entities.Create(context.Background(), tGroup.ID, EntityCreate{
		Name:         "Test Drill",
		Description:  "Power drill",
		ParentID:     container2.ID,
		EntityTypeID: itemET.ID,
		TagIDs:       []uuid.UUID{tag2.ID},
	})
	require.NoError(t, err)

	entity3, err := tRepos.Entities.Create(context.Background(), tGroup.ID, EntityCreate{
		Name:         "Test Monitor",
		Description:  "Computer monitor",
		ParentID:     container1.ID,
		EntityTypeID: itemET.ID,
		TagIDs:       []uuid.UUID{tag1.ID},
	})
	require.NoError(t, err)

	// 4. Create maintenance entries
	_, err = tRepos.MaintEntry.Create(context.Background(), entity1.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Laptop cleaning",
		Description:   "Cleaned keyboard and screen",
		Cost:          0,
	})
	require.NoError(t, err)

	_, err = tRepos.MaintEntry.Create(context.Background(), entity2.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Drill maintenance",
		Description:   "Oiled motor",
		Cost:          5.00,
	})
	require.NoError(t, err)

	_, err = tRepos.MaintEntry.Create(context.Background(), entity3.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Monitor calibration",
		Description:   "Color calibration",
		Cost:          0,
	})
	require.NoError(t, err)

	// 5. Verify entities exist
	allEntities, err := tRepos.Entities.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(allEntities), 3, "Should have at least 3 entities")

	// 6. Verify maintenance entries exist
	maint1List, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, entity1.ID, MaintenanceFilters{})
	require.NoError(t, err)
	assert.NotEmpty(t, maint1List, "Entity 1 should have maintenance records")

	maint2List, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, entity2.ID, MaintenanceFilters{})
	require.NoError(t, err)
	assert.NotEmpty(t, maint2List, "Entity 2 should have maintenance records")

	// 7. Test wipe inventory with all options enabled
	deleted, err := tRepos.Entities.WipeInventory(context.Background(), tGroup.ID, true, true, true)
	require.NoError(t, err)
	assert.Positive(t, deleted, "Should have deleted entities")

	// 8. Verify all entities are deleted
	allEntitiesAfter, err := tRepos.Entities.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)
	assert.Empty(t, allEntitiesAfter, "All entities should be deleted")

	// 9. Verify maintenance entries are deleted
	maint1After, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, entity1.ID, MaintenanceFilters{})
	require.NoError(t, err)
	assert.Empty(t, maint1After, "Entity 1 maintenance records should be deleted")

	// 10. Verify tags are deleted
	_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tag1.ID)
	require.Error(t, err, "Tag 1 should be deleted")

	_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tag2.ID)
	require.Error(t, err, "Tag 2 should be deleted")
}

// TestWipeInventory_SelectiveWipe tests wiping only certain entity types
func TestWipeInventory_SelectiveWipe(t *testing.T) {
	containerET := useContainerEntityType(t)
	itemET := useItemEntityType(t)

	// Create test data
	container, err := tRepos.Entities.Create(context.Background(), tGroup.ID, EntityCreate{
		Name:         "Test Office",
		Description:  "Office container",
		EntityTypeID: containerET.ID,
	})
	require.NoError(t, err)

	tagObj, err := tRepos.Tags.Create(context.Background(), tGroup.ID, TagCreate{
		Name:        "Test Important",
		Description: "Important tag",
	})
	require.NoError(t, err)

	e, err := tRepos.Entities.Create(context.Background(), tGroup.ID, EntityCreate{
		Name:         "Test Computer",
		Description:  "Desktop computer",
		ParentID:     container.ID,
		EntityTypeID: itemET.ID,
		TagIDs:       []uuid.UUID{tagObj.ID},
	})
	require.NoError(t, err)

	_, err = tRepos.MaintEntry.Create(context.Background(), e.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "System update",
		Description:   "OS update",
		Cost:          0,
	})
	require.NoError(t, err)

	// Test: Wipe only items (keep tags and containers)
	deleted, err := tRepos.Entities.WipeInventory(context.Background(), tGroup.ID, false, false, false)
	require.NoError(t, err)
	assert.Positive(t, deleted, "Should have deleted at least entities")

	// Verify entity is deleted
	_, err = tRepos.Entities.GetOneByGroup(context.Background(), tGroup.ID, e.ID)
	require.Error(t, err, "Entity should be deleted")

	// Verify maintenance is cascade deleted
	maintList, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, e.ID, MaintenanceFilters{})
	require.NoError(t, err)
	assert.Empty(t, maintList, "Maintenance should be cascade deleted")

	// Verify tag still exists
	_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tagObj.ID)
	require.NoError(t, err, "Tag should still exist")

	// Cleanup
	_ = tRepos.Tags.DeleteByGroup(context.Background(), tGroup.ID, tagObj.ID)
}
