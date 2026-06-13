package search

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
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

// TestDatabaseEngine_Facets exercises the Faceter implementation end-to-end
// against a real (in-memory SQLite) database, including per-group scoping,
// entity counts, ranking, and case-insensitive narrowing.
func TestDatabaseEngine_Facets(t *testing.T) {
	assertFacets(t, newTestEntClient(t))
}

// TestDatabaseEngine_FacetsPostgres runs the same scenario against PostgreSQL,
// which—unlike SQLite—stores ids in a native uuid column. This is what proves
// the raw-SQL group filter binds the uuid argument correctly on pgx. Set
// TEST_POSTGRES_URL (e.g. from a throwaway docker container) to run it.
func TestDatabaseEngine_FacetsPostgres(t *testing.T) {
	assertFacets(t, newTestPostgresClient(t))
}

// assertFacets seeds a fixed two-group dataset and asserts the facet behavior;
// shared so SQLite and PostgreSQL are held to identical expectations.
func assertFacets(t *testing.T, db *ent.Client) {
	t.Helper()
	ctx := context.Background()
	e := NewDatabaseEngine(db)

	g1 := db.Group.Create().SetName("g1").SaveX(ctx)
	g2 := db.Group.Create().SetName("g2").SaveX(ctx)
	et1 := db.EntityType.Create().SetName("Item").SetGroup(g1).SaveX(ctx)
	et2 := db.EntityType.Create().SetName("Item").SetGroup(g2).SaveX(ctx)

	elec := db.Tag.Create().SetName("Electronics").SetGroup(g1).SaveX(ctx)
	tools := db.Tag.Create().SetName("Tools").SetGroup(g1).SaveX(ctx)
	elec2 := db.Tag.Create().SetName("Electronics").SetGroup(g2).SaveX(ctx)

	newItem := func(g *ent.Group, et *ent.EntityType, name string) *ent.EntityCreate {
		return db.Entity.Create().SetName(name).SetGroup(g).SetEntityType(et)
	}
	field := func(name, val string) *ent.EntityField {
		return db.EntityField.Create().SetName(name).SetType("text").SetTextValue(val).SaveX(ctx)
	}

	newItem(g1, et1, "Phone").AddTag(elec).AddFields(field("Condition", "Clean")).SaveX(ctx)
	newItem(g1, et1, "Laptop").AddTag(elec).AddFields(field("Condition", "Clean"), field("Color", "Red")).SaveX(ctx)
	newItem(g1, et1, "Hammer").AddTag(tools).AddFields(field("Condition", "Dirty")).SaveX(ctx)
	// g2 carries the same tag/field names to prove scoping isolates groups
	newItem(g2, et2, "Tablet").AddTag(elec2).AddFields(field("Condition", "Clean")).SaveX(ctx)

	t.Run("tag facets ranked by entity count", func(t *testing.T) {
		tags, err := e.SearchTags(ctx, g1.ID, "")
		require.NoError(t, err)
		require.Len(t, tags, 2)
		assert.Equal(t, TagFacet{Name: "Electronics", Count: 2}, tags[0])
		assert.Equal(t, TagFacet{Name: "Tools", Count: 1}, tags[1])
	})

	t.Run("tag facet narrowing is scoped and case-insensitive", func(t *testing.T) {
		tags, err := e.SearchTags(ctx, g1.ID, "elec")
		require.NoError(t, err)
		require.Len(t, tags, 1)
		// 2, not 3: g2's Electronics tag must not be counted
		assert.Equal(t, TagFacet{Name: "Electronics", Count: 2}, tags[0])
	})

	t.Run("field facets discovery", func(t *testing.T) {
		facets, err := e.FieldFacets(ctx, g1.ID)
		require.NoError(t, err)
		require.Contains(t, facets, "Condition")
		require.Contains(t, facets, "Color", "each field is its own facet")

		counts := map[string]int{}
		for _, f := range facets["Condition"] {
			counts[f.Value] = f.Count
		}
		assert.Equal(t, 2, counts["Clean"])
		assert.Equal(t, 1, counts["Dirty"])
	})

	t.Run("field value narrowing and scoping", func(t *testing.T) {
		vals, err := e.SearchFieldValues(ctx, g1.ID, "Condition", "cle")
		require.NoError(t, err)
		require.Len(t, vals, 1)
		assert.Equal(t, FieldFacet{Value: "Clean", Count: 2}, vals[0])

		g2vals, err := e.SearchFieldValues(ctx, g2.ID, "Condition", "")
		require.NoError(t, err)
		require.Len(t, g2vals, 1)
		assert.Equal(t, FieldFacet{Value: "Clean", Count: 1}, g2vals[0], "facets must be scoped to the group")
	})
}

// newTestPostgresClient connects an ent client to the PostgreSQL instance in
// TEST_POSTGRES_URL (skipping when unset), using the same pgx driver the app
// uses, and creates the schema. Run one with:
//
//	docker run -d --rm -p 5433:5432 -e POSTGRES_PASSWORD=pw postgres
//	TEST_POSTGRES_URL=postgres://postgres:pw@localhost:5433/postgres?sslmode=disable go test ./internal/data/search/
func newTestPostgresClient(t *testing.T) *ent.Client {
	t.Helper()
	dsn := os.Getenv("TEST_POSTGRES_URL")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_URL not set; skipping Postgres integration test")
	}
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Schema.Create(context.Background()))
	return client
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
