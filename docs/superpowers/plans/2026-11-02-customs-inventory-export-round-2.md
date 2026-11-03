# Customs Inventory Export, Round 2: Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the four findings from live testing: piece counts (R4), first-word merge (R1), containers with their contents (R2), and a single entry point on the search page (R3).

**Architecture:** R1, R2 and R4 change the pure `customsinventory` package (`Build`, `Verify`, `WriteXLSX`). R2 also changes the service, which loads the contents. All three then add a dialog option or KPI. R3 is frontend only. The preview and the export keep sharing one `Build` call.

**Tech Stack:** Go 1.x, excelize v2, ent repos; Nuxt 4, Vue 3, vitest, Playwright.

**Spec:** [`docs/superpowers/specs/2026-11-02-customs-inventory-export-round-2-design.md`](../specs/2026-11-02-customs-inventory-export-round-2-design.md) (round 2). It builds on [`2026-10-31-customs-inventory-export-design.md`](../specs/2026-10-31-customs-inventory-export-design.md) (the first spec).

## Global Constraints

- Every sum is in integer cents (CALC-1 to CALC-3). Quantities are rounded with `RoundQty` (CALC-5).
- Text cells never become formulas (FR-16). New sheet text is written with `text(...)`, never as a formula.
- Never hand-edit `frontend/lib/api/types`. Run `task generate` after changing the API types.
- All UI text goes into `frontend/locales/en.json` and `de.json`, with the exact copy from round 2 §3.
- `task go:lint`, `task ui:fix` (at most 1 warning in `lint:ci`), `task ui:check`: clean before every PR.
- Tests never use the local dev database.
- **PR slicing.** Four PRs, each single-purpose, each opened against `main`. Pull Request CI only runs on PRs into `main`, so each PR is cut from `main` after the previous one is merged:
  - PR A: Tasks 1–2 (R4);
  - PR B: Tasks 3–4 (R1);
  - PR C: Tasks 5–8 (R2);
  - PR D: Task 9 (R3). It touches none of the other PRs' files, so it can be cut from `main` at any time.
- PR bodies follow the repo's PR conventions (persona first line, signature).

## Review Focus

1. **Bad parent data.** A cycle in the parent links (A → B → A) must not hang `Build`. Expected: the walk stops, and every entity is still accounted for once. Test: Task 5, `TestBuild_RollUp_ParentCycleTerminates`.
2. **A dutiable child that the search also matched.** It is in scope and has a duty-free container. Expected: it gets its own dutiable line once, never twice. Test: Task 5, `TestBuild_RollUp_ScopeChildSplitOnce`.
3. **Merged containers.** Two equal boxes, each with contents, merge into one line. Expected: each member's contents stay tied to that member, and (i) holds for both. Test: Task 5, `TestBuild_RollUp_MergedContainersKeepTheirContents`.
4. **Fractional quantities in piece sums**, such as 2.5 + 0.1 + 0.2. Expected: `2,8` in the text, 2.8 in column B, with no float noise. Test: Task 1, `TestFormatQtyDE` and `TestBuild_PieceCounts`.
5. **Stale selection on the search page.** Row selection is kept by row index. After a new search, the old selection must not carry over to different rows. Expected: the table clears its selection when its rows change. Test: Task 9, the e2e step "a new search clears the selection".

---

### Task 1: Piece counts and quantity sums in report, preview and sheet (R4, backend)

**Files:**
- Modify: `backend/internal/core/services/reporting/customsinventory/model.go`, `build.go`, `verify.go`, `money.go`, `xlsx.go`
- Test: `money_test.go`, `build_test.go`, `verify_test.go`, `xlsx_test.go`, `testdata/golden_report.json`

**Interfaces:**
- Produces:
  - `Group.SubtotalQty float64`
  - `Report.DutyFreeQty, DutiableQty, TotalQty float64`
  - `Preview.PieceCount float64` (`json:"pieceCount"`), `Preview.DutiablePieceCount float64` (`json:"dutiablePieceCount"`)
  - `func FormatQtyDE(q float64) string`
  - Verify invariant letter `(h)`

- [ ] **Step 1: Write the failing tests**

```go
func TestFormatQtyDE(t *testing.T) {
	a, b, c := 2.5, 0.1, 0.2 // variables, so the sums carry float noise (constants would be exact)
	cases := []struct {
		in   float64
		want string
	}{{7, "7"}, {a, "2,5"}, {b + c, "0,3"}, {a + b + c, "2,8"}, {1e-7, "0"}, {1234.5, "1234,5"}}
	for _, tc := range cases {
		assert.Equal(t, tc.want, FormatQtyDE(tc.in), "%v", tc.in)
	}
}

func TestBuild_PieceCounts(t *testing.T) {
	in := smallInput(
		furniture(1, 1, "Box books", 3, 10, "2019-01-01", locLiving, pathLiving),
		furniture(2, 2, "Cable", 2.5, 1, "2019-01-01", locLiving, pathLiving),
		furniture(3, 3, "Rope", 0.3, 1, "2019-01-01", locLiving, pathLiving),
		furniture(4, 4, "New lamp", 4, 20, "2026-10-01", locLiving, pathLiving), // dutiable
	)
	r := mustBuild(t, in)
	assert.InDelta(t, 5.8, r.DutyFreeQty, 0)
	assert.InDelta(t, 5.8, r.Groups[0].SubtotalQty, 0)
	assert.InDelta(t, 4.0, r.DutiableQty, 0)
	assert.InDelta(t, 9.8, r.TotalQty, 0)
	p := r.Preview()
	assert.InDelta(t, 9.8, p.PieceCount, 0)
	assert.InDelta(t, 4.0, p.DutiablePieceCount, 0)
}
```

Add to the `TestVerify_RejectsCorruption` table: `"invariant (h)": func(r *Report) { r.Groups[0].SubtotalQty++ }`, and a second case `r.TotalQty++`.

In `xlsx_test.go`:
- `TestWriteXLSX_LayoutAndLabels`: row 5 now contains `· {lines} Positionen / lines · {pieces} Stück / pieces` for the golden report, built from `FormatQtyDE(r.TotalQty)` and the line count. It no longer contains `Artikel / items`.
- New `TestWriteXLSX_QuantityTotals`: column B of the first subtotal row has the formula `SUBTOTAL(9,B{first}:B{last})` and the value `r.Groups[0].SubtotalQty`. The grand total row's B formula is `B{dutyFreeRow}+B{dutiableRow}`. With no dutiable lines, the dutiable total's B formula is `0`.
- Extend `TestWriteXLSX_FormulasRecalculateToCachedValues` so that every formula cell in column B also recalculates to its stored value.

- [ ] **Step 2: Run the tests and see them fail**

Run: `task go:test -- ./internal/core/services/reporting/customsinventory/...`
Expected: FAIL. The build reports `undefined: FormatQtyDE`, and the fields are missing.

- [ ] **Step 3: Implement**

- `FormatQtyDE` in `money.go`: `strconv.FormatFloat(RoundQty(q), 'f', -1, 64)`, then replace `.` with `,`.
- `Build`: accumulate `SubtotalQty` per group, and `DutyFreeQty`, `DutiableQty`, `TotalQty`. Round each sum with `RoundQty`.
- `Preview`: `PieceCount = r.TotalQty`, `DutiablePieceCount = r.DutiableQty`.
- `Verify` (h): compare with `RoundQty` on both sides. Messages: `group %d quantity doesn't match its lines`, `duty-free quantity doesn't match the groups`, `dutiable quantity doesn't match its lines`, `total quantity isn't duty-free + dutiable`.
- `xlsx.go`:
  - change `total` to `total(label string, qty float64, qtyFormula string, cents int64, formula string) int`, and write B as a bold cell with that formula and value;
  - in `section`, build the B formula the same way as the I formula;
  - for the duty-free total, use `SUM(B…)`, or the SUMIF fallback over column B;
  - for the grand total, use `B{df}+B{dt}`;
  - in row 5, replace the `%d Artikel / items` part with `%d Positionen / lines · %s Stück / pieces`, using the line count and `FormatQtyDE(r.TotalQty)`.
- Add a bold style for quantities with the same number format that column B lines use today. If lines use none, add none.

- [ ] **Step 4: Refresh the golden file and check the diff**

Run: `cd backend && go test ./internal/core/services/reporting/customsinventory/ -run TestBuild_Golden -update && git diff --stat testdata`
Expected: only the new `*Qty` fields change in `golden_report.json`.

- [ ] **Step 5: Run the package tests and see them pass**

Run: `task go:test -- ./internal/core/services/reporting/customsinventory/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/core/services/reporting/customsinventory
git commit -m "feat(customs): count pieces and sum quantities in the sheet's total rows"
```

### Task 2: Show pieces in the dialog (R4, frontend)

**Files:**
- Modify: `frontend/lib/customs-inventory.ts`, `frontend/components/Reports/CustomsInventoryDialog.vue`, `frontend/locales/en.json`, `frontend/locales/de.json`
- Generated: `backend/app/api/static/docs/*`, `frontend/lib/api/types/*` (via `task generate`)
- Test: `frontend/lib/customs-inventory.test.ts`

**Interfaces:**
- Consumes: `Preview.pieceCount`, `Preview.dutiablePieceCount` (Task 1)
- Produces: `export function formatQty(q: number, locale: string): string`

- [ ] **Step 1: Regenerate the API types**

Run: `task generate`
Expected: `CustomsInventoryPreview` in `frontend/lib/api/types/data-contracts.ts` has `pieceCount: number` and `dutiablePieceCount: number`.

- [ ] **Step 2: Write the failing test**

```ts
describe("formatQty", () => {
  it("formats pieces for the locale without float noise", () => {
    expect(formatQty(12, "en")).toBe("12");
    expect(formatQty(2.5, "en")).toBe("2.5");
    expect(formatQty(2.5, "de")).toBe("2,5");
    expect(formatQty(0.1 + 0.2, "en")).toBe("0.3");
  });
});
```

- [ ] **Step 3: Run it and see it fail**

Run: `cd frontend && pnpm exec vitest run --config ./test/vitest.config.ts lib/customs-inventory.test.ts`
Expected: FAIL, `formatQty` is not exported.

- [ ] **Step 4: Implement**

- `formatQty`: `new Intl.NumberFormat(locale, { maximumFractionDigits: 6 }).format(q)`.
- Dialog:
  - the positions KPI shows `{{ preview.lineCount }} / {{ formatQty(preview.pieceCount, locale) }}`;
  - `kpi_dutiable` gets `{ count: formatQty(preview.dutiablePieceCount, locale) }`.
- Locales: `kpi_positions` and `kpi_dutiable` take the round 2 §3 copy.

- [ ] **Step 5: Run the tests and checks**

Run:
- `cd frontend && pnpm exec vitest run --config ./test/vitest.config.ts lib/customs-inventory.test.ts`
- `task ui:fix`
- `task ui:check`

Expected: PASS, with no new lint warnings and no type errors.

- [ ] **Step 6: Commit and open PR A**

```bash
git add backend/app/api/static/docs docs/public/api frontend
git commit -m "feat(customs): show pieces instead of entries in the export dialog"
```

PR A goes against `main`. In the body, name R4 and the round 2 spec.

### Task 3: Merge by first word (R1, backend)

**Files:**
- Modify: `backend/internal/core/services/reporting/customsinventory/model.go`, `build.go`
- Test: `build_test.go`

**Interfaces:**
- Produces:
  - `Options.MergeFirstWord bool` (`json:"mergeFirstWord"`)
  - `func FirstWord(name string) string`
  - `entry.firstWord string`
  - `makeLines(entries []*entry, merge, firstWord bool) []Line`

- [ ] **Step 1: Write the failing tests**

```go
func TestFirstWord(t *testing.T) {
	for in, want := range map[string]string{
		"Books & deco": "Books", "  books   and deco": "books", "Bücher/Deko": "Bücher",
		"3D-Drucker": "3D", "& more": "more", "—": "—", "": "",
	} {
		assert.Equal(t, want, FirstWord(in), "%q", in)
	}
}

func bookBox(n int, name string, qty, price float64) Item {
	it := furniture(n, int64(n), name, qty, price, "2019-01-01", locLiving, pathLiving)
	it.EntityTypeID, it.EntityTypeName = typeBox, "Box"
	return it
}

func booksInput(first bool) Input {
	in := smallInput(
		bookBox(1, "Books & deco", 2, 50), bookBox(2, "books and deco", 3, 50),
		bookBox(3, "Books & material", 1, 50), bookBox(4, "Books & deco", 1, 80),
		bookBox(5, "Linen", 1, 50),
	)
	in.Options.MergeFirstWord = first
	return in
}

func TestBuild_MergeFirstWord(t *testing.T) {
	r := mustBuild(t, booksInput(true))
	lines := allLines(r)
	require.Len(t, lines, 3)
	books := lineByName(t, r, "Books")
	assert.InDelta(t, 6.0, books.Quantity, 0)
	assert.Equal(t, int64(5000), books.UnitCents)
	assert.Equal(t, int64(30000), books.LineCents)
	assert.Equal(t, []string{"000-001", "000-002", "000-003"}, books.AssetIDs)
	assert.Equal(t, int64(8000), lineByName(t, r, "Books & deco").UnitCents) // own price, one member
	assert.Equal(t, []string{"Books", "Books & deco", "Linen"}, []string{lines[0].Name, lines[1].Name, lines[2].Name})
}

func TestBuild_MergeFirstWordKeepsSharedFullName(t *testing.T) {
	in := smallInput(bookBox(1, "Books & deco", 1, 50), bookBox(6, "books & DECO", 1, 50))
	in.Options.MergeFirstWord = true
	lines := allLines(mustBuild(t, in))
	require.Len(t, lines, 1)
	assert.Equal(t, "Books & deco", lines[0].Name)
}

func TestBuild_MergeFirstWordNeedsMerge(t *testing.T) {
	in := booksInput(true)
	in.Options.MergeIdentical = false
	assert.Len(t, allLines(mustBuild(t, in)), 5)
}
```

- [ ] **Step 2: Run them and see them fail**

Run: `task go:test -- ./internal/core/services/reporting/customsinventory/... -run 'FirstWord'`
Expected: FAIL, with `undefined: FirstWord` and the missing field.

- [ ] **Step 3: Implement**

- `FirstWord`: run `normalizeName` first. Then scan runes: skip runes that are neither `unicode.IsLetter` nor `unicode.IsDigit`, and take the first run of letters and digits. With no such run, return the normalized name.
- `newEntry` sets `firstWord`.
- When `MergeIdentical && MergeFirstWord`, sort the group's entries by `compareFold(a.firstWord, b.firstWord)` first, then `compareEntries`. Otherwise the order is unchanged.
- `mergeKey.name` is `strings.ToLower(e.firstWord)` in first-word mode.
- `lineFrom` takes a flag for the first-word mode. In that mode, it uses the first member's `firstWord` as the name when the members' `strings.ToLower(name)` differ.
- `MergeFirstWord` without `MergeIdentical` changes nothing.

- [ ] **Step 4: Run the package tests and see them pass**

Run: `task go:test -- ./internal/core/services/reporting/customsinventory/...`
Expected: PASS. `TestBuild_Golden` passes without `-update`, because the option is off there.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/core/services/reporting/customsinventory
git commit -m "feat(customs): optionally merge lines by the first word of the name"
```

### Task 4: First-word option in the dialog (R1, frontend)

**Files:**
- Modify: `frontend/lib/customs-inventory.ts`, `frontend/components/Reports/CustomsInventoryDialog.vue`, `frontend/locales/en.json`, `frontend/locales/de.json`
- Generated: via `task generate`
- Test: `frontend/lib/customs-inventory.test.ts`

**Interfaces:**
- Consumes: `CustomsInventoryOptions.mergeFirstWord` (Task 3, after `task generate`)
- Produces: `CustomsInventoryPrefs.mergeFirstWord: boolean`, default `false`

- [ ] **Step 1: Regenerate the API types**

Run: `task generate`
Expected: `CustomsInventoryOptions.mergeFirstWord: boolean` exists.

- [ ] **Step 2: Write the failing tests**

```ts
it("defaults mergeFirstWord to false and keeps a saved value", () => {
  expect(sanitizePrefs(undefined, new Set()).mergeFirstWord).toBe(false);
  expect(sanitizePrefs({ mergeFirstWord: true }, new Set()).mergeFirstWord).toBe(true);
  expect(sanitizePrefs({ mergeFirstWord: "yes" as unknown as boolean }, new Set()).mergeFirstWord).toBe(false);
});

it("sends mergeFirstWord", () => {
  const prefs = { ...sanitizePrefs(undefined, new Set()), mergeFirstWord: true };
  expect(buildRequest({ scope: "selected", itemIds: ["a"], search: null, prefs }).options.mergeFirstWord).toBe(true);
});
```

- [ ] **Step 3: Run them and see them fail**

Run: `cd frontend && pnpm exec vitest run --config ./test/vitest.config.ts lib/customs-inventory.test.ts`
Expected: FAIL.

- [ ] **Step 4: Implement**

- Prefs, defaults, `sanitizePrefs` and `buildRequest` get the new flag.
- Dialog: below the merge checkbox, add an indented `Checkbox` with `id="customs-merge-first-word"`, bound to `prefs.mergeFirstWord` and `:disabled="!prefs.mergeIdentical"`. Its label and hint use the keys `merge_first_word` and `merge_first_word_hint`.

- [ ] **Step 5: Run the tests and checks**

Run: vitest as in Step 3, then `task ui:fix`, then `task ui:check`.
Expected: PASS, clean.

- [ ] **Step 6: Commit and open PR B**

```bash
git add backend/app/api/static/docs docs/public/api frontend
git commit -m "feat(customs): add the first-word merge option to the export dialog"
```

### Task 5: Containers absorb their contents (R2, pure package)

**Files:**
- Create: `backend/internal/core/services/reporting/customsinventory/contents.go`
- Modify: `model.go`, `build.go`, `verify.go`
- Test: create `contents_test.go`; modify `verify_test.go`

**Interfaces:**
- Consumes: `makeLines(entries, merge, firstWord)` (Task 3); `FormatQtyDE` (Task 1)
- Produces:
  - `Item.ParentID uuid.UUID`
  - `Options.RollUpContents bool` (`json:"rollUpContents"`)
  - `IssueContentSplit IssueCode = "content_split"` (warning)
  - `IssueContainerQuantity IssueCode = "container_quantity"` (info, in `SeverityOf`). Both go into `IssueOrder` right after `IssueReceiptIsLink`.
  - `type Content struct { ID, OwnerID uuid.UUID; Quantity float64; UnitCents int64 }`
  - `Line.MemberOwnCents []int64` (parallel to `MemberIDs`), `Line.Contents []Content`
  - `Preview.ContentCount int` (`json:"contentCount"`)
  - `entry.parent *entry` (the included parent, or nil); `entry.contents []*entry`; `entry.lineUnitCents int64`
  - Verify invariant letter `(i)`

- [ ] **Step 1: Write the failing tests**

Fixture in `contents_test.go`:

```go
var (
	boxID, bookID, vaseID, pouchID, penID = id(1001), id(1002), id(1003), id(1004), id(1005)
)

// containerInput: a duty-free box (own price 0) that holds books, a vase and a pouch with pens.
// Only the box is in scope. Contents value: 3×10.00 + 25.50 + 5.00 + 2×1.25 = 63.00, 7 pieces.
func containerInput() Input {
	mk := func(n int, name string, qty, price float64, parent uuid.UUID) Item {
		it := furniture(n, int64(n-1000), name, qty, price, "2019-01-01", locLiving, pathLiving)
		it.ParentID = parent
		return it
	}
	box := mk(1001, "Box books", 1, 0, locLiving)
	in := smallInput(box)
	in.Items = append(in.Items,
		mk(1002, "Book", 3, 10, boxID), mk(1003, "Vase", 1, 25.50, boxID),
		mk(1004, "Pouch", 1, 5, boxID), mk(1005, "Pen", 2, 1.25, pouchID))
	in.Options.RollUpContents = true
	return in
}
```

Tests, each built from `containerInput()` with one change:

| test | change | asserts |
|---|---|---|
| `TestBuild_RollUp_ListsOnlyTopContainer` | — | one line "Box books"; `UnitCents` 6300; `LineCents` 6300; `ItemCount` 5; `Preview().ContentCount` 4; `Comments` contains `inkl. Inhalt (7 Stk.)`; `ScopeIDs` has all 5 IDs |
| `TestBuild_RollUp_ParentPriceAdded` | box price 12.00 | `UnitCents` 7500 |
| `TestBuild_RollUp_ChildrenInScopeNotListedTwice` | `ScopeIDs` = box, book, pen | one line; `Verify` passes |
| `TestBuild_RollUp_Off` | `RollUpContents=false`, `ScopeIDs` = box, book | two lines; `ScopeIDs` has 2 IDs; no `inkl. Inhalt` comment |
| `TestBuild_RollUp_DutiableChildSplitOut` | vase bought `2026-10-01` | box `UnitCents` 3750; one dutiable line "Vase" with `Inhalt von 000-001` in its comments; `issueNames(r, IssueContentSplit)` = `["Vase"]` |
| `TestBuild_RollUp_ContainerQuantityNotOne` | box quantity 2 | four lines: box, book, vase, and pouch (the pen is still absorbed into the pouch, `UnitCents` 750); `issueNames(..., IssueContainerQuantity)` = `["Box books"]`; `issueNames(..., IssueContentSplit)` is empty |
| `TestBuild_RollUp_QuantityParentStopsWalk` | pouch quantity 2 | three lines: box with `UnitCents` 5550, pouch, and pen with `Inhalt von 000-004` (the pen doesn't skip the pouch into the box); `issueNames(..., IssueContainerQuantity)` = `["Pouch"]`; `issueNames(..., IssueContentSplit)` is empty |
| `TestBuild_RollUp_DifferentStatusParentStopsWalk` | add "Bag" (asset 6, 1 × 40.00, bought `2026-10-01`, parent box) and "Scarf" (asset 7, 1 × 15.00, bought `2019-01-01`, parent bag) | box `UnitCents` 6300; a dutiable line "Bag" with `Inhalt von 000-001`; a duty-free line "Scarf" with `UnitCents` 1500 and `Inhalt von 000-006` (not rolled into the box); `issueNames(..., IssueContentSplit)` = `["Bag", "Scarf"]`; `Verify` passes |
| `TestBuild_RollUp_MissingPrice` | book price 0 | `issueNames(..., IssueMissingPrice)` = `["Book"]` (the box isn't listed); box comments contain `Wert fehlt` |
| `TestBuild_RollUp_ReceiptsOfContents` | `IncludeReceipts=true`; the vase has `receipt(1, "r.pdf", "application/pdf", "2024-01-01T00:00:00Z")` | the box line's `ReceiptNames` = `["000-003_Rechnung.pdf"]`; `r.Receipts` has that name with `ItemID` = vase |
| `TestBuild_RollUp_SoldContentExcluded` | book `SoldDate` set | `Excluded` has the book with `IssueExcludedSold`; `UnitCents` 3300 |
| `TestBuild_RollUp_ContentUnderArchivedIgnored` | pouch `Archived` | the pouch is in `Excluded` (`IssueExcludedArchived`); the pen ID is not in `r.ScopeIDs`; `UnitCents` 5550 |
| `TestBuild_RollUp_ParentCycleTerminates` | the box's `ParentID` = pen (cycle box → pen → pouch → box) | `Build` returns with no error; one line "Box books" with `UnitCents` 6300 (the cycle's link is cut at the box, the member with the smallest ID) |
| `TestBuild_RollUp_ScopeChildSplitOnce` | vase bought `2026-10-01` and also in `ScopeIDs` | exactly one line contains the vase's ID among its members |
| `TestBuild_RollUp_MergedContainersKeepTheirContents` | a second box "Box books" (asset 9, scope) holding "Plate" at 63.00 | one merged line, quantity 2, `UnitCents` 6300, `LineCents` 12600; `Contents` owners are both box IDs; `Verify` passes |

In `verify_test.go`, add `"invariant (i)"` cases. They use a report built from `containerInput()`, not the golden one:
- `l.UnitCents++` together with `l.LineCents = LineCents(l.Quantity, l.UnitCents)`;
- `l.Contents[0].OwnerID = uuid.New()`.

Add `"invariant (a)"` for a content ID that is duplicated into `l.Contents`.

- [ ] **Step 2: Run them and see them fail**

Run: `task go:test -- ./internal/core/services/reporting/customsinventory/... -run 'RollUp|Verify'`
Expected: FAIL, the fields and codes don't exist.

- [ ] **Step 3: Implement in `contents.go` and `build.go`**

Build order inside `Build`:
1. Resolve the universe.
2. Make the entries.
3. Link parents.
4. Classify.
5. Assign owners.
6. Compute values.
7. Finish comments and `hasPrice`.
8. Raise issues.
9. Group and merge.
10. Number the positions.

How each step works:

- `universe(in Input, scope []uuid.UUID) []uuid.UUID` returns the scope IDs, deduplicated and in order. When `RollUpContents` is on, it appends, in `in.Items` order, every item that isn't in scope and whose `ParentID` is an included item that is in scope or is itself content. Use a memo plus a visited set, so a cycle ends. `exclusion` decides "included". `Report.ScopeIDs` is the universe.
- `linkParents(entries)`: when the option is on, `e.parent` = the entry of `e.it.ParentID` when there is one, else nil. With the option off, it stays nil everywhere.
  - It then breaks cycles: for each entry, walk up `parent`. If the walk comes back to an entry already on the path, set `parent = nil` on the cycle member with the smallest ID (string compare). After that, every chain ends, and later walks need no guard for correctness. Keep a visited guard anyway.
- Classification moves out of `newEntry`, to after `linkParents`. In this task it still uses the item's own tags.
- `assignOwners(entries)`: for each entry, step up through `parent`, with a visited guard, while the parent `a` has `a.class.Status == e.class.Status && a.it.Quantity == 1`. Stop at the first parent that fails; never skip it (FR-30). If at least one step was taken, append `e` to the `contents` of the last `a` reached. Sort each `contents` with `compareEntries`.
  - The owner is never absorbed itself: its parent fails the same test for it, because the owner has `e`'s status. No second pass is needed.
- `lineUnitCents` = `unitCents` + Σ `LineCents(c.it.Quantity, c.unitCents)` over `contents`.
  - For a member, `hasPrice` = `lineUnitCents != 0`.
  - For absorbed content, `hasPrice` = `unitCents != 0`.
- The comment is built after the values. Join these parts, leaving out empty ones, with `"; "`:
  - `ReasonComment`;
  - `Wert fehlt` when `!hasPrice`;
  - `Inhalt von <parent.assetLabel>` when `e` is a member and `e.parent != nil`.
- Issues, in `itemIssues`:
  - `missing_price`: when `unitCents == 0 && !(len(contents) > 0 && lineUnitCents != 0)`.
  - `content_split`: an entry with `parent != nil && parent.class.Status != e.class.Status` (always a member, since its walk stops at once). The detail is the parent's label.
  - `container_quantity`: an entry with `Quantity != 1` that is the `parent` of at least one entry. The detail is `FormatQtyDE(Quantity)`.
- Only members (entries with no owner) go into `groupEntries`. The merge key uses `lineUnitCents`.
- `lineFrom`:
  - `UnitCents = first.lineUnitCents`;
  - for each member, append `MemberOwnCents` (the member's `unitCents`) and the member's `Contents`;
  - receipt names: the member's, then its contents'.
  - comments, each once: the member's, then its contents';
  - when the line has contents, add `inkl. Inhalt (<FormatQtyDE(Σ content qty)> Stk.)` last.
- `collectReceipts`: for each member, its receipts, then those of its contents in `Contents` order.
- `Preview.ContentCount` = Σ `len(l.Contents)`.
- `Verify`:
  - (a) also counts `l.Contents[].ID`.
  - (i) for each line: `len(MemberOwnCents) == len(MemberIDs)`; every `Content.OwnerID` is in `MemberIDs`; for each member `k`, `UnitCents == MemberOwnCents[k] + Σ LineCents(c.Quantity, c.UnitCents)` over its contents; a member with contents has `MemberQuantities[k] == 1`.
  - Messages carry the position only.

- [ ] **Step 4: Run the package tests and see them pass**

Run: `task go:test -- ./internal/core/services/reporting/customsinventory/...`
Expected: `TestBuild_Golden` fails only because the new `Line` fields show up. Rerun it with `-update`. The diff in `testdata/golden_report.json` may contain only `MemberOwnCents` (equal to `UnitCents` on every line) and `Contents: null`. Then the whole package passes.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/core/services/reporting/customsinventory
git commit -m "feat(customs): roll contents into their container line"
```

### Task 6: Contents inherit their containers' customs tags (R2, FR-35)

**Files:**
- Modify: `backend/internal/core/services/reporting/customsinventory/contents.go`, `build.go`
- Test: `contents_test.go`

**Interfaces:**
- Consumes: `entry.parent` (Task 5)
- Produces: `func effectiveTags(e *entry) []uuid.UUID`, the item's own tags followed by every ancestor's tags along `parent`, deduplicated

- [ ] **Step 1: Write the failing tests**

All use `containerInput()`. The tags come from the golden fixture's tag list: `tagP` (parent of `tagP2`, "owned > 2 years") and `tagZ` (always dutiable). Set `in.Tags` to the golden tags.

| test | change | asserts |
|---|---|---|
| `TestBuild_RollUp_ContentInheritsDutyFreeTag` | `DutyFreeTagIDs={tagP}`; the box is tagged `tagP2`; the book has no purchase date | one line; no `missing_purchase_date` issue; the line's comments contain `zollfrei lt. Tag „owned > 2 years“` |
| `TestBuild_RollUp_ContentInheritsDutiableTag` | `DutiableTagIDs={tagZ}`; the box is tagged `tagZ` | no duty-free group; one dutiable line "Box books" with `UnitCents` 6300 |
| `TestBuild_RollUp_DateBeatsInheritedTag` | as the first test, and the vase is bought `2026-10-01` | the vase is on its own dutiable line; `issueNames(..., IssueConflictDateTag)` = `["Vase"]` |
| `TestBuild_NoInheritanceWhenRollUpOff` | as the first test, with `RollUpContents=false` and `ScopeIDs` = box, book | the book is dutiable with `missing_purchase_date` |

- [ ] **Step 2: Run them and see them fail**

Run: `task go:test -- ./internal/core/services/reporting/customsinventory/... -run 'Inherit'`
Expected: FAIL, the book is dutiable.

- [ ] **Step 3: Implement**

`Classify` takes the item with `TagIDs` replaced by `effectiveTags(e)`. Copy the `Item` and don't mutate `e.it`, because the tag grouping (FR-4) keeps using the item's own tags. Walk `parent` with a visited guard.

- [ ] **Step 4: Run the package tests and see them pass**

Run: `task go:test -- ./internal/core/services/reporting/customsinventory/...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/core/services/reporting/customsinventory
git commit -m "feat(customs): classify contents with their containers' customs tags"
```

### Task 7: The service loads contents (R2, FR-28)

**Files:**
- Modify: `backend/internal/core/services/service_customs_inventory.go`
- Test: `backend/internal/core/services/service_customs_inventory_test.go`

**Interfaces:**
- Consumes: `Item.ParentID`, `Options.RollUpContents` (Task 5); `repo.EntityQuery.ParentItemIDs`
- Produces: `func (s *CustomsInventoryService) loadContents(ctx context.Context, gid uuid.UUID, frontier []uuid.UUID, in *customsinventory.Input, paths map[uuid.UUID]string, locationPaths map[uuid.UUID][]string, budget int) error`

- [ ] **Step 1: Write the failing tests**

Fixture changes:
- Add `itemType` and `living` (`uuid.UUID`) to `customsFixture`, set in `newCustomsFixture`.
- Add `newItem` as a method, so new tests can create items under any parent.
- New helper `withBox(t, f) (box, book, pouch, pen uuid.UUID)`. It creates:
  - "Moving box" in `living`: asset 5, quantity 1, price 0, bought 2019-01-01, tag Shipment A;
  - "Book" in the box: asset 6, quantity 3, price 10.00, bought 2019-01-01, no tags;
  - "Pouch" in the box: asset 7, quantity 1, price 2.00, bought 2019-01-01;
  - "Pen" in the pouch: asset 8, quantity 2, price 1.25, bought 2019-01-01.

Existing tests keep using `newCustomsFixture` without the box.

| test | request | asserts |
|---|---|---|
| `TestCustomsInventory_RollUpContents` | `request(box)` with `RollUpContents=true` | `LineCount` 1; `ContentCount` 3; `ItemCount` 4; `GrandTotalCents` 3450 |
| `TestCustomsInventory_RollUpByQuery` | query `TagIDs={shipment}` with `RollUpContents=true` | `ContentCount` 3; `LineCount` 4 (clock, chairs, espresso, box) |
| `TestCustomsInventory_RollUpCountsTowardLimit` | `request(box)` with `RollUpContents=true` on a service with `maxScope: 3` | `errors.Is(err, customsinventory.ErrScopeTooLarge)` |
| `TestCustomsInventory_RollUpOffByDefault` | `request(box)` | `ContentCount` 0; `GrandTotalCents` 0; a `missing_price` issue lists "Moving box" |

- [ ] **Step 2: Run them and see them fail**

Run: `task go:test -- ./internal/core/services/... -run 'CustomsInventory_RollUp'`
Expected: FAIL, `ContentCount` is 0.

- [ ] **Step 3: Implement**

- `toCustomsItem` sets `ParentID` from `e.Parent.ID` when `e.Parent != nil`.
- Factor the per-entity loading in `buildInput` (`GetOneByGroup`, location path cache, attachment paths) into a helper that both loops share.
- When `req.Options.RollUpContents` is on, call `loadContents` after the scope loop.
  - The frontier is the loaded scope entities that aren't locations, aren't archived and have a zero `SoldDate`.
  - For each level, run `QueryByGroup(ctx, gid, repo.EntityQuery{Page: -1, PageSize: -1, ParentItemIDs: frontier})`.
  - Skip children already loaded, and children whose `EntityType` is a location.
  - Load each child with the shared helper.
  - The next frontier is the loaded children that aren't archived and aren't sold.
  - Stop when the frontier is empty.
  - Return `ErrScopeTooLarge`, with the same message format, as soon as scope + contents exceed `s.limit()`.
- Add the span attribute `contents.count`.

- [ ] **Step 4: Run the service and package tests and see them pass**

Run:
- `task go:test -- ./internal/core/services/...`
- `task go:test -- ./app/api/handlers/v1/... -run CustomsInventory`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/core/services
git commit -m "feat(customs): load container contents for the export"
```

### Task 8: Rollup option, content count and new issues in the dialog (R2, frontend)

**Files:**
- Modify: `frontend/lib/customs-inventory.ts`, `frontend/components/Reports/CustomsInventoryDialog.vue`, `frontend/locales/en.json`, `frontend/locales/de.json`
- Generated: via `task generate`
- Test: `frontend/lib/customs-inventory.test.ts`

**Interfaces:**
- Consumes: `CustomsInventoryOptions.rollUpContents`, `CustomsInventoryPreview.contentCount` (Tasks 5–7, after `task generate`)
- Produces: `CustomsInventoryPrefs.rollUpContents: boolean`, default `true`

- [ ] **Step 1: Regenerate the API types**

Run: `task generate`
Expected: both fields exist in `data-contracts.ts`.

- [ ] **Step 2: Write the failing tests**

```ts
it("defaults rollUpContents to true and keeps a saved false", () => {
  expect(sanitizePrefs(undefined, new Set()).rollUpContents).toBe(true);
  expect(sanitizePrefs({ rollUpContents: false }, new Set()).rollUpContents).toBe(false);
});

it("sends rollUpContents", () => {
  const prefs = sanitizePrefs(undefined, new Set());
  expect(buildRequest({ scope: "selected", itemIds: ["a"], search: null, prefs }).options.rollUpContents).toBe(true);
});
```

- [ ] **Step 3: Run them and see them fail**

Run: `cd frontend && pnpm exec vitest run --config ./test/vitest.config.ts lib/customs-inventory.test.ts`
Expected: FAIL.

- [ ] **Step 4: Implement**

- Prefs, defaults, `sanitizePrefs` and `buildRequest` get the flag.
- Dialog:
  - add a `Checkbox` with `id="customs-roll-up"` in the grouping fieldset, after the merge options, with the keys `roll_up` and `roll_up_hint`;
  - under the positions KPI, add a line `kpi_contents` with `{ count: preview.contentCount }`, shown when `preview.contentCount > 0`.
- Locales: add `issue_content_split` and `issue_container_quantity`.

- [ ] **Step 5: Run the tests and checks**

Run: vitest as in Step 3, then `task ui:fix`, then `task ui:check`.
Expected: PASS, clean.

- [ ] **Step 6: Commit and open PR C**

```bash
git add backend/app/api/static/docs docs/public/api frontend
git commit -m "feat(customs): add the contents rollup option to the export dialog"
```

### Task 9: One entry point on the search page (R3)

**Files:**
- Modify:
  - `frontend/components/Item/View/table/data-table-dropdown.vue`
  - `frontend/components/Item/View/table/data-table.vue`
  - `frontend/components/Item/View/Selectable.vue`
  - `frontend/pages/search.vue`
  - `frontend/locales/en.json`, `frontend/locales/de.json`
- Test: `frontend/test/e2e/customs-inventory.browser.spec.ts`

**Interfaces:**
- Produces:
  - `data-table.vue` emits `selectionChange: [ids: string[]]`;
  - `Selectable.vue` re-emits `selection-change` and has a named slot `actions`, rendered in the section header next to the view toggle.

- [ ] **Step 1: Rewrite the e2e spec so it fails**

- `openExportDialog(page)` clicks `page.getByRole("button", { name: "Export inventory list …" })` and returns the dialog. It no longer selects rows or opens a menu.
- The main test:
  1. log in;
  2. `goto("/search")`;
  3. click "Table";
  4. wait for `networkidle`;
  5. select every row with the header checkbox, using the existing retry;
  6. open the actions menu (`header.getByRole("button", { name: "Open menu" })`) and assert `page.getByRole("menuitem", { name: /Customs Inventory/ })` has count 0; press Escape;
  7. open the dialog with the button and assert that the radio "Only selected items (N)" is checked, where N is the number in "N of N rows selected";
  8. run the existing export steps.
- New step, "a new search clears the selection": after the export, type `zzz-no-match` into the search field and submit. Open the dialog with the button, and assert that the "Only selected items (0)" radio is disabled.

Run: `task test:e2e -- customs-inventory`
Expected: FAIL. There is no "Export inventory list …" button. (This task uses its own server and database, never the local dev database.)

- [ ] **Step 2: Implement**

- `data-table-dropdown.vue`: remove the customs menu item, `useCustomsInventorySearch()` and its constant.
- `data-table.vue`:
  - after `rowSelection` changes, emit `selectionChange` with `table.getSelectedRowModel().rows.map(r => r.original.id)`;
  - reset `rowSelection` to `{}` when `props.data` changes (Review Focus 5).
- `Selectable.vue`: forward the event, and render `<slot name="actions" />` before the view-toggle `ButtonGroup`, outside its `v-if`.
- `search.vue`:
  - add `const selectedIds = ref<string[]>([])`;
  - bind `@selection-change` on `ItemViewSelectable`;
  - in the `#actions` slot, add a `Button size="sm" variant="outline"` that is `:disabled="!customsInventorySearch"` and calls `openDialog(DialogID.CustomsInventoryExport, { params: { itemIds: selectedIds, search: customsInventorySearch } })`, with the label `search_button`.
- Locales: remove `menu_entry` and add `search_button`, in en and de.

- [ ] **Step 3: Run the checks**

Run:
- `task ui:fix`
- `task ui:check`
- `task test:e2e -- customs-inventory`

Expected: lint and typecheck are clean, and the e2e spec passes in chromium, firefox and webkit.

- [ ] **Step 4: Commit and open PR D**

```bash
git add frontend
git commit -m "feat(customs): open the export from a search page button only"
```
