package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func templateFactory() EntityTemplateCreate {
	return EntityTemplateCreate{
		Name:                    fk.Str(10),
		Description:             fk.Str(100),
		Notes:                   fk.Str(50),
		DefaultQuantity:         lo.ToPtr(1.0),
		DefaultInsured:          false,
		DefaultName:             lo.ToPtr(fk.Str(20)),
		DefaultDescription:      lo.ToPtr(fk.Str(50)),
		DefaultManufacturer:     lo.ToPtr(fk.Str(15)),
		DefaultModelNumber:      lo.ToPtr(fk.Str(10)),
		DefaultLifetimeWarranty: false,
		DefaultWarrantyDetails:  lo.ToPtr(""),
		IncludeWarrantyFields:   false,
		IncludePurchaseFields:   false,
		IncludeSoldFields:       false,
		Fields:                  []TemplateField{},
	}
}

func TestEntityTemplatesRepository_Create(t *testing.T) {
	data := templateFactory()

	result, err := tRepos.EntityTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, data.Name, result.Name)
	assert.Equal(t, data.Description, result.Description)

	// Cleanup
	err = tRepos.EntityTemplates.Delete(context.Background(), tGroup.ID, result.ID)
	require.NoError(t, err)
}

func TestEntityTemplatesRepository_GetAll(t *testing.T) {
	data := templateFactory()

	created, err := tRepos.EntityTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)

	results, err := tRepos.EntityTemplates.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)

	found := false
	for _, r := range results {
		if r.ID == created.ID {
			found = true
			assert.Equal(t, data.Name, r.Name)
		}
	}
	assert.True(t, found)

	// Cleanup
	err = tRepos.EntityTemplates.Delete(context.Background(), tGroup.ID, created.ID)
	require.NoError(t, err)
}

func TestEntityTemplatesRepository_GetOne(t *testing.T) {
	data := templateFactory()

	created, err := tRepos.EntityTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)

	result, err := tRepos.EntityTemplates.GetOne(context.Background(), tGroup.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, data.Name, result.Name)
	assert.Equal(t, data.Description, result.Description)

	// Cleanup
	err = tRepos.EntityTemplates.Delete(context.Background(), tGroup.ID, created.ID)
	require.NoError(t, err)
}

func TestEntityTemplatesRepository_Update(t *testing.T) {
	data := templateFactory()

	created, err := tRepos.EntityTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)

	updateData := EntityTemplateUpdate{
		ID:          created.ID,
		Name:        fk.Str(10),
		Description: fk.Str(100),
		Notes:       fk.Str(50),
	}

	result, err := tRepos.EntityTemplates.Update(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, updateData.Name, result.Name)
	assert.Equal(t, updateData.Description, result.Description)

	// Cleanup
	err = tRepos.EntityTemplates.Delete(context.Background(), tGroup.ID, created.ID)
	require.NoError(t, err)
}

func TestEntityTemplatesRepository_Delete(t *testing.T) {
	data := templateFactory()

	created, err := tRepos.EntityTemplates.Create(context.Background(), tGroup.ID, data)
	require.NoError(t, err)

	err = tRepos.EntityTemplates.Delete(context.Background(), tGroup.ID, created.ID)
	require.NoError(t, err)

	_, err = tRepos.EntityTemplates.GetOne(context.Background(), tGroup.ID, created.ID)
	require.Error(t, err)
}
