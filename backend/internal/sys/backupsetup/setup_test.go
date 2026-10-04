package backupsetup

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

const backupsDir = "/backups"

func full() Params {
	return Params{
		BaseURL:              "https://homebox.example.com",
		BehindProxy:          true,
		LocalRoot:            backupsDir,
		AllowCustomEndpoints: true,
		Google:               App{"1234-abc.apps.googleusercontent.com", "GOCSPX-abc_DEF-123"},
		Microsoft:            App{"11111111-2222-3333-4444-555555555555", "abc~DEF.123_x-y"},
		MicrosoftTenant:      "organizations",
		Dropbox:              App{"dbxkey123", "dbxsecret456"},
	}
}

func errs(issues []Issue) []string {
	var out []string
	for _, i := range issues {
		if i.Level == Error {
			out = append(out, i.Message)
		}
	}
	return out
}

func TestGenerateKey(t *testing.T) {
	a, b := GenerateKey(), GenerateKey()
	assert.NotEqual(t, a, b)
	raw, err := base64.StdEncoding.DecodeString(a)
	require.NoError(t, err)
	assert.Len(t, raw, 32)
	assert.True(t, safeValue.MatchString(a), "a generated key is safe in a .env file")
}

func TestValidate(t *testing.T) {
	t.Run("a complete, valid setup has no errors", func(t *testing.T) {
		assert.Empty(t, errs(full().Validate()))
	})
	t.Run("nothing but a key and local root needs no address", func(t *testing.T) {
		assert.Empty(t, errs(Params{LocalRoot: backupsDir, AllowCustomEndpoints: true}.Validate()))
	})

	bad := func(name string, mut func(*Params), want string) {
		t.Run(name, func(t *testing.T) {
			p := full()
			mut(&p)
			got := errs(p.Validate())
			require.NotEmpty(t, got)
			assert.Contains(t, strings.Join(got, "|"), want)
		})
	}
	bad("address required for a cloud drive", func(p *Params) { p.BaseURL = "" }, "address is required")
	bad("address with a path", func(p *Params) { p.BaseURL = "https://h.example/homebox" }, "without a path")
	bad("address with credentials", func(p *Params) { p.BaseURL = "https://u:p@h.example" }, "username or password")
	bad("not a URL", func(p *Params) { p.BaseURL = "homebox.example.com" }, "not a URL")
	bad("https without proxy trust", func(p *Params) { p.BehindProxy = false }, "--behind-proxy")
	bad("plain http for a public host", func(p *Params) { p.BaseURL = "http://homebox.example.com" }, "https")
	bad("id without secret", func(p *Params) { p.Google.ClientSecret = "" }, "without a client secret")
	bad("secret without id", func(p *Params) { p.Dropbox.ClientID = "" }, "without a client ID")
	bad("unsafe characters", func(p *Params) { p.Dropbox.ClientSecret = `a b#c"` }, "not safe in a .env")
	bad("newline smuggling", func(p *Params) { p.Google.ClientSecret = "abc\nHBOX_DEMO=true" }, "not safe in a .env")
	bad("bad tenant", func(p *Params) { p.MicrosoftTenant = "not a tenant!" }, "tenant")
	bad("relative local root", func(p *Params) { p.LocalRoot = "backups" }, "absolute")
	bad("traversal in local root", func(p *Params) { p.LocalRoot = "/data/../etc" }, "'..'")
	bad("filesystem root", func(p *Params) { p.LocalRoot = "/" }, "filesystem root")
	bad("spaces in local root", func(p *Params) { p.LocalRoot = "/my backups" }, "not safe")
	bad("short key", func(p *Params) { p.EncryptionKey = "short" }, "at least 16")

	t.Run("proxy trust configured elsewhere is accepted", func(t *testing.T) {
		p := full()
		p.BehindProxy = false
		p.TrustProxyAlready = true
		assert.Empty(t, errs(p.Validate()))
	})
	t.Run("http is fine for localhost", func(t *testing.T) {
		for _, u := range []string{"http://localhost:3000", "http://127.0.0.1:7745", "http://[::1]:3000"} {
			p := full()
			p.BaseURL = u
			assert.Empty(t, errs(p.Validate()), u)
		}
	})
	t.Run("tenants", func(t *testing.T) {
		for _, tenant := range []string{"common", "consumers", "organizations", "contoso.onmicrosoft.com", "9188040d-6c67-4c5b-b112-36a304b66dad", ""} {
			p := full()
			p.MicrosoftTenant = tenant
			assert.Empty(t, errs(p.Validate()), tenant)
		}
	})
	t.Run("warnings do not block", func(t *testing.T) {
		p := full()
		p.Google.ClientID = "1234-abc"
		p.EncryptionKey = "sixteen-chars-ok."
		issues := p.Validate()
		assert.Empty(t, errs(issues))
		var warnings int
		for _, i := range issues {
			if i.Level == Warning {
				warnings++
			}
		}
		assert.GreaterOrEqual(t, warnings, 3) // client id shape, short key, proxy
	})
}

func TestParseBaseURL(t *testing.T) {
	s, h, err := ParseBaseURL("https://homebox.example.com:8443/")
	require.NoError(t, err)
	assert.Equal(t, "https", s)
	assert.Equal(t, "homebox.example.com:8443", h)
	assert.Equal(t, "https://homebox.example.com/api/v1/group/backup-oauth/callback", RedirectURI("https://homebox.example.com/"))
}

func TestBuildAndKeyHandling(t *testing.T) {
	get := func(p Plan, k string) string {
		for _, kv := range p.Vars {
			if kv.Key == k {
				return kv.Value
			}
		}
		return ""
	}

	t.Run("a key is generated when there is none", func(t *testing.T) {
		plan := Build(full(), nil)
		assert.True(t, plan.KeyGenerated)
		assert.False(t, plan.KeyKept)
		assert.NotEmpty(t, get(plan, EnvEncryptionKey))
	})
	t.Run("an existing key is kept, never silently replaced", func(t *testing.T) {
		plan := Build(full(), map[string]string{EnvEncryptionKey: "existing-key-existing-key-existing"})
		assert.True(t, plan.KeyKept)
		assert.Equal(t, "existing-key-existing-key-existing", get(plan, EnvEncryptionKey))
	})
	t.Run("rotation is explicit and flagged", func(t *testing.T) {
		p := full()
		p.RotateKey = true
		plan := Build(p, map[string]string{EnvEncryptionKey: "existing-key-existing-key-existing"})
		assert.True(t, plan.KeyRotated)
		assert.NotEqual(t, "existing-key-existing-key-existing", get(plan, EnvEncryptionKey))
	})
	t.Run("a supplied key that differs from the existing one counts as a rotation", func(t *testing.T) {
		p := full()
		p.EncryptionKey = "a-brand-new-key-a-brand-new-key-123"
		assert.True(t, Build(p, map[string]string{EnvEncryptionKey: "old-old-old-old-old-old-old-old"}).KeyRotated)
		assert.False(t, Build(p, map[string]string{EnvEncryptionKey: p.EncryptionKey}).KeyRotated)
	})

	t.Run("variables", func(t *testing.T) {
		plan := Build(full(), nil)
		assert.Equal(t, backupsDir, get(plan, EnvLocalRoot))
		assert.Equal(t, "organizations", get(plan, EnvMicrosoftTenant))
		assert.Equal(t, "homebox.example.com", get(plan, EnvHostname), "a bare host: the OIDC code prepends the scheme")
		assert.Equal(t, "true", get(plan, EnvTrustProxy))
		assert.Equal(t, "true", get(plan, EnvCustomEndpoints), "written when allowed, since the server default is off")

		p := full()
		p.AllowCustomEndpoints = false
		p.BehindProxy = false
		p.MicrosoftTenant = "common"
		plan = Build(p, nil)
		assert.Empty(t, get(plan, EnvCustomEndpoints), "left at the server's safe default when not allowed")
		assert.Empty(t, get(plan, EnvTrustProxy), "trust proxy is opt-in")
		assert.Empty(t, get(plan, EnvMicrosoftTenant), "the default tenant is not written")

		onlyKey := Build(Params{AllowCustomEndpoints: true}, nil)
		assert.Empty(t, get(onlyKey, EnvHostname), "no hostname without a cloud drive")
		assert.Empty(t, get(onlyKey, EnvGoogleID))
	})
}

func TestMerge(t *testing.T) {
	existing := "# my config\nHBOX_MODE=production\nHBOX_BACKUP_LOCAL_ROOT=/old\nexport HBOX_OPTIONS_HOSTNAME=old.example.com\n\nHBOX_BACKUP_LOCAL_ROOT=/dup\n# HBOX_BACKUP_GOOGLE_CLIENT_ID=commented\nOTHER=1"
	out := string(Merge([]byte(existing), []KV{
		{EnvLocalRoot, backupsDir},
		{EnvHostname, "homebox.example.com"},
		{EnvGoogleID, "gid"},
	}))
	lines := strings.Split(strings.TrimSpace(out), "\n")

	assert.Contains(t, lines, "# my config", "comments are kept")
	assert.Contains(t, lines, "HBOX_MODE=production", "unrelated variables are untouched")
	assert.Contains(t, lines, "OTHER=1")
	assert.Contains(t, lines, "# HBOX_BACKUP_GOOGLE_CLIENT_ID=commented", "commented-out lines are not treated as set")
	assert.Equal(t, 1, strings.Count(out, EnvLocalRoot+"="), "duplicate definitions collapse to one")
	assert.Contains(t, lines, "HBOX_BACKUP_LOCAL_ROOT=/backups")
	assert.Contains(t, lines, "HBOX_OPTIONS_HOSTNAME=homebox.example.com", "export-prefixed lines are replaced too")
	assert.NotContains(t, out, "old.example.com")
	assert.Contains(t, lines, "HBOX_BACKUP_GOOGLE_CLIENT_ID=gid", "new keys are appended")
	assert.Contains(t, out, sectionHeader)

	// Idempotent: merging the same updates again changes nothing.
	again := string(Merge([]byte(out), []KV{{EnvLocalRoot, backupsDir}, {EnvHostname, "homebox.example.com"}, {EnvGoogleID, "gid"}}))
	assert.Equal(t, out, again)

	// A fresh file is just the header and the variables.
	fresh := string(Merge(nil, []KV{{EnvEncryptionKey, "k"}}))
	assert.Equal(t, sectionHeader+"\n"+EnvEncryptionKey+"=k\n", fresh)
}

func TestParseEnv(t *testing.T) {
	got := ParseEnv([]byte("# c\nA=1\nexport B=\"two\"\nC='three'\n\nbad line\nD = spaced \n"))
	assert.Equal(t, map[string]string{"A": "1", "B": "two", "C": "three", "D": "spaced"}, got)
}

func TestMaskAndDiff(t *testing.T) {
	assert.Equal(t, "abc****", Mask(EnvGoogleSecret, "abcdefghij"))
	assert.Equal(t, "****", Mask(EnvGoogleSecret, "abc"))
	assert.Equal(t, backupsDir, Mask(EnvLocalRoot, backupsDir), "non-secrets are shown")
	assert.Equal(t, "gid-123", Mask(EnvGoogleID, "gid-123"))

	d := Diff(map[string]string{EnvLocalRoot: "/old", EnvGoogleID: "same"}, Plan{Vars: []KV{
		{EnvLocalRoot, backupsDir}, {EnvGoogleID, "same"}, {EnvGoogleSecret, "supersecretvalue"},
	}})
	joined := strings.Join(d, "\n")
	assert.Contains(t, joined, "~ HBOX_BACKUP_LOCAL_ROOT=/backups (was /old)")
	assert.Contains(t, joined, "= HBOX_BACKUP_GOOGLE_CLIENT_ID (unchanged)")
	assert.Contains(t, joined, "+ HBOX_BACKUP_GOOGLE_CLIENT_SECRET=sup****")
	assert.NotContains(t, joined, "supersecretvalue", "secrets never appear in a diff")
}

func TestGuides(t *testing.T) {
	redirect := RedirectURI("https://homebox.example.com")
	for _, p := range []string{"google", "microsoft", "dropbox"} {
		g := ProviderGuide(p, redirect)
		assert.Contains(t, g, redirect, p)
	}
	assert.Contains(t, ProviderGuide("google", redirect), "In production")
	assert.Contains(t, ProviderGuide("microsoft", redirect), "Files.ReadWrite.AppFolder")
	assert.Empty(t, ProviderGuide("nope", redirect))

	c := ComposeSnippet(".env", backupsDir)
	assert.Contains(t, c, "env_file:")
	assert.Contains(t, c, ":/backups")
	assert.NotContains(t, ComposeSnippet(".env", ""), "volumes:")
}

// TestEnvNamesMatchTheRealConfig loads the generator's own output through the
// server's config loader, so a rename of a setting cannot silently break the
// helper.
func TestEnvNamesMatchTheRealConfig(t *testing.T) {
	plan := Build(full(), nil)
	for _, kv := range plan.Vars {
		t.Setenv(kv.Key, kv.Value)
	}
	old := os.Args
	os.Args = []string{"homebox"}
	t.Cleanup(func() { os.Args = old })

	cfg, err := config.New("test", "test")
	require.NoError(t, err)

	assert.Equal(t, backupsDir, cfg.Backup.LocalRoot)
	assert.NotEmpty(t, cfg.Backup.EncryptionKey)
	assert.Equal(t, full().Google.ClientID, cfg.Backup.GoogleClientID)
	assert.Equal(t, full().Google.ClientSecret, cfg.Backup.GoogleClientSecret)
	assert.Equal(t, full().Microsoft.ClientID, cfg.Backup.MicrosoftClientID)
	assert.Equal(t, full().Microsoft.ClientSecret, cfg.Backup.MicrosoftClientSecret)
	assert.Equal(t, "organizations", cfg.Backup.MicrosoftTenant)
	assert.Equal(t, full().Dropbox.ClientID, cfg.Backup.DropboxClientID)
	assert.Equal(t, full().Dropbox.ClientSecret, cfg.Backup.DropboxClientSecret)
	assert.Equal(t, "homebox.example.com", cfg.Options.Hostname)
	assert.True(t, cfg.Options.TrustProxy)
}

func TestCustomEndpointOptInMatchesTheRealConfig(t *testing.T) {
	plan := Build(Params{AllowCustomEndpoints: true}, nil)
	for _, kv := range plan.Vars {
		t.Setenv(kv.Key, kv.Value)
	}
	old := os.Args
	os.Args = []string{"homebox"}
	t.Cleanup(func() { os.Args = old })

	cfg, err := config.New("test", "test")
	require.NoError(t, err)
	assert.True(t, cfg.Backup.AllowCustomEndpoints)
}

func TestCustomEndpointsStayOffUnlessAllowed(t *testing.T) {
	plan := Build(Params{}, nil)
	for _, kv := range plan.Vars {
		t.Setenv(kv.Key, kv.Value)
	}
	old := os.Args
	os.Args = []string{"homebox"}
	t.Cleanup(func() { os.Args = old })

	cfg, err := config.New("test", "test")
	require.NoError(t, err)
	assert.False(t, cfg.Backup.AllowCustomEndpoints, "the server default is off, and the helper does not turn it on")
}

func TestAllowingCustomEndpointsWarns(t *testing.T) {
	var warned bool
	for _, i := range (Params{AllowCustomEndpoints: true}).Validate() {
		if i.Level == Warning && strings.Contains(i.Message, "custom endpoints are on") {
			warned = true
		}
	}
	assert.True(t, warned)
	for _, i := range (Params{}).Validate() {
		assert.NotContains(t, i.Message, "custom endpoints are on")
	}
}
