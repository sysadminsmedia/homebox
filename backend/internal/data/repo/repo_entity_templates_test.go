package repo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
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

func TestEntityTemplatesRepository_SetDefaultImage(t *testing.T) {
	ctx := context.Background()

	created, err := tRepos.EntityTemplates.Create(ctx, tGroup.ID, templateFactory())
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = tRepos.EntityTemplates.Delete(ctx, tGroup.ID, created.ID)
	})

	result, err := tRepos.EntityTemplates.SetDefaultImage(ctx, tGroup.ID, created.ID, ItemCreateAttachment{
		Title:   "photo.png",
		Content: strings.NewReader("template image"),
	})
	require.NoError(t, err)
	require.NotNil(t, result.DefaultImage)
	assert.Equal(t, "photo.png", result.DefaultImage.Title)

	// The image must survive a round trip through GetOne
	fetched, err := tRepos.EntityTemplates.GetOne(ctx, tGroup.ID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched.DefaultImage)
	assert.Equal(t, result.DefaultImage.ID, fetched.DefaultImage.ID)

	// Replacing the image must not leave the old attachment behind
	old := result.DefaultImage.ID
	replaced, err := tRepos.EntityTemplates.SetDefaultImage(ctx, tGroup.ID, created.ID, ItemCreateAttachment{
		Title:   "replacement.png",
		Content: strings.NewReader("replacement image"),
	})
	require.NoError(t, err)
	require.NotNil(t, replaced.DefaultImage)
	assert.NotEqual(t, old, replaced.DefaultImage.ID)

	_, err = tRepos.Attachments.Get(ctx, tGroup.ID, old)
	require.Error(t, err)
}

func TestEntityTemplatesRepository_ClearDefaultImage(t *testing.T) {
	ctx := context.Background()

	created, err := tRepos.EntityTemplates.Create(ctx, tGroup.ID, templateFactory())
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = tRepos.EntityTemplates.Delete(ctx, tGroup.ID, created.ID)
	})

	withImage, err := tRepos.EntityTemplates.SetDefaultImage(ctx, tGroup.ID, created.ID, ItemCreateAttachment{
		Title:   "photo.png",
		Content: strings.NewReader("template image"),
	})
	require.NoError(t, err)
	require.NotNil(t, withImage.DefaultImage)

	cleared, err := tRepos.EntityTemplates.ClearDefaultImage(ctx, tGroup.ID, created.ID)
	require.NoError(t, err)
	assert.Nil(t, cleared.DefaultImage)

	// Clearing again is a no-op rather than an error
	cleared, err = tRepos.EntityTemplates.ClearDefaultImage(ctx, tGroup.ID, created.ID)
	require.NoError(t, err)
	assert.Nil(t, cleared.DefaultImage)
}

func TestEntityTemplatesRepository_CreateFromTemplateAppliesDefaultImage(t *testing.T) {
	ctx := context.Background()

	created, err := tRepos.EntityTemplates.Create(ctx, tGroup.ID, templateFactory())
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = tRepos.EntityTemplates.Delete(ctx, tGroup.ID, created.ID)
	})

	withImage, err := tRepos.EntityTemplates.SetDefaultImage(ctx, tGroup.ID, created.ID, ItemCreateAttachment{
		Title:   "photo.png",
		Content: strings.NewReader("template image"),
	})
	require.NoError(t, err)
	require.NotNil(t, withImage.DefaultImage)

	templateImage, err := tRepos.EntityTemplates.GetDefaultImage(ctx, tGroup.ID, created.ID)
	require.NoError(t, err)

	entity, err := tRepos.Entities.CreateFromTemplate(ctx, tGroup.ID, EntityCreateFromTemplate{
		TemplateID: created.ID,
		Name:       "From Template",
		Quantity:   1,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = tRepos.Entities.Delete(ctx, entity.ID)
	})

	require.Len(t, entity.Attachments, 1)
	applied := entity.Attachments[0]
	assert.Equal(t, attachment.TypePhoto.String(), applied.Type)
	assert.True(t, applied.Primary)
	assert.Equal(t, templateImage.Title, applied.Title)
	// The file is stored once and shared, not copied per entity
	assert.Equal(t, templateImage.Path, applied.Path)

	// Deleting the entity's copy must not take the shared file with it while
	// the template still points at that path.
	err = tRepos.Attachments.Delete(ctx, tGroup.ID, applied.ID)
	require.NoError(t, err)

	// Blob keys are relative to the bucket root, which the test harness backs
	// with the temp dir.
	stored := filepath.Join(os.TempDir(), tRepos.Attachments.GetFullPath(templateImage.Path))
	_, err = os.Stat(stored)
	require.NoError(t, err, "shared file was deleted while the template still references it")
}

func TestEntityTemplatesRepository_CreateFromTemplateWithoutImage(t *testing.T) {
	ctx := context.Background()

	created, err := tRepos.EntityTemplates.Create(ctx, tGroup.ID, templateFactory())
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = tRepos.EntityTemplates.Delete(ctx, tGroup.ID, created.ID)
	})

	entity, err := tRepos.Entities.CreateFromTemplate(ctx, tGroup.ID, EntityCreateFromTemplate{
		TemplateID: created.ID,
		Name:       "No Image",
		Quantity:   1,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = tRepos.Entities.Delete(ctx, entity.ID)
	})

	assert.Empty(t, entity.Attachments)
}
