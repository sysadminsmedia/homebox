package repo

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

func itemFactory() ItemCreate {
	return ItemCreate{
		Name:        fk.Str(10),
		Description: fk.Str(100),
	}
}

func useItems(t *testing.T, len int) []ItemOut {
	t.Helper()

	location, err := tRepos.Locations.Create(context.Background(), tGroup.ID, locationFactory())
	require.NoError(t, err)

	items := make([]ItemOut, len)
	for i := 0; i < len; i++ {
		itm := itemFactory()
		itm.LocationID = location.ID

		item, err := tRepos.Items.Create(context.Background(), tGroup.ID, itm)
		require.NoError(t, err)
		items[i] = item
	}

	t.Cleanup(func() {
		for _, item := range items {
			_ = tRepos.Items.Delete(context.Background(), item.ID)
		}

		_ = tRepos.Locations.delete(context.Background(), location.ID)
	})

	return items
}

func TestItemsRepository_RecursiveRelationships(t *testing.T) {
	parent := useItems(t, 1)[0]

	children := useItems(t, 3)

	for _, child := range children {
		update := ItemUpdate{
			ID:          child.ID,
			ParentID:    parent.ID,
			Name:        "note-important",
			Description: "This is a note",
			LocationID:  child.Location.ID,
		}

		// Append Parent ID
		_, err := tRepos.Items.UpdateByGroup(context.Background(), tGroup.ID, update)
		require.NoError(t, err)

		// Check Parent ID
		updated, err := tRepos.Items.GetOne(context.Background(), child.ID)
		require.NoError(t, err)
		assert.Equal(t, parent.ID, updated.Parent.ID)

		// Remove Parent ID
		update.ParentID = uuid.Nil
		_, err = tRepos.Items.UpdateByGroup(context.Background(), tGroup.ID, update)
		require.NoError(t, err)

		// Check Parent ID
		updated, err = tRepos.Items.GetOne(context.Background(), child.ID)
		require.NoError(t, err)
		assert.Nil(t, updated.Parent)
	}
}

func TestItemsRepository_GetOne(t *testing.T) {
	entity := useItems(t, 3)

	for _, item := range entity {
		result, err := tRepos.Items.GetOne(context.Background(), item.ID)
		require.NoError(t, err)
		assert.Equal(t, item.ID, result.ID)
	}
}

func TestItemsRepository_GetAll(t *testing.T) {
	length := 10
	expected := useItems(t, length)

	results, err := tRepos.Items.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)

	assert.Len(t, results, length)

	for _, item := range results {
		for _, expectedItem := range expected {
			if item.ID == expectedItem.ID {
				assert.Equal(t, expectedItem.ID, item.ID)
				assert.Equal(t, expectedItem.Name, item.Name)
				assert.Equal(t, expectedItem.Description, item.Description)
			}
		}
	}
}

func TestItemsRepository_Create(t *testing.T) {
	location, err := tRepos.Locations.Create(context.Background(), tGroup.ID, locationFactory())
	require.NoError(t, err)

	itm := itemFactory()
	itm.LocationID = location.ID

	result, err := tRepos.Items.Create(context.Background(), tGroup.ID, itm)
	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)

	// Cleanup - Also deletes item
	err = tRepos.Locations.delete(context.Background(), location.ID)
	require.NoError(t, err)
}

func TestItemsRepository_Create_Location(t *testing.T) {
	location, err := tRepos.Locations.Create(context.Background(), tGroup.ID, locationFactory())
	require.NoError(t, err)
	assert.NotEmpty(t, location.ID)

	item := itemFactory()
	item.LocationID = location.ID

	// Create Resource
	result, err := tRepos.Items.Create(context.Background(), tGroup.ID, item)
	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)

	// Get Resource
	foundItem, err := tRepos.Items.GetOne(context.Background(), result.ID)
	require.NoError(t, err)
	assert.Equal(t, result.ID, foundItem.ID)
	assert.Equal(t, location.ID, foundItem.Location.ID)

	// Cleanup - Also deletes item
	err = tRepos.Locations.delete(context.Background(), location.ID)
	require.NoError(t, err)
}

func TestItemsRepository_Delete(t *testing.T) {
	entities := useItems(t, 3)

	for _, item := range entities {
		err := tRepos.Items.Delete(context.Background(), item.ID)
		require.NoError(t, err)
	}

	results, err := tRepos.Items.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestItemsRepository_Update_Tags(t *testing.T) {
	entity := useItems(t, 1)[0]
	tags := useTags(t, 3)

	tagsIDs := []uuid.UUID{tags[0].ID, tags[1].ID, tags[2].ID}

	type args struct {
		tagIds []uuid.UUID
	}

	tests := []struct {
		name string
		args args
		want []uuid.UUID
	}{
		{
			name: "add all tags",
			args: args{
				tagIds: tagsIDs,
			},
			want: tagsIDs,
		},
		{
			name: "update with one tag",
			args: args{
				tagIds: tagsIDs[:1],
			},
			want: tagsIDs[:1],
		},
		{
			name: "add one new tag to existing single tag",
			args: args{
				tagIds: tagsIDs[1:],
			},
			want: tagsIDs[1:],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply all tags to entity
			updateData := ItemUpdate{
				ID:         entity.ID,
				Name:       entity.Name,
				LocationID: entity.Location.ID,
				TagIDs:     tt.args.tagIds,
			}

			updated, err := tRepos.Items.UpdateByGroup(context.Background(), tGroup.ID, updateData)
			require.NoError(t, err)
			assert.Len(t, tt.want, len(updated.Tags))

			for _, tag := range updated.Tags {
				assert.Contains(t, tt.want, tag.ID)
			}
		})
	}
}

func TestItemsRepository_Update(t *testing.T) {
	entities := useItems(t, 3)

	entity := entities[0]

	updateData := ItemUpdate{
		ID:               entity.ID,
		Name:             entity.Name,
		LocationID:       entity.Location.ID,
		SerialNumber:     fk.Str(10),
		TagIDs:           nil,
		ModelNumber:      fk.Str(10),
		Manufacturer:     fk.Str(10),
		PurchaseTime:     types.DateFromTime(time.Now()),
		PurchaseFrom:     fk.Str(10),
		PurchasePrice:    300.99,
		SoldTime:         types.DateFromTime(time.Now()),
		SoldTo:           fk.Str(10),
		SoldPrice:        300.99,
		SoldNotes:        fk.Str(10),
		Notes:            fk.Str(10),
		WarrantyExpires:  types.DateFromTime(time.Now()),
		WarrantyDetails:  fk.Str(10),
		LifetimeWarranty: true,
	}

	updatedEntity, err := tRepos.Items.UpdateByGroup(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)

	got, err := tRepos.Items.GetOne(context.Background(), updatedEntity.ID)
	require.NoError(t, err)

	assert.Equal(t, updateData.ID, got.ID)
	assert.Equal(t, updateData.Name, got.Name)
	assert.Equal(t, updateData.LocationID, got.Location.ID)
	assert.Equal(t, updateData.SerialNumber, got.SerialNumber)
	assert.Equal(t, updateData.ModelNumber, got.ModelNumber)
	assert.Equal(t, updateData.Manufacturer, got.Manufacturer)
	// assert.Equal(t, updateData.PurchaseTime, got.PurchaseTime)
	assert.Equal(t, updateData.PurchaseFrom, got.PurchaseFrom)
	assert.InDelta(t, updateData.PurchasePrice, got.PurchasePrice, 0.01)
	// assert.Equal(t, updateData.SoldTime, got.SoldTime)
	assert.Equal(t, updateData.SoldTo, got.SoldTo)
	assert.InDelta(t, updateData.SoldPrice, got.SoldPrice, 0.01)
	assert.Equal(t, updateData.SoldNotes, got.SoldNotes)
	assert.Equal(t, updateData.Notes, got.Notes)
	// assert.Equal(t, updateData.WarrantyExpires, got.WarrantyExpires)
	assert.Equal(t, updateData.WarrantyDetails, got.WarrantyDetails)
	assert.Equal(t, updateData.LifetimeWarranty, got.LifetimeWarranty)
}

func TestItemRepository_GetAllCustomFields(t *testing.T) {
	const FieldsCount = 5

	entity := useItems(t, 1)[0]

	fields := make([]ItemField, FieldsCount)
	names := make([]string, FieldsCount)
	values := make([]string, FieldsCount)

	for i := 0; i < FieldsCount; i++ {
		name := fk.Str(10)
		fields[i] = ItemField{
			Name:      name,
			Type:      "text",
			TextValue: fk.Str(10),
		}
		names[i] = name
		values[i] = fields[i].TextValue
	}

	_, err := tRepos.Items.UpdateByGroup(context.Background(), tGroup.ID, ItemUpdate{
		ID:         entity.ID,
		Name:       entity.Name,
		LocationID: entity.Location.ID,
		Fields:     fields,
	})

	require.NoError(t, err)

	// Test getting all fields
	{
		results, err := tRepos.Items.GetAllCustomFieldNames(context.Background(), tGroup.ID)
		require.NoError(t, err)
		assert.ElementsMatch(t, names, results)
	}

	// Test getting all values from field
	{
		results, err := tRepos.Items.GetAllCustomFieldValues(context.Background(), tUser.DefaultGroupID, names[0])

		require.NoError(t, err)
		assert.ElementsMatch(t, values[:1], results)
	}
}

func TestItemsRepository_DeleteWithAttachments(t *testing.T) {
	// Create an item with an attachment
	item := useItems(t, 1)[0]

	// Add an attachment to the item
	attachment, err := tRepos.Attachments.Create(
		context.Background(),
		item.ID,
		ItemCreateAttachment{
			Title:   "test-attachment.txt",
			Content: strings.NewReader("test content for attachment deletion"),
		},
		attachment.TypePhoto,
		true,
	)
	require.NoError(t, err)
	assert.NotNil(t, attachment)

	// Verify the attachment exists
	retrievedAttachment, err := tRepos.Attachments.Get(context.Background(), tGroup.ID, attachment.ID)
	require.NoError(t, err)
	assert.Equal(t, attachment.ID, retrievedAttachment.ID)

	// Verify the attachment is linked to the item
	itemWithAttachments, err := tRepos.Items.GetOne(context.Background(), item.ID)
	require.NoError(t, err)
	assert.Len(t, itemWithAttachments.Attachments, 1)
	assert.Equal(t, attachment.ID, itemWithAttachments.Attachments[0].ID)

	// Delete the item
	err = tRepos.Items.Delete(context.Background(), item.ID)
	require.NoError(t, err)

	// Verify the item is deleted
	_, err = tRepos.Items.GetOne(context.Background(), item.ID)
	require.Error(t, err)

	// Verify the attachment is also deleted
	_, err = tRepos.Attachments.Get(context.Background(), tGroup.ID, attachment.ID)
	require.Error(t, err)
}

func TestItemsRepository_DeleteByGroupWithAttachments(t *testing.T) {
	// Create an item with an attachment
	item := useItems(t, 1)[0]

	// Add an attachment to the item
	attachment, err := tRepos.Attachments.Create(
		context.Background(),
		item.ID,
		ItemCreateAttachment{
			Title:   "test-attachment-by-group.txt",
			Content: strings.NewReader("test content for attachment deletion by group"),
		},
		attachment.TypePhoto,
		true,
	)
	require.NoError(t, err)
	assert.NotNil(t, attachment)

	// Verify the attachment exists
	retrievedAttachment, err := tRepos.Attachments.Get(context.Background(), tGroup.ID, attachment.ID)
	require.NoError(t, err)
	assert.Equal(t, attachment.ID, retrievedAttachment.ID)

	// Delete the item using DeleteByGroup
	err = tRepos.Items.DeleteByGroup(context.Background(), tGroup.ID, item.ID)
	require.NoError(t, err)

	// Verify the item is deleted
	_, err = tRepos.Items.GetOneByGroup(context.Background(), tGroup.ID, item.ID)
	require.Error(t, err)

	// Verify the attachment is also deleted
	_, err = tRepos.Attachments.Get(context.Background(), tGroup.ID, attachment.ID)
	require.Error(t, err)
}

func TestItemsRepository_WipeInventory(t *testing.T) {
	// Create test data: items, tags, locations, and maintenance entries
	// Create locations
	loc1, err := tRepos.Locations.Create(context.Background(), tGroup.ID, LocationCreate{
		Name:        "Test Location 1",
		Description: "Test location for wipe test",
	})
	require.NoError(t, err)

	loc2, err := tRepos.Locations.Create(context.Background(), tGroup.ID, LocationCreate{
		Name:        "Test Location 2",
		Description: "Another test location",
	})
	require.NoError(t, err)

	// Create tags
	tag1, err := tRepos.Tags.Create(context.Background(), tGroup.ID, TagCreate{
		Name:        "Test Tag 1",
		Description: "Test tag for wipe test",
	})
	require.NoError(t, err)

	tag2, err := tRepos.Tags.Create(context.Background(), tGroup.ID, TagCreate{
		Name:        "Test Tag 2",
		Description: "Another test tag",
	})
	require.NoError(t, err)

	// Create items
	item1, err := tRepos.Items.Create(context.Background(), tGroup.ID, ItemCreate{
		Name:        "Test Item 1",
		Description: "Test item for wipe test",
		LocationID:  loc1.ID,
		TagIDs:      []uuid.UUID{tag1.ID},
	})
	require.NoError(t, err)

	item2, err := tRepos.Items.Create(context.Background(), tGroup.ID, ItemCreate{
		Name:        "Test Item 2",
		Description: "Another test item",
		LocationID:  loc2.ID,
		TagIDs:      []uuid.UUID{tag2.ID},
	})
	require.NoError(t, err)

	// Create maintenance entries for items
	_, err = tRepos.MaintEntry.Create(context.Background(), item1.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Test Maintenance 1",
		Description:   "Test maintenance entry",
		Cost:          100.0,
	})
	require.NoError(t, err)

	_, err = tRepos.MaintEntry.Create(context.Background(), item2.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Test Maintenance 2",
		Description:   "Another test maintenance entry",
		Cost:          200.0,
	})
	require.NoError(t, err)

	// Test 1: Wipe inventory with all options enabled
	t.Run("wipe all including tags, locations, and maintenance", func(t *testing.T) {
		deleted, err := tRepos.Items.WipeInventory(context.Background(), tGroup.ID, true, true, true)
		require.NoError(t, err)
		assert.Positive(t, deleted, "Should have deleted at least some entities")
		// Verify items are deleted
		_, err = tRepos.Items.GetOneByGroup(context.Background(), tGroup.ID, item1.ID)
		require.Error(t, err, "Item 1 should be deleted")

		_, err = tRepos.Items.GetOneByGroup(context.Background(), tGroup.ID, item2.ID)
		require.Error(t, err, "Item 2 should be deleted")

		// Verify maintenance entries are deleted (query by item ID, should return empty)
		maint1List, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, item1.ID, MaintenanceFilters{})
		require.NoError(t, err)
		assert.Empty(t, maint1List, "Maintenance entry 1 should be deleted")

		maint2List, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, item2.ID, MaintenanceFilters{})
		require.NoError(t, err)
		assert.Empty(t, maint2List, "Maintenance entry 2 should be deleted")

		// Verify tags are deleted
		_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tag1.ID)
		require.Error(t, err, "Tag 1 should be deleted")

		_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tag2.ID)
		require.Error(t, err, "Tag 2 should be deleted")
		// Verify locations are deleted
		_, err = tRepos.Locations.Get(context.Background(), loc1.ID)
		require.Error(t, err, "Location 1 should be deleted")

		_, err = tRepos.Locations.Get(context.Background(), loc2.ID)
		require.Error(t, err, "Location 2 should be deleted")
	})
}

func TestItemsRepository_WipeInventory_OnlyItems(t *testing.T) {
	// Create test data
	loc, err := tRepos.Locations.Create(context.Background(), tGroup.ID, LocationCreate{
		Name:        "Test Location",
		Description: "Test location for wipe test",
	})
	require.NoError(t, err)

	tag, err := tRepos.Tags.Create(context.Background(), tGroup.ID, TagCreate{
		Name:        "Test Tag",
		Description: "Test tag for wipe test",
	})
	require.NoError(t, err)

	item, err := tRepos.Items.Create(context.Background(), tGroup.ID, ItemCreate{
		Name:        "Test Item",
		Description: "Test item for wipe test",
		LocationID:  loc.ID,
		TagIDs:      []uuid.UUID{tag.ID},
	})
	require.NoError(t, err)

	_, err = tRepos.MaintEntry.Create(context.Background(), item.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Test Maintenance",
		Description:   "Test maintenance entry",
		Cost:          100.0,
	})
	require.NoError(t, err)

	// Test: Wipe inventory with only items (no tags, locations, or maintenance)
	deleted, err := tRepos.Items.WipeInventory(context.Background(), tGroup.ID, false, false, false)
	require.NoError(t, err)
	assert.Positive(t, deleted, "Should have deleted at least the item")
	// Verify item is deleted
	_, err = tRepos.Items.GetOneByGroup(context.Background(), tGroup.ID, item.ID)
	require.Error(t, err, "Item should be deleted")

	// Verify maintenance entry is deleted due to cascade
	maintList, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, item.ID, MaintenanceFilters{})
	require.NoError(t, err)
	assert.Empty(t, maintList, "Maintenance entry should be cascade deleted with item")

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
