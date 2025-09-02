package ent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildGenericNormalizeExpression(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{
			name:     "Simple field name",
			field:    "name",
			expected: "name", // Should be wrapped in many REPLACE functions
		},
		{
			name:     "Complex field name",
			field:    "description",
			expected: "description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildGenericNormalizeExpression(tt.field)

			// Should contain the original field
			assert.Contains(t, result, tt.field)

			// Should contain REPLACE functions for accent normalization
			assert.Contains(t, result, "REPLACE(")

			// Should handle common accented characters
			assert.Contains(t, result, "'á'", "Should handle Spanish á")
			assert.Contains(t, result, "'é'", "Should handle Spanish é")
			assert.Contains(t, result, "'ñ'", "Should handle Spanish ñ")
			assert.Contains(t, result, "'ü'", "Should handle German ü")

			// Should handle uppercase accents too
			assert.Contains(t, result, "'Á'", "Should handle uppercase Spanish Á")
			assert.Contains(t, result, "'É'", "Should handle uppercase Spanish É")
		})
	}
}

func TestSQLiteNormalizeExpression(t *testing.T) {
	result := buildSQLiteNormalizeExpression("test_field")

	// Should contain the field name and REPLACE functions
	assert.Contains(t, result, "test_field")
	assert.Contains(t, result, "REPLACE(")
	// Check for some specific accent replacements (order doesn't matter)
	assert.Contains(t, result, "'á'", "Should handle Spanish á")
	assert.Contains(t, result, "'ó'", "Should handle Spanish ó")
}

func TestAccentInsensitivePredicateCreation(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		searchValue string
		description string
	}{
		{
			name:        "Normal search value",
			field:       "name",
			searchValue: "electronica",
			description: "Should create predicate for normal search",
		},
		{
			name:        "Accented search value",
			field:       "description",
			searchValue: "electrónica",
			description: "Should create predicate for accented search",
		},
		{
			name:        "Empty search value",
			field:       "name",
			searchValue: "",
			description: "Should handle empty search gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate := AccentInsensitiveContains(tt.field, tt.searchValue)
			assert.NotNil(t, predicate, tt.description)
		})
	}
}

func TestSpecificItemPredicates(t *testing.T) {
	tests := []struct {
		name          string
		predicateFunc func(string) interface{}
		searchValue   string
		description   string
	}{
		{
			name:          "ItemNameAccentInsensitiveContains",
			predicateFunc: func(val string) interface{} { return ItemNameAccentInsensitiveContains(val) },
			searchValue:   "electronica",
			description:   "Should create accent-insensitive name search predicate",
		},
		{
			name:          "ItemDescriptionAccentInsensitiveContains",
			predicateFunc: func(val string) interface{} { return ItemDescriptionAccentInsensitiveContains(val) },
			searchValue:   "descripcion",
			description:   "Should create accent-insensitive description search predicate",
		},
		{
			name:          "ItemManufacturerAccentInsensitiveContains",
			predicateFunc: func(val string) interface{} { return ItemManufacturerAccentInsensitiveContains(val) },
			searchValue:   "compañia",
			description:   "Should create accent-insensitive manufacturer search predicate",
		},
		{
			name:          "ItemSerialNumberAccentInsensitiveContains",
			predicateFunc: func(val string) interface{} { return ItemSerialNumberAccentInsensitiveContains(val) },
			searchValue:   "sn123",
			description:   "Should create accent-insensitive serial number search predicate",
		},
		{
			name:          "ItemModelNumberAccentInsensitiveContains",
			predicateFunc: func(val string) interface{} { return ItemModelNumberAccentInsensitiveContains(val) },
			searchValue:   "model456",
			description:   "Should create accent-insensitive model number search predicate",
		},
		{
			name:          "ItemNotesAccentInsensitiveContains",
			predicateFunc: func(val string) interface{} { return ItemNotesAccentInsensitiveContains(val) },
			searchValue:   "notas importantes",
			description:   "Should create accent-insensitive notes search predicate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate := tt.predicateFunc(tt.searchValue)
			assert.NotNil(t, predicate, tt.description)
		})
	}
}
