package validate

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/nicholas-fedor/shoutrrr/pkg/router"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// maxNotifierRedirects caps redirect hops for outbound delivery, matching net/http's
// default so behavior is unchanged for legitimate redirect chains.
const maxNotifierRedirects = 10

// notifierHTTPTimeout bounds a single notifier delivery.
const notifierHTTPTimeout = 30 * time.Second

// InstallNotifierRedirectGuard hardens http.DefaultClient so redirects made through
// it are re-validated against the notifier SSRF policy on every hop.
//
// This is only a backstop. shoutrrr does not use http.DefaultClient: its services
// build their own clients, and shoutrrr.Send routes through a package-level router
// that never injects one, so a guard installed here never ran on the delivery path.
// SendNotifierMessage is the enforced path — it supplies NotifierHTTPClient, which
// carries the same redirect guard plus a policy-checked dialer. This call remains so
// that any other code reaching for http.DefaultClient is covered too.
func InstallNotifierRedirectGuard(cfg *config.NotifierConf) {
	http.DefaultClient.CheckRedirect = NotifierRedirectGuard(cfg)
}

// SendNotifierMessage validates rawURL against the SSRF policy and, if it passes,
// delivers message through NotifierHTTPClient.
//
// Validation and delivery are deliberately bundled: shoutrrr.Send would otherwise
// send through an unguarded client, so a caller that forgot the validation step —
// or a scheme the initial gate did not cover — would reach the network unchecked.
func SendNotifierMessage(rawURL, message string, cfg *config.NotifierConf) error {
	if err := ValidateNotifierURL(rawURL, cfg); err != nil {
		return err
	}

	sender, err := router.NewWithOptions(nil, types.SenderOptions{
		HTTPClient: NotifierHTTPClient(cfg),
		Timeout:    notifierHTTPTimeout,
	})
	if err != nil {
		return fmt.Errorf("creating notifier sender: %w", err)
	}

	return sender.Route(rawURL, message)
}

// NotifierHTTPClient returns the HTTP client used to deliver notifications. It
// enforces the SSRF policy at two points the initial URL check cannot cover:
//
//   - CheckRedirect re-validates the target of every redirect hop, so a host that
//     passes the URL gate cannot 30x the request onward to a blocked destination.
//   - DialContext resolves each hostname itself, validates the addresses it got
//     back, and then dials one of those exact addresses. Re-using the validated
//     address instead of letting the transport resolve the name again closes the
//     DNS-rebinding window between validation and connection.
func NotifierHTTPClient(cfg *config.NotifierConf) *http.Client {
	client := &http.Client{
		Timeout:       notifierHTTPTimeout,
		CheckRedirect: NotifierRedirectGuard(cfg),
	}

	// With no policy configured there is nothing to enforce at dial time; keep the
	// default transport rather than failing every connection.
	if cfg == nil {
		return client
	}

	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return client
	}

	guarded := transport.Clone()
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}

	guarded.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid notifier address %q: %w", addr, err)
		}

		ips, err := resolveHostForPolicy(ctx, host, cfg)
		if err != nil {
			return nil, fmt.Errorf("notifier destination %q blocked by SSRF policy: %w", host, err)
		}

		var lastErr error
		for _, ip := range ips {
			conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if dialErr == nil {
				return conn, nil
			}
			lastErr = dialErr
		}

		if lastErr == nil {
			lastErr = fmt.Errorf("no addresses available for %q", host)
		}

		return nil, lastErr
	}

	client.Transport = guarded

	return client
}

// NotifierRedirectGuard returns an http.Client CheckRedirect hook that refuses any
// redirect whose target resolves to an address blocked by cfg, and caps the number
// of hops. Returning a non-nil error aborts the request with that error rather than
// following the redirect, so a blocked hop surfaces as a delivery failure.
func NotifierRedirectGuard(cfg *config.NotifierConf) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxNotifierRedirects {
			return fmt.Errorf("stopped after %d redirects", maxNotifierRedirects)
		}

		// Defensive: with no policy configured, preserve default behavior (only the
		// hop cap above applies) rather than blocking every redirect.
		if cfg == nil {
			return nil
		}

		if err := validateHostAgainstPolicy(req.URL.Hostname(), cfg); err != nil {
			return fmt.Errorf("redirect to %s blocked by notifier SSRF policy: %w", req.URL.Redacted(), err)
		}
		return nil
	}
}
