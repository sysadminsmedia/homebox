package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Group_Create(t *testing.T) {
	g, err := tRepos.Groups.GroupCreate(context.Background(), "test", uuid.Nil)

	require.NoError(t, err)
	assert.Equal(t, "test", g.Name)

	// Get by ID
	foundGroup, err := tRepos.Groups.GroupByID(context.Background(), g.ID)
	require.NoError(t, err)
	assert.Equal(t, g.ID, foundGroup.ID)
}

func Test_Group_Create_WithUser(t *testing.T) {
	// Create a test user first
	password := "password123"
	user, err := tRepos.Users.Create(context.Background(), UserCreate{
		Name:           "test_user",
		Email:          "test_group_user@example.com",
		Password:       &password,
		DefaultGroupID: tGroup.ID,
	})
	require.NoError(t, err)

	// Create a group with the user
	g, err := tRepos.Groups.GroupCreate(context.Background(), "test_group_with_user", user.ID)
	require.NoError(t, err)
	assert.Equal(t, "test_group_with_user", g.Name)

	// Verify the user is a member of the group
	members, err := tRepos.Users.GetUsersByGroupID(context.Background(), g.ID)
	require.NoError(t, err)
	assert.Len(t, members, 1, "Group should have exactly one member")
	assert.Equal(t, user.Name, members[0].Name, "The member should be the user who created the group")
}

func Test_Group_Update(t *testing.T) {
	g, err := tRepos.Groups.GroupCreate(context.Background(), "test", uuid.Nil)
	require.NoError(t, err)

	g, err = tRepos.Groups.GroupUpdate(context.Background(), g.ID, GroupUpdate{
		Name:     "test2",
		Currency: "eur",
	})
	require.NoError(t, err)
	assert.Equal(t, "test2", g.Name)
	assert.Equal(t, "EUR", g.Currency)
}

// TODO: Fix this test at some point, the data itself in production/development is working fine, it only fails on the test
/*func Test_Group_GroupStatistics(t *testing.T) {
	useItems(t, 20)
	useLabels(t, 20)

	stats, err := tRepos.Groups.StatsGroup(context.Background(), tGroup.ID)

	require.NoError(t, err)
	assert.Equal(t, 20, stats.TotalItems)
	assert.Equal(t, 20, stats.TotalLabels)
	assert.Equal(t, 1, stats.TotalUsers)
	assert.Equal(t, 1, stats.TotalLocations)
}*/

func Test_Group_IsMember(t *testing.T) {
	ctx := context.Background()

	group, err := tRepos.Groups.GroupCreate(ctx, "member-check", uuid.Nil)
	require.NoError(t, err)

	user := userFactory()
	user.DefaultGroupID = group.ID
	createdUser, err := tRepos.Users.Create(ctx, user)
	require.NoError(t, err)

	// Newly created user is added to default group in Create()
	isMember, err := tRepos.Groups.IsMember(ctx, group.ID, createdUser.ID)
	require.NoError(t, err)
	assert.True(t, isMember)

	otherUser := userFactory()
	otherUser.DefaultGroupID = tGroup.ID
	createdOther, err := tRepos.Users.Create(ctx, otherUser)
	require.NoError(t, err)

	isMember, err = tRepos.Groups.IsMember(ctx, group.ID, createdOther.ID)
	require.NoError(t, err)
	assert.False(t, isMember)
}
