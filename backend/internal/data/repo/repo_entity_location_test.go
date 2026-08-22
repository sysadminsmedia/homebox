package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// locationFixture gives each test its own group with two locations, so tests
// don't see each other's tree.
type locationFixture struct {
	gid      uuid.UUID
	itemType uuid.UUID
	attic    EntityOut
	basement EntityOut
}

func newLocationFixture(t *testing.T, name string) locationFixture {
	t.Helper()
	ctx := context.Background()

	g, err := tRepos.Groups.GroupCreate(ctx, name, uuid.Nil)
	require.NoError(t, err)

	itemType, err := tRepos.EntityTypes.GetDefault(ctx, g.ID, false)
	require.NoError(t, err)
	locType, err := tRepos.EntityTypes.GetDefault(ctx, g.ID, true)
	require.NoError(t, err)

	attic, err := tRepos.Entities.Create(ctx, g.ID, EntityCreate{
		Name:         "Attic",
		EntityTypeID: locType.ID,
	})
	require.NoError(t, err)

	basement, err := tRepos.Entities.Create(ctx, g.ID, EntityCreate{
		Name:         "Basement",
		EntityTypeID: locType.ID,
	})
	require.NoError(t, err)

	return locationFixture{gid: g.ID, itemType: itemType.ID, attic: attic, basement: basement}
}

// #1688 verbatim: a spare part belonging to a product but stored on a different
// shelf. Step 4 used to drag the child into the parent's location, and step 5
// couldn't put it back.
func TestEntityLocation_ChildKeepsOwnLocation(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "loc-child-own-location")

	// 1. Parent in the Attic.
	parent, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name:         "Parent",
		EntityTypeID: f.itemType,
		ParentID:     f.attic.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, parent.Location)
	assert.Equal(t, f.attic.ID, parent.Location.ID)

	// 2. Child in the Basement.
	child, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name:         "Child",
		EntityTypeID: f.itemType,
		ParentID:     f.basement.ID,
	})
	require.NoError(t, err)

	// 4. Make Parent the parent of Child, while keeping Child in the Basement.
	updated, err := tRepos.Entities.UpdateByGroup(ctx, f.gid, EntityUpdate{
		ID:         child.ID,
		Name:       "Child",
		Quantity:   1,
		ParentID:   parent.ID,
		LocationID: f.basement.ID,
	})
	require.NoError(t, err)

	// 5. Both relationships hold at once.
	require.NotNil(t, updated.Parent)
	assert.Equal(t, parent.ID, updated.Parent.ID, "child should be parented to the parent item")
	require.NotNil(t, updated.Location)
	assert.Equal(t, f.basement.ID, updated.Location.ID, "child should still be stored in the Basement")

	// And it survives a re-read rather than only looking right in the response.
	reread, err := tRepos.Entities.GetOneByGroup(ctx, f.gid, child.ID)
	require.NoError(t, err)
	require.NotNil(t, reread.Location)
	assert.Equal(t, f.basement.ID, reread.Location.ID)

	// The parent is unaffected.
	parentAfter, err := tRepos.Entities.GetOneByGroup(ctx, f.gid, parent.ID)
	require.NoError(t, err)
	require.NotNil(t, parentAfter.Location)
	assert.Equal(t, f.attic.ID, parentAfter.Location.ID)
}

// The default is unchanged: no locationId means the child still follows its
// parent. This is what makes the upgrade a no-op for existing rows.
func TestEntityLocation_InheritsWhenNoOverride(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "loc-inherit-default")

	parent, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Parent", EntityTypeID: f.itemType, ParentID: f.attic.ID,
	})
	require.NoError(t, err)

	child, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Child", EntityTypeID: f.itemType, ParentID: parent.ID,
	})
	require.NoError(t, err)

	require.NotNil(t, child.Location)
	assert.Equal(t, f.attic.ID, child.Location.ID, "child with no override inherits the parent's location")

	// Moving the parent moves the inheriting child with it.
	_, err = tRepos.Entities.UpdateByGroup(ctx, f.gid, EntityUpdate{
		ID: parent.ID, Name: "Parent", Quantity: 1, ParentID: f.basement.ID,
	})
	require.NoError(t, err)

	childAfter, err := tRepos.Entities.GetOneByGroup(ctx, f.gid, child.ID)
	require.NoError(t, err)
	require.NotNil(t, childAfter.Location)
	assert.Equal(t, f.basement.ID, childAfter.Location.ID, "inheriting child follows the parent")
}

// Naming a location with no parent item is just a parent change, so no override
// is written and there stays one way to express it.
func TestEntityLocation_DirectLocationStoresNoOverride(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "loc-normalizes-direct")

	item, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Item", EntityTypeID: f.itemType, LocationID: f.attic.ID,
	})
	require.NoError(t, err)

	require.NotNil(t, item.Parent)
	assert.Equal(t, f.attic.ID, item.Parent.ID, "location with no parent item becomes the parent")
	require.NotNil(t, item.Location)
	assert.Equal(t, f.attic.ID, item.Location.ID)

	// No override row was written, so the entity still resolves by ancestry.
	overrideID, err := tClient.Entity.Query().
		Where(entity.ID(item.ID)).
		QueryLocation().
		OnlyID(ctx)
	require.Error(t, err, "no override should be stored for a direct location")
	assert.Equal(t, uuid.Nil, overrideID)
}

// Parent is one location, locationId is another. Both claim to say where this
// lives; silently honoring one is how #1691-class data loss happens.
func TestEntityLocation_RejectsConflictingLocation(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "loc-conflict")

	item, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Item", EntityTypeID: f.itemType, ParentID: f.attic.ID,
	})
	require.NoError(t, err)

	_, err = tRepos.Entities.UpdateByGroup(ctx, f.gid, EntityUpdate{
		ID: item.ID, Name: "Item", Quantity: 1,
		ParentID:   f.attic.ID,
		LocationID: f.basement.ID,
	})
	require.Error(t, err)
	assert.True(t, validate.IsFieldError(err), "should surface as a 422 field error, not a 500")

	// The entity is untouched.
	after, err := tRepos.Entities.GetOneByGroup(ctx, f.gid, item.ID)
	require.NoError(t, err)
	require.NotNil(t, after.Location)
	assert.Equal(t, f.attic.ID, after.Location.ID)
}

// An item can't be used as a location — that would build a second, competing
// containment hierarchy.
func TestEntityLocation_RejectsNonLocationTarget(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "loc-non-location-target")

	parent, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Parent", EntityTypeID: f.itemType, ParentID: f.attic.ID,
	})
	require.NoError(t, err)
	other, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Other", EntityTypeID: f.itemType, ParentID: f.attic.ID,
	})
	require.NoError(t, err)

	_, err = tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Child", EntityTypeID: f.itemType,
		ParentID:   parent.ID,
		LocationID: other.ID,
	})
	require.Error(t, err)
	assert.True(t, validate.IsFieldError(err))
}

// The override must not reach another tenant's location: that corrupts their
// tree and confirms a guessed UUID.
func TestEntityLocation_RejectsCrossGroupLocation(t *testing.T) {
	ctx := context.Background()
	a := newLocationFixture(t, "loc-tenant-a")
	b := newLocationFixture(t, "loc-tenant-b")

	parent, err := tRepos.Entities.Create(ctx, a.gid, EntityCreate{
		Name: "Parent", EntityTypeID: a.itemType, ParentID: a.attic.ID,
	})
	require.NoError(t, err)

	_, err = tRepos.Entities.Create(ctx, a.gid, EntityCreate{
		Name: "Child", EntityTypeID: a.itemType,
		ParentID:   parent.ID,
		LocationID: b.basement.ID,
	})
	require.Error(t, err, "must not accept another group's location")
}

// The "Sync child items' locations" toggle had nothing to act on between 0.26
// and this change. Turning it on drops the children's own locations — and must
// leave their parent edge alone, which is what flattened the tree before #1591.
func TestEntityLocation_SyncChildLocationsClearsOverrides(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "loc-sync-toggle")

	parent, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Parent", EntityTypeID: f.itemType, ParentID: f.attic.ID,
	})
	require.NoError(t, err)

	child, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Child", EntityTypeID: f.itemType,
		ParentID:   parent.ID,
		LocationID: f.basement.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, child.Location)
	require.Equal(t, f.basement.ID, child.Location.ID)

	// Turn the toggle on for the parent.
	_, err = tRepos.Entities.UpdateByGroup(ctx, f.gid, EntityUpdate{
		ID: parent.ID, Name: "Parent", Quantity: 1, ParentID: f.attic.ID,
		SyncChildEntityLocations: true,
	})
	require.NoError(t, err)

	childAfter, err := tRepos.Entities.GetOneByGroup(ctx, f.gid, child.ID)
	require.NoError(t, err)
	require.NotNil(t, childAfter.Location)
	assert.Equal(t, f.attic.ID, childAfter.Location.ID, "child should follow the parent once sync is on")
	require.NotNil(t, childAfter.Parent)
	assert.Equal(t, parent.ID, childAfter.Parent.ID, "sync must not reparent the child (#1591)")
}

// The location page must list the spare part stored in the Basement even though
// it hangs off a product in the Attic.
func TestEntityLocation_QueryByLocationFindsOverriddenChild(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "loc-query-filter")

	parent, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Parent", EntityTypeID: f.itemType, ParentID: f.attic.ID,
	})
	require.NoError(t, err)

	child, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Spare Part", EntityTypeID: f.itemType,
		ParentID:   parent.ID,
		LocationID: f.basement.ID,
	})
	require.NoError(t, err)

	inBasement, err := tRepos.Entities.QueryByGroup(ctx, f.gid, EntityQuery{
		ParentIDs: []uuid.UUID{f.basement.ID},
		Page:      -1,
	})
	require.NoError(t, err)
	ids := make([]uuid.UUID, 0, len(inBasement.Items))
	for _, e := range inBasement.Items {
		ids = append(ids, e.ID)
	}
	assert.Contains(t, ids, child.ID, "Basement should list the item stored there")

	inAttic, err := tRepos.Entities.QueryByGroup(ctx, f.gid, EntityQuery{
		ParentIDs: []uuid.UUID{f.attic.ID},
		Page:      -1,
	})
	require.NoError(t, err)
	atticIDs := make([]uuid.UUID, 0, len(inAttic.Items))
	for _, e := range inAttic.Items {
		atticIDs = append(atticIDs, e.ID)
	}
	assert.Contains(t, atticIDs, parent.ID)
	assert.NotContains(t, atticIDs, child.ID, "Attic must not list an item stored in the Basement")
}

// The recursive tree SQL has to honor the override in both arms. The spare part
// belongs under the Basement, exactly once — handling only one arm lists it
// twice, which is worse than the bug we're fixing.
func TestEntityLocation_TreePlacesOverriddenChildUnderItsLocation(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "loc-tree-placement")

	parent, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Product", EntityTypeID: f.itemType, ParentID: f.attic.ID,
	})
	require.NoError(t, err)

	// Stored in the Basement despite belonging to the product in the Attic.
	spare, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Spare Part", EntityTypeID: f.itemType,
		ParentID:   parent.ID,
		LocationID: f.basement.ID,
	})
	require.NoError(t, err)

	// A sibling with no override still nests under the product.
	bundled, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Bundled Part", EntityTypeID: f.itemType, ParentID: parent.ID,
	})
	require.NoError(t, err)

	tree, err := tRepos.Entities.Tree(ctx, f.gid, TreeQuery{WithItems: true})
	require.NoError(t, err)

	// Flatten to id -> parent name, counting appearances.
	parentOf := map[uuid.UUID]string{}
	seen := map[uuid.UUID]int{}
	var walk func(nodes []*TreeItem, parentName string)
	walk = func(nodes []*TreeItem, parentName string) {
		for _, n := range nodes {
			seen[n.ID]++
			parentOf[n.ID] = parentName
			walk(n.Children, n.Name)
		}
	}
	roots := make([]*TreeItem, 0, len(tree))
	for i := range tree {
		roots = append(roots, &tree[i])
	}
	walk(roots, "")

	assert.Equal(t, 1, seen[spare.ID], "overridden item must appear exactly once")
	assert.Equal(t, "Basement", parentOf[spare.ID], "overridden item belongs under the location it is stored in")
	assert.Equal(t, "Product", parentOf[bundled.ID], "an inheriting sibling still nests under the parent item")
	assert.Equal(t, "Attic", parentOf[parent.ID])
}

// #1691's hazard applied to the new field: GET a body, PUT it straight back, and
// the entity must not move. An inherited location must not become a pinned
// override, and an override must not be dropped.
func TestEntityLocation_RoundTripPreservesPlacement(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "loc-roundtrip")

	parent, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Parent", EntityTypeID: f.itemType, ParentID: f.attic.ID,
	})
	require.NoError(t, err)

	inheriting, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Inheriting", EntityTypeID: f.itemType, ParentID: parent.ID,
	})
	require.NoError(t, err)
	assert.Nil(t, inheriting.LocationID, "an inherited location exposes no locationId")

	overridden, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Overridden", EntityTypeID: f.itemType,
		ParentID: parent.ID, LocationID: f.basement.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, overridden.LocationID)
	assert.Equal(t, f.basement.ID, *overridden.LocationID)

	// Echo each record back the way a client would.
	roundTrip := func(e EntityOut) EntityOut {
		t.Helper()
		u := EntityUpdate{ID: e.ID, Name: e.Name, Quantity: e.Quantity}
		if e.Parent != nil {
			u.ParentID = e.Parent.ID
		}
		if e.LocationID != nil {
			u.LocationID = *e.LocationID
		}
		out, err := tRepos.Entities.UpdateByGroup(ctx, f.gid, u)
		require.NoError(t, err)
		return out
	}

	afterInherit := roundTrip(inheriting)
	assert.Nil(t, afterInherit.LocationID, "round trip must not pin an inherited location")
	require.NotNil(t, afterInherit.Location)
	assert.Equal(t, f.attic.ID, afterInherit.Location.ID)

	afterOverride := roundTrip(overridden)
	require.NotNil(t, afterOverride.LocationID, "round trip must not drop an override")
	assert.Equal(t, f.basement.ID, *afterOverride.LocationID)

	// The inheriting one still tracks the parent after its round trip.
	_, err = tRepos.Entities.UpdateByGroup(ctx, f.gid, EntityUpdate{
		ID: parent.ID, Name: "Parent", Quantity: 1, ParentID: f.basement.ID,
	})
	require.NoError(t, err)
	reread, err := tRepos.Entities.GetOneByGroup(ctx, f.gid, inheriting.ID)
	require.NoError(t, err)
	require.NotNil(t, reread.Location)
	assert.Equal(t, f.basement.ID, reread.Location.ID)
}

// ON DELETE SET NULL: deleting a location that items only point at must not take
// them with it. They fall back to inheriting through their parent.
func TestEntityLocation_DeletingLocationFallsBackToInheritance(t *testing.T) {
	ctx := context.Background()
	f := newLocationFixture(t, "loc-delete-target")

	parent, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Parent", EntityTypeID: f.itemType, ParentID: f.attic.ID,
	})
	require.NoError(t, err)

	child, err := tRepos.Entities.Create(ctx, f.gid, EntityCreate{
		Name: "Child", EntityTypeID: f.itemType,
		ParentID: parent.ID, LocationID: f.basement.ID,
	})
	require.NoError(t, err)

	require.NoError(t, tRepos.Entities.DeleteByGroup(ctx, f.gid, f.basement.ID))

	survived, err := tRepos.Entities.GetOneByGroup(ctx, f.gid, child.ID)
	require.NoError(t, err, "the item must survive deletion of the location it pointed at")
	assert.Nil(t, survived.LocationID, "the dangling override must be cleared")
	require.NotNil(t, survived.Location)
	assert.Equal(t, f.attic.ID, survived.Location.ID, "it falls back to the parent's location")
	require.NotNil(t, survived.Parent)
	assert.Equal(t, parent.ID, survived.Parent.ID)
}
