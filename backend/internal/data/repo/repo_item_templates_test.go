package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func templateFactory() ItemTemplateCreate {
	qty := 1
	name := fk.Str(20)
	desc := fk.Str(50)
	mfr := fk.Str(15)
	model := fk.Str(10)
	warranty := ""

	return ItemTemplateCreate{
		Name:                    fk.Str(10),
		Description:             fk.Str(100),
		Notes:                   fk.Str(50),
		DefaultQuantity:         &qty,
		DefaultInsured:          false,
		DefaultName:             &name,
		DefaultDescription:      &desc,
		DefaultManufacturer:     &mfr,
		DefaultModelNumber:      &model,
		DefaultLifetimeWarranty: false,
		DefaultWarrantyDetails:  &warranty,
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
	assert.Equal(t, *data.DefaultQuantity, template.DefaultQuantity)
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

	qty := 5
	defaultName := "Default Item Name"
	defaultDesc := "Default Item Description"
	defaultMfr := "Updated Manufacturer"
	defaultModel := "MODEL-123"
	defaultWarranty := "Lifetime coverage"

	updateData := ItemTemplateUpdate{
		ID:                      template.ID,
		Name:                    "Updated Name",
		Description:             "Updated Description",
		Notes:                   "Updated Notes",
		DefaultQuantity:         &qty,
		DefaultInsured:          true,
		DefaultName:             &defaultName,
		DefaultDescription:      &defaultDesc,
		DefaultManufacturer:     &defaultMfr,
		DefaultModelNumber:      &defaultModel,
		DefaultLifetimeWarranty: true,
		DefaultWarrantyDetails:  &defaultWarranty,
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
	assert.Equal(t, 5, updated.DefaultQuantity)
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
	qty := template.DefaultQuantity
	updateData := ItemTemplateUpdate{
		ID:              template.ID,
		Name:            template.Name,
		Description:     template.Description,
		DefaultQuantity: &qty,
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

func TestItemTemplatesRepository_CreateWithLabels(t *testing.T) {
	// Create some labels
	label1, err := tRepos.Labels.Create(context.Background(), tGroup.ID, LabelCreate{
		Name:        fk.Str(10),
		Description: fk.Str(50),
	})
	require.NoError(t, err)

	label2, err := tRepos.Labels.Create(context.Background(), tGroup.ID, LabelCreate{
		Name:        fk.Str(10),
		Description: fk.Str(50),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tRepos.Labels.delete(context.Background(), label1.ID)
		_ = tRepos.Labels.delete(context.Background(), label2.ID)
	})

	// Create template with labels
	data := templateFactory()
	labelIDs := []uuid.UUID{label1.ID, label2.ID}
	data.DefaultLabelIDs = &labelIDs

	template, err := tRepos.ItemTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)

	assert.Len(t, template.DefaultLabels, 2)

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
	qty := template.DefaultQuantity
	updateData := ItemTemplateUpdate{
		ID:                template.ID,
		Name:              template.Name,
		DefaultQuantity:   &qty,
		DefaultLocationID: uuid.Nil, // Remove location
	}

	updated, err := tRepos.ItemTemplates.Update(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)

	assert.Nil(t, updated.DefaultLocation)

	// Cleanup
	err = tRepos.ItemTemplates.Delete(context.Background(), tGroup.ID, template.ID)
	require.NoError(t, err)
}
