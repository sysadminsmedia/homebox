package services

import (
	"context"
	"crypto/rand"
	"io"
	"io/fs"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

const driveRedirect = "https://homebox.example/api/v1/group/backup-oauth/callback"

// newDriveSvc returns services with all three cloud providers configured and
// pointed at the fake.
func newDriveSvc(t *testing.T, cloud *fakeCloud) *AllServices {
	t.Helper()
	svc := New(tRepos,
		WithExportPlumbing(tbus, tClient, config.Storage{PrefixPath: "/", ConnString: "file://" + os.TempDir()}, "mem://{{ .Topic }}", "sqlite3"),
		WithBackupConfig(config.BackupConf{
			Enabled: true, AllowCustomEndpoints: true, EncryptionKey: "drive-test-key",
			GoogleClientID: cloud.clientID, GoogleClientSecret: cloud.clientSecret,
			MicrosoftClientID: cloud.clientID, MicrosoftClientSecret: cloud.clientSecret,
			DropboxClientID: cloud.clientID, DropboxClientSecret: cloud.clientSecret,
		}),
	)
	svc.Backups.endpointOverrides = cloud.endpoints()
	return svc
}

func driveSettings(typ string) repo.BackupSettings {
	return repo.BackupSettings{
		Name: "drive " + typ, Type: typ, Prefix: "homebox-backups", Enabled: true, Frequency: freqDaily,
		IntervalHours: 1, AtHour: 3, DayOfMonth: 1, SkipIfUnchanged: true, KeepDaily: 7, KeepWeekly: 4, KeepMonthly: 6,
		HealthIntervalMinutes: 15, AlertFailureThreshold: 2,
	}
}

// connectAccount runs the whole authorization dance against the fake and
// returns the outcome of the callback.
func connectAccount(t *testing.T, svc *AllServices, cloud *fakeCloud, gid, uid uuid.UUID, provider string) OAuthOutcome {
	t.Helper()
	authURL, err := svc.Backups.OAuthStart(gid, uid, provider, driveRedirect, "")
	require.NoError(t, err)
	u, err := url.Parse(authURL)
	require.NoError(t, err)
	q := u.Query()
	assert.Equal(t, cloud.clientID, q.Get("client_id"))
	assert.Equal(t, driveRedirect, q.Get("redirect_uri"))
	assert.Equal(t, "code", q.Get("response_type"))
	assert.Equal(t, "S256", q.Get("code_challenge_method"))
	assert.NotEmpty(t, q.Get("code_challenge"))
	assert.GreaterOrEqual(t, len(q.Get("state")), 40, "state is unguessable")
	code := cloud.authorize(q.Get("code_challenge"))
	return svc.Backups.OAuthCallback(context.Background(), driveRedirect, q.Get("state"), code, "")
}

var driveCases = []struct {
	provider, typ, account, prefix string
}{
	{providerGoogle, destTypeGDrive, "me@example.com", ""},
	{providerMicrosoft, destTypeOneDrive, "me@contoso.example", ""},
	{providerDropbox, destTypeDropbox, "me@dropbox.example", ""},
}

func TestOAuthFlowSecurity(t *testing.T) {
	ctx := context.Background()
	cloud := newFakeCloud(t)
	svc := newDriveSvc(t, cloud)
	gidA, uid := uuid.New(), uuid.New()
	grpA, err := tRepos.Groups.GroupCreate(ctx, "oauth-a-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)
	grpB, err := tRepos.Groups.GroupCreate(ctx, "oauth-b-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)
	_ = gidA

	t.Run("happy path reports the account and a one-time ticket", func(t *testing.T) {
		out := connectAccount(t, svc, cloud, grpA.ID, uid, providerGoogle)
		require.True(t, out.OK, out.Error)
		assert.Equal(t, "me@example.com", out.Account)
		assert.NotEmpty(t, out.Ticket)
	})

	t.Run("state is single use", func(t *testing.T) {
		authURL, err := svc.Backups.OAuthStart(grpA.ID, uid, providerGoogle, driveRedirect, "")
		require.NoError(t, err)
		u, _ := url.Parse(authURL)
		state, ch := u.Query().Get("state"), u.Query().Get("code_challenge")
		first := svc.Backups.OAuthCallback(ctx, driveRedirect, state, cloud.authorize(ch), "")
		require.True(t, first.OK)
		replay := svc.Backups.OAuthCallback(ctx, driveRedirect, state, cloud.authorize(ch), "")
		assert.False(t, replay.OK, "a replayed state is refused")
	})

	t.Run("unknown state, provider errors and missing codes are refused", func(t *testing.T) {
		assert.False(t, svc.Backups.OAuthCallback(ctx, driveRedirect, "nope", "code", "").OK)

		authURL, _ := svc.Backups.OAuthStart(grpA.ID, uid, providerDropbox, driveRedirect, "")
		u, _ := url.Parse(authURL)
		denied := svc.Backups.OAuthCallback(ctx, driveRedirect, u.Query().Get("state"), "", "access_denied")
		assert.False(t, denied.OK)
		assert.Contains(t, denied.Error, "access_denied")

		authURL, _ = svc.Backups.OAuthStart(grpA.ID, uid, providerDropbox, driveRedirect, "")
		u, _ = url.Parse(authURL)
		assert.False(t, svc.Backups.OAuthCallback(ctx, driveRedirect, u.Query().Get("state"), "", "").OK)
	})

	t.Run("a code that fails PKCE or was not issued is refused", func(t *testing.T) {
		authURL, _ := svc.Backups.OAuthStart(grpA.ID, uid, providerGoogle, driveRedirect, "")
		u, _ := url.Parse(authURL)
		out := svc.Backups.OAuthCallback(ctx, driveRedirect, u.Query().Get("state"), cloud.authorize("a-different-challenge"), "")
		assert.False(t, out.OK)
		assert.Contains(t, out.Error, "PKCE")

		authURL, _ = svc.Backups.OAuthStart(grpA.ID, uid, providerGoogle, driveRedirect, "")
		u, _ = url.Parse(authURL)
		assert.False(t, svc.Backups.OAuthCallback(ctx, driveRedirect, u.Query().Get("state"), "forged-code", "").OK)
	})

	t.Run("an expired flow is refused", func(t *testing.T) {
		authURL, _ := svc.Backups.OAuthStart(grpA.ID, uid, providerGoogle, driveRedirect, "")
		u, _ := url.Parse(authURL)
		state := u.Query().Get("state")
		svc.Backups.oauth.mu.Lock()
		f := svc.Backups.oauth.flows[state]
		f.created = time.Now().Add(-2 * oauthFlowTTL)
		svc.Backups.oauth.flows[state] = f
		svc.Backups.oauth.mu.Unlock()
		out := svc.Backups.OAuthCallback(ctx, driveRedirect, state, cloud.authorize(u.Query().Get("code_challenge")), "")
		assert.False(t, out.OK)
	})

	t.Run("a ticket is bound to its collection and provider and used once", func(t *testing.T) {
		out := connectAccount(t, svc, cloud, grpA.ID, uid, providerGoogle)
		require.True(t, out.OK)

		in := repo.BackupInput{BackupSettings: driveSettings(destTypeGDrive), OAuthTicket: out.Ticket}
		_, err := svc.Backups.CreateDestination(ctx, grpB.ID, in)
		require.ErrorIs(t, err, ErrBackupInvalid, "another collection cannot use the ticket")

		wrong := in
		wrong.Type = destTypeDropbox
		_, err = svc.Backups.CreateDestination(ctx, grpA.ID, wrong)
		require.ErrorIs(t, err, ErrBackupInvalid, "a Google ticket cannot create a Dropbox destination")

		d, err := svc.Backups.CreateDestination(ctx, grpA.ID, in)
		require.NoError(t, err)
		assert.Equal(t, "me@example.com", d.Username)

		_, err = svc.Backups.CreateDestination(ctx, grpA.ID, in)
		require.ErrorIs(t, err, ErrBackupInvalid, "the ticket is consumed")
	})

	t.Run("saving needs a connected account", func(t *testing.T) {
		_, err := svc.Backups.CreateDestination(ctx, grpA.ID, repo.BackupInput{BackupSettings: driveSettings(destTypeGDrive)})
		require.ErrorIs(t, err, ErrBackupInvalid)
		_, err = svc.Backups.CreateDestination(ctx, grpA.ID, repo.BackupInput{BackupSettings: driveSettings(destTypeGDrive), OAuthTicket: "made-up"})
		require.ErrorIs(t, err, ErrBackupInvalid)
	})

	t.Run("unconfigured providers and a missing key are refused", func(t *testing.T) {
		bare := &BackupService{cfg: config.BackupConf{Enabled: true, AllowCustomEndpoints: true}}
		_, err := bare.OAuthStart(grpA.ID, uid, providerGoogle, driveRedirect, "")
		require.ErrorIs(t, err, ErrBackupInvalid)

		noKey := New(tRepos, WithBackupConfig(config.BackupConf{Enabled: true, GoogleClientID: "x", GoogleClientSecret: "y"}))
		assert.Empty(t, noKey.Backups.OAuthProviderKeys(), "cloud drives need an encryption key")
		assert.False(t, noKey.Backups.Options().RemoteEnabled)

		onlyGoogle := New(tRepos, WithBackupConfig(config.BackupConf{Enabled: true, EncryptionKey: "k", GoogleClientID: "x", GoogleClientSecret: "y"}))
		assert.Equal(t, []string{providerGoogle}, onlyGoogle.Backups.OAuthProviderKeys())
		_, err = onlyGoogle.Backups.NormalizeSettings(driveSettings(destTypeDropbox))
		require.ErrorIs(t, err, ErrBackupInvalid)

		_, err = svc.Backups.OAuthStart(grpA.ID, uid, "unknown", driveRedirect, "")
		require.ErrorIs(t, err, ErrBackupInvalid)
	})
}

func TestCloudDriveEndToEnd(t *testing.T) {
	for _, tc := range driveCases {
		t.Run(tc.provider, func(t *testing.T) {
			ctx := context.Background()
			cloud := newFakeCloud(t)
			svc := newDriveSvc(t, cloud)
			uid := uuid.New()
			grp, err := tRepos.Groups.GroupCreate(ctx, "drive-"+tc.provider+"-"+fk.Str(4), uuid.Nil)
			require.NoError(t, err)

			// Testing before saving works through the ticket without consuming it.
			out := connectAccount(t, svc, cloud, grp.ID, uid, tc.provider)
			require.True(t, out.OK, out.Error)
			in := repo.BackupInput{BackupSettings: driveSettings(tc.typ), OAuthTicket: out.Ticket}
			res, err := svc.Backups.TestSettings(ctx, grp.ID, in)
			require.NoError(t, err)
			require.True(t, res.OK, res.Message)

			dest, err := svc.Backups.CreateDestination(ctx, grp.ID, in)
			require.NoError(t, err)
			assert.Equal(t, tc.account, dest.Username)
			assert.True(t, dest.HasSecret)

			// The refresh token is sealed at rest and never in the API shape.
			row, err := tRepos.BackupDestinations.Get(ctx, grp.ID, dest.ID)
			require.NoError(t, err)
			plain, err := svc.Backups.secrets.open(secretAAD(grp.ID, dest.ID), row.Secret)
			require.NoError(t, err)
			require.NotEmpty(t, plain.RefreshToken)
			assert.NotContains(t, row.Secret, plain.RefreshToken)

			// A backup lands in the provider and reads back byte for byte.
			exp, err := tRepos.Exports.CreateForDestination(ctx, grp.ID, dest.ID, "scheduled")
			require.NoError(t, err)
			svc.Exports.RunExport(ctx, exp.ID, grp.ID)
			exp, err = tRepos.Exports.Get(ctx, grp.ID, exp.ID)
			require.NoError(t, err)
			require.Equal(t, "completed", exp.Status, exp.Error)

			stored, ok := cloud.stored(tc.provider, exp.ArtifactPath)
			require.True(t, ok, "artifact %s is stored at the provider", exp.ArtifactPath)
			assert.Equal(t, exp.SizeBytes, int64(len(stored)))
			assert.Equal(t, []byte("PK"), stored[:2])

			store, key, err := svc.Backups.ArtifactLocation(ctx, grp.ID, exp)
			require.NoError(t, err)
			rc, err := store.Open(ctx, key)
			require.NoError(t, err)
			got, err := io.ReadAll(rc)
			require.NoError(t, err)
			_ = rc.Close()
			_ = store.Close()
			assert.Equal(t, stored, got)

			// An expired access token is refreshed transparently.
			cloud.expireAccessTokens()
			res2, err := svc.Backups.TestDestination(ctx, grp.ID, dest.ID)
			require.NoError(t, err)
			assert.True(t, res2.OK, res2.Message)

			// Unchanged edits keep the account without a new ticket. Read the row
			// again: a provider that rotates refresh tokens has updated it.
			row, err = tRepos.BackupDestinations.Get(ctx, grp.ID, dest.ID)
			require.NoError(t, err)
			edit := driveSettings(tc.typ)
			edit.Name = "renamed"
			upd, err := svc.Backups.UpdateDestination(ctx, grp.ID, dest.ID, repo.BackupInput{BackupSettings: edit})
			require.NoError(t, err)
			assert.Equal(t, tc.account, upd.Username)
			assert.Equal(t, row.Secret, upd.Secret)

			// Deleting a version removes it from the provider; repeating is fine.
			require.NoError(t, svc.Backups.DeleteVersion(ctx, grp.ID, exp))
			_, ok = cloud.stored(tc.provider, exp.ArtifactPath)
			assert.False(t, ok)
			require.NoError(t, svc.Backups.DeleteVersion(ctx, grp.ID, exp))

			// Revoked access surfaces as a clear failure, and reconnecting fixes it.
			cloud.expireAccessTokens()
			cloud.revokeRefreshTokens()
			svc.Backups.tokens = nil
			res3, err := svc.Backups.TestDestination(ctx, grp.ID, dest.ID)
			require.NoError(t, err)
			assert.False(t, res3.OK)
			assert.Contains(t, res3.Message, "connect the account again")

			out2 := connectAccount(t, svc, cloud, grp.ID, uid, tc.provider)
			require.True(t, out2.OK)
			_, err = svc.Backups.UpdateDestination(ctx, grp.ID, dest.ID, repo.BackupInput{BackupSettings: edit, OAuthTicket: out2.Ticket})
			require.NoError(t, err)
			res4, err := svc.Backups.TestDestination(ctx, grp.ID, dest.ID)
			require.NoError(t, err)
			assert.True(t, res4.OK, res4.Message)

			// A type change needs a fresh connection.
			other := driveSettings(destTypeDropbox)
			if tc.typ == destTypeDropbox {
				other = driveSettings(destTypeGDrive)
			}
			_, err = svc.Backups.UpdateDestination(ctx, grp.ID, dest.ID, repo.BackupInput{BackupSettings: other})
			require.ErrorIs(t, err, ErrBackupInvalid)
		})
	}
}

func TestMicrosoftRefreshTokenRotationIsPersisted(t *testing.T) {
	ctx := context.Background()
	cloud := newFakeCloud(t) // microsoft rotates refresh tokens
	svc := newDriveSvc(t, cloud)
	uid := uuid.New()
	grp, err := tRepos.Groups.GroupCreate(ctx, "rotate-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	out := connectAccount(t, svc, cloud, grp.ID, uid, providerMicrosoft)
	require.True(t, out.OK)
	dest, err := svc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: driveSettings(destTypeOneDrive), OAuthTicket: out.Ticket})
	require.NoError(t, err)
	before, err := svc.Backups.secrets.open(secretAAD(grp.ID, dest.ID), dest.Secret)
	require.NoError(t, err)

	res, err := svc.Backups.TestDestination(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	require.True(t, res.OK, res.Message)

	row, err := tRepos.BackupDestinations.Get(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	after, err := svc.Backups.secrets.open(secretAAD(grp.ID, dest.ID), row.Secret)
	require.NoError(t, err)
	assert.NotEqual(t, before.RefreshToken, after.RefreshToken, "the rotated token was stored")

	// A fresh process (no cache) still works with only what is in the database.
	svc.Backups.tokens = nil
	cloud.expireAccessTokens()
	res, err = svc.Backups.TestDestination(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	assert.True(t, res.OK, res.Message)
}

func TestDriveLargeUploadsAreChunked(t *testing.T) {
	oldOne, oldOneChunk, oldDB, oldDBChunk := oneDriveSimpleMax, oneDriveChunk, dropboxSimpleMax, dropboxChunk
	t.Cleanup(func() {
		oneDriveSimpleMax, oneDriveChunk, dropboxSimpleMax, dropboxChunk = oldOne, oldOneChunk, oldDB, oldDBChunk
	})
	oneDriveSimpleMax, oneDriveChunk = 100*1024, 2*320*1024
	dropboxSimpleMax, dropboxChunk = 100*1024, 256*1024

	data := make([]byte, 1500*1024+123) // several chunks and an odd tail
	_, err := rand.Read(data)
	require.NoError(t, err)

	for _, tc := range driveCases {
		t.Run(tc.provider, func(t *testing.T) {
			ctx := context.Background()
			cloud := newFakeCloud(t)
			svc := newDriveSvc(t, cloud)
			grp, err := tRepos.Groups.GroupCreate(ctx, "big-"+tc.provider+"-"+fk.Str(4), uuid.Nil)
			require.NoError(t, err)
			out := connectAccount(t, svc, cloud, grp.ID, uuid.New(), tc.provider)
			require.True(t, out.OK)
			dest, err := svc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: driveSettings(tc.typ), OAuthTicket: out.Ticket})
			require.NoError(t, err)

			st, _, err := svc.Backups.openStore(ctx, dest, nil)
			require.NoError(t, err)
			defer func() { _ = st.Close() }()

			key := "homebox-backups/" + grp.ID.String() + "/backups/big.zip"
			f, err := os.CreateTemp(t.TempDir(), "big")
			require.NoError(t, err)
			_, err = f.Write(data)
			require.NoError(t, err)
			_, err = f.Seek(0, io.SeekStart)
			require.NoError(t, err)
			require.NoError(t, st.Write(ctx, key, f, int64(len(data)), "application/zip"))
			_ = f.Close()

			stored, ok := cloud.stored(tc.provider, key)
			require.True(t, ok)
			assert.Equal(t, data, stored)

			// A large upload survives an expired token mid-session start.
			cloud.expireAccessTokens()
			key2 := "homebox-backups/" + grp.ID.String() + "/backups/big2.zip"
			require.NoError(t, st.Write(ctx, key2, bytesReader(data), int64(len(data)), "application/zip"))
			stored2, ok := cloud.stored(tc.provider, key2)
			require.True(t, ok)
			assert.Equal(t, data, stored2)
		})
	}
}

func TestDriveStoreMissingObjects(t *testing.T) {
	for _, tc := range driveCases {
		t.Run(tc.provider, func(t *testing.T) {
			ctx := context.Background()
			cloud := newFakeCloud(t)
			svc := newDriveSvc(t, cloud)
			grp, err := tRepos.Groups.GroupCreate(ctx, "miss-"+tc.provider+"-"+fk.Str(4), uuid.Nil)
			require.NoError(t, err)
			out := connectAccount(t, svc, cloud, grp.ID, uuid.New(), tc.provider)
			require.True(t, out.OK)
			dest, err := svc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: driveSettings(tc.typ), OAuthTicket: out.Ticket})
			require.NoError(t, err)
			st, _, err := svc.Backups.openStore(ctx, dest, nil)
			require.NoError(t, err)
			defer func() { _ = st.Close() }()

			_, err = st.Open(ctx, "homebox-backups/nope/missing.zip")
			require.ErrorIs(t, err, fs.ErrNotExist)
			require.NoError(t, st.Delete(ctx, "homebox-backups/nope/missing.zip"), "deleting a missing file is not an error")
		})
	}
}

func TestDropboxArgIsASCII(t *testing.T) {
	arg, err := dropboxArg(map[string]any{"path": "/sauvegardes/été/日本語/😀.zip"})
	require.NoError(t, err)
	for _, r := range arg {
		require.Less(t, r, rune(0x80))
	}
	var back map[string]string
	require.NoError(t, jsonUnmarshal(arg, &back))
	assert.Equal(t, "/sauvegardes/été/日本語/😀.zip", back["path"])
}

func TestDriveQueryEscaping(t *testing.T) {
	assert.Equal(t, `it\'s a \\ name`, driveQueryEscaper.Replace(`it's a \ name`))
}

func TestOIDCProviderKey(t *testing.T) {
	cases := map[string]string{
		"https://accounts.google.com":                                providerGoogle,
		"https://accounts.google.com/":                               providerGoogle,
		"accounts.google.com":                                        providerGoogle,
		"https://login.microsoftonline.com/9188040d-6c67/v2.0":       providerMicrosoft,
		"https://login.microsoftonline.com/common/v2.0":              providerMicrosoft,
		"https://sts.windows.net/9188040d-6c67/":                     providerMicrosoft,
		"https://auth.example.com/application/o/homebox/":            "",
		"https://keycloak.lan/realms/home":                           "",
		"https://accounts.google.com.evil.example":                   "",
		"https://evil.example/accounts.google.com":                   "",
		"https://login.microsoftonline.com.evil.example/common/v2.0": "",
		"": "",
	}
	for issuer, want := range cases {
		assert.Equal(t, want, oidcProviderKey(issuer), issuer)
	}
}

func TestOIDCSuggestion(t *testing.T) {
	cloud := newFakeCloud(t)
	svc := newDriveSvc(t, cloud)
	str := func(s string) *string { return &s }

	t.Run("a Google login suggests Google Drive", func(t *testing.T) {
		got := svc.Backups.OIDCSuggestionFor(str("https://accounts.google.com"), "me@example.com")
		require.NotNil(t, got)
		assert.Equal(t, OIDCSuggestion{Provider: providerGoogle, DestType: destTypeGDrive, Email: "me@example.com"}, *got)
	})
	t.Run("a Microsoft login suggests OneDrive", func(t *testing.T) {
		got := svc.Backups.OIDCSuggestionFor(str("https://login.microsoftonline.com/tid/v2.0"), "me@contoso.example")
		require.NotNil(t, got)
		assert.Equal(t, destTypeOneDrive, got.DestType)
	})
	t.Run("no offer without an OIDC login, an email, or a storage-less provider", func(t *testing.T) {
		assert.Nil(t, svc.Backups.OIDCSuggestionFor(nil, "me@example.com"), "password login")
		assert.Nil(t, svc.Backups.OIDCSuggestionFor(str("https://accounts.google.com"), ""))
		assert.Nil(t, svc.Backups.OIDCSuggestionFor(str("https://auth.example.com/application/o/homebox/"), "me@example.com"), "Authentik has no cloud storage")
	})
	t.Run("no offer when the matching backup app is not configured or backups are off", func(t *testing.T) {
		onlyDropbox := New(tRepos, WithBackupConfig(config.BackupConf{
			Enabled: true, EncryptionKey: "k", DropboxClientID: "x", DropboxClientSecret: "y",
		}))
		assert.Nil(t, onlyDropbox.Backups.OIDCSuggestionFor(str("https://accounts.google.com"), "me@example.com"))

		noKey := New(tRepos, WithBackupConfig(config.BackupConf{Enabled: true, GoogleClientID: "x", GoogleClientSecret: "y"}))
		assert.Nil(t, noKey.Backups.OIDCSuggestionFor(str("https://accounts.google.com"), "me@example.com"))

		off := New(tRepos, WithBackupConfig(config.BackupConf{EncryptionKey: "k", GoogleClientID: "x", GoogleClientSecret: "y"}))
		assert.Nil(t, off.Backups.OIDCSuggestionFor(str("https://accounts.google.com"), "me@example.com"))
	})
}

func TestOAuthStartLoginHint(t *testing.T) {
	cloud := newFakeCloud(t)
	svc := newDriveSvc(t, cloud)
	hint := func(provider, h string) string {
		u, err := svc.Backups.OAuthStart(uuid.New(), uuid.New(), provider, driveRedirect, h)
		require.NoError(t, err)
		parsed, err := url.Parse(u)
		require.NoError(t, err)
		return parsed.Query().Get("login_hint")
	}
	assert.Equal(t, "me@example.com", hint(providerGoogle, "me@example.com"))
	assert.Equal(t, "me@contoso.example", hint(providerMicrosoft, "me@contoso.example"))
	assert.Empty(t, hint(providerGoogle, ""), "no hint unless asked")
	assert.Empty(t, hint(providerDropbox, "me@example.com"), "Dropbox takes no login hint")
}
