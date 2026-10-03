package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Cloud-drive destination types. Their OAuth refresh token is stored sealed in
// the destination's secret, and the connected account's name in username.
// OAuth provider keys, as accepted by the start endpoint and listed in options.
const (
	providerGoogle    = "google"
	providerMicrosoft = "microsoft"
	providerDropbox   = "dropbox"
)

const (
	destTypeGDrive   = "gdrive"
	destTypeOneDrive = "onedrive"
	destTypeDropbox  = providerDropbox

	oauthFlowTTL     = 10 * time.Minute
	oauthMaxPending  = 1000
	tokenExpirySlack = 60 * time.Second
)

func isDriveType(t string) bool {
	return t == destTypeGDrive || t == destTypeOneDrive || t == destTypeDropbox
}

// providerEndpoints are the URLs a cloud provider is reached at. Tests point
// them at local fake servers.
type providerEndpoints struct {
	Auth   string
	Token  string
	API    string
	Upload string // Google upload host, Dropbox content host
}

// oauthProvider describes one cloud drive and the OAuth app the operator
// registered for it.
type oauthProvider struct {
	Key          string // providerGoogle, providerMicrosoft, providerDropbox
	DestType     string
	Label        string
	ClientID     string
	ClientSecret string
	Scopes       string
	AuthExtra    url.Values
	EP           providerEndpoints
}

// oauthProviders returns the providers the operator configured. They are only
// usable when credentials can be stored, i.e. an encryption key is set.
func (s *BackupService) oauthProviders() []oauthProvider {
	if s.secrets == nil {
		return nil
	}
	var out []oauthProvider
	for _, key := range []string{providerGoogle, providerMicrosoft, providerDropbox} {
		if p, ok := s.providerByKey(key); ok {
			out = append(out, p)
		}
	}
	return out
}

func (s *BackupService) providerByKey(key string) (oauthProvider, bool) {
	var p oauthProvider
	switch key {
	case providerGoogle:
		p = oauthProvider{
			Key: key, DestType: destTypeGDrive, Label: "Google Drive",
			ClientID: s.cfg.GoogleClientID, ClientSecret: s.cfg.GoogleClientSecret,
			// drive.file limits access to files this app created.
			Scopes:    "https://www.googleapis.com/auth/drive.file",
			AuthExtra: url.Values{"access_type": {"offline"}, "prompt": {"consent"}},
			EP: providerEndpoints{
				Auth:   "https://accounts.google.com/o/oauth2/v2/auth",
				Token:  "https://oauth2.googleapis.com/token",
				API:    "https://www.googleapis.com",
				Upload: "https://www.googleapis.com",
			},
		}
	case providerMicrosoft:
		tenant := s.cfg.MicrosoftTenant
		if tenant == "" {
			tenant = "common"
		}
		base := "https://login.microsoftonline.com/" + url.PathEscape(tenant) + "/oauth2/v2.0"
		p = oauthProvider{
			Key: key, DestType: destTypeOneDrive, Label: "OneDrive",
			ClientID: s.cfg.MicrosoftClientID, ClientSecret: s.cfg.MicrosoftClientSecret,
			// The app folder is the narrowest OneDrive scope.
			Scopes:    "Files.ReadWrite.AppFolder offline_access",
			AuthExtra: url.Values{"response_mode": {"query"}},
			EP:        providerEndpoints{Auth: base + "/authorize", Token: base + "/token", API: "https://graph.microsoft.com/v1.0"},
		}
	case providerDropbox:
		p = oauthProvider{
			Key: key, DestType: destTypeDropbox, Label: "Dropbox",
			ClientID: s.cfg.DropboxClientID, ClientSecret: s.cfg.DropboxClientSecret,
			Scopes:    "files.content.write files.content.read account_info.read",
			AuthExtra: url.Values{"token_access_type": {"offline"}},
			EP: providerEndpoints{
				Auth:   "https://www.dropbox.com/oauth2/authorize",
				Token:  "https://api.dropboxapi.com/oauth2/token",
				API:    "https://api.dropboxapi.com",
				Upload: "https://content.dropboxapi.com",
			},
		}
	default:
		return oauthProvider{}, false
	}
	if ov, ok := s.endpointOverrides[key]; ok {
		p.EP = ov
	}
	if p.ClientID == "" || p.ClientSecret == "" {
		return oauthProvider{}, false
	}
	return p, true
}

func (s *BackupService) providerByType(t string) (oauthProvider, bool) {
	for _, key := range []string{providerGoogle, providerMicrosoft, providerDropbox} {
		if p, ok := s.providerByKey(key); ok && p.DestType == t {
			return p, true
		}
	}
	return oauthProvider{}, false
}

// OAuthProviderKeys lists the configured provider keys, for the UI.
func (s *BackupService) OAuthProviderKeys() []string {
	ps := s.oauthProviders()
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.Key)
	}
	return out
}

// ---------------------------------------------------------------------------
// Pending authorizations

type oauthFlow struct {
	gid, uid uuid.UUID
	provider string
	verifier string
	created  time.Time
}

type oauthResult struct {
	gid, uid     uuid.UUID
	provider     string
	refreshToken string
	account      string
	// accessToken is the token from the code exchange. Testing an unsaved
	// destination uses it, so the refresh token is not spent: providers that
	// rotate refresh tokens (Microsoft) invalidate the old one on first use.
	accessToken  string
	accessExpiry time.Time
	created      time.Time
}

// oauthPending holds in-flight authorizations and the connected accounts
// waiting to be saved. It is in memory and single-use: a restart or a second
// replica simply makes the user connect again.
type oauthPending struct {
	mu      sync.Mutex
	flows   map[string]oauthFlow
	results map[string]oauthResult
}

func (p *oauthPending) sweep(now time.Time) {
	for k, f := range p.flows {
		if now.Sub(f.created) > oauthFlowTTL {
			delete(p.flows, k)
		}
	}
	for k, r := range p.results {
		if now.Sub(r.created) > oauthFlowTTL {
			delete(p.results, k)
		}
	}
}

func (p *oauthPending) putFlow(state string, f oauthFlow) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.flows == nil {
		p.flows = map[string]oauthFlow{}
	}
	p.sweep(f.created)
	if len(p.flows)+len(p.results) >= oauthMaxPending {
		return errors.New("too many pending authorizations; try again in a few minutes")
	}
	p.flows[state] = f
	return nil
}

// takeFlow removes and returns a flow: every state is single-use.
func (p *oauthPending) takeFlow(state string) (oauthFlow, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	f, ok := p.flows[state]
	delete(p.flows, state)
	if !ok || time.Since(f.created) > oauthFlowTTL {
		return oauthFlow{}, false
	}
	return f, true
}

func (p *oauthPending) putResult(ticket string, r oauthResult) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.results == nil {
		p.results = map[string]oauthResult{}
	}
	p.sweep(r.created)
	p.results[ticket] = r
}

func (p *oauthPending) getResult(ticket string, take bool) (oauthResult, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	r, ok := p.results[ticket]
	if ok && take {
		delete(p.results, ticket)
	}
	if !ok || time.Since(r.created) > oauthFlowTTL {
		return oauthResult{}, false
	}
	return r, true
}

func randomToken(n int) string {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// OAuthStart begins connecting an account and returns the provider's
// authorization URL. redirectURI is where the provider sends the user back; it
// must match the one registered with the OAuth app. loginHint, when set, asks
// the provider to preselect that account.
func (s *BackupService) OAuthStart(gid, uid uuid.UUID, providerKey, redirectURI, loginHint string) (string, error) {
	if !s.cfg.Enabled {
		return "", ErrBackupDisabled
	}
	p, ok := s.providerByKey(providerKey)
	if !ok || s.secrets == nil {
		return "", invalid("this cloud provider is not configured on the server")
	}

	verifier := randomToken(32)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	state := randomToken(32)
	if err := s.oauth.putFlow(state, oauthFlow{gid: gid, uid: uid, provider: p.Key, verifier: verifier, created: time.Now()}); err != nil {
		return "", invalid("%v", err)
	}

	q := url.Values{}
	for k, v := range p.AuthExtra {
		q[k] = v
	}
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", p.Scopes)
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	if loginHint != "" && p.Key != providerDropbox {
		q.Set("login_hint", loginHint)
	}
	return p.EP.Auth + "?" + q.Encode(), nil
}

// OAuthOutcome is what the callback page reports to the opener window.
type OAuthOutcome struct {
	OK       bool   `json:"ok"`
	Ticket   string `json:"ticket,omitempty"`
	Account  string `json:"account,omitempty"`
	Provider string `json:"provider,omitempty"`
	Error    string `json:"error,omitempty"`
}

// OAuthCallback finishes an authorization: it validates the single-use state,
// exchanges the code, looks up the account and parks the refresh token under a
// one-time ticket until the destination is saved.
func (s *BackupService) OAuthCallback(ctx context.Context, redirectURI, state, code, providerError string) OAuthOutcome {
	flow, ok := s.oauth.takeFlow(state)
	if !ok {
		return OAuthOutcome{Error: "This authorization link is unknown or has expired. Close this window and try again."}
	}
	p, ok := s.providerByKey(flow.provider)
	if !ok {
		return OAuthOutcome{Error: "This cloud provider is no longer configured."}
	}
	if providerError != "" {
		return OAuthOutcome{Provider: p.Key, Error: "The provider refused access (" + sanitizeProviderText(providerError) + ")."}
	}
	if code == "" {
		return OAuthOutcome{Provider: p.Key, Error: "The provider did not return an authorization code."}
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	tok, err := postToken(ctx, s.httpClient(), p, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {flow.verifier},
	})
	if err != nil {
		log.Warn().Err(err).Str("provider", p.Key).Msg("backup oauth: code exchange failed")
		return OAuthOutcome{Provider: p.Key, Error: "Could not complete sign-in: " + err.Error()}
	}
	if tok.RefreshToken == "" {
		return OAuthOutcome{Provider: p.Key, Error: "The provider did not return a refresh token, so backups could not keep running. Remove Homebox from the account's connected apps and try again."}
	}

	account, err := s.accountName(ctx, p, tok.AccessToken)
	if err != nil {
		log.Warn().Err(err).Str("provider", p.Key).Msg("backup oauth: account lookup failed")
		account = p.Label + " account"
	}

	ticket := randomToken(32)
	s.oauth.putResult(ticket, oauthResult{
		gid: flow.gid, uid: flow.uid, provider: p.Key,
		refreshToken: tok.RefreshToken, account: account, created: time.Now(),
		accessToken: tok.AccessToken, accessExpiry: time.Now().Add(time.Duration(max(tok.ExpiresIn, 60)) * time.Second),
	})
	return OAuthOutcome{OK: true, Ticket: ticket, Account: account, Provider: p.Key}
}

// ticketCredentials resolves a callback ticket for a destination of type typ in
// group gid. When consume is true the ticket cannot be used again.
func (s *BackupService) ticketCredentials(gid uuid.UUID, ticket, typ string, consume bool) (backupSecret, string, error) {
	// Validate before consuming, so a ticket presented by the wrong collection
	// or for the wrong type does not burn the real owner's ticket.
	r, ok := s.oauth.getResult(ticket, false)
	if !ok {
		return backupSecret{}, "", invalid("the account connection expired; connect the account again")
	}
	p, ok := s.providerByType(typ)
	if !ok || p.Key != r.provider || r.gid != gid {
		return backupSecret{}, "", invalid("the connected account does not match this destination type")
	}
	if consume {
		if r, ok = s.oauth.getResult(ticket, true); !ok {
			return backupSecret{}, "", invalid("the account connection expired; connect the account again")
		}
	}
	return backupSecret{RefreshToken: r.refreshToken, accessToken: r.accessToken, accessExpiry: r.accessExpiry}, r.account, nil
}

// ---------------------------------------------------------------------------
// Tokens

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

// errReauthorize means the refresh token no longer works: the user revoked
// access or it expired, and the account must be connected again.
var errReauthorize = errors.New("the cloud account's authorization was revoked or expired; edit the destination and connect the account again")

// sanitizeProviderText trims provider-supplied text to something safe to show.
func sanitizeProviderText(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, s)
	if len(s) > 200 {
		s = s[:200]
	}
	return strings.TrimSpace(s)
}

// postToken calls the provider's token endpoint with the app's credentials.
func postToken(ctx context.Context, hc *http.Client, p oauthProvider, form url.Values) (tokenResponse, error) {
	form.Set("client_id", p.ClientID)
	form.Set("client_secret", p.ClientSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.EP.Token, strings.NewReader(form.Encode()))
	if err != nil {
		return tokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var tr tokenResponse
	_ = json.Unmarshal(body, &tr)
	if resp.StatusCode != http.StatusOK || tr.AccessToken == "" {
		if tr.Error == "invalid_grant" && form.Get("grant_type") == "refresh_token" {
			return tokenResponse{}, errReauthorize
		}
		detail := tr.Error
		if tr.ErrorDesc != "" {
			detail += ": " + tr.ErrorDesc
		}
		if detail == "" {
			detail = http.StatusText(resp.StatusCode)
		}
		return tokenResponse{}, fmt.Errorf("token endpoint returned %d (%s)", resp.StatusCode, sanitizeProviderText(detail))
	}
	return tr, nil
}

// tokenSource hands out access tokens for one destination, refreshing them as
// needed. Providers that rotate refresh tokens (Microsoft) report the new one
// through onRotate so it can be stored.
type tokenSource struct {
	mu       sync.Mutex
	prov     oauthProvider
	hc       *http.Client
	refresh  string
	access   string
	expiry   time.Time
	onRotate func(newRefresh string)
}

func (t *tokenSource) token(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.access != "" && time.Until(t.expiry) > tokenExpirySlack {
		return t.access, nil
	}
	tr, err := postToken(ctx, t.hc, t.prov, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {t.refresh},
	})
	if err != nil {
		return "", err
	}
	t.access = tr.AccessToken
	exp := tr.ExpiresIn
	if exp <= 0 {
		exp = 3600
	}
	t.expiry = time.Now().Add(time.Duration(exp) * time.Second)
	if tr.RefreshToken != "" && tr.RefreshToken != t.refresh {
		t.refresh = tr.RefreshToken
		if t.onRotate != nil {
			t.onRotate(tr.RefreshToken)
		}
	}
	return t.access, nil
}

// invalidate drops the cached access token, forcing a refresh on next use.
func (t *tokenSource) invalidate() {
	t.mu.Lock()
	t.access = ""
	t.mu.Unlock()
}

// ---------------------------------------------------------------------------
// OIDC-assisted offer

// OIDCSuggestion offers a cloud drive that matches the identity provider the
// user signed in with: a Google login suggests Google Drive, a Microsoft login
// OneDrive. It is only an offer. The account is still connected through the
// normal OAuth flow, with the login's email as a hint so the provider
// preselects it.
type OIDCSuggestion struct {
	Provider string `json:"provider"`
	DestType string `json:"destType"`
	Email    string `json:"email"`
}

// oidcProviderKey maps an OIDC issuer URL to the cloud provider that issued the
// login, or "" when the identity provider has no cloud storage (Authentik,
// Keycloak, Authelia and similar).
func oidcProviderKey(issuer string) string {
	issuer = strings.TrimSuffix(strings.TrimSpace(issuer), "/")
	issuer = strings.TrimPrefix(strings.TrimPrefix(issuer, "https://"), "http://")
	host, _, _ := strings.Cut(issuer, "/")
	switch strings.ToLower(host) {
	case "accounts.google.com":
		return providerGoogle
	case "login.microsoftonline.com", "sts.windows.net", "login.windows.net":
		return providerMicrosoft
	}
	return ""
}

// OIDCSuggestionFor returns the offer for a user, or nil when they did not sign
// in through a Google or Microsoft identity provider or the matching backup
// OAuth app is not configured.
func (s *BackupService) OIDCSuggestionFor(issuer *string, email string) *OIDCSuggestion {
	if issuer == nil || email == "" || !s.cfg.Enabled {
		return nil
	}
	key := oidcProviderKey(*issuer)
	if key == "" {
		return nil
	}
	p, ok := s.providerByKey(key)
	if !ok || s.secrets == nil {
		return nil
	}
	return &OIDCSuggestion{Provider: key, DestType: p.DestType, Email: email}
}
