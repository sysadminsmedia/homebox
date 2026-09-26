package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// TestSwaggerDocHostNotAttackerControlled guards the unauthenticated
// /swagger/doc.json route. The `host` it emits is the target of every "Try it
// out" request the Swagger UI issues, Authorization header included. Honoring a
// raw X-Forwarded-Host let any unauthenticated client choose that host; behind a
// cache that does not key on the header, the poisoned spec would then be served
// to operators. X-Forwarded-Host must be ignored unless the operator opted into
// TrustProxy.
func TestSwaggerDocHostNotAttackerControlled(t *testing.T) {
	const (
		realHost = "homebox.example:3007"
		evilHost = "evil.attacker.example"
	)

	docHost := func(t *testing.T, opts config.Options, headers map[string]string) string {
		t.Helper()

		a := &app{conf: &config.Config{Options: opts}}

		req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
		req.Host = realHost
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		w := httptest.NewRecorder()
		a.swaggerDocHandler(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var spec struct {
			Host string `json:"host"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
			t.Fatalf("decode spec: %v", err)
		}
		return spec.Host
	}

	t.Run("NoHeaderUsesRequestHost", func(t *testing.T) {
		if got := docHost(t, config.Options{}, nil); got != realHost {
			t.Errorf("host = %q, want %q", got, realHost)
		}
	})

	t.Run("ForwardedHostIgnoredByDefault", func(t *testing.T) {
		got := docHost(t, config.Options{}, map[string]string{"X-Forwarded-Host": evilHost})
		if got == evilHost {
			t.Fatalf("X-Forwarded-Host was honored without TrustProxy: host = %q", got)
		}
		if got != realHost {
			t.Errorf("host = %q, want %q", got, realHost)
		}
	})

	t.Run("ConfiguredHostnameWins", func(t *testing.T) {
		opts := config.Options{Hostname: "https://canonical.example"}
		got := docHost(t, opts, map[string]string{"X-Forwarded-Host": evilHost})
		if got != "canonical.example" {
			t.Errorf("host = %q, want %q", got, "canonical.example")
		}
	})

	t.Run("ForwardedHostHonoredWithTrustProxy", func(t *testing.T) {
		opts := config.Options{TrustProxy: true}
		got := docHost(t, opts, map[string]string{"X-Forwarded-Host": "proxy.example"})
		if got != "proxy.example" {
			t.Errorf("host = %q, want %q", got, "proxy.example")
		}
	})

	t.Run("MalformedForwardedHostRejectedUnderTrustProxy", func(t *testing.T) {
		opts := config.Options{TrustProxy: true}
		got := docHost(t, opts, map[string]string{"X-Forwarded-Host": "evil.example/path?x=1"})
		if got != realHost {
			t.Errorf("host = %q, want fallback %q", got, realHost)
		}
	})
}
