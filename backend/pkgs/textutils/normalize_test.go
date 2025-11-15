package textutils

import (
	"strings"
	"testing"
)

func TestRemoveAccents(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Spanish accented characters",
			input:    "electrónica",
			expected: "electronica",
		},
		{
			name:     "Spanish accented characters with tilde",
			input:    "café",
			expected: "cafe",
		},
		{
			name:     "French accented characters",
			input:    "père",
			expected: "pere",
		},
		{
			name:     "German umlauts",
			input:    "Björk",
			expected: "Bjork",
		},
		{
			name:     "Mixed accented characters",
			input:    "résumé",
			expected: "resume",
		},
		{
			name:     "Portuguese accented characters",
			input:    "João",
			expected: "Joao",
		},
		{
			name:     "No accents",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Numbers and symbols",
			input:    "123!@#",
			expected: "123!@#",
		},
		{
			name:     "Multiple accents in one word",
			input:    "été",
			expected: "ete",
		},
		{
			name:     "Complex Unicode characters",
			input:    "français",
			expected: "francais",
		},
		{
			name:     "Unicode diacritics",
			input:    "naïve",
			expected: "naive",
		},
		{
			name:     "Unicode combining characters",
			input:    "e\u0301", // e with combining acute accent
			expected: "e",
		},
		{
			name:     "Very long string with accents",
			input:    strings.Repeat("café", 1000),
			expected: strings.Repeat("cafe", 1000),
		},
		{
			name:     "All French accents",
			input:    "àâäéèêëïîôöùûüÿç",
			expected: "aaaeeeeiioouuuyc",
		},
		{
			name:     "All Spanish accents",
			input:    "áéíóúñüÁÉÍÓÚÑÜ",
			expected: "aeiounuAEIOUNU",
		},
		{
			name:     "All German umlauts",
			input:    "äöüÄÖÜß",
			expected: "aouAOUß",
		},
		{
			name:     "Mixed languages",
			input:    "Français café España niño",
			expected: "Francais cafe Espana nino",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := RemoveAccents(tc.input)
			if result != tc.expected {
				t.Errorf("RemoveAccents(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestNormalizeSearchQuery(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Uppercase with accents",
			input:    "ELECTRÓNICA",
			expected: "electronica",
		},
		{
			name:     "Mixed case with accents",
			input:    "Electrónica",
			expected: "electronica",
		},
		{
			name:     "Multiple words with accents",
			input:    "Café París",
			expected: "cafe paris",
		},
		{
			name:     "No accents mixed case",
			input:    "Hello World",
			expected: "hello world",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeSearchQuery(tc.input)
			if result != tc.expected {
				t.Errorf("NormalizeSearchQuery(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}
