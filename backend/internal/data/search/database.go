package search

import (
	"context"
	"strings"
	"sync"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entityfield"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/tag"
	"github.com/sysadminsmedia/homebox/backend/pkgs/textutils"
)

// entityColumns are the entity table columns matched against each token.
var entityColumns = []string{
	entity.FieldName,
	entity.FieldDescription,
	entity.FieldSerialNumber,
	entity.FieldModelNumber,
	entity.FieldManufacturer,
	entity.FieldNotes,
	entity.FieldPurchaseFrom,
}

// DatabaseEngine implements Engine with tokenized substring matching executed
// by the database itself, so it needs no external services or index
// maintenance.
//
// The query is split into tokens (see Tokenize); an entity matches when every
// token matches at least one searched column, tag name, or custom field
// value. All matching is case-insensitive across the full Unicode range and
// accent-insensitive where the dialect allows:
//
//   - SQLite: both sides of the comparison go through hb_fold, a Go-defined
//     SQL function (registered by pkgs/cgofreesqlite) that applies Unicode
//     case folding and strips diacritics. SQLite's native LIKE/lower() are
//     ASCII-only and silently fail for Cyrillic, Greek, etc.
//   - PostgreSQL: ILIKE provides Unicode case-insensitivity natively, and the
//     unaccent extension is used for accent-insensitivity when available
//     (the engine tries to enable it once and degrades gracefully when the
//     database user lacks the privilege).
type DatabaseEngine struct {
	dialect string
	db      *ent.Client

	unaccentOnce sync.Once
	unaccent     bool
}

// NewDatabaseEngine returns a database-backed search engine querying through
// the given ent client.
func NewDatabaseEngine(db *ent.Client) *DatabaseEngine {
	return &DatabaseEngine{dialect: db.Dialect(), db: db}
}

// Predicate implements Engine.
func (e *DatabaseEngine) Predicate(ctx context.Context, _ uuid.UUID, query string) (predicate.Entity, error) {
	tokens := Tokenize(query)
	if len(tokens) == 0 {
		return nil, nil
	}

	match := e.matcher(ctx)

	tokenPreds := make([]predicate.Entity, 0, len(tokens))
	for _, token := range tokens {
		fieldPreds := make([]predicate.Entity, 0, len(entityColumns)+2)
		for _, col := range entityColumns {
			fieldPreds = append(fieldPreds, predicate.Entity(match(col, token)))
		}
		fieldPreds = append(fieldPreds,
			// Tag names and custom field values are searchable too
			// (requested in #1509 and #1380).
			entity.HasTagWith(predicate.Tag(match(tag.FieldName, token))),
			entity.HasFieldsWith(predicate.EntityField(match(entityfield.FieldTextValue, token))),
		)
		tokenPreds = append(tokenPreds, entity.Or(fieldPreds...))
	}
	return entity.And(tokenPreds...), nil
}

// matcher returns a function that builds a dialect-appropriate
// "column contains token" SQL condition. The returned closure is generic over
// the table being selected (entity, tag, entity_fields), qualifying the
// column through the active selector.
func (e *DatabaseEngine) matcher(ctx context.Context) func(col, token string) func(*entsql.Selector) {
	if e.dialect == dialect.Postgres {
		unaccent := e.unaccentAvailable(ctx)
		return func(col, token string) func(*entsql.Selector) {
			pattern := "%" + escapeLike(token) + "%"
			return func(s *entsql.Selector) {
				s.Where(entsql.P(func(b *entsql.Builder) {
					if unaccent {
						b.WriteString("unaccent(").WriteString(s.C(col)).WriteString(") ILIKE unaccent(")
						b.Arg(pattern)
						b.WriteString(")")
					} else {
						b.WriteString(s.C(col)).WriteString(" ILIKE ")
						b.Arg(pattern)
					}
				}))
			}
		}
	}

	// SQLite
	return func(col, token string) func(*entsql.Selector) {
		pattern := "%" + escapeLike(textutils.Fold(token)) + "%"
		return func(s *entsql.Selector) {
			s.Where(entsql.P(func(b *entsql.Builder) {
				b.WriteString("hb_fold(").WriteString(s.C(col)).WriteString(") LIKE ")
				b.Arg(pattern)
				b.WriteString(" ESCAPE '\\'")
			}))
		}
	}
}

// unaccentAvailable reports whether the PostgreSQL unaccent extension can be
// used. On first call it tries to enable the extension (ignoring permission
// errors) and caches the result for the lifetime of the engine.
func (e *DatabaseEngine) unaccentAvailable(ctx context.Context) bool {
	e.unaccentOnce.Do(func() {
		if e.db == nil {
			return
		}
		if _, err := e.db.Sql().ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS unaccent"); err != nil {
			log.Debug().Err(err).Msg("could not create unaccent extension (insufficient privileges?), checking if it already exists")
		}

		var count int
		row := e.db.Sql().QueryRowContext(ctx, "SELECT COUNT(*) FROM pg_extension WHERE extname = 'unaccent'")
		if err := row.Scan(&count); err != nil {
			log.Warn().Err(err).Msg("failed to check for unaccent extension; search will be accent-sensitive")
			return
		}

		e.unaccent = count > 0
		if e.unaccent {
			log.Info().Msg("postgres unaccent extension available; search is accent-insensitive")
		} else {
			log.Info().Msg("postgres unaccent extension not available; search will be accent-sensitive (install it with: CREATE EXTENSION unaccent)")
		}
	})
	return e.unaccent
}

// escapeLike escapes the LIKE wildcards in a literal token so user input
// cannot inject wildcard matching. Backslash is the escape character on both
// dialects (PostgreSQL's default; SQLite via an explicit ESCAPE clause).
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
