package search

import (
	"strings"
	"unicode"
)

// maxTokens bounds the number of tokens a single query can expand into so a
// pathological query cannot generate an unbounded SQL statement.
const maxTokens = 8

// Tokenize splits a free-text query into match tokens.
//
// Tokens are separated by whitespace. A double-quoted span is kept together
// as a single token (without the quotes) so users can search for exact
// phrases, e.g. `red "tool box"` yields ["red", "tool box"]. Duplicate tokens
// are dropped, and at most maxTokens tokens are returned.
func Tokenize(query string) []string {
	var (
		tokens   []string
		current  strings.Builder
		inQuotes bool
	)

	seen := make(map[string]struct{})
	flush := func() {
		tok := current.String()
		current.Reset()
		if tok == "" {
			return
		}
		if _, dup := seen[tok]; dup {
			return
		}
		seen[tok] = struct{}{}
		tokens = append(tokens, tok)
	}

	for _, r := range query {
		switch {
		case r == '"':
			if inQuotes {
				flush()
			}
			inQuotes = !inQuotes
		case !inQuotes && unicode.IsSpace(r):
			flush()
		default:
			current.WriteRune(r)
		}
	}
	flush()

	if len(tokens) > maxTokens {
		tokens = tokens[:maxTokens]
	}
	return tokens
}
