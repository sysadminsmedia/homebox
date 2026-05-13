package v1

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// fixtureAppURL is the canonical URL used across SecureBaseURL test cases.
// Extracted to satisfy goconst.
const fixtureAppURL = "https://app.example.com"

func TestSecureBaseURL(t *testing.T) {
	cases := []struct {
		name        string
		hostname    string
		trustProxy  bool
		xfHost      string
		xfProto     string
		expectURL   string
		expectEmpty bool
	}{
		{
			name:      "Hostname configured wins outright",
			hostname:  fixtureAppURL,
			expectURL: fixtureAppURL,
		},
		{
			name:        "no hostname, no trust proxy → refuse",
			expectEmpty: true,
		},
		{
			name:        "trust proxy but no XF-Host → refuse",
			trustProxy:  true,
			expectEmpty: true,
		},
		{
			name:       "trust proxy + XF-Host (https)",
			trustProxy: true,
			xfHost:     "app.example.com",
			xfProto:    schemeHTTPS,
			expectURL:  fixtureAppURL,
		},
		{
			name:       "trust proxy + XF-Host with port",
			trustProxy: true,
			xfHost:     "app.example.com:8443",
			xfProto:    schemeHTTPS,
			expectURL:  "https://app.example.com:8443",
		},
		{
			name:       "trust proxy + IPv6 XF-Host",
			trustProxy: true,
			xfHost:     "[2001:db8::1]:8443",
			xfProto:    schemeHTTPS,
			expectURL:  "https://[2001:db8::1]:8443",
		},

		// Multi-hop / multi-value handling: take the first comma-separated
		// value, mirroring the leftmost-is-original convention used by
		// reverse proxies.
		{
			name:       "comma-separated XF-Host: take first",
			trustProxy: true,
			xfHost:     "app.example.com, internal.example.com",
			xfProto:    schemeHTTPS,
			expectURL:  fixtureAppURL,
		},
		{
			name:       "comma-separated XF-Proto: take first",
			trustProxy: true,
			xfHost:     "app.example.com",
			xfProto:    "https, http",
			expectURL:  fixtureAppURL,
		},

		// Validation rejections — these are the attack-relevant cases. A
		// misconfigured proxy that forwards client-supplied X-Forwarded-Host
		// without overwriting could otherwise let an attacker plant any of
		// these into the password-reset link.
		{
			name:        "XF-Host with embedded scheme is rejected",
			trustProxy:  true,
			xfHost:      "https://evil.com",
			expectEmpty: true,
		},
		{
			name:        "XF-Host with path is rejected",
			trustProxy:  true,
			xfHost:      "evil.com/path",
			expectEmpty: true,
		},
		{
			name:        "XF-Host with query is rejected",
			trustProxy:  true,
			xfHost:      "evil.com?x=1",
			expectEmpty: true,
		},
		{
			name:        "XF-Host with fragment is rejected",
			trustProxy:  true,
			xfHost:      "evil.com#frag",
			expectEmpty: true,
		},
		{
			name:        "XF-Host with whitespace is rejected",
			trustProxy:  true,
			xfHost:      "evil .com",
			expectEmpty: true,
		},
		{
			name:        "XF-Host with backslash is rejected",
			trustProxy:  true,
			xfHost:      `evil.com\foo`,
			expectEmpty: true,
		},
		{
			name:        "XF-Host with embedded CR is rejected",
			trustProxy:  true,
			xfHost:      "evil.com\rfoo",
			expectEmpty: true,
		},
		{
			name:        "empty XF-Host is rejected",
			trustProxy:  true,
			xfHost:      "",
			expectEmpty: true,
		},

		// Without TrustProxy, even legitimately-set XF-Host is ignored.
		{
			name:        "untrusted proxy headers are ignored",
			trustProxy:  false,
			xfHost:      "app.example.com",
			xfProto:     schemeHTTPS,
			expectEmpty: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/users/forgot-password", nil)
			if tc.xfHost != "" {
				req.Header.Set("X-Forwarded-Host", tc.xfHost)
			}
			if tc.xfProto != "" {
				req.Header.Set("X-Forwarded-Proto", tc.xfProto)
			}

			opts := &config.Options{
				Hostname:   tc.hostname,
				TrustProxy: tc.trustProxy,
			}

			got := SecureBaseURL(req, opts)
			if tc.expectEmpty {
				assert.Empty(t, got, "expected empty (refuse), got %q", got)
				return
			}
			assert.Equal(t, tc.expectURL, got)
		})
	}
}

func TestValidProxyHost(t *testing.T) {
	good := []string{
		"example.com",
		"example.com:8080",
		"sub.example.com",
		"127.0.0.1",
		"127.0.0.1:8080",
		"[::1]",
		"[::1]:8080",
		"[2001:db8::1]:8443",
	}
	for _, h := range good {
		t.Run("good/"+h, func(t *testing.T) {
			assert.True(t, validProxyHost(h), "expected valid: %q", h)
		})
	}

	bad := []string{
		"",
		"https://example.com",
		"example.com/path",
		"example.com?q=1",
		"example.com#frag",
		"example .com",
		"example.com\r\nX-Other: foo",
		"example.com\n",
		"foo\\bar",
		"example.com:abc:8080:more", // malformed
	}
	for _, h := range bad {
		t.Run("bad/"+h, func(t *testing.T) {
			assert.False(t, validProxyHost(h), "expected invalid: %q", h)
		})
	}
}

func TestFirstHeaderValue(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"foo", "foo"},
		{"foo, bar", "foo"},
		{"foo,bar", "foo"},
		{"  foo , bar", "foo"},
		{"https, http", schemeHTTPS},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			assert.Equal(t, c.want, firstHeaderValue(c.in))
		})
	}
}
