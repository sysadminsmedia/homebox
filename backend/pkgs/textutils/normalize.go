// Package textutils provides text normalization helpers used by the search
// system to implement case- and accent-insensitive matching across scripts.
package textutils

import (
	"unicode"

	"golang.org/x/text/cases"
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

// Fold returns a canonical caseless, accent-less representation of text for
// search comparison. Two strings match case- and accent-insensitively iff
// their folded forms are equal (or one contains the other).
//
// Unicode case folding is used instead of lowercasing so that scripts with
// non-trivial case rules compare correctly (e.g. Greek final sigma "ς" and
// "σ" both fold to "σ", "Σ" included; Cyrillic "Тест" folds to "тест").
// Folding can introduce new combining marks (e.g. "İ" folds to "i" + U+0307),
// so accents are stripped after folding as well as before.
func Fold(text string) string {
	return RemoveAccents(cases.Fold().String(RemoveAccents(text)))
}
