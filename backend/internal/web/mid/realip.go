package mid

import (
	"net/http"
	"strings"
)

var (
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
)

// RealIP rewrites the request's RemoteAddr from the X-Real-IP or
// X-Forwarded-For header, but ONLY when trustProxy is enabled. It replaces
// chi's middleware.RealIP, which is deprecated precisely because it trusts
// those headers unconditionally: without a reverse proxy that strips and sets
// them, any client can spoof its source IP, defeating rate limiting and
// poisoning request logs (see GHSA-9g5q-2w5x-hmxf).
//
// When trustProxy is false the request's RemoteAddr (the real TCP peer) is left
// untouched. This mirrors the precedence used by the auth rate limiter's
// extractClientIP so both see the same client address.
func RealIP(trustProxy bool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		if !trustProxy {
			return h
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip := realIP(r); ip != "" {
				r.RemoteAddr = ip
			}
			h.ServeHTTP(w, r)
		})
	}
}

// realIP returns the client IP from the proxy headers: X-Real-IP first, then
// the leftmost (original client) entry of X-Forwarded-For.
func realIP(r *http.Request) string {
	if xrip := strings.TrimSpace(r.Header.Get(xRealIP)); xrip != "" {
		return xrip
	}
	if xff := r.Header.Get(xForwardedFor); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	return ""
}
