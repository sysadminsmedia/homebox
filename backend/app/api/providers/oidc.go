package providers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/pkgs/hasher"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	config    *config.OIDCConf
	userSvc   *services.UserService
	oauth2Cfg *oauth2.Config
	verifier  *oidc.IDTokenVerifier
	provider  *oidc.Provider
	states    map[string]pkceState // state -> verifier+nonce+expiry (use Redis in production)
	statesMu  sync.RWMutex
}

type pkceState struct {
	Verifier string
	Nonce    string
	Expires  time.Time
}

func NewOIDCProvider(ctx context.Context, cfg *config.OIDCConf, userSvc *services.UserService) (*OIDCProvider, error) {
	if !cfg.Enabled {
		return nil, errors.New("OIDC is not enabled")
	}

	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC provider: %w", err)
	}

	scopes := strings.Split(cfg.Scopes, " ")
	oauth2Cfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	return &OIDCProvider{
		config:    cfg,
		userSvc:   userSvc,
		oauth2Cfg: oauth2Cfg,
		verifier:  verifier,
		provider:  provider,
		states:    make(map[string]pkceState),
	}, nil
}

func (p *OIDCProvider) Name() string {
	return "oidc"
}

func (p *OIDCProvider) Authenticate(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Handle different OIDC flows based on request
	if strings.Contains(r.URL.Path, "callback") {
		return p.handleCallback(w, r)
	}
	return p.handleLogin(w, r)
}

func (p *OIDCProvider) handleLogin(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Generate state parameter
	state, err := p.generateState()
	if err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to generate state: %w", err)
	}

	// PKCE: generate code_verifier and S256 code_challenge, plus nonce
	verifier := mustRandomString(64)
	challenge := codeChallengeS256(verifier)
	nonce := mustRandomString(32)

	// Store state with expiration (write lock)
	p.statesMu.Lock()
	p.states[state] = pkceState{Verifier: verifier, Nonce: nonce, Expires: time.Now().Add(10 * time.Minute)}
	p.statesMu.Unlock()

	// Generate authorization URL with PKCE and nonce
	authURL := p.oauth2Cfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", nonce),
	)

	// Redirect to OIDC provider
	http.Redirect(w, r, authURL, http.StatusFound)

	// Return empty token since this is a redirect
	return services.UserAuthTokenDetail{}, nil
}

func (p *OIDCProvider) handleCallback(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Verify state parameter and retrieve PKCE/nonce
	state := r.URL.Query().Get("state")
	ps, ok := p.popState(state)
	if !ok {
		return services.UserAuthTokenDetail{}, errors.New("invalid state parameter")
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		return services.UserAuthTokenDetail{}, errors.New("authorization code not found")
	}

	// Exchange code for token with PKCE code_verifier (bounded context)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	token, err := p.oauth2Cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", ps.Verifier))
	if err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to exchange token: %w", err)
	}

	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return services.UserAuthTokenDetail{}, errors.New("no id_token field in oauth2 token")
	}

	// Verify ID token
	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract claims
	var claims struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		Subject       string `json:"sub"`
		EmailVerified bool   `json:"email_verified"`
		Nonce         string `json:"nonce"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to parse claims: %w", err)
	}

	// Verify email is verified
	if !claims.EmailVerified {
		return services.UserAuthTokenDetail{}, errors.New("email not verified")
	}
	// Verify nonce matches what we generated
	if claims.Nonce == "" || claims.Nonce != ps.Nonce {
		return services.UserAuthTokenDetail{}, errors.New("invalid id token nonce")
	}

	// Extract roles from configured claim and determine user role
	roles := p.extractRolesFromIDToken(idToken, p.config.RolesClaim)
	role := p.determineUserRole(roles)

	// Create or get user
	userOut, err := p.createOrGetUser(ctx, claims.Email, claims.Name, role)
	if err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to create or get user: %w", err)
	}

	// Generate authentication token for OIDC authenticated user
	authToken, err := p.userSvc.CreateSessionTokenForUser(ctx, userOut.ID, true) // Extended session for OIDC
	if err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to generate session token: %w", err)
	}

	return authToken, nil
}

func (p *OIDCProvider) generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// mustRandomString returns a URL-safe base64 string of n random bytes or panics
func mustRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

// codeChallengeS256 returns the base64url-encoded SHA256 of the verifier
func codeChallengeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h[:])
}

// popState returns the pkce state and removes it atomically if present and not expired
func (p *OIDCProvider) popState(state string) (pkceState, bool) {
	p.statesMu.Lock()
	defer p.statesMu.Unlock()
	ps, exists := p.states[state]
	if !exists {
		return pkceState{}, false
	}
	if time.Now().After(ps.Expires) {
		delete(p.states, state)
		return pkceState{}, false
	}
	delete(p.states, state)
	return ps, true
}

func (p *OIDCProvider) determineUserRole(roles []string) string {
	for _, r := range roles {
		if r == p.config.AdminRole {
			return "owner"
		}
	}
	return "user"
}

// extractRolesFromIDToken reads roles from a dynamic claim. Supports:
// - string
// - []string
// - nested dotted paths (e.g., "realm_access.roles")
func (p *OIDCProvider) extractRolesFromIDToken(idToken *oidc.IDToken, claimPath string) []string {
	var raw map[string]any
	if err := idToken.Claims(&raw); err != nil {
		return nil
	}
	val := getByDottedPath(raw, claimPath)
	switch v := val.(type) {
	case string:
		return []string{v}
	case []any:
		out := make([]string, 0, len(v))
		for _, it := range v {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func getByDottedPath(m map[string]any, path string) any {
	if path == "" {
		return nil
	}
	cur := any(m)
	for _, part := range strings.Split(path, ".") {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = obj[part]
		if !ok {
			return nil
		}
	}
	return cur
}

func (p *OIDCProvider) createOrGetUser(ctx context.Context, email, name, role string) (repo.UserOut, error) {
	// Check if user already exists by email
	existing, err := p.userSvc.GetByEmail(ctx, email)
	if err == nil {
		return existing, nil
	}

	// Create new OIDC user registration
	randomPassword := hasher.GenerateToken()
	registration := services.UserRegistration{
		Email:    email,
		Name:     name,
		Password: randomPassword.Raw,
	}

	createdUser, err := p.userSvc.RegisterUser(ctx, registration)
	if err != nil {
		// If creation fails because user exists, fetch and return
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "unique") || strings.Contains(errStr, "users.email") || strings.Contains(errStr, "email") {
			if existing, gerr := p.userSvc.GetByEmail(ctx, email); gerr == nil {
				return existing, nil
			}
		}
		return repo.UserOut{}, fmt.Errorf("failed to register OIDC user: %w", err)
	}

	log.Info().Str("email", email).Str("name", name).Str("role", role).Msg("created new OIDC user")
	return createdUser, nil
}

// CleanupExpiredStates should be called periodically to clean up expired states
func (p *OIDCProvider) CleanupExpiredStates() {
	now := time.Now()
	// Write lock for iteration + deletion safety
	p.statesMu.Lock()
	for state, expiration := range p.states {
		if now.After(expiration) {
			delete(p.states, state)
		}
	}
	p.statesMu.Unlock()
}
