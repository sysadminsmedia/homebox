package validate_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/nicholas-fedor/shoutrrr"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/notifier"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// The servers sit on different loopback addresses on purpose. Both have to be
// reachable for the first hop, so "block all localhost" would stop the request
// before it reached the redirect under test. Blocking only 127.0.0.2/32 leaves
// the redirector permitted and its target not.
const (
	redirectorAddr = "127.0.0.1:0"
	victimIP       = "127.0.0.2"
)

// blockVictimPolicy blocks the victim only, leaving the redirector reachable.
func blockVictimPolicy() *config.NotifierConf {
	return &config.NotifierConf{BlockNets: []string{victimIP + "/32"}}
}

// serveOn starts an httptest server on a specific address.
func serveOn(t *testing.T, addr string, h http.Handler) *httptest.Server {
	t.Helper()
	l, err := net.Listen("tcp", addr)
	require.NoError(t, err, "binding %s", addr)

	srv := httptest.NewUnstartedServer(h)
	_ = srv.Listener.Close()
	srv.Listener = l
	srv.Start()
	t.Cleanup(srv.Close)
	return srv
}

// startVictimAndRedirector puts a victim on the blocked address that records
// hits, and a redirector on a permitted one that 307s to it. Returns the
// notifier URL for the redirector and the victim's hit count.
func startVictimAndRedirector(t *testing.T) (notifierURL string, victimHits func() int32) {
	t.Helper()

	var hits int32
	victim := serveOn(t, victimIP+":0", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))

	redirector := serveOn(t, redirectorAddr, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", victim.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))

	// shoutrrr's generic service wraps a plain http(s) URL as generic+http(s)://...
	return "generic+" + redirector.URL, func() int32 { return atomic.LoadInt32(&hits) }
}

// sendVia is what notifier.Sender does for a generic webhook, minus the URL
// gate — which would otherwise reject these loopback servers before the hop
// under test.
func sendVia(client *http.Client, rawURL, message string) error {
	sender, err := shoutrrr.CreateSenderWithOptions(types.SenderOptions{HTTPClient: client}, rawURL)
	if err != nil {
		return err
	}
	for _, sendErr := range sender.Send(message, nil) {
		if sendErr != nil {
			return sendErr
		}
	}
	return nil
}

// Documents the hole: shoutrrr follows redirects with no policy re-check, so a
// host that passes the URL gate can 307 to a blocked destination. Sends through
// the package-level shoutrrr.Send, which is what we must not do.
func TestNotifierRedirectSSRF_Unguarded(t *testing.T) {
	notifierURL, victimHits := startVictimAndRedirector(t)

	err := shoutrrr.Send(notifierURL, "Test message from Homebox")
	require.NoError(t, err)
	assert.Equal(t, int32(1), victimHits(), "unguarded: redirect to the blocked victim IS followed (SSRF)")
}

// The hardened client must refuse the redirect and never reach the victim. It
// has to be injected into shoutrrr, not set on http.DefaultClient — since
// v0.17.1 the generic service builds its own client per send, so anything on
// DefaultClient is never consulted.
func TestNotifierRedirectSSRF_Guarded(t *testing.T) {
	notifierURL, victimHits := startVictimAndRedirector(t)

	client := validate.NotifierHTTPClient(blockVictimPolicy())
	err := sendVia(client, notifierURL, "Test message from Homebox")

	require.Error(t, err, "guarded: redirect to a blocked destination must be refused and the send must fail")
	assert.Equal(t, int32(0), victimHits(), "guarded: the blocked redirect target must never be reached")
}

// The second layer. The dial check runs on the address of every connect, so it
// holds however the request got there — including a name that resolved to
// something permitted at validation time and blocked at connect (rebinding),
// which CheckRedirect can't see.
func TestNotifierDialGuard_BlocksDirectConnection(t *testing.T) {
	var hits int32
	victim := serveOn(t, victimIP+":0", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))

	client := validate.NotifierHTTPClient(blockVictimPolicy())
	err := sendVia(client, "generic+"+victim.URL, "Test message from Homebox")

	require.Error(t, err, "a direct send to a blocked address must fail at dial time")
	assert.Equal(t, int32(0), atomic.LoadInt32(&hits), "the blocked address must never be connected to")
}

// The counterweight: a guard that blocks everything would pass the tests above.
func TestNotifierDialGuard_AllowsPermittedDestination(t *testing.T) {
	var hits int32
	receiver := serveOn(t, redirectorAddr, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))

	client := validate.NotifierHTTPClient(blockVictimPolicy())
	err := sendVia(client, "generic+"+receiver.URL, "Test message from Homebox")

	require.NoError(t, err, "a permitted destination must still be delivered to")
	assert.Equal(t, int32(1), atomic.LoadInt32(&hits))
}

// Redirects to permitted hosts still work, so webhook endpoints that redirect
// keep working.
func TestNotifierRedirectGuard_FollowsPermittedRedirect(t *testing.T) {
	var hits int32
	final := serveOn(t, redirectorAddr, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))

	redirector := serveOn(t, redirectorAddr, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", final.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))

	client := validate.NotifierHTTPClient(blockVictimPolicy())
	err := sendVia(client, "generic+"+redirector.URL, "Test message from Homebox")

	require.NoError(t, err, "a redirect to a permitted host must still be followed")
	assert.Equal(t, int32(1), atomic.LoadInt32(&hits))
}

// The first gate is still wired into the shared path: a URL pointing straight at
// a blocked destination is refused before any request goes out.
func TestNotifierSend_RejectsBlockedURLBeforeSending(t *testing.T) {
	var hits int32
	victim := serveOn(t, victimIP+":0", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))

	err := notifier.Send(blockVictimPolicy(), "generic+"+victim.URL, "Test message from Homebox")

	require.Error(t, err)
	var vErr *notifier.ValidationError
	require.ErrorAs(t, err, &vErr, "should be reported as a validation failure, not a delivery failure")
	assert.Equal(t, int32(0), atomic.LoadInt32(&hits))
}

// Pins the policy's scope. notifier.* covers generic webhooks; everything else
// has a fixed or operator-chosen endpoint. Applying the network rules there
// would break a LAN Gotify on a bogon address and prevent nothing.
func TestNotifierSend_NonGenericBypassesNetworkPolicy(t *testing.T) {
	cfg := blockVictimPolicy()

	// Same address, non-generic scheme.
	require.False(t, validate.IsGenericNotifier("gotify://"+victimIP+":1234/token"),
		"gotify:// must not be treated as a generic webhook")
	require.NoError(t, validate.ValidateNotifierURL("gotify://"+victimIP+":1234/token", cfg),
		"the URL gate must not apply network policy to non-generic services")

	// The generic form of the same address is still refused, so the difference is
	// the scheme and not a hole.
	require.True(t, validate.IsGenericNotifier("generic+http://"+victimIP+":1234"))
	require.Error(t, validate.ValidateNotifierURL("generic+http://"+victimIP+":1234", cfg))
}
