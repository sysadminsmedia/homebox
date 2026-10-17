package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

// longHTTPClient has no overall timeout, since backups can be hundreds of
// megabytes, but bounds connection setup and the wait for response headers.
func longHTTPClient() *http.Client {
	return &http.Client{Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: dialTimeout}).DialContext,
		TLSHandshakeTimeout:   dialTimeout,
		ResponseHeaderTimeout: 3 * time.Minute,
		IdleConnTimeout:       30 * time.Second,
	}}
}

// httpClient is the client used for provider calls.
func (s *BackupService) httpClient() *http.Client {
	if s.httpc != nil {
		return s.httpc
	}
	return longHTTPClient()
}

// driveAPI sends authenticated requests, refreshing the token once on a 401 and
// backing off on 429/503. Request bodies are rebuilt by the caller for every
// attempt, so uploads can be replayed.
type driveAPI struct {
	ts *tokenSource
	hc *http.Client
}

func (a *driveAPI) do(ctx context.Context, build func() (*http.Request, error)) (*http.Response, error) {
	var resp *http.Response
	for attempt := 0; attempt < 4; attempt++ {
		tok, err := a.ts.token(ctx)
		if err != nil {
			return nil, err
		}
		req, err := build()
		if err != nil {
			return nil, err
		}
		req = req.WithContext(ctx)
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, err = a.hc.Do(req)
		if err != nil {
			return nil, err
		}
		switch {
		case resp.StatusCode == http.StatusUnauthorized && attempt == 0:
			drain(resp)
			a.ts.invalidate()
		case (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable) && attempt < 3:
			wait := time.Duration(attempt+1) * 500 * time.Millisecond
			if ra, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && ra > 0 && ra <= 10 {
				wait = time.Duration(ra) * time.Second
			}
			drain(resp)
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		default:
			return resp, nil
		}
	}
	return resp, nil
}

func drain(resp *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	_ = resp.Body.Close()
}

// apiError reads a short, sanitized description of a failed response.
func apiError(op string, resp *http.Response) error {
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("%s: %d %s", op, resp.StatusCode, sanitizeProviderText(string(b)))
}

func okStatus(resp *http.Response, codes ...int) bool {
	for _, c := range codes {
		if resp.StatusCode == c {
			return true
		}
	}
	return false
}

// readerAt makes r randomly readable so an upload can be replayed or sent in
// chunks. Files and byte readers already are; anything else is buffered.
func readerAt(r io.Reader, size int64) (io.ReaderAt, int64, error) {
	if ra, ok := r.(io.ReaderAt); ok && size >= 0 {
		return ra, size, nil
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, 0, err
	}
	return bytes.NewReader(b), int64(len(b)), nil
}

func notFound(what string) error { return fmt.Errorf("%s: %w", what, fs.ErrNotExist) }

func jsonBody(v any) (io.Reader, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// accountName looks up a human-readable name for the connected account.
func (s *BackupService) accountName(ctx context.Context, p oauthProvider, accessToken string) (string, error) {
	hc := s.httpClient()
	get := func(method, u string, body io.Reader, out any) error {
		req, err := http.NewRequestWithContext(ctx, method, u, body)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := hc.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("account lookup: %d", resp.StatusCode)
		}
		return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
	}

	switch p.Key {
	case providerGoogle:
		var out struct {
			User struct {
				Email string `json:"emailAddress"`
				Name  string `json:"displayName"`
			} `json:"user"`
		}
		if err := get(http.MethodGet, p.EP.API+"/drive/v3/about?fields=user(emailAddress,displayName)", nil, &out); err != nil {
			return "", err
		}
		return firstNonEmpty(out.User.Email, out.User.Name), nil
	case providerMicrosoft:
		var out struct {
			UPN  string `json:"userPrincipalName"`
			Mail string `json:"mail"`
		}
		if err := get(http.MethodGet, p.EP.API+"/me?$select=userPrincipalName,mail", nil, &out); err != nil {
			return "", err
		}
		return firstNonEmpty(out.Mail, out.UPN), nil
	default: // dropbox
		var out struct {
			Email string `json:"email"`
			Name  struct {
				Display string `json:"display_name"`
			} `json:"name"`
		}
		if err := get(http.MethodPost, p.EP.API+"/2/users/get_current_account", nil, &out); err != nil {
			return "", err
		}
		return firstNonEmpty(out.Email, out.Name.Display), nil
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Google Drive

type googleStore struct {
	api  *driveAPI
	base string // API host
	up   string // upload host

	mu      sync.Mutex
	folders map[string]string // directory path -> folder ID
}

func newGoogleStore(api *driveAPI, ep providerEndpoints) *googleStore {
	return &googleStore{api: api, base: ep.API, up: firstNonEmpty(ep.Upload, ep.API), folders: map[string]string{}}
}

var driveQueryEscaper = strings.NewReplacer(`\`, `\\`, `'`, `\'`)

func (g *googleStore) findChild(ctx context.Context, parent, name string, folder bool) (string, error) {
	op := "!="
	if folder {
		op = "="
	}
	q := fmt.Sprintf("name = '%s' and '%s' in parents and trashed = false and mimeType %s 'application/vnd.google-apps.folder'",
		driveQueryEscaper.Replace(name), parent, op)
	u := g.base + "/drive/v3/files?" + url.Values{"q": {q}, "fields": {"files(id)"}, "pageSize": {"1"}, "spaces": {"drive"}}.Encode()
	resp, err := g.api.do(ctx, func() (*http.Request, error) { return http.NewRequest(http.MethodGet, u, nil) })
	if err != nil {
		return "", err
	}
	if !okStatus(resp, http.StatusOK) {
		return "", apiError("drive lookup", resp)
	}
	defer func() { _ = resp.Body.Close() }()
	var out struct {
		Files []struct {
			ID string `json:"id"`
		} `json:"files"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Files) == 0 {
		return "", nil
	}
	return out.Files[0].ID, nil
}

func (g *googleStore) createFolder(ctx context.Context, parent, name string) (string, error) {
	resp, err := g.api.do(ctx, func() (*http.Request, error) {
		body, err := jsonBody(map[string]any{"name": name, "mimeType": "application/vnd.google-apps.folder", "parents": []string{parent}})
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodPost, g.base+"/drive/v3/files?fields=id", body)
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return req, err
	})
	if err != nil {
		return "", err
	}
	if !okStatus(resp, http.StatusOK, http.StatusCreated) {
		return "", apiError("create folder", resp)
	}
	defer func() { _ = resp.Body.Close() }()
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&out); err != nil || out.ID == "" {
		return "", errors.New("create folder: unexpected response")
	}
	return out.ID, nil
}

// folderID resolves (and optionally creates) a directory path under the root.
func (g *googleStore) folderID(ctx context.Context, dir string, create bool) (string, error) {
	dir = strings.Trim(path.Clean("/"+dir), "/")
	if dir == "" {
		return "root", nil
	}
	parent, built := "root", ""
	for _, seg := range strings.Split(dir, "/") {
		built += "/" + seg
		g.mu.Lock()
		id, ok := g.folders[built]
		g.mu.Unlock()
		if !ok {
			var err error
			if id, err = g.findChild(ctx, parent, seg, true); err != nil {
				return "", err
			}
			if id == "" {
				if !create {
					return "", nil
				}
				if id, err = g.createFolder(ctx, parent, seg); err != nil {
					return "", err
				}
			}
			g.mu.Lock()
			g.folders[built] = id
			g.mu.Unlock()
		}
		parent = id
	}
	return parent, nil
}

func (g *googleStore) Write(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	ra, total, err := readerAt(r, size)
	if err != nil {
		return err
	}
	dir, name := path.Split(key)
	parent, err := g.folderID(ctx, dir, true)
	if err != nil {
		return err
	}

	// Start a resumable session, then send the file in one request. The file
	// only appears in the folder once the upload completes.
	resp, err := g.api.do(ctx, func() (*http.Request, error) {
		body, err := jsonBody(map[string]any{"name": name, "parents": []string{parent}})
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodPost, g.up+"/upload/drive/v3/files?uploadType=resumable&fields=id", body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Set("X-Upload-Content-Type", contentType)
		req.Header.Set("X-Upload-Content-Length", strconv.FormatInt(total, 10))
		return req, nil
	})
	if err != nil {
		return err
	}
	if !okStatus(resp, http.StatusOK) {
		return apiError("start upload", resp)
	}
	session := resp.Header.Get("Location")
	drain(resp)
	if session == "" {
		return errors.New("start upload: no session URL returned")
	}

	put, err := g.api.do(ctx, func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodPut, session, io.NewSectionReader(ra, 0, total))
		if err != nil {
			return nil, err
		}
		req.ContentLength = total
		req.Header.Set("Content-Type", contentType)
		return req, nil
	})
	if err != nil {
		return err
	}
	if !okStatus(put, http.StatusOK, http.StatusCreated) {
		return apiError("upload", put)
	}
	drain(put)
	return nil
}

func (g *googleStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	dir, name := path.Split(key)
	parent, err := g.folderID(ctx, dir, false)
	if err != nil {
		return nil, err
	}
	if parent == "" {
		return nil, notFound(key)
	}
	id, err := g.findChild(ctx, parent, name, false)
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, notFound(key)
	}
	u := g.base + "/drive/v3/files/" + url.PathEscape(id) + "?alt=media"
	resp, err := g.api.do(ctx, func() (*http.Request, error) { return http.NewRequest(http.MethodGet, u, nil) })
	if err != nil {
		return nil, err
	}
	if !okStatus(resp, http.StatusOK) {
		return nil, apiError("download", resp)
	}
	return resp.Body, nil
}

func (g *googleStore) Delete(ctx context.Context, key string) error {
	dir, name := path.Split(key)
	parent, err := g.folderID(ctx, dir, false)
	if err != nil || parent == "" {
		return err
	}
	id, err := g.findChild(ctx, parent, name, false)
	if err != nil || id == "" {
		return err
	}
	u := g.base + "/drive/v3/files/" + url.PathEscape(id)
	resp, err := g.api.do(ctx, func() (*http.Request, error) { return http.NewRequest(http.MethodDelete, u, nil) })
	if err != nil {
		return err
	}
	if !okStatus(resp, http.StatusNoContent, http.StatusOK, http.StatusNotFound) {
		return apiError("delete", resp)
	}
	drain(resp)
	return nil
}

func (g *googleStore) Close() error { return nil }

// ---------------------------------------------------------------------------
// OneDrive (Microsoft Graph, app folder)

// Upload thresholds. They are variables so tests can exercise the chunked
// paths with small files.
var (
	oneDriveSimpleMax int64 = 4 << 20         // simple PUT limit
	oneDriveChunk     int64 = 32 * 320 * 1024 // session chunks must be multiples of 320 KiB
)

const oneDriveChunkTries = 3

type onedriveStore struct {
	api  *driveAPI
	base string
	hc   *http.Client // upload session URLs are pre-authenticated and must not receive our token
}

func newOneDriveStore(api *driveAPI, ep providerEndpoints, hc *http.Client) *onedriveStore {
	return &onedriveStore{api: api, base: ep.API, hc: hc}
}

func (o *onedriveStore) item(key, suffix string) string {
	segs := strings.Split(strings.Trim(path.Clean("/"+key), "/"), "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	return o.base + "/me/drive/special/approot:/" + strings.Join(segs, "/") + ":" + suffix
}

func (o *onedriveStore) Write(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	ra, total, err := readerAt(r, size)
	if err != nil {
		return err
	}

	if total <= oneDriveSimpleMax {
		u := o.item(key, "/content") + "?@microsoft.graph.conflictBehavior=replace"
		resp, err := o.api.do(ctx, func() (*http.Request, error) {
			req, err := http.NewRequest(http.MethodPut, u, io.NewSectionReader(ra, 0, total))
			if err != nil {
				return nil, err
			}
			req.ContentLength = total
			req.Header.Set("Content-Type", contentType)
			return req, nil
		})
		if err != nil {
			return err
		}
		if !okStatus(resp, http.StatusOK, http.StatusCreated) {
			return apiError("upload", resp)
		}
		drain(resp)
		return nil
	}

	resp, err := o.api.do(ctx, func() (*http.Request, error) {
		body, err := jsonBody(map[string]any{"item": map[string]any{"@microsoft.graph.conflictBehavior": "replace"}})
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodPost, o.item(key, "/createUploadSession"), body)
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return req, err
	})
	if err != nil {
		return err
	}
	if !okStatus(resp, http.StatusOK) {
		return apiError("start upload", resp)
	}
	var sess struct {
		UploadURL string `json:"uploadUrl"`
	}
	err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&sess)
	_ = resp.Body.Close()
	if err != nil || sess.UploadURL == "" {
		return errors.New("start upload: no upload URL returned")
	}

	for off := int64(0); off < total; off += oneDriveChunk {
		n := min(oneDriveChunk, total-off)
		var last error
		for try := 0; try < oneDriveChunkTries; try++ {
			req, err := http.NewRequestWithContext(ctx, http.MethodPut, sess.UploadURL, io.NewSectionReader(ra, off, n))
			if err != nil {
				return err
			}
			req.ContentLength = n
			req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", off, off+n-1, total))
			cresp, err := o.hc.Do(req)
			if err != nil {
				last = err
				continue
			}
			if okStatus(cresp, http.StatusAccepted, http.StatusOK, http.StatusCreated) {
				drain(cresp)
				last = nil
				break
			}
			last = apiError("upload chunk", cresp)
		}
		if last != nil {
			return last
		}
	}
	return nil
}

func (o *onedriveStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	u := o.item(key, "/content")
	resp, err := o.api.do(ctx, func() (*http.Request, error) { return http.NewRequest(http.MethodGet, u, nil) })
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		drain(resp)
		return nil, notFound(key)
	}
	if !okStatus(resp, http.StatusOK) {
		return nil, apiError("download", resp)
	}
	return resp.Body, nil
}

func (o *onedriveStore) Delete(ctx context.Context, key string) error {
	u := o.item(key, "")
	resp, err := o.api.do(ctx, func() (*http.Request, error) { return http.NewRequest(http.MethodDelete, u, nil) })
	if err != nil {
		return err
	}
	if !okStatus(resp, http.StatusNoContent, http.StatusOK, http.StatusNotFound) {
		return apiError("delete", resp)
	}
	drain(resp)
	return nil
}

func (o *onedriveStore) Close() error { return nil }

// ---------------------------------------------------------------------------
// Dropbox

var (
	dropboxSimpleMax int64 = 100 << 20
	dropboxChunk     int64 = 64 << 20
)

type dropboxStore struct {
	api     *driveAPI
	apiBase string
	content string
}

func newDropboxStore(api *driveAPI, ep providerEndpoints) *dropboxStore {
	return &dropboxStore{api: api, apiBase: ep.API, content: firstNonEmpty(ep.Upload, ep.API)}
}

// dropboxArg encodes the Dropbox-API-Arg header, escaping non-ASCII as \uXXXX
// since header values must be ASCII.
func dropboxArg(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, r := range string(b) {
		switch {
		case r < 0x80:
			sb.WriteRune(r)
		case r > 0xFFFF:
			r -= 0x10000
			fmt.Fprintf(&sb, `\u%04x\u%04x`, 0xD800+(r>>10), 0xDC00+(r&0x3FF))
		default:
			fmt.Fprintf(&sb, `\u%04x`, r)
		}
	}
	return sb.String(), nil
}

func dropboxPath(key string) string { return "/" + strings.Trim(path.Clean("/"+key), "/") }

func (d *dropboxStore) content_(ctx context.Context, endpoint string, arg any, body func() io.Reader, length int64) (*http.Response, error) {
	hdr, err := dropboxArg(arg)
	if err != nil {
		return nil, err
	}
	return d.api.do(ctx, func() (*http.Request, error) {
		var b io.Reader = http.NoBody
		if body != nil {
			b = body()
		}
		req, err := http.NewRequest(http.MethodPost, d.content+endpoint, b)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Dropbox-API-Arg", hdr)
		req.Header.Set("Content-Type", "application/octet-stream")
		if body != nil {
			req.ContentLength = length
		}
		return req, nil
	})
}

func (d *dropboxStore) Write(ctx context.Context, key string, r io.Reader, size int64, _ string) error {
	ra, total, err := readerAt(r, size)
	if err != nil {
		return err
	}
	p := dropboxPath(key)
	commit := map[string]any{"path": p, "mode": "overwrite", "mute": true}

	if total <= dropboxSimpleMax {
		resp, err := d.content_(ctx, "/2/files/upload", commit, func() io.Reader { return io.NewSectionReader(ra, 0, total) }, total)
		if err != nil {
			return err
		}
		if !okStatus(resp, http.StatusOK) {
			return apiError("upload", resp)
		}
		drain(resp)
		return nil
	}

	// Large files go through an upload session, which only commits at the end.
	first := min(dropboxChunk, total)
	resp, err := d.content_(ctx, "/2/files/upload_session/start", map[string]any{"close": false},
		func() io.Reader { return io.NewSectionReader(ra, 0, first) }, first)
	if err != nil {
		return err
	}
	if !okStatus(resp, http.StatusOK) {
		return apiError("start upload", resp)
	}
	var sess struct {
		ID string `json:"session_id"`
	}
	err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&sess)
	_ = resp.Body.Close()
	if err != nil || sess.ID == "" {
		return errors.New("start upload: no session returned")
	}

	off := first
	for total-off > dropboxChunk {
		cur := off
		resp, err := d.content_(ctx, "/2/files/upload_session/append_v2",
			map[string]any{"cursor": map[string]any{"session_id": sess.ID, "offset": cur}, "close": false},
			func() io.Reader { return io.NewSectionReader(ra, cur, dropboxChunk) }, dropboxChunk)
		if err != nil {
			return err
		}
		if !okStatus(resp, http.StatusOK) {
			return apiError("upload chunk", resp)
		}
		drain(resp)
		off += dropboxChunk
	}

	rest := total - off
	resp, err = d.content_(ctx, "/2/files/upload_session/finish",
		map[string]any{"cursor": map[string]any{"session_id": sess.ID, "offset": off}, "commit": commit},
		func() io.Reader { return io.NewSectionReader(ra, off, rest) }, rest)
	if err != nil {
		return err
	}
	if !okStatus(resp, http.StatusOK) {
		return apiError("finish upload", resp)
	}
	drain(resp)
	return nil
}

func (d *dropboxStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	resp, err := d.content_(ctx, "/2/files/download", map[string]any{"path": dropboxPath(key)}, nil, 0)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusConflict {
		drain(resp)
		return nil, notFound(key)
	}
	if !okStatus(resp, http.StatusOK) {
		return nil, apiError("download", resp)
	}
	return resp.Body, nil
}

func (d *dropboxStore) Delete(ctx context.Context, key string) error {
	arg := map[string]any{"path": dropboxPath(key)}
	resp, err := d.api.do(ctx, func() (*http.Request, error) {
		body, err := jsonBody(arg)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodPost, d.apiBase+"/2/files/delete_v2", body)
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return req, err
	})
	if err != nil {
		return err
	}
	if okStatus(resp, http.StatusOK) {
		drain(resp)
		return nil
	}
	if resp.StatusCode == http.StatusConflict {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		if strings.Contains(string(b), "not_found") {
			return nil
		}
		return fmt.Errorf("delete: 409 %s", sanitizeProviderText(string(b)))
	}
	return apiError("delete", resp)
}

func (d *dropboxStore) Close() error { return nil }
