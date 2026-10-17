// Package backupsetup generates and validates the server settings that
// scheduled backups need: the credential encryption key, the local backup
// root, and the OAuth apps for the cloud-drive destinations. It is the logic
// behind `homebox backup-config`, kept free of terminal I/O so it can be tested.
package backupsetup

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"path"
	"regexp"
	"strings"
)

// Environment variable names. They are asserted against the real config
// struct in tests, so a rename there cannot silently break the generator.
const (
	EnvEncryptionKey    = "HBOX_BACKUP_ENCRYPTION_KEY"
	EnvLocalRoot        = "HBOX_BACKUP_LOCAL_ROOT"
	EnvCustomEndpoints  = "HBOX_BACKUP_ALLOW_CUSTOM_ENDPOINTS"
	EnvGoogleID         = "HBOX_BACKUP_GOOGLE_CLIENT_ID"
	EnvGoogleSecret     = "HBOX_BACKUP_GOOGLE_CLIENT_SECRET"
	EnvMicrosoftID      = "HBOX_BACKUP_MICROSOFT_CLIENT_ID"
	EnvMicrosoftSecret  = "HBOX_BACKUP_MICROSOFT_CLIENT_SECRET"
	EnvMicrosoftTenant  = "HBOX_BACKUP_MICROSOFT_TENANT"
	EnvDropboxID        = "HBOX_BACKUP_DROPBOX_CLIENT_ID"
	EnvDropboxSecret    = "HBOX_BACKUP_DROPBOX_CLIENT_SECRET"
	EnvHostname         = "HBOX_OPTIONS_HOSTNAME"
	EnvTrustProxy       = "HBOX_OPTIONS_TRUST_PROXY"
	CallbackPath        = "/api/v1/group/backup-oauth/callback"
	DefaultMicrosoftTen = "common"
)

// Level is how serious a finding is.
type Level int

const (
	Warning Level = iota
	Error
)

// Issue is one validation finding.
type Issue struct {
	Level   Level
	Message string
}

// App is an OAuth app registered at a cloud provider.
type App struct {
	ClientID     string
	ClientSecret string
}

// Set reports whether the app has any credentials entered.
func (a App) Set() bool { return a.ClientID != "" || a.ClientSecret != "" }

// Params are the values the wizard collects.
type Params struct {
	// BaseURL is how users reach Homebox, e.g. https://homebox.example.com. It is
	// needed whenever a cloud drive is configured, to form the redirect URI.
	BaseURL string
	// BehindProxy marks a TLS-terminating reverse proxy in front of Homebox, so
	// the server trusts X-Forwarded-* headers to know it is served over HTTPS.
	BehindProxy bool
	// TrustProxyAlready is set when proxy trust is already configured outside
	// this run (the existing env file or the process environment).
	TrustProxyAlready bool
	// LocalRoot is where local-directory destinations may write, as seen by the
	// Homebox process (inside the container). Empty disables them.
	LocalRoot string
	// AllowCustomEndpoints lets collection owners use SFTP, WebDAV and SMB
	// servers and S3-compatible endpoints they type in. It defaults to false:
	// each is a connection the server makes on the owner's behalf, to wherever
	// the server can reach.
	AllowCustomEndpoints bool
	// EncryptionKey seals stored logins. Empty means "generate one" (or keep the
	// existing one when merging).
	EncryptionKey string
	// RotateKey replaces an existing key instead of keeping it.
	RotateKey bool

	Google          App
	Microsoft       App
	MicrosoftTenant string
	Dropbox         App
}

// Providers lists which cloud providers have credentials entered.
func (p Params) Providers() []string {
	var out []string
	if p.Google.Set() {
		out = append(out, "google")
	}
	if p.Microsoft.Set() {
		out = append(out, "microsoft")
	}
	if p.Dropbox.Set() {
		out = append(out, "dropbox")
	}
	return out
}

var (
	// Values are written unquoted to a .env file, which compose, systemd and
	// shells all read slightly differently, so accept only characters that
	// mean the same everywhere. Real client IDs and secrets fit.
	safeValue = regexp.MustCompile(`^[A-Za-z0-9._~+/=:@-]+$`)
	guidRE    = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	domainRE  = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
)

// GenerateKey returns a new random encryption key: 32 bytes, base64 encoded.
func GenerateKey() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.StdEncoding.EncodeToString(b)
}

// ParseBaseURL splits a base URL into scheme and host (host:port), rejecting
// anything that is not a plain origin.
func ParseBaseURL(raw string) (scheme, host string, err error) {
	raw = strings.TrimSpace(raw)
	u, perr := url.Parse(raw)
	switch {
	case perr != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https"):
		return "", "", fmt.Errorf("%q is not a URL like https://homebox.example.com", raw)
	case u.User != nil:
		return "", "", fmt.Errorf("the URL must not contain a username or password")
	case (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "":
		return "", "", fmt.Errorf("give only the address, without a path (got %q)", raw)
	}
	return u.Scheme, u.Host, nil
}

func isLoopback(host string) bool {
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		h = host
	}
	h = strings.Trim(h, "[]")
	if h == "localhost" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}

// RedirectURI is the address to register with each OAuth app.
func RedirectURI(baseURL string) string {
	return strings.TrimSuffix(strings.TrimSpace(baseURL), "/") + CallbackPath
}

// Validate checks p and returns every problem it finds.
func (p Params) Validate() []Issue {
	var out []Issue
	out = append(out, p.validateAddress()...)
	out = append(out, p.validateApps()...)
	out = append(out, p.validateLocalRoot()...)
	out = append(out, p.validateKey()...)
	if p.AllowCustomEndpoints {
		out = append(out, warnf("custom endpoints are on: collection owners can make the server connect to any address it can reach (SFTP, WebDAV and SMB hosts, S3-compatible endpoints). Turn this off unless you trust every collection owner or the server's network is restricted"))
	}
	if p.BehindProxy {
		out = append(out, warnf("behind a proxy, Homebox trusts X-Forwarded-Host and X-Forwarded-Proto: make sure the proxy sets them and strips any client-supplied copies"))
	}
	return out
}

func errf(f string, a ...any) Issue  { return Issue{Error, fmt.Sprintf(f, a...)} }
func warnf(f string, a ...any) Issue { return Issue{Warning, fmt.Sprintf(f, a...)} }

// validateAddress checks the public address, which only matters once a cloud
// drive is configured because it forms the OAuth redirect URI.
func (p Params) validateAddress() []Issue {
	if len(p.Providers()) == 0 {
		return nil
	}
	if strings.TrimSpace(p.BaseURL) == "" {
		return []Issue{errf("the Homebox address is required to set up a cloud drive (it forms the redirect URI)")}
	}
	scheme, host, err := ParseBaseURL(p.BaseURL)
	switch {
	case err != nil:
		return []Issue{errf("Homebox address: %v", err)}
	case scheme == "http" && !isLoopback(host):
		return []Issue{errf("Google, Microsoft and Dropbox only accept an https redirect URI (plain http is allowed for localhost only)")}
	case scheme == "https" && !p.BehindProxy && !p.TrustProxyAlready:
		// Homebox serves plain HTTP, so an https address means a reverse proxy
		// terminates TLS in front of it. Without trusting that proxy's
		// X-Forwarded-Proto, Homebox builds an http:// redirect URI and the
		// provider rejects the sign-in.
		return []Issue{errf("an https address means a reverse proxy terminates TLS in front of Homebox, which must then trust its X-Forwarded-* headers or the redirect URI is built as http:// and rejected: pass --behind-proxy")}
	}
	return nil
}

// validateApps checks the OAuth app credentials of each provider.
func (p Params) validateApps() []Issue {
	var out []Issue
	check := func(name, val string) {
		if val != "" && !safeValue.MatchString(val) {
			out = append(out, errf("%s contains characters that are not safe in a .env file (allowed: letters, digits and . _ ~ + / = : @ -)", name))
		}
	}
	pair := func(name string, a App) {
		switch {
		case a.ClientID == "" && a.ClientSecret != "":
			out = append(out, errf("%s: a client secret was given without a client ID", name))
		case a.ClientID != "" && a.ClientSecret == "":
			out = append(out, errf("%s: a client ID was given without a client secret", name))
		}
		check(name+" client ID", a.ClientID)
		check(name+" client secret", a.ClientSecret)
	}
	pair("Google", p.Google)
	pair("Microsoft", p.Microsoft)
	pair("Dropbox", p.Dropbox)

	if p.Google.ClientID != "" && !strings.HasSuffix(p.Google.ClientID, ".apps.googleusercontent.com") {
		out = append(out, warnf("Google client IDs normally end in .apps.googleusercontent.com; double-check you copied the OAuth client ID"))
	}
	if t := p.MicrosoftTenant; t != "" && t != DefaultMicrosoftTen && p.Microsoft.Set() {
		if t != "consumers" && t != "organizations" && !guidRE.MatchString(t) && !domainRE.MatchString(t) {
			out = append(out, errf("Microsoft tenant must be common, consumers, organizations, a tenant ID or a domain (got %q)", t))
		}
	}
	return out
}

func (p Params) validateLocalRoot() []Issue {
	switch {
	case p.LocalRoot == "":
		return nil
	case !strings.HasPrefix(p.LocalRoot, "/"):
		return []Issue{errf("the local backup root must be an absolute path inside the container, such as /backups")}
	case strings.Contains(p.LocalRoot, ".."):
		return []Issue{errf("the local backup root must not contain '..'")}
	case path.Clean(p.LocalRoot) == "/":
		return []Issue{errf("the local backup root must not be the filesystem root")}
	case strings.ContainsAny(p.LocalRoot, " \t#\"'$\\"):
		return []Issue{errf("the local backup root contains characters that are not safe in a .env file")}
	}
	return nil
}

func (p Params) validateKey() []Issue {
	k := p.EncryptionKey
	if k == "" {
		return nil
	}
	var out []Issue
	switch {
	case len(k) < 16:
		out = append(out, errf("the encryption key must be at least 16 characters (32 or more is recommended)"))
	case len(k) < 32:
		out = append(out, warnf("the encryption key is shorter than 32 characters; a generated key is stronger"))
	}
	if !safeValue.MatchString(k) {
		out = append(out, errf("the encryption key contains characters that are not safe in a .env file (allowed: letters, digits and . _ ~ + / = : @ -)"))
	}
	return out
}

// HasErrors reports whether any issue is an error.
func HasErrors(issues []Issue) bool {
	for _, i := range issues {
		if i.Level == Error {
			return true
		}
	}
	return false
}
