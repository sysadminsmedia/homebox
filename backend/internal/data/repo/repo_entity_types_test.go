package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityTypeRepository_GetAllDoesNotCreateDefaults(t *testing.T) {
	ctx := context.Background()
	group, err := tRepos.Groups.GroupCreate(ctx, "entity-types-empty-"+uuid.NewString(), uuid.Nil)
	require.NoError(t, err)

	entityTypes, err := tRepos.EntityTypes.GetAll(ctx, group.ID)
	require.NoError(t, err)
	assert.Empty(t, entityTypes)
}

func TestEntityTypeRepository_EnsureDefaults(t *testing.T) {
	ctx := context.Background()
	group, err := tRepos.Groups.GroupCreate(ctx, "entity-types-defaults-"+uuid.NewString(), uuid.Nil)
	require.NoError(t, err)

	require.NoError(t, tRepos.EntityTypes.EnsureDefaults(ctx, group.ID))

	entityTypes, err := tRepos.EntityTypes.GetAll(ctx, group.ID)
	require.NoError(t, err)
	require.Len(t, entityTypes, 2)

	seen := map[string]bool{}
	for _, entityType := range entityTypes {
		seen[entityType.Name] = entityType.IsLocation
	}
	assert.Contains(t, seen, "Item")
	assert.False(t, seen["Item"])
	assert.Contains(t, seen, "Location")
	assert.True(t, seen["Location"])
}
