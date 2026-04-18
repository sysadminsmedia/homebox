package repo

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

func containerFactory() EntityCreate {
	return EntityCreate{
		Name:        fk.Str(10),
		Description: fk.Str(100),
	}
}

func entityFactory() EntityCreate {
	return EntityCreate{
		Name:        fk.Str(10),
		Description: fk.Str(100),
	}
}

// useContainerEntityType creates or gets a default location entity type for the test group.
func useContainerEntityType(t *testing.T) EntityTypeSummary {
	t.Helper()
	et, err := tRepos.EntityTypes.GetDefault(context.Background(), tGroup.ID, true)
	require.NoError(t, err)
	return et
}

// useItemEntityType creates or gets a default item entity type for the test group.
func useItemEntityType(t *testing.T) EntityTypeSummary {
	t.Helper()
	et, err := tRepos.EntityTypes.GetDefault(context.Background(), tGroup.ID, false)
	require.NoError(t, err)
	return et
}

func useEntities(t *testing.T, count int) []EntityOut {
	t.Helper()

	containerET := useContainerEntityType(t)
	itemET := useItemEntityType(t)

	// Create a container entity
	cf := containerFactory()
	cf.EntityTypeID = containerET.ID
	container, err := tRepos.Entities.Create(context.Background(), tGroup.ID, cf)
	require.NoError(t, err)

	entities := make([]EntityOut, count)
	for i := 0; i < count; i++ {
		itm := entityFactory()
		itm.ParentID = container.ID
		itm.EntityTypeID = itemET.ID

		e, err := tRepos.Entities.Create(context.Background(), tGroup.ID, itm)
		require.NoError(t, err)
		entities[i] = e
	}

	t.Cleanup(func() {
		for _, e := range entities {
			_ = tRepos.Entities.Delete(context.Background(), e.ID)
		}
		_ = tRepos.Entities.Delete(context.Background(), container.ID)
	})

	return entities
}

func TestEntityRepository_RecursiveRelationships(t *testing.T) {
	parent := useEntities(t, 1)[0]

	children := useEntities(t, 3)

	for _, child := range children {
		update := EntityUpdate{
			ID:          child.ID,
			ParentID:    parent.ID,
			Name:        "note-important",
			Description: "This is a note",
		}
		if child.EntityType != nil {
			update.EntityTypeID = child.EntityType.ID
		}

		// Append Parent ID
		_, err := tRepos.Entities.UpdateByGroup(context.Background(), tGroup.ID, update)
		require.NoError(t, err)

		// Check Parent ID
		updated, err := tRepos.Entities.GetOne(context.Background(), child.ID)
		require.NoError(t, err)
		assert.Equal(t, parent.ID, updated.Parent.ID)

		// Remove Parent ID
		update.ParentID = uuid.Nil
		_, err = tRepos.Entities.UpdateByGroup(context.Background(), tGroup.ID, update)
		require.NoError(t, err)

		// Check Parent ID
		updated, err = tRepos.Entities.GetOne(context.Background(), child.ID)
		require.NoError(t, err)
		assert.Nil(t, updated.Parent)
	}
}

func TestEntityRepository_GetOne(t *testing.T) {
	entities := useEntities(t, 3)

	for _, e := range entities {
		result, err := tRepos.Entities.GetOne(context.Background(), e.ID)
		require.NoError(t, err)
		assert.Equal(t, e.ID, result.ID)
	}
}

func TestEntityRepository_GetAll(t *testing.T) {
	length := 10
	expected := useEntities(t, length)

	results, err := tRepos.Entities.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)

	// Results include the container + the items
	assert.GreaterOrEqual(t, len(results), length)

	for _, e := range expected {
		found := false
		for _, r := range results {
			if e.ID == r.ID {
				found = true
				assert.Equal(t, e.Name, r.Name)
				assert.Equal(t, e.Description, r.Description)
			}
		}
		assert.True(t, found, "expected entity not found in results")
	}
}

func TestEntityRepository_Create(t *testing.T) {
	containerET := useContainerEntityType(t)
	itemET := useItemEntityType(t)

	cf := containerFactory()
	cf.EntityTypeID = containerET.ID
	container, err := tRepos.Entities.Create(context.Background(), tGroup.ID, cf)
	require.NoError(t, err)

	itm := entityFactory()
	itm.ParentID = container.ID
	itm.EntityTypeID = itemET.ID

	result, err := tRepos.Entities.Create(context.Background(), tGroup.ID, itm)
	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)

	// Cleanup
	err = tRepos.Entities.Delete(context.Background(), result.ID)
	require.NoError(t, err)
	err = tRepos.Entities.Delete(context.Background(), container.ID)
	require.NoError(t, err)
}

func TestEntityRepository_Create_WithFractionalQuantity(t *testing.T) {
	containerET := useContainerEntityType(t)
	itemET := useItemEntityType(t)

	cf := containerFactory()
	cf.EntityTypeID = containerET.ID
	container, err := tRepos.Entities.Create(context.Background(), tGroup.ID, cf)
	require.NoError(t, err)

	itm := entityFactory()
	itm.ParentID = container.ID
	itm.EntityTypeID = itemET.ID
	itm.Quantity = 1.25

	result, err := tRepos.Entities.Create(context.Background(), tGroup.ID, itm)
	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.InDelta(t, 1.25, result.Quantity, 0.000001)

	fetched, err := tRepos.Entities.GetOne(context.Background(), result.ID)
	require.NoError(t, err)
	assert.InDelta(t, 1.25, fetched.Quantity, 0.000001)

	// Cleanup
	err = tRepos.Entities.Delete(context.Background(), result.ID)
	require.NoError(t, err)
	err = tRepos.Entities.Delete(context.Background(), container.ID)
	require.NoError(t, err)
}

func TestEntityRepository_Create_RejectsNonFiniteQuantity(t *testing.T) {
	containerET := useContainerEntityType(t)
	itemET := useItemEntityType(t)

	cf := containerFactory()
	cf.EntityTypeID = containerET.ID
	container, err := tRepos.Entities.Create(context.Background(), tGroup.ID, cf)
	require.NoError(t, err)

	itm := entityFactory()
	itm.ParentID = container.ID
	itm.EntityTypeID = itemET.ID
	itm.Quantity = math.NaN()

	_, err = tRepos.Entities.Create(context.Background(), tGroup.ID, itm)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid quantity: must be a finite number")

	// Cleanup
	err = tRepos.Entities.Delete(context.Background(), container.ID)
	require.NoError(t, err)
}

func TestEntityRepository_Create_WithParent(t *testing.T) {
	containerET := useContainerEntityType(t)
	itemET := useItemEntityType(t)

	cf := containerFactory()
	cf.EntityTypeID = containerET.ID
	container, err := tRepos.Entities.Create(context.Background(), tGroup.ID, cf)
	require.NoError(t, err)
	assert.NotEmpty(t, container.ID)

	itm := entityFactory()
	itm.ParentID = container.ID
	itm.EntityTypeID = itemET.ID

	// Create Resource
	result, err := tRepos.Entities.Create(context.Background(), tGroup.ID, itm)
	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)

	// Get Resource
	foundEntity, err := tRepos.Entities.GetOne(context.Background(), result.ID)
	require.NoError(t, err)
	assert.Equal(t, result.ID, foundEntity.ID)
	assert.NotNil(t, foundEntity.Parent)
	assert.Equal(t, container.ID, foundEntity.Parent.ID)

	// Cleanup
	err = tRepos.Entities.Delete(context.Background(), result.ID)
	require.NoError(t, err)
	err = tRepos.Entities.Delete(context.Background(), container.ID)
	require.NoError(t, err)
}

func TestEntityRepository_Delete(t *testing.T) {
	entities := useEntities(t, 3)

	for _, e := range entities {
		err := tRepos.Entities.Delete(context.Background(), e.ID)
		require.NoError(t, err)
	}

	results, err := tRepos.Entities.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)
	// After deleting items, only container(s) remain
	for _, e := range entities {
		for _, r := range results {
			assert.NotEqual(t, e.ID, r.ID)
		}
	}
}

func TestEntityRepository_Update_Tags(t *testing.T) {
	e := useEntities(t, 1)[0]
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
			updateData := EntityUpdate{
				ID:     e.ID,
				Name:   e.Name,
				TagIDs: tt.args.tagIds,
			}
			if e.EntityType != nil {
				updateData.EntityTypeID = e.EntityType.ID
			}

			updated, err := tRepos.Entities.UpdateByGroup(context.Background(), tGroup.ID, updateData)
			require.NoError(t, err)
			assert.Len(t, tt.want, len(updated.Tags))

			for _, tag := range updated.Tags {
				assert.Contains(t, tt.want, tag.ID)
			}
		})
	}
}

func TestEntityRepository_QueryByGroup_TagFilter(t *testing.T) {
	// Set up 3 entities and 2 tags:
	//   entity1 -> tagA only
	//   entity2 -> tagA + tagB
	//   entity3 -> tagB only
	entities := useEntities(t, 3)
	tags := useTags(t, 2)
	tagA, tagB := tags[0], tags[1]

	assignTags := func(e EntityOut, tagIDs []uuid.UUID) {
		t.Helper()
		update := EntityUpdate{ID: e.ID, Name: e.Name, TagIDs: tagIDs}
		if e.EntityType != nil {
			update.EntityTypeID = e.EntityType.ID
		}
		_, err := tRepos.Entities.UpdateByGroup(context.Background(), tGroup.ID, update)
		require.NoError(t, err)
	}

	assignTags(entities[0], []uuid.UUID{tagA.ID})
	assignTags(entities[1], []uuid.UUID{tagA.ID, tagB.ID})
	assignTags(entities[2], []uuid.UUID{tagB.ID})

	containsID := func(results []EntitySummary, id uuid.UUID) bool {
		for _, r := range results {
			if r.ID == id {
				return true
			}
		}
		return false
	}

	t.Run("OR mode returns entities matching any tag", func(t *testing.T) {
		result, err := tRepos.Entities.QueryByGroup(context.Background(), tGroup.ID, EntityQuery{
			TagIDs:  []uuid.UUID{tagA.ID, tagB.ID},
			TagsAND: false,
		})
		require.NoError(t, err)

		// All three entities have at least one of the two tags
		assert.True(t, containsID(result.Items, entities[0].ID), "entity1 (tagA only) should be included")
		assert.True(t, containsID(result.Items, entities[1].ID), "entity2 (tagA+tagB) should be included")
		assert.True(t, containsID(result.Items, entities[2].ID), "entity3 (tagB only) should be included")
	})

	t.Run("AND mode returns only entities matching all tags", func(t *testing.T) {
		result, err := tRepos.Entities.QueryByGroup(context.Background(), tGroup.ID, EntityQuery{
			TagIDs:  []uuid.UUID{tagA.ID, tagB.ID},
			TagsAND: true,
		})
		require.NoError(t, err)

		// Only entity2 has both tags
		assert.False(t, containsID(result.Items, entities[0].ID), "entity1 (tagA only) should be excluded")
		assert.True(t, containsID(result.Items, entities[1].ID), "entity2 (tagA+tagB) should be included")
		assert.False(t, containsID(result.Items, entities[2].ID), "entity3 (tagB only) should be excluded")
	})
}

func TestEntityRepository_Update(t *testing.T) {
	entities := useEntities(t, 3)

	e := entities[0]

	updateData := EntityUpdate{
		ID:               e.ID,
		Name:             e.Name,
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
	if e.EntityType != nil {
		updateData.EntityTypeID = e.EntityType.ID
	}

	updatedEntity, err := tRepos.Entities.UpdateByGroup(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)

	got, err := tRepos.Entities.GetOne(context.Background(), updatedEntity.ID)
	require.NoError(t, err)

	assert.Equal(t, updateData.ID, got.ID)
	assert.Equal(t, updateData.Name, got.Name)
	assert.Equal(t, updateData.SerialNumber, got.SerialNumber)
	assert.Equal(t, updateData.ModelNumber, got.ModelNumber)
	assert.Equal(t, updateData.Manufacturer, got.Manufacturer)
	assert.Equal(t, updateData.PurchaseFrom, got.PurchaseFrom)
	assert.InDelta(t, updateData.PurchasePrice, got.PurchasePrice, 0.01)
	assert.Equal(t, updateData.SoldTo, got.SoldTo)
	assert.InDelta(t, updateData.SoldPrice, got.SoldPrice, 0.01)
	assert.Equal(t, updateData.SoldNotes, got.SoldNotes)
	assert.Equal(t, updateData.Notes, got.Notes)
	assert.Equal(t, updateData.WarrantyDetails, got.WarrantyDetails)
	assert.Equal(t, updateData.LifetimeWarranty, got.LifetimeWarranty)
}

func TestEntityRepository_Update_WithFractionalQuantity(t *testing.T) {
	e := useEntities(t, 1)[0]

	updateData := EntityUpdate{
		ID:       e.ID,
		Name:     e.Name,
		Quantity: 2.75,
	}
	if e.EntityType != nil {
		updateData.EntityTypeID = e.EntityType.ID
	}

	updatedEntity, err := tRepos.Entities.UpdateByGroup(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)

	got, err := tRepos.Entities.GetOne(context.Background(), updatedEntity.ID)
	require.NoError(t, err)

	assert.InDelta(t, 2.75, got.Quantity, 0.000001)
}

func TestEntityRepository_Update_RejectsNonFiniteQuantity(t *testing.T) {
	e := useEntities(t, 1)[0]

	updateData := EntityUpdate{
		ID:       e.ID,
		Name:     e.Name,
		Quantity: math.Inf(1),
	}

	_, err := tRepos.Entities.UpdateByGroup(context.Background(), tGroup.ID, updateData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid quantity: must be a finite number")
}

func TestEntityRepository_Patch_RejectsNonFiniteQuantity(t *testing.T) {
	e := useEntities(t, 1)[0]

	quantity := math.Inf(-1)
	err := tRepos.Entities.Patch(context.Background(), tGroup.ID, e.ID, EntityPatch{Quantity: &quantity})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid quantity: must be a finite number")
}

func TestEntityRepository_CreateFromTemplate_RejectsNonFiniteQuantity(t *testing.T) {
	containerET := useContainerEntityType(t)

	cf := containerFactory()
	cf.EntityTypeID = containerET.ID
	container, err := tRepos.Entities.Create(context.Background(), tGroup.ID, cf)
	require.NoError(t, err)

	_, err = tRepos.Entities.CreateFromTemplate(context.Background(), tGroup.ID, EntityCreateFromTemplate{
		Name:        fk.Str(10),
		Description: fk.Str(20),
		Quantity:    math.NaN(),
		ParentID:    container.ID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid quantity: must be a finite number")

	// Cleanup
	err = tRepos.Entities.Delete(context.Background(), container.ID)
	require.NoError(t, err)
}

func TestEntityRepository_GetAllCustomFields(t *testing.T) {
	const FieldsCount = 5

	e := useEntities(t, 1)[0]

	fields := make([]EntityFieldData, FieldsCount)
	names := make([]string, FieldsCount)
	values := make([]string, FieldsCount)

	for i := 0; i < FieldsCount; i++ {
		name := fk.Str(10)
		fields[i] = EntityFieldData{
			Name:      name,
			Type:      "text",
			TextValue: fk.Str(10),
		}
		names[i] = name
		values[i] = fields[i].TextValue
	}

	updateData := EntityUpdate{
		ID:     e.ID,
		Name:   e.Name,
		Fields: fields,
	}
	if e.EntityType != nil {
		updateData.EntityTypeID = e.EntityType.ID
	}

	_, err := tRepos.Entities.UpdateByGroup(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)

	// Test getting all fields
	{
		results, err := tRepos.Entities.GetAllCustomFieldNames(context.Background(), tGroup.ID)
		require.NoError(t, err)
		assert.ElementsMatch(t, names, results)
	}

	// Test getting all values from field
	{
		results, err := tRepos.Entities.GetAllCustomFieldValues(context.Background(), tUser.DefaultGroupID, names[0])

		require.NoError(t, err)
		assert.ElementsMatch(t, values[:1], results)
	}
}

func TestEntityRepository_DeleteWithAttachments(t *testing.T) {
	// Create an entity with an attachment
	e := useEntities(t, 1)[0]

	// Add an attachment to the entity
	att, err := tRepos.Attachments.Create(
		context.Background(),
		e.ID,
		ItemCreateAttachment{
			Title:   "test-attachment.txt",
			Content: strings.NewReader("test content for attachment deletion"),
		},
		attachment.TypePhoto,
		true,
	)
	require.NoError(t, err)
	assert.NotNil(t, att)

	// Verify the attachment exists
	retrievedAttachment, err := tRepos.Attachments.Get(context.Background(), tGroup.ID, att.ID)
	require.NoError(t, err)
	assert.Equal(t, att.ID, retrievedAttachment.ID)

	// Verify the attachment is linked to the entity
	entityWithAttachments, err := tRepos.Entities.GetOne(context.Background(), e.ID)
	require.NoError(t, err)
	assert.Len(t, entityWithAttachments.Attachments, 1)
	assert.Equal(t, att.ID, entityWithAttachments.Attachments[0].ID)

	// Delete the entity
	err = tRepos.Entities.Delete(context.Background(), e.ID)
	require.NoError(t, err)

	// Verify the entity is deleted
	_, err = tRepos.Entities.GetOne(context.Background(), e.ID)
	require.Error(t, err)

	// Verify the attachment is also deleted
	_, err = tRepos.Attachments.Get(context.Background(), tGroup.ID, att.ID)
	require.Error(t, err)
}

func TestEntityRepository_DeleteByGroupWithAttachments(t *testing.T) {
	// Create an entity with an attachment
	e := useEntities(t, 1)[0]

	// Add an attachment to the entity
	att, err := tRepos.Attachments.Create(
		context.Background(),
		e.ID,
		ItemCreateAttachment{
			Title:   "test-attachment-by-group.txt",
			Content: strings.NewReader("test content for attachment deletion by group"),
		},
		attachment.TypePhoto,
		true,
	)
	require.NoError(t, err)
	assert.NotNil(t, att)

	// Verify the attachment exists
	retrievedAttachment, err := tRepos.Attachments.Get(context.Background(), tGroup.ID, att.ID)
	require.NoError(t, err)
	assert.Equal(t, att.ID, retrievedAttachment.ID)

	// Delete the entity using DeleteByGroup
	err = tRepos.Entities.DeleteByGroup(context.Background(), tGroup.ID, e.ID)
	require.NoError(t, err)

	// Verify the entity is deleted
	_, err = tRepos.Entities.GetOneByGroup(context.Background(), tGroup.ID, e.ID)
	require.Error(t, err)

	// Verify the attachment is also deleted
	_, err = tRepos.Attachments.Get(context.Background(), tGroup.ID, att.ID)
	require.Error(t, err)
}

func TestEntityRepository_WipeInventory(t *testing.T) {
	containerET := useContainerEntityType(t)
	itemET := useItemEntityType(t)

	// Create containers
	c1f := containerFactory()
	c1f.EntityTypeID = containerET.ID
	c1f.Name = "Test Container 1"
	c1f.Description = "Test container for wipe test"
	container1, err := tRepos.Entities.Create(context.Background(), tGroup.ID, c1f)
	require.NoError(t, err)

	c2f := containerFactory()
	c2f.EntityTypeID = containerET.ID
	c2f.Name = "Test Container 2"
	c2f.Description = "Another test container"
	container2, err := tRepos.Entities.Create(context.Background(), tGroup.ID, c2f)
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
	i1f := entityFactory()
	i1f.ParentID = container1.ID
	i1f.EntityTypeID = itemET.ID
	i1f.Name = "Test Item 1"
	i1f.Description = "Test item for wipe test"
	i1f.TagIDs = []uuid.UUID{tag1.ID}
	entity1, err := tRepos.Entities.Create(context.Background(), tGroup.ID, i1f)
	require.NoError(t, err)

	i2f := entityFactory()
	i2f.ParentID = container2.ID
	i2f.EntityTypeID = itemET.ID
	i2f.Name = "Test Item 2"
	i2f.Description = "Another test item"
	i2f.TagIDs = []uuid.UUID{tag2.ID}
	entity2, err := tRepos.Entities.Create(context.Background(), tGroup.ID, i2f)
	require.NoError(t, err)

	// Create maintenance entries
	_, err = tRepos.MaintEntry.Create(context.Background(), entity1.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Test Maintenance 1",
		Description:   "Test maintenance entry",
		Cost:          100.0,
	})
	require.NoError(t, err)

	_, err = tRepos.MaintEntry.Create(context.Background(), entity2.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Test Maintenance 2",
		Description:   "Another test maintenance entry",
		Cost:          200.0,
	})
	require.NoError(t, err)

	// Test: Wipe inventory with all options enabled
	t.Run("wipe all including tags, containers, and maintenance", func(t *testing.T) {
		deleted, err := tRepos.Entities.WipeInventory(context.Background(), tGroup.ID, true, true, true)
		require.NoError(t, err)
		assert.Positive(t, deleted, "Should have deleted at least some entities")

		// Verify items are deleted
		_, err = tRepos.Entities.GetOneByGroup(context.Background(), tGroup.ID, entity1.ID)
		require.Error(t, err, "Entity 1 should be deleted")

		_, err = tRepos.Entities.GetOneByGroup(context.Background(), tGroup.ID, entity2.ID)
		require.Error(t, err, "Entity 2 should be deleted")

		// Verify maintenance entries are deleted
		maint1List, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, entity1.ID, MaintenanceFilters{})
		require.NoError(t, err)
		assert.Empty(t, maint1List, "Maintenance entry 1 should be deleted")

		maint2List, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, entity2.ID, MaintenanceFilters{})
		require.NoError(t, err)
		assert.Empty(t, maint2List, "Maintenance entry 2 should be deleted")

		// Verify tags are deleted
		_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tag1.ID)
		require.Error(t, err, "Tag 1 should be deleted")

		_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tag2.ID)
		require.Error(t, err, "Tag 2 should be deleted")
	})
}

func TestEntityRepository_WipeInventory_OnlyItems(t *testing.T) {
	containerET := useContainerEntityType(t)
	itemET := useItemEntityType(t)

	// Create test data
	cf := containerFactory()
	cf.EntityTypeID = containerET.ID
	cf.Name = "Test Container"
	cf.Description = "Test container for wipe test"
	container, err := tRepos.Entities.Create(context.Background(), tGroup.ID, cf)
	require.NoError(t, err)

	tagObj, err := tRepos.Tags.Create(context.Background(), tGroup.ID, TagCreate{
		Name:        "Test Tag",
		Description: "Test tag for wipe test",
	})
	require.NoError(t, err)

	ef := entityFactory()
	ef.ParentID = container.ID
	ef.EntityTypeID = itemET.ID
	ef.Name = "Test Item"
	ef.Description = "Test item for wipe test"
	ef.TagIDs = []uuid.UUID{tagObj.ID}
	e, err := tRepos.Entities.Create(context.Background(), tGroup.ID, ef)
	require.NoError(t, err)

	_, err = tRepos.MaintEntry.Create(context.Background(), e.ID, MaintenanceEntryCreate{
		CompletedDate: types.DateFromTime(time.Now()),
		Name:          "Test Maintenance",
		Description:   "Test maintenance entry",
		Cost:          100.0,
	})
	require.NoError(t, err)

	// Test: Wipe inventory with only items (no tags, containers, or maintenance)
	deleted, err := tRepos.Entities.WipeInventory(context.Background(), tGroup.ID, false, false, false)
	require.NoError(t, err)
	assert.Positive(t, deleted, "Should have deleted at least the entity")

	// Verify item entity is deleted
	_, err = tRepos.Entities.GetOneByGroup(context.Background(), tGroup.ID, e.ID)
	require.Error(t, err, "Entity should be deleted")

	// Verify maintenance entry is deleted due to cascade
	maintList, err := tRepos.MaintEntry.GetMaintenanceByItemID(context.Background(), tGroup.ID, e.ID, MaintenanceFilters{})
	require.NoError(t, err)
	assert.Empty(t, maintList, "Maintenance entry should be cascade deleted with entity")

	// Verify tag still exists
	_, err = tRepos.Tags.GetOneByGroup(context.Background(), tGroup.ID, tagObj.ID)
	require.NoError(t, err, "Tag should still exist")

	// Cleanup
	_ = tRepos.Tags.DeleteByGroup(context.Background(), tGroup.ID, tagObj.ID)
}
