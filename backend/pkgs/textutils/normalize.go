package textutils

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// RemoveAccents removes accents from text by normalizing Unicode characters
// and removing diacritical marks. This allows for accent-insensitive search.
//
// Example:
// - "electrónica" becomes "electronica"
// - "café" becomes "cafe"
// - "père" becomes "pere"
func RemoveAccents(text string) string {
	// Create a transformer that:
	// 1. Normalizes to NFD (canonical decomposition)
	// 2. Removes diacritical marks (combining characters)
	// 3. Normalizes back to NFC (canonical composition)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	result, _, err := transform.String(t, text)
	if err != nil {
		// If transformation fails, return the original text
		return text
	}

	return result
}

// NormalizeSearchQuery normalizes a search query for accent-insensitive matching.
// This function removes accents and converts to lowercase for consistent search behavior.
func NormalizeSearchQuery(query string) string {
	normalized := RemoveAccents(query)
	return strings.ToLower(normalized)
}
