package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/passwordresettokens"
)

// TestPasswordResetTokens_AtomicClaim_RejectsExpired exercises the
// defense-in-depth expiry predicate on the atomic UPDATE inside MarkUsed and
// ConsumeAndChangePassword. Service callers normally filter expired tokens
// via GetValidByHash before reaching the claim, but if the token expires in
// the gap between the lookup and the claim, the UPDATE must still refuse to
// burn it.
func TestPasswordResetTokens_AtomicClaim_RejectsExpired(t *testing.T) {
	ctx := context.Background()

	// Create a token that's already expired by the time we try to claim it.
	tok, err := tRepos.PasswordResetTokens.Create(ctx, tUser.ID, []byte("hash-1-claim-test"), time.Now().Add(-time.Minute))
	require.NoError(t, err)

	err = tRepos.PasswordResetTokens.MarkUsed(ctx, tok.ID, time.Now())
	require.ErrorIs(t, err, ErrPasswordResetTokenAlreadyClaimed,
		"MarkUsed must refuse to burn an expired token even when used_at is still NULL")

	// Sanity: the row still has used_at = NULL, since the conditional UPDATE
	// matched zero rows.
	row, err := tClient.PasswordResetTokens.Query().
		Where(passwordresettokens.ID(tok.ID)).
		Only(ctx)
	require.NoError(t, err)
	assert.Nil(t, row.UsedAt, "used_at must remain NULL when the claim is rejected for expiry")
}

func TestPasswordResetTokens_ConsumeAndChangePassword_RejectsExpired(t *testing.T) {
	ctx := context.Background()

	tok, err := tRepos.PasswordResetTokens.Create(ctx, tUser.ID, []byte("hash-2-consume-test"), time.Now().Add(-time.Minute))
	require.NoError(t, err)

	originalHash, err := tRepos.Users.GetOneID(ctx, tUser.ID)
	require.NoError(t, err)

	err = tRepos.PasswordResetTokens.ConsumeAndChangePassword(ctx, tok.ID, tUser.ID, "new-hash-should-not-stick")
	require.ErrorIs(t, err, ErrPasswordResetTokenAlreadyClaimed)

	// The user's password must NOT have changed — the whole tx should have
	// rolled back when the conditional UPDATE matched zero rows.
	after, err := tRepos.Users.GetOneID(ctx, tUser.ID)
	require.NoError(t, err)
	assert.Equal(t, originalHash.PasswordHash, after.PasswordHash,
		"password must be unchanged when the token claim is rejected")
}
