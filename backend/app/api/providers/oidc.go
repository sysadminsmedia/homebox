package providers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	config    *config.OIDCConf
	userSvc   *services.UserService
	oauth2Cfg *oauth2.Config
	verifier  *oidc.IDTokenVerifier
	provider  *oidc.Provider
	states    map[string]time.Time // Simple state storage - in production, use Redis or database
}

func NewOIDCProvider(cfg *config.OIDCConf, userSvc *services.UserService) (*OIDCProvider, error) {
	if !cfg.Enabled {
		return nil, errors.New("OIDC is not enabled")
	}

	ctx := context.Background()
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
		states:    make(map[string]time.Time),
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

	// Store state with expiration
	p.states[state] = time.Now().Add(10 * time.Minute)

	// Generate authorization URL
	authURL := p.oauth2Cfg.AuthCodeURL(state)

	// Redirect to OIDC provider
	http.Redirect(w, r, authURL, http.StatusFound)

	// Return empty token since this is a redirect
	return services.UserAuthTokenDetail{}, nil
}

func (p *OIDCProvider) handleCallback(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Verify state parameter
	state := r.URL.Query().Get("state")
	if !p.verifyState(state) {
		return services.UserAuthTokenDetail{}, errors.New("invalid state parameter")
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		return services.UserAuthTokenDetail{}, errors.New("authorization code not found")
	}

	// Exchange code for token
	ctx := r.Context()
	token, err := p.oauth2Cfg.Exchange(ctx, code)
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
		Email         string   `json:"email"`
		Name          string   `json:"name"`
		Subject       string   `json:"sub"`
		EmailVerified bool     `json:"email_verified"`
		Groups        []string `json:"groups"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to parse claims: %w", err)
	}

	// Verify email is verified
	if !claims.EmailVerified {
		return services.UserAuthTokenDetail{}, errors.New("email not verified")
	}

	// Determine user role based on groups
	role := p.determineUserRole(claims.Groups)

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

func (p *OIDCProvider) verifyState(state string) bool {
	expiration, exists := p.states[state]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(expiration) {
		delete(p.states, state)
		return false
	}

	// Clean up state after use
	delete(p.states, state)
	return true
}

func (p *OIDCProvider) determineUserRole(groups []string) string {
	for _, group := range groups {
		if group == p.config.AdminRole {
			return "owner"
		}
	}
	return "user"
}

func (p *OIDCProvider) createOrGetUser(ctx context.Context, email, name, role string) (repo.UserOut, error) {
	// Try to get existing user by email
	// Note: We'll need to implement GetByEmail in the user service if it doesn't exist
	// For now, we'll create the user directly through the registration process

	// Create new OIDC user registration
	registration := services.UserRegistration{
		Email:    email,
		Name:     name,
		Password: "oidc-user", // Placeholder password that won't be used
	}

	// Try to register the user - this will fail if user already exists
	createdUser, err := p.userSvc.RegisterUser(ctx, registration)
	if err != nil {
		// If registration fails due to existing email, try to login instead
		if strings.Contains(err.Error(), "email") {
			// User exists, we need to get them - for now return error
			// TODO: Implement proper user lookup by email in repo/services
			return repo.UserOut{}, fmt.Errorf("OIDC user already exists but cannot retrieve: %w", err)
		}
		return repo.UserOut{}, fmt.Errorf("failed to register OIDC user: %w", err)
	}

	log.Info().
		Str("email", email).
		Str("name", name).
		Str("role", role).
		Msg("created new OIDC user")

	return createdUser, nil
}

// CleanupExpiredStates should be called periodically to clean up expired states
func (p *OIDCProvider) CleanupExpiredStates() {
	now := time.Now()
	for state, expiration := range p.states {
		if now.After(expiration) {
			delete(p.states, state)
		}
	}
}
