package hasher

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ITERATIONS = 200

func init() {
	// Tests need a pepper installed for HashAPIKey/GenerateAPIKey paths.
	SetAPIKeyPepper([]byte("test-pepper-not-for-production-use-only"))
}

func Test_NewToken(t *testing.T) {
	t.Parallel()
	tokens := make([]Token, ITERATIONS)
	for i := 0; i < ITERATIONS; i++ {
		tokens[i] = GenerateToken()
	}

	// Check if they are unique
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			if tokens[i].Raw == tokens[j].Raw {
				t.Errorf("NewToken() failed to generate unique tokens")
			}
		}
	}
}

func Test_HashToken_CheckTokenHash(t *testing.T) {
	t.Parallel()
	for i := 0; i < ITERATIONS; i++ {
		token := GenerateToken()

		// Check raw text is reltively random
		for j := 0; j < 5; j++ {
			assert.NotEqual(t, token.Raw, GenerateToken().Raw)
		}

		// Check token length is less than 32 characters
		assert.Less(t, len(token.Raw), 32)

		// Check hash is the same
		assert.Equal(t, token.Hash, HashToken(token.Raw))
	}
}

func Test_GenerateAPIKey_Format(t *testing.T) {
	t.Parallel()
	for i := 0; i < ITERATIONS; i++ {
		k := GenerateAPIKey()
		assert.True(t, strings.HasPrefix(k.Raw, APIKeyPrefix), "raw key must carry the hb_ prefix")
		assert.Len(t, k.Hash, 32, "HMAC-SHA256 output is 32 bytes")
		assert.Equal(t, k.Hash, HashAPIKey(k.Raw), "stored hash must match HashAPIKey(raw)")
	}
}

func Test_HashAPIKey_DiffersFromSHA256(t *testing.T) {
	t.Parallel()
	// The whole point of HMAC-keyed hashing is that an attacker with the DB
	// but not the pepper can't precompute or verify hashes via plain SHA-256.
	k := GenerateAPIKey()
	assert.NotEqual(t, HashToken(k.Raw), HashAPIKey(k.Raw))
}

func Test_HashAPIKey_IsKeyed(t *testing.T) {
	// Mutates the package-level pepper, so cannot run in parallel with the
	// other tests that also depend on it being stable.
	original := apiKeyPepper.Load()
	t.Cleanup(func() {
		require.NotNil(t, original)
		apiKeyPepper.Store(original)
	})

	SetAPIKeyPepper([]byte("pepper-A-pepper-A-pepper-A-pepperA"))
	a := HashAPIKey("hb_sample-token-value")
	SetAPIKeyPepper([]byte("pepper-B-pepper-B-pepper-B-pepperB"))
	b := HashAPIKey("hb_sample-token-value")
	assert.NotEqual(t, a, b)
}

func Test_HashAPIKey_PanicsWithoutPepper(t *testing.T) {
	// This test MUST NOT run in parallel — it temporarily clears global state.
	original := apiKeyPepper.Load()
	t.Cleanup(func() {
		require.NotNil(t, original)
		apiKeyPepper.Store(original)
	})
	apiKeyPepper.Store(nil)
	assert.Panics(t, func() { HashAPIKey("anything") })
}
