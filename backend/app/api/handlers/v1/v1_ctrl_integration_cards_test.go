package v1

import "testing"

const paperlessTestBaseURL = "https://paperless.local"

func TestParsePaperlessAttachmentLink(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		baseURL string
		docID   string
		openURL string
		matches bool
	}{
		{
			name:    "document page",
			rawURL:  "https://paperless.local/documents/42/details",
			baseURL: paperlessTestBaseURL,
			docID:   "42",
			openURL: "https://paperless.local/documents/42/details",
			matches: true,
		},
		{
			name:    "document page without details suffix",
			rawURL:  "https://paperless.local/documents/42",
			baseURL: paperlessTestBaseURL,
			docID:   "42",
			openURL: "https://paperless.local/documents/42",
			matches: true,
		},
		{
			name:    "preview endpoint with base path",
			rawURL:  "https://example.local/paperless/api/documents/500/preview/?page=1#view",
			baseURL: "https://example.local/paperless",
			docID:   "500",
			openURL: "https://example.local/paperless/api/documents/500/preview/?page=1#view",
			matches: true,
		},
		{
			name:    "download endpoint",
			rawURL:  "https://paperless.local/api/documents/7/download",
			baseURL: paperlessTestBaseURL,
			docID:   "7",
			openURL: "https://paperless.local/api/documents/7/download",
			matches: true,
		},
		{
			name:    "foreign host",
			rawURL:  "https://paperless.local.evil/documents/42",
			baseURL: paperlessTestBaseURL,
			matches: false,
		},
		{
			name:    "wrong scheme",
			rawURL:  "http://paperless.local/documents/42",
			baseURL: paperlessTestBaseURL,
			matches: false,
		},
		{
			name:    "wrong base path",
			rawURL:  "https://example.local/other/documents/42",
			baseURL: "https://example.local/paperless",
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, ok := parsePaperlessAttachmentLink(tt.rawURL, paperlessConfig{
				baseURL: tt.baseURL,
				scope:   integrationScope("paperless", tt.baseURL),
			})
			if ok != tt.matches {
				t.Fatalf("match = %v, want %v", ok, tt.matches)
			}
			if !ok {
				return
			}
			if ref.docID != tt.docID {
				t.Fatalf("docID = %q, want %q", ref.docID, tt.docID)
			}
			if ref.openURL != tt.openURL {
				t.Fatalf("openURL = %q, want %q", ref.openURL, tt.openURL)
			}
		})
	}
}

func TestIntegrationScopeChangesWithBaseURL(t *testing.T) {
	a := integrationScope("paperless", paperlessTestBaseURL)
	b := integrationScope("paperless", "https://paperless.local/")
	c := integrationScope("paperless", "https://other.local")

	if a != b {
		t.Fatalf("trailing slash changed scope: %q != %q", a, b)
	}
	if a == c {
		t.Fatal("different base URLs should produce different scopes")
	}
}

func TestParseExternalHTTPURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		matches bool
	}{
		{name: "http", rawURL: "http://example.com/path", matches: true},
		{name: "https with query and fragment", rawURL: "https://example.com/path?q=1#section", matches: true},
		{name: "missing scheme", rawURL: "www.example.com/path", matches: false},
		{name: "unsupported scheme", rawURL: "ftp://example.com/path", matches: false},
		{name: "userinfo", rawURL: "https://user@example.com/path", matches: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := parseExternalHTTPURL(tt.rawURL)
			if ok != tt.matches {
				t.Fatalf("match = %v, want %v", ok, tt.matches)
			}
		})
	}
}
