package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

var errJWTRejected = errors.New("bearer JWT rejected")

const (
	// jwtFetchTimeout bounds OIDC discovery and every JWKS fetch.
	jwtFetchTimeout = 10 * time.Second
	// jwtDiscoveryCooldown is how long a failed discovery is remembered before
	// the next attempt, so an unreachable identity provider is not hammered.
	jwtDiscoveryCooldown = 30 * time.Second
)

var errJWTDiscoveryBusy = errors.New("identity provider discovery in progress")

// jwtAuthenticator validates bearer JWTs issued by an external identity
// provider (HBOX_AUTH_JWT_*) and maps them to an existing user by email.
// It never creates users: the person must have signed in to Homebox before.
type jwtAuthenticator struct {
	conf  config.AuthJWTConf
	users *repo.UserRepository

	// fetchTimeout and cooldown default to jwtFetchTimeout and
	// jwtDiscoveryCooldown; tests shorten them.
	fetchTimeout time.Duration
	cooldown     time.Duration

	mu          sync.Mutex // guards the fields below; never held across network calls
	verifier    *oidc.IDTokenVerifier
	discovering bool
	lastFail    time.Time
	lastErr     error
}

func newJWTAuthenticator(conf config.AuthJWTConf, users *repo.UserRepository) *jwtAuthenticator {
	if !conf.Enabled() {
		return nil
	}
	return &jwtAuthenticator{conf: conf, users: users, fetchTimeout: jwtFetchTimeout, cooldown: jwtDiscoveryCooldown}
}

// looksLikeJWT is a cheap shape check so session tokens and API keys never
// reach the verifier.
func looksLikeJWT(token string) bool {
	return strings.Count(token, ".") == 2 && strings.HasPrefix(token, "ey")
}

// getVerifier builds the verifier on first use, so a slow or unavailable
// identity provider never blocks startup. Discovery runs at most once at a
// time, is bounded by fetchTimeout, and does not hold the lock; requests that
// arrive meanwhile, or within cooldown of a failed attempt, are refused at
// once instead of waiting. A successful verifier is cached.
func (j *jwtAuthenticator) getVerifier() (*oidc.IDTokenVerifier, error) {
	j.mu.Lock()
	if j.verifier != nil {
		defer j.mu.Unlock()
		return j.verifier, nil
	}

	// Discovery and the key set's lazy JWKS fetches both use this client; the
	// context must outlive any single request.
	client := &http.Client{Timeout: j.fetchTimeout}
	clientCtx := oidc.ClientContext(context.Background(), client)
	cfg := &oidc.Config{ClientID: j.conf.Audience}

	if j.conf.JWKSURL != "" {
		defer j.mu.Unlock()
		cfg.SupportedSigningAlgs = []string{oidc.RS256, oidc.RS384, oidc.RS512, oidc.ES256, oidc.ES384, oidc.ES512}
		j.verifier = oidc.NewVerifier(j.conf.IssuerURL, oidc.NewRemoteKeySet(clientCtx, j.conf.JWKSURL), cfg)
		return j.verifier, nil
	}

	if !j.lastFail.IsZero() && time.Since(j.lastFail) < j.cooldown {
		err := j.lastErr
		j.mu.Unlock()
		return nil, err
	}
	if j.discovering {
		j.mu.Unlock()
		return nil, errJWTDiscoveryBusy
	}
	j.discovering = true
	j.mu.Unlock()

	ctx, cancel := context.WithTimeout(clientCtx, j.fetchTimeout)
	provider, err := oidc.NewProvider(ctx, j.conf.IssuerURL)
	cancel()

	j.mu.Lock()
	defer j.mu.Unlock()
	j.discovering = false
	if err != nil {
		j.lastFail, j.lastErr = time.Now(), err
		return nil, err
	}
	j.lastFail, j.lastErr = time.Time{}, nil
	j.verifier = provider.Verifier(cfg)
	return j.verifier, nil
}

// authenticate verifies the token and returns the user it belongs to.
func (j *jwtAuthenticator) authenticate(ctx context.Context, raw string) (repo.UserOut, error) {
	verifier, err := j.getVerifier()
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
