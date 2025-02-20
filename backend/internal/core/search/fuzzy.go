package search

import (
	"strings"
	"unicode"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/mozillazg/go-slugify"
)

// SearchResult represents a matched string with its score
type SearchResult struct {
	Text     string
	Distance int
}

// FuzzySearcher provides fuzzy search capabilities
type FuzzySearcher struct {
	// MaxDistance is the maximum Levenshtein distance allowed for a match
	MaxDistance int
}

// NewFuzzySearcher creates a new FuzzySearcher with default settings
func NewFuzzySearcher() *FuzzySearcher {
	return &FuzzySearcher{
		MaxDistance: 2, // Allow for up to 2 character differences by default
	}
}

// Search performs a fuzzy search on the given text against a list of candidates
// It returns matches sorted by their Levenshtein distance (closest matches first)
func (fs *FuzzySearcher) Search(query string, candidates []string) []SearchResult {
	// Normalize the query
	query = normalizeText(query)
	
	var results []SearchResult
	
	// Check each candidate
	for _, candidate := range candidates {
		normalizedCandidate := normalizeText(candidate)
		
		// First try exact substring match (case-insensitive)
		if strings.Contains(normalizedCandidate, query) {
			results = append(results, SearchResult{
				Text:     candidate,
				Distance: 0,
			})
			continue
		}
		
		// If no exact match, try fuzzy match
		distance := fuzzy.LevenshteinDistance(query, normalizedCandidate)
		if distance <= fs.MaxDistance {
			results = append(results, SearchResult{
				Text:     candidate,
				Distance: distance,
			})
		}
		
		// Also try matching individual words for multi-word queries
		queryWords := strings.Fields(query)
		if len(queryWords) > 1 {
			candidateWords := strings.Fields(normalizedCandidate)
			matchCount := 0
			
			for _, qWord := range queryWords {
				for _, cWord := range candidateWords {
					if distance := fuzzy.LevenshteinDistance(qWord, cWord); distance <= fs.MaxDistance {
						matchCount++
						break
					}
				}
			}
			
			// If we matched all query words, add this as a result
			if matchCount == len(queryWords) {
				results = append(results, SearchResult{
					Text:     candidate,
					Distance: 1, // Prioritize these matches above pure fuzzy matches
				})
			}
		}
		
		// Try phonetic matching
		if FuzzyMatch(query, candidate) {
			results = append(results, SearchResult{
				Text:     candidate,
				Distance: 1, // Prioritize phonetic matches
			})
		}
	}
	
	return results
}

// normalizeText prepares text for fuzzy matching by:
// - Converting to lowercase
// - Removing extra whitespace
// - Removing special characters
func normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)
	
	// Remove extra whitespace
	text = strings.Join(strings.Fields(text), " ")
	
	// Remove special characters (keep letters, numbers, and spaces)
	var result strings.Builder
	for _, ch := range text {
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			ch == ' ' {
			result.WriteRune(ch)
		}
	}
	
	return result.String()
}

// FuzzyMatch checks if needle matches haystack using both fuzzy string matching
// and phonetic matching for better search results
func FuzzyMatch(needle, haystack string) bool {
	// Normalize strings
	needle = strings.ToLower(strings.TrimSpace(needle))
	haystack = strings.ToLower(strings.TrimSpace(haystack))

	// Early return if exact match after normalization
	if needle == haystack {
		return true
	}

	// Check for substring match
	if strings.Contains(haystack, needle) {
		return true
	}

	// Generate phonetic versions
	needlePhonetic := slugify.Slugify(needle)
	haystackPhonetic := slugify.Slugify(haystack)

	// Check for phonetic match
	if needlePhonetic == haystackPhonetic {
		return true
	}

	// Check for fuzzy match with original strings
	if fuzzy.LevenshteinDistance(needle, haystack) <= 2 {
		return true
	}

	// Check for fuzzy match with phonetic versions
	if fuzzy.LevenshteinDistance(needlePhonetic, haystackPhonetic) <= 2 {
		return true
	}

	// Split into words and check each
	needleWords := splitIntoWords(needle)
	haystackWords := splitIntoWords(haystack)

	for _, nWord := range needleWords {
		for _, hWord := range haystackWords {
			// Skip very short words
			if len(nWord) < 3 || len(hWord) < 3 {
				continue
			}

			// Check phonetic match for individual words
			nWordPhonetic := slugify.Slugify(nWord)
			hWordPhonetic := slugify.Slugify(hWord)

			if nWordPhonetic == hWordPhonetic {
				return true
			}

			// Check fuzzy match for individual words
			if fuzzy.LevenshteinDistance(nWord, hWord) <= 2 {
				return true
			}

			// Check fuzzy match for phonetic versions of words
			if fuzzy.LevenshteinDistance(nWordPhonetic, hWordPhonetic) <= 2 {
				return true
			}
		}
	}

	return false
}

// splitIntoWords splits a string into words, handling various separators
func splitIntoWords(s string) []string {
	var words []string
	var currentWord strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}
