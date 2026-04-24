package mid

import (
	"net/http"
)

// SecurityHeaders is a middleware that will set security headers on the response
// It includes recommended headers from OWASP that are safe for self-hosted applications.
// Reference: https://owasp.org/www-project-secure-headers/
//
// In demo mode the clipboard-read directive is relaxed to `self` so E2E tests
// (and anyone interacting with the public demo) can read the clipboard for
// copy/paste flows. Production-style deployments keep clipboard-read disabled.
func SecurityHeaders(demo bool) func(http.Handler) http.Handler {
	clipboardRead := "clipboard-read=()"
	if demo {
		clipboardRead = "clipboard-read=(self)"
	}
	permissionsPolicy := "accelerometer=(), autoplay=(), camera=(self), cross-origin-isolated=(), display-capture=(), encrypted-media=(), fullscreen=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), payment=(), picture-in-picture=(), publickey-credentials-get=(), screen-wake-lock=(), sync-xhr=(self), usb=(), web-share=(), xr-spatial-tracking=(), " + clipboardRead + ", clipboard-write=(self), gamepad=(), hid=(), idle-detection=(), interest-cohort=(), serial=(), unload=()"

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Origin-Embedder-Policy", "require-corp")
			w.Header().Set("Content-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Content-Origin-Resource-Policy", "same-site")
			w.Header().Set("Permissions-Policy", permissionsPolicy)
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
			h.ServeHTTP(w, r)
		})
	}
}

// MaxBodySize is a middleware that limits the size of the request body.
// If the request body exceeds the specified maxBytes, the middleware will
// return a 413 Request Entity Too Large response.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes*1024*1024) // maxBytes in MB
			h.ServeHTTP(w, r)
		})
	}
}
