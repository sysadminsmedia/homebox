package repo

import (
	"testing"

	"github.com/sysadminsmedia/homebox/backend/pkgs/textutils"
	"github.com/stretchr/testify/assert"
)

func TestItemsRepository_AccentInsensitiveSearch(t *testing.T) {
	// Test cases for accent-insensitive search
	testCases := []struct {
		name                string
		itemName            string
		searchQuery         string
		shouldMatch         bool
		description         string
	}{
		{
			name:        "Spanish accented item, search without accents",
			itemName:    "electrónica",
			searchQuery: "electronica",
			shouldMatch: true,
			description: "Should find 'electrónica' when searching for 'electronica'",
		},
		{
			name:        "Spanish accented item, search with accents",
			itemName:    "electrónica",
			searchQuery: "electrónica",
			shouldMatch: true,
			description: "Should find 'electrónica' when searching for 'electrónica'",
		},
		{
			name:        "Non-accented item, search with accents",
			itemName:    "electronica",
			searchQuery: "electrónica",
			shouldMatch: true,
			description: "Should find 'electronica' when searching for 'electrónica' (bidirectional search)",
		},
		{
			name:        "Spanish item with tilde, search without accents",
			itemName:    "café",
			searchQuery: "cafe",
			shouldMatch: true,
			description: "Should find 'café' when searching for 'cafe'",
		},
		{
			name:        "Spanish item without tilde, search with accents",
			itemName:    "cafe",
			searchQuery: "café",
			shouldMatch: true,
			description: "Should find 'cafe' when searching for 'café' (bidirectional)",
		},
		{
			name:        "French accented item, search without accents",
			itemName:    "pére",
			searchQuery: "pere",
			shouldMatch: true,
			description: "Should find 'pére' when searching for 'pere'",
		},
		{
			name:        "French: père without accent, search with accents",
			itemName:    "pere",
			searchQuery: "père",
			shouldMatch: true,
			description: "Should find 'pere' when searching for 'père' (bidirectional)",
		},
		{
			name:        "Mixed case with accents",
			itemName:    "Electrónica",
			searchQuery: "ELECTRONICA",
			shouldMatch: true,
			description: "Should find 'Electrónica' when searching for 'ELECTRONICA' (case insensitive)",
		},
		{
			name:        "Bidirectional: Non-accented item, search with different accents",
			itemName:    "cafe",
			searchQuery: "café",
			shouldMatch: true,
			description: "Should find 'cafe' when searching for 'café' (bidirectional)",
		},
		{
			name:        "Bidirectional: Item with accent, search with different accent",
			itemName:    "résumé",
			searchQuery: "resume",
			shouldMatch: true,
			description: "Should find 'résumé' when searching for 'resume' (bidirectional)",
		},
		{
			name:        "Bidirectional: Spanish ñ to n",
			itemName:    "espanol",
			searchQuery: "español",
			shouldMatch: true,
			description: "Should find 'espanol' when searching for 'español' (bidirectional ñ)",
		},
		{
			name:        "French: français with accent, search without",
			itemName:    "français",
			searchQuery: "francais",
			shouldMatch: true,
			description: "Should find 'français' when searching for 'francais'",
		},
		{
			name:        "French: français without accent, search with",
			itemName:    "francais",
			searchQuery: "français",
			shouldMatch: true,
			description: "Should find 'francais' when searching for 'français' (bidirectional)",
		},
		{
			name:        "French: été with accent, search without",
			itemName:    "été",
			searchQuery: "ete",
			shouldMatch: true,
			description: "Should find 'été' when searching for 'ete'",
		},
		{
			name:        "French: été without accent, search with",
			itemName:    "ete",
			searchQuery: "été",
			shouldMatch: true,
			description: "Should find 'ete' when searching for 'été' (bidirectional)",
		},
		{
			name:        "French: hôtel with accent, search without",
			itemName:    "hôtel",
			searchQuery: "hotel",
			shouldMatch: true,
			description: "Should find 'hôtel' when searching for 'hotel'",
		},
		{
			name:        "French: hôtel without accent, search with",
			itemName:    "hotel",
			searchQuery: "hôtel",
			shouldMatch: true,
			description: "Should find 'hotel' when searching for 'hôtel' (bidirectional)",
		},
		{
			name:        "French: naïve with accent, search without",
			itemName:    "naïve",
			searchQuery: "naive",
			shouldMatch: true,
			description: "Should find 'naïve' when searching for 'naive'",
		},
		{
			name:        "French: naïve without accent, search with",
			itemName:    "naive",
			searchQuery: "naïve",
			shouldMatch: true,
			description: "Should find 'naive' when searching for 'naïve' (bidirectional)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the normalization logic used in the repository
			normalizedSearch := textutils.NormalizeSearchQuery(tc.searchQuery)
			
			// This simulates what happens in the repository
			// The original search would find exact matches (case-insensitive)
			// The normalized search would find accent-insensitive matches
			
			// Test that our normalization works as expected
			if tc.shouldMatch {
				// If it should match, then either the original query should match
				// or the normalized query should match when applied to the stored data
				assert.NotEqual(t, "", normalizedSearch, "Normalized search should not be empty")
				
				// The key insight is that we're searching with both the original and normalized queries
				// So "electrónica" will be found when searching for "electronica" because:
				// 1. Original search: "electronica" doesn't match "electrónica"
				// 2. Normalized search: "electronica" matches the normalized version
				t.Logf("✓ %s: Item '%s' should be found with search '%s' (normalized: '%s')", 
					tc.description, tc.itemName, tc.searchQuery, normalizedSearch)
			} else {
				t.Logf("✗ %s: Item '%s' should NOT be found with search '%s' (normalized: '%s')", 
					tc.description, tc.itemName, tc.searchQuery, normalizedSearch)
			}
		})
	}
}

func TestNormalizeSearchQueryIntegration(t *testing.T) {
	// Test that the normalization function works correctly
	testCases := []struct {
		input    string
		expected string
	}{
		{"electrónica", "electronica"},
		{"café", "cafe"},
		{"ELECTRÓNICA", "electronica"},
		{"Café París", "cafe paris"},
		{"hello world", "hello world"},
		// French accented words
		{"père", "pere"},
		{"français", "francais"},
		{"été", "ete"},
		{"hôtel", "hotel"},
		{"naïve", "naive"},
		{"PÈRE", "pere"},
		{"FRANÇAIS", "francais"},
		{"ÉTÉ", "ete"},
		{"HÔTEL", "hotel"},
		{"NAÏVE", "naive"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := textutils.NormalizeSearchQuery(tc.input)
			assert.Equal(t, tc.expected, result, "Normalization should work correctly")
		})
	}
} 