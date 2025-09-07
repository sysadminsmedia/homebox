package providers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	service      *services.UserService
	config       *config.OIDCConf
	options      *config.Options
	cookieSecure bool
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	endpoint     oauth2.Endpoint
}

type DiscoveryDocument struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JwksURI               string `json:"jwks_uri"`
	Issuer                string `json:"issuer"`
}

type OIDCClaims struct {
	Email   string
	Groups  []string
	Name    string
	Subject string
}

func NewOIDCProvider(service *services.UserService, config *config.OIDCConf, options *config.Options, cookieSecure bool) (*OIDCProvider, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("OIDC is not enabled")
	}

	// Validate required configuration
	if config.ClientID == "" {
		return nil, fmt.Errorf("OIDC client ID is required when OIDC is enabled (set HBOX_OIDC_CLIENT_ID)")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("OIDC client secret is required when OIDC is enabled (set HBOX_OIDC_CLIENT_SECRET)")
	}
	if config.IssuerURL == "" {
		return nil, fmt.Errorf("OIDC issuer URL is required when OIDC is enabled (set HBOX_OIDC_ISSUER_URL)")
	}

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider from issuer URL: %w", err)
	}

	// Create ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.ClientID,
	})

	log.Info().
		Str("issuer", config.IssuerURL).
		Str("client_id", config.ClientID).
		Str("scope", config.Scope).
		Msg("OIDC provider initialized successfully with discovery")

	return &OIDCProvider{
		service:      service,
		config:       config,
		options:      options,
		cookieSecure: cookieSecure,
		provider:     provider,
		verifier:     verifier,
		endpoint:     provider.Endpoint(),
	}, nil
}

func (p *OIDCProvider) Name() string {
	return "oidc"
}

// Authenticate implements the AuthProvider interface but is not used for OIDC
// OIDC uses dedicated endpoints: GET /api/v1/users/login/oidc and GET /api/v1/users/login/oidc/callback
func (p *OIDCProvider) Authenticate(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	return services.UserAuthTokenDetail{}, fmt.Errorf("OIDC authentication uses dedicated endpoints: /api/v1/users/login/oidc")
}

// AuthenticateWithBaseURL is the main authentication method that requires baseURL
// This is now only called from handleCallback after state verification
func (p *OIDCProvider) AuthenticateWithBaseURL(baseURL string, w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return services.UserAuthTokenDetail{}, fmt.Errorf("missing authorization code")
	}

	// Get OAuth2 config for this request
	oauth2Config := p.getOAuth2Config(baseURL)

	// Exchange code for token with timeout
	ctx, cancel := context.WithTimeout(r.Context(), p.config.RequestTimeout)
	defer cancel()

	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Err(err).Msg("failed to exchange OIDC code for token")
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to exchange code for token")
	}

	// Extract ID token
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return services.UserAuthTokenDetail{}, fmt.Errorf("no id_token in response")
	}

	// Parse and validate the ID token using the library's verifier with timeout
	verifyCtx, verifyCancel := context.WithTimeout(r.Context(), p.config.RequestTimeout)
	defer verifyCancel()

	idTokenStruct, err := p.verifier.Verify(verifyCtx, idToken)
	if err != nil {
		log.Err(err).Msg("failed to verify ID token")
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to verify ID token")
	}

	// Extract claims from the verified token using dynamic parsing
	var rawClaims map[string]interface{}
	if err := idTokenStruct.Claims(&rawClaims); err != nil {
		log.Err(err).Msg("failed to extract claims from ID token")
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to extract claims from ID token")
	}

	// Parse claims using configurable claim names
	claims, err := p.parseOIDCClaims(rawClaims)
	if err != nil {
		log.Err(err).Msg("failed to parse OIDC claims")
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to parse OIDC claims: %w", err)
	}

	// Check group authorization if configured
	if p.config.AllowedGroups != "" {
		allowedGroups := strings.Split(p.config.AllowedGroups, ",")
		if !p.hasAllowedGroup(claims.Groups, allowedGroups) {
			log.Warn().
				Strs("user_groups", claims.Groups).
				Strs("allowed_groups", allowedGroups).
				Str("user", claims.Email).
				Msg("user not in allowed groups")
			return services.UserAuthTokenDetail{}, fmt.Errorf("user not in allowed groups")
		}
	}

	// Determine username from claims
	email := claims.Email
	if email == "" {
		return services.UserAuthTokenDetail{}, fmt.Errorf("no email found in token claims")
	}

	// Use the dedicated OIDC login method
	sessionToken, err := p.service.LoginOIDC(r.Context(), email, claims.Name)
	if err != nil {
		log.Err(err).Str("email", email).Msg("OIDC login failed")
		return services.UserAuthTokenDetail{}, fmt.Errorf("OIDC login failed: %w", err)
	}

	return sessionToken, nil
}

func (p *OIDCProvider) parseOIDCClaims(rawClaims map[string]interface{}) (OIDCClaims, error) {
	var claims OIDCClaims

	// Parse email claim
	if emailValue, exists := rawClaims[p.config.EmailClaim]; exists {
		if email, ok := emailValue.(string); ok {
			claims.Email = email
		}
	}

	// Parse name claim
	if nameValue, exists := rawClaims[p.config.NameClaim]; exists {
		if name, ok := nameValue.(string); ok {
			claims.Name = name
		}
	}

	// Parse groups claim
	if groupsValue, exists := rawClaims[p.config.GroupClaim]; exists {
		switch groups := groupsValue.(type) {
		case []interface{}:
			for _, group := range groups {
				if groupStr, ok := group.(string); ok {
					claims.Groups = append(claims.Groups, groupStr)
				}
			}
		case []string:
			claims.Groups = groups
		case string:
			// Single group as string
			claims.Groups = []string{groups}
		}
	}

	// Parse subject claim (always "sub")
	if subValue, exists := rawClaims["sub"]; exists {
		if subject, ok := subValue.(string); ok {
			claims.Subject = subject
		}
	}

	return claims, nil
}

func (p *OIDCProvider) hasAllowedGroup(userGroups, allowedGroups []string) bool {
	if len(allowedGroups) == 0 {
		return true
	}

	allowedGroupsMap := make(map[string]bool)
	for _, group := range allowedGroups {
		allowedGroupsMap[strings.TrimSpace(group)] = true
	}

	for _, userGroup := range userGroups {
		if allowedGroupsMap[userGroup] {
			return true
		}
	}

	return false
}

func (p *OIDCProvider) GetAuthURL(baseURL, state string) string {
	oauth2Config := p.getOAuth2Config(baseURL)
	return oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *OIDCProvider) getOAuth2Config(baseURL string) oauth2.Config {
	// Construct full redirect URL with dedicated callback endpoint
	redirectURL, err := url.JoinPath(baseURL, "/api/v1/users/login/oidc/callback")
	if err != nil {
		log.Err(err).Msg("failed to construct redirect URL")
		return oauth2.Config{}
	}

	return oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     p.endpoint,
		Scopes:       strings.Split(p.config.Scope, " "),
	}
}

// initiateOIDCFlow handles the initial OIDC authentication request by redirecting to the provider
func (p *OIDCProvider) initiateOIDCFlow(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Generate state parameter for CSRF protection
	state, err := generateState()
	if err != nil {
		log.Err(err).Msg("failed to generate OIDC state parameter")
		return services.UserAuthTokenDetail{}, fmt.Errorf("internal server error")
	}

	// Get base URL from request
	baseURL := p.getBaseURL(r)
	u, _ := url.Parse(baseURL)
	domain := u.Hostname()
	if domain == "" {
		domain = noPort(r.Host)
	}

	// Store state in session cookie for validation
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		Expires:  time.Now().Add(p.config.StateExpiry),
		Domain:   domain,
		Secure:   p.isSecure(r),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Generate auth URL and redirect
	authURL := p.GetAuthURL(baseURL, state)
	http.Redirect(w, r, authURL, http.StatusFound)

	// Return empty token since this is a redirect response
	return services.UserAuthTokenDetail{}, nil
}

// handleCallback processes the OAuth2 callback from the OIDC provider
func (p *OIDCProvider) handleCallback(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Check for OAuth error responses first
	if errCode := r.URL.Query().Get("error"); errCode != "" {
		errDesc := r.URL.Query().Get("error_description")
		log.Warn().Str("error", errCode).Str("description", errDesc).Msg("OIDC provider returned error")
		return services.UserAuthTokenDetail{}, fmt.Errorf("OIDC provider error: %s - %s", errCode, errDesc)
	}

	// Verify state parameter
	stateCookie, err := r.Cookie("oidc_state")
	if err != nil {
		log.Warn().Err(err).Msg("OIDC state cookie not found - possible CSRF attack or expired session")
		return services.UserAuthTokenDetail{}, fmt.Errorf("state cookie not found")
	}

	stateParam := r.URL.Query().Get("state")
	if stateParam == "" {
		log.Warn().Msg("OIDC state parameter missing from callback")
		return services.UserAuthTokenDetail{}, fmt.Errorf("state parameter missing")
	}

	if stateParam != stateCookie.Value {
		log.Warn().Str("received", stateParam).Str("expected", stateCookie.Value).Msg("OIDC state mismatch - possible CSRF attack")
		return services.UserAuthTokenDetail{}, fmt.Errorf("state parameter mismatch")
	}

	// Clear state cookie
	baseURL := p.getBaseURL(r)
	u, _ := url.Parse(baseURL)
	domain := u.Hostname()
	if domain == "" {
		domain = noPort(r.Host)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    "",
		Expires:  time.Unix(0, 0),
		Domain:   domain,
		MaxAge:   -1,
		Secure:   p.isSecure(r),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Get base URL from request
	baseURL := p.getBaseURL(r)

	// Use the existing callback logic but return the token instead of redirecting
	return p.AuthenticateWithBaseURL(baseURL, w, r)
}

// Helper functions
func generateState() (string, error) {
	// Generate 32 bytes of cryptographically secure random data
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random state: %w", err)
	}
	// Use URL-safe base64 encoding without padding for clean URLs
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func noPort(host string) string {
	return strings.Split(host, ":")[0]
}

func (p *OIDCProvider) getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	} else if p.options.TrustProxy && r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	host := r.Host
	if p.options.Hostname != "" {
		host = p.options.Hostname
	}

	return scheme + "://" + host
}

func (p *OIDCProvider) isSecure(r *http.Request) bool {
	return p.cookieSecure
}

// InitiateOIDCFlow starts the OIDC authentication flow by redirecting to the provider
func (p *OIDCProvider) InitiateOIDCFlow(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	return p.initiateOIDCFlow(w, r)
}

// HandleCallback processes the OIDC callback and returns the authenticated user token
func (p *OIDCProvider) HandleCallback(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	return p.handleCallback(w, r)
}
