package hasher

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
)

// APIKeyPrefix is prepended to every static API key so they can be identified
// at a glance (e.g. in logs or secret scanners).
const APIKeyPrefix = "hb_"

type Token struct {
	Raw  string
	Hash []byte
}

// apiKeyPepper holds the HMAC key applied to API key hashes. Stored as an
// atomic.Pointer so SetAPIKeyPepper at startup is visible to verify paths
// without a lock. A nil load means the app forgot to install it.
var apiKeyPepper atomic.Pointer[[]byte]

// SetAPIKeyPepper installs the server-side pepper used for HMAC-keyed API key
// hashing. Call once at startup before any API key is hashed or verified.
// Rotating the pepper invalidates every previously issued API key, so the
// caller must persist the same value across restarts.
func SetAPIKeyPepper(pepper []byte) {
	cp := make([]byte, len(pepper))
	copy(cp, pepper)
	apiKeyPepper.Store(&cp)
}

// GenerateToken generates a cryptographically random token. The non-context variant is
// kept for callers that don't have a context handy; both go through GenerateTokenCtx.
func GenerateToken() Token {
	return GenerateTokenCtx(context.Background())
}

func GenerateTokenCtx(ctx context.Context) Token {
	_, span := hasherTracer().Start(ctx, "hasher.GenerateToken")
	defer span.End()

	randomBytes := make([]byte, 16)
	_, _ = rand.Read(randomBytes)

	plainText := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := HashToken(plainText)

	span.SetAttributes(
		attribute.Int("token.raw.length", len(plainText)),
		attribute.Int("token.hash.length", len(hash)),
	)
	return Token{
		Raw:  plainText,
		Hash: hash,
	}
}

// HashToken hashes a token. This is fast (single SHA-256) so we don't trace it by
// default — adding a span on every middleware request would be more noise than signal.
func HashToken(plainTextToken string) []byte {
	hash := sha256.Sum256([]byte(plainTextToken))
	return hash[:]
}

// HashAPIKey returns HMAC-SHA256(pepper, plain). Indexed equality lookups stay
// O(1) while a DB-only leak — without the pepper held in app config — yields
// no usable hashes. Panics if SetAPIKeyPepper has not been called: missing it
// would silently fall back to a constant key, which is the failure mode this
// function exists to prevent.
func HashAPIKey(plain string) []byte {
	p := apiKeyPepper.Load()
	if p == nil || len(*p) == 0 {
		panic("hasher: API key pepper not configured (call SetAPIKeyPepper at startup)")
	}
	mac := hmac.New(sha256.New, *p)
	mac.Write([]byte(plain))
	return mac.Sum(nil)
}

// GenerateAPIKey produces a static API key with 256 bits of entropy and a
// recognizable prefix. The format is `hb_<base64url(32 bytes)>` — long enough
// to discourage guessing, mixed-case + symbols for visual distinctness from
// short session tokens, and identifiable by secret-scanning tools.
func GenerateAPIKey() Token {
	return GenerateAPIKeyCtx(context.Background())
}

func GenerateAPIKeyCtx(ctx context.Context) Token {
	_, span := hasherTracer().Start(ctx, "hasher.GenerateAPIKey")
	defer span.End()

	randomBytes := make([]byte, 32)
	_, _ = rand.Read(randomBytes)

	plainText := APIKeyPrefix + base64.RawURLEncoding.EncodeToString(randomBytes)
	hash := HashAPIKey(plainText)

	span.SetAttributes(
		attribute.Int("token.raw.length", len(plainText)),
		attribute.Int("token.hash.length", len(hash)),
	)
	return Token{
		Raw:  plainText,
		Hash: hash,
	}
}
