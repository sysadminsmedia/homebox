package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/usergroup"
)

// assertHasUsableTenant asserts the user is still a member of their default
// group, which is exactly what mwTenant requires for every authenticated
// route (including logout and self-deletion). It returns that default group.
func assertHasUsableTenant(t *testing.T, ctx Context, userID uuid.UUID) uuid.UUID {
	t.Helper()

	u, err := tRepos.Users.GetOneID(ctx, userID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, u.DefaultGroupID, "user must keep a default group")
	require.Contains(t, u.GroupIDs, u.DefaultGroupID, "user must be a member of their default group")

	return u.DefaultGroupID
}

func TestGroupService_RemoveMember_OnlyGroupGetsPersonalCollection(t *testing.T) {
	f := newOwnershipFixture(t)

	require.NoError(t, tSvc.Group.RemoveMember(f.ownerCtx, f.memberCtx.UID))

	gid := assertHasUsableTenant(t, f.ownerCtx, f.memberCtx.UID)
	assert.NotEqual(t, f.group.ID, gid)

	// The removed user owns the new collection, so they can use and delete it.
	isOwner, err := tRepos.Groups.IsOwnerOf(f.ownerCtx, f.memberCtx.UID, gid)
	require.NoError(t, err)
	assert.True(t, isOwner)

	// And they are really out of the owner's collection.
	isMember, err := tRepos.Groups.IsMember(f.ownerCtx, f.group.ID, f.memberCtx.UID)
	require.NoError(t, err)
	assert.False(t, isMember)
}

func TestGroupService_RemoveMember_MultiGroupMemberMovesToExistingGroup(t *testing.T) {
	f := newOwnershipFixture(t)

	other, err := tRepos.Groups.GroupCreate(f.ownerCtx, fk.Str(10), f.memberCtx.UID)
	require.NoError(t, err)

	require.NoError(t, tSvc.Group.RemoveMember(f.ownerCtx, f.memberCtx.UID))

	gid := assertHasUsableTenant(t, f.ownerCtx, f.memberCtx.UID)
	assert.Equal(t, other.ID, gid)

	count, err := tClient.UserGroup.Query().Where(usergroup.UserID(f.memberCtx.UID)).Count(f.ownerCtx)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "no extra collection may be created when one remains")
}

func TestGroupService_RemoveMember_NonMemberNotFound(t *testing.T) {
	f := newOwnershipFixture(t)

	// Removing a user who isn't a member must fail rather than spawn a
	// collection for them.
	err := tSvc.Group.RemoveMember(f.ownerCtx, uuid.New())
	require.Error(t, err)
}

func TestGroupService_DeleteGroup_OrphanedMembersKeepATenant(t *testing.T) {
	f := newOwnershipFixture(t)

	// A second member who also belongs to another collection.
	multi := newOwnershipFixture(t)
	require.NoError(t, tClient.UserGroup.Create().
		SetUserID(multi.memberCtx.UID).
		SetGroupID(f.group.ID).
		SetRole(usergroup.RoleUser).
		Exec(f.ownerCtx))
	require.NoError(t, tRepos.Users.UpdateDefaultGroup(f.ownerCtx, multi.memberCtx.UID, f.group.ID))

	require.NoError(t, tSvc.Group.DeleteGroup(f.ownerCtx))

	// Single-group member and the owner each get a personal collection.
	for _, uid := range []uuid.UUID{f.memberCtx.UID, f.ownerCtx.UID} {
		gid := assertHasUsableTenant(t, f.ownerCtx, uid)
		assert.NotEqual(t, f.group.ID, gid)
	}

	// The multi-group member falls back to the collection they already had,
	// rather than being left with a NULL default by the FK.
	gid := assertHasUsableTenant(t, f.ownerCtx, multi.memberCtx.UID)
	assert.Equal(t, multi.group.ID, gid)
}
