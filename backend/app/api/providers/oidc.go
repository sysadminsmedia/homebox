package providers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	config   *config.OIDCConf
	service  *services.UserService
	provider *oidc.Provider
	oauth2   oauth2.Config
}

type OIDCAuthRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func NewOIDCProvider(cfg *config.OIDCConf, service *services.UserService) (*OIDCProvider, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("OIDC is not enabled")
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	scopes := strings.Split(cfg.Scopes, " ")
	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	return &OIDCProvider{
		config:   cfg,
		service:  service,
		provider: provider,
		oauth2:   oauth2Config,
	}, nil
}

func (p *OIDCProvider) Name() string {
	return "oidc"
}

func (p *OIDCProvider) Authenticate(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Check if this is an authorization code callback
	code := r.URL.Query().Get("code")
	if code != "" {
		return p.HandleCallback(w, r)
	}

	// For OIDC, we need to parse the request to check if it's a redirect request
	// If no code is present, this means we should initiate the OAuth flow
	return services.UserAuthTokenDetail{}, fmt.Errorf("OIDC authentication requires redirection to provider")
}

// InitiateAuth starts the OIDC authentication flow by redirecting to the provider
func (p *OIDCProvider) InitiateAuth(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Generate a random state
	state, err := generateRandomState()
	if err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to generate state: %w", err)
	}

	// Store state in session/cookie for verification
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		MaxAge:   600, // 10 minutes
	})

	// Redirect to OIDC provider
	authURL := p.oauth2.AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)

	// Return empty token detail as this is a redirect
	return services.UserAuthTokenDetail{}, nil
}

// HandleCallback processes the OIDC callback with authorization code
func (p *OIDCProvider) HandleCallback(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Verify state parameter
	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie("oidc_state")
	if err != nil || cookie.Value != state {
		return services.UserAuthTokenDetail{}, fmt.Errorf("invalid state parameter")
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		MaxAge:   -1,
	})

	// Exchange authorization code for token
	code := r.URL.Query().Get("code")
	token, err := p.oauth2.Exchange(r.Context(), code)
	if err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return services.UserAuthTokenDetail{}, fmt.Errorf("no id_token in token response")
	}

	// Verify ID token
	verifier := p.provider.Verifier(&oidc.Config{ClientID: p.config.ClientID})
	idToken, err := verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract claims
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Extract user information
	username := p.getClaimValue(claims, p.config.UsernameClaim)
	email := p.getClaimValue(claims, p.config.EmailClaim)
	name := p.getClaimValue(claims, p.config.NameClaim)

	if username == "" && email == "" {
		return services.UserAuthTokenDetail{}, fmt.Errorf("no username or email found in claims")
	}

	// Use email as username if username is not available
	if username == "" {
		username = email
	}

	// Create or get user
	userDetail, err := p.service.GetOrCreateOIDCUser(r.Context(), username, email, name)
	if err != nil {
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to create or get user: %w", err)
	}

	return userDetail, nil
}

func (p *OIDCProvider) getClaimValue(claims map[string]interface{}, claimName string) string {
	if value, ok := claims[claimName]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthURL returns the OIDC authorization URL for frontend initiation
func (p *OIDCProvider) GetAuthURL() (string, error) {
	state, err := generateRandomState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	authURL := p.oauth2.AuthCodeURL(state)
	return authURL, nil
}