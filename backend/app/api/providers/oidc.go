package providers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
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

type OIDCClaims struct {
	Email         string
	Groups        []string
	Name          string
	Subject       string
	Issuer        string
	EmailVerified *bool
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

	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()

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
	_ = w
	_ = r
	return services.UserAuthTokenDetail{}, fmt.Errorf("OIDC authentication uses dedicated endpoints: /api/v1/users/login/oidc")
}

// AuthenticateWithBaseURL is the main authentication method that requires baseURL
// Called from handleCallback after state, nonce, and PKCE verification
func (p *OIDCProvider) AuthenticateWithBaseURL(baseURL, expectedNonce, pkceVerifier string, _ http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return services.UserAuthTokenDetail{}, fmt.Errorf("missing authorization code")
	}

	// Get OAuth2 config for this request
	oauth2Config := p.getOAuth2Config(baseURL)

	// Exchange code for token with timeout and PKCE verifier
	ctx, cancel := context.WithTimeout(r.Context(), p.config.RequestTimeout)
	defer cancel()

	token, err := oauth2Config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", pkceVerifier))
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

	// Attempt to retrieve UserInfo claims; use them as primary, fallback to ID token claims.
	finalClaims := rawClaims
	userInfoCtx, uiCancel := context.WithTimeout(r.Context(), p.config.RequestTimeout)
	defer uiCancel()

	userInfo, uiErr := p.provider.UserInfo(userInfoCtx, oauth2.StaticTokenSource(token))
	if uiErr != nil {
		log.Debug().Err(uiErr).Msg("OIDC UserInfo fetch failed; falling back to ID token claims")
	} else {
		var uiClaims map[string]interface{}
		if err := userInfo.Claims(&uiClaims); err != nil {
			log.Debug().Err(err).Msg("failed to decode UserInfo claims; falling back to ID token claims")
		} else {
			finalClaims = mergeOIDCClaims(uiClaims, rawClaims) // UserInfo first, then fill gaps from ID token
			log.Debug().Int("userinfo_claims", len(uiClaims)).Int("id_token_claims", len(rawClaims)).Int("merged_claims", len(finalClaims)).Msg("merged UserInfo and ID token claims")
		}
	}

	// Parse claims using configurable claim names (after merge)
	claims, err := p.parseOIDCClaims(finalClaims)
	if err != nil {
		log.Err(err).Msg("failed to parse OIDC claims")
		return services.UserAuthTokenDetail{}, fmt.Errorf("failed to parse OIDC claims: %w", err)
	}

	// Verify nonce claim matches expected value (nonce only from ID token; ensure preserved in merged map)
	tokenNonce, exists := finalClaims["nonce"]
	if !exists {
		log.Warn().Msg("nonce claim missing from ID token - possible replay attack")
		return services.UserAuthTokenDetail{}, fmt.Errorf("nonce claim missing from token")
	}

	tokenNonceStr, ok := tokenNonce.(string)
	if !ok {
		log.Warn().Msg("nonce claim is not a string in ID token")
		return services.UserAuthTokenDetail{}, fmt.Errorf("invalid nonce claim format")
	}

	if tokenNonceStr != expectedNonce {
		log.Warn().Str("received", tokenNonceStr).Str("expected", expectedNonce).Msg("OIDC nonce mismatch - possible replay attack")
		return services.UserAuthTokenDetail{}, fmt.Errorf("nonce parameter mismatch")
	}

	// Check if email is verified
	if p.config.VerifyEmail {
		if claims.EmailVerified == nil {
			return services.UserAuthTokenDetail{}, fmt.Errorf("email verification status not found in token claims")
		}

		if !*claims.EmailVerified {
			return services.UserAuthTokenDetail{}, fmt.Errorf("email not verified")
		}
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
	if claims.Subject == "" {
		return services.UserAuthTokenDetail{}, fmt.Errorf("no subject (sub) claim present")
	}
	if claims.Issuer == "" {
		claims.Issuer = p.config.IssuerURL // fallback to configured issuer, though spec requires 'iss'
	}

	// Use the dedicated OIDC login method (issuer + subject identity)
	sessionToken, err := p.service.LoginOIDC(r.Context(), claims.Issuer, claims.Subject, email, claims.Name)
	if err != nil {
		log.Err(err).Str("email", email).Str("issuer", claims.Issuer).Str("subject", claims.Subject).Msg("OIDC login failed")
		return services.UserAuthTokenDetail{}, fmt.Errorf("OIDC login failed: %w", err)
	}

	return sessionToken, nil
}

func (p *OIDCProvider) parseOIDCClaims(rawClaims map[string]interface{}) (OIDCClaims, error) {
	var claims OIDCClaims

	// Parse email claim
	key := p.config.EmailClaim
	if key == "" {
		key = "email"
	}
	if emailValue, exists := rawClaims[key]; exists {
		if email, ok := emailValue.(string); ok {
			claims.Email = email
		}
	}

	// Parse email_verified claim
	if p.config.VerifyEmail {
		key = p.config.EmailVerifiedClaim
		if key == "" {
			key = "email_verified"
		}
		if emailVerifiedValue, exists := rawClaims[key]; exists {
			switch v := emailVerifiedValue.(type) {
			case bool:
				claims.EmailVerified = &v
			case string:
				if b, err := strconv.ParseBool(v); err == nil {
					claims.EmailVerified = &b
				}
			}
		}
	}

	// Parse name claim
	key = p.config.NameClaim
	if key == "" {
		key = "name"
	}
	if nameValue, exists := rawClaims[key]; exists {
		if name, ok := nameValue.(string); ok {
			claims.Name = name
		}
	}

	// Parse groups claim
	key = p.config.GroupClaim
	if key == "" {
		key = "groups"
	}
	if groupsValue, exists := rawClaims[key]; exists {
		switch groups := groupsValue.(type) {
		case []interface{}:
			claims.Groups = lo.FilterMap(groups, func(group interface{}, _ int) (string, bool) {
				groupStr, ok := group.(string)
				return groupStr, ok
			})
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
	// Parse issuer claim ("iss")
	if issValue, exists := rawClaims["iss"]; exists {
		if iss, ok := issValue.(string); ok {
			claims.Issuer = iss
		}
	}

	return claims, nil
}

func (p *OIDCProvider) hasAllowedGroup(userGroups, allowedGroups []string) bool {
	if len(allowedGroups) == 0 {
		return true
	}

	allowedSet := lo.SliceToMap(allowedGroups, func(group string) (string, bool) {
		return strings.TrimSpace(group), true
	})

	return lo.SomeBy(userGroups, func(userGroup string) bool {
		return allowedSet[userGroup]
	})
}

func (p *OIDCProvider) GetAuthURL(baseURL, state, nonce, pkceVerifier string) string {
	oauth2Config := p.getOAuth2Config(baseURL)
	pkceChallenge := generatePKCEChallenge(pkceVerifier)
	return oauth2Config.AuthCodeURL(state,
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"))
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
		Scopes:       strings.Fields(p.config.Scope),
	}
}

// initiateOIDCFlow handles the initial OIDC authentication request by redirecting to the provider
func (p *OIDCProvider) initiateOIDCFlow(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Generate state parameter for CSRF protection
	state, err := generateSecureToken()
	if err != nil {
		log.Err(err).Msg("failed to generate OIDC state parameter")
		return services.UserAuthTokenDetail{}, fmt.Errorf("internal server error")
	}

	// Generate nonce parameter for replay attack protection
	nonce, err := generateSecureToken()
	if err != nil {
		log.Err(err).Msg("failed to generate OIDC nonce parameter")
		return services.UserAuthTokenDetail{}, fmt.Errorf("internal server error")
	}

	// Generate PKCE verifier for code interception protection
	pkceVerifier, err := generatePKCEVerifier()
	if err != nil {
		log.Err(err).Msg("failed to generate OIDC PKCE verifier")
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

	// Store nonce in session cookie for validation
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_nonce",
		Value:    nonce,
		Expires:  time.Now().Add(p.config.StateExpiry),
		Domain:   domain,
		Secure:   p.isSecure(r),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Store PKCE verifier in session cookie for token exchange
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_pkce_verifier",
		Value:    pkceVerifier,
		Expires:  time.Now().Add(p.config.StateExpiry),
		Domain:   domain,
		Secure:   p.isSecure(r),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Generate auth URL and redirect
	authURL := p.GetAuthURL(baseURL, state, nonce, pkceVerifier)
	http.Redirect(w, r, authURL, http.StatusFound)

	// Return empty token since this is a redirect response
	return services.UserAuthTokenDetail{}, nil
}

// handleCallback processes the OAuth2 callback from the OIDC provider
func (p *OIDCProvider) handleCallback(w http.ResponseWriter, r *http.Request) (services.UserAuthTokenDetail, error) {
	// Helper to clear state cookie using computed domain
	baseURL := p.getBaseURL(r)
	u, _ := url.Parse(baseURL)
	domain := u.Hostname()
	if domain == "" {
		domain = noPort(r.Host)
	}
	clearCookies := func() {
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
		http.SetCookie(w, &http.Cookie{
			Name:     "oidc_nonce",
			Value:    "",
			Expires:  time.Unix(0, 0),
			Domain:   domain,
			MaxAge:   -1,
			Secure:   p.isSecure(r),
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "oidc_pkce_verifier",
			Value:    "",
			Expires:  time.Unix(0, 0),
			Domain:   domain,
			MaxAge:   -1,
			Secure:   p.isSecure(r),
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})
	}

	// Check for OAuth error responses first
	if errCode := r.URL.Query().Get("error"); errCode != "" {
		errDesc := r.URL.Query().Get("error_description")
		log.Warn().Str("error", errCode).Str("description", errDesc).Msg("OIDC provider returned error")
		clearCookies()
		return services.UserAuthTokenDetail{}, fmt.Errorf("OIDC provider error: %s - %s", errCode, errDesc)
	}

	// Verify state parameter
	stateCookie, err := r.Cookie("oidc_state")
	if err != nil {
		log.Warn().Err(err).Msg("OIDC state cookie not found - possible CSRF attack or expired session")
		clearCookies()
		return services.UserAuthTokenDetail{}, fmt.Errorf("state cookie not found")
	}

	stateParam := r.URL.Query().Get("state")
	if stateParam == "" {
		log.Warn().Msg("OIDC state parameter missing from callback")
		clearCookies()
		return services.UserAuthTokenDetail{}, fmt.Errorf("state parameter missing")
	}

	if stateParam != stateCookie.Value {
		log.Warn().Str("received", stateParam).Str("expected", stateCookie.Value).Msg("OIDC state mismatch - possible CSRF attack")
		clearCookies()
		return services.UserAuthTokenDetail{}, fmt.Errorf("state parameter mismatch")
	}

	// Verify nonce parameter
	nonceCookie, err := r.Cookie("oidc_nonce")
	if err != nil {
		log.Warn().Err(err).Msg("OIDC nonce cookie not found - possible replay attack or expired session")
		clearCookies()
		return services.UserAuthTokenDetail{}, fmt.Errorf("nonce cookie not found")
	}

	// Verify PKCE verifier parameter
	pkceCookie, err := r.Cookie("oidc_pkce_verifier")
	if err != nil {
		log.Warn().Err(err).Msg("OIDC PKCE verifier cookie not found - possible code interception attack or expired session")
		clearCookies()
		return services.UserAuthTokenDetail{}, fmt.Errorf("PKCE verifier cookie not found")
	}

	// Clear cookies before proceeding to token verification
	clearCookies()

	// Use the existing callback logic but return the token instead of redirecting
	return p.AuthenticateWithBaseURL(baseURL, nonceCookie.Value, pkceCookie.Value, w, r)
}

// Helper functions
func generateSecureToken() (string, error) {
	// Generate 32 bytes of cryptographically secure random data
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random token: %w", err)
	}
	// Use URL-safe base64 encoding without padding for clean URLs
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// generatePKCEVerifier generates a cryptographically secure code verifier for PKCE
func generatePKCEVerifier() (string, error) {
	// PKCE verifier must be 43-128 characters, we'll use 43 for efficiency
	// 32 bytes = 43 characters when base64url encoded without padding
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// generatePKCEChallenge generates a code challenge from a verifier using S256 method
func generatePKCEChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
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
	} else if p.options.TrustProxy {
		if xfHost := r.Header.Get("X-Forwarded-Host"); xfHost != "" {
			host = xfHost
		}
	}

	return scheme + "://" + host
}

func (p *OIDCProvider) isSecure(r *http.Request) bool {
	_ = r
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

func mergeOIDCClaims(primary, secondary map[string]interface{}) map[string]interface{} {
	// primary has precedence; fill missing/empty values from secondary.
	merged := make(map[string]interface{}, len(primary)+len(secondary))
	for k, v := range primary {
		merged[k] = v
	}
	for k, v := range secondary {
		if existing, ok := merged[k]; !ok || isEmptyClaim(existing) {
			merged[k] = v
		}
	}
	return merged
}

func isEmptyClaim(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []interface{}:
		return len(val) == 0
	case []string:
		return len(val) == 0
	default:
		return false
	}
}
