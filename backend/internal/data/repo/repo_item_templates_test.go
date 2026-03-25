package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func templateFactory() ItemTemplateCreate {
	return ItemTemplateCreate{
		Name:                    fk.Str(10),
		Description:             fk.Str(100),
		Notes:                   fk.Str(50),
		DefaultQuantity:         new(1.0),
		DefaultInsured:          false,
		DefaultName:             new(fk.Str(20)),
		DefaultDescription:      new(fk.Str(50)),
		DefaultManufacturer:     new(fk.Str(15)),
		DefaultModelNumber:      new(fk.Str(10)),
		DefaultLifetimeWarranty: false,
		DefaultWarrantyDetails:  new(""),
		IncludeWarrantyFields:   false,
		IncludePurchaseFields:   false,
		IncludeSoldFields:       false,
		Fields:                  []TemplateField{},
	}
}

func useTemplates(t *testing.T, count int) []ItemTemplateOut {
	t.Helper()

	templates := make([]ItemTemplateOut, count)
	for i := 0; i < count; i++ {
		data := templateFactory()

		template, err := tRepos.ItemTemplates.Create(context.Background(), tGroup.ID, data)
		require.NoError(t, err)
		templates[i] = template
	}

	t.Cleanup(func() {
		for _, template := range templates {
			_ = tRepos.ItemTemplates.Delete(context.Background(), tGroup.ID, template.ID)
		}
	})

	return templates
}

func TestItemTemplatesRepository_GetAll(t *testing.T) {
	useTemplates(t, 5)

	all, err := tRepos.ItemTemplates.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(all), 5)
}

func TestItemTemplatesRepository_GetOne(t *testing.T) {
	templates := useTemplates(t, 1)
	template := templates[0]

	found, err := tRepos.ItemTemplates.GetOne(context.Background(), tGroup.ID, template.ID)
	require.NoError(t, err)
	assert.Equal(t, template.ID, found.ID)
	assert.Equal(t, template.Name, found.Name)
	assert.Equal(t, template.Description, found.Description)
}

func TestItemTemplatesRepository_Create(t *testing.T) {
	data := templateFactory()

	template, err := tRepos.ItemTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, template.ID)
	assert.Equal(t, data.Name, template.Name)
	assert.Equal(t, data.Description, template.Description)
	assert.InDelta(t, *data.DefaultQuantity, template.DefaultQuantity, 0.0001)
	assert.Equal(t, data.DefaultInsured, template.DefaultInsured)
	assert.Equal(t, *data.DefaultName, template.DefaultName)
	assert.Equal(t, *data.DefaultDescription, template.DefaultDescription)
	assert.Equal(t, *data.DefaultManufacturer, template.DefaultManufacturer)
	assert.Equal(t, *data.DefaultModelNumber, template.DefaultModelNumber)

	// Cleanup
	err = tRepos.ItemTemplates.Delete(context.Background(), tGroup.ID, template.ID)
	require.NoError(t, err)
}

func TestItemTemplatesRepository_CreateWithFields(t *testing.T) {
	data := templateFactory()
	data.Fields = []TemplateField{
		{Name: "Field 1", Type: "text", TextValue: "Value 1"},
		{Name: "Field 2", Type: "text", TextValue: "Value 2"},
	}

	template, err := tRepos.ItemTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)

	assert.Len(t, template.Fields, 2)
	assert.Equal(t, "Field 1", template.Fields[0].Name)
	assert.Equal(t, "Value 1", template.Fields[0].TextValue)
	assert.Equal(t, "Field 2", template.Fields[1].Name)
	assert.Equal(t, "Value 2", template.Fields[1].TextValue)

	// Cleanup
	err = tRepos.ItemTemplates.Delete(context.Background(), tGroup.ID, template.ID)
	require.NoError(t, err)
}

func TestItemTemplatesRepository_Update(t *testing.T) {
	templates := useTemplates(t, 1)
	template := templates[0]

	updateData := ItemTemplateUpdate{
		ID:                      template.ID,
		Name:                    "Updated Name",
		Description:             "Updated Description",
		Notes:                   "Updated Notes",
		DefaultQuantity:         new(5.0),
		DefaultInsured:          true,
		DefaultName:             new("Default Item Name"),
		DefaultDescription:      new("Default Item Description"),
		DefaultManufacturer:     new("Updated Manufacturer"),
		DefaultModelNumber:      new("MODEL-123"),
		DefaultLifetimeWarranty: true,
		DefaultWarrantyDetails:  new("Lifetime coverage"),
		IncludeWarrantyFields:   true,
		IncludePurchaseFields:   true,
		IncludeSoldFields:       false,
		Fields:                  []TemplateField{},
	}

	updated, err := tRepos.ItemTemplates.Update(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)

	assert.Equal(t, template.ID, updated.ID)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated Description", updated.Description)
	assert.Equal(t, "Updated Notes", updated.Notes)
	assert.InDelta(t, 5.0, updated.DefaultQuantity, 0.0001)
	assert.True(t, updated.DefaultInsured)
	assert.Equal(t, "Default Item Name", updated.DefaultName)
	assert.Equal(t, "Default Item Description", updated.DefaultDescription)
	assert.Equal(t, "Updated Manufacturer", updated.DefaultManufacturer)
	assert.Equal(t, "MODEL-123", updated.DefaultModelNumber)
	assert.True(t, updated.DefaultLifetimeWarranty)
	assert.Equal(t, "Lifetime coverage", updated.DefaultWarrantyDetails)
	assert.True(t, updated.IncludeWarrantyFields)
	assert.True(t, updated.IncludePurchaseFields)
	assert.False(t, updated.IncludeSoldFields)
}

func TestItemTemplatesRepository_UpdateWithFields(t *testing.T) {
	data := templateFactory()
	data.Fields = []TemplateField{
		{Name: "Original Field", Type: "text", TextValue: "Original Value"},
	}

	template, err := tRepos.ItemTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)
	require.Len(t, template.Fields, 1)

	// Update with new fields
	updateData := ItemTemplateUpdate{
		ID:              template.ID,
		Name:            template.Name,
		Description:     template.Description,
		DefaultQuantity: new(template.DefaultQuantity),
		Fields: []TemplateField{
			{ID: template.Fields[0].ID, Name: "Updated Field", Type: "text", TextValue: "Updated Value"},
			{Name: "New Field", Type: "text", TextValue: "New Value"},
		},
	}

	updated, err := tRepos.ItemTemplates.Update(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)

	assert.Len(t, updated.Fields, 2)

	// Cleanup
	err = tRepos.ItemTemplates.Delete(context.Background(), tGroup.ID, template.ID)
	require.NoError(t, err)
}

func TestItemTemplatesRepository_Delete(t *testing.T) {
	data := templateFactory()

	template, err := tRepos.ItemTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)

	err = tRepos.ItemTemplates.Delete(context.Background(), tGroup.ID, template.ID)
	require.NoError(t, err)

	// Verify it's deleted
	_, err = tRepos.ItemTemplates.GetOne(context.Background(), tGroup.ID, template.ID)
	require.Error(t, err)
}

func TestItemTemplatesRepository_DeleteCascadesFields(t *testing.T) {
	data := templateFactory()
	data.Fields = []TemplateField{
		{Name: "Field 1", Type: "text", TextValue: "Value 1"},
		{Name: "Field 2", Type: "text", TextValue: "Value 2"},
	}

	template, err := tRepos.ItemTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)
	require.Len(t, template.Fields, 2)

	// Delete template - fields should be cascade deleted
	err = tRepos.ItemTemplates.Delete(context.Background(), tGroup.ID, template.ID)
	require.NoError(t, err)

	// Verify template is deleted
	_, err = tRepos.ItemTemplates.GetOne(context.Background(), tGroup.ID, template.ID)
	require.Error(t, err)
}

func TestItemTemplatesRepository_CreateWithLocation(t *testing.T) {
	// First create a location
	loc, err := tRepos.Locations.Create(context.Background(), tGroup.ID, LocationCreate{
		Name:        fk.Str(10),
		Description: fk.Str(50),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tRepos.Locations.delete(context.Background(), loc.ID)
	})

	// Create template with location
	data := templateFactory()
	data.DefaultLocationID = loc.ID

	template, err := tRepos.ItemTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)

	assert.NotNil(t, template.DefaultLocation)
	assert.Equal(t, loc.ID, template.DefaultLocation.ID)
	assert.Equal(t, loc.Name, template.DefaultLocation.Name)

	// Cleanup
	err = tRepos.ItemTemplates.Delete(context.Background(), tGroup.ID, template.ID)
	require.NoError(t, err)
}

func TestItemTemplatesRepository_CreateWithTags(t *testing.T) {
	// Create some tags
	tag1, err := tRepos.Tags.Create(context.Background(), tGroup.ID, TagCreate{
		Name:        fk.Str(10),
		Description: fk.Str(50),
	})
	require.NoError(t, err)

	tag2, err := tRepos.Tags.Create(context.Background(), tGroup.ID, TagCreate{
		Name:        fk.Str(10),
		Description: fk.Str(50),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tRepos.Tags.delete(context.Background(), tag1.ID)
		_ = tRepos.Tags.delete(context.Background(), tag2.ID)
	})

	// Create template with tags
	data := templateFactory()
	tagIDs := []uuid.UUID{tag1.ID, tag2.ID}
	data.DefaultTagIDs = &tagIDs

	template, err := tRepos.ItemTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)

	assert.Len(t, template.DefaultTags, 2)

	// Cleanup
	err = tRepos.ItemTemplates.Delete(context.Background(), tGroup.ID, template.ID)
	require.NoError(t, err)
}

func TestItemTemplatesRepository_UpdateRemoveLocation(t *testing.T) {
	// First create a location
	loc, err := tRepos.Locations.Create(context.Background(), tGroup.ID, LocationCreate{
		Name:        fk.Str(10),
		Description: fk.Str(50),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tRepos.Locations.delete(context.Background(), loc.ID)
	})

	// Create template with location
	data := templateFactory()
	data.DefaultLocationID = loc.ID

	template, err := tRepos.ItemTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)
	require.NotNil(t, template.DefaultLocation)

	// Update to remove location
	updateData := ItemTemplateUpdate{
		ID:                template.ID,
		Name:              template.Name,
		DefaultQuantity:   new(template.DefaultQuantity),
		DefaultLocationID: uuid.Nil, // Remove location
	}

	updated, err := tRepos.ItemTemplates.Update(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)

	assert.Nil(t, updated.DefaultLocation)

	// Cleanup
	err = tRepos.ItemTemplates.Delete(context.Background(), tGroup.ID, template.ID)
	require.NoError(t, err)
}
