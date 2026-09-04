package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// useSearchableItem creates an item with the given name and applies an
// optional full update so tests can populate any searchable field.
func useSearchableItem(t *testing.T, name string, mutate func(u *EntityUpdate)) EntityOut {
	t.Helper()
	ctx := context.Background()
	itemET := useItemEntityType(t)

	e, err := tRepos.Entities.Create(ctx, tGroup.ID, EntityCreate{
		Name:         name,
		EntityTypeID: itemET.ID,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = tRepos.Entities.Delete(context.Background(), e.ID) })

	if mutate != nil {
		u := EntityUpdate{
			ID:           e.ID,
			Name:         name,
			Quantity:     1,
			EntityTypeID: itemET.ID,
		}
		mutate(&u)
		_, err = tRepos.Entities.UpdateByGroup(ctx, tGroup.ID, u)
		require.NoError(t, err)
	}
	return e
}

// searchIDs runs a search query and returns the set of matched entity IDs.
// The test group is shared across the package, so assertions check membership
// instead of exact result counts.
func searchIDs(t *testing.T, q EntityQuery) map[uuid.UUID]bool {
	t.Helper()
	q.Page, q.PageSize = -1, -1
	res, err := tRepos.Entities.QueryByGroup(context.Background(), tGroup.ID, q)
	require.NoError(t, err)

	ids := make(map[uuid.UUID]bool, len(res.Items))
	for _, item := range res.Items {
		ids[item.ID] = true
	}
	return ids
}

func assertSearchFinds(t *testing.T, query string, item EntityOut, want bool) {
	t.Helper()
	found := searchIDs(t, EntityQuery{Search: query})[item.ID]
	if want {
		assert.True(t, found, "search %q should find item %q", query, item.Name)
	} else {
		assert.False(t, found, "search %q should NOT find item %q", query, item.Name)
	}
}

func TestEntitySearch_UnicodeCaseInsensitive(t *testing.T) {
	ukrainian := useSearchableItem(t, "Тестовий Запис", nil)
	greek := useSearchableItem(t, "Υπολογιστής", nil)

	// Cyrillic: lowercase, uppercase, and partial queries must match
	// uppercase stored text (issue #1021).
	assertSearchFinds(t, "тест", ukrainian, true)
	assertSearchFinds(t, "ТЕСТ", ukrainian, true)
	assertSearchFinds(t, "тестовий запис", ukrainian, true)
	assertSearchFinds(t, "запис", ukrainian, true)

	// Greek, including the final-sigma form difference (issue #1367).
	assertSearchFinds(t, "Υπολογιστής", greek, true)
	assertSearchFinds(t, "υπολογιστής", greek, true)
	assertSearchFinds(t, "ΥΠΟΛΟΓΙΣΤΗΣ", greek, true)
	assertSearchFinds(t, "υπολογιστης", greek, true)

	assertSearchFinds(t, "холодильник", ukrainian, false)
}

func TestEntitySearch_AccentInsensitive(t *testing.T) {
	accented := useSearchableItem(t, "Electrónica de café", nil)
	plain := useSearchableItem(t, "electronica cafe pere", nil)

	assertSearchFinds(t, "electronica", accented, true)
	assertSearchFinds(t, "café", accented, true)
	assertSearchFinds(t, "CAFE", accented, true)
	assertSearchFinds(t, "electrónica", plain, true)
	assertSearchFinds(t, "père", plain, true)
}

func TestEntitySearch_MultiTokenAnd(t *testing.T) {
	item := useSearchableItem(t, "Red Tool Box", nil)

	// every token must match, in any order
	assertSearchFinds(t, "box red", item, true)
	assertSearchFinds(t, "red tool", item, true)
	assertSearchFinds(t, "red hammer", item, false)

	// quoted phrases match as a unit
	assertSearchFinds(t, `"tool box"`, item, true)
	assertSearchFinds(t, `"box tool"`, item, false)
}

func TestEntitySearch_MatchesAcrossFields(t *testing.T) {
	item := useSearchableItem(t, "Multifield", func(u *EntityUpdate) {
		u.SerialNumber = "SN-998877"
		u.ModelNumber = "MX-1000"
		u.Manufacturer = "Acme Corp"
		u.Notes = "stored in the attic"
		u.PurchaseFrom = "Conrad Electronic"
	})

	assertSearchFinds(t, "998877", item, true)
	assertSearchFinds(t, "mx-1000", item, true)
	assertSearchFinds(t, "acme", item, true)
	assertSearchFinds(t, "attic", item, true)
	assertSearchFinds(t, "conrad", item, true)

	// tokens may match across different fields of the same item
	assertSearchFinds(t, "acme attic", item, true)
}

func TestEntitySearch_MatchesTagNames(t *testing.T) {
	ctx := context.Background()

	tagOut, err := tRepos.Tags.Create(ctx, tGroup.ID, TagCreate{Name: "Электроника-поиск"})
	require.NoError(t, err)
	t.Cleanup(func() { _ = tRepos.Tags.delete(context.Background(), tagOut.ID) })

	itemET := useItemEntityType(t)
	tagged, err := tRepos.Entities.Create(ctx, tGroup.ID, EntityCreate{
		Name:         "Tagged thing",
		EntityTypeID: itemET.ID,
		TagIDs:       []uuid.UUID{tagOut.ID},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = tRepos.Entities.Delete(context.Background(), tagged.ID) })

	untagged := useSearchableItem(t, "Untagged thing", nil)

	// tag names are searchable from the search bar (#1509), with the same
	// UTF-8 case folding as other fields
	assertSearchFinds(t, "электроника-поиск", tagged, true)
	assertSearchFinds(t, "электроника-поиск", untagged, false)
}

func TestEntitySearch_MatchesCustomFieldValues(t *testing.T) {
	item := useSearchableItem(t, "Phone", func(u *EntityUpdate) {
		u.Fields = []EntityFieldData{
			{Type: "text", Name: "IMEI", TextValue: "351234567891011"},
		}
	})
	other := useSearchableItem(t, "Other phone", nil)

	// custom field values are searchable from the search bar (#1380)
	assertSearchFinds(t, "351234567891011", item, true)
	assertSearchFinds(t, "3512345", item, true)
	assertSearchFinds(t, "351234567891011", other, false)
}

func TestEntitySearch_LikeWildcardsAreLiteral(t *testing.T) {
	percent := useSearchableItem(t, "100% cotton", nil)
	plain := useSearchableItem(t, "100x cotton", nil)

	assertSearchFinds(t, "100%", percent, true)
	assertSearchFinds(t, "100%", plain, false)

	underscore := useSearchableItem(t, "a_b pattern", nil)
	noUnderscore := useSearchableItem(t, "axb pattern", nil)

	assertSearchFinds(t, "a_b", underscore, true)
	assertSearchFinds(t, "a_b", noUnderscore, false)
}

func TestQueryByGroup_MatchAllTags(t *testing.T) {
	ctx := context.Background()
	tags := useTags(t, 2)

	itemET := useItemEntityType(t)
	both, err := tRepos.Entities.Create(ctx, tGroup.ID, EntityCreate{
		Name:         "Has both tags",
		EntityTypeID: itemET.ID,
		TagIDs:       []uuid.UUID{tags[0].ID, tags[1].ID},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = tRepos.Entities.Delete(context.Background(), both.ID) })

	one, err := tRepos.Entities.Create(ctx, tGroup.ID, EntityCreate{
		Name:         "Has one tag",
		EntityTypeID: itemET.ID,
		TagIDs:       []uuid.UUID{tags[0].ID},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = tRepos.Entities.Delete(context.Background(), one.ID) })

	tagIDs := []uuid.UUID{tags[0].ID, tags[1].ID}

	// default OR behavior: any selected tag matches
	anyMatch := searchIDs(t, EntityQuery{TagIDs: tagIDs})
	assert.True(t, anyMatch[both.ID], "OR mode should match item with both tags")
	assert.True(t, anyMatch[one.ID], "OR mode should match item with one tag")

	// matchAllTags: every selected tag must be present (#1454)
	allMatch := searchIDs(t, EntityQuery{TagIDs: tagIDs, MatchAllTags: true})
	assert.True(t, allMatch[both.ID], "AND mode should match item with both tags")
	assert.False(t, allMatch[one.ID], "AND mode should NOT match item with only one tag")
}
