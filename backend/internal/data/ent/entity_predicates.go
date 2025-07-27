package ent

import (
	"entgo.io/ent/dialect/sql"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/pkgs/textutils"
)

// AccentInsensitiveContains creates a predicate that performs accent-insensitive text search.
// It normalizes both the database field value and the search value for comparison.
func AccentInsensitiveContains(field string, searchValue string) predicate.Entity {
	if searchValue == "" {
		return predicate.Entity(func(s *sql.Selector) {
			// Return a predicate that never matches if search is empty
			s.Where(sql.False())
		})
	}

	// Normalize the search value
	normalizedSearch := textutils.NormalizeSearchQuery(searchValue)

	return predicate.Entity(func(s *sql.Selector) {
		dialect := s.Dialect()

		switch dialect {
		case "sqlite3":
			// For SQLite, we'll create a custom normalization function using REPLACE
			// to handle common accented characters
			normalizeFunc := buildSQLiteNormalizeExpression(s.C(field))
			s.Where(sql.ExprP(
				"LOWER("+normalizeFunc+") LIKE ?",
				"%"+normalizedSearch+"%",
			))
		case "postgres":
			// For PostgreSQL, try to use unaccent extension if available
			// Fall back to REPLACE-based normalization if not available
			normalizeFunc := buildPostgreSQLNormalizeExpression(s.C(field))
			s.Where(sql.ExprP(
				"LOWER("+normalizeFunc+") LIKE ?",
				"%"+normalizedSearch+"%",
			))
		default:
			// Default fallback using REPLACE for common accented characters
			normalizeFunc := buildGenericNormalizeExpression(s.C(field))
			s.Where(sql.ExprP(
				"LOWER("+normalizeFunc+") LIKE ?",
				"%"+normalizedSearch+"%",
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
	// Use a CASE statement to check if unaccent function exists before using it
	// This prevents errors when the unaccent extension is not installed
	return "CASE WHEN EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'unaccent') " +
		"THEN unaccent(" + fieldExpr + ") " +
		"ELSE " + buildGenericNormalizeExpression(fieldExpr) + " END"
}

// buildGenericNormalizeExpression creates a database-agnostic expression to normalize common accented characters
func buildGenericNormalizeExpression(fieldExpr string) string {
	// Chain REPLACE functions to handle the most common accented characters
	// Focused on the most frequently used accents in Spanish, French, and Portuguese
	// Ordered by frequency of use for better performance
	normalized := fieldExpr

	// Most common accented characters ordered by frequency
	commonAccents := []struct {
		from, to string
	}{
		// Spanish - most common
		{"á", "a"}, {"é", "e"}, {"í", "i"}, {"ó", "o"}, {"ú", "u"}, {"ñ", "n"},
		{"Á", "A"}, {"É", "E"}, {"Í", "I"}, {"Ó", "O"}, {"Ú", "U"}, {"Ñ", "N"},

		// French - most common
		{"è", "e"}, {"ê", "e"}, {"à", "a"}, {"ç", "c"},
		{"È", "E"}, {"Ê", "E"}, {"À", "A"}, {"Ç", "C"},

		// German umlauts and Portuguese - common
		{"ä", "a"}, {"ö", "o"}, {"ü", "u"}, {"ã", "a"}, {"õ", "o"},
		{"Ä", "A"}, {"Ö", "O"}, {"Ü", "U"}, {"Ã", "A"}, {"Õ", "O"},
	}

	for _, accent := range commonAccents {
		normalized = "REPLACE(" + normalized + ", '" + accent.from + "', '" + accent.to + "')"
	}

	return normalized
}

// ItemNameAccentInsensitiveContains creates an accent-insensitive search predicate for the item name field.
func ItemNameAccentInsensitiveContains(value string) predicate.Entity {
	return AccentInsensitiveContains(entity.FieldName, value)
}

// ItemDescriptionAccentInsensitiveContains creates an accent-insensitive search predicate for the item description field.
func ItemDescriptionAccentInsensitiveContains(value string) predicate.Entity {
	return AccentInsensitiveContains(entity.FieldDescription, value)
}

// ItemSerialNumberAccentInsensitiveContains creates an accent-insensitive search predicate for the item serial number field.
func ItemSerialNumberAccentInsensitiveContains(value string) predicate.Entity {
	return AccentInsensitiveContains(entity.FieldSerialNumber, value)
}

// ItemModelNumberAccentInsensitiveContains creates an accent-insensitive search predicate for the item model number field.
func ItemModelNumberAccentInsensitiveContains(value string) predicate.Entity {
	return AccentInsensitiveContains(entity.FieldModelNumber, value)
}

// ItemManufacturerAccentInsensitiveContains creates an accent-insensitive search predicate for the item manufacturer field.
func ItemManufacturerAccentInsensitiveContains(value string) predicate.Entity {
	return AccentInsensitiveContains(entity.FieldManufacturer, value)
}

// ItemNotesAccentInsensitiveContains creates an accent-insensitive search predicate for the item notes field.
func ItemNotesAccentInsensitiveContains(value string) predicate.Entity {
	return AccentInsensitiveContains(entity.FieldNotes, value)
}
