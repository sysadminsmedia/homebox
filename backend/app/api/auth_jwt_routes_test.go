package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/web/mid"
)

// TestBearerJWTCannotManageCredentials mounts the real routes and checks that
// an identity provider JWT reaches everyday data but is refused (403) on every
// route that creates or manages credentials, sessions, the account or
// collection membership. A short-lived token must never be able to mint a
// long-lived credential such as an API key.
func TestBearerJWTCannotManageCredentials(t *testing.T) {
	idp := newTestIdP(t)
	a, tokens := newAuthTestApp(t)
	a.jwtAuth = newJWTAuthenticator(config.AuthJWTConf{
		IssuerURL:   idp.issuer(),
		Audience:    "api-things",
		EmailClaim:  claimEmail,
		VerifyEmail: true,
	}, a.repos.Users)

	// mountRoutes needs the pieces run() normally sets up.
	full := new(a.conf)
	a.bus, a.authLimiter, a.notifierTestLimiter, a.registerLimiter = eventbus.New(), full.authLimiter, full.notifierTestLimiter, full.registerLimiter

	router := chi.NewMux()
	a.mountRoutes(router, errchain.New(mid.Errors(zerolog.Nop())), a.repos)

	do := func(method, path, token string) int {
		req := httptest.NewRequest(method, path, strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}
	jwt := idp.token(t, nil)

	denied := []struct{ method, path string }{
		{http.MethodPut, "/api/v1/users/self"},
		{http.MethodDelete, "/api/v1/users/self"},
		{http.MethodPost, "/api/v1/users/logout"},
		{http.MethodPost, "/api/v1/users/logout/all"},
		{http.MethodGet, "/api/v1/users/refresh"},
		{http.MethodPut, "/api/v1/users/self/change-password"},
		{http.MethodGet, "/api/v1/users/self/api-keys"},
		{http.MethodPost, "/api/v1/users/self/api-keys"},
		{http.MethodDelete, "/api/v1/users/self/api-keys/00000000-0000-0000-0000-000000000001"},
		{http.MethodPost, "/api/v1/groups"},
		{http.MethodPut, "/api/v1/groups"},
		{http.MethodDelete, "/api/v1/groups"},
		{http.MethodDelete, "/api/v1/groups/members/00000000-0000-0000-0000-000000000001"},
		{http.MethodGet, "/api/v1/groups/invitations"},
		{http.MethodPost, "/api/v1/groups/invitations"},
		{http.MethodDelete, "/api/v1/groups/invitations/00000000-0000-0000-0000-000000000001"},
		{http.MethodPost, "/api/v1/groups/invitations/00000000-0000-0000-0000-000000000001"},
	}
	for _, d := range denied {
		t.Run("JWTDenied "+d.method+" "+d.path, func(t *testing.T) {
			assert.Equal(t, http.StatusForbidden, do(d.method, d.path, jwt))
		})
	}

	t.Run("JWTAllowedOnEverydayData", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, do(http.MethodGet, "/api/v1/entities", jwt))
		assert.Equal(t, http.StatusOK, do(http.MethodGet, "/api/v1/users/self", jwt))
	})

	t.Run("SessionTokenStillManagesAPIKeys", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, do(http.MethodGet, "/api/v1/users/self/api-keys", tokens.Raw))
	})
}

// TestBearerJWTDiscoveryBounded checks that an unreachable identity provider
// cannot hang requests: discovery times out, is not retried during the
// cooldown, and concurrent requests are refused instead of queueing.
func TestBearerJWTDiscoveryBounded(t *testing.T) {
	var hits atomic.Int32
	release := make(chan struct{})
	hang := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(func() { close(release); hang.Close() })

	a, _ := newAuthTestApp(t)
	j := newJWTAuthenticator(config.AuthJWTConf{IssuerURL: hang.URL, Audience: "api-things", EmailClaim: claimEmail}, a.repos.Users)
	j.fetchTimeout = 300 * time.Millisecond
	j.cooldown = time.Hour

	// While the first discovery hangs, a second request is refused at once.
	done := make(chan error, 1)
	go func() { _, err := j.getVerifier(); done <- err }()
	require.Eventually(t, func() bool { return hits.Load() == 1 }, 2*time.Second, 10*time.Millisecond)
	start := time.Now()
	_, err := j.getVerifier()
	require.ErrorIs(t, err, errJWTDiscoveryBusy)
	assert.Less(t, time.Since(start), 100*time.Millisecond)

	// The hanging discovery gives up after fetchTimeout.
	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("discovery was not bounded by fetchTimeout")
	}

	// Within the cooldown the failure is returned without another fetch.
	_, err = j.getVerifier()
	require.Error(t, err)
	assert.Equal(t, int32(1), hits.Load())

	// After the cooldown a working issuer is discovered and cached.
	idp := newTestIdP(t)
	j.conf.IssuerURL = idp.issuer()
	j.cooldown = 0
	v1, err := j.getVerifier()
	require.NoError(t, err)
	v2, err := j.getVerifier()
	require.NoError(t, err)
	assert.Same(t, v1, v2)
}
