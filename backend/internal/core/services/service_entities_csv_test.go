package services

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
)

func TestEntityService_CsvImportParentImportRef(t *testing.T) {
	ctx := context.Background()

	g, err := tRepos.Groups.GroupCreate(ctx, "csv-parent-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	data := strings.NewReader(strings.Join([]string{
		"HB.location,HB.import_ref,HB.parent_import_ref,HB.name,HB.quantity",
		"Garage,child,parent,Child,1",
		"Garage,parent,,Parent,1",
		"",
	}, "\n"))

	imported, err := tSvc.Entities.CsvImport(ctx, g.ID, data)
	require.NoError(t, err)
	assert.Equal(t, 2, imported)

	parentItem, err := tClient.Entity.Query().
		Where(entity.HasGroupWith(group.ID(g.ID)), entity.Name("Parent")).
		Only(ctx)
	require.NoError(t, err)

	childItem, err := tClient.Entity.Query().
		Where(entity.HasGroupWith(group.ID(g.ID)), entity.Name("Child")).
		Only(ctx)
	require.NoError(t, err)

	childParent, err := childItem.QueryParent().Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, parentItem.ID, childParent.ID)

	itemLocation, err := parentItem.QueryParent().Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Garage", itemLocation.Name)
}
