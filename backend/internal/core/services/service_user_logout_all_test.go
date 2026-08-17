package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

// TestSessionPersistenceGap documents the reported behavior: authenticating again
// does not invalidate prior tokens, and logging out only revokes the single token
// used for the logout request. A token the user no longer trusts therefore survives
// both re-login and logout. (Concurrent sessions are intentional; this asserts the
// pre-existing behavior that motivates a revoke-all control.)
func TestSessionPersistenceGap(t *testing.T) {
	ctx := context.Background()
	usr := newTestUserWithPassword(t, "persist-pw-1")

	tok1, err := tSvc.User.createSessionToken(ctx, usr.ID, false)
	require.NoError(t, err)
	tok2, err := tSvc.User.createSessionToken(ctx, usr.ID, false)
	require.NoError(t, err)

	// Logging out with token 2 leaves token 1 fully valid.
	require.NoError(t, tSvc.User.Logout(ctx, tok2.Raw))

	_, err = tRepos.AuthTokens.GetUserFromToken(ctx, hasher.HashToken(tok1.Raw))
	require.NoError(t, err, "prior token survives logout of another session")
}

// TestLogoutAll_RevokesEverySession verifies the new revoke-all control: it
// invalidates every session token for the user (all devices), so a leaked/stolen
// token can be unilaterally revoked. API keys live in a separate table and are
// intentionally unaffected.
func TestLogoutAll_RevokesEverySession(t *testing.T) {
	ctx := context.Background()
	usr := newTestUserWithPassword(t, "logout-all-pw")

	tok1, err := tSvc.User.createSessionToken(ctx, usr.ID, false)
	require.NoError(t, err)
	tok2, err := tSvc.User.createSessionToken(ctx, usr.ID, false)
	require.NoError(t, err)
	_, err = tSvc.User.createSessionToken(ctx, usr.ID, false)
	require.NoError(t, err)
	require.Equal(t, 3, countUserSessions(t, ctx, usr.ID))

	revoked, err := tSvc.User.LogoutAll(ctx, usr.ID)
	require.NoError(t, err)
	assert.Positive(t, revoked, "should report revoked token count")

	// No session tokens remain.
	assert.Equal(t, 0, countUserSessions(t, ctx, usr.ID))

	// None of the previously issued tokens authenticate any longer.
	_, err = tRepos.AuthTokens.GetUserFromToken(ctx, hasher.HashToken(tok1.Raw))
	require.Error(t, err, "token 1 must be revoked")
	_, err = tRepos.AuthTokens.GetUserFromToken(ctx, hasher.HashToken(tok2.Raw))
	require.Error(t, err, "token 2 must be revoked")
}
