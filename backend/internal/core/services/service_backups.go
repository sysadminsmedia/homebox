package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gocloud.dev/blob"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// ErrBackupInvalid wraps every user-correctable problem with a destination's
// settings so handlers can answer 422 instead of 500.
var ErrBackupInvalid = errors.New("invalid backup destination")

// ErrBackupDisabled is returned when scheduled backups are switched off by
// configuration.
var ErrBackupDisabled = errors.New("scheduled backups are disabled")

const (
	destTypePrimary = "primary"
	destTypeLocal   = "local"
	destTypeSFTP    = "sftp"
	destTypeWebDAV  = "webdav"
	destTypeSMB     = "smb"

	originManual    = "manual"
	originScheduled = "scheduled"

	healthHealthy     = "healthy"
	healthUnreachable = "unreachable"

	// failedRowRetention bounds how long failed backup rows (which hold no
	// artifact) stay in the history.
	failedRowRetention = 14 * 24 * time.Hour
	checkTimeout       = 30 * time.Second

	maxDestinationsPerGroup = 20

	// activeRunTimeout is how long a pending or running backup may go without any
	// update before it is treated as interrupted. A live run updates its row as it
	// progresses, so this only has to outlast the longest silent step.
	activeRunTimeout = 6 * time.Hour
)

// BackupService owns scheduled backups: destination management, the
// scheduler, retention pruning, health checks and alerting. The actual zip
// building and upload is still ExportService's job; scheduled runs are just
// export rows bound to a destination.
type BackupService struct {
	repos          *repo.AllRepos
	db             *ent.Client
	exports        *ExportService
	cfg            config.BackupConf
	notifierConfig *config.NotifierConf
	dialect        string
	bus            *eventbus.EventBus
	// secrets seals sftp/webdav/cloud-drive credentials; nil when no
	// encryption key is configured, which disables those destination types.
	secrets *secretBox

	oauth oauthPending
	// endpointOverrides and httpc let tests point providers at fake servers.
	endpointOverrides map[string]providerEndpoints
	httpc             *http.Client

	tokensMu sync.Mutex
	tokens   map[uuid.UUID]*tokenSource
}

// Enabled reports whether scheduled backups are switched on.
func (s *BackupService) Enabled() bool { return s.cfg.Enabled }

// BackupOptions describes what the server allows, for the UI.
type BackupOptions struct {
	Enabled              bool
	LocalEnabled         bool
	AllowCustomEndpoints bool
	// RemoteEnabled reports whether sftp and webdav destinations can be used:
	// they need an encryption key for their credentials and custom endpoints.
	RemoteEnabled bool
	// OAuthProviders are the cloud drives the operator configured: any of
	// providerGoogle, providerMicrosoft and providerDropbox.
	OAuthProviders []string
}

// Options returns the server-side backup switches.
func (s *BackupService) Options() BackupOptions {
	return BackupOptions{
		Enabled:              s.cfg.Enabled,
		LocalEnabled:         s.cfg.LocalRoot != "",
		AllowCustomEndpoints: s.cfg.AllowCustomEndpoints,
		RemoteEnabled:        s.secrets != nil && s.cfg.AllowCustomEndpoints,
		OAuthProviders:       s.OAuthProviderKeys(),
	}
}

// TestResult is the outcome of a connection test or health probe.
type TestResult struct {
	OK        bool   `json:"ok"`
	LatencyMs int64  `json:"latencyMs"`
	Message   string `json:"message"`
	// HostKey is the SSH host key fingerprint the server presented when it
	// was missing or did not match, so the UI can offer to trust it.
	HostKey string `json:"hostKey,omitempty"`
}

// invalid wraps a message as ErrBackupInvalid, the class of error a user can fix.
func invalid(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrBackupInvalid, fmt.Sprintf(format, a...))
}

func (s *BackupService) NormalizeSettings(in repo.BackupSettings) (repo.BackupSettings, error) {
	return s.normalize(in, true)
}

// isRemoteType reports whether a type needs a stored username and password.
func isRemoteType(t string) bool { return t == destTypeSFTP || t == destTypeWebDAV || t == destTypeSMB }

// passwordOnly reports whether a remote type authenticates by password alone.
func passwordOnly(t string) bool { return t == destTypeWebDAV || t == destTypeSMB }

func (s *BackupService) requireRemote() error {
	if s.secrets == nil {
		return invalid("set HBOX_BACKUP_ENCRYPTION_KEY to use SFTP or WebDAV destinations")
	}
	if !s.cfg.AllowCustomEndpoints {
		return invalid("custom endpoints are disabled on this server")
	}
	return nil
}

// normalize validates in and fills defaults. requireHostKey is false only when
// testing, so a first connection can read the server's host key.
func (s *BackupService) normalize(in repo.BackupSettings, requireHostKey bool) (repo.BackupSettings, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return in, invalid("name is required")
	}

	prefix, err := cleanPrefix(in.Prefix)
	if err != nil {
		return in, err
	}
	in.Prefix = prefix

	if in.Frequency == "cron" {
		in.CronExpr = strings.TrimSpace(in.CronExpr)
		if _, err := parseCron(in.CronExpr); err != nil {
			return in, invalid("schedule: %v", err)
		}
	} else {
		in.CronExpr = ""
	}

	in.ConnString = strings.TrimSpace(in.ConnString)
	in.Username = strings.TrimSpace(in.Username)
	in.HostKey = strings.TrimSpace(in.HostKey)
	if !isRemoteType(in.Type) {
		in.Username, in.HostKey = "", ""
	}

	switch in.Type {
	case destTypePrimary:
		in.ConnString = ""
	case destTypeLocal:
		if s.cfg.LocalRoot == "" {
			return in, invalid("local destinations are disabled; set HBOX_BACKUP_LOCAL_ROOT")
		}
		if _, err := s.localDir(in.ConnString); err != nil {
			return in, err
		}
	case "s3", "gcs", "azblob":
		if err := s.validateCloudURL(in.Type, in.ConnString); err != nil {
			return in, err
		}
	case destTypeSFTP:
		if err := s.requireRemote(); err != nil {
			return in, err
		}
		t, err := parseSFTPURL(in.ConnString)
		if err != nil {
			return in, invalid("address: %v", err)
		}
		in.ConnString = "sftp://" + t.addr + t.base
		if in.Username == "" {
			return in, invalid("a username is required")
		}
		if requireHostKey && !strings.HasPrefix(in.HostKey, "SHA256:") {
			return in, invalid("the SSH host key fingerprint is required; run Test connection to read it and confirm it")
		}
	case destTypeWebDAV:
		if err := s.requireRemote(); err != nil {
			return in, err
		}
		u, err := parseWebDAVURL(in.ConnString)
		if err != nil {
			return in, invalid("address: %v", err)
		}
		in.ConnString = u.String()
		in.HostKey = "" // SSH only
		if in.Username == "" {
			return in, invalid("a username is required")
		}
	case destTypeSMB:
		if err := s.requireRemote(); err != nil {
			return in, err
		}
		t, err := parseSMBURL(in.ConnString)
		if err != nil {
			return in, invalid("address: %v", err)
		}
		in.ConnString = "smb://" + t.addr + "/" + t.share
		if t.base != "" {
			in.ConnString += "/" + t.base
		}
		in.HostKey = "" // SSH only
		if in.Username == "" {
			return in, invalid("a username is required")
		}
	case destTypeGDrive, destTypeOneDrive, destTypeDropbox:
		if _, ok := s.providerByType(in.Type); !ok || s.secrets == nil {
			return in, invalid("this cloud provider is not configured on the server")
		}
		// The address is the provider's; the account name is set when the
		// account is connected, never taken from the request.
		in.ConnString = ""
	default:
		return in, invalid("unknown destination type %q", in.Type)
	}
	return in, nil
}

// cleanPrefix normalizes the key prefix: no leading/trailing slashes, no
// traversal segments, no backslashes.
func cleanPrefix(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "homebox-backups", nil
	}
	if strings.ContainsAny(p, "\\\x00") {
		return "", invalid("prefix contains invalid characters")
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return "", invalid("prefix must not contain '..'")
		}
	}
	p = strings.Trim(path.Clean("/"+p), "/")
	if p == "" {
		return "homebox-backups", nil
	}
	return p, nil
}

// localDir resolves a local destination's sub-directory under the configured
// root, refusing anything that would escape it.
func (s *BackupService) localDir(sub string) (string, error) {
	sub = strings.TrimSpace(sub)
	if sub == "" {
		return "", invalid("a directory name is required for local destinations")
	}
	if filepath.IsAbs(sub) || strings.ContainsAny(sub, "\x00:") {
		return "", invalid("local directory must be a relative path")
	}
	root, err := filepath.Abs(s.cfg.LocalRoot)
	if err != nil {
		return "", invalid("invalid local backup root")
	}
	dir := filepath.Join(root, filepath.FromSlash(sub))
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", invalid("local directory must stay inside the backup root")
	}
	return dir, nil
}

// validateCloudURL checks a cloud connection string: the right scheme for the type, no embedded credentials, a named bucket, and no custom endpoint unless the operator allowed them.
func (s *BackupService) validateCloudURL(typ, raw string) error {
	if raw == "" {
		return invalid("a connection string is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return invalid("connection string is not a valid URL")
	}
	if u.Scheme != typ {
		return invalid("connection string must start with %s://", typ)
	}
	if u.User != nil {
		return invalid("credentials must not be embedded in the connection string; use the provider's environment variables")
	}
	if u.Host == "" {
		return invalid("connection string must name a bucket or container")
	}
	if !s.cfg.AllowCustomEndpoints {
		q := u.Query()
		for _, k := range []string{"endpoint", "domain", "protocol", "hostname_immutable", "disable_https", "use_path_style"} {
			if q.Has(k) {
				return invalid("custom endpoints are disabled on this server (%s)", k)
			}
		}
	}
	return nil
}

// redact replaces the server's absolute local backup root in msg, so error
// text shown to group owners never reveals the host's directory layout.
func (s *BackupService) redact(msg string) string {
	if s.cfg.LocalRoot == "" {
		return msg
	}
	root := s.cfg.LocalRoot
	if abs, err := filepath.Abs(root); err == nil {
		msg = strings.ReplaceAll(msg, abs, "<backup root>")
	}
	return strings.ReplaceAll(msg, root, "<backup root>")
}

func (s *BackupService) openStore(ctx context.Context, d repo.BackupDestinationOut, sec *backupSecret) (objectStore, func(string) string, error) {
	identity := func(p string) string { return p }
	switch d.Type {
	case destTypePrimary:
		b, err := blob.OpenBucket(ctx, s.repos.Attachments.GetConnString())
		if err != nil {
			return nil, nil, err
		}
		return blobStore{b}, s.repos.Attachments.GetFullPath, nil
	case destTypeLocal:
		dir, err := s.localDir(d.ConnString)
		if err != nil {
			return nil, nil, err
		}
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, nil, fmt.Errorf("create backup directory: %s", s.redact(err.Error()))
		}
		slash := filepath.ToSlash(dir)
		if !strings.HasPrefix(slash, "/") {
			slash = "/" + slash
		}
		u := url.URL{Scheme: "file", Path: slash, RawQuery: "no_tmp_dir=true"}
		b, err := blob.OpenBucket(ctx, u.String())
		if err != nil {
			return nil, nil, errors.New(s.redact(err.Error()))
		}
		return blobStore{b}, identity, nil
	case destTypeSFTP, destTypeWebDAV, destTypeSMB:
		cred, err := s.credentials(d, sec)
		if err != nil {
			return nil, nil, err
		}
		if d.Type == destTypeSMB {
			t, err := parseSMBURL(d.ConnString)
			if err != nil {
				return nil, nil, err
			}
			st, err := dialSMB(t, d.Username, cred.Password)
			if err != nil {
				return nil, nil, err
			}
			return st, identity, nil
		}
		if d.Type == destTypeWebDAV {
			st, err := newDAVStore(d.ConnString, d.Username, cred.Password)
			if err != nil {
				return nil, nil, err
			}
			return st, identity, nil
		}
		t, err := parseSFTPURL(d.ConnString)
		if err != nil {
			return nil, nil, err
		}
		st, err := dialSFTP(t, d.Username, cred, d.HostKey)
		if err != nil {
			return nil, nil, err
		}
		return st, identity, nil
	case destTypeGDrive, destTypeOneDrive, destTypeDropbox:
		st, err := s.driveStore(d, sec)
		if err != nil {
			return nil, nil, err
		}
		return st, identity, nil
	default:
		if err := s.validateCloudURL(d.Type, d.ConnString); err != nil {
			return nil, nil, err
		}
		b, err := blob.OpenBucket(ctx, d.ConnString)
		if err != nil {
			return nil, nil, err
		}
		return blobStore{b}, identity, nil
	}
}

// credentials returns the login for an sftp or webdav destination: sec when
// supplied (a test of unsaved settings), otherwise the decrypted stored secret.
func (s *BackupService) credentials(d repo.BackupDestinationOut, sec *backupSecret) (backupSecret, error) {
	if sec != nil {
		return *sec, nil
	}
	if s.secrets == nil {
		return backupSecret{}, errors.New("HBOX_BACKUP_ENCRYPTION_KEY is not set, so stored credentials cannot be read")
	}
	if d.Secret == "" {
		return backupSecret{}, errors.New("no credentials are stored for this destination")
	}
	return s.secrets.open(secretAAD(d.GroupID, d.ID), d.Secret)
}

func secretAAD(gid, id uuid.UUID) string { return gid.String() + "/" + id.String() }

// sealSecret seals the credentials in in for a remote destination of type typ.
func (s *BackupService) sealSecret(gid, id uuid.UUID, typ string, in repo.BackupInput) (string, error) {
	sec := backupSecret{Password: in.Password, PrivateKey: in.PrivateKey, Passphrase: in.Passphrase}
	if passwordOnly(typ) {
		sec.PrivateKey, sec.Passphrase = "", ""
	}
	if sec.empty() {
		if passwordOnly(typ) {
			return "", invalid("a password is required")
		}
		return "", invalid("a password or private key is required")
	}
	return s.secrets.seal(secretAAD(gid, id), sec)
}

// artifactPath builds the artifact path recorded on the export row. Primary
// storage keeps the historical "{gid}/exports/{id}.zip" layout so existing
// download and sweep logic applies unchanged; other destinations get a
// browsable "{prefix}/{gid}/backups/{timestamp}-{id}.zip".
func artifactPath(d repo.BackupDestinationOut, gid, exportID uuid.UUID, at time.Time) string {
	if d.Type == destTypePrimary {
		return fmt.Sprintf("%s/exports/%s.zip", gid, exportID)
	}
	return fmt.Sprintf("%s/%s/backups/%s-%s.zip", d.Prefix, gid, at.UTC().Format("20060102-150405"), exportID.String()[:8])
}

// artifactPrefix is the only key prefix a destination's artifacts may live
// under; handlers use it as a defence-in-depth check against tampered rows.
func artifactPrefix(d repo.BackupDestinationOut, gid uuid.UUID) string {
	if d.Type == destTypePrimary {
		return gid.String() + "/exports/"
	}
	return fmt.Sprintf("%s/%s/backups/", d.Prefix, gid)
}

// driveStore builds the client for a cloud-drive destination. Access tokens
// are cached per destination, and a refresh token the provider rotates is
// written back (sealed) so the next refresh still works.
func (s *BackupService) driveStore(d repo.BackupDestinationOut, sec *backupSecret) (objectStore, error) {
	p, ok := s.providerByType(d.Type)
	if !ok {
		return nil, errors.New("this cloud provider is not configured on the server")
	}
	cred, err := s.credentials(d, sec)
	if err != nil {
		return nil, err
	}
	if cred.RefreshToken == "" {
		return nil, errors.New("no account is connected; edit the destination and connect the account")
	}

	hc := s.httpClient()
	ts := s.tokenSourceFor(d, p, cred.RefreshToken, sec != nil, hc)
	if sec != nil && sec.accessToken != "" {
		ts.access, ts.expiry = sec.accessToken, sec.accessExpiry
	}
	api := &driveAPI{ts: ts, hc: hc}
	switch d.Type {
	case destTypeGDrive:
		return newGoogleStore(api, p.EP), nil
	case destTypeOneDrive:
		return newOneDriveStore(api, p.EP, hc), nil
	default:
		return newDropboxStore(api, p.EP), nil
	}
}

func (s *BackupService) tokenSourceFor(d repo.BackupDestinationOut, p oauthProvider, refresh string, transient bool, hc *http.Client) *tokenSource {
	if transient {
		return &tokenSource{prov: p, hc: hc, refresh: refresh}
	}
	s.tokensMu.Lock()
	defer s.tokensMu.Unlock()
	if ts, ok := s.tokens[d.ID]; ok && ts.prov.Key == p.Key {
		ts.mu.Lock()
		same := ts.refresh == refresh
		ts.mu.Unlock()
		if same {
			return ts
		}
	}
	ts := &tokenSource{prov: p, hc: hc, refresh: refresh}
	gid, id := d.GroupID, d.ID
	ts.onRotate = func(newRefresh string) {
		sealed, err := s.secrets.seal(secretAAD(gid, id), backupSecret{RefreshToken: newRefresh})
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err = s.repos.BackupDestinations.SetSecret(ctx, id, sealed)
		}
		if err != nil {
			log.Err(err).Stringer("destination_id", id).Msg("backup: could not store the rotated refresh token")
		}
	}
	if s.tokens == nil {
		s.tokens = map[uuid.UUID]*tokenSource{}
	}
	s.tokens[id] = ts
	return ts
}

// ArtifactLocation resolves where an export's artifact lives. The caller owns
// the returned store and must Close it.
func (s *BackupService) ArtifactLocation(ctx context.Context, gid uuid.UUID, exp repo.ExportOut) (objectStore, string, error) {
	if exp.ArtifactPath == "" {
		return nil, "", errors.New("export has no artifact")
	}
	clean := path.Clean(exp.ArtifactPath)
	var dest repo.BackupDestinationOut
	if exp.DestinationID != nil {
		d, err := s.repos.BackupDestinations.Get(ctx, gid, *exp.DestinationID)
		if err != nil {
			return nil, "", err
		}
		dest = d
	} else {
		dest = repo.BackupDestinationOut{BackupSettings: repo.BackupSettings{Type: destTypePrimary}}
	}
	if !strings.HasPrefix(clean, artifactPrefix(dest, gid)) {
		return nil, "", errors.New("artifact outside the expected prefix")
	}
	st, key, err := s.openStore(ctx, dest, nil)
	if err != nil {
		return nil, "", err
	}
	return st, key(clean), nil
}

func (s *BackupService) Check(ctx context.Context, d repo.BackupDestinationOut, gid uuid.UUID) TestResult {
	return s.check(ctx, d, gid, nil)
}

// check probes d. sec supplies credentials for unsaved settings.
func (s *BackupService) check(ctx context.Context, d repo.BackupDestinationOut, gid uuid.UUID, sec *backupSecret) TestResult {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	start := time.Now()
	fail := func(err error) TestResult {
		res := TestResult{OK: false, LatencyMs: time.Since(start).Milliseconds(), Message: s.redact(err.Error())}
		var hk *hostKeyError
		if errors.As(err, &hk) {
			res.HostKey = hk.Fingerprint
		}
		return res
	}

	st, key, err := s.openStore(ctx, d, sec)
	if err != nil {
		return fail(err)
	}
	defer func() { _ = st.Close() }()

	probe := fmt.Sprintf(".homebox-healthcheck-%s", uuid.NewString())
	if d.Type == destTypePrimary {
		probe = fmt.Sprintf("%s/exports/%s", gid, probe)
	} else if d.Prefix != "" {
		probe = d.Prefix + "/" + probe
	}
	k := key(probe)
	if err := st.Write(ctx, k, probeBytes(), 1024, "application/octet-stream"); err != nil {
		return fail(fmt.Errorf("write test object: %w", err))
	}
	if err := st.Delete(ctx, k); err != nil {
		return fail(fmt.Errorf("delete test object: %w", err))
	}
	return TestResult{OK: true, LatencyMs: time.Since(start).Milliseconds(), Message: "Wrote and deleted a 1 KB test file"}
}

// ---------------------------------------------------------------------------
// Destination CRUD

// ListDestinations returns a collection's backup destinations, oldest first.
func (s *BackupService) ListDestinations(ctx context.Context, gid uuid.UUID) ([]repo.BackupDestinationOut, error) {
	return s.repos.BackupDestinations.ListByGroup(ctx, gid)
}

// GetDestination returns one destination of the collection.
func (s *BackupService) GetDestination(ctx context.Context, gid, id uuid.UUID) (repo.BackupDestinationOut, error) {
	return s.repos.BackupDestinations.Get(ctx, gid, id)
}

// CreateDestination validates and stores a new destination, scheduling its first run when scheduling is on.
func (s *BackupService) CreateDestination(ctx context.Context, gid uuid.UUID, in repo.BackupInput) (repo.BackupDestinationOut, error) {
	if !s.cfg.Enabled {
		return repo.BackupDestinationOut{}, ErrBackupDisabled
	}
	existing, err := s.repos.BackupDestinations.ListByGroup(ctx, gid)
	if err != nil {
		return repo.BackupDestinationOut{}, err
	}
	if len(existing) >= maxDestinationsPerGroup {
		return repo.BackupDestinationOut{}, invalid("a collection can have at most %d backup destinations", maxDestinationsPerGroup)
	}
	st, err := s.normalize(in.BackupSettings, true)
	if err != nil {
		return repo.BackupDestinationOut{}, err
	}
	id := uuid.New()
	var secret string
	switch {
	case isRemoteType(st.Type):
		if secret, err = s.sealSecret(gid, id, st.Type, in); err != nil {
			return repo.BackupDestinationOut{}, err
		}
	case isDriveType(st.Type):
		if in.OAuthTicket == "" {
			return repo.BackupDestinationOut{}, invalid("connect your account first")
		}
		sec, account, terr := s.ticketCredentials(gid, in.OAuthTicket, st.Type, true)
		if terr != nil {
			return repo.BackupDestinationOut{}, terr
		}
		if secret, err = s.secrets.seal(secretAAD(gid, id), sec); err != nil {
			return repo.BackupDestinationOut{}, err
		}
		st.Username = account
	}
	return s.repos.BackupDestinations.Create(ctx, gid, id, st, secret, s.initialNextRun(st))
}

// UpdateDestination replaces a destination's settings. Its type, address and prefix cannot change while it holds backups, since each stored file is located through them.
func (s *BackupService) UpdateDestination(ctx context.Context, gid, id uuid.UUID, in repo.BackupInput) (repo.BackupDestinationOut, error) {
	if !s.cfg.Enabled {
		return repo.BackupDestinationOut{}, ErrBackupDisabled
	}
	cur, err := s.repos.BackupDestinations.Get(ctx, gid, id)
	if err != nil {
		return repo.BackupDestinationOut{}, err
	}
	st, err := s.normalize(in.BackupSettings, true)
	if err != nil {
		return repo.BackupDestinationOut{}, err
	}
	relocated := cur.Type != st.Type || cur.ConnString != st.ConnString || cur.Prefix != st.Prefix
	if relocated {
		// Each stored backup records its path relative to the destination, and is
		// found again by resolving that path against the destination's current
		// settings. Changing where the destination points would leave those files
		// unreachable for download, pruning and deletion.
		n, cerr := s.repos.Exports.CountArtifactsForDestination(ctx, gid, id)
		if cerr != nil {
			return repo.BackupDestinationOut{}, cerr
		}
		if n > 0 {
			return repo.BackupDestinationOut{}, invalid("this destination already holds %d backup(s) at its current location, so its type, address and folder prefix can't change; delete those backups first, or add a new destination", n)
		}
	}
	moved := relocated || cur.Username != st.Username || cur.HostKey != st.HostKey

	var secret string
	if isDriveType(st.Type) {
		switch {
		case in.OAuthTicket != "":
			sec, account, terr := s.ticketCredentials(gid, in.OAuthTicket, st.Type, true)
			if terr != nil {
				return repo.BackupDestinationOut{}, terr
			}
			if secret, err = s.secrets.seal(secretAAD(gid, id), sec); err != nil {
				return repo.BackupDestinationOut{}, err
			}
			st.Username = account
			moved = true
		case cur.Secret != "" && cur.Type == st.Type:
			secret, st.Username = cur.Secret, cur.Username
		default:
			return repo.BackupDestinationOut{}, invalid("connect your account first")
		}
	}
	if isRemoteType(st.Type) {
		switch {
		case in.Password != "" || in.PrivateKey != "":
			if secret, err = s.sealSecret(gid, id, st.Type, in); err != nil {
				return repo.BackupDestinationOut{}, err
			}
		case cur.Secret != "" && cur.Type == st.Type && cur.ConnString == st.ConnString && cur.Username == st.Username:
			// Unchanged target: keep the stored credentials. They are never
			// carried over to a different address or login, so a changed
			// destination cannot be pointed at a server that harvests them.
			secret = cur.Secret
		default:
			return repo.BackupDestinationOut{}, invalid("re-enter the credentials when changing the type, address or username")
		}
	}
	out, err := s.repos.BackupDestinations.Update(ctx, gid, id, st, secret, s.initialNextRun(st), moved)
	if err == nil {
		s.publishMutation(gid)
	}
	return out, err
}

// initialNextRun returns the first scheduled run for new or edited settings, or nil when the destination is disabled or not scheduled.
func (s *BackupService) initialNextRun(in repo.BackupSettings) *time.Time {
	if !in.Enabled || !in.ScheduleEnabled {
		return nil
	}
	t := NextRun(in, time.Now())
	return &t
}

// DeleteDestination removes the destination and its history rows. Backup
// files already written to the destination are left untouched.
func (s *BackupService) DeleteDestination(ctx context.Context, gid, id uuid.UUID) error {
	if _, err := s.repos.BackupDestinations.Get(ctx, gid, id); err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return err
	}
	if _, err := s.repos.Exports.DeleteByDestination(ctx, gid, id); err != nil {
		return err
	}
	if _, err := s.repos.BackupDestinations.Delete(ctx, gid, id); err != nil {
		return err
	}
	s.tokensMu.Lock()
	delete(s.tokens, id)
	s.tokensMu.Unlock()
	s.publishMutation(gid)
	return nil
}

func (s *BackupService) TestSettings(ctx context.Context, gid uuid.UUID, in repo.BackupInput) (TestResult, error) {
	if !s.cfg.Enabled {
		return TestResult{}, ErrBackupDisabled
	}
	st, err := s.normalize(in.BackupSettings, false)
	if err != nil {
		return TestResult{}, err
	}
	d := repo.BackupDestinationOut{GroupID: gid, BackupSettings: st}

	var sec *backupSecret
	if isDriveType(st.Type) {
		switch {
		case in.OAuthTicket != "":
			cred, _, terr := s.ticketCredentials(gid, in.OAuthTicket, st.Type, false)
			if terr != nil {
				return TestResult{}, terr
			}
			sec = &cred
		case in.DestinationID != "":
			id, perr := uuid.Parse(in.DestinationID)
			if perr != nil {
				return TestResult{}, invalid("invalid destination id")
			}
			cur, gerr := s.repos.BackupDestinations.Get(ctx, gid, id)
			if gerr != nil {
				return TestResult{}, gerr
			}
			if cur.Secret == "" || cur.Type != st.Type {
				return TestResult{}, invalid("connect your account first")
			}
			d.ID, d.Secret = cur.ID, cur.Secret
		default:
			return TestResult{}, invalid("connect your account first")
		}
	}
	if isRemoteType(st.Type) {
		switch {
		case in.Password != "" || in.PrivateKey != "":
			sec = &backupSecret{Password: in.Password, PrivateKey: in.PrivateKey, Passphrase: in.Passphrase}
			if passwordOnly(st.Type) {
				sec.PrivateKey, sec.Passphrase = "", ""
			}
		case in.DestinationID != "":
			id, perr := uuid.Parse(in.DestinationID)
			if perr != nil {
				return TestResult{}, invalid("invalid destination id")
			}
			cur, gerr := s.repos.BackupDestinations.Get(ctx, gid, id)
			if gerr != nil {
				return TestResult{}, gerr
			}
			if cur.Secret == "" || cur.Type != st.Type || cur.ConnString != st.ConnString || cur.Username != st.Username {
				return TestResult{}, invalid("re-enter the credentials to test a changed type, address or username")
			}
			d.ID, d.Secret = cur.ID, cur.Secret
		default:
			return TestResult{}, invalid("enter a password or private key to test the connection")
		}
	}
	return s.check(ctx, d, gid, sec), nil
}

// TestDestination checks a saved destination and records the result as its
// health status.
func (s *BackupService) TestDestination(ctx context.Context, gid, id uuid.UUID) (TestResult, error) {
	d, err := s.repos.BackupDestinations.Get(ctx, gid, id)
	if err != nil {
		return TestResult{}, err
	}
	res := s.Check(ctx, d, gid)
	s.recordHealth(ctx, d, res)
	return res, nil
}

// RunNow starts an on-demand backup to a destination.
func (s *BackupService) RunNow(ctx context.Context, gid, id uuid.UUID) (repo.ExportOut, error) {
	if !s.cfg.Enabled {
		return repo.ExportOut{}, ErrBackupDisabled
	}
	d, err := s.repos.BackupDestinations.Get(ctx, gid, id)
	if err != nil {
		return repo.ExportOut{}, err
	}
	if !d.Enabled {
		return repo.ExportOut{}, invalid("destination is disabled")
	}
	s.failInterrupted(ctx, gid, id)
	active, err := s.repos.Exports.HasActiveForDestination(ctx, gid, id)
	if err != nil {
		return repo.ExportOut{}, err
	}
	if active {
		return repo.ExportOut{}, invalid("a backup to this destination is already running")
	}
	return s.exports.EnqueueForDestination(ctx, gid, id, originManual)
}

// failInterrupted fails runs that stopped updating, so a crash mid-backup cannot
// block the destination forever.
func (s *BackupService) failInterrupted(ctx context.Context, gid, id uuid.UUID) {
	n, err := s.repos.Exports.FailStaleActive(ctx, gid, id, time.Now().Add(-activeRunTimeout))
	if err != nil {
		log.Err(err).Stringer("destination_id", id).Msg("backup: failing interrupted runs")
		return
	}
	if n > 0 {
		log.Warn().Stringer("destination_id", id).Int("runs", n).Msg("backup: marked interrupted runs as failed")
		s.publishMutation(gid)
	}
}

// ListVersions returns the backups written to a destination, newest first.
func (s *BackupService) ListVersions(ctx context.Context, gid, id uuid.UUID) ([]repo.ExportOut, error) {
	if _, err := s.repos.BackupDestinations.Get(ctx, gid, id); err != nil {
		return nil, err
	}
	return s.repos.Exports.ListByDestination(ctx, gid, id)
}

// publishMutation tells connected clients of the collection to refresh their backup views.
func (s *BackupService) publishMutation(gid uuid.UUID) {
	if s.bus != nil {
		s.bus.Publish(eventbus.EventExportMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

// ---------------------------------------------------------------------------
// Schedule

// NextRun returns the first scheduled time strictly after `after`, in the
// server's local time zone.
func NextRun(c repo.BackupSettings, after time.Time) time.Time {
	loc := time.Local
	after = after.In(loc)
	y, m, day := after.Date()

	switch c.Frequency {
	case "cron":
		if sched, err := parseCron(c.CronExpr); err == nil {
			return sched.Next(after)
		}
		return after.Add(time.Hour) // unreachable for validated settings
	case "hourly":
		t := time.Date(y, m, day, after.Hour(), c.AtMinute, 0, 0, loc)
		if !t.After(after) {
			t = t.Add(time.Hour)
		}
		if c.IntervalHours > 1 {
			t = t.Add(time.Duration(c.IntervalHours-1) * time.Hour)
		}
		return t
	case "weekly":
		t := time.Date(y, m, day, c.AtHour, c.AtMinute, 0, 0, loc)
		for i := 0; i < 8; i++ {
			if int(t.Weekday()) == c.Weekday && t.After(after) {
				return t
			}
			t = t.AddDate(0, 0, 1)
		}
		return t
	case "monthly":
		t := time.Date(y, m, c.DayOfMonth, c.AtHour, c.AtMinute, 0, 0, loc)
		if !t.After(after) {
			t = time.Date(y, m+1, c.DayOfMonth, c.AtHour, c.AtMinute, 0, 0, loc)
		}
		return t
	default: // daily
		t := time.Date(y, m, day, c.AtHour, c.AtMinute, 0, 0, loc)
		if !t.After(after) {
			t = t.AddDate(0, 0, 1)
		}
		return t
	}
}

// SchedulerTick starts every backup that has come due and raises "no recent
// successful backup" alerts. It is safe to call often; each destination's
// next_run_at is advanced before its backup is enqueued so a slow run is
// never started twice.
func (s *BackupService) SchedulerTick(ctx context.Context) {
	dests, err := s.repos.BackupDestinations.ListScheduled(ctx)
	if err != nil {
		log.Err(err).Msg("backup scheduler: list destinations")
		return
	}
	now := time.Now()
	for _, d := range dests {
		s.checkStale(ctx, d, now)

		if d.NextRunAt == nil {
			if err := s.repos.BackupDestinations.SetNextRun(ctx, d.ID, NextRun(d.BackupSettings, now), nil); err != nil {
				log.Err(err).Stringer("destination_id", d.ID).Msg("backup scheduler: set next run")
			}
			continue
		}
		if d.NextRunAt.After(now) {
			continue
		}

		next := NextRun(d.BackupSettings, now)
		s.failInterrupted(ctx, d.GroupID, d.ID)
		active, err := s.repos.Exports.HasActiveForDestination(ctx, d.GroupID, d.ID)
		if err != nil {
			log.Err(err).Stringer("destination_id", d.ID).Msg("backup scheduler: check active")
			continue
		}
		if active {
			_ = s.repos.BackupDestinations.SetNextRun(ctx, d.ID, next, nil)
			continue
		}

		if d.SkipIfUnchanged && d.LastFingerprint != "" && d.LastSuccessAt != nil {
			fp, err := s.Fingerprint(ctx, d.GroupID)
			if err == nil && fp == d.LastFingerprint {
				_ = s.repos.BackupDestinations.MarkSkipped(ctx, d.ID, now)
				_ = s.repos.BackupDestinations.SetNextRun(ctx, d.ID, next, nil)
				s.publishMutation(d.GroupID)
				continue
			}
		}

		if err := s.repos.BackupDestinations.SetNextRun(ctx, d.ID, next, &now); err != nil {
			log.Err(err).Stringer("destination_id", d.ID).Msg("backup scheduler: advance next run")
			continue
		}
		if _, err := s.exports.EnqueueForDestination(ctx, d.GroupID, d.ID, originScheduled); err != nil {
			log.Err(err).Stringer("destination_id", d.ID).Msg("backup scheduler: enqueue")
			s.recordFailure(ctx, d, "failed to start backup: "+err.Error())
		}
	}
}

// schedulePeriod is the longest normal gap between two scheduled runs, used to
// keep the "no recent backup" alert from firing between runs of a sparse
// schedule.
func schedulePeriod(c repo.BackupSettings) time.Duration {
	switch c.Frequency {
	case "cron":
		if sched, err := parseCron(c.CronExpr); err == nil {
			return cronPeriod(sched)
		}
		return 24 * time.Hour
	case "hourly":
		return time.Duration(max(c.IntervalHours, 1)) * time.Hour
	case "weekly":
		return 7 * 24 * time.Hour
	case "monthly":
		return 31 * 24 * time.Hour
	default: // daily
		return 24 * time.Hour
	}
}

// staleThreshold is how long a destination may go without a successful (or
// legitimately skipped) backup before it alerts: the configured hours, but never
// less than two schedule periods. A weekly or monthly schedule with the default
// 48 hours would otherwise alert on every healthy gap, since one missed run is
// not yet a problem.
func staleThreshold(d repo.BackupDestinationOut) time.Duration {
	return max(time.Duration(d.AlertStaleHours)*time.Hour, 2*schedulePeriod(d.BackupSettings))
}

// checkStale raises the "no recent backup" alert. A run skipped because nothing
// changed counts as fresh: the data is already backed up, so a quiet collection
// must not alert just because it had nothing new to save.
func (s *BackupService) checkStale(ctx context.Context, d repo.BackupDestinationOut, now time.Time) {
	if !d.AlertsEnabled || d.AlertStaleHours <= 0 || d.AlertedStale {
		return
	}
	ref := d.CreatedAt
	if d.LastSuccessAt != nil {
		ref = *d.LastSuccessAt
	}
	if d.LastSkippedAt != nil && d.LastSkippedAt.After(ref) {
		ref = *d.LastSkippedAt
	}
	limit := staleThreshold(d)
	if now.Sub(ref) < limit {
		return
	}
	s.alert(ctx, d, fmt.Sprintf("no successful backup in the last %d hours.", int(limit.Hours())))
	_ = s.repos.BackupDestinations.SetAlertedStale(ctx, d.ID, true)
}

// ---------------------------------------------------------------------------
// Change detection

// fingerprintSpecs lists what a backup captures. Every table with an
// updated_at column contributes its row count and newest timestamp, so edits,
// additions and deletions all change the fingerprint. tag_entities is a bare
// join table, so only its count is used.
type fingerprintSpec struct {
	table      string
	scope      string
	hasUpdated bool
}

// fingerprintSpecs lists the tables, and the rows within them, that make up a backup.
func fingerprintSpecs() []fingerprintSpec {
	specs := make([]fingerprintSpec, 0, len(exportTables)+1)
	specs = append(specs, fingerprintSpec{table: "groups", scope: "id = ?", hasUpdated: true})
	for _, t := range exportTables {
		specs = append(specs, fingerprintSpec{table: t.name, scope: t.scope, hasUpdated: t.name != "tag_entities"})
	}
	return specs
}

// Fingerprint returns a hash that changes whenever the group's backed-up data
// or settings change.
func (s *BackupService) Fingerprint(ctx context.Context, gid uuid.UUID) (string, error) {
	h := sha256.New()
	db := s.db.Sql()
	for _, spec := range fingerprintSpecs() {
		q := "SELECT COUNT(*)"
		if spec.hasUpdated {
			q += ", MAX(updated_at)"
		}
		q += " FROM " + spec.table
		var args []any
		if spec.scope != "" {
			q += " WHERE " + rebindPlaceholders(spec.scope, s.dialect)
			for i := 0; i < strings.Count(spec.scope, "?"); i++ {
				args = append(args, gid.String())
			}
		}
		var (
			count int64
			maxTS any
		)
		var err error
		if spec.hasUpdated {
			err = db.QueryRowContext(ctx, q, args...).Scan(&count, &maxTS)
		} else {
			err = db.QueryRowContext(ctx, q, args...).Scan(&count)
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("fingerprint %s: %w", spec.table, err)
		}
		_, _ = fmt.Fprintf(h, "%s|%d|%v\n", spec.table, count, normalizeScan(maxTS))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ---------------------------------------------------------------------------
// Run bookkeeping (called by ExportService)

// beforeRun captures the data fingerprint a destination backup starts from.
func (s *BackupService) beforeRun(ctx context.Context, gid uuid.UUID) string {
	fp, err := s.Fingerprint(ctx, gid)
	if err != nil {
		log.Warn().Err(err).Msg("backup: fingerprint failed; skip-if-unchanged will not apply to this run")
		return ""
	}
	return fp
}

// afterRun records the outcome of a destination-bound export, prunes old
// versions on success and raises or clears alerts.
func (s *BackupService) afterRun(ctx context.Context, gid uuid.UUID, exp repo.ExportOut, fingerprint string, runErr error) {
	if exp.DestinationID == nil {
		return
	}
	d, err := s.repos.BackupDestinations.Get(ctx, gid, *exp.DestinationID)
	if err != nil {
		return // destination deleted mid-run
	}
	if runErr != nil {
		s.recordFailure(ctx, d, runErr.Error())
		return
	}

	now := time.Now()
	if err := s.repos.BackupDestinations.MarkRunSucceeded(ctx, d.ID, now); err != nil {
		log.Err(err).Msg("backup: record success")
	}
	if fingerprint != "" {
		_ = s.repos.BackupDestinations.SetFingerprint(ctx, d.ID, fingerprint)
	}
	if d.AlertedFailure && d.AlertsEnabled {
		s.alert(ctx, d, "backups are working again.")
	}
	// MarkRunSucceeded already cleared the failure and stale flags.
	s.Prune(ctx, d)
	s.publishMutation(gid)
}

// recordFailure stores a failed run and sends the failure alert once per streak.
func (s *BackupService) recordFailure(ctx context.Context, d repo.BackupDestinationOut, msg string) {
	msg = s.redact(msg)
	if err := s.repos.BackupDestinations.MarkRunFailed(ctx, d.ID, time.Now(), msg); err != nil {
		log.Err(err).Msg("backup: record failure")
	}
	if d.AlertsEnabled && !d.AlertedFailure {
		s.alert(ctx, d, "backup failed: "+msg)
		_ = s.repos.BackupDestinations.SetAlertedFailure(ctx, d.ID, true)
	}
	s.publishMutation(d.GroupID)
}

// ---------------------------------------------------------------------------
// Retention

// SelectPrunable returns the completed backups a destination's retention
// policy no longer keeps. The newest backup is always kept. For each of the
// daily, weekly and monthly policies the newest backup in each of the most
// recent N calendar buckets is kept; a policy of 0 is unused. When all three
// are 0 nothing is pruned.
func SelectPrunable(rows []repo.ExportOut, daily, weekly, monthly int) []repo.ExportOut {
	if daily <= 0 && weekly <= 0 && monthly <= 0 {
		return nil
	}
	completed := make([]repo.ExportOut, 0, len(rows))
	for _, r := range rows {
		if r.Status == "completed" {
			completed = append(completed, r)
		}
	}
	sort.Slice(completed, func(i, j int) bool { return completed[i].CreatedAt.After(completed[j].CreatedAt) })

	keep := make(map[uuid.UUID]bool, len(completed))
	if len(completed) > 0 {
		keep[completed[0].ID] = true
	}
	mark := func(limit int, bucket func(time.Time) string) {
		if limit <= 0 {
			return
		}
		seen := map[string]bool{}
		for _, r := range completed {
			b := bucket(r.CreatedAt.In(time.Local))
			if seen[b] {
				continue
			}
			if len(seen) >= limit {
				break
			}
			seen[b] = true
			keep[r.ID] = true
		}
	}
	mark(daily, func(t time.Time) string { return t.Format("2006-01-02") })
	mark(weekly, func(t time.Time) string {
		y, w := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", y, w)
	})
	mark(monthly, func(t time.Time) string { return t.Format("2006-01") })

	var out []repo.ExportOut
	for _, r := range completed {
		if !keep[r.ID] {
			out = append(out, r)
		}
	}
	return out
}

// Prune applies the destination's retention policy, deleting each pruned
// artifact before its row, and drops long-stale failed rows.
func (s *BackupService) Prune(ctx context.Context, d repo.BackupDestinationOut) {
	rows, err := s.repos.Exports.ListByDestination(ctx, d.GroupID, d.ID)
	if err != nil {
		log.Err(err).Msg("backup prune: list")
		return
	}

	victims := SelectPrunable(rows, d.KeepDaily, d.KeepWeekly, d.KeepMonthly)
	cutoff := time.Now().Add(-failedRowRetention)
	for _, r := range rows {
		if r.Status == "failed" && r.CreatedAt.Before(cutoff) {
			victims = append(victims, r)
		}
	}
	for _, v := range victims {
		if err := s.DeleteVersion(ctx, d.GroupID, v); err != nil {
			log.Warn().Err(err).Stringer("export_id", v.ID).Msg("backup prune: delete failed; leaving for next run")
		}
	}
}

// DeleteVersion removes a backup's artifact and then its row. A missing
// artifact counts as already deleted.
func (s *BackupService) DeleteVersion(ctx context.Context, gid uuid.UUID, exp repo.ExportOut) error {
	if exp.ArtifactPath != "" {
		st, key, err := s.ArtifactLocation(ctx, gid, exp)
		if err != nil {
			return err
		}
		err = st.Delete(ctx, key)
		_ = st.Close()
		if err != nil {
			return err
		}
	}
	_, err := s.repos.Exports.Delete(ctx, gid, exp.ID)
	return err
}

// ---------------------------------------------------------------------------
// Health

// HealthTick probes every enabled destination whose check interval elapsed.
func (s *BackupService) HealthTick(ctx context.Context) {
	dests, err := s.repos.BackupDestinations.ListEnabled(ctx)
	if err != nil {
		log.Err(err).Msg("backup health: list destinations")
		return
	}
	now := time.Now()
	for _, d := range dests {
		if d.HealthCheckedAt != nil && now.Sub(*d.HealthCheckedAt) < time.Duration(d.HealthIntervalMinutes)*time.Minute {
			continue
		}
		s.recordHealth(ctx, d, s.Check(ctx, d, d.GroupID))
	}
}

// recordHealth stores a probe result and raises or clears the unreachable
// alert. An outage alerts once, after AlertFailureThreshold consecutive
// failed probes, and a recovery notice follows when the destination returns.
func (s *BackupService) recordHealth(ctx context.Context, d repo.BackupDestinationOut, res TestResult) {
	now := time.Now()
	if res.OK {
		if err := s.repos.BackupDestinations.SetHealth(ctx, d.ID, healthHealthy, now, "", 0); err != nil {
			log.Err(err).Msg("backup health: record")
			return
		}
		if d.AlertedUnreachable {
			if d.AlertsEnabled {
				s.alert(ctx, d, "destination is reachable again.")
			}
			_ = s.repos.BackupDestinations.SetAlertedUnreachable(ctx, d.ID, false)
		}
		s.publishMutation(d.GroupID)
		return
	}

	failures := d.HealthFailures + 1
	if err := s.repos.BackupDestinations.SetHealth(ctx, d.ID, healthUnreachable, now, res.Message, failures); err != nil {
		log.Err(err).Msg("backup health: record")
		return
	}
	if d.AlertsEnabled && !d.AlertedUnreachable && failures >= d.AlertFailureThreshold {
		s.alert(ctx, d, fmt.Sprintf("destination is unreachable after %d failed checks: %s", failures, res.Message))
		_ = s.repos.BackupDestinations.SetAlertedUnreachable(ctx, d.ID, true)
	}
	s.publishMutation(d.GroupID)
}

// ---------------------------------------------------------------------------
// Alerts

// alert sends msg to the group's active notifiers. The in-app banner is
// driven by the persisted destination state, not by this call.
func (s *BackupService) alert(ctx context.Context, d repo.BackupDestinationOut, msg string) {
	log.Warn().Stringer("destination_id", d.ID).Str("destination", d.Name).Msg("backup alert: " + msg)

	notifiers, err := s.repos.Notifiers.GetActiveByGroup(ctx, d.GroupID)
	if err != nil {
		log.Err(err).Msg("backup alert: list notifiers")
		return
	}
	text := fmt.Sprintf("Homebox backup alert for %q: %s", d.Name, msg)
	for i := range notifiers {
		if err := validate.SendNotifierMessage(notifiers[i].URL, text, s.notifierConfig); err != nil {
			log.Err(err).Str("notifier", notifiers[i].Name).Msg("backup alert: send failed")
		}
	}
}

// AuthorizeExportAccess refuses access to a backup that belongs to a destination
// unless the caller owns the collection. Destinations are managed by owners, so
// the backups they hold are owner-only too; ordinary manual exports stay
// available to every member. Without this, the generic export endpoints would
// expose (and let members delete) the backups the owner-only routes protect.
func (s *BackupService) AuthorizeExportAccess(ctx Context, exp repo.ExportOut) error {
	if exp.DestinationID == nil {
		return nil
	}
	isOwner, err := s.repos.Groups.IsOwnerOf(ctx.Context, ctx.UID, ctx.GID)
	if err != nil {
		return err
	}
	if !isOwner {
		return validate.NewRequestError(ErrNotGroupOwner, http.StatusForbidden)
	}
	return nil
}

// VisibleExports drops destination-bound backups from a listing unless the
// caller owns the collection.
func (s *BackupService) VisibleExports(ctx Context, rows []repo.ExportOut) ([]repo.ExportOut, error) {
	isOwner, err := s.repos.Groups.IsOwnerOf(ctx.Context, ctx.UID, ctx.GID)
	if err != nil {
		return nil, err
	}
	if isOwner {
		return rows, nil
	}
	out := make([]repo.ExportOut, 0, len(rows))
	for _, r := range rows {
		if r.DestinationID == nil {
			out = append(out, r)
		}
	}
	return out, nil
}
