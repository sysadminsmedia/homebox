package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// ownershipFixture builds a collection with a real owner plus an invited member
// holding the ordinary per-collection role, mirroring the shape produced by
// accepting an invitation.
type ownershipFixture struct {
	group     repo.Group
	ownerCtx  Context
	memberCtx Context
}

func newOwnershipFixture(t *testing.T) ownershipFixture {
	t.Helper()

	ctx := context.Background()

	group, err := tRepos.Groups.GroupCreate(ctx, fk.Str(10), uuid.Nil)
	require.NoError(t, err)

	owner, err := tRepos.Users.Create(ctx, repo.UserCreate{
		Name:           fk.Str(10),
		Email:          fk.Email(),
		Password:       new(fk.Str(10)),
		DefaultGroupID: group.ID,
		IsOwner:        true,
	})
	require.NoError(t, err)

	member, err := tRepos.Users.Create(ctx, repo.UserCreate{
		Name:           fk.Str(10),
		Email:          fk.Email(),
		Password:       new(fk.Str(10)),
		DefaultGroupID: group.ID,
		IsOwner:        false,
	})
	require.NoError(t, err)

	isOwner, err := tRepos.Groups.IsOwnerOf(ctx, owner.ID, group.ID)
	require.NoError(t, err)
	require.True(t, isOwner, "fixture owner must hold role=owner")

	isOwner, err = tRepos.Groups.IsOwnerOf(ctx, member.ID, group.ID)
	require.NoError(t, err)
	require.False(t, isOwner, "fixture member must not hold role=owner")

	return ownershipFixture{
		group:     group,
		ownerCtx:  Context{Context: ctx, GID: group.ID, UID: owner.ID},
		memberCtx: Context{Context: ctx, GID: group.ID, UID: member.ID},
	}
}

// assertForbidden asserts the error is a 403 rather than a generic failure, so
// a member is told they lack permission instead of the action half-succeeding.
func assertForbidden(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)

	var reqErr *validate.RequestError
	require.ErrorAs(t, err, &reqErr, "expected a validate.RequestError, got %T: %v", err, err)
	assert.Equal(t, http.StatusForbidden, reqErr.Status)
	assert.Equal(t, ErrNotGroupOwner.Error(), reqErr.Error())
}

func TestGroupService_UpdateGroup_MemberForbidden(t *testing.T) {
	f := newOwnershipFixture(t)

	_, err := tSvc.Group.UpdateGroup(f.memberCtx, repo.GroupUpdate{
		Name:     "member-controlled",
		Currency: "eur",
	})
	assertForbidden(t, err)

	// The collection must be untouched by the rejected request.
	after, err := tRepos.Groups.GroupByID(f.memberCtx, f.group.ID)
	require.NoError(t, err)
	assert.Equal(t, f.group.Name, after.Name)
	assert.Equal(t, f.group.Currency, after.Currency)
}

func TestGroupService_UpdateGroup_OwnerAllowed(t *testing.T) {
	f := newOwnershipFixture(t)

	updated, err := tSvc.Group.UpdateGroup(f.ownerCtx, repo.GroupUpdate{
		Name:     "owner-renamed",
		Currency: "eur",
	})
	require.NoError(t, err)
	assert.Equal(t, "owner-renamed", updated.Name)
	assert.Equal(t, "EUR", updated.Currency)
}

func TestGroupService_NewInvitation_MemberForbidden(t *testing.T) {
	f := newOwnershipFixture(t)

	_, token, err := tSvc.Group.NewInvitation(f.memberCtx, 1, time.Now().Add(time.Hour))
	assertForbidden(t, err)
	assert.Empty(t, token, "no invitation token may be minted for a non-owner")

	invitations, err := tRepos.Groups.InvitationGetAll(f.ownerCtx, f.group.ID)
	require.NoError(t, err)
	assert.Empty(t, invitations)
}

func TestGroupService_NewInvitation_OwnerAllowed(t *testing.T) {
	f := newOwnershipFixture(t)

	invitation, token, err := tSvc.Group.NewInvitation(f.ownerCtx, 1, time.Now().Add(time.Hour))
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, 1, invitation.Uses)
}

func TestGroupService_DeleteInvitation_MemberForbidden(t *testing.T) {
	f := newOwnershipFixture(t)

	invitation, _, err := tSvc.Group.NewInvitation(f.ownerCtx, 1, time.Now().Add(time.Hour))
	require.NoError(t, err)

	err = tSvc.Group.DeleteInvitation(f.memberCtx, invitation.ID)
	assertForbidden(t, err)

	// The owner's invitation must survive the revoke attempt.
	invitations, err := tRepos.Groups.InvitationGetAll(f.ownerCtx, f.group.ID)
	require.NoError(t, err)
	require.Len(t, invitations, 1)
	assert.Equal(t, invitation.ID, invitations[0].ID)
}

func TestGroupService_DeleteInvitation_OwnerAllowed(t *testing.T) {
	f := newOwnershipFixture(t)

	invitation, _, err := tSvc.Group.NewInvitation(f.ownerCtx, 1, time.Now().Add(time.Hour))
	require.NoError(t, err)

	require.NoError(t, tSvc.Group.DeleteInvitation(f.ownerCtx, invitation.ID))

	invitations, err := tRepos.Groups.InvitationGetAll(f.ownerCtx, f.group.ID)
	require.NoError(t, err)
	assert.Empty(t, invitations)
}

func TestGroupService_RemoveMember_MemberForbidden(t *testing.T) {
	f := newOwnershipFixture(t)

	err := tSvc.Group.RemoveMember(f.memberCtx, f.ownerCtx.UID)
	assertForbidden(t, err)

	// The owner must still be a member, and still the owner.
	isOwner, err := tRepos.Groups.IsOwnerOf(f.ownerCtx, f.ownerCtx.UID, f.group.ID)
	require.NoError(t, err)
	assert.True(t, isOwner)
}

func TestGroupService_DeleteGroup_MemberForbidden(t *testing.T) {
	f := newOwnershipFixture(t)

	err := tSvc.Group.DeleteGroup(f.memberCtx)
	assertForbidden(t, err)

	// The collection must still exist.
	after, err := tRepos.Groups.GroupByID(f.memberCtx, f.group.ID)
	require.NoError(t, err)
	assert.Equal(t, f.group.ID, after.ID)
}

func TestGroupService_DeleteGroup_OwnerAllowed(t *testing.T) {
	f := newOwnershipFixture(t)

	require.NoError(t, tSvc.Group.DeleteGroup(f.ownerCtx))

	_, err := tRepos.Groups.GroupByID(f.ownerCtx, f.group.ID)
	require.Error(t, err)
}
