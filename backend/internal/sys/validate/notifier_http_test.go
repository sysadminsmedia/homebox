package validate

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// redirectTo builds an http.Request for a redirect target so the guard can be
// exercised without a live server. IP-literal targets resolve via net.LookupIP
// without touching DNS, keeping these cases deterministic.
func redirectTo(rawURL string) *http.Request {
	u, _ := url.Parse(rawURL)
	return &http.Request{URL: u}
}

func TestNotifierRedirectGuard_BlocksByPolicy(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.NotifierConf
		target  string
		blocked bool
	}{
		{"loopback blocked", &config.NotifierConf{BlockLocalhost: true}, "http://127.0.0.1/x", true},
		{"loopback allowed when flag off", &config.NotifierConf{}, "http://127.0.0.1/x", false},
		{"cloud metadata blocked", &config.NotifierConf{BlockCloudMetadata: true}, "http://169.254.169.254/latest/meta-data", true},
		{"rfc1918 blocked", &config.NotifierConf{BlockLocalNets: true}, "http://10.1.2.3/x", true},
		{"bogon link-local blocked", &config.NotifierConf{BlockBogonNets: true}, "http://169.254.1.1/x", true},
		{"blocknet CIDR blocked", &config.NotifierConf{BlockNets: []string{"203.0.113.0/24"}}, "http://203.0.113.7/x", true},
		{"public address allowed", &config.NotifierConf{BlockLocalhost: true, BlockLocalNets: true, BlockBogonNets: true, BlockCloudMetadata: true}, "http://8.8.8.8/x", false},
		{"allowlist rejects outside", &config.NotifierConf{AllowNets: []string{"8.8.8.0/24"}}, "http://127.0.0.1/x", true},
		{"allowlist permits inside", &config.NotifierConf{AllowNets: []string{"8.8.8.0/24"}}, "http://8.8.8.8/x", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			guard := NotifierRedirectGuard(tc.cfg)
			err := guard(redirectTo(tc.target), nil)
			if tc.blocked {
				require.Error(t, err, "target should be refused")
			} else {
				require.NoError(t, err, "target should be allowed")
			}
		})
	}
}

func TestNotifierRedirectGuard_CapsHops(t *testing.T) {
	guard := NotifierRedirectGuard(&config.NotifierConf{})
	via := make([]*http.Request, maxNotifierRedirects)
	err := guard(redirectTo("http://8.8.8.8/x"), via)
	assert.Error(t, err, "should stop once the redirect cap is reached")
}

func TestNotifierRedirectGuard_NilConfigOnlyCaps(t *testing.T) {
	guard := NotifierRedirectGuard(nil)
	// With no policy, a normal hop is allowed...
	require.NoError(t, guard(redirectTo("http://127.0.0.1/x"), nil))
	// ...but the hop cap still applies.
	assert.Error(t, guard(redirectTo("http://127.0.0.1/x"), make([]*http.Request, maxNotifierRedirects)))
}
