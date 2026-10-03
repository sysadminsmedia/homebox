package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// fakeCloud imitates the OAuth and file APIs of Google Drive, OneDrive and
// Dropbox closely enough to exercise the real clients. It enforces what the
// real services enforce and the clients must get right: PKCE, valid bearer
// tokens, upload length headers, and no credentials on pre-signed upload URLs.
type fakeCloud struct {
	t   *testing.T
	srv *httptest.Server

	mu sync.Mutex
	// OAuth
	clientID, clientSecret string
	rotate                 map[string]bool // provider -> rotates refresh tokens
	codes                  map[string]string
	refresh                map[string]string // refresh token -> provider
	access                 map[string]string // access token -> provider
	tokenCalls             map[string]int
	seq                    int
	// Files, keyed by provider then path.
	files map[string]map[string][]byte
	// Google keeps an ID tree.
	gnodes   map[string]*gnode
	gsession map[string]gupload
	// OneDrive / Dropbox upload sessions.
	sessions map[string]*chunked
	// Number of authenticated API calls, to prove refreshes happen.
	apiCalls int
}

type gnode struct {
	name, parent string
	folder       bool
	data         []byte
}

type gupload struct {
	name, parent string
	length       int64
}

type chunked struct {
	path string
	buf  []byte
}

func newFakeCloud(t *testing.T) *fakeCloud {
	f := &fakeCloud{
		t: t, clientID: "client-id", clientSecret: "client-secret",
		rotate:     map[string]bool{providerMicrosoft: true},
		codes:      map[string]string{},
		refresh:    map[string]string{},
		access:     map[string]string{},
		tokenCalls: map[string]int{},
		files:      map[string]map[string][]byte{providerGoogle: {}, providerMicrosoft: {}, providerDropbox: {}},
		gnodes:     map[string]*gnode{"root": {name: "root", folder: true}},
		gsession:   map[string]gupload{},
		sessions:   map[string]*chunked{},
	}
	mux := http.NewServeMux()
	for _, p := range []string{providerGoogle, providerMicrosoft, providerDropbox} {
		mux.HandleFunc("/"+p+"/token", func(w http.ResponseWriter, r *http.Request) { f.token(p, w, r) })
		// The consent screen: approve immediately and send the user back.
		mux.HandleFunc("/"+p+"/auth", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			code := f.authorize(q.Get("code_challenge"))
			http.Redirect(w, r, q.Get("redirect_uri")+"?"+url.Values{"code": {code}, "state": {q.Get("state")}}.Encode(), http.StatusFound)
		})
	}
	mux.HandleFunc("/google/", f.google)
	mux.HandleFunc("/microsoft/", f.microsoft)
	mux.HandleFunc("/dropbox/", f.dropbox)
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeCloud) endpoints() map[string]providerEndpoints {
	u := f.srv.URL
	return map[string]providerEndpoints{
		providerGoogle:    {Auth: u + "/google/auth", Token: u + "/google/token", API: u + "/google", Upload: u + "/google"},
		providerMicrosoft: {Auth: u + "/microsoft/auth", Token: u + "/microsoft/token", API: u + "/microsoft"},
		providerDropbox:   {Auth: u + "/dropbox/auth", Token: u + "/dropbox/token", API: u + "/dropbox", Upload: u + "/dropbox"},
	}
}

func (f *fakeCloud) next(prefix string) string {
	f.seq++
	return fmt.Sprintf("%s-%d", prefix, f.seq)
}

// authorize plays the user approving access: it returns the code the provider
// would redirect back with, bound to the PKCE challenge from the auth URL.
func (f *fakeCloud) authorize(challenge string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	code := f.next("code")
	f.codes[code] = challenge
	return code
}

// expireAccessTokens invalidates every issued access token server-side.
func (f *fakeCloud) expireAccessTokens() {
	f.mu.Lock()
	f.access = map[string]string{}
	f.mu.Unlock()
}

// revokeRefreshTokens simulates the user removing the app from their account.
func (f *fakeCloud) revokeRefreshTokens() {
	f.mu.Lock()
	f.refresh = map[string]string{}
	f.mu.Unlock()
}

func (f *fakeCloud) token(provider string, w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tokenCalls[provider]++
	_ = r.ParseForm()
	fail := func(code, desc string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "error_description": desc})
	}
	if r.PostForm.Get("client_id") != f.clientID || r.PostForm.Get("client_secret") != f.clientSecret {
		fail("invalid_client", "bad client credentials")
		return
	}
	resp := map[string]any{"token_type": "Bearer", "expires_in": 3600}
	switch r.PostForm.Get("grant_type") {
	case "authorization_code":
		challenge, ok := f.codes[r.PostForm.Get("code")]
		delete(f.codes, r.PostForm.Get("code"))
		if !ok {
			fail("invalid_grant", "unknown or used code")
			return
		}
		sum := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
		if base64.RawURLEncoding.EncodeToString(sum[:]) != challenge {
			fail("invalid_grant", "PKCE verification failed")
			return
		}
		if r.PostForm.Get("redirect_uri") == "" {
			fail("invalid_request", "redirect_uri required")
			return
		}
		rt := f.next("rt")
		f.refresh[rt] = provider
		resp["refresh_token"] = rt
	case "refresh_token":
		old := r.PostForm.Get("refresh_token")
		if f.refresh[old] != provider {
			fail("invalid_grant", "refresh token revoked")
			return
		}
		if f.rotate[provider] {
			delete(f.refresh, old)
			rt := f.next("rt")
			f.refresh[rt] = provider
			resp["refresh_token"] = rt
		}
	default:
		fail("unsupported_grant_type", "")
		return
	}
	at := f.next("at")
	f.access[at] = provider
	resp["access_token"] = at
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// authed checks the bearer token; it answers 401 itself when invalid.
func (f *fakeCloud) authed(provider string, w http.ResponseWriter, r *http.Request) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.apiCalls++
	tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if tok == "" || f.access[tok] != provider {
		http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// ---------------------------------------------------------------------------
// Google Drive

var gQuery = regexp.MustCompile(`^name = '((?:[^'\\]|\\.)*)' and '([^']+)' in parents and trashed = false and mimeType (=|!=) 'application/vnd\.google-apps\.folder'$`)

func (f *fakeCloud) google(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/google")
	if strings.HasPrefix(p, "/upload-session/") {
		f.googleSessionPut(w, r, strings.TrimPrefix(p, "/upload-session/"))
		return
	}
	if !f.authed(providerGoogle, w, r) {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case p == "/drive/v3/about":
		writeJSON(w, map[string]any{"user": map[string]string{"emailAddress": "me@example.com", "displayName": "Me"}})
	case p == "/drive/v3/files" && r.Method == http.MethodGet:
		m := gQuery.FindStringSubmatch(r.URL.Query().Get("q"))
		if m == nil {
			http.Error(w, "bad query", http.StatusBadRequest)
			return
		}
		name := strings.NewReplacer(`\'`, `'`, `\\`, `\`).Replace(m[1])
		wantFolder := m[3] == "="
		var files []map[string]string
		for id, n := range f.gnodes {
			if n.name == name && n.parent == m[2] && n.folder == wantFolder && id != "root" {
				files = append(files, map[string]string{"id": id})
			}
		}
		writeJSON(w, map[string]any{"files": files})
	case p == "/drive/v3/files" && r.Method == http.MethodPost:
		var in struct {
			Name    string   `json:"name"`
			Parents []string `json:"parents"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		id := f.next("gf")
		f.gnodes[id] = &gnode{name: in.Name, parent: in.Parents[0], folder: true}
		writeJSON(w, map[string]string{"id": id})
	case p == "/upload/drive/v3/files" && r.Method == http.MethodPost:
		var in struct {
			Name    string   `json:"name"`
			Parents []string `json:"parents"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		n, _ := strconv.ParseInt(r.Header.Get("X-Upload-Content-Length"), 10, 64)
		sid := f.next("gs")
		f.gsession[sid] = gupload{name: in.Name, parent: in.Parents[0], length: n}
		w.Header().Set("Location", f.srv.URL+"/google/upload-session/"+sid)
	case strings.HasPrefix(p, "/drive/v3/files/"):
		id := strings.TrimPrefix(p, "/drive/v3/files/")
		n, ok := f.gnodes[id]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write(n.data)
		case http.MethodDelete:
			delete(f.gnodes, id)
			w.WriteHeader(http.StatusNoContent)
		}
	default:
		http.Error(w, "unhandled "+r.Method+" "+p, http.StatusNotFound)
	}
}

func (f *fakeCloud) googleSessionPut(w http.ResponseWriter, r *http.Request, sid string) {
	if !f.authed(providerGoogle, w, r) {
		return
	}
	body, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.gsession[sid]
	if !ok || int64(len(body)) != s.length || r.ContentLength != s.length {
		http.Error(w, "bad upload", http.StatusBadRequest)
		return
	}
	delete(f.gsession, sid)
	id := f.next("gf")
	f.gnodes[id] = &gnode{name: s.name, parent: s.parent, data: body}
	writeJSON(w, map[string]string{"id": id})
}

// googleFile returns a stored file's bytes by walking the tree from the root.
func (f *fakeCloud) googleFile(key string) ([]byte, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	parent := "root"
	segs := strings.Split(key, "/")
	for i, seg := range segs {
		found := false
		for id, n := range f.gnodes {
			if id != "root" && n.parent == parent && n.name == seg && n.folder == (i < len(segs)-1) {
				if i == len(segs)-1 {
					return n.data, true
				}
				parent, found = id, true
				break
			}
		}
		if !found {
			return nil, false
		}
	}
	return nil, false
}

// ---------------------------------------------------------------------------
// OneDrive

func (f *fakeCloud) microsoft(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.EscapedPath(), "/microsoft")
	if strings.HasPrefix(p, "/session/") {
		f.onedriveSession(w, r, strings.TrimPrefix(p, "/session/"))
		return
	}
	if !f.authed(providerMicrosoft, w, r) {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if p == "/me" {
		writeJSON(w, map[string]string{"userPrincipalName": "me@contoso.example", "mail": "me@contoso.example"})
		return
	}
	const prefix = "/me/drive/special/approot:/"
	if !strings.HasPrefix(p, prefix) {
		http.Error(w, "unhandled "+p, http.StatusNotFound)
		return
	}
	rest := strings.TrimPrefix(p, prefix)
	i := strings.LastIndex(rest, ":")
	rawPath, suffix := rest[:i], rest[i+1:]
	key, _ := url.PathUnescape(rawPath)

	switch {
	case suffix == "/content" && r.Method == http.MethodPut:
		body, _ := io.ReadAll(r.Body)
		f.files[providerMicrosoft][key] = body
		w.WriteHeader(http.StatusCreated)
	case suffix == "/content" && r.Method == http.MethodGet:
		b, ok := f.files[providerMicrosoft][key]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_, _ = w.Write(b)
	case suffix == "/createUploadSession" && r.Method == http.MethodPost:
		sid := f.next("ms")
		f.sessions[sid] = &chunked{path: key}
		writeJSON(w, map[string]string{"uploadUrl": f.srv.URL + "/microsoft/session/" + sid})
	case suffix == "" && r.Method == http.MethodDelete:
		if _, ok := f.files[providerMicrosoft][key]; !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		delete(f.files[providerMicrosoft], key)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "unhandled", http.StatusNotFound)
	}
}

var rangeRe = regexp.MustCompile(`^bytes (\d+)-(\d+)/(\d+)$`)

func (f *fakeCloud) onedriveSession(w http.ResponseWriter, r *http.Request, sid string) {
	if r.Header.Get("Authorization") != "" {
		http.Error(w, "pre-signed upload URLs must not receive the access token", http.StatusBadRequest)
		return
	}
	body, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[sid]
	m := rangeRe.FindStringSubmatch(r.Header.Get("Content-Range"))
	if !ok || m == nil {
		http.Error(w, "bad session request", http.StatusBadRequest)
		return
	}
	start, _ := strconv.ParseInt(m[1], 10, 64)
	end, _ := strconv.ParseInt(m[2], 10, 64)
	total, _ := strconv.ParseInt(m[3], 10, 64)
	if start != int64(len(s.buf)) || int64(len(body)) != end-start+1 || (end-start+1)%(320*1024) != 0 && end+1 != total {
		http.Error(w, "range mismatch", http.StatusRequestedRangeNotSatisfiable)
		return
	}
	s.buf = append(s.buf, body...)
	if int64(len(s.buf)) == total {
		f.files[providerMicrosoft][s.path] = s.buf
		delete(f.sessions, sid)
		w.WriteHeader(http.StatusCreated)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// ---------------------------------------------------------------------------
// Dropbox

func (f *fakeCloud) dropbox(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/dropbox")
	if !f.authed(providerDropbox, w, r) {
		return
	}
	var arg map[string]any
	if h := r.Header.Get("Dropbox-API-Arg"); h != "" {
		for _, c := range h {
			if c > 0x7f {
				http.Error(w, "Dropbox-API-Arg must be ASCII", http.StatusBadRequest)
				return
			}
		}
		_ = json.Unmarshal([]byte(h), &arg)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	pathArg := func() string { s, _ := arg["path"].(string); return strings.TrimPrefix(s, "/") }

	switch p {
	case "/2/users/get_current_account":
		writeJSON(w, map[string]any{"email": "me@dropbox.example", "name": map[string]string{"display_name": "Me"}})
	case "/2/files/upload":
		body, _ := io.ReadAll(r.Body)
		f.files[providerDropbox][pathArg()] = body
		writeJSON(w, map[string]string{"name": path.Base(pathArg())})
	case "/2/files/upload_session/start":
		body, _ := io.ReadAll(r.Body)
		sid := f.next("ds")
		f.sessions[sid] = &chunked{buf: body}
		writeJSON(w, map[string]string{"session_id": sid})
	case "/2/files/upload_session/append_v2", "/2/files/upload_session/finish":
		body, _ := io.ReadAll(r.Body)
		cur, _ := arg["cursor"].(map[string]any)
		sid, _ := cur["session_id"].(string)
		off, _ := cur["offset"].(float64)
		s, ok := f.sessions[sid]
		if !ok || int64(off) != int64(len(s.buf)) {
			http.Error(w, `{"error_summary":"incorrect_offset"}`, http.StatusConflict)
			return
		}
		s.buf = append(s.buf, body...)
		if p == "/2/files/upload_session/finish" {
			commit, _ := arg["commit"].(map[string]any)
			cp, _ := commit["path"].(string)
			f.files[providerDropbox][strings.TrimPrefix(cp, "/")] = s.buf
			delete(f.sessions, sid)
			writeJSON(w, map[string]string{"name": path.Base(cp)})
			return
		}
		w.WriteHeader(http.StatusOK)
	case "/2/files/download":
		b, ok := f.files[providerDropbox][pathArg()]
		if !ok {
			http.Error(w, `{"error_summary":"path/not_found/"}`, http.StatusConflict)
			return
		}
		_, _ = w.Write(b)
	case "/2/files/delete_v2":
		var in struct {
			Path string `json:"path"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		key := strings.TrimPrefix(in.Path, "/")
		if _, ok := f.files[providerDropbox][key]; !ok {
			http.Error(w, `{"error_summary":"path_lookup/not_found/.."}`, http.StatusConflict)
			return
		}
		delete(f.files[providerDropbox], key)
		writeJSON(w, map[string]string{})
	default:
		http.Error(w, "unhandled "+p, http.StatusNotFound)
	}
}

// stored returns the bytes the fake holds for a provider path.
func (f *fakeCloud) stored(provider, key string) ([]byte, bool) {
	if provider == providerGoogle {
		return f.googleFile(key)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.files[provider][key]
	return b, ok
}

func bytesReader(b []byte) io.Reader { return bytes.NewReader(b) }

func jsonUnmarshal(s string, v any) error { return json.Unmarshal([]byte(s), v) }
