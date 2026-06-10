package textutils

import (
	"strings"
	"unicode"
)

// TokenizeSearchQuery splits a search query into tokens for multi-term matching.
// Tokens are separated by Unicode whitespace. A double-quoted segment is kept
// as a single token with the quotes stripped and inner whitespace preserved,
// allowing exact-phrase searches. An unbalanced quote consumes the remainder
// of the query as one token. Empty tokens are dropped.
//
// Example:
// - `item blue` becomes ["item", "blue"]
// - `red "blue box"` becomes ["red", "blue box"]
// - `  ` becomes []
func TokenizeSearchQuery(query string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}

	for _, r := range query {
		switch {
		case r == '"':
			inQuotes = !inQuotes
			flush()
		case unicode.IsSpace(r) && !inQuotes:
			flush()
		default:
			current.WriteRune(r)
		}
	}
	flush()

	return tokens
}
