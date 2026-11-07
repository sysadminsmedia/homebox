// Package textutils provides text normalization helpers used by the search
// system to implement case- and accent-insensitive matching across scripts.
package textutils

import (
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

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
	// Fast path: for pure ASCII input, Unicode case folding is plain
	// lowercasing and there are no diacritics to strip. This is the common case
	// for hb_fold, which SQLite calls once per row per searched column, and
	// strings.ToLower doesn't allocate when the input is already lowercase.
	if isASCII(text) {
		return strings.ToLower(text)
	}

	t := foldPool.Get().(transform.Transformer)
	defer foldPool.Put(t)
	t.Reset()
	result, _, err := transform.String(t, text)
	if err != nil {
		return text
	}
	return result
}

// foldPool reuses the fold transformer chain. Building it allocates several
// KB of internal buffers, which dominated search cost when done per call;
// a Chain is stateful, so instances are pooled rather than shared.
var foldPool = sync.Pool{
	New: func() any {
		// Strip accents, fold case, then strip again (folding can introduce
		// combining marks, e.g. "İ" -> "i" + U+0307), and recompose.
		return transform.Chain(
			norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC,
			cases.Fold(),
			norm.NFD, runes.Remove(runes.In(unicode.Mn)),
			norm.NFC,
		)
	},
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
