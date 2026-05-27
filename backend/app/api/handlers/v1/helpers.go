package v1

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// GetHBURL determines the base URL of the Homebox instance using the following priority:
// 1. Configured hostname from Options.Hostname
// 2. X-Forwarded headers (if TrustProxy is enabled)
// 3. Referer header
// 4. Fallback URL (ctrl.url)
func GetHBURL(r *http.Request, options *config.Options, fallback string) string {
	// 1. Use configured hostname if set
	if options.Hostname != "" {
		return ensureScheme(options.Hostname, r, options.TrustProxy)
	}

	// 2. Use X-Forwarded headers if TrustProxy is enabled
	if options.TrustProxy {
		if xfHost := r.Header.Get("X-Forwarded-Host"); xfHost != "" {
			scheme := getScheme(r, options.TrustProxy)
			return scheme + "://" + xfHost
		}
	}

	// 3. Fall back to Referer header
	if referer := r.Header.Get("Referer"); referer != "" {
		return stripPathFromURL(referer)
	}

	// 4. Fall back to the controller's URL
	if fallback != "" {
		return stripPathFromURL(fallback)
	}

	return ""
}

// ensureScheme ensures the hostname has a proper URL scheme
func ensureScheme(hostname string, r *http.Request, trustProxy bool) string {
	// If hostname already has a scheme, use it as-is
	if strings.HasPrefix(hostname, "http://") || strings.HasPrefix(hostname, "https://") {
		return strings.TrimSuffix(hostname, "/")
	}

	// Otherwise, determine scheme from request
	scheme := getScheme(r, trustProxy)
	return scheme + "://" + hostname
}

// getScheme determines the appropriate URL scheme based on request and proxy settings
func getScheme(r *http.Request, trustProxy bool) string {
	if r.TLS != nil {
		return schemeHTTPS
	}
	if trustProxy {
		// X-Forwarded-Proto may be a comma-separated list (one entry per hop);
		// the leftmost entry is the original client-facing protocol — that's
		// what we want for user-facing URL construction. A literal equality
		// check fails on legitimate "https, http" multi-hop values.
		proto := strings.ToLower(firstHeaderValue(r.Header.Get("X-Forwarded-Proto")))
		if proto == schemeHTTPS {
			return schemeHTTPS
		}
	}
	return "http"
}

// firstHeaderValue returns the first comma-separated value from a header
// field, trimmed. RFC 7230 §3.2.2 allows multiple instances of a field to
// be combined with commas; net/http's Header.Get returns only the first
// occurrence but does NOT split combined values, so a header like
// `X-Forwarded-Host: a, b` would otherwise be embedded verbatim into URLs.
func firstHeaderValue(v string) string {
	if i := strings.IndexByte(v, ','); i >= 0 {
		v = v[:i]
	}
	return strings.TrimSpace(v)
}

// validProxyHost reports whether s is a syntactically valid host:port (or
// host) value, with no scheme, path, query, fragment, whitespace, or control
// characters. Callers building URLs from untrusted X-Forwarded-Host must run
// this check — without it, a misconfigured proxy that forwards client-set
// headers could let an attacker inject path/CRLF/full-URL payloads into a
// password-reset email link.
func validProxyHost(s string) bool {
	if s == "" {
		return false
	}
	// Reject control characters, whitespace, and structural URL characters
	// outright. Hosts legitimately include letters, digits, dot, dash, colon
	// (port separator), and brackets (IPv6).
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	if strings.ContainsAny(s, " \t/?#\\") {
		return false
	}
	if strings.Contains(s, "://") {
		return false
	}
	// Final structural check: url.Parse should round-trip the string as the
	// authority component. If it doesn't, something funny is going on.
	u, err := url.Parse("http://" + s)
	if err != nil || u.Host != s || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	return true
}

// SecureBaseURL returns a base URL safe to embed in security-sensitive emails
// (password reset, etc.). Unlike GetHBURL it deliberately omits the Referer
// fallback, since Referer is unauthenticated client input — an attacker who
// can reach /forgot-password could otherwise poison the link in the victim's
// reset email and phish the new password.
//
// X-Forwarded-Host is honored only when the operator has opted into
// TrustProxy. Even then, the value is parsed via firstHeaderValue (multi-hop
// safe) and validated via validProxyHost so a misconfigured proxy that
// forwards client-supplied headers can't inject schemes, paths, CRLF, or
// extra hosts into the link.
//
// Returns "" when no trusted source is available; callers must refuse the
// operation in that case.
func SecureBaseURL(r *http.Request, options *config.Options) string {
	if options.Hostname != "" {
		return ensureScheme(options.Hostname, r, options.TrustProxy)
	}
	if options.TrustProxy {
		host := firstHeaderValue(r.Header.Get("X-Forwarded-Host"))
		if !validProxyHost(host) {
			return ""
		}
		return getScheme(r, options.TrustProxy) + "://" + host
	}
	return ""
}

// stripPathFromURL removes the path from a URL.
// ex. https://example.com/tools -> https://example.com
func stripPathFromURL(rawURL string) string {
	// Validate that the URL has a scheme; if not, return empty string
	if !strings.Contains(rawURL, "://") {
		log.Warn().Str("url", rawURL).Msg("URL missing scheme")
		return ""
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		log.Err(err).Msg("failed to parse URL")
		return ""
	}

	strippedURL := url.URL{Scheme: parsedURL.Scheme, Host: parsedURL.Host}

	return strippedURL.String()
}
