package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tagFactory() TagCreate {
	return TagCreate{
		Name:        fk.Str(10),
		Description: fk.Str(100),
	}
}

func useTags(t *testing.T, len int) []TagOut {
	t.Helper()

	tags := make([]TagOut, len)
	for i := 0; i < len; i++ {
		itm := tagFactory()

		item, err := tRepos.Tags.Create(context.Background(), tGroup.ID, itm)
		require.NoError(t, err)
		tags[i] = item
	}

	t.Cleanup(func() {
		for _, item := range tags {
			_ = tRepos.Tags.delete(context.Background(), item.ID)
		}
	})

	return tags
}

func TestTagRepository_Get(t *testing.T) {
	tags := useTags(t, 1)
	tag := tags[0]

	// Get by ID
	foundLoc, err := tRepos.Tags.GetOne(context.Background(), tag.ID)
	require.NoError(t, err)
	assert.Equal(t, tag.ID, foundLoc.ID)
}

func TestTagRepositoryGetAll(t *testing.T) {
	useTags(t, 10)

	all, err := tRepos.Tags.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)
	assert.Len(t, all, 10)
}

func TestTagRepository_Create(t *testing.T) {
	loc, err := tRepos.Tags.Create(context.Background(), tGroup.ID, tagFactory())
	require.NoError(t, err)

	// Get by ID
	foundLoc, err := tRepos.Tags.GetOne(context.Background(), loc.ID)
	require.NoError(t, err)
	assert.Equal(t, loc.ID, foundLoc.ID)

	err = tRepos.Tags.delete(context.Background(), loc.ID)
	require.NoError(t, err)
}

func TestTagRepository_Update(t *testing.T) {
	loc, err := tRepos.Tags.Create(context.Background(), tGroup.ID, tagFactory())
	require.NoError(t, err)

	updateData := TagUpdate{
		ID:          loc.ID,
		Name:        fk.Str(10),
		Description: fk.Str(100),
	}

	update, err := tRepos.Tags.UpdateByGroup(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)

	foundLoc, err := tRepos.Tags.GetOne(context.Background(), loc.ID)
	require.NoError(t, err)

	assert.Equal(t, update.ID, foundLoc.ID)
	assert.Equal(t, update.Name, foundLoc.Name)
	assert.Equal(t, update.Description, foundLoc.Description)

	err = tRepos.Tags.delete(context.Background(), loc.ID)
	require.NoError(t, err)
}

func TestTagRepository_Delete(t *testing.T) {
	loc, err := tRepos.Tags.Create(context.Background(), tGroup.ID, tagFactory())
	require.NoError(t, err)

	err = tRepos.Tags.delete(context.Background(), loc.ID)
	require.NoError(t, err)

	_, err = tRepos.Tags.GetOne(context.Background(), loc.ID)
	require.Error(t, err)
}
