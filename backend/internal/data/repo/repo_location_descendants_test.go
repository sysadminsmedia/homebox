package repo

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	batteryChargerName = "01 Battery Charger"
	resinCleanerName   = "02 Resin Cleaner"
	filamentDryerName  = "03 Filament Dryer"
)

func makeNestedLocation(t *testing.T, f locationFixture, name string, parentID uuid.UUID) EntityOut {
	t.Helper()
	ct, err := tRepos.EntityTypes.GetDefault(context.Background(), f.gid, true)
	require.NoError(t, err)
	out, err := tRepos.Entities.Create(context.Background(), f.gid, EntityCreate{
		Name: name, EntityTypeID: ct.ID, ParentID: parentID,
	})
	require.NoError(t, err)
	return out
}

func makeNestedItem(t *testing.T, f locationFixture, name string, parentID uuid.UUID, override uuid.UUID) EntityOut {
	t.Helper()
	out, err := tRepos.Entities.Create(context.Background(), f.gid, EntityCreate{
		Name: name, EntityTypeID: f.itemType, ParentID: parentID, LocationID: override,
	})
	require.NoError(t, err)
	return out
}

func TestLocationDescendants_DepthsPathsAndPagination(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "location-depths-and-pagination")

	rootItem := makeNestedItem(t, f, batteryChargerName, f.attic.ID, uuid.Nil)
	shelf := makeNestedLocation(t, f, "Upper Storage Shelf", f.attic.ID)
	one := makeNestedItem(t, f, resinCleanerName, shelf.ID, uuid.Nil)
	box := makeNestedLocation(t, f, "3D Printing Box", shelf.ID)
	two := makeNestedItem(t, f, filamentDryerName, box.ID, uuid.Nil)
	tray := makeNestedLocation(t, f, "Maintenance Tray", box.ID)
	three := makeNestedItem(t, f, "04 Silicone Socks", tray.ID, uuid.Nil)
	compartment := makeNestedLocation(t, f, "Tiny Tray", tray.ID)
	all := makeNestedItem(t, f, "05 Spare Nozzle", compartment.ID, uuid.Nil)

	// The existing direct-only endpoint is deliberately unchanged.
	direct, err := tRepos.Entities.QueryByGroup(ctx, f.gid, EntityQuery{
		ParentIDs: []uuid.UUID{f.attic.ID}, Page: -1, PageSize: -1,
	})
	require.NoError(t, err)
	require.Len(t, direct.Items, 1)
	assert.Equal(t, rootItem.ID, direct.Items[0].ID)

	for _, tc := range []struct {
		depth int
		names []string
	}{
		{1, []string{batteryChargerName, resinCleanerName}},
		{2, []string{batteryChargerName, resinCleanerName, filamentDryerName}},
		{3, []string{batteryChargerName, resinCleanerName, filamentDryerName, "04 Silicone Socks"}},
		{maxAncestorDepth, []string{batteryChargerName, resinCleanerName, filamentDryerName, "04 Silicone Socks", "05 Spare Nozzle"}},
	} {
		got, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, tc.depth, 1, 24)
		require.NoError(t, err)
		names := make([]string, 0, len(got.Items))
		for _, item := range got.Items {
			names = append(names, item.Name)
		}
		assert.Equal(t, tc.names, names, "depth %d", tc.depth)
		assert.Equal(t, len(tc.names), got.Total, "depth %d", tc.depth)
		assert.Equal(t, 1, got.DirectTotal)
	}

	page1, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, maxAncestorDepth, 1, 2)
	require.NoError(t, err)
	page2, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, maxAncestorDepth, 2, 2)
	require.NoError(t, err)
	page3, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, maxAncestorDepth, 3, 2)
	require.NoError(t, err)
	assert.Equal(t, []string{batteryChargerName, resinCleanerName}, []string{page1.Items[0].Name, page1.Items[1].Name})
	assert.Equal(t, []string{filamentDryerName, "04 Silicone Socks"}, []string{page2.Items[0].Name, page2.Items[1].Name})
	assert.Equal(t, all.ID, page3.Items[0].ID)
	assert.Equal(t, 5, page1.Total)
	assert.Equal(t, 5, page2.Total)
	assert.Equal(t, 5, page3.Total)
	assert.Equal(t, "Attic / Upper Storage Shelf / 3D Printing Box", page2.Items[0].LocationPath)
	assert.Equal(t, "Attic / Upper Storage Shelf / 3D Printing Box / Maintenance Tray", page2.Items[1].LocationPath)
	assert.Equal(t, "Attic", page1.Items[0].LocationPath)
	assert.Equal(t, one.ID, page1.Items[1].ID)
	assert.Equal(t, three.ID, page2.Items[1].ID)
	assert.Equal(t, two.ID, page2.Items[0].ID)
}

func TestLocationDescendants_OverridesAndNonLocationContainers(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "location-depth-override")
	shelf := makeNestedLocation(t, f, "Shelf", f.attic.ID)
	basket := makeNestedItem(t, f, "Container Item", f.attic.ID, uuid.Nil)
	nested := makeNestedItem(t, f, "Nested screws", basket.ID, uuid.Nil)

	// Although logically nested inside Container Item, the item is physically
	// in the Basement. It must not appear in an Attic recursive view.
	misplaced := makeNestedItem(t, f, "Other room part", basket.ID, f.basement.ID)

	parentElsewhere := makeNestedItem(t, f, "Machine in Basement", f.basement.ID, uuid.Nil)
	physicallyOnShelf := makeNestedItem(t, f, "Spare nozzle on shelf", parentElsewhere.ID, shelf.ID)

	got, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, 1, 1, 24)
	require.NoError(t, err)
	names := map[string]bool{}
	paths := map[string]string{}
	for _, item := range got.Items {
		names[item.Name] = true
		paths[item.Name] = item.LocationPath
	}
	assert.True(t, names["Container Item"])
	assert.True(t, names["Nested screws"], "item stored inside a non-location container should be included")
	assert.True(t, names["Spare nozzle on shelf"], "item physically on shelf with a parent in another room should be included")
	assert.False(t, names["Other room part"])
	assert.False(t, names["Machine in Basement"])
	assert.Equal(t, "Attic / Shelf", paths["Spare nozzle on shelf"])
	assert.Equal(t, "Attic / Container Item", paths["Nested screws"])
	assert.Equal(t, 3, got.Total)
	assert.Equal(t, 1, got.DirectTotal)
	require.NotEqual(t, nested.ID, misplaced.ID)
	require.NotEqual(t, shelf.ID, f.basement.ID)
	require.NotEqual(t, physicallyOnShelf.ID, misplaced.ID)
}

func TestLocationDescendants_GroupIsolationAndArchived(t *testing.T) {
	ctx := context.Background()
	first := newLocationFixture(t, "location-depth-group1")
	second := newLocationFixture(t, "location-depth-group2")
	shelf := makeNestedLocation(t, first, "Shelf", first.attic.ID)
	item := makeNestedItem(t, first, "Retained", shelf.ID, uuid.Nil)
	archived := makeNestedItem(t, first, "Archived", shelf.ID, uuid.Nil)
	_, err := tClient.Entity.UpdateOneID(archived.ID).SetArchived(true).Save(ctx)
	require.NoError(t, err)
	got, err := tRepos.Entities.QueryLocationDescendants(ctx, first.gid, first.attic.ID, 1, 1, 24)
	require.NoError(t, err)
	require.Equal(t, 1, got.Total)
	assert.Equal(t, item.ID, got.Items[0].ID)
	assert.Equal(t, 0, got.DirectTotal)
	// Never show any data (or match a root) outside the requesting group.
	_, err = tRepos.Entities.QueryLocationDescendants(ctx, second.gid, first.attic.ID, 64, 1, 24)
	require.Error(t, err)
	secondView, err := tRepos.Entities.QueryLocationDescendants(ctx, second.gid, second.attic.ID, 64, 1, 24)
	require.NoError(t, err)
	assert.Zero(t, secondView.Total)
}

func TestLocationDescendants_RejectsInvalidDepthAndPage(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "location-depth-invalid")
	for _, tc := range []struct{ depth, page, size int }{
		{0, 1, 24}, {-1, 1, 24}, {65, 1, 24}, {1, 0, 24},
		{1, 1, 0}, {1, 1, 97},
		{1, int(^uint(0) >> 1), 96},
	} {
		_, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, tc.depth, tc.page, tc.size)
		require.ErrorIs(t, err, ErrInvalidLocationQuery)
	}
	// A normal item cannot be used as the root location.
	item := makeNestedItem(t, f, "Not a location", f.attic.ID, uuid.Nil)
	_, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, item.ID, 1, 1, 24)
	require.ErrorIs(t, err, ErrLocationRootNotLocation)
	// A nonexistent root stays distinct from the invalid-location sentinel.
	_, err = tRepos.Entities.QueryLocationDescendants(ctx, f.gid, uuid.New(), 1, 1, 24)
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrLocationRootNotLocation)
	require.NotErrorIs(t, err, ErrInvalidLocationQuery)
}

// A corrupt two-location parent cycle must not repeat entities or inflate page
// totals. The root has already been visited before traversal begins.
func TestLocationDescendants_CyclicLocations(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "location-depth-cycle")
	shelf := makeNestedLocation(t, f, "Cycle Shelf", f.attic.ID)
	direct := makeNestedItem(t, f, "One Direct Item", f.attic.ID, uuid.Nil)
	nested := makeNestedItem(t, f, "Two Nested Item", shelf.ID, uuid.Nil)

	// Existing entity updates can form a loop. Protect the reader even when
	// that corrupt state exists, without changing ordinary update semantics.
	_, err := tClient.Entity.UpdateOneID(f.attic.ID).SetParentID(shelf.ID).Save(ctx)
	require.NoError(t, err)

	for _, depth := range []int{1, 2, 3, MaxLocationDepth} {
		got, queryErr := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, depth, 1, 24)
		require.NoError(t, queryErr)
		require.Equal(t, 2, got.Total)
		require.Equal(t, 1, got.DirectTotal)
		require.Len(t, got.Items, 2)
		assert.Equal(t, []uuid.UUID{direct.ID, nested.ID}, []uuid.UUID{got.Items[0].ID, got.Items[1].ID})
		assert.Equal(t, "Attic", got.Items[0].LocationPath)
		assert.Equal(t, "Attic / Cycle Shelf", got.Items[1].LocationPath)
	}

	first, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, MaxLocationDepth, 1, 1)
	require.NoError(t, err)
	second, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, MaxLocationDepth, 2, 1)
	require.NoError(t, err)
	require.Len(t, first.Items, 1)
	require.Len(t, second.Items, 1)
	assert.NotEqual(t, first.Items[0].ID, second.Items[0].ID)
	assert.Equal(t, 2, second.Total)
}

// Deeply nested item containers do not increment the location depth, and must
// remain visible even when they exceed the original 64-edge traversal guard.
func TestLocationDescendants_DeepNonLocationContainerChain(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "location-depth-deep-items")

	const containerCount = 70
	parentID := f.attic.ID
	for i := range containerCount {
		container := makeNestedItem(t, f, fmt.Sprintf("Box %03d", i), parentID, uuid.Nil)
		parentID = container.ID
	}
	leaf := makeNestedItem(t, f, "Deepest spare part", parentID, uuid.Nil)

	for _, depth := range []int{1, MaxLocationDepth} {
		got, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, depth, 1, 24)
		require.NoError(t, err)
		require.Equal(t, containerCount+1, got.Total)
		require.Equal(t, 1, got.DirectTotal)
		assert.Len(t, got.Items, 24)
		// The nested leaf sorts after the numbered containers.
		lastPage, err := tRepos.Entities.QueryLocationDescendants(ctx, f.gid, f.attic.ID, depth, 3, 24)
		require.NoError(t, err)
		require.Len(t, lastPage.Items, containerCount+1-2*24)
		assert.Equal(t, leaf.ID, lastPage.Items[len(lastPage.Items)-1].ID)
		assert.Contains(t, lastPage.Items[len(lastPage.Items)-1].LocationPath, "Box 069")
	}
}
