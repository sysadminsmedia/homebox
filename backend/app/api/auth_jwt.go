package main

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

var errJWTRejected = errors.New("bearer JWT rejected")

// jwtAuthenticator validates bearer JWTs issued by an external identity
// provider (HBOX_AUTH_JWT_*) and maps them to an existing user by email.
// It never creates users: the person must have signed in to Homebox before.
type jwtAuthenticator struct {
	conf     config.AuthJWTConf
	users    *repo.UserRepository
	mu       sync.Mutex
	verifier *oidc.IDTokenVerifier
}

func newJWTAuthenticator(conf config.AuthJWTConf, users *repo.UserRepository) *jwtAuthenticator {
	if !conf.Enabled() {
		return nil
	}
	return &jwtAuthenticator{conf: conf, users: users}
}

// looksLikeJWT is a cheap shape check so session tokens and API keys never
// reach the verifier.
func looksLikeJWT(token string) bool {
	return strings.Count(token, ".") == 2 && strings.HasPrefix(token, "ey")
}

// getVerifier builds the verifier on first use, so a slow or unavailable
// identity provider never blocks startup. A failed discovery is retried on
// the next request.
func (j *jwtAuthenticator) getVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.verifier != nil {
		return j.verifier, nil
	}

	cfg := &oidc.Config{ClientID: j.conf.Audience}
	if j.conf.JWKSURL != "" {
		cfg.SupportedSigningAlgs = []string{oidc.RS256, oidc.RS384, oidc.RS512, oidc.ES256, oidc.ES384, oidc.ES512}
		// The key set fetches lazily and must outlive this request.
		keys := oidc.NewRemoteKeySet(context.Background(), j.conf.JWKSURL)
		j.verifier = oidc.NewVerifier(j.conf.IssuerURL, keys, cfg)
		return j.verifier, nil
	}

	provider, err := oidc.NewProvider(oidc.ClientContext(ctx, nil), j.conf.IssuerURL)
	if err != nil {
		return nil, err
	}
	j.verifier = provider.Verifier(cfg)
	return j.verifier, nil
}

// authenticate verifies the token and returns the user it belongs to.
func (j *jwtAuthenticator) authenticate(ctx context.Context, raw string) (repo.UserOut, error) {
	verifier, err := j.getVerifier(ctx)
	if err != nil {
		return repo.UserOut{}, err
	}

	tok, err := verifier.Verify(ctx, raw)
	if err != nil {
		return repo.UserOut{}, errJWTRejected
	}

	var claims map[string]any
	if err := tok.Claims(&claims); err != nil {
		return repo.UserOut{}, errJWTRejected
	}

	email, _ := claims[j.conf.EmailClaim].(string)
	if email == "" {
		return repo.UserOut{}, errJWTRejected
	}
	if j.conf.VerifyEmail {
		if verified, _ := claims["email_verified"].(bool); !verified {
			return repo.UserOut{}, errJWTRejected
		}
	}

	return j.users.GetOneEmail(ctx, email)
}
