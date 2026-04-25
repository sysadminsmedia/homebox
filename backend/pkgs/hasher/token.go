package hasher

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"

	"go.opentelemetry.io/otel/attribute"
)

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
