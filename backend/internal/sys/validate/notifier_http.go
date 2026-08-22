package validate

import (
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// maxNotifierRedirects caps redirect hops for outbound delivery, matching net/http's
// default so behavior is unchanged for legitimate redirect chains.
const maxNotifierRedirects = 10

// shoutrrr's own client has no timeout, so a hung receiver would pin the
// background maintenance job indefinitely.
const notifierHTTPTimeout = 30 * time.Second

// NotifierHTTPClient builds the client used for notifier delivery. The URL gate
// only sees the saved URL, so the policy is re-checked on every 30x hop and
// again on the address of every connect (which also covers DNS resolving
// differently than it did at validation time).
//
// The dial check is the backstop on purpose: this guard used to live on
// http.DefaultClient and worked only because shoutrrr happened to use that
// client. When v0.17.1 started building its own, it silently stopped running.
func NotifierHTTPClient(cfg *config.NotifierConf) *http.Client {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   notifierDialControl(cfg),
	}

	transport, _ := http.DefaultTransport.(*http.Transport)
	if transport != nil {
		transport = transport.Clone()
	} else {
		transport = &http.Transport{}
	}
	transport.DialContext = dialer.DialContext

	return &http.Client{
		Transport:     transport,
		CheckRedirect: NotifierRedirectGuard(cfg),
		Timeout:       notifierHTTPTimeout,
	}
}

// notifierDialControl rejects a connection whose address is blocked by policy.
// By this point host is an IP literal, which is what makes it immune to the DNS
// answer changing after validation.
func notifierDialControl(cfg *config.NotifierConf) func(network, address string, c syscall.RawConn) error {
	return func(_, address string, _ syscall.RawConn) error {
		// No policy: dial normally rather than failing closed.
		if cfg == nil {
			return nil
		}

		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return fmt.Errorf("notifier dial: cannot parse address %q: %w", address, err)
		}

		ip := net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("notifier dial: %q is not an IP address", host)
		}

		if err := ValidateResolvedIPs([]net.IP{ip}, cfg); err != nil {
			return fmt.Errorf("connection to %s blocked by notifier SSRF policy: %w", ip, err)
		}
		return nil
	}
}

// NotifierRedirectGuard is a CheckRedirect hook refusing any redirect whose
// target is blocked by cfg, plus a hop cap. A blocked hop aborts the request and
// surfaces as a delivery failure.
func NotifierRedirectGuard(cfg *config.NotifierConf) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxNotifierRedirects {
			return fmt.Errorf("stopped after %d redirects", maxNotifierRedirects)
		}

		// No policy: only the hop cap above applies.
		if cfg == nil {
			return nil
		}

		if err := validateHostAgainstPolicy(req.URL.Hostname(), cfg); err != nil {
			return fmt.Errorf("redirect to %s blocked by notifier SSRF policy: %w", req.URL.Redacted(), err)
		}
		return nil
	}
}
