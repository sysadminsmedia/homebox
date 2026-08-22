package search

import (
	"context"
	"errors"
	"fmt"
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

	unaccentMu      sync.Mutex
	unaccentChecked bool
	unaccent        bool
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
	return func(col, token string) func(*entsql.Selector) {
		return func(s *entsql.Selector) {
			s.Where(e.foldContains(ctx, s.C(col), token))
		}
	}
}

// foldContains builds a dialect-appropriate, case- and (where available)
// accent-insensitive "qualified column contains token" predicate. col must
// already be qualified (e.g. via Selector.C or SelectTable.C). It is the shared
// core behind both free-text matching and facet-value narrowing.
func (e *DatabaseEngine) foldContains(ctx context.Context, col, token string) *entsql.Predicate {
	if e.dialect == dialect.Postgres {
		unaccent := e.unaccentAvailable(ctx)
		pattern := "%" + escapeLike(token) + "%"
		return entsql.P(func(b *entsql.Builder) {
			if unaccent {
				b.WriteString("unaccent(").WriteString(col).WriteString(") ILIKE unaccent(")
				b.Arg(pattern)
				b.WriteString(")")
			} else {
				b.WriteString(col).WriteString(" ILIKE ")
				b.Arg(pattern)
			}
		})
	}

	// SQLite
	pattern := "%" + escapeLike(textutils.Fold(token)) + "%"
	return entsql.P(func(b *entsql.Builder) {
		b.WriteString("hb_fold(").WriteString(col).WriteString(") LIKE ")
		b.Arg(pattern)
		b.WriteString(" ESCAPE '\\'")
	})
}

// unaccentAvailable reports whether the PostgreSQL unaccent extension can be
// used. On first call it tries to enable the extension (ignoring permission
// errors) and caches the result for the lifetime of the engine.
func (e *DatabaseEngine) unaccentAvailable(ctx context.Context) bool {
	e.unaccentMu.Lock()
	defer e.unaccentMu.Unlock()

	if e.unaccentChecked {
		return e.unaccent
	}
	if e.db == nil {
		return false
	}

	if _, err := e.db.Sql().ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS unaccent"); err != nil {
		log.Debug().Err(err).Msg("could not create unaccent extension (insufficient privileges?), checking if it already exists")
	}

	var count int
	row := e.db.Sql().QueryRowContext(ctx, "SELECT COUNT(*) FROM pg_extension WHERE extname = 'unaccent'")
	if err := row.Scan(&count); err != nil {
		// A caller-context cancellation or deadline is transient: leave the
		// probe unmarked so a later call retries, rather than permanently
		// caching accent-sensitive search from a one-off timeout.
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			log.Debug().Err(err).Msg("unaccent probe interrupted by context; will retry on next search")
			return false
		}
		// Any other failure is treated as a definitive negative result.
		log.Warn().Err(err).Msg("failed to check for unaccent extension; search will be accent-sensitive")
		e.unaccentChecked = true
		return false
	}

	e.unaccent = count > 0
	e.unaccentChecked = true
	if e.unaccent {
		log.Info().Msg("postgres unaccent extension available; search is accent-insensitive")
	} else {
		log.Info().Msg("postgres unaccent extension not available; search will be accent-sensitive (install it with: CREATE EXTENSION unaccent)")
	}
	return e.unaccent
}

// escapeLike escapes the LIKE wildcards in a literal token so user input
// cannot inject wildcard matching. Backslash is the escape character on both
// dialects (PostgreSQL's default; SQLite via an explicit ESCAPE clause).
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

// --- Faceter implementation -------------------------------------------------
//
// These mirror the Meilisearch engine's facet methods so the search UI behaves
// the same regardless of driver. Counts are entity counts within the group; the
// grouping key (tag name / field value) keeps its original casing while the
// optional narrowing query matches case- and accent-insensitively, matching
// both the free-text search and Meilisearch's facetQuery.

// notEmpty is the "<col> is a non-empty string" predicate, used to mirror the
// Meilisearch document builder, which only facets fields that have a text value.
func notEmpty(col string) *entsql.Predicate {
	return entsql.P(func(b *entsql.Builder) {
		b.WriteString(col).WriteString(" <> ''")
	})
}

// SearchTags implements Faceter.
func (e *DatabaseEngine) SearchTags(ctx context.Context, gid uuid.UUID, query string) ([]TagFacet, error) {
	t := entsql.Table(tag.Table).As("t")
	te := entsql.Table(entity.TagTable).As("te")
	// entity.TagPrimaryKey is {tag_id, entity_id}; count distinct entities.
	cnt := entsql.Count(entsql.Distinct(te.C(entity.TagPrimaryKey[1])))

	sel := entsql.Dialect(e.dialect).
		Select(t.C(tag.FieldName), entsql.As(cnt, "count")).
		From(t).
		Join(te).On(te.C(entity.TagPrimaryKey[0]), t.C(tag.FieldID)).
		Where(entsql.EQ(t.C(tag.GroupColumn), gid)).
		GroupBy(t.C(tag.FieldName)).
		OrderBy(entsql.Desc(cnt), entsql.Asc(t.C(tag.FieldName)))
	if q := strings.TrimSpace(query); q != "" {
		sel.Where(e.foldContains(ctx, t.C(tag.FieldName), q))
	}

	rows, err := e.scanFacets(ctx, sel)
	if err != nil {
		return nil, fmt.Errorf("database tag facets: %w", err)
	}
	out := make([]TagFacet, len(rows))
	for i, r := range rows {
		out[i] = TagFacet{Name: r.key, Count: r.count}
	}
	return out, nil
}

// SearchFieldValues implements Faceter.
func (e *DatabaseEngine) SearchFieldValues(ctx context.Context, gid uuid.UUID, field, query string) ([]FieldFacet, error) {
	f := entsql.Table(entityfield.Table).As("f")
	en := entsql.Table(entity.Table).As("e")
	cnt := entsql.Count(entsql.Distinct(f.C(entityfield.EntityColumn)))

	sel := entsql.Dialect(e.dialect).
		Select(f.C(entityfield.FieldTextValue), entsql.As(cnt, "count")).
		From(f).
		Join(en).On(f.C(entityfield.EntityColumn), en.C(entity.FieldID)).
		Where(entsql.EQ(en.C(entity.GroupColumn), gid)).
		Where(entsql.EQ(f.C(entityfield.FieldName), field)).
		Where(notEmpty(f.C(entityfield.FieldTextValue))).
		GroupBy(f.C(entityfield.FieldTextValue)).
		OrderBy(entsql.Desc(cnt), entsql.Asc(f.C(entityfield.FieldTextValue)))
	if q := strings.TrimSpace(query); q != "" {
		sel.Where(e.foldContains(ctx, f.C(entityfield.FieldTextValue), q))
	}

	rows, err := e.scanFacets(ctx, sel)
	if err != nil {
		return nil, fmt.Errorf("database field value facets: %w", err)
	}
	out := make([]FieldFacet, len(rows))
	for i, r := range rows {
		out[i] = FieldFacet{Value: r.key, Count: r.count}
	}
	return out, nil
}

// FieldFacets implements Faceter. A single grouped query yields every
// (field name, value) pair with its entity count, which is then bucketed by
// field name.
func (e *DatabaseEngine) FieldFacets(ctx context.Context, gid uuid.UUID) (map[string][]FieldFacet, error) {
	f := entsql.Table(entityfield.Table).As("f")
	en := entsql.Table(entity.Table).As("e")
	cnt := entsql.Count(entsql.Distinct(f.C(entityfield.EntityColumn)))

	sel := entsql.Dialect(e.dialect).
		Select(f.C(entityfield.FieldName), f.C(entityfield.FieldTextValue), entsql.As(cnt, "count")).
		From(f).
		Join(en).On(f.C(entityfield.EntityColumn), en.C(entity.FieldID)).
		Where(entsql.EQ(en.C(entity.GroupColumn), gid)).
		Where(notEmpty(f.C(entityfield.FieldTextValue))).
		GroupBy(f.C(entityfield.FieldName), f.C(entityfield.FieldTextValue)).
		OrderBy(entsql.Asc(f.C(entityfield.FieldName)), entsql.Desc(cnt), entsql.Asc(f.C(entityfield.FieldTextValue)))

	q, args := sel.Query()
	rows, err := e.db.Sql().QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("database field facets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string][]FieldFacet)
	for rows.Next() {
		var name, value string
		var count int
		if err := rows.Scan(&name, &value, &count); err != nil {
			return nil, fmt.Errorf("database field facets: %w", err)
		}
		out[name] = append(out[name], FieldFacet{Value: value, Count: count})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("database field facets: %w", err)
	}
	return out, nil
}

// facetRow is a (grouping key, entity count) pair from a two-column facet query.
type facetRow struct {
	key   string
	count int
}

func (e *DatabaseEngine) scanFacets(ctx context.Context, sel *entsql.Selector) ([]facetRow, error) {
	q, args := sel.Query()
	rows, err := e.db.Sql().QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []facetRow
	for rows.Next() {
		var r facetRow
		if err := rows.Scan(&r.key, &r.count); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
