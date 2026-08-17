package validate

import (
	"fmt"
	"net/http"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// maxNotifierRedirects caps redirect hops for outbound delivery, matching net/http's
// default so behavior is unchanged for legitimate redirect chains.
const maxNotifierRedirects = 10

// InstallNotifierRedirectGuard hardens http.DefaultClient so redirects are
// re-validated against the notifier SSRF policy on every hop.
//
// shoutrrr's generic service performs its HTTP request with http.DefaultClient and
// no CheckRedirect hook, so a host that passes the initial ValidateNotifierURL gate
// can respond with a 30x redirect to localhost / link-local / cloud-metadata / any
// other blocked destination, and the follow-up hop is delivered without re-checking
// the policy — bypassing the SSRF guards. Re-validating each hop closes that hole.
//
// This affects every http.DefaultClient consumer in the process, but in Homebox
// that is only shoutrrr: all other outbound HTTP clients (labelmaker, analytics,
// otel, product search, the GitHub release check) are constructed explicitly. The
// guard only rejects redirects whose target is blocked by policy; ordinary
// redirects to permitted hosts continue to be followed.
func InstallNotifierRedirectGuard(cfg *config.NotifierConf) {
	http.DefaultClient.CheckRedirect = NotifierRedirectGuard(cfg)
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
