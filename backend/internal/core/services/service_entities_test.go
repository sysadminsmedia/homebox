package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

// TestEntityService_CsvImport_AssetIDIdempotent verifies that re-importing the
// same CSV file (rows matched by their existing import ref) does not consume a
// new auto-incremented asset ID on every run.
//
// Regression test for https://github.com/sysadminsmedia/homebox/issues/1100
// ("Importing a File with import_ref increments asset-id"): the first import
// assigns an asset ID, and subsequent imports of the unchanged file must leave
// that asset ID untouched.
func TestEntityService_CsvImport_AssetIDIdempotent(t *testing.T) {
	ctx := context.Background()

	grp, err := tRepos.Groups.GroupCreate(ctx, "csv-assetid-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	csv := "HB.import_ref,HB.location,HB.name,HB.quantity\n" +
		"ref-1,Loc A,Widget,1\n"

	// First import: creates the item and assigns an auto-incremented asset ID.
	n, err := tSvc.Entities.CsvImport(ctx, grp.ID, strings.NewReader(csv))
	require.NoError(t, err)
	require.Equal(t, 1, n)

	first, err := tRepos.Entities.GetByRef(ctx, grp.ID, "ref-1")
	require.NoError(t, err)
	require.False(t, first.AssetID.Nil(), "first import should assign an asset ID")

	// Re-import the exact same file. The row is matched by its import ref and
	// updated in place, so the asset ID must remain stable.
	n, err = tSvc.Entities.CsvImport(ctx, grp.ID, strings.NewReader(csv))
	require.NoError(t, err)
	require.Equal(t, 1, n)

	second, err := tRepos.Entities.GetByRef(ctx, grp.ID, "ref-1")
	require.NoError(t, err)

	require.Equal(t, second.ID, first.ID, "re-import must update the same entity")
	require.Equal(t, first.AssetID, second.AssetID, "re-import must not change the asset ID")
}

// TestEntityService_CsvImport_AssetIDAutoIncrement guards the auto-increment
// path that issue #1100's fix must preserve: genuinely new items imported in a
// single file still receive distinct, sequential asset IDs.
func TestEntityService_CsvImport_AssetIDAutoIncrement(t *testing.T) {
	ctx := context.Background()

	grp, err := tRepos.Groups.GroupCreate(ctx, "csv-assetid-inc-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	csv := "HB.import_ref,HB.location,HB.name,HB.quantity\n" +
		"ref-a,Loc A,Widget A,1\n" +
		"ref-b,Loc A,Widget B,1\n"

	n, err := tSvc.Entities.CsvImport(ctx, grp.ID, strings.NewReader(csv))
	require.NoError(t, err)
	require.Equal(t, 2, n)

	a, err := tRepos.Entities.GetByRef(ctx, grp.ID, "ref-a")
	require.NoError(t, err)
	b, err := tRepos.Entities.GetByRef(ctx, grp.ID, "ref-b")
	require.NoError(t, err)

	require.False(t, a.AssetID.Nil())
	require.False(t, b.AssetID.Nil())
	require.NotEqual(t, a.AssetID, b.AssetID, "distinct new items must get distinct asset IDs")
}

// TestEntityService_CsvImport_ReimportExportedInventoryWithImportRefs verifies
// that exporting an inventory after EnsureImportRef (which includes top-level
// location rows with an empty HB.location) and re-importing the unmodified CSV
// succeeds and updates entities in place without failing to create a location (#1856).
func TestEntityService_CsvImport_ReimportExportedInventoryWithImportRefs(t *testing.T) {
	ctx := context.Background()

	grp, err := tRepos.Groups.GroupCreate(ctx, "csv-reimport-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	topLoc, err := tRepos.Entities.CreateContainer(ctx, grp.ID, repo.EntityCreate{
		Name: "Kitchen",
	})
	require.NoError(t, err)

	subLoc, err := tRepos.Entities.CreateContainer(ctx, grp.ID, repo.EntityCreate{
		Name:     "Pantry",
		ParentID: topLoc.ID,
	})
	require.NoError(t, err)

	item, err := tRepos.Entities.Create(ctx, grp.ID, repo.EntityCreate{
		Name:        "Olive Oil",
		Description: "extra virgin",
		ParentID:    subLoc.ID,
		Quantity:    1,
	})
	require.NoError(t, err)

	_, err = tSvc.Entities.EnsureImportRef(ctx, grp.ID)
	require.NoError(t, err)

	rows, err := tSvc.Entities.ExportCSV(ctx, grp.ID, "https://homebox.example")
	require.NoError(t, err)
	require.Len(t, rows, 4) // header + topLoc + subLoc + item

	var csvBuf bytes.Buffer
	writer := csv.NewWriter(&csvBuf)
	require.NoError(t, writer.WriteAll(rows))
	require.NoError(t, writer.Error())

	n, err := tSvc.Entities.CsvImport(ctx, grp.ID, bytes.NewReader(csvBuf.Bytes()))
	require.NoError(t, err)
	require.Equal(t, 3, n)

	allAfter, err := tRepos.Entities.GetAll(ctx, grp.ID)
	require.NoError(t, err)
	require.Len(t, allAfter, 3)

	topAfter, err := tRepos.Entities.GetOneByGroup(ctx, grp.ID, topLoc.ID)
	require.NoError(t, err)
	require.NotNil(t, topAfter.EntityType)
	require.True(t, topAfter.EntityType.IsLocation)
	require.Nil(t, topAfter.Parent)

	itemAfter, err := tRepos.Entities.GetOneByGroup(ctx, grp.ID, item.ID)
	require.NoError(t, err)
	require.NotNil(t, itemAfter.Parent)
	require.Equal(t, subLoc.ID, itemAfter.Parent.ID)
}

// TestEntityService_CsvImport_UpdateByRefWithoutLocationColumn verifies that
// importing a partial CSV matched by HB.import_ref when HB.location is omitted
// does not fail with "failed to create location" and preserves the existing
// entity's parent/location override (#1856).
func TestEntityService_CsvImport_UpdateByRefWithoutLocationColumn(t *testing.T) {
	ctx := context.Background()

	grp, err := tRepos.Groups.GroupCreate(ctx, "csv-partial-update-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	loc, err := tRepos.Entities.CreateContainer(ctx, grp.ID, repo.EntityCreate{
		Name: "Kitchen",
	})
	require.NoError(t, err)

	item, err := tRepos.Entities.Create(ctx, grp.ID, repo.EntityCreate{
		ImportRef:   "f7378d7a",
		Name:        "Vorräte Küche",
		Description: "old description",
		ParentID:    loc.ID,
		Quantity:    1,
	})
	require.NoError(t, err)

	csvData := strings.Join([]string{
		"HB.import_ref,HB.url,HB.name,HB.description",
		"f7378d7a,https://homebox.example/item/" + item.ID.String() + ",Vorräte Küche,über Geschirrspüler",
		"",
	}, "\n")

	n, err := tSvc.Entities.CsvImport(ctx, grp.ID, strings.NewReader(csvData))
	require.NoError(t, err)
	require.Equal(t, 1, n)

	updated, err := tRepos.Entities.GetOneByGroup(ctx, grp.ID, item.ID)
	require.NoError(t, err)
	require.Equal(t, "über Geschirrspüler", updated.Description)
	require.NotNil(t, updated.Parent)
	require.Equal(t, loc.ID, updated.Parent.ID)

	// Explicitly including an empty HB.location column clears the parent location.
	clearLocCSV := strings.Join([]string{
		"HB.import_ref,HB.location,HB.name,HB.description",
		"f7378d7a,,Vorräte Küche,über Geschirrspüler",
		"",
	}, "\n")

	n, err = tSvc.Entities.CsvImport(ctx, grp.ID, strings.NewReader(clearLocCSV))
	require.NoError(t, err)
	require.Equal(t, 1, n)

	cleared, err := tRepos.Entities.GetOneByGroup(ctx, grp.ID, item.ID)
	require.NoError(t, err)
	require.Nil(t, cleared.Parent)
}
