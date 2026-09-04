package search

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	_ "github.com/sysadminsmedia/homebox/backend/pkgs/cgofreesqlite"
)

// meiliTestConfig returns the Meilisearch connection settings for integration
// tests, skipping the test when no instance is configured. Run one with:
//
//	docker run -d --rm -p 7700:7700 -e MEILI_MASTER_KEY=test-master-key getmeili/meilisearch
//	TEST_MEILISEARCH_URL=http://localhost:7700 TEST_MEILISEARCH_KEY=test-master-key go test ./internal/data/search/
func meiliTestConfig(t *testing.T) config.MeilisearchConf {
	t.Helper()
	host := os.Getenv("TEST_MEILISEARCH_URL")
	if host == "" {
		t.Skip("TEST_MEILISEARCH_URL not set; skipping Meilisearch integration test")
	}
	return config.MeilisearchConf{
		Host:   host,
		APIKey: os.Getenv("TEST_MEILISEARCH_KEY"),
		// unique index per run so concurrent/repeated runs don't interfere
		Index:   "homebox_test_" + uuid.NewString(),
		MaxHits: 1000,
	}
}

func newTestEntClient(t *testing.T) *ent.Client {
	t.Helper()
	client, err := ent.Open("sqlite3", "file:"+uuid.NewString()+"?mode=memory&cache=shared&_fk=1&_time_format=sqlite")
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Schema.Create(context.Background()))
	return client
}

func TestMeilisearchEngine_Integration(t *testing.T) {
	cfg := meiliTestConfig(t)
	db := newTestEntClient(t)
	ctx := context.Background()

	g1 := db.Group.Create().SetName("group-one").SaveX(ctx)
	g2 := db.Group.Create().SetName("group-two").SaveX(ctx)
	et1 := db.EntityType.Create().SetName("Item").SetGroup(g1).SaveX(ctx)
	et2 := db.EntityType.Create().SetName("Item").SetGroup(g2).SaveX(ctx)

	electronicsTag := db.Tag.Create().SetName("Электроника").SetGroup(g1).SaveX(ctx)
	imeiField := db.EntityField.Create().SetName("IMEI").SetType("text").SetTextValue("351234567891011").SaveX(ctx)

	newItem := func(g *ent.Group, et *ent.EntityType, name string) *ent.EntityCreate {
		return db.Entity.Create().SetName(name).SetGroup(g).SetEntityType(et)
	}

	ukrainian := newItem(g1, et1, "Тестовий Запис").SaveX(ctx)
	greek := newItem(g1, et1, "Υπολογιστής").SaveX(ctx)
	tagged := newItem(g1, et1, "Tagged item").AddTag(electronicsTag).SaveX(ctx)
	phone := newItem(g1, et1, "Smartphone").AddFields(imeiField).SaveX(ctx)
	toolbox := newItem(g1, et1, "Red Tool Box").SaveX(ctx)
	foreign := newItem(g2, et2, "Тестовий Запис").SaveX(ctx)

	engine, err := NewMeilisearchEngine(cfg, db, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = engine.client.DeleteIndex(cfg.Index) })

	require.NoError(t, engine.ReindexAll(ctx))

	// search applies only the engine predicate (no extra group filter) so the
	// assertions also verify Meilisearch-side group scoping
	search := func(q string) map[uuid.UUID]bool {
		t.Helper()
		pred, err := engine.Predicate(ctx, g1.ID, q)
		require.NoError(t, err)
		require.NotNil(t, pred)
		rows, err := db.Entity.Query().Where(pred).All(ctx)
		require.NoError(t, err)
		ids := make(map[uuid.UUID]bool, len(rows))
		for _, r := range rows {
			ids[r.ID] = true
		}
		return ids
	}

	t.Run("unicode case-insensitive", func(t *testing.T) {
		assert.True(t, search("тест")[ukrainian.ID], "lowercase Cyrillic query should match uppercase name")
		assert.True(t, search("ТЕСТОВИЙ")[ukrainian.ID])
		assert.True(t, search("υπολογιστής")[greek.ID])
		assert.True(t, search("ΥΠΟΛΟΓΙΣΤΗΣ")[greek.ID])
	})

	t.Run("group scoping", func(t *testing.T) {
		ids := search("тестовий")
		assert.True(t, ids[ukrainian.ID])
		assert.False(t, ids[foreign.ID], "results must be scoped to the queried group")
	})

	t.Run("tag names searchable", func(t *testing.T) {
		ids := search("электроника")
		assert.True(t, ids[tagged.ID])
		assert.False(t, ids[ukrainian.ID])
	})

	t.Run("custom field values searchable", func(t *testing.T) {
		assert.True(t, search("351234567891011")[phone.ID])
		// field names are stored for inspectability but intentionally not
		// searchable, matching the database engine
		assert.False(t, search("IMEI")[phone.ID])
	})

	t.Run("multi-word requires all terms", func(t *testing.T) {
		assert.True(t, search("box red")[toolbox.ID], "word order should not matter")
		assert.False(t, search("red hammer")[toolbox.ID], "all terms must match")
	})

	t.Run("typo tolerance", func(t *testing.T) {
		assert.True(t, search("smartphnoe")[phone.ID], "single-word typo should still match")
	})

	t.Run("incremental reindex adds new entities", func(t *testing.T) {
		bicycle := newItem(g1, et1, "Blue Bicycle").SaveX(ctx)
		require.NoError(t, engine.ReindexGroup(ctx, g1.ID))
		assert.True(t, search("bicycle")[bicycle.ID])
	})

	t.Run("reindex prunes deleted entities", func(t *testing.T) {
		require.True(t, search("tool box")[toolbox.ID])
		db.Entity.DeleteOneID(toolbox.ID).ExecX(ctx)
		require.NoError(t, engine.ReindexGroup(ctx, g1.ID))

		// assert against the raw index: the document itself must be gone
		// (going through the predicate + DB would hide staleness, since the
		// deleted row can never be selected anyway)
		resp, err := engine.index.SearchWithContext(ctx, "tool box", &meilisearch.SearchRequest{Limit: 100})
		require.NoError(t, err)
		for _, hit := range resp.Hits {
			var doc struct {
				ID string `json:"id"`
			}
			require.NoError(t, hit.DecodeInto(&doc))
			assert.NotEqual(t, toolbox.ID.String(), doc.ID, "deleted entity's document should be pruned from the index")
		}
	})

	t.Run("empty query yields no predicate", func(t *testing.T) {
		pred, err := engine.Predicate(ctx, g1.ID, "   ")
		require.NoError(t, err)
		assert.Nil(t, pred)
	})

	t.Run("tag facet search", func(t *testing.T) {
		// the only tagged entity in g1 carries "Электроника"
		facets, err := engine.SearchTags(ctx, g1.ID, "")
		require.NoError(t, err)
		byName := make(map[string]int, len(facets))
		for _, f := range facets {
			byName[f.Name] = f.Count
		}
		assert.Equal(t, 1, byName["Электроника"], "facet should report the tag and its entity count")

		// facetQuery narrows by tag name (case-insensitive substring)
		filtered, err := engine.SearchTags(ctx, g1.ID, "электро")
		require.NoError(t, err)
		require.Len(t, filtered, 1)
		assert.Equal(t, "Электроника", filtered[0].Name)

		// the tag belongs to g1, so g2's facets must not include it
		other, err := engine.SearchTags(ctx, g2.ID, "")
		require.NoError(t, err)
		for _, f := range other {
			assert.NotEqual(t, "Электроника", f.Name, "facets must be scoped to the group")
		}
	})

	t.Run("custom field facets", func(t *testing.T) {
		// three entities sharing one field name ("Condition") with two values,
		// plus a field name with a space to exercise nested facet attributes
		clean1 := db.EntityField.Create().SetName("Condition").SetType("text").SetTextValue("Clean").SaveX(ctx)
		clean2 := db.EntityField.Create().SetName("Condition").SetType("text").SetTextValue("Clean").SaveX(ctx)
		dirty := db.EntityField.Create().SetName("Condition").SetType("text").SetTextValue("Dirty").SaveX(ctx)
		special := db.EntityField.Create().SetName("Special Field").SetType("text").SetTextValue("Clean").SaveX(ctx)
		newItem(g1, et1, "Sofa").AddFields(clean1).SaveX(ctx)
		newItem(g1, et1, "Rug").AddFields(clean2, special).SaveX(ctx)
		newItem(g1, et1, "Doormat").AddFields(dirty).SaveX(ctx)
		require.NoError(t, engine.ReindexGroup(ctx, g1.ID))

		// discovery: each field is its own facet with per-value counts
		facets, err := engine.FieldFacets(ctx, g1.ID)
		require.NoError(t, err)
		require.Contains(t, facets, "Condition")
		require.Contains(t, facets, "Special Field", "field names with spaces are faceted")
		require.Contains(t, facets, "IMEI", "fields are independent of one another")

		counts := map[string]int{}
		for _, f := range facets["Condition"] {
			counts[f.Value] = f.Count
		}
		assert.Equal(t, 2, counts["Clean"])
		assert.Equal(t, 1, counts["Dirty"])

		// per-field value autocomplete, scoped and narrowed by query
		vals, err := engine.SearchFieldValues(ctx, g1.ID, "Condition", "cle")
		require.NoError(t, err)
		require.Len(t, vals, 1)
		assert.Equal(t, "Clean", vals[0].Value)
		assert.Equal(t, 2, vals[0].Count)

		// g2 has none of these fields
		g2facets, err := engine.FieldFacets(ctx, g2.ID)
		require.NoError(t, err)
		assert.NotContains(t, g2facets, "Condition", "facets must be scoped to the group")
	})
}

func TestMeilisearchEngine_EventDrivenReindex(t *testing.T) {
	cfg := meiliTestConfig(t)
	db := newTestEntClient(t)
	ctx := context.Background()

	oldDebounce := meiliReindexDebounce
	meiliReindexDebounce = 100 * time.Millisecond
	t.Cleanup(func() { meiliReindexDebounce = oldDebounce })

	bus := eventbus.New()
	busCtx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	go func() { _ = bus.Run(busCtx) }()

	g := db.Group.Create().SetName("group-bus").SaveX(ctx)
	et := db.EntityType.Create().SetName("Item").SetGroup(g).SaveX(ctx)

	engine, err := NewMeilisearchEngine(cfg, db, bus)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = engine.client.DeleteIndex(cfg.Index) })

	// created after engine startup, so only the mutation event can index it
	lamp := db.Entity.Create().SetName("Vintage Lamp").SetGroup(g).SetEntityType(et).SaveX(ctx)
	bus.Publish(eventbus.EventEntityMutation, eventbus.GroupMutationEvent{GID: g.ID})

	assert.Eventually(t, func() bool {
		pred, err := engine.Predicate(ctx, g.ID, "vintage lamp")
		if err != nil || pred == nil {
			return false
		}
		ids, err := db.Entity.Query().Where(pred).IDs(ctx)
		return err == nil && len(ids) == 1 && ids[0] == lamp.ID
	}, 15*time.Second, 200*time.Millisecond, "mutation event should trigger a debounced group reindex")
}

func TestNewEngine_UnknownDriver(t *testing.T) {
	_, err := NewEngine(config.SearchConf{Driver: "sphinx"}, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported search driver")
}

func TestNewEngine_MeilisearchUnreachable(t *testing.T) {
	_, err := NewEngine(config.SearchConf{
		Driver:      DriverMeilisearch,
		Meilisearch: config.MeilisearchConf{Host: "http://127.0.0.1:1", MaxHits: 10},
	}, nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not reachable")
}
