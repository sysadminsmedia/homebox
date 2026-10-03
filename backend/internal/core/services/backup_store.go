package services

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/hirochachacha/go-smb2"
	"github.com/pkg/sftp"
	"github.com/studio-b12/gowebdav"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"
	"golang.org/x/crypto/ssh"
)

// x509IncorrectPassword is what x/crypto returns for a wrong key passphrase.
var x509IncorrectPassword = x509.IncorrectPasswordError

// objectStore is the small slice of storage the backup service needs. The
// gocloud buckets, SFTP and WebDAV all sit behind it.
type objectStore interface {
	// Write stores r under key. size is the byte count when known, or -1.
	Write(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete removes key. A missing key is not an error.
	Delete(ctx context.Context, key string) error
	Close() error
}

const (
	dialTimeout = 15 * time.Second
	partSuffix  = ".part"
)

// ---------------------------------------------------------------------------
// gocloud buckets

type blobStore struct{ b *blob.Bucket }

func (s blobStore) Write(ctx context.Context, key string, r io.Reader, _ int64, contentType string) error {
	w, err := s.b.NewWriter(ctx, key, &blob.WriterOptions{ContentType: contentType})
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func (s blobStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.b.NewReader(ctx, key, nil)
}

func (s blobStore) Delete(ctx context.Context, key string) error {
	err := s.b.Delete(ctx, key)
	if err != nil && gcerrors.Code(err) == gcerrors.NotFound {
		return nil
	}
	return err
}

func (s blobStore) Close() error { return s.b.Close() }

// ---------------------------------------------------------------------------
// SFTP

// hostKeyError reports a host key that is missing from, or does not match, the
// destination's settings. Fingerprint is what the server presented, so the UI
// can offer it for confirmation.
type hostKeyError struct {
	Fingerprint string
	Expected    string
}

func (e *hostKeyError) Error() string {
	if e.Expected == "" {
		return "the server's SSH host key has not been verified yet (" + e.Fingerprint + ")"
	}
	return fmt.Sprintf("the server's SSH host key changed: expected %s, got %s", e.Expected, e.Fingerprint)
}

// sftpTarget is a parsed sftp:// destination.
type sftpTarget struct {
	addr string // host:port
	base string // absolute directory on the server
}

func parseSFTPURL(raw string) (sftpTarget, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "sftp" || u.Hostname() == "" {
		return sftpTarget{}, errors.New("expected sftp://host[:port]/path")
	}
	if u.User != nil {
		return sftpTarget{}, errors.New("enter the username in its own field, not in the URL")
	}
	port := u.Port()
	if port == "" {
		port = "22"
	}
	base := path.Clean("/" + strings.TrimPrefix(u.Path, "/"))
	return sftpTarget{addr: net.JoinHostPort(u.Hostname(), port), base: base}, nil
}

// parsePrivateKey reads a PEM private key, unlocking it with passphrase when it
// is encrypted. The errors are written for the user and never echo the key.
func parsePrivateKey(pemKey, passphrase string) (ssh.Signer, error) {
	key := []byte(pemKey)
	if passphrase != "" {
		signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
		if err == nil {
			return signer, nil
		}
		if errors.Is(err, x509IncorrectPassword) {
			return nil, errors.New("the private key's passphrase is wrong")
		}
		// An unencrypted key given a passphrase is still usable.
		if signer, perr := ssh.ParsePrivateKey(key); perr == nil {
			return signer, nil
		}
		return nil, errors.New("the private key could not be read (check the key and its passphrase)")
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		var missing *ssh.PassphraseMissingError
		if errors.As(err, &missing) {
			return nil, errors.New("the private key is passphrase-protected; enter its passphrase")
		}
		return nil, errors.New("the private key could not be read (it must be a PEM private key)")
	}
	return signer, nil
}

type sftpStore struct {
	ssh    *ssh.Client
	client *sftp.Client
	base   string
}

// dialSFTP connects and authenticates. expectedHostKey is the trusted
// fingerprint; when empty the connection is refused with a hostKeyError that
// carries the presented fingerprint.
func dialSFTP(t sftpTarget, user string, sec backupSecret, expectedHostKey string) (*sftpStore, error) {
	var methods []ssh.AuthMethod
	if sec.PrivateKey != "" {
		signer, err := parsePrivateKey(sec.PrivateKey, sec.Passphrase)
		if err != nil {
			return nil, err
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if sec.Password != "" {
		methods = append(methods, ssh.Password(sec.Password))
	}
	if len(methods) == 0 {
		return nil, errors.New("no password or private key is set")
	}

	var hkErr *hostKeyError
	cfg := &ssh.ClientConfig{
		User: user,
		Auth: methods,
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			fp := ssh.FingerprintSHA256(key)
			if expectedHostKey == "" || fp != expectedHostKey {
				hkErr = &hostKeyError{Fingerprint: fp, Expected: expectedHostKey}
				return hkErr
			}
			return nil
		},
		Timeout: dialTimeout,
	}
	conn, err := ssh.Dial("tcp", t.addr, cfg)
	if err != nil {
		if hkErr != nil {
			return nil, hkErr
		}
		return nil, fmt.Errorf("ssh connect: %w", err)
	}
	client, err := sftp.NewClient(conn)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("sftp session: %w", err)
	}
	return &sftpStore{ssh: conn, client: client, base: t.base}, nil
}

func (s *sftpStore) full(key string) (string, error) {
	p := path.Clean(path.Join(s.base, key))
	if s.base != "/" && p != s.base && !strings.HasPrefix(p, s.base+"/") {
		return "", errors.New("path escapes the destination directory")
	}
	return p, nil
}

func (s *sftpStore) Write(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	p, err := s.full(key)
	if err != nil {
		return err
	}
	if err := s.client.MkdirAll(path.Dir(p)); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	tmp := p + partSuffix
	f, err := s.client.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("open remote file: %w", err)
	}
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		_ = s.client.Remove(tmp)
		return fmt.Errorf("upload: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = s.client.Remove(tmp)
		return fmt.Errorf("finish upload: %w", err)
	}
	if err := s.client.PosixRename(tmp, p); err != nil {
		// Servers without the posix-rename extension refuse to overwrite.
		_ = s.client.Remove(p)
		if err := s.client.Rename(tmp, p); err != nil {
			_ = s.client.Remove(tmp)
			return fmt.Errorf("finalize upload: %w", err)
		}
	}
	return nil
}

func (s *sftpStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	p, err := s.full(key)
	if err != nil {
		return nil, err
	}
	return s.client.Open(p)
}

func (s *sftpStore) Delete(_ context.Context, key string) error {
	p, err := s.full(key)
	if err != nil {
		return err
	}
	if err := s.client.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (s *sftpStore) Close() error {
	cerr := s.client.Close()
	return errors.Join(cerr, s.ssh.Close())
}

// ---------------------------------------------------------------------------
// WebDAV

func parseWebDAVURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, errors.New("expected http(s)://host/path")
	}
	if u.User != nil {
		return nil, errors.New("enter the username in its own field, not in the URL")
	}
	return u, nil
}

type davStore struct {
	c *gowebdav.Client

	mu     sync.Mutex
	putLen int64
}

func newDAVStore(raw, user, password string) (*davStore, error) {
	u, err := parseWebDAVURL(raw)
	if err != nil {
		return nil, err
	}
	s := &davStore{c: gowebdav.NewClient(u.String(), user, password)}
	s.c.SetTransport(&http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: dialTimeout}).DialContext,
		TLSHandshakeTimeout:   dialTimeout,
		ResponseHeaderTimeout: 2 * time.Minute,
		IdleConnTimeout:       30 * time.Second,
	})
	// Many servers (Synology, nginx) reject chunked PUT bodies, so declare the
	// length when it is known.
	s.c.SetInterceptor(func(method string, rq *http.Request) {
		if method != http.MethodPut {
			return
		}
		s.mu.Lock()
		n := s.putLen
		s.mu.Unlock()
		if n > 0 {
			rq.ContentLength = n
		}
	})
	return s, nil
}

func (s *davStore) Write(_ context.Context, key string, r io.Reader, size int64, _ string) error {
	if err := s.c.MkdirAll(path.Dir(key), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	tmp := key + partSuffix
	s.mu.Lock()
	s.putLen = size
	s.mu.Unlock()
	if err := s.c.WriteStream(tmp, r, 0o644); err != nil {
		_ = s.c.Remove(tmp)
		return fmt.Errorf("upload: %w", err)
	}
	if err := s.c.Rename(tmp, key, true); err != nil {
		_ = s.c.Remove(tmp)
		return fmt.Errorf("finalize upload: %w", err)
	}
	return nil
}

func (s *davStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	return s.c.ReadStream(key)
}

func (s *davStore) Delete(_ context.Context, key string) error {
	if err := s.c.Remove(key); err != nil && !gowebdav.IsErrNotFound(err) {
		return err
	}
	return nil
}

func (s *davStore) Close() error { return nil }

// probeBytes is the 1 KB payload health checks write.
func probeBytes() *bytes.Reader { return bytes.NewReader(make([]byte, 1024)) }

// ---------------------------------------------------------------------------
// SMB

// smbTarget is a parsed smb://host[:port]/share/path destination.
type smbTarget struct {
	addr  string // host:port
	share string
	base  string // directory inside the share, "" for the share root
}

func parseSMBURL(raw string) (smbTarget, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "smb" || u.Hostname() == "" {
		return smbTarget{}, errors.New("expected smb://host[:port]/share[/folder]")
	}
	if u.User != nil {
		return smbTarget{}, errors.New("enter the username in its own field, not in the URL")
	}
	port := u.Port()
	if port == "" {
		port = "445"
	}
	segs := strings.Split(strings.Trim(path.Clean("/"+u.Path), "/"), "/")
	if len(segs) == 0 || segs[0] == "" {
		return smbTarget{}, errors.New("name the share: smb://host/share")
	}
	return smbTarget{addr: net.JoinHostPort(u.Hostname(), port), share: segs[0], base: strings.Join(segs[1:], "/")}, nil
}

type smbStore struct {
	conn    net.Conn
	session *smb2.Session
	share   *smb2.Share
	base    string
}

// dialSMB connects with NTLM. user may be "name", "DOMAIN\\name" or an email-style
// "name@domain" (sent as written, which servers accept as a UPN).
func dialSMB(t smbTarget, user, password string) (*smbStore, error) {
	domain := ""
	if i := strings.IndexAny(user, `\/`); i > 0 {
		domain, user = user[:i], user[i+1:]
	}
	conn, err := net.DialTimeout("tcp", t.addr, dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("smb connect: %w", err)
	}
	_ = conn.SetDeadline(time.Now().Add(2 * dialTimeout)) // bound the handshake only
	d := &smb2.Dialer{Initiator: &smb2.NTLMInitiator{User: user, Password: password, Domain: domain}}
	sess, err := d.Dial(conn)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("smb login: %w", err)
	}
	share, err := sess.Mount(t.share)
	if err != nil {
		_ = sess.Logoff()
		_ = conn.Close()
		return nil, fmt.Errorf("smb share %q: %w", t.share, err)
	}
	_ = conn.SetDeadline(time.Time{})
	return &smbStore{conn: conn, session: sess, share: share, base: t.base}, nil
}

func (s *smbStore) full(key string) (string, error) {
	p := strings.TrimPrefix(path.Clean(path.Join("/", s.base, key)), "/")
	b := strings.Trim(s.base, "/")
	if b != "" && p != b && !strings.HasPrefix(p, b+"/") {
		return "", errors.New("path escapes the destination folder")
	}
	if p == "" {
		return "", errors.New("empty path")
	}
	return p, nil
}

func (s *smbStore) Write(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	p, err := s.full(key)
	if err != nil {
		return err
	}
	if dir := path.Dir(p); dir != "." {
		if err := s.share.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create folder: %w", err)
		}
	}
	tmp := p + partSuffix
	f, err := s.share.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open remote file: %w", err)
	}
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		_ = s.share.Remove(tmp)
		return fmt.Errorf("upload: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = s.share.Remove(tmp)
		return fmt.Errorf("finish upload: %w", err)
	}
	// SMB rename does not replace an existing file.
	if err := s.share.Remove(p); err != nil && !os.IsNotExist(err) {
		_ = s.share.Remove(tmp)
		return fmt.Errorf("replace existing file: %w", err)
	}
	if err := s.share.Rename(tmp, p); err != nil {
		_ = s.share.Remove(tmp)
		return fmt.Errorf("finalize upload: %w", err)
	}
	return nil
}

func (s *smbStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	p, err := s.full(key)
	if err != nil {
		return nil, err
	}
	return s.share.Open(p)
}

func (s *smbStore) Delete(_ context.Context, key string) error {
	p, err := s.full(key)
	if err != nil {
		return err
	}
	if err := s.share.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *smbStore) Close() error {
	_ = s.share.Umount()
	_ = s.session.Logoff()
	return s.conn.Close()
}
