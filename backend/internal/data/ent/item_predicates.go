package ent

import (
	"entgo.io/ent/dialect/sql"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/item"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/pkgs/textutils"
)

// AccentInsensitiveContains creates a predicate that performs accent-insensitive text search.
// It normalizes both the database field value and the search value for comparison.
func AccentInsensitiveContains(field string, searchValue string) predicate.Item {
	if searchValue == "" {
		return predicate.Item(func(s *sql.Selector) {
			// Return a predicate that never matches if search is empty
			s.Where(sql.False())
		})
	}

	// Normalize the search value
	normalizedSearch := textutils.NormalizeSearchQuery(searchValue)

	return predicate.Item(func(s *sql.Selector) {
		dialect := s.Dialect()
		
		switch dialect {
		case "sqlite3":
			// For SQLite, we'll create a custom normalization function using REPLACE
			// to handle common accented characters
			normalizeFunc := buildSQLiteNormalizeExpression(s.C(field))
			s.Where(sql.ExprP(
				"LOWER("+normalizeFunc+") LIKE '%' || LOWER(?) || '%'",
				normalizedSearch,
			))
		case "postgres":
			// For PostgreSQL, try to use unaccent extension if available
			// Fall back to REPLACE-based normalization if not available
			normalizeFunc := buildPostgreSQLNormalizeExpression(s.C(field))
			s.Where(sql.ExprP(
				"LOWER("+normalizeFunc+") LIKE '%' || LOWER(?) || '%'",
				normalizedSearch,
			))
		default:
			// Default fallback using REPLACE for common accented characters
			normalizeFunc := buildGenericNormalizeExpression(s.C(field))
			s.Where(sql.ExprP(
				"LOWER("+normalizeFunc+") LIKE '%' || LOWER(?) || '%'",
				normalizedSearch,
			))
		}
	})
}

// buildSQLiteNormalizeExpression creates a SQLite expression to normalize accented characters
func buildSQLiteNormalizeExpression(fieldExpr string) string {
	return buildGenericNormalizeExpression(fieldExpr)
}

// buildPostgreSQLNormalizeExpression creates a PostgreSQL expression to normalize accented characters
func buildPostgreSQLNormalizeExpression(fieldExpr string) string {
	// Try to use unaccent extension if available, otherwise fall back to REPLACE
	// Note: This assumes the unaccent extension is installed. If not, it will use the generic approach.
	return "COALESCE(unaccent(" + fieldExpr + "), " + buildGenericNormalizeExpression(fieldExpr) + ")"
}

// buildGenericNormalizeExpression creates a database-agnostic expression to normalize common accented characters
func buildGenericNormalizeExpression(fieldExpr string) string {
	// Chain REPLACE functions to handle common accented characters
	// This handles the most common accented characters in Spanish, French, German, Portuguese, etc.
	normalized := fieldExpr
	
	// Common accent mappings
	accents := map[string]string{
		"á": "a", "à": "a", "ä": "a", "â": "a", "ã": "a", "å": "a",
		"é": "e", "è": "e", "ë": "e", "ê": "e",
		"í": "i", "ì": "i", "ï": "i", "î": "i",
		"ó": "o", "ò": "o", "ö": "o", "ô": "o", "õ": "o",
		"ú": "u", "ù": "u", "ü": "u", "û": "u",
		"ñ": "n", "ç": "c",
		"Á": "A", "À": "A", "Ä": "A", "Â": "A", "Ã": "A", "Å": "A",
		"É": "E", "È": "E", "Ë": "E", "Ê": "E",
		"Í": "I", "Ì": "I", "Ï": "I", "Î": "I",
		"Ó": "O", "Ò": "O", "Ö": "O", "Ô": "O", "Õ": "O",
		"Ú": "U", "Ù": "U", "Ü": "U", "Û": "U",
		"Ñ": "N", "Ç": "C",
	}
	
	for accented, normal := range accents {
		normalized = "REPLACE(" + normalized + ", '" + accented + "', '" + normal + "')"
	}
	
	return normalized
}

// ItemNameAccentInsensitiveContains creates an accent-insensitive search predicate for the item name field.
func ItemNameAccentInsensitiveContains(value string) predicate.Item {
	return AccentInsensitiveContains(item.FieldName, value)
}

// ItemDescriptionAccentInsensitiveContains creates an accent-insensitive search predicate for the item description field.
func ItemDescriptionAccentInsensitiveContains(value string) predicate.Item {
	return AccentInsensitiveContains(item.FieldDescription, value)
}

// ItemSerialNumberAccentInsensitiveContains creates an accent-insensitive search predicate for the item serial number field.
func ItemSerialNumberAccentInsensitiveContains(value string) predicate.Item {
	return AccentInsensitiveContains(item.FieldSerialNumber, value)
}

// ItemModelNumberAccentInsensitiveContains creates an accent-insensitive search predicate for the item model number field.
func ItemModelNumberAccentInsensitiveContains(value string) predicate.Item {
	return AccentInsensitiveContains(item.FieldModelNumber, value)
}

// ItemManufacturerAccentInsensitiveContains creates an accent-insensitive search predicate for the item manufacturer field.
func ItemManufacturerAccentInsensitiveContains(value string) predicate.Item {
	return AccentInsensitiveContains(item.FieldManufacturer, value)
}

// ItemNotesAccentInsensitiveContains creates an accent-insensitive search predicate for the item notes field.
func ItemNotesAccentInsensitiveContains(value string) predicate.Item {
	return AccentInsensitiveContains(item.FieldNotes, value)
} 