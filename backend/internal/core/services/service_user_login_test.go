package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogin_EndToEnd verifies that after the account-enumeration timing fix,
// legitimate login still works and every rejection path returns the same generic
// error, regardless of whether the account exists.
func TestLogin_EndToEnd(t *testing.T) {
	ctx := context.Background()

	const password = "correct-horse-battery-staple"
	reg := UserRegistration{
		Name:     fk.Str(8),
		Email:    fk.Email(),
		Password: password,
	}
	_, err := tSvc.User.RegisterUser(ctx, reg)
	require.NoError(t, err)

	t.Run("valid credentials succeed", func(t *testing.T) {
		tok, err := tSvc.User.Login(ctx, reg.Email, password, false)
		require.NoError(t, err)
		assert.NotEmpty(t, tok.Raw)
	})

	t.Run("wrong password is rejected with generic error", func(t *testing.T) {
		_, err := tSvc.User.Login(ctx, reg.Email, "wrong-password", false)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrorInvalidLogin)
	})

	t.Run("nonexistent user is rejected with the same generic error", func(t *testing.T) {
		_, err := tSvc.User.Login(ctx, "does-not-exist@example.test", "any-password", false)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrorInvalidLogin)
	})
}
