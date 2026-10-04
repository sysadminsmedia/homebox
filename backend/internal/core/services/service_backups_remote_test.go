package services

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/webdav"

	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

const (
	testUser = "backup"
	testPass = "s3cret-Passw0rd"
)

// startSFTPServer runs an in-process SFTP server on loopback and returns its
// address and host key fingerprint. It serves the real filesystem, so tests
// point the destination at a temp directory.
func startSFTPServer(t *testing.T, authorized ssh.PublicKey) (addr, fingerprint string) {
	t.Helper()
	_, hostPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	hostSigner, err := ssh.NewSignerFromKey(hostPriv)
	require.NoError(t, err)

	cfg := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, p []byte) (*ssh.Permissions, error) {
			if c.User() == testUser && string(p) == testPass {
				return nil, nil
			}
			return nil, errors.New("denied")
		},
		PublicKeyCallback: func(c ssh.ConnMetadata, k ssh.PublicKey) (*ssh.Permissions, error) {
			if authorized != nil && c.User() == testUser && string(k.Marshal()) == string(authorized.Marshal()) {
				return nil, nil
			}
			return nil, errors.New("denied")
		},
	}
	cfg.AddHostKey(hostSigner)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go serveSFTPConn(c, cfg)
		}
	}()
	return l.Addr().String(), ssh.FingerprintSHA256(hostSigner.PublicKey())
}

func serveSFTPConn(c net.Conn, cfg *ssh.ServerConfig) {
	_, chans, reqs, err := ssh.NewServerConn(c, cfg)
	if err != nil {
		_ = c.Close()
		return
	}
	go ssh.DiscardRequests(reqs)
	for nc := range chans {
		if nc.ChannelType() != "session" {
			_ = nc.Reject(ssh.UnknownChannelType, "unsupported")
			continue
		}
		ch, requests, err := nc.Accept()
		if err != nil {
			continue
		}
		go func(in <-chan *ssh.Request) {
			for r := range in {
				ok := r.Type == "subsystem" && len(r.Payload) > 4 && string(r.Payload[4:]) == "sftp"
				_ = r.Reply(ok, nil)
				if ok {
					srv, err := sftp.NewServer(ch)
					if err == nil {
						_ = srv.Serve()
					}
					_ = ch.Close()
				}
			}
		}(requests)
	}
}

func startWebDAV(t *testing.T, dir string) *httptest.Server {
	t.Helper()
	h := &webdav.Handler{FileSystem: webdav.Dir(dir), LockSystem: webdav.NewMemLS()}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != testUser || p != testPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="x"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func remoteSettings(typ, conn string) repo.BackupSettings {
	return repo.BackupSettings{
		Name: "remote", Type: typ, ConnString: conn, Username: testUser, Prefix: "homebox-backups",
		Enabled: true, Frequency: freqDaily, IntervalHours: 1, AtHour: 3, DayOfMonth: 1, SkipIfUnchanged: true,
		KeepDaily: 7, KeepWeekly: 4, KeepMonthly: 6, HealthIntervalMinutes: 15, AlertFailureThreshold: 2,
	}
}

func TestSecretBox(t *testing.T) {
	box, err := newSecretBox("k1")
	require.NoError(t, err)
	require.NotNil(t, box)

	none, err := newSecretBox("")
	require.NoError(t, err)
	assert.Nil(t, none, "no key disables credential storage")

	sealed, err := box.seal("g/d", backupSecret{Password: "pw", PrivateKey: "key"})
	require.NoError(t, err)
	assert.NotContains(t, sealed, "pw")

	got, err := box.open("g/d", sealed)
	require.NoError(t, err)
	assert.Equal(t, backupSecret{Password: "pw", PrivateKey: "key"}, got)

	sealed2, err := box.seal("g/d", backupSecret{Password: "pw"})
	require.NoError(t, err)
	assert.NotEqual(t, sealed, sealed2, "each seal uses a fresh nonce")

	_, err = box.open("g/other", sealed)
	require.Error(t, err, "a sealed secret is bound to its destination")

	other, _ := newSecretBox("k2")
	_, err = other.open("g/d", sealed)
	require.Error(t, err, "a different key cannot open it")

	tampered := []byte(sealed)
	tampered[len(tampered)/2] ^= 0x01
	_, err = box.open("g/d", string(tampered))
	require.Error(t, err)
	_, err = box.open("g/d", "!!not-base64!!")
	require.Error(t, err)
}

func TestNormalizeRemoteSettings(t *testing.T) {
	box, _ := newSecretBox("k")
	svc := &BackupService{cfg: config.BackupConf{Enabled: true, AllowCustomEndpoints: true}, secrets: box}

	t.Run("sftp normalizes the address and needs a host key", func(t *testing.T) {
		in := remoteSettings("sftp", "sftp://nas.lan/backups/homebox")
		_, err := svc.NormalizeSettings(in)
		require.ErrorIs(t, err, ErrBackupInvalid)

		in.HostKey = "SHA256:abc"
		out, err := svc.NormalizeSettings(in)
		require.NoError(t, err)
		assert.Equal(t, "sftp://nas.lan:22/backups/homebox", out.ConnString)

		// Testing may omit the host key so the first connection can read it.
		_, err = svc.normalize(remoteSettings("sftp", "sftp://nas.lan"), false)
		require.NoError(t, err)
	})
	t.Run("rejects embedded credentials, missing user and bad schemes", func(t *testing.T) {
		bad := func(typ, conn string, mut func(*repo.BackupSettings)) {
			in := remoteSettings(typ, conn)
			in.HostKey = "SHA256:abc"
			if mut != nil {
				mut(&in)
			}
			_, err := svc.NormalizeSettings(in)
			require.ErrorIs(t, err, ErrBackupInvalid, "%s %s", typ, conn)
		}
		bad("sftp", "sftp://u:p@nas.lan/x", nil)
		bad("sftp", "http://nas.lan/x", nil)
		bad("sftp", "sftp://", nil)
		bad("sftp", "sftp://nas.lan", func(s *repo.BackupSettings) { s.Username = "" })
		bad("webdav", "ftp://nas.lan", nil)
		bad("webdav", "https://u:p@nas.lan/dav", nil)
		bad("webdav", "https://nas.lan/dav", func(s *repo.BackupSettings) { s.Username = "" })
	})
	t.Run("webdav keeps its URL and clears the host key", func(t *testing.T) {
		in := remoteSettings("webdav", "https://cloud.example/remote.php/dav/files/me/")
		in.HostKey = "SHA256:ignored"
		out, err := svc.NormalizeSettings(in)
		require.NoError(t, err)
		assert.Empty(t, out.HostKey)
		assert.Equal(t, "https://cloud.example/remote.php/dav/files/me/", out.ConnString)
	})
	t.Run("needs an encryption key and custom endpoints", func(t *testing.T) {
		in := remoteSettings("webdav", "https://cloud.example/dav")
		noKey := &BackupService{cfg: config.BackupConf{Enabled: true, AllowCustomEndpoints: true}}
		_, err := noKey.NormalizeSettings(in)
		require.ErrorIs(t, err, ErrBackupInvalid)
		assert.Contains(t, err.Error(), "HBOX_BACKUP_ENCRYPTION_KEY")

		locked := &BackupService{cfg: config.BackupConf{Enabled: true}, secrets: box}
		_, err = locked.NormalizeSettings(in)
		require.ErrorIs(t, err, ErrBackupInvalid)

		assert.False(t, noKey.Options().RemoteEnabled)
		assert.False(t, locked.Options().RemoteEnabled)
		assert.True(t, svc.Options().RemoteEnabled)
	})
}

func TestSFTPDestinationEndToEnd(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	userPub, userPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sshPub, err := ssh.NewPublicKey(userPub)
	require.NoError(t, err)
	addr, fp := startSFTPServer(t, sshPub)
	conn := "sftp://" + addr + root

	grp, err := tRepos.Groups.GroupCreate(ctx, "sftp-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	// 1. First contact: the host key is unknown, so the test refuses but
	//    reports the fingerprint for the user to confirm.
	res, err := tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{
		BackupSettings: remoteSettings("sftp", conn), Password: testPass,
	})
	require.NoError(t, err)
	assert.False(t, res.OK)
	assert.Equal(t, fp, res.HostKey)

	// 2. A wrong password fails after the host key is trusted.
	settings := remoteSettings("sftp", conn)
	settings.HostKey = fp
	res, err = tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, Password: "nope"})
	require.NoError(t, err)
	assert.False(t, res.OK)
	assert.Empty(t, res.HostKey)

	// 3. The right password passes, and so does key authentication.
	res, err = tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, Password: testPass})
	require.NoError(t, err)
	require.True(t, res.OK, res.Message)

	pemBlock, err := ssh.MarshalPrivateKey(userPriv, "")
	require.NoError(t, err)
	keyPEM := string(pem.EncodeToMemory(pemBlock))
	res, err = tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, PrivateKey: keyPEM})
	require.NoError(t, err)
	require.True(t, res.OK, res.Message)

	// 4. Saving requires the fingerprint and credentials.
	_, err = tSvc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: remoteSettings("sftp", conn), Password: testPass})
	require.ErrorIs(t, err, ErrBackupInvalid)
	_, err = tSvc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: settings})
	require.ErrorIs(t, err, ErrBackupInvalid)

	dest, err := tSvc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, Password: testPass})
	require.NoError(t, err)
	assert.True(t, dest.HasSecret)

	// The credentials never leave the service: not in the API shape, and not
	// in plaintext at rest.
	js, err := json.Marshal(dest)
	require.NoError(t, err)
	assert.NotContains(t, string(js), testPass)
	assert.NotContains(t, string(js), "secret")
	assert.NotContains(t, dest.Secret, testPass)
	row, err := tRepos.BackupDestinations.Get(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	assert.NotContains(t, row.Secret, testPass)

	// 5. A backup lands on the server, downloads back, and prunes cleanly.
	exp, err := tRepos.Exports.CreateForDestination(ctx, grp.ID, dest.ID, "scheduled")
	require.NoError(t, err)
	tSvc.Exports.RunExport(ctx, exp.ID, grp.ID)
	exp, err = tRepos.Exports.Get(ctx, grp.ID, exp.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", exp.Status, exp.Error)

	onDisk := filepath.Join(root, filepath.FromSlash(exp.ArtifactPath))
	st, err := os.Stat(onDisk)
	require.NoError(t, err)
	assert.Equal(t, exp.SizeBytes, st.Size())
	_, err = os.Stat(onDisk + partSuffix)
	assert.True(t, os.IsNotExist(err), "no partial file is left behind")

	store, key, err := tSvc.Backups.ArtifactLocation(ctx, grp.ID, exp)
	require.NoError(t, err)
	rc, err := store.Open(ctx, key)
	require.NoError(t, err)
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	_ = rc.Close()
	_ = store.Close()
	want, err := os.ReadFile(onDisk)
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, []byte("PK"), got[:2], "the artifact is a zip")

	require.NoError(t, tSvc.Backups.DeleteVersion(ctx, grp.ID, exp))
	_, err = os.Stat(onDisk)
	assert.True(t, os.IsNotExist(err))

	// 6. A changed host key is refused and reported.
	bad := settings
	bad.HostKey = "SHA256:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	res, err = tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{BackupSettings: bad, Password: testPass})
	require.NoError(t, err)
	assert.False(t, res.OK)
	assert.Equal(t, fp, res.HostKey)
	assert.Contains(t, res.Message, "changed")

	// 7. Updates keep the secret for the same target but never carry it to a
	//    different address or login.
	same := settings
	same.Name = "renamed"
	upd, err := tSvc.Backups.UpdateDestination(ctx, grp.ID, dest.ID, repo.BackupInput{BackupSettings: same})
	require.NoError(t, err)
	assert.True(t, upd.HasSecret)
	assert.Equal(t, row.Secret, upd.Secret, "unchanged target keeps the sealed secret")

	moved := same
	moved.ConnString = "sftp://" + addr + root + "/elsewhere"
	_, err = tSvc.Backups.UpdateDestination(ctx, grp.ID, dest.ID, repo.BackupInput{BackupSettings: moved})
	require.ErrorIs(t, err, ErrBackupInvalid)
	upd, err = tSvc.Backups.UpdateDestination(ctx, grp.ID, dest.ID, repo.BackupInput{BackupSettings: moved, Password: testPass})
	require.NoError(t, err)
	assert.NotEqual(t, row.Secret, upd.Secret)

	// Testing an edit can reuse stored credentials only for an unchanged target.
	res, err = tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{BackupSettings: moved, DestinationID: dest.ID.String()})
	require.NoError(t, err)
	assert.True(t, res.OK, res.Message)
	other := moved
	other.Username = "someone-else"
	_, err = tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{BackupSettings: other, DestinationID: dest.ID.String()})
	require.ErrorIs(t, err, ErrBackupInvalid)

	// 8. A sealed secret copied to another destination row cannot be opened.
	second, err := tSvc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, Password: testPass})
	require.NoError(t, err)
	clone := second
	clone.Secret = upd.Secret
	_, err = tSvc.Backups.credentials(clone, nil)
	require.Error(t, err)
}

func TestWebDAVDestinationEndToEnd(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	srv := startWebDAV(t, dir)

	grp, err := tRepos.Groups.GroupCreate(ctx, "dav-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	settings := remoteSettings("webdav", srv.URL+"/")

	res, err := tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, Password: "wrong"})
	require.NoError(t, err)
	assert.False(t, res.OK)

	res, err = tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, Password: testPass})
	require.NoError(t, err)
	require.True(t, res.OK, res.Message)

	_, err = tSvc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: settings})
	require.ErrorIs(t, err, ErrBackupInvalid, "a password is required")

	dest, err := tSvc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, Password: testPass})
	require.NoError(t, err)

	exp, err := tRepos.Exports.CreateForDestination(ctx, grp.ID, dest.ID, "scheduled")
	require.NoError(t, err)
	tSvc.Exports.RunExport(ctx, exp.ID, grp.ID)
	exp, err = tRepos.Exports.Get(ctx, grp.ID, exp.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", exp.Status, exp.Error)

	onDisk := filepath.Join(dir, filepath.FromSlash(exp.ArtifactPath))
	st, err := os.Stat(onDisk)
	require.NoError(t, err)
	assert.Equal(t, exp.SizeBytes, st.Size())
	_, err = os.Stat(onDisk + partSuffix)
	assert.True(t, os.IsNotExist(err))

	store, key, err := tSvc.Backups.ArtifactLocation(ctx, grp.ID, exp)
	require.NoError(t, err)
	rc, err := store.Open(ctx, key)
	require.NoError(t, err)
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	_ = rc.Close()
	_ = store.Close()
	assert.Len(t, got, int(exp.SizeBytes))

	// Deleting twice is fine; a missing artifact counts as deleted.
	require.NoError(t, tSvc.Backups.DeleteVersion(ctx, grp.ID, exp))
	_, err = os.Stat(onDisk)
	assert.True(t, os.IsNotExist(err))
	require.NoError(t, tSvc.Backups.DeleteVersion(ctx, grp.ID, exp))

	// The password is not stored in the clear.
	row, err := tRepos.BackupDestinations.Get(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	assert.NotContains(t, row.Secret, testPass)
}

func TestRemoteStoresRejectPathEscape(t *testing.T) {
	s := &sftpStore{base: "/srv/backups"}
	for _, k := range []string{"../etc/passwd", "a/../../x", "/../x"} {
		_, err := s.full(k)
		require.Error(t, err, k)
	}
	p, err := s.full("p/g/backups/x.zip")
	require.NoError(t, err)
	assert.Equal(t, "/srv/backups/p/g/backups/x.zip", p)
}

func TestSFTPPassphraseProtectedKey(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sshPub, err := ssh.NewPublicKey(pub)
	require.NoError(t, err)
	addr, fp := startSFTPServer(t, sshPub)
	grp, err := tRepos.Groups.GroupCreate(ctx, "pass-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	const passphrase = "correct horse battery staple"
	block, err := ssh.MarshalPrivateKeyWithPassphrase(priv, "", []byte(passphrase))
	require.NoError(t, err)
	locked := string(pem.EncodeToMemory(block))
	plainBlock, err := ssh.MarshalPrivateKey(priv, "")
	require.NoError(t, err)
	plain := string(pem.EncodeToMemory(plainBlock))

	settings := remoteSettings("sftp", "sftp://"+addr+root)
	settings.HostKey = fp
	test := func(in repo.BackupInput) TestResult {
		in.BackupSettings = settings
		res, err := tSvc.Backups.TestSettings(ctx, grp.ID, in)
		require.NoError(t, err)
		return res
	}

	res := test(repo.BackupInput{PrivateKey: locked, Passphrase: passphrase})
	assert.True(t, res.OK, res.Message)

	res = test(repo.BackupInput{PrivateKey: locked})
	assert.False(t, res.OK)
	assert.Contains(t, res.Message, "passphrase")

	res = test(repo.BackupInput{PrivateKey: locked, Passphrase: "wrong"})
	assert.False(t, res.OK)
	assert.Contains(t, res.Message, "wrong")
	assert.NotContains(t, res.Message, "BEGIN", "the key never appears in messages")

	res = test(repo.BackupInput{PrivateKey: plain, Passphrase: "not needed"})
	assert.True(t, res.OK, "a passphrase given for an unencrypted key is harmless: %s", res.Message)

	res = test(repo.BackupInput{PrivateKey: "not a key"})
	assert.False(t, res.OK)

	// The passphrase is stored sealed with the key and survives a round trip.
	dest, err := tSvc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, PrivateKey: locked, Passphrase: passphrase})
	require.NoError(t, err)
	assert.NotContains(t, dest.Secret, passphrase)
	sec, err := tSvc.Backups.credentials(dest, nil)
	require.NoError(t, err)
	assert.Equal(t, passphrase, sec.Passphrase)
	got, err := tSvc.Backups.TestDestination(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	assert.True(t, got.OK, got.Message)
}

func TestParseSMBURL(t *testing.T) {
	ok := func(raw string, want smbTarget) {
		got, err := parseSMBURL(raw)
		require.NoError(t, err, raw)
		assert.Equal(t, want, got, raw)
	}
	ok("smb://nas.lan/backups", smbTarget{addr: "nas.lan:445", share: "backups"})
	ok("smb://nas.lan:1445/backups/homebox/daily", smbTarget{addr: "nas.lan:1445", share: "backups", base: "homebox/daily"})
	ok("smb://nas.lan/backups/../backups/x", smbTarget{addr: "nas.lan:445", share: "backups", base: "x"})
	for _, raw := range []string{"smb://nas.lan", "smb://nas.lan/", "smb://user:pw@nas.lan/share", "http://nas.lan/share", "smb:///share", "nas.lan/share"} {
		_, err := parseSMBURL(raw)
		require.Error(t, err, raw)
	}
}

func TestSMBPathGuard(t *testing.T) {
	s := &smbStore{base: "homebox/daily"}
	for _, k := range []string{"../x", "a/../../x", "/../etc"} {
		_, err := s.full(k)
		require.Error(t, err, k)
	}
	p, err := s.full("p/g/backups/x.zip")
	require.NoError(t, err)
	assert.Equal(t, "homebox/daily/p/g/backups/x.zip", p)

	root := &smbStore{}
	p, err = root.full("a/b.zip")
	require.NoError(t, err)
	assert.Equal(t, "a/b.zip", p)
}

func TestNormalizeSMBSettings(t *testing.T) {
	box, _ := newSecretBox("k")
	svc := &BackupService{cfg: config.BackupConf{Enabled: true, AllowCustomEndpoints: true}, secrets: box}
	in := remoteSettings("smb", "smb://nas.lan/backups/homebox")
	in.HostKey = "SHA256:ignored"
	out, err := svc.NormalizeSettings(in)
	require.NoError(t, err)
	assert.Equal(t, "smb://nas.lan:445/backups/homebox", out.ConnString)
	assert.Empty(t, out.HostKey)

	for _, conn := range []string{"smb://nas.lan", "smb://u:p@nas.lan/s", "ftp://nas.lan/s"} {
		in := remoteSettings("smb", conn)
		_, err := svc.NormalizeSettings(in)
		require.ErrorIs(t, err, ErrBackupInvalid, conn)
	}
	noName := remoteSettings("smb", "smb://nas.lan/s")
	noName.Username = ""
	_, err = svc.NormalizeSettings(noName)
	require.ErrorIs(t, err, ErrBackupInvalid)

	// A key is never stored for a password-only type.
	sealed, err := svc.sealSecret(uuid.New(), uuid.New(), "smb", repo.BackupInput{Password: "pw", PrivateKey: "k", Passphrase: "p"})
	require.NoError(t, err)
	assert.NotEmpty(t, sealed)
	_, err = svc.sealSecret(uuid.New(), uuid.New(), "smb", repo.BackupInput{PrivateKey: "k"})
	require.ErrorIs(t, err, ErrBackupInvalid)
}

// TestSMBDestinationAgainstServer runs the real pipeline against a live SMB
// server. Point it at a disposable share:
//
//	HBOX_TEST_SMB_URL=smb://127.0.0.1:1445/backups HBOX_TEST_SMB_USER=hbuser HBOX_TEST_SMB_PASS=... go test -run SMBDestination
func TestSMBDestinationAgainstServer(t *testing.T) {
	conn, user, pass := os.Getenv("HBOX_TEST_SMB_URL"), os.Getenv("HBOX_TEST_SMB_USER"), os.Getenv("HBOX_TEST_SMB_PASS")
	if conn == "" {
		t.Skip("set HBOX_TEST_SMB_URL, HBOX_TEST_SMB_USER and HBOX_TEST_SMB_PASS to run against a disposable SMB share")
	}
	ctx := context.Background()
	grp, err := tRepos.Groups.GroupCreate(ctx, "smb-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	settings := remoteSettings("smb", conn)
	settings.Username = user

	res, err := tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, Password: "wrong-" + pass})
	require.NoError(t, err)
	assert.False(t, res.OK, "a wrong password is refused")

	res, err = tSvc.Backups.TestSettings(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, Password: pass})
	require.NoError(t, err)
	require.True(t, res.OK, res.Message)

	dest, err := tSvc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: settings, Password: pass})
	require.NoError(t, err)
	assert.NotContains(t, dest.Secret, pass)

	exp, err := tRepos.Exports.CreateForDestination(ctx, grp.ID, dest.ID, "scheduled")
	require.NoError(t, err)
	tSvc.Exports.RunExport(ctx, exp.ID, grp.ID)
	exp, err = tRepos.Exports.Get(ctx, grp.ID, exp.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", exp.Status, exp.Error)

	store, key, err := tSvc.Backups.ArtifactLocation(ctx, grp.ID, exp)
	require.NoError(t, err)
	rc, err := store.Open(ctx, key)
	require.NoError(t, err)
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	_ = rc.Close()
	_ = store.Close()
	assert.Len(t, got, int(exp.SizeBytes))
	assert.Equal(t, []byte("PK"), got[:2])

	// Overwriting works (SMB rename will not replace), and deleting twice is fine.
	st, _, err := tSvc.Backups.openStore(ctx, dest, nil)
	require.NoError(t, err)
	require.NoError(t, st.Write(ctx, key, bytesReader([]byte("replacement")), 11, "application/zip"))
	require.NoError(t, st.Write(ctx, key, bytesReader([]byte("replaced again")), 14, "application/zip"))
	rc, err = st.Open(ctx, key)
	require.NoError(t, err)
	got, err = io.ReadAll(rc)
	require.NoError(t, err)
	_ = rc.Close()
	assert.Equal(t, "replaced again", string(got))
	_ = st.Close()

	require.NoError(t, tSvc.Backups.DeleteVersion(ctx, grp.ID, exp))
	require.NoError(t, tSvc.Backups.DeleteVersion(ctx, grp.ID, exp))
}
