package config

import "testing"

func Test_normalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "/"},
		{"/", "/"},
		{"homebox", "/homebox/"},
		{"/homebox", "/homebox/"},
		{"homebox/", "/homebox/"},
		{"/homebox/", "/homebox/"},
		{"/app/inventory", "/app/inventory/"},
		{"app/inventory/", "/app/inventory/"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := normalizePath(tt.input)
			if err != nil {
				t.Fatalf("normalizePath(%q) returned unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func Test_normalizePath_rejectsInvalidChars(t *testing.T) {
	invalids := []string{
		"<script>",
		"path with spaces",
		"path?query",
		"path#fragment",
	}
	for _, input := range invalids {
		t.Run(input, func(t *testing.T) {
			_, err := normalizePath(input)
			if err == nil {
				t.Errorf("normalizePath(%q) did not return error", input)
			}
		})
	}
}
