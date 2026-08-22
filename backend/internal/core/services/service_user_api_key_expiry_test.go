package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

// API key hashing is peppered and panics if the pepper was never configured
// (see hasher.HashAPIKey); the app sets it at startup, so tests must too.
func init() {
	hasher.SetAPIKeyPepper([]byte("test-api-key-pepper"))
}

// TestCreateAPIKey_DefaultsToBoundedLifetime pins the default API key TTL. A
// password change revokes sessions but intentionally leaves API keys alone, so
// a key created without an explicit expiry must not live forever — otherwise a
// leaked key stays valid for the life of the account with no lifecycle event
// that would ever retire it.
func TestCreateAPIKey_DefaultsToBoundedLifetime(t *testing.T) {
	ctx := context.Background()
	usr := newTestUserWithPassword(t, "api-key-default-ttl")

	before := time.Now()
	out, err := tSvc.User.CreateAPIKey(ctx, usr.ID, repo.APIKeyCreate{Name: "no-expiry-requested"})
	require.NoError(t, err)

	require.NotNil(t, out.ExpiresAt, "a key created without an explicit expiry must still expire")
	assert.WithinDuration(t, before.Add(defaultAPIKeyTTL), *out.ExpiresAt, time.Minute)
	assert.True(t, out.ExpiresAt.After(time.Now()), "default expiry must be in the future")

	// The key is usable now...
	_, _, err = tRepos.APIKeys.GetUserFromToken(ctx, hasher.HashAPIKey(out.Token))
	require.NoError(t, err, "key must be valid before its expiry")
}

// TestCreateAPIKey_HonorsExplicitExpiry ensures the default does not override a
// caller-supplied expiry in either direction.
func TestCreateAPIKey_HonorsExplicitExpiry(t *testing.T) {
	ctx := context.Background()
	usr := newTestUserWithPassword(t, "api-key-explicit-ttl")

	t.Run("LongerThanDefault", func(t *testing.T) {
		want := time.Now().Add(defaultAPIKeyTTL * 12).UTC()
		out, err := tSvc.User.CreateAPIKey(ctx, usr.ID, repo.APIKeyCreate{Name: "long", ExpiresAt: &want})
		require.NoError(t, err)
		require.NotNil(t, out.ExpiresAt)
		assert.WithinDuration(t, want, *out.ExpiresAt, time.Second)
	})

	t.Run("ShorterThanDefault", func(t *testing.T) {
		want := time.Now().Add(time.Hour).UTC()
		out, err := tSvc.User.CreateAPIKey(ctx, usr.ID, repo.APIKeyCreate{Name: "short", ExpiresAt: &want})
		require.NoError(t, err)
		require.NotNil(t, out.ExpiresAt)
		assert.WithinDuration(t, want, *out.ExpiresAt, time.Second)
	})
}

// TestCreateAPIKey_ExpiredKeyRejected confirms the stored expiry is actually
// enforced at authentication time, so the default TTL has teeth.
func TestCreateAPIKey_ExpiredKeyRejected(t *testing.T) {
	ctx := context.Background()
	usr := newTestUserWithPassword(t, "api-key-expired")

	past := time.Now().Add(-time.Hour).UTC()
	out, err := tSvc.User.CreateAPIKey(ctx, usr.ID, repo.APIKeyCreate{Name: "expired", ExpiresAt: &past})
	require.NoError(t, err)

	_, _, err = tRepos.APIKeys.GetUserFromToken(ctx, hasher.HashAPIKey(out.Token))
	require.Error(t, err, "an expired key must not authenticate")
}
