package providers

import (
	"context"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"chainguard.dev/go-oidctest/pkg/oidctest"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

// newDiscoveryServer serves the issuer's discovery document from another
// address, so its URL differs from the issuer named in the document.
func newDiscoveryServer(t *testing.T, issuer string) string {
	t.Helper()
	target, err := url.Parse(issuer)
	require.NoError(t, err)
	srv := httptest.NewServer(httputil.NewSingleHostReverseProxy(target))
	t.Cleanup(srv.Close)
	return srv.URL
}

// idToken signs an ID token for the homebox client, as issued by issuer.
func idToken(t *testing.T, signer jose.Signer, issuer string) string {
	t.Helper()
	token, err := jwt.Signed(signer).Claims(jwt.Claims{
		Issuer:   issuer,
		Subject:  "user",
		Audience: jwt.Audience{"homebox"},
		Expiry:   jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).Serialize()
	require.NoError(t, err)
	return token
}

func oidcConf(issuer, discovery string) *config.OIDCConf {
	return &config.OIDCConf{
		Enabled:        true,
		IssuerURL:      issuer,
		DiscoveryURL:   discovery,
		ClientID:       "homebox",
		ClientSecret:   "secret",
		RequestTimeout: 5 * time.Second,
	}
}

func TestNewOIDCProviderDiscoversFromDiscoveryURL(t *testing.T) {
	_, issuer := oidctest.NewIssuer(t)

	p, err := NewOIDCProvider(nil, oidcConf(issuer, newDiscoveryServer(t, issuer)), nil, true)
	require.NoError(t, err)
	assert.Equal(t, issuer+"/token", p.endpoint.TokenURL)
	assert.Equal(t, issuer+"/authz", p.endpoint.AuthURL)
}

func TestNewOIDCProviderWithDiscoveryURLExpectsTheIssuersTokens(t *testing.T) {
	signer, issuer := oidctest.NewIssuer(t)

	p, err := NewOIDCProvider(nil, oidcConf(issuer, newDiscoveryServer(t, issuer)), nil, true)
	require.NoError(t, err)

	_, err = p.verifier.Verify(context.Background(), idToken(t, signer, issuer))
	require.NoError(t, err)

	_, err = p.verifier.Verify(context.Background(), idToken(t, signer, "https://other.example.com"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "issued by a different provider")
}
