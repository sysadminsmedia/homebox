package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
)

const claimEmail = "email"

// testIdP is a minimal OIDC issuer: discovery + JWKS, and a signer for tokens.
type testIdP struct {
	srv    *httptest.Server
	signer jose.Signer
}

func newTestIdP(t *testing.T) *testIdP {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "k1", Algorithm: "RS256", Use: "sig"}}}
	idp := &testIdP{}
	mux := http.NewServeMux()
	mux.HandleFunc("/jwks/", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	})
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                idp.issuer(),
			"jwks_uri":                              idp.srv.URL + "/jwks/",
			"authorization_endpoint":                idp.srv.URL + "/authorize/",
			"token_endpoint":                        idp.srv.URL + "/token/",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})
	idp.srv = httptest.NewServer(mux)
	t.Cleanup(idp.srv.Close)

	idp.signer, err = jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: key, KeyID: "k1"}}, nil)
	require.NoError(t, err)
	return idp
}

func (p *testIdP) issuer() string { return p.srv.URL }

func (p *testIdP) token(t *testing.T, claims map[string]any) string {
	t.Helper()
	base := map[string]any{
		"iss":            p.issuer(),
		"aud":            "api-things",
		"sub":            "abc",
		"exp":            time.Now().Add(10 * time.Minute).Unix(),
		"iat":            time.Now().Unix(),
		claimEmail:       "auth-test@example.com",
		"email_verified": true,
	}
	for k, v := range claims {
		if v == nil {
			delete(base, k)
		} else {
			base[k] = v
		}
	}
	payload, err := json.Marshal(base)
	require.NoError(t, err)
	obj, err := p.signer.Sign(payload)
	require.NoError(t, err)
	raw, err := obj.CompactSerialize()
	require.NoError(t, err)
	return raw
}

func TestAuthTokenBearerJWT(t *testing.T) {
	idp := newTestIdP(t)

	for _, mode := range []struct {
		name string
		jwks string
	}{
		{"ExplicitJWKS", idp.srv.URL + "/jwks/"},
		{"Discovery", ""},
	} {
		t.Run(mode.name, func(t *testing.T) {
			a, tokens := newAuthTestApp(t)
			a.jwtAuth = newJWTAuthenticator(config.AuthJWTConf{
				IssuerURL:   idp.issuer(),
				Audience:    "api-things",
				JWKSURL:     mode.jwks,
				EmailClaim:  claimEmail,
				VerifyEmail: true,
			}, a.repos.Users)
			require.NotNil(t, a.jwtAuth)

			var seen *services.Context
			h := a.mwAuthToken(errchain.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) error {
				c := services.NewContext(r.Context())
				seen = &c
				return nil
			}))
			serve := func(token string) error {
				seen = nil
				req := httptest.NewRequest(http.MethodGet, "/api/v1/entities", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				return h.ServeHTTP(httptest.NewRecorder(), req)
			}
			unauthorized := func(t *testing.T, err error) {
				t.Helper()
				var reqErr *validate.RequestError
				require.ErrorAs(t, err, &reqErr)
				assert.Equal(t, http.StatusUnauthorized, reqErr.Status)
				assert.Nil(t, seen)
			}

			t.Run("ValidTokenMapsToExistingUser", func(t *testing.T) {
				require.NoError(t, serve(idp.token(t, nil)))
				require.NotNil(t, seen)
				require.NotNil(t, seen.User)
				assert.Equal(t, "auth-test@example.com", seen.User.Email)
				assert.NotEmpty(t, seen.User.GroupIDs)
			})

			t.Run("TreatedLikeAnAPIKey", func(t *testing.T) {
				var isKey bool
				h := a.mwAuthToken(errchain.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) error {
					isKey = services.IsAPIKeyAuth(r.Context())
					return nil
				}))
				req := httptest.NewRequest(http.MethodGet, "/api/v1/entities", nil)
				req.Header.Set("Authorization", "Bearer "+idp.token(t, nil))
				require.NoError(t, h.ServeHTTP(httptest.NewRecorder(), req))
				assert.True(t, isKey)
			})

			t.Run("EmailIsCaseInsensitive", func(t *testing.T) {
				require.NoError(t, serve(idp.token(t, map[string]any{claimEmail: "Auth-Test@Example.com"})))
			})

			t.Run("WrongAudienceRejected", func(t *testing.T) {
				unauthorized(t, serve(idp.token(t, map[string]any{"aud": "api-vehicles"})))
			})

			t.Run("WrongIssuerRejected", func(t *testing.T) {
				unauthorized(t, serve(idp.token(t, map[string]any{"iss": "https://elsewhere.example/"})))
			})

			t.Run("ExpiredRejected", func(t *testing.T) {
				unauthorized(t, serve(idp.token(t, map[string]any{"exp": time.Now().Add(-time.Minute).Unix()})))
			})

			t.Run("TamperedRejected", func(t *testing.T) {
				tok := idp.token(t, nil)
				unauthorized(t, serve(tok[:len(tok)-4]+"AAAA"))
			})

			t.Run("UnknownEmailRejectedAndNotCreated", func(t *testing.T) {
				unauthorized(t, serve(idp.token(t, map[string]any{claimEmail: "stranger@example.com"})))
				_, err := a.repos.Users.GetOneEmail(t.Context(), "stranger@example.com")
				assert.Error(t, err)
			})

			t.Run("UnverifiedEmailRejected", func(t *testing.T) {
				unauthorized(t, serve(idp.token(t, map[string]any{"email_verified": false})))
				unauthorized(t, serve(idp.token(t, map[string]any{"email_verified": nil})))
			})

			t.Run("MissingEmailRejected", func(t *testing.T) {
				unauthorized(t, serve(idp.token(t, map[string]any{claimEmail: nil})))
			})

			t.Run("SessionTokensStillWork", func(t *testing.T) {
				require.NoError(t, serve(tokens.Raw))
			})
		})
	}
}

func TestAuthTokenBearerJWTDisabledByDefault(t *testing.T) {
	// With JWT auth off the token falls through to the API key lookup.
	hasher.SetAPIKeyPepper([]byte("test-api-key-pepper-test-api-key-pepper"))
	idp := newTestIdP(t)
	a, _ := newAuthTestApp(t)
	assert.Nil(t, newJWTAuthenticator(config.AuthJWTConf{}, a.repos.Users))
	assert.Nil(t, newJWTAuthenticator(config.AuthJWTConf{IssuerURL: idp.issuer()}, a.repos.Users))

	h := a.mwAuthToken(errchain.HandlerFunc(func(http.ResponseWriter, *http.Request) error { return nil }))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/entities", nil)
	req.Header.Set("Authorization", "Bearer "+idp.token(t, nil))
	err := h.ServeHTTP(httptest.NewRecorder(), req)
	var reqErr *validate.RequestError
	require.ErrorAs(t, err, &reqErr)
	assert.Equal(t, http.StatusUnauthorized, reqErr.Status)
}
