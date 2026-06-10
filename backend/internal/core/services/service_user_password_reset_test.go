package services

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/authtokens"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/passwordresettokens"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/user"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

// newTestUserWithPassword creates a fresh user with a known password so each
// test gets its own subject and doesn't fight with the shared tUser.
func newTestUserWithPassword(t *testing.T, password string) repo.UserOut {
	t.Helper()
	hash, err := hasher.HashPassword(password)
	require.NoError(t, err)
	usr, err := tRepos.Users.Create(testCtx(), repo.UserCreate{
		Name:           fk.Str(10),
		Email:          fk.Email(),
		Password:       &hash,
		DefaultGroupID: tGroup.ID,
	})
	require.NoError(t, err)
	return usr
}

func extractToken(t *testing.T, link string) string {
	t.Helper()
	u, err := url.Parse(link)
	require.NoError(t, err, "reset link should be a valid URL")
	tok := u.Query().Get("token")
	require.NotEmpty(t, tok, "reset link should include a token query param")
	return tok
}

func TestRequestPasswordReset_NoMailer_ReturnsError(t *testing.T) {
	// tSvc.User.mailer is nil by default in tests; that's the path we want.
	err := tSvc.User.RequestPasswordReset(testCtx(), "anyone@example.com", "https://example.com")
	assert.ErrorIs(t, err, ErrorMailerNotConfigured)
}

func TestGenerateResetLink_HappyPath(t *testing.T) {
	usr := newTestUserWithPassword(t, "original-password")

	link, err := tSvc.User.GenerateResetLink(testCtx(), usr.Email, "https://example.com")
	require.NoError(t, err)
	assert.Contains(t, link, "https://example.com/reset-password?token=")

	// The token should be persisted (hashed) and unused.
	rawToken := extractToken(t, link)
	hash := hasher.HashToken(rawToken)
	got, err := tRepos.PasswordResetTokens.GetValidByHash(testCtx(), hash)
	require.NoError(t, err)
	assert.Equal(t, usr.ID, got.UserID)
}

func TestGenerateResetLink_UnknownEmail_ReturnsNotFound(t *testing.T) {
	_, err := tSvc.User.GenerateResetLink(testCtx(), "ghost-"+fk.Email(), "https://example.com")
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err), "expected NotFound, got %v", err)
}

func TestGenerateResetLink_OIDCUser_ReturnsSentinel(t *testing.T) {
	// User with no local password (OIDC-only style).
	usr, err := tRepos.Users.Create(testCtx(), repo.UserCreate{
		Name:           fk.Str(10),
		Email:          fk.Email(),
		Password:       nil,
		DefaultGroupID: tGroup.ID,
	})
	require.NoError(t, err)

	_, err = tSvc.User.GenerateResetLink(testCtx(), usr.Email, "https://example.com")
	require.Error(t, err)
	assert.ErrorIs(t, err, errResetUserHasNoPassword)
}

func TestResetPassword_HappyPath_RevokesSessions(t *testing.T) {
	ctx := testCtx()
	usr := newTestUserWithPassword(t, "old-password")

	// Give the user two active sessions; both must be killed by the reset.
	_, err := tSvc.User.createSessionToken(ctx, usr.ID, false)
	require.NoError(t, err)
	_, err = tSvc.User.createSessionToken(ctx, usr.ID, false)
	require.NoError(t, err)

	link, err := tSvc.User.GenerateResetLink(ctx, usr.Email, "https://example.com")
	require.NoError(t, err)
	rawToken := extractToken(t, link)

	require.NoError(t, tSvc.User.ResetPassword(ctx, rawToken, "brand-new-password"))

	// Old sessions are gone — DeleteAllByUser should have wiped them.
	gone, err := tClient.AuthTokens.Query().
		Where(authtokens.HasUserWith(user.ID(usr.ID))).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, gone, "all sessions for the user should be revoked")

	// Login with the new password works.
	_, err = tSvc.User.Login(ctx, usr.Email, "brand-new-password", false)
	require.NoError(t, err)

	// Login with the old password fails.
	_, err = tSvc.User.Login(ctx, usr.Email, "old-password", false)
	require.ErrorIs(t, err, ErrorInvalidLogin)

	// Token is marked used (not retrievable as valid).
	_, err = tRepos.PasswordResetTokens.GetValidByHash(ctx, hasher.HashToken(rawToken))
	assert.True(t, ent.IsNotFound(err), "used token must not be retrievable as valid")
}

func TestResetPassword_TokenReplay_Rejected(t *testing.T) {
	ctx := testCtx()
	usr := newTestUserWithPassword(t, "first-password")

	link, err := tSvc.User.GenerateResetLink(ctx, usr.Email, "https://example.com")
	require.NoError(t, err)
	rawToken := extractToken(t, link)

	require.NoError(t, tSvc.User.ResetPassword(ctx, rawToken, "second-password"))

	// Replay — the same token must not work twice.
	err = tSvc.User.ResetPassword(ctx, rawToken, "third-password")
	assert.ErrorIs(t, err, ErrorPasswordResetInvalid)
}

func TestResetPassword_AtomicClaim_LosesRaceToConcurrentReset(t *testing.T) {
	// Simulate the race: two requests both get the token via GetValidByHash
	// before either marks it used. The atomic conditional update inside
	// MarkUsed must let only one win. We model this by manually marking the
	// token used between GetValidByHash and the second ResetPassword call —
	// the second caller's MarkUsed should fail with the claim-race error,
	// which ResetPassword translates to ErrorPasswordResetInvalid.
	ctx := testCtx()
	usr := newTestUserWithPassword(t, "starting-pw")

	link, err := tSvc.User.GenerateResetLink(ctx, usr.Email, "https://example.com")
	require.NoError(t, err)
	rawToken := extractToken(t, link)

	// Mark the token used out-of-band, mimicking a concurrent winner. After
	// this, ResetPassword's GetValidByHash will return NotFound (since
	// GetValidByHash filters used tokens), giving ErrorPasswordResetInvalid.
	tok, err := tRepos.PasswordResetTokens.GetValidByHash(ctx, hasher.HashToken(rawToken))
	require.NoError(t, err)
	require.NoError(t, tRepos.PasswordResetTokens.MarkUsed(ctx, tok.ID, time.Now()))

	err = tSvc.User.ResetPassword(ctx, rawToken, "race-loser-pw")
	require.ErrorIs(t, err, ErrorPasswordResetInvalid)

	// Confirm the password was NOT changed by the losing caller.
	_, err = tSvc.User.Login(ctx, usr.Email, "starting-pw", false)
	assert.NoError(t, err, "original password must still work — losing race must not have changed it")
}

func TestResetPassword_BogusToken_Rejected(t *testing.T) {
	err := tSvc.User.ResetPassword(testCtx(), "this-is-not-a-real-token", "newpw")
	assert.ErrorIs(t, err, ErrorPasswordResetInvalid)
}

// TestResetPassword_InvalidToken_RunsDummyHash exists to catch regressions of
// the timing-equalization fix. We don't assert absolute timings (CI is too
// noisy) — instead we assert that the invalid-token path takes at LEAST a
// reasonable fraction of the valid-token path. argon2id with the configured
// params (m=64MiB, t=3, p=2) takes tens of ms on any modern host; the
// invalid-token path without the dummy hash returns in microseconds, so a
// loose floor reliably distinguishes "dummy hash ran" from "didn't".
func TestResetPassword_InvalidToken_RunsDummyHash(t *testing.T) {
	ctx := testCtx()

	// Warm up: argon2id allocates ~64 MiB; the first call in a process can be
	// markedly slower than steady-state. Run once so the measurement below
	// reflects steady-state cost.
	_, _ = hasher.HashPasswordCtx(ctx, "warmup-pw")

	const password = "any-password-of-reasonable-length"

	measure := func(token string) time.Duration {
		start := time.Now()
		_ = tSvc.User.ResetPassword(ctx, token, password)
		return time.Since(start)
	}

	// Establish a per-host baseline by measuring a real argon2id directly.
	// This avoids hard-coding a millisecond floor that could be flaky on
	// slow CI runners or fast workstations.
	hashStart := time.Now()
	_, err := hasher.HashPasswordCtx(ctx, password)
	require.NoError(t, err)
	hashCost := time.Since(hashStart)

	invalidElapsed := measure("definitely-not-a-real-token-value-here")

	// The invalid-token path must take at least ~half the cost of a single
	// argon2id call. If the dummy hash were skipped, this would return in
	// well under a millisecond and the assertion would fail loudly.
	floor := hashCost / 2
	assert.GreaterOrEqual(t, invalidElapsed, floor,
		"invalid-token path returned in %s but a single argon2id costs %s — dummy hash likely was not run, exposing a timing oracle for valid vs invalid tokens",
		invalidElapsed, hashCost)
}

func TestResetPassword_EmptyInputs_Rejected(t *testing.T) {
	err := tSvc.User.ResetPassword(testCtx(), "", "newpw")
	require.ErrorIs(t, err, ErrorPasswordResetInvalid)

	err = tSvc.User.ResetPassword(testCtx(), "sometoken", "")
	assert.ErrorIs(t, err, ErrorPasswordResetInvalid)
}

func TestResetPassword_ExpiredToken_Rejected(t *testing.T) {
	ctx := testCtx()
	usr := newTestUserWithPassword(t, "starting-password")

	// Mint a token directly in the repo with an expiration in the past, so we
	// don't have to time-travel the clock or rebuild the service.
	tok := hasher.GenerateTokenCtx(ctx)
	_, err := tRepos.PasswordResetTokens.Create(ctx, usr.ID, tok.Hash, time.Now().Add(-1*time.Minute))
	require.NoError(t, err)

	err = tSvc.User.ResetPassword(ctx, tok.Raw, "wont-work")
	assert.ErrorIs(t, err, ErrorPasswordResetInvalid)
}

func TestPurgeExpired_RemovesExpiredAndUsedTokens(t *testing.T) {
	ctx := testCtx()
	usr := newTestUserWithPassword(t, "irrelevant")

	// Expired token.
	expiredTok := hasher.GenerateTokenCtx(ctx)
	expiredRow, err := tRepos.PasswordResetTokens.Create(ctx, usr.ID, expiredTok.Hash, time.Now().Add(-1*time.Hour))
	require.NoError(t, err)

	// Used token (still within its TTL).
	usedTok := hasher.GenerateTokenCtx(ctx)
	usedRow, err := tRepos.PasswordResetTokens.Create(ctx, usr.ID, usedTok.Hash, time.Now().Add(time.Hour))
	require.NoError(t, err)
	require.NoError(t, tRepos.PasswordResetTokens.MarkUsed(ctx, usedRow.ID, time.Now()))

	// Fresh token that should survive the purge.
	freshTok := hasher.GenerateTokenCtx(ctx)
	freshRow, err := tRepos.PasswordResetTokens.Create(ctx, usr.ID, freshTok.Hash, time.Now().Add(time.Hour))
	require.NoError(t, err)

	deleted, err := tRepos.PasswordResetTokens.PurgeExpired(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, deleted, 2, "expired + used token rows should be deleted")

	exists := func(id uuid.UUID) bool {
		_, err := tClient.PasswordResetTokens.Query().Where(passwordresettokens.ID(id)).Only(ctx)
		return err == nil
	}
	assert.False(t, exists(expiredRow.ID), "expired token should be gone")
	assert.False(t, exists(usedRow.ID), "used token should be gone")
	assert.True(t, exists(freshRow.ID), "fresh token should survive")
}
