# Batch Print Sort Order and Asset ID Filter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let the user sort the batch print list, print in that order, and filter items by asset ID comparisons such as `#>=100`.

**Architecture:** The server already sorts; the batch page just never asks it to. R1 adds the controls and a pure helper that re-orders the selection, plus a stable tie-break in the repository. R3 adds one parser in `repo`, used by both places that read the search term today (the entities handler and the customs service), and a range filter in `QueryByGroup`.

**Tech Stack:** Go, ent; Nuxt 4, Vue 3, vitest, Playwright.

**Spec:** [`docs/superpowers/specs/2026-11-03-batch-print-sort-and-asset-filter-design.md`](../specs/2026-11-03-batch-print-sort-and-asset-filter-design.md)

## Global Constraints

- Never hand-edit `frontend/lib/api/types` or generated docs. After changing swaggo annotations, run `task --force generate` (plain `task generate` can report "up to date") and revert the unrelated `backend/go.sum` changes it makes.
- All UI text goes into `frontend/locales/en.json` and `de.json`, with the exact copy from spec §3.
- `task go:lint`, `task ui:fix` (at most 1 warning in `lint:ci`), `task ui:check`: clean before every PR.
- For a scoped Go run use `cd backend && go test <pkg>`; `task go:test` always runs every package.
- Tests never use the local dev database.
- **PR slicing.** Two PRs:
  - PR A: Tasks 1–2 (R1, R2), against `main`;
  - PR B: Tasks 3–4 (R3), stacked on PR A, because both change `batch.vue`.
- PR bodies follow the repo's PR conventions (persona first line, signature).

## Review Focus

1. **A search or filter change after picking a custom order.** Expected: the order survives; only a sort change replaces it. Test: Task 2, e2e steps 5–7.
2. **Selected items filtered out of the list, then a sort change.** Expected: they stay selected, after the listed ones, in their old order. Test: Task 2, `keeps selected items outside the list at the end`.
3. **`#<1` or `#>=0`.** Expected: no item without an asset ID ever matches, and `#>=0` isn't "no filter". Test: Task 3, `TestParseAssetIDSearch` and `TestEntityRepository_QueryByGroup_AssetIDRange`.
4. **The customs export from a search of `#>=100`.** Expected: the export covers the same items as the search page, not a text search for that string. Test: Task 3, `TestToEntityQuery_AssetIDRange`.
5. **Many items with the same name, across pages.** Expected: no row repeats or goes missing between pages. Test: Task 1, `TestEntityRepository_QueryByGroup_TiesOrderByID`.

---

### Task 1: Stable order for ties (FR-6)

**Files:**
- Modify: `backend/internal/data/repo/repo_entities.go` (`applyEntityOrder`)
- Test: `backend/internal/data/repo/repo_entities_test.go`

**Interfaces:**
- Produces: every `QueryByGroup` result ordered by the chosen field, then `entity.FieldID` ascending.

- [ ] **Step 1: Write the failing test**

`TestEntityRepository_QueryByGroup_TiesOrderByID`: create five items in one location, all named `"Same name"` and with equal quantity. Query with `OrderBy` set to `""` (name), `"quantity"` and `"location"`, `Page: -1, PageSize: -1`. Each time assert the returned IDs are in ascending `ID.String()` order. UUIDs are random, so insertion order won't pass by chance.

- [ ] **Step 2: Run it and see it fail**

Run: `cd backend && go test ./internal/data/repo/ -run TiesOrderByID`
Expected: FAIL, the IDs aren't sorted.

- [ ] **Step 3: Implement**

In `applyEntityOrder`, append `ent.Asc(entity.FieldID)` as the last order term on all three return paths, including the `location` one. Direction of the main field stays as it is; the tie-break is always ascending.

- [ ] **Step 4: Run the repository tests and see them pass**

Run: `cd backend && go test ./internal/data/repo/`
Expected: PASS, including `OrderByLocation` and `TimestampOrderDefaultsDesc`.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/data/repo
git commit -m "fix(repo): break entity order ties by ID"
```

### Task 2: Sort controls, selection order and limit message (FR-1 to FR-5, FR-7)

**Files:**
- Create: `frontend/lib/label-batch.ts`, `frontend/lib/label-batch.test.ts`
- Modify: `frontend/pages/label-templates/batch.vue`, `frontend/locales/en.json`, `frontend/locales/de.json`
- Test: create `frontend/test/e2e/label-batch.browser.spec.ts`

**Interfaces:**
- Produces:
  - `export function orderSelection(selectedIds: string[], resultIds: string[]): string[]`
  - `export const BATCH_ORDER_FIELDS` — the FR-2 values, in spec order: `assetId, name, location, createdAt, updatedAt, quantity, purchasePrice, insured`

- [ ] **Step 1: Write the failing unit tests**

```ts
describe("orderSelection", () => {
  test("follows the list order", () => {
    expect(orderSelection(["c", "a", "b"], ["a", "b", "c", "d"])).toEqual(["a", "b", "c"]);
  });
  test("keeps selected items outside the list at the end", () => {
    expect(orderSelection(["x", "c", "y", "a"], ["a", "b", "c"])).toEqual(["a", "c", "x", "y"]);
  });
  test("keeps the order when nothing was reordered", () => {
    expect(orderSelection(["a", "b"], ["a", "b"])).toEqual(["a", "b"]);
  });
  test("handles an empty selection", () => {
    expect(orderSelection([], ["a"])).toEqual([]);
  });
});
```

- [ ] **Step 2: Write the failing e2e spec**

`label-batch.browser.spec.ts`, with the API `login` helper from `customs-inventory.browser.spec.ts`:
1. `goto("/label-templates/batch")`, wait for the first row.
2. Read the `#…` asset IDs of the rows and assert they are in ascending order (FR-3).
3. Choose "Descending" in the direction select. Wait for the `/api/v1/entities` response with `orderDirection=desc`, and assert the IDs are in descending order.
4. Choose "Ascending" again and wait for the response.
5. Tick the second row, then the first, so the order badges read 2 on the first row and 1 on the second.
6. Type the first row's exact name into the search box and wait for the response, then clear the box and wait again. Assert the badges are unchanged.
7. Choose "Descending" and wait for the response. Assert the badges now follow the list: the row higher up has 1 (FR-5).
8. Click "Locations" and assert the order-by select is gone (FR-1).

- [ ] **Step 3: Run them and see them fail**

Run:
- `cd frontend && pnpm exec vitest run --config ./test/vitest.config.ts lib/label-batch.test.ts`
- `task test:e2e -- --project=chromium label-batch`

Expected: FAIL. `orderSelection` doesn't exist, and the rows aren't in asset ID order.

- [ ] **Step 4: Implement**

- `label-batch.ts`: the helper and the constant.
- `batch.vue`:
  - `const orderBy = ref("assetId")`, `const orderDirection = ref("asc")`, sent in the `getAll` call; `total` from `data.total`.
  - Two `Select`s in the filter row, inside the existing `v-if="kind === 'items'"`, with `items.order_by` as the first one's accessible label. Option labels reuse the `items.*` keys the search page uses (`items.asset_id`, `items.name`, `items.location`, `items.created_at`, `items.updated_at`, `items.quantity`, `items.purchase_price`, `items.insured`, `items.asc`, `items.desc`).
  - A separate `watch([orderBy, orderDirection], …)`: await `search()`, then `selectedIds.value = orderSelection(selectedIds.value, results.value.map(r => r.id))`. Don't add these refs to the existing `watchDebounced`, which must keep leaving the selection alone.
  - Under the results line, show `results_truncated` with `{ shown: results.length, total }` when `kind === 'items' && total > results.length`.
- Locales: `results_truncated` in en and de. `de.json` has no `pages.label_templates` block, so create it with this key. Add `items.asc` and `items.desc` to `de.json` (spec §3).

- [ ] **Step 5: Run the tests and checks**

Run: the commands from Step 3, then `task ui:fix`, `task ui:check`.
Expected: PASS, with no new lint warnings and no type errors.

- [ ] **Step 6: Commit and open PR A**

```bash
git add frontend
git commit -m "feat(labels): sort the batch print list and print in that order"
```

PR A goes against `main`. In the body, name R1, R2 and the spec.

### Task 3: Asset ID comparisons in the entity search (FR-8 to FR-13)

**Files:**
- Modify: `backend/internal/data/repo/asset_id_type.go`, `backend/internal/data/repo/repo_entities.go`, `backend/app/api/handlers/v1/v1_ctrl_entities.go`, `backend/internal/core/services/service_customs_inventory.go`
- Test: `backend/internal/data/repo/asset_id_type_test.go`, `backend/internal/data/repo/repo_entities_test.go`, `backend/internal/core/services/service_customs_inventory_test.go`
- Generated: `backend/app/api/static/docs/*`, `docs/public/api/*`, `frontend/lib/api/types/*`

**Interfaces:**
- Produces:
  - `type AssetIDRange struct { Min, Max AssetID }`, both inclusive
  - `func ParseAssetIDSearch(s string) (AssetIDRange, bool)`
  - `EntityQuery.AssetIDRange *AssetIDRange`
  - `func (q *EntityQuery) ResolveAssetIDSearch()`: turns `q.Search` into `q.AssetID` or `q.AssetIDRange` and clears `q.Search`; otherwise leaves `q` alone

- [ ] **Step 1: Write the failing tests**

```go
func TestParseAssetIDSearch(t *testing.T) {
	const top = AssetID(math.MaxInt64)
	valid := map[string]AssetIDRange{
		"#>=100":              {100, top},
		"#>= 100":             {100, top},
		"# >= 000-100":        {100, top},
		"asset-id >= 000-100": {100, top},
		"Asset ID>=100":       {100, top},
		"assetid >= 100":      {100, top},
		"#>100":               {101, top},
		"#<=200":              {1, 200},
		"#<200":               {1, 199},
		"#>=100 <=200":        {100, 200},
		"#<=200 >=100":        {100, 200},
		"#>=0":                {1, top},
		"#<1":                 {1, 0},
		"#>=200 <=100":        {200, 100},
	}
	for in, want := range valid {
		got, ok := ParseAssetIDSearch(in)
		assert.True(t, ok, in)
		assert.Equal(t, want, got, in)
	}
	for _, in := range []string{"", "#", "#000-123", "#abc", "#>=", "#>=abc", "#>=100 books", "books #>=100", "asset-id", "asset-id 100", "#>=1 <=2 <=3"} {
		_, ok := ParseAssetIDSearch(in)
		assert.False(t, ok, in)
	}
}

func TestEntityQuery_ResolveAssetIDSearch(t *testing.T) {
	exact := EntityQuery{Search: "#000-123"}
	exact.ResolveAssetIDSearch()
	assert.Equal(t, AssetID(123), exact.AssetID)
	assert.Empty(t, exact.Search)
	assert.Nil(t, exact.AssetIDRange)

	ranged := EntityQuery{Search: "#>=100"}
	ranged.ResolveAssetIDSearch()
	require.NotNil(t, ranged.AssetIDRange)
	assert.Empty(t, ranged.Search)

	text := EntityQuery{Search: "#abc"}
	text.ResolveAssetIDSearch()
	assert.Equal(t, "#abc", text.Search)
	assert.Nil(t, text.AssetIDRange)
}
```

`TestEntityRepository_QueryByGroup_AssetIDRange`: items with asset IDs 5, 10, 15 and one with none (0). Assert the returned asset IDs for `{10, MaxInt64}` → 10, 15; `{5, 10}` → 5, 10; `{1, MaxInt64}` → 5, 10, 15; `{1, 0}` → none; `{15, 5}` → none, without an error.

`TestToEntityQuery_AssetIDRange` in the service test file: `toEntityQuery(customsinventory.ScopeQuery{Q: "#>=100"})` has an empty `Search` and `AssetIDRange` `{100, MaxInt64}`. The existing `#000-123` assertion at `service_customs_inventory_test.go:302` must keep passing.

- [ ] **Step 2: Run them and see them fail**

Run: `cd backend && go test ./internal/data/repo/ ./internal/core/services/ -run 'AssetIDSearch|AssetIDRange'`
Expected: FAIL, `undefined: ParseAssetIDSearch`.

- [ ] **Step 3: Implement**

- `ParseAssetIDSearch`, in `asset_id_type.go`: match the trimmed term case-insensitively against a package-level regexp:

  ```go
  `(?i)^(?:#|asset[- ]?id)\s*(>=|<=|>|<)\s*([0-9-]+)(?:\s+(>=|<=|>|<)\s*([0-9-]+))?$`
  ```

  Start from `AssetIDRange{1, math.MaxInt64}`. Read each value with `ParseAssetID`; a failure means `false`. Apply each comparison as a tightening: `>=n` raises `Min` to `n`, `>n` to `n+1`, `<=n` lowers `Max` to `n`, `<n` to `n-1`, keeping `Min` at 1 or more. Those tightenings give the test table's results.
- `ResolveAssetIDSearch`: the exact `#` handling moved here from the handler, then `ParseAssetIDSearch`.
- `QueryByGroup`: next to the `q.AssetID` filter, add `entity.AssetIDGTE` and `entity.AssetIDLTE` when `q.AssetIDRange != nil`.
- Handler `extractQuery` and customs `toEntityQuery`: replace their `#` blocks with `v.ResolveAssetIDSearch()`. Update the comment above `toEntityQuery`.
- Handler `@Param q` annotation: after "Use #<assetId> to look up by asset ID", add "#>=100, #<=200 or #>=100 <=200 to filter by asset ID". Then run `task --force generate` and revert `backend/go.sum`.

- [ ] **Step 4: Run the tests and see them pass**

Run: `cd backend && go test ./internal/data/repo/ ./internal/core/services/... ./app/api/handlers/v1/...`, then `task go:lint`.
Expected: PASS, 0 issues.

- [ ] **Step 5: Commit**

```bash
git add backend docs/public/api frontend/lib/api/types
git commit -m "feat(search): filter entities by asset ID comparisons"
```

### Task 4: Asset ID filter in the UI (FR-14, FR-15)

**Files:**
- Modify: `frontend/pages/label-templates/batch.vue`, `frontend/pages/search.vue`, `frontend/lib/label-batch.ts`, `frontend/locales/en.json`, `frontend/locales/de.json`
- Test: `frontend/lib/label-batch.test.ts`, `frontend/test/e2e/label-batch.browser.spec.ts`

**Interfaces:**
- Produces: `export function isExactAssetIdSearch(q: string): boolean`: true for `#` followed by digits and dashes only

- [ ] **Step 1: Write the failing tests**

Unit: `isExactAssetIdSearch` is true for `#000-123` and `#123`; false for `#>=100`, `#>=100 <=200`, `#abc`, `#`, `books`.

E2E, added to `label-batch.browser.spec.ts`:
1. Assert the hint text is visible.
2. Note the third row's asset ID, type `#>=<that ID>` into the search box, and wait for the `/api/v1/entities` response.
3. Assert every listed ID is at or above it, and the first row is that ID.
4. Type `#>=200000 <=100` and assert the "No items found" text appears.

- [ ] **Step 2: Run them and see them fail**

Run: the vitest and `task test:e2e -- --project=chromium label-batch` commands.
Expected: FAIL, the helper and the hint don't exist. The filter steps pass already if Task 3 is in.

- [ ] **Step 3: Implement**

- `batch.vue`: render `asset_filter_hint` under the filter row, only when `kind === 'items'`.
- `search.vue`: `byAssetId` uses `isExactAssetIdSearch(query.value ?? "")`, so the "Query ID" line stays away for comparisons. A term like `#abc` no longer shows "invalid"; `parsedAssetId` keeps its check for values that fit the pattern but don't parse.
- Locales: `asset_filter_hint` in en and de, in the block Task 2 added.

- [ ] **Step 4: Run the checks**

Run: `task ui:fix`, `task ui:check`, then `task test:e2e -- label-batch customs-inventory`.
Expected: clean; both specs pass in chromium and webkit. Firefox doesn't launch on the local machine, so it's left to CI.

- [ ] **Step 5: Commit and open PR B**

```bash
git add frontend
git commit -m "feat(labels): explain the asset ID filter on the batch print page"
```

PR B is stacked on PR A. In the body, name R3 and the spec, and say that CI runs once it retargets to `main`.
