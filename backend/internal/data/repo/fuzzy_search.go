package repo

import (
	"strings"

	"github.com/agext/levenshtein"
)

// FuzzyMatch checks if a string matches a query with fuzzy matching
// using Levenshtein distance algorithm
func FuzzyMatch(str, query string, threshold float64) bool {
	if query == "" {
		return true
	}

	// Convert to lowercase for case-insensitive comparison
	str = strings.ToLower(str)
	query = strings.ToLower(query)

	// Check if the query is a substring of the string
	if strings.Contains(str, query) {
		return true
	}

	// Calculate Levenshtein distance
	distance := levenshtein.Distance(str, query, nil)
	
	// Normalize the distance based on the length of the longer string
	maxLen := len(str)
	if len(query) > maxLen {
		maxLen = len(query)
	}
	
	if maxLen == 0 {
		return true
	}
	
	// Calculate similarity ratio (0.0 to 1.0)
	similarity := 1.0 - float64(distance)/float64(maxLen)
	
	// Return true if similarity is above the threshold
	return similarity >= threshold
}
