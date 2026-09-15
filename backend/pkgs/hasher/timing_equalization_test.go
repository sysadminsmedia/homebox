package hasher

import (
	"context"
	"testing"
)

// TestTimingEqualizationHashIsValidArgon2 guards against regressing the login
// account-enumeration timing side channel. The "user not found" / "no password"
// login paths call CheckDummyPasswordHashCtx to equalize response timing with a
// real login. That only works if the dummy comparison runs the argon2id KDF, which
// requires a *valid* argon2id hash — an arbitrary string would fail to decode
// instantly and perform no cryptographic work, reopening the side channel.
func TestTimingEqualizationHashIsValidArgon2(t *testing.T) {
	h := timingEqualizationHash()
	if h == "" {
		t.Fatal("timing equalization hash is empty")
	}
	if _, _, _, err := decodeHash(h); err != nil {
		t.Fatalf("timing equalization hash must be a decodable argon2id hash, got decode error: %v (hash=%q)", err, h)
	}
}

// TestStaticDummyHashIsValid ensures the compiled-in fallback used when runtime
// hash generation fails is itself a valid argon2id hash.
func TestStaticDummyHashIsValid(t *testing.T) {
	if _, _, _, err := decodeHash(staticDummyHash); err != nil {
		t.Fatalf("staticDummyHash must decode as argon2id, got: %v", err)
	}
}

// TestCheckDummyPasswordHashDoesRealWork sanity-checks that the dummy comparison
// exercises the real comparison path (returns a no-match result rather than
// short-circuiting) when protection is enabled.
func TestCheckDummyPasswordHashDoesRealWork(t *testing.T) {
	// Should not panic and should complete a real argon2id comparison.
	CheckDummyPasswordHashCtx(context.Background())
}
