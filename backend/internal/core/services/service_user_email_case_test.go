package services

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegister_EmailCaseInsensitiveUniqueness proves that registering the same
// email in different cases is rejected as a duplicate, and that login still works
// for the single account (the case-variant account-takeover/DoS fix).
func TestRegister_EmailCaseInsensitiveUniqueness(t *testing.T) {
	ctx := context.Background()

	base := strings.ToLower(fk.Str(10))
	lower := base + "@example.com"
	upper := strings.ToUpper(base) + "@EXAMPLE.COM"
	const password = "correct-horse-battery-staple"

	_, err := tSvc.User.RegisterUser(ctx, UserRegistration{
		Name:     fk.Str(8),
		Email:    lower,
		Password: password,
	})
	require.NoError(t, err, "first registration should succeed")

	// Registering a case variant must be rejected as a duplicate.
	_, err = tSvc.User.RegisterUser(ctx, UserRegistration{
		Name:     fk.Str(8),
		Email:    upper,
		Password: password,
	})
	require.Error(t, err, "registering a case-variant of an existing email must be rejected")

	// Login must still work for both case forms — a single account exists.
	tok, err := tSvc.User.Login(ctx, lower, password, false)
	require.NoError(t, err, "login with the lowercase email should succeed")
	assert.NotEmpty(t, tok.Raw)

	tok, err = tSvc.User.Login(ctx, upper, password, false)
	require.NoError(t, err, "login with an uppercase variant should succeed (case-insensitive)")
	assert.NotEmpty(t, tok.Raw)
}
