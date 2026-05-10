package hasher

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"

	"go.opentelemetry.io/otel/attribute"
)

// APIKeyPrefix is prepended to every static API key so they can be identified
// at a glance (e.g. in logs or secret scanners).
const APIKeyPrefix = "hb_"

type Token struct {
	Raw  string
	Hash []byte
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
