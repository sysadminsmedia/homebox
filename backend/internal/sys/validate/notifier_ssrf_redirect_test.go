package validate_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// startVictim spins up a loopback server that records whether it was reached.
func startVictim(t *testing.T) (serverURL string, hits func() int32) {
	t.Helper()

	var count int32
	victim := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(victim.Close)

	return victim.URL, func() int32 { return atomic.LoadInt32(&count) }
}

// TestNotifierRedirectGuard_RefusesBlockedHop verifies the CheckRedirect hook refuses
// a hop whose target is blocked by policy. Without it, a host that passes the initial
// URL gate could 30x the request onward to localhost / link-local / cloud metadata.
func TestNotifierRedirectGuard_RefusesBlockedHop(t *testing.T) {
	guard := validate.NotifierRedirectGuard(&config.NotifierConf{BlockLocalhost: true})

	blocked, err := url.Parse("http://127.0.0.1:8080/webhook")
	require.NoError(t, err)

	err = guard(&http.Request{URL: blocked}, nil)
	assert.Error(t, err, "a redirect to a blocked destination must be refused")
}

// TestNotifierRedirectGuard_AllowsPermittedHop ensures ordinary redirects to
// permitted hosts continue to be followed.
func TestNotifierRedirectGuard_AllowsPermittedHop(t *testing.T) {
	guard := validate.NotifierRedirectGuard(&config.NotifierConf{})

	allowed, err := url.Parse("http://127.0.0.1:8080/webhook")
	require.NoError(t, err)

	err = guard(&http.Request{URL: allowed}, nil)
	assert.NoError(t, err, "with no blocks configured the redirect must be followed")
}

// TestNotifierRedirectGuard_CapsHops verifies the hop cap still applies.
func TestNotifierRedirectGuard_CapsHops(t *testing.T) {
	guard := validate.NotifierRedirectGuard(&config.NotifierConf{})

	target, err := url.Parse("http://example.com/webhook")
	require.NoError(t, err)

	err = guard(&http.Request{URL: target}, make([]*http.Request, 10))
	assert.Error(t, err, "the redirect chain must be capped")
}

// TestNotifierHTTPClient_BlocksDestination verifies the client's dialer enforces the
// policy on the very first hop. The previous guard lived only on CheckRedirect, which
// net/http invokes before *following* a redirect and never for the initial request,
// so the first hop was never checked at connection time.
func TestNotifierHTTPClient_BlocksDestination(t *testing.T) {
	victimURL, victimHits := startVictim(t)

	client := validate.NotifierHTTPClient(&config.NotifierConf{BlockLocalhost: true})

	resp, err := client.Get(victimURL) //nolint:noctx // exercising the dialer, not a real request path
	if err == nil {
		_ = resp.Body.Close()
	}

	require.Error(t, err, "a blocked destination must be refused at dial time")
	assert.Equal(t, int32(0), victimHits(), "the blocked destination must never be reached")
}

// TestNotifierHTTPClient_AllowsPermittedDestination ensures the guarded client still
// delivers to destinations the policy permits.
func TestNotifierHTTPClient_AllowsPermittedDestination(t *testing.T) {
	victimURL, victimHits := startVictim(t)

	client := validate.NotifierHTTPClient(&config.NotifierConf{})

	resp, err := client.Get(victimURL) //nolint:noctx // exercising the dialer, not a real request path
	require.NoError(t, err)
	_ = resp.Body.Close()

	assert.Equal(t, int32(1), victimHits(), "a permitted destination must still be reached")
}

// TestSendNotifierMessage_EnforcesPolicy verifies the send path validates the URL
// before delivering. shoutrrr.Send would otherwise deliver through an unguarded
// client of its own.
func TestSendNotifierMessage_EnforcesPolicy(t *testing.T) {
	victimURL, victimHits := startVictim(t)

	err := validate.SendNotifierMessage(
		"generic+"+victimURL,
		"Test message from Homebox",
		&config.NotifierConf{BlockLocalhost: true},
	)

	require.Error(t, err, "a blocked destination must not be delivered to")
	assert.Equal(t, int32(0), victimHits(), "the blocked destination must never be reached")
}

// TestSendNotifierMessage_DeliversPermitted ensures the guarded send path still
// delivers notifications the policy permits.
func TestSendNotifierMessage_DeliversPermitted(t *testing.T) {
	victimURL, victimHits := startVictim(t)

	err := validate.SendNotifierMessage(
		"generic+"+victimURL,
		"Test message from Homebox",
		&config.NotifierConf{},
	)

	require.NoError(t, err)
	assert.Equal(t, int32(1), victimHits(), "a permitted notifier must still be delivered")
}
