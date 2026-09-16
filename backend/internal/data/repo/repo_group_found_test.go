package repo

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/usergroup"
)

// assetIDSeq hands out strictly increasing asset IDs so tests never reuse a
// value. FoundContact data (opted-in groups and their items) persists for
// the lifetime of the test binary, so a hardcoded literal reused across
// -count=N reruns of the same test would collide with data left behind by
// the previous run.
var assetIDSeq int64 = 900000

func nextAssetID() AssetID {
	return AssetID(atomic.AddInt64(&assetIDSeq, 1))
}

// foundContactGroup creates a fresh group with its own dedicated owner user.
//
// The shared tGroup/tUser fixtures cannot be reused here: bootstrap() creates
// tGroup via GroupCreate(ctx, "test-group", uuid.Nil), which skips creating
// any owner membership, and tUser is created via userFactory(), which does
// not set UserCreate.IsOwner. So tUser is never an owner of tGroup, which
// FoundContact's owner lookup requires.
func foundContactGroup(t *testing.T) (Group, UserOut) {
	t.Helper()
	ctx := context.Background()

	g, err := tRepos.Groups.GroupCreate(ctx, "found-contact-"+fk.Str(8), uuid.Nil)
	require.NoError(t, err)

	owner, err := tRepos.Users.Create(ctx, UserCreate{
		Name:           fk.Str(10),
		Email:          fk.Email(),
		Password:       new(fk.Str(10)),
		DefaultGroupID: g.ID,
		IsOwner:        true,
	})
	require.NoError(t, err)

	return g, owner
}

// enableFoundContact opts a group into found-item contact with the given
// message. Name/Currency must be re-supplied because GroupUpdate overwrites
// them unconditionally.
func enableFoundContact(t *testing.T, g Group, message string) {
	t.Helper()
	ctx := context.Background()

	enabled := true
	_, err := tRepos.Groups.GroupUpdate(ctx, g.ID, GroupUpdate{
		Name:                g.Name,
		Currency:            g.Currency,
		FoundContactEnabled: &enabled,
		FoundContactMessage: &message,
	})
	require.NoError(t, err)
}

// createFoundContactItem creates a single item entity in the given group.
func createFoundContactItem(t *testing.T, groupID uuid.UUID) EntityOut {
	t.Helper()
	ctx := context.Background()

	itemET, err := tRepos.EntityTypes.GetDefault(ctx, groupID, false)
	require.NoError(t, err)

	item, err := tRepos.Entities.Create(ctx, groupID, EntityCreate{
		Name:         fk.Str(10),
		Description:  fk.Str(50),
		EntityTypeID: itemET.ID,
	})
	require.NoError(t, err)

	return item
}

// createOwnerAt creates an owner user for groupID with an explicit CreatedAt,
// bypassing the repo layer (which always stamps CreatedAt with time.Now()).
// Relying on two sequential repo calls landing in distinguishably-ordered
// sqlite-stored instants would be flaky, so tests that need a deterministic
// "earliest owner" set the timestamps explicitly instead.
func createOwnerAt(t *testing.T, groupID uuid.UUID, createdAt time.Time) *ent.User {
	t.Helper()
	ctx := context.Background()

	u, err := tClient.User.Create().
		SetName(fk.Str(10)).
		SetEmail(fk.Email()).
		SetIsSuperuser(false).
		SetDefaultGroupID(groupID).
		SetCreatedAt(createdAt).
		Save(ctx)
	require.NoError(t, err)

	_, err = tClient.UserGroup.Create().
		SetUserID(u.ID).
		SetGroupID(groupID).
		SetRole(usergroup.RoleOwner).
		Save(ctx)
	require.NoError(t, err)

	return u
}

func TestGroupRepository_FoundContactByItemID_HappyPath(t *testing.T) {
	ctx := context.Background()

	g, owner := foundContactGroup(t)
	enableFoundContact(t, g, "call me")

	item := createFoundContactItem(t, g.ID)

	fc, err := tRepos.Groups.FoundContactByItemID(ctx, item.ID)
	require.NoError(t, err)

	assert.Equal(t, item.ID, fc.ItemID)
	assert.NotEmpty(t, fc.ItemName)
	assert.Equal(t, "call me", fc.Message)
	assert.Equal(t, owner.Name, fc.OwnerName)
	assert.Equal(t, owner.Email, fc.OwnerEmail)
}

func TestGroupRepository_FoundContactByItemID_NotOptedIn(t *testing.T) {
	ctx := context.Background()

	g, _ := foundContactGroup(t)
	// found-contact left at its default (disabled).

	item := createFoundContactItem(t, g.ID)

	_, err := tRepos.Groups.FoundContactByItemID(ctx, item.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

func TestGroupRepository_FoundContactByItemID_NonExistent(t *testing.T) {
	ctx := context.Background()

	g, _ := foundContactGroup(t)
	enableFoundContact(t, g, "call me")

	_, err := tRepos.Groups.FoundContactByItemID(ctx, uuid.New())
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

func TestGroupRepository_FoundContactByAssetID_HappyPath(t *testing.T) {
	ctx := context.Background()

	g, owner := foundContactGroup(t)
	enableFoundContact(t, g, "call me too")

	item := createFoundContactItem(t, g.ID)
	assetID := nextAssetID()
	require.NoError(t, tClient.Entity.UpdateOneID(item.ID).SetAssetID(int64(assetID)).Exec(ctx))

	fc, err := tRepos.Groups.FoundContactByAssetID(ctx, assetID)
	require.NoError(t, err)

	assert.Equal(t, item.ID, fc.ItemID)
	assert.NotEmpty(t, fc.ItemName)
	assert.Equal(t, "call me too", fc.Message)
	assert.Equal(t, owner.Name, fc.OwnerName)
	assert.Equal(t, owner.Email, fc.OwnerEmail)
}

func TestGroupRepository_FoundContactByAssetID_NoMatch(t *testing.T) {
	ctx := context.Background()

	g, _ := foundContactGroup(t)
	enableFoundContact(t, g, "call me")

	_, err := tRepos.Groups.FoundContactByAssetID(ctx, nextAssetID())
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

func TestGroupRepository_FoundContactByAssetID_Ambiguous(t *testing.T) {
	ctx := context.Background()
	assetID := nextAssetID()

	g1, _ := foundContactGroup(t)
	enableFoundContact(t, g1, "call me")
	item1 := createFoundContactItem(t, g1.ID)
	require.NoError(t, tClient.Entity.UpdateOneID(item1.ID).SetAssetID(int64(assetID)).Exec(ctx))

	g2, _ := foundContactGroup(t)
	enableFoundContact(t, g2, "call me")
	item2 := createFoundContactItem(t, g2.ID)
	require.NoError(t, tClient.Entity.UpdateOneID(item2.ID).SetAssetID(int64(assetID)).Exec(ctx))

	_, err := tRepos.Groups.FoundContactByAssetID(ctx, assetID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// Archived (sold/disposed) items must read as not-found for both lookups,
// even though their group is opted in — the item is no longer something
// anyone should be able to return.
func TestGroupRepository_FoundContact_ExcludesArchivedItems(t *testing.T) {
	ctx := context.Background()

	g, _ := foundContactGroup(t)
	enableFoundContact(t, g, "call me")

	item := createFoundContactItem(t, g.ID)
	assetID := nextAssetID()
	require.NoError(t, tClient.Entity.UpdateOneID(item.ID).SetAssetID(int64(assetID)).Exec(ctx))

	// Confirm both lookups succeed before archiving, so the test is
	// self-contained and doesn't rely on other tests to prove the happy path.
	_, err := tRepos.Groups.FoundContactByItemID(ctx, item.ID)
	require.NoError(t, err)
	_, err = tRepos.Groups.FoundContactByAssetID(ctx, assetID)
	require.NoError(t, err)

	require.NoError(t, tClient.Entity.UpdateOneID(item.ID).SetArchived(true).Exec(ctx))

	_, err = tRepos.Groups.FoundContactByItemID(ctx, item.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	_, err = tRepos.Groups.FoundContactByAssetID(ctx, assetID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// An opted-in group with no owner membership must not leak contact details;
// the owner-resolution query in foundContactFromEntity has to fail closed.
func TestGroupRepository_FoundContact_NoOwner(t *testing.T) {
	ctx := context.Background()

	g, err := tRepos.Groups.GroupCreate(ctx, "found-contact-no-owner-"+fk.Str(8), uuid.Nil)
	require.NoError(t, err)
	enableFoundContact(t, g, "call me")

	item := createFoundContactItem(t, g.ID)
	assetID := nextAssetID()
	require.NoError(t, tClient.Entity.UpdateOneID(item.ID).SetAssetID(int64(assetID)).Exec(ctx))

	_, err = tRepos.Groups.FoundContactByItemID(ctx, item.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	_, err = tRepos.Groups.FoundContactByAssetID(ctx, assetID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// When a group somehow has two owner memberships, the earliest-created owner
// wins deterministically rather than depending on query/storage order.
func TestGroupRepository_FoundContact_MultipleOwners_EarliestWins(t *testing.T) {
	ctx := context.Background()

	g, err := tRepos.Groups.GroupCreate(ctx, "found-contact-multi-owner-"+fk.Str(8), uuid.Nil)
	require.NoError(t, err)
	enableFoundContact(t, g, "call me")

	base := time.Now().Add(-time.Hour)
	earliest := createOwnerAt(t, g.ID, base)
	_ = createOwnerAt(t, g.ID, base.Add(time.Minute))

	item := createFoundContactItem(t, g.ID)

	fc, err := tRepos.Groups.FoundContactByItemID(ctx, item.ID)
	require.NoError(t, err)

	assert.Equal(t, earliest.Name, fc.OwnerName)
	assert.Equal(t, earliest.Email, fc.OwnerEmail)
}
