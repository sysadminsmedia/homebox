package search

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/entity"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
)

// renderPredicate applies a predicate to a bare entity selector and returns
// the generated SQL and bound arguments. The end-to-end behavior against a
// real SQLite database is covered by the repo package tests; these tests pin
// the SQL shape per dialect, including the PostgreSQL form that cannot run in
// unit tests.
func renderPredicate(t *testing.T, dialectName string, p predicate.Entity) (string, []any) {
	t.Helper()
	s := entsql.Dialect(dialectName).
		Select(entity.FieldID).
		From(entsql.Table(entity.Table))
	p(s)
	query, args := s.Query()
	return query, args
}

func TestDatabaseEngine_SQLiteSQL(t *testing.T) {
	e := &DatabaseEngine{dialect: dialect.SQLite}

	pred, err := e.Predicate(context.Background(), uuid.Nil, "Straße")
	require.NoError(t, err)
	require.NotNil(t, pred)

	query, args := renderPredicate(t, dialect.SQLite, pred)

	// both sides folded: the column through hb_fold, the pattern in Go
	assert.Contains(t, query, "hb_fold(`entities`.`name`) LIKE ? ESCAPE '\\'")
	assert.Contains(t, args, "%strasse%")
}

func TestDatabaseEngine_PostgresSQL(t *testing.T) {
	// no db handle: unaccent probing reports unavailable, exercising the
	// plain ILIKE fallback
	e := &DatabaseEngine{dialect: dialect.Postgres}

	pred, err := e.Predicate(context.Background(), uuid.Nil, "café")
	require.NoError(t, err)
	require.NotNil(t, pred)

	query, args := renderPredicate(t, dialect.Postgres, pred)

	// ILIKE is Unicode case-insensitive natively; without unaccent the token
	// keeps its accents so accented data still matches accented queries
	assert.Contains(t, query, `"entities"."name" ILIKE $`)
	assert.Contains(t, args, "%café%")
}

func TestDatabaseEngine_PostgresUnaccentSQL(t *testing.T) {
	e := &DatabaseEngine{dialect: dialect.Postgres, unaccent: true}
	e.unaccentOnce.Do(func() {}) // mark probed

	pred, err := e.Predicate(context.Background(), uuid.Nil, "café")
	require.NoError(t, err)
	require.NotNil(t, pred)

	query, args := renderPredicate(t, dialect.Postgres, pred)

	assert.Contains(t, query, `unaccent("entities"."name") ILIKE unaccent($`)
	assert.Contains(t, args, "%café%")
}

func TestDatabaseEngine_EmptyQuery(t *testing.T) {
	e := &DatabaseEngine{dialect: dialect.SQLite}

	pred, err := e.Predicate(context.Background(), uuid.Nil, "   ")
	require.NoError(t, err)
	assert.Nil(t, pred)
}

func TestDatabaseEngine_MultiTokenStructure(t *testing.T) {
	e := &DatabaseEngine{dialect: dialect.SQLite}

	pred, err := e.Predicate(context.Background(), uuid.Nil, "red box")
	require.NoError(t, err)

	query, args := renderPredicate(t, dialect.SQLite, pred)

	// one AND-ed group per token, each ORing all searched surfaces,
	// including tag names and custom field values
	assert.Contains(t, args, "%red%")
	assert.Contains(t, args, "%box%")
	assert.Contains(t, query, "`tags`")
	assert.Contains(t, query, "`entity_fields`")
}
