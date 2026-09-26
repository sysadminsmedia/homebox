package mid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The cross-origin isolation headers were spelled "Content-Origin-*" instead of
// "Cross-Origin-*". Browsers ignore header names they do not recognise, so all
// three policies were silently inert rather than merely misconfigured, and
// nothing in the response looked wrong enough to notice. This is issue #1665.
func TestSecurityHeadersUsesStandardCrossOriginNames(t *testing.T) {
	rec := httptest.NewRecorder()

	handler := SecurityHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	want := map[string]string{
		"Cross-Origin-Embedder-Policy":      "require-corp",
		"Cross-Origin-Opener-Policy":        "same-origin",
		"Cross-Origin-Resource-Policy":      "same-site",
		"Referrer-Policy":                   "no-referrer",
		"X-Content-Type-Options":            "nosniff",
		"X-Frame-Options":                   "DENY",
		"X-Permitted-Cross-Domain-Policies": "none",
	}
	for name, value := range want {
		assert.Equal(t, value, rec.Header().Get(name), "%s should be set", name)
	}

	// Guard against the misspelling coming back: a stray "Content-Origin-*"
	// header is indistinguishable from a working one without checking the
	// exact name, which is what let this go unnoticed in the first place.
	for _, name := range []string{
		"Content-Origin-Embedder-Policy",
		"Content-Origin-Opener-Policy",
		"Content-Origin-Resource-Policy",
	} {
		assert.Empty(t, rec.Header().Values(name), "%s is not a real header and should not be sent", name)
	}
}
