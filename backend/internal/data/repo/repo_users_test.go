package repo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func userFactory() UserCreate {
	password := fk.Str(10)
	return UserCreate{
		Name:        fk.Str(10),
		Email:       fk.Email(),
		Password:    &password,
		IsSuperuser: fk.Bool(),
		GroupID:     tGroup.ID,
	}
}

func TestUserRepo_GetOneEmail(t *testing.T) {
	assert := assert.New(t)
	user := userFactory()
	ctx := context.Background()

	_, err := tRepos.Users.Create(ctx, user)
	require.NoError(t, err)

	foundUser, err := tRepos.Users.GetOneEmail(ctx, user.Email)

	assert.NotNil(foundUser)
	require.NoError(t, err)
	assert.Equal(user.Email, foundUser.Email)
	assert.Equal(user.Name, foundUser.Name)

	// Cleanup
	err = tRepos.Users.DeleteAll(ctx)
	require.NoError(t, err)
}

func TestUserRepo_GetOneId(t *testing.T) {
	assert := assert.New(t)
	user := userFactory()
	ctx := context.Background()

	userOut, _ := tRepos.Users.Create(ctx, user)
	foundUser, err := tRepos.Users.GetOneID(ctx, userOut.ID)

	assert.NotNil(foundUser)
	require.NoError(t, err)
	assert.Equal(user.Email, foundUser.Email)
	assert.Equal(user.Name, foundUser.Name)

	// Cleanup
	err = tRepos.Users.DeleteAll(ctx)
	require.NoError(t, err)
}

func TestUserRepo_GetAll(t *testing.T) {
	// Setup
	toCreate := []UserCreate{
		userFactory(),
		userFactory(),
		userFactory(),
		userFactory(),
	}

	ctx := context.Background()

	created := []UserOut{}

	for _, usr := range toCreate {
		usrOut, _ := tRepos.Users.Create(ctx, usr)
		created = append(created, usrOut)
	}

	// Validate
	allUsers, err := tRepos.Users.GetAll(ctx)

	require.NoError(t, err)
	assert.Len(t, allUsers, len(created))

	for _, usr := range created {
		for _, usr2 := range allUsers {
			if usr.ID == usr2.ID {
				assert.Equal(t, usr.Email, usr2.Email)

				// Check groups are loaded
				assert.NotNil(t, usr2.GroupID)
			}
		}
	}

	for _, usr := range created {
		_ = tRepos.Users.Delete(ctx, usr.ID)
	}

	// Cleanup
	err = tRepos.Users.DeleteAll(ctx)
	require.NoError(t, err)
}

func TestUserRepo_Update(t *testing.T) {
	user, err := tRepos.Users.Create(context.Background(), userFactory())
	require.NoError(t, err)

	updateData := UserUpdate{
		Name:  fk.Str(10),
		Email: fk.Email(),
	}

	// Update
	err = tRepos.Users.Update(context.Background(), user.ID, updateData)
	require.NoError(t, err)

	// Validate
	updated, err := tRepos.Users.GetOneID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.NotEqual(t, user.Name, updated.Name)
	assert.NotEqual(t, user.Email, updated.Email)
}

func TestUserRepo_Delete(t *testing.T) {
	// Create 10 Users
	for i := 0; i < 10; i++ {
		user := userFactory()
		ctx := context.Background()
		_, _ = tRepos.Users.Create(ctx, user)
	}

	// Delete all
	ctx := context.Background()
	allUsers, _ := tRepos.Users.GetAll(ctx)

	assert.NotEmpty(t, allUsers)
	err := tRepos.Users.DeleteAll(ctx)
	require.NoError(t, err)

	allUsers, _ = tRepos.Users.GetAll(ctx)
	assert.Empty(t, allUsers)
}

func TestUserRepo_GetSuperusers(t *testing.T) {
	// Create 10 Users
	superuser := 0
	users := 0

	for i := 0; i < 10; i++ {
		user := userFactory()
		ctx := context.Background()
		_, _ = tRepos.Users.Create(ctx, user)

		if user.IsSuperuser {
			superuser++
		} else {
			users++
		}
	}

	// Delete all
	ctx := context.Background()

	superUsers, err := tRepos.Users.GetSuperusers(ctx)
	require.NoError(t, err)

	for _, usr := range superUsers {
		assert.True(t, usr.IsSuperuser)
	}

	// Cleanup
	err = tRepos.Users.DeleteAll(ctx)
	require.NoError(t, err)
}
