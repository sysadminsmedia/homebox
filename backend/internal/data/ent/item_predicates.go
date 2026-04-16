package ent

import (
	"entgo.io/ent/dialect/sql"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/item"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	conf "github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	// "github.com/sysadminsmedia/homebox/backend/pkgs/textutils"
	"strings"
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
	// normalizedSearch := textutils.NormalizeSearchQuery(searchValue)
	normalizedSearch := NormalizeString(searchValue)

	return predicate.Item(func(s *sql.Selector) {
		dialect := s.Dialect()

		switch dialect {
		case conf.DriverSqlite3:
			// For SQLite, we'll create a custom normalization function using REPLACE
			// to handle common accented characters
			normalizeFunc := buildSQLiteNormalizeExpression(s.C(field))
			s.Where(sql.ExprP(
				"LOWER("+normalizeFunc+") LIKE ?",
				"%"+normalizedSearch+"%",
			))
		case conf.DriverPostgres:
			// For PostgreSQL, use REPLACE-based normalization to avoid unaccent dependency
			normalizeFunc := buildGenericNormalizeExpression(s.C(field))
			// Use sql.P() for proper PostgreSQL parameter binding ($1, $2, etc.)
			s.Where(sql.P(func(b *sql.Builder) {
				b.WriteString("LOWER(")
				b.WriteString(normalizeFunc)
				b.WriteString(") LIKE ")
				b.Arg("%" + normalizedSearch + "%")
			}))
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

type accentPair struct {
	from string
	to   string
}

var accentReplacements = []accentPair{
	// Spanish
	{"á", "a"}, {"é", "e"}, {"í", "i"}, {"ó", "o"}, {"ú", "u"}, {"ñ", "n"},
	{"Á", "A"}, {"É", "E"}, {"Í", "I"}, {"Ó", "O"}, {"Ú", "U"}, {"Ñ", "N"},

	// French
	{"è", "e"}, {"ê", "e"}, {"à", "a"}, {"ç", "c"},
	{"È", "E"}, {"Ê", "E"}, {"À", "A"}, {"Ç", "C"},

	// German / Portuguese
	{"ä", "a"}, {"ö", "o"}, {"ü", "u"}, {"ã", "a"}, {"õ", "o"},
	{"Ä", "A"}, {"Ö", "O"}, {"Ü", "U"}, {"Ã", "A"}, {"Õ", "O"},

	// Greek lowercase
	{"α", "a"}, {"β", "v"}, {"γ", "g"}, {"δ", "d"}, {"ε", "e"},
	{"ζ", "z"}, {"η", "i"}, {"θ", "th"}, {"ι", "i"}, {"κ", "k"},
	{"λ", "l"}, {"μ", "m"}, {"ν", "n"}, {"ξ", "x"}, {"ο", "o"},
	{"π", "p"}, {"ρ", "r"}, {"σ", "s"}, {"ς", "s"}, {"τ", "t"},
	{"υ", "y"}, {"φ", "f"}, {"χ", "ch"}, {"ψ", "ps"}, {"ω", "o"},

	// Greek accented lowercase
	{"ά", "a"}, {"έ", "e"}, {"ή", "i"}, {"ί", "i"}, {"ϊ", "i"}, {"ΐ", "i"},
	{"ό", "o"}, {"ώ", "o"}, {"ύ", "y"}, {"ϋ", "y"}, {"ΰ", "y"},

	// Greek uppercase
	{"Α", "A"}, {"Β", "V"}, {"Γ", "G"}, {"Δ", "D"}, {"Ε", "E"},
	{"Ζ", "Z"}, {"Η", "I"}, {"Θ", "TH"}, {"Ι", "I"}, {"Κ", "K"},
	{"Λ", "L"}, {"Μ", "M"}, {"Ν", "N"}, {"Ξ", "X"}, {"Ο", "O"},
	{"Π", "P"}, {"Ρ", "R"}, {"Σ", "S"}, {"Τ", "T"}, {"Υ", "Y"},
	{"Φ", "F"}, {"Χ", "CH"}, {"Ψ", "PS"}, {"Ω", "O"},

	// Greek accented uppercase
	{"Ά", "A"}, {"Έ", "E"}, {"Ή", "I"}, {"Ί", "I"}, {"Ϊ", "I"},
	{"Ό", "O"}, {"Ώ", "O"}, {"Ύ", "Y"}, {"Ϋ", "Y"},
}

func buildGenericNormalizeExpression(fieldExpr string) string {
	normalized := fieldExpr

	for _, r := range accentReplacements {
		normalized = "REPLACE(" + normalized + ", '" + r.from + "', '" + r.to + "')"
	}
	return normalized
}

var accentReplacer = strings.NewReplacer(flattenReplacements(accentReplacements)...)

func flattenReplacements(pairs []accentPair) []string {
	out := make([]string, 0, len(pairs)*2)
	for _, p := range pairs {
		out = append(out, p.from, p.to)
	}
	return out
}

func NormalizeString(input string) string {
	if input == "" {
		return ""
	}
	return strings.ToLower(accentReplacer.Replace(input))
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
