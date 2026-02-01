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
		return "https"
	}
	if trustProxy && r.Header.Get("X-Forwarded-Proto") == "https" {
		return "https"
	}
	return "http"
}

// stripPathFromURL removes the path from a URL.
// ex. https://example.com/tools -> https://example.com
func stripPathFromURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		log.Err(err).Msg("failed to parse URL")
		return ""
	}

	strippedURL := url.URL{Scheme: parsedURL.Scheme, Host: parsedURL.Host}

	return strippedURL.String()
}
