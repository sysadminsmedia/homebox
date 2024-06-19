package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

func TestAuthTokenRepo_CreateToken(t *testing.T) {
	ctx := context.Background()
	user := userFactory()

	userOut, err := tRepos.Users.Create(ctx, user)
	require.NoError(t, err)

	expiresAt := time.Now().Add(time.Hour)

	generatedToken := hasher.GenerateToken()

	token, err := tRepos.AuthTokens.CreateToken(ctx, UserAuthTokenCreate{
		TokenHash: generatedToken.Hash,
		ExpiresAt: expiresAt,
		UserID:    userOut.ID,
	})

	require.NoError(t, err)
	assert.Equal(t, userOut.ID, token.UserID)
	assert.Equal(t, expiresAt, token.ExpiresAt)

	// Cleanup
	require.NoError(t, tRepos.Users.Delete(ctx, userOut.ID))
	_, err = tRepos.AuthTokens.DeleteAll(ctx)
	require.NoError(t, err)
}

func TestAuthTokenRepo_DeleteToken(t *testing.T) {
	ctx := context.Background()
	user := userFactory()

	userOut, err := tRepos.Users.Create(ctx, user)
	require.NoError(t, err)

	expiresAt := time.Now().Add(time.Hour)

	generatedToken := hasher.GenerateToken()

	_, err = tRepos.AuthTokens.CreateToken(ctx, UserAuthTokenCreate{
		TokenHash: generatedToken.Hash,
		ExpiresAt: expiresAt,
		UserID:    userOut.ID,
	})
	require.NoError(t, err)

	// Delete token
	err = tRepos.AuthTokens.DeleteToken(ctx, []byte(generatedToken.Raw))
	require.NoError(t, err)
}

func TestAuthTokenRepo_GetUserByToken(t *testing.T) {
	ctx := context.Background()

	user := userFactory()
	userOut, _ := tRepos.Users.Create(ctx, user)

	expiresAt := time.Now().Add(time.Hour)
	generatedToken := hasher.GenerateToken()

	token, err := tRepos.AuthTokens.CreateToken(ctx, UserAuthTokenCreate{
		TokenHash: generatedToken.Hash,
		ExpiresAt: expiresAt,
		UserID:    userOut.ID,
	})

	require.NoError(t, err)

	// Get User from token
	foundUser, err := tRepos.AuthTokens.GetUserFromToken(ctx, token.TokenHash)

	require.NoError(t, err)
	assert.Equal(t, userOut.ID, foundUser.ID)
	assert.Equal(t, userOut.Name, foundUser.Name)
	assert.Equal(t, userOut.Email, foundUser.Email)

	// Cleanup
	require.NoError(t, tRepos.Users.Delete(ctx, userOut.ID))
	_, err = tRepos.AuthTokens.DeleteAll(ctx)
	require.NoError(t, err)
}

func TestAuthTokenRepo_PurgeExpiredTokens(t *testing.T) {
	ctx := context.Background()

	user := userFactory()
	userOut, _ := tRepos.Users.Create(ctx, user)

	createdTokens := []UserAuthToken{}

	for i := 0; i < 5; i++ {
		expiresAt := time.Now()
		generatedToken := hasher.GenerateToken()

		createdToken, err := tRepos.AuthTokens.CreateToken(ctx, UserAuthTokenCreate{
			TokenHash: generatedToken.Hash,
			ExpiresAt: expiresAt,
			UserID:    userOut.ID,
		})

		require.NoError(t, err)
		assert.NotNil(t, createdToken)

		createdTokens = append(createdTokens, createdToken)
	}

	// Purge expired tokens
	tokensDeleted, err := tRepos.AuthTokens.PurgeExpiredTokens(ctx)

	require.NoError(t, err)
	assert.Equal(t, 5, tokensDeleted)

	// Check if tokens are deleted
	for _, token := range createdTokens {
		_, err := tRepos.AuthTokens.GetUserFromToken(ctx, token.TokenHash)
		require.Error(t, err)
	}

	// Cleanup
	require.NoError(t, tRepos.Users.Delete(ctx, userOut.ID))
	_, err = tRepos.AuthTokens.DeleteAll(ctx)
	require.NoError(t, err)
}
