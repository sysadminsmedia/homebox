# Customs Inventory Export Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** From the item search, export selected (or all matching) items as a customs inventory list in `.xlsx`, for export and duty declarations, optionally zipped with the invoices, with a live pre-export check. Each line's value is quantity × purchase price per piece.

**Architecture:** A pure Go package (`customsinventory`) turns plain item data into a verified `Report`, then writes the `.xlsx` and the ZIP. A thin service loads group-scoped data from the existing repos and feeds the package. Two POST endpoints (preview, export) share the same `Build` call. A Nuxt dialog, opened from the item table's actions menu, calls preview on every option change and export on click.

**Tech Stack:** Go 1.27, chi/errchain handlers with swaggo annotations, ent repos, `github.com/xuri/excelize/v2 v2.11.0`, `archive/zip`, gocloud `blob`; Nuxt 4 / Vue 3 / ShadCN, vitest, Playwright.

**Spec:** [`docs/superpowers/specs/2026-10-31-customs-inventory-export-design.md`](../specs/2026-10-31-customs-inventory-export-design.md) (FR-/CALC-/NFR- numbers below refer to it). **Mockups:** [`…-mockups.html`](../specs/2026-10-31-customs-inventory-export-mockups.html).

## Global Constraints

- New dependency, exactly: `github.com/xuri/excelize/v2 v2.11.0`. No other new Go or npm dependencies.
- `backend/internal/core/services/reporting/customsinventory` imports nothing from `ent`, `repo`, `services`, `net/http` or `blob`. Allowed: stdlib, `github.com/google/uuid`, `internal/data/types`, excelize.
- Money is `int64` cents everywhere inside the package. A float appears only as input (`Item.PurchasePrice`) and as the value written to an Excel cell (`cents/100`).
- `Item.PurchasePrice` is the price **per piece**. A line's value is always `LineCents(quantity, unitCents)`; never use the price alone as a line value.
- Dates cross the API as `"YYYY-MM-DD"` strings. Internally they are `types.Date`; never compare times of day.
- Every repo call passes `services.Context.GID`. IDs, tags and attachments outside the group must never produce data (NFR-5).
- User text goes into cells only via `SetCellStr` (FR-16).
- Sheet labels and UI copy are exactly as in spec §8 and the T9 copy table.
- Don't hand-edit `frontend/lib/api/types/*` or `backend/app/api/static/docs/*`. Regenerate them with `task swag && task typescript-types`.
- Frontend work is done only when `task ui:fix`, `task ui:check` and `cd frontend && pnpm run lint:ci` (≤ 1 warning) all pass.
- Never set `UNSAFE_DISABLE_PASSWORD_PROJECTION` for tests.
- No test, golden or fixture may use the local dev database. It is incomplete (entity types are missing). Fixtures only.
- Docs, comments, commit messages and PR text stay neutral: no personal background, and the person using the feature is "the user".
- **Git:**
  - Never commit to `main`.
  - Commit messages end with `Co-Authored-By: Claude <model name> <noreply@anthropic.com>`.
  - PR bodies end with `Assisted by **Agent Blue** (Claude Code, model: <model name> — \`<model id>\`)`.
  - One PR per lane (see Execution Lanes).

## Review Focus

1. **Price per piece.** An item with quantity 4 and price 89.50. Expected: unit value 89.50, line value 358.00, and the totals include 358.00, not 89.50. HomeBox's own value-by-location and value-by-tag charts skip the quantity, so don't copy their queries. (Tests in Tasks 3 and 4.)
2. **One item, several group tags.** In tag grouping, an item carries a group-B tag and a child of group-A's tag, with A ordered first. Expected: it appears once, in group A. (Test in Task 3.)
3. **Shared asset ID.** Two exported items share asset ID 123 and both have receipts. Expected: ZIP names `000-123_Rechnung.pdf` and `000-123_b_Rechnung.pdf`, each referenced on its own line. (Tests in Task 3 for names, Task 7 for the ZIP.)
4. **Asset-ID search syntax.** The user searched with `#000-123` and chose "all matching". Expected: the same items the search page shows. (Test in Task 7, `toEntityQuery`.)
5. **Formula-like text.** Names or locations like `=HYPERLINK("http://x")`, `+1 555 0100`, `-`, or containing umlauts. Expected: stored as plain text, shown verbatim, never a formula. (Test in Task 4.)

---

## Execution Lanes (multi-agent)

```
T1 ──┬─> T2 ──┐
     ├─> T5 ──┼─> T3 ──┐
     ├─> T4 ──┴────────┼─> T7 ──┐
     └─> T6 ──┬────────┘        ├─> T10
              └─> T8 ─> T9 ─────┘
```

| wave | tasks (parallel) | needs |
|---|---|---|
| 0 | T1 | – |
| 1 | T2, T4, T5, T6 | T1 |
| 2 | T3, T8 | T2+T4+T5 / T6 |
| 3 | T7, T9 | T3+T4+T5+T6 / T8 |
| 4 | T10 | T7+T9 |

**File ownership.** No two tasks in the same wave touch the same file.
- `go.mod`/`go.sum`: T4 only.
- `routes.go`, `typegen/main.go`, generated docs and types: T6 only.
- `services/all.go`: T7 only.
- Locale files and `layouts/default.vue`: T9 only.
- `use-preferences.ts` and `search.vue`: T8 only.

**Branches and PRs.** Parallel tasks run in their own worktrees, cut from the lane branch tip. The coordinator merges them back into the lane branch; conflicts aren't expected.

| PR | branch | tasks | base |
|---|---|---|---|
| A | `feat/customs-inventory-core` | T1–T5 | `main` |
| B | `feat/customs-inventory-api` | T6, T7 | PR-A |
| C | `feat/customs-inventory-ui` | T8, T9 | T6 commit on PR-B |
| D | `test/customs-inventory-e2e` | T10 | B + C |

**Commands** (backend paths are relative to `backend/`, as `task go:test` expects):
- Package tests: `task go:test -- ./internal/core/services/reporting/customsinventory/... -run <Name>`
- Service tests: `task go:test -- ./internal/core/services/ -run <Name>`
- Frontend unit tests: `cd frontend && pnpm exec vitest run --config ./test/vitest.config.ts <file>`

---

### Task 1: Package foundations (contract types, money, dates)

**Files:**
- Create: `backend/internal/core/services/reporting/customsinventory/model.go`
- Create: `backend/internal/core/services/reporting/customsinventory/money.go`
- Create: `backend/internal/core/services/reporting/customsinventory/dates.go`
- Test: `backend/internal/core/services/reporting/customsinventory/money_test.go`, `dates_test.go`, `model_test.go`

**Interfaces:**
- Consumes: `types.Date` (`internal/data/types`).
- Produces (all in `model.go` unless noted):

```go
const MaxScope = 5000
const DefaultMinMonths = 6 // used when Options.MinMonths == 0
const MaxMinMonths = 60
var DefaultReceiptTypes = []string{"receipt"}
var AllowedReceiptTypes = []string{"receipt", "attachment", "warranty", "manual"}
var ErrInvalidRequest = errors.New("customs inventory: invalid request")
var ErrScopeTooLarge  = errors.New("customs inventory: scope too large")

type GroupBy string   // "location" | "entityType" | "tag"
const (GroupByLocation GroupBy = "location"; GroupByEntityType GroupBy = "entityType"; GroupByTag GroupBy = "tag")
type Status string    // "dutyFree" | "dutiable"
type Reason string    // "date" | "tag_duty_free" | "owned_too_short" | "purchased_after_reference" | "missing_purchase_date" | "tag_dutiable"
type IssueCode string // the FR-11 codes, verbatim, as constants IssueMissingPurchaseDate … IssueExcludedNotFound
type Severity string  // "warning" | "info"

// --- API contract (JSON) ---
type Request struct {
    Scope   Scope   `json:"scope"`
    Options Options `json:"options"`
    Header  Header  `json:"header"`
}
type Scope struct {
    ItemIDs []uuid.UUID `json:"itemIds"`
    Query   *ScopeQuery `json:"query,omitempty" extensions:"x-nullable"`
}
type ScopeQuery struct {
    Q                string      `json:"q"`
    ParentIDs        []uuid.UUID `json:"parentIds"`
    EntityTypeIDs    []uuid.UUID `json:"entityTypeIds"`
    TagIDs           []uuid.UUID `json:"tags"`
    NegateTags       bool        `json:"negateTags"`
    MatchAllTags     bool        `json:"matchAllTags"`
    OnlyWithPhoto    bool        `json:"onlyWithPhoto"`
    OnlyWithoutPhoto bool        `json:"onlyWithoutPhoto"`
    IncludeArchived  bool        `json:"includeArchived"`
    Fields           []string    `json:"fields"` // "name=value"
}
type Options struct {
    GroupBy         GroupBy     `json:"groupBy"`
    GroupTagIDs     []uuid.UUID `json:"groupTagIds"`
    DutyFreeTagIDs  []uuid.UUID `json:"dutyFreeTagIds"`
    DutiableTagIDs  []uuid.UUID `json:"dutiableTagIds"`
    ReferenceDate   string      `json:"referenceDate"` // YYYY-MM-DD
    MinMonths       int         `json:"minMonths"`     // 0 → DefaultMinMonths; allowed 1..MaxMinMonths
    MergeIdentical  bool        `json:"mergeIdentical"`
    IncludeReceipts bool        `json:"includeReceipts"`
    ReceiptTypes    []string    `json:"receiptTypes"`
}
type Header struct {
    Title   string `json:"title"   validate:"max=200"`
    Name    string `json:"name"    validate:"max=200"`
    Address string `json:"address" validate:"max=1000"` // newline-separated
    Phone   string `json:"phone"   validate:"max=100"`
    Email   string `json:"email"   validate:"max=200"`
}
func (r Request) Validate() error // wraps ErrInvalidRequest; called automatically by adapters.DecodeBody

type Preview struct {
    Currency           string  `json:"currency"`
    ReferenceDate      string  `json:"referenceDate"`
    MinMonths          int     `json:"minMonths"` // after applying the default
    Cutoff             string  `json:"cutoff"`
    ItemCount          int     `json:"itemCount"`
    LineCount          int     `json:"lineCount"`
    DutiableItemCount  int     `json:"dutiableItemCount"`
    ExcludedCount      int     `json:"excludedCount"`
    ReceiptCount       int     `json:"receiptCount"`
    DutyFreeTotalCents int64   `json:"dutyFreeTotalCents"`
    DutiableTotalCents int64   `json:"dutiableTotalCents"`
    GrandTotalCents    int64   `json:"grandTotalCents"`
    Issues             []Issue `json:"issues"`
}
type Issue struct {
    Code     IssueCode   `json:"code"`
    Severity Severity    `json:"severity"`
    Items    []IssueItem `json:"items"`
}
type IssueItem struct {
    ID      uuid.UUID `json:"id"`
    AssetID string    `json:"assetId"`
    Name    string    `json:"name"`
    Detail  string    `json:"detail"`
}

// --- domain input (built by the service) ---
type Tag struct{ ID, ParentID uuid.UUID; Name string }
type Attachment struct{ ID uuid.UUID; Type, Title, MimeType string; CreatedAt time.Time; IsLink bool }
type Item struct {
    ID uuid.UUID; AssetID int64; Name string
    EntityTypeID uuid.UUID; EntityTypeName string; IsLocation bool
    Archived bool; SoldDate types.Date
    Quantity, PurchasePrice float64; PurchaseDate types.Date // PurchasePrice is per piece
    LocationID uuid.UUID; LocationPath []string // root first; empty = no location
    TagIDs []uuid.UUID; Attachments []Attachment
}
type Input struct {
    Options Options; Header Header; Currency string; GeneratedAt time.Time
    Tags []Tag; Items []Item; ScopeIDs []uuid.UUID // every requested/resolved ID, found or not
}

// --- output ---
type Report struct {
    Header Header; Currency string; ReferenceDate, Cutoff types.Date; MinMonths int; GeneratedAt time.Time
    ItemCount int; Groups []Group; Dutiable []Line
    DutyFreeTotalCents, DutiableTotalCents, GrandTotalCents int64
    Excluded []Excluded; Issues []Issue; Receipts []ReceiptFile; ScopeIDs []uuid.UUID
}
type Group struct{ Title string; Lines []Line; SubtotalCents int64 }
type Line struct {
    Pos int; Quantity float64; Name, EntityTypeName string; Locations, AssetIDs []string
    PurchaseDate types.Date; PurchaseDateText string // text only for merged ranges
    UnitCents int64; HasPrice bool; LineCents int64 // UnitCents = price per piece; LineCents = LineCents(Quantity, UnitCents)
    ReceiptNames, Comments []string; Status Status; Reason Reason
    MemberIDs []uuid.UUID; MemberQuantities []float64
}
type Excluded struct{ ID uuid.UUID; AssetID, Name string; Code IssueCode }
type ReceiptFile struct{ Name string; AttachmentID, ItemID uuid.UUID }
```

- `money.go`:
  - `ToCents(p float64) int64`
  - `LineCents(qty float64, unitCents int64) int64`
  - `RoundQty(q float64) float64` (6 decimals)
  - `FormatCents(c int64) string` (`"1234.50"`, `"-1.01"`)
- `dates.go`:
  - `ParseDate(s string) (types.Date, error)`, accepting only `YYYY-MM-DD`
  - `SubtractMonthsClamped(d types.Date, months int) types.Date`
  - `Cutoff(ref types.Date, months int) types.Date` (`months` 0 → `DefaultMinMonths`)
  - `FormatDE(d types.Date) string` (`"15.11.2026"`)

- [ ] **Step 1: Write the failing tests**

```go
func TestToCents(t *testing.T) {
    cases := map[float64]int64{0: 0, 19.99: 1999, 1.005: 101, 2.675: 268, 0.1 + 0.2: 30, -1.005: -101,
        1234567.891: 123456789, 0.004: 0, 0.005: 1}
    for in, want := range cases { assert.Equal(t, want, ToCents(in), "ToCents(%v)", in) }
}
func TestLineCents(t *testing.T) {
    type c struct{ q float64; u, want int64 }
    for _, tc := range []c{{1, 1999, 1999}, {4, 8950, 35800}, {2.5, 33, 83}, {0.1, 33, 3}, {3, 333, 999}, {14, 0, 0}, {1.5, 101, 152}, {-1, 500, -500}} {
        assert.Equal(t, tc.want, LineCents(tc.q, tc.u), "%v×%v", tc.q, tc.u)
    }
}
func TestFormatCents(t *testing.T) { // 0→"0.00", 5→"0.05", -101→"-1.01", 123456789→"1234567.89"
}
func TestRoundQty(t *testing.T) { assert.Equal(t, 0.3, RoundQty(0.1+0.2)) }

func TestSubtractMonthsClamped(t *testing.T) {
    // 2026-08-31→2026-02-28, 2028-08-31→2028-02-29, 2026-03-15→2025-09-15,
    // 2026-12-31→2026-06-30, 2026-01-31→2025-07-31 (all with months=6)
}
func TestCutoff(t *testing.T) { /* Cutoff(2026-11-15, 6) == 2026-05-15; Cutoff(2026-11-15, 0) == 2026-05-15; Cutoff(2026-11-15, 12) == 2025-11-15 */ }
func TestParseDate(t *testing.T) { /* "2026-11-15" ok; "15.11.2026", "", "2026-13-01" → error */ }
func TestFormatDE(t *testing.T) { /* 2026-11-15 → "15.11.2026" */ }

func TestRequestValidate(t *testing.T) {
    // valid: Scope{ItemIDs:[1 id]}, Options{GroupBy:"location", ReferenceDate:"2026-11-15"} → nil
    // each of these → errors.Is(err, ErrInvalidRequest):
    //   ReferenceDate "" ; ReferenceDate "2026-13-01" ; MinMonths -1 ; MinMonths 61 ;
    //   ItemIDs AND Query set ; neither set (empty ItemIDs, nil Query) ;
    //   GroupBy "room" ; GroupBy "tag" with empty GroupTagIDs ; ReceiptTypes ["photo"]
    // ReceiptTypes nil → valid; MinMonths 0 → valid (defaults applied later)
}
```

- [ ] **Step 2: Run them and confirm they fail.** Run: `task go:test -- ./internal/core/services/reporting/customsinventory/...`. Expected: FAIL (undefined identifiers).
- [ ] **Step 3: Implement `model.go`, `money.go` and `dates.go` with the signatures above.** Rounding is half away from zero, applied after snapping to 6 decimals (`math.Round(x*1e6)/1e6`, then `math.Round`). `SubtractMonthsClamped` computes the target month and clamps the day to that month's length (don't rely on `time.AddDate` normalization).
- [ ] **Step 4: Run them and confirm they pass.** Same command; expect PASS.
- [ ] **Step 5: Commit.** `git add backend/internal/core/services/reporting/customsinventory && git commit -m "feat(reporting): customs inventory contract, money and date rules"`

---

### Task 2: Tag index and customs classification

**Files:**
- Create: `…/customsinventory/tags.go`, `…/customsinventory/classify.go`
- Test: `…/customsinventory/tags_test.go`, `…/customsinventory/classify_test.go`

**Interfaces:**
- Consumes: `Tag`, `Item`, `Status`, `Reason`, `IssueCode` (Task 1).
- Produces:

```go
type TagIndex struct{ /* unexported */ }
func NewTagIndex(tags []Tag) TagIndex
func (ti TagIndex) Expand(ids []uuid.UUID) map[uuid.UUID]struct{}           // ids ∉ index dropped; adds all descendants; cycle-safe
func (ti TagIndex) Name(id uuid.UUID) string
func (ti TagIndex) FirstGroup(itemTags, groupTags []uuid.UUID) (uuid.UUID, bool) // first of groupTags (in that order) matched by itemTags directly or via descendant

type Rules struct {
    Reference, Cutoff  types.Date
    MinMonths          int                    // after applying the default
    DutyFree, Dutiable map[uuid.UUID]struct{} // already expanded
    Tags               TagIndex
}
type Classification struct {
    Status    Status
    Reason    Reason
    ReasonTag string      // name of the item's own matching tag (alphabetically first), for tag_* reasons
    Issues    []IssueCode // in FR-11 order
}
func Classify(it Item, r Rules) Classification
func ReasonComment(c Classification, minMonths int) string // FR-10 copy, without the "; Wert fehlt" suffix
```

- [ ] **Step 1: Write the failing tests.** Fixture: reference date 2026-11-15, 6 months, cutoff 2026-05-15. Tags: `P` = "owned" (parent), `P2` = "owned > 2 years" (child of P), `Z` = "dutiable". `Rules.DutyFree = Expand([P])`, `Rules.Dutiable = Expand([Z])`.

```go
func TestClassify(t *testing.T) {
    cases := []struct{ name string; date string; tags []uuid.UUID; want Classification }{
        {"dutiable tag wins", "2019-03-12", []uuid.UUID{Z}, Classification{StatusDutiable, ReasonTagDutiable, "dutiable", nil}},
        {"both tags", "", []uuid.UUID{Z, P2}, Classification{StatusDutiable, ReasonTagDutiable, "dutiable", []IssueCode{IssueConflictTags}}},
        {"after reference date", "2026-11-16", nil, Classification{StatusDutiable, ReasonPurchasedAfterReference, "", []IssueCode{IssuePurchasedAfterReference}}},
        {"reference day itself", "2026-11-15", nil, Classification{StatusDutiable, ReasonOwnedTooShort, "", nil}},
        {"day after cutoff", "2026-05-16", nil, Classification{StatusDutiable, ReasonOwnedTooShort, "", nil}},
        {"date beats tag", "2026-05-16", []uuid.UUID{P2}, Classification{StatusDutiable, ReasonOwnedTooShort, "", []IssueCode{IssueConflictDateTag}}},
        {"on cutoff", "2026-05-15", nil, Classification{StatusDutyFree, ReasonDate, "", nil}},
        {"old + tag, no conflict", "2019-03-12", []uuid.UUID{P2}, Classification{StatusDutyFree, ReasonDate, "", nil}},
        {"no date, child of duty-free tag", "", []uuid.UUID{P2}, Classification{StatusDutyFree, ReasonTagDutyFree, "owned > 2 years", nil}},
        {"no date, no tag", "", nil, Classification{StatusDutiable, ReasonMissingPurchaseDate, "", []IssueCode{IssueMissingPurchaseDate}}},
    }
    // run each through Classify and assert.Equal
    // extra: MinMonths 12 (cutoff 2025-11-15): "2026-01-10" → owned_too_short; "2025-11-15" → date
}
func TestReasonComment(t *testing.T) {
    // owned_too_short, 6 → "Kauf nach Stichtag (< 6 Monate)"; owned_too_short, 12 → "Kauf nach Stichtag (< 12 Monate)";
    // missing_purchase_date → "Kaufdatum fehlt";
    // tag_dutiable/"dutiable" → "zu verzollen lt. Tag „dutiable“"; purchased_after_reference → "Kauf nach Referenzdatum";
    // tag_duty_free/"owned > 2 years" → "zollfrei lt. Tag „owned > 2 years“"; date → ""
}
func TestTagIndexExpand(t *testing.T) {
    // grandchild included; unknown id dropped; two-tag cycle terminates and returns both
}
func TestTagIndexFirstGroup(t *testing.T) {
    // groups [A, B]; item tags [B, childOfA] → A, true ; item tags [X] → uuid.Nil, false
}
```

- [ ] **Step 2: Run them and confirm they fail.** `task go:test -- ./internal/core/services/reporting/customsinventory/... -run 'TestClassify|TestReasonComment|TestTagIndex'` → FAIL.
- [ ] **Step 3: Implement `tags.go` and `classify.go`.** The order of checks is FR-9 steps 1–6, exactly.
- [ ] **Step 4: Run them and confirm they pass.** Same command → PASS.
- [ ] **Step 5: Commit.** `git commit -m "feat(reporting): customs classification and tag expansion"`

---

### Task 3: Build report, verify invariants, golden

**Files:**
- Create: `…/customsinventory/build.go`, `…/customsinventory/verify.go`
- Test: `…/customsinventory/build_test.go`, `…/customsinventory/verify_test.go`, `…/customsinventory/fixture_test.go`, `…/customsinventory/testdata/golden_report.json`

**Interfaces:**
- Consumes:
  - Task 1: all types, `ToCents`, `LineCents`, `RoundQty`, `ParseDate`, `Cutoff`, `DefaultReceiptTypes`, `DefaultMinMonths`;
  - Task 2: `NewTagIndex`, `Classify`, `ReasonComment`;
  - Task 4: `LabelNoLocation`, `LabelOther`;
  - Task 5: `FormatAssetID`, `ReceiptName`.
- Produces:

```go
func Build(in Input) (Report, error) // calls Verify before returning; error wraps the failed invariant
func Verify(r Report) error          // NFR-1 (a)–(g); message starts with "invariant (x):"
func (r Report) Preview() Preview
```

- [ ] **Step 1: Write the shared fixture and the failing tests.** `fixture_test.go` defines `goldenInput()` with GeneratedAt 2026-10-31T10:00:00Z, currency `"eur"`, reference date 2026-11-15, MinMonths 0 (so the default 6 applies), group-by location and merge on. Prices are per piece. Items (asset IDs in brackets):
  - Grandfather clock [101]: 1 × 1200.00, bought 2019-03-12, Home › Living room, one receipt PDF.
  - Sofa [102]: 1 × 890.00, bought 2021-06-01, Home › Living room, plus one receipt that is an external link (`IsLink`) and one `photo` attachment.
  - Dining chair [103]: 4 × 89.50, bought 2020-05-01, Home › Living room (Review Focus 1).
  - Shelf [110]: 2.5 × 0.33, bought 2020-01-01, Home › Office.
  - 14 × "Files" / " files " [131–144]: type Box, no date, no price, tag P2, Home › Office.
  - Files [145]: price 5.00, tag P2, Home › Office (stays a separate line).
  - Espresso machine [146]: 649.90, bought 2026-09-02, tag P2, Home › Kitchen, receipts: a PDF (`receipt`) and a JPG (`attachment`). Options.ReceiptTypes = `["receipt","attachment"]`.
  - Office chair [150]: no date, no price, Home › Office.
  - Toolbox [0]: 1 × 30.00, bought 2018-01-01, no location.
  - Floor lamp [163] (1 × 79.00) and Desk lamp [163] (1 × 45.50), a shared asset ID: each has a receipt, both bought 2026-10-01, both in Home › Living room.
  - A location entity "Basement".
  - An archived item.
  - A sold item.
  - One scope ID not present in `Items`.

```go
func TestBuild_EveryScopeItemOnceOrExcluded(t *testing.T) { /* count of MemberIDs across Groups+Dutiable + len(Excluded) == len(ScopeIDs); no ID twice; Excluded codes: Basement→excluded_location, archived→excluded_archived, sold→excluded_sold, missing→excluded_not_found */ }
func TestBuild_Totals(t *testing.T) {
    r := mustBuild(t, goldenInput())
    // DutyFree: 120000 + 89000 + 35800 + 83 + 0 + 500 + 3000 = 248383 ; Dutiable: 64990 + 0 + 7900 + 4550 = 77440
    assert.Equal(t, int64(248383), r.DutyFreeTotalCents)
    assert.Equal(t, int64(77440), r.DutiableTotalCents)
    assert.Equal(t, int64(325823), r.GrandTotalCents)
}
func TestBuild_LineValueIsQuantityTimesUnitPrice(t *testing.T) { /* Review Focus 1: Dining chair line Quantity 4, UnitCents 8950, LineCents 35800; "Home › Living room" SubtotalCents == 120000 + 89000 + 35800 == 244800 */ }
func TestBuild_MergeIdentical(t *testing.T) { /* one line Name "Files" Qty 14, AssetIDs "000-131"…"000-144" ascending, HasPrice false; Files[145] separate line with UnitCents 500 */ }
func TestBuild_MergeOffKeepsLines(t *testing.T) { /* MergeIdentical=false → 15 Files lines */ }
func TestBuild_MergedDateRange(t *testing.T) { /* two merged items 2019-01-01 & 2021-06-30 → PurchaseDateText "2019-01-01 – 2021-06-30", PurchaseDate zero */ }
func TestBuild_MergedValueUsesSummedQuantity(t *testing.T) { /* own small Input: 3 identical items, qty 0.5 each, price 0.33 → one line Qty 1.5, UnitCents 33, LineCents 50 (CALC-2; summing rounded member values would give 51) */ }
func TestBuild_GroupOrder(t *testing.T) {
    // location mode titles == ["Home › Living room", "Home › Office", "Ohne Lagerort / no location"] (Kitchen holds only dutiable items → no duty-free group)
    // tag mode with GroupTagIDs [B, A]: titles in that order, then "Übrige / other"
}
func TestBuild_TagGroupingFirstMatchWins(t *testing.T) { /* Review Focus 2 */ }
func TestBuild_PositionsContinuous(t *testing.T) { /* Pos 1..n across groups then Dutiable, no gaps */ }
func TestBuild_DutiableSection(t *testing.T) { /* Espresso machine Comments ["Kauf nach Stichtag (< 6 Monate)"]; Office chair Comments ["Kaufdatum fehlt; Wert fehlt"]; Locations shown */ }
func TestBuild_MinMonthsChangesCutoff(t *testing.T) { /* goldenInput with MinMonths 1 → Cutoff 2026-10-15; Espresso machine (2026-09-02) now duty-free; lamps (2026-10-01) now duty-free */ }
func TestBuild_ReceiptNames(t *testing.T) {
    // Grandfather clock → ["Belege/000-101_Rechnung.pdf"]; Espresso machine → ["Belege/000-146_Rechnung.pdf", "Belege/000-146_Rechnung_2.jpg"];
    // shared 163: lower internal ID → "Belege/000-163_Rechnung.pdf", other → "Belege/000-163_b_Rechnung.pdf"  (Review Focus 3)
    // link attachment → not in Receipts, issue receipt_is_link ; photo attachment never included
    // IncludeReceipts=false → r.Receipts empty and all ReceiptNames empty
}
func TestBuild_Issues(t *testing.T) { /* every FR-11 code except negative_value/zero_quantity present with expected item names; issues in FR-11 table order; items sorted by name then asset ID; Toolbox in missing_asset_id; only the two 163 items in duplicate_asset_id; Office chair in missing_price, missing_purchase_date and dutiable_without_receipt; Sofa in receipt_is_link; Espresso machine in conflict_date_tag */ }
func TestBuild_NegativeAndZero(t *testing.T) { /* own small Input: qty 0 × 10.00 → zero_quantity, LineCents 0; price -10.00 → negative_value, LineCents -1000 and totals include it */ }
func TestBuild_Golden(t *testing.T) { /* json.MarshalIndent(Build(goldenInput())) == testdata/golden_report.json; regenerate with -update flag */ }
func TestBuild_UnknownTagIDsIgnored(t *testing.T) { /* DutyFreeTagIDs contains an ID not in Input.Tags → treated as absent */ }

func TestVerify_RejectsCorruption(t *testing.T) {
    // from a valid report, each mutation → error containing "invariant (a)" … "(g)":
    // (a) duplicate a MemberID into another line; (b) LineCents+1; (c) SubtotalCents+1; (d) DutyFreeTotalCents+1;
    // (e) GrandTotalCents+1; (f) remove one entry from Receipts; (g) duplicate a Receipts entry
}
func TestPreview_MatchesReport(t *testing.T) { /* counts and cents equal Report fields; ReferenceDate "2026-11-15"; MinMonths 6; Cutoff "2026-05-15"; Currency "EUR" */ }
```

- [ ] **Step 2: Run them and confirm they fail.** `task go:test -- ./internal/core/services/reporting/customsinventory/... -run 'TestBuild|TestVerify|TestPreview'` → FAIL.
- [ ] **Step 3: Implement `Build`, following FR-2, FR-4…FR-11, FR-18, FR-19.**
  - **Exclusion first:** location, then archived, then sold.
  - **Defaults:** `ReceiptTypes` falls back to `DefaultReceiptTypes` when empty; `MinMonths` 0 becomes `DefaultMinMonths`.
  - **Values:** `UnitCents = ToCents(PurchasePrice)` (per piece); `LineCents = LineCents(Quantity, UnitCents)`, for merged lines with the summed quantity.
  - **Currency:** upper-cased.
  - **Comments:** add `"; Wert fehlt"` when `!HasPrice`. `Build` ends with `return r, Verify(r)`.
- [ ] **Step 4: Implement `Verify` and `Preview`.**
- [ ] **Step 5: Generate the golden and inspect it by hand.** Run `task go:test -- ./internal/core/services/reporting/customsinventory/... -run TestBuild_Golden -update`, then read `testdata/golden_report.json` and check its totals against `TestBuild_Totals`.
- [ ] **Step 6: Run the whole package and confirm it passes.** `task go:test -- ./internal/core/services/reporting/customsinventory/...` → PASS.
- [ ] **Step 7: Commit.** `git commit -m "feat(reporting): build and self-verify customs inventory report"`

---

### Task 4: XLSX writer

**Files:**
- Create: `…/customsinventory/xlsx.go`
- Modify: `backend/go.mod`, `backend/go.sum` (`go get github.com/xuri/excelize/v2@v2.11.0`)
- Test: `…/customsinventory/xlsx_test.go`

**Interfaces:**
- Consumes: `Report`, `Group`, `Line`, `FormatCents`, `FormatDE` (Task 1). It builds its own `testReport()` fixture by hand, so it doesn't wait for Task 3.
- Produces:

```go
const SheetName = "Inventarliste"
const (
    LabelSubtotal      = "Zwischensumme / subtotal"
    LabelDutyFreeTotal = "Total abgabenfrei / total duty-free"
    LabelDutiable      = "Zu verzollende Waren / goods subject to duty"
    LabelDutiableTotal = "Total zu verzollen / total subject to duty"
    LabelGrandTotal    = "Gesamtwert / total value"
    LabelNoLocation    = "Ohne Lagerort / no location"
    LabelOther         = "Übrige / other"
)
func ColumnHeaders(currency string) []string // the 11 FR-14 headers, {CUR} = currency
func WriteXLSX(r Report) ([]byte, error)
```

The column headers, in order:
1. `Pos.`
2. `Menge / quantity`
3. `Bezeichnung / description`
4. `Typ / type`
5. `Lagerort / location`
6. `Inventar-Nr. / asset ID`
7. `Kaufdatum / purchase date`
8. `Einzelwert / unit value {CUR}` (price per piece)
9. `Wiederbeschaffungswert / replacement value {CUR}` (quantity × price per piece)
10. `Beleg / receipt`
11. `Bemerkungen / comments`

(Task 3 uses `LabelNoLocation` and `LabelOther` as group titles. Task 4 is in wave 1, so it has merged before Task 3 starts.)

- [ ] **Step 1: Write the failing tests.** `testReport()` contains:
  - 2 groups: one has a merged line (qty 14, no price) and a qty-4 × 8950-cent line; the other has a qty-2.5 × 33-cent line;
  - 2 dutiable lines: one has no price;
  - ReferenceDate 2026-11-15, MinMonths 6, Cutoff 2026-05-15;
  - header name `=HYPERLINK("http://x")`, phone `+1 555 0100`, a location `-`, title "Möbel – Übersicht".

```go
func TestWriteXLSX_LayoutAndLabels(t *testing.T) {
    // A1 == title; A4 contains "Telefon / phone: +1 555 0100";
    // A5 contains "Referenzdatum / reference date: 15.11.2026" and "Stichtag / cut-off (6 Monate / months): 15.05.2026" and "Währung / currency: EUR";
    // first group title in A7; row 8 == ColumnHeaders("EUR"); subtotal label row contains LabelSubtotal; last row LabelGrandTotal
}
func TestWriteXLSX_LineValueFormula(t *testing.T) { /* Review Focus 1: the qty-4 line has B == 4, H == 89.5, I formula "ROUND(B{r}*H{r},2)" and calc value 358 */ }
func TestWriteXLSX_FormulasRecalculateToCachedValues(t *testing.T) { // NFR-4
    // for every cell in column I with a formula: f.GetCellFormula != "";
    // cached := f.GetCellValue(SheetName, cell, excelize.Options{RawCellValue: true});
    // calc := f.CalcCellValue(SheetName, cell, excelize.Options{RawCellValue: true});
    // both parse to float and equal the expected cents/100 (compare via ToCents(parsed) == expected cents)
    // formulas: lines "ROUND(B{r}*H{r},2)", subtotals "SUBTOTAL(9,I{a}:I{b})", grand "I{x}+I{y}"
}
func TestWriteXLSX_TextNeverFormula(t *testing.T) { /* Review Focus 5: the header/location cells have GetCellFormula=="" and value verbatim */ }
func TestWriteXLSX_CellTypes(t *testing.T) { /* B and H numeric; G of a single-date line is numeric with number format showing "15.05.2019"-style via GetCellValue; G of merged range is the text "2019-01-01 – 2021-06-30"; H empty when !HasPrice */ }
func TestWriteXLSX_EmptyDutiableSection(t *testing.T) { /* Dutiable nil → dutiable total formula "0", cached 0; grand == duty-free */ }
func TestWriteXLSX_FullCalcOnLoad(t *testing.T) { /* f.GetCalcProps() FullCalcOnLoad == true */ }
func TestWriteXLSX_Deterministic(t *testing.T) { /* two writes of the same Report → identical cell values for every used cell */ }
```

- [ ] **Step 2: Run them and confirm they fail.** `task go:test -- ./internal/core/services/reporting/customsinventory/... -run TestWriteXLSX` → FAIL.
- [ ] **Step 3: Implement `WriteXLSX` (FR-13…FR-17).**
  - **Formulas:** comma separators.
  - **Formatting:** number format `#,##0.00` for H and I, and a `dd.mm.yyyy` custom format for G; bold rows; merge A1:K1 only; set column widths.
  - **Cached values:** every formula cell must keep its computed value next to the formula. The test pins this; how you achieve it with excelize is your choice. Set `FullCalcOnLoad`.
- [ ] **Step 4: Run them and confirm they pass.** Same command → PASS. Then run `task go:tidy` and confirm `go.mod` lists excelize `v2.11.0` and nothing else new.
- [ ] **Step 5: Commit.** `git add backend/go.mod backend/go.sum backend/internal/core/services/reporting/customsinventory && git commit -m "feat(reporting): write customs inventory as xlsx with live formulas"`

---

### Task 5: File names and ZIP bundle

**Files:**
- Create: `…/customsinventory/files.go`, `…/customsinventory/bundle.go`
- Test: `…/customsinventory/files_test.go`, `…/customsinventory/bundle_test.go`

**Interfaces:**
- Consumes: `ReceiptFile` (Task 1).
- Produces:

```go
func FormatAssetID(id int64) string                         // same output as repo.AssetID.String(): 123→"000-123", 123456→"123-456", 1234567→"123-4567", ≤0→""
func Slug(title string) string                              // FR-20
func BaseName(title string, generated time.Time) string     // "Inventarliste_{slug}_{yyyy-mm-dd UTC}"
func FileExt(title, mime string) string                     // FR-19
func ReceiptName(assetLabel string, dupIndex, index int, title, mime string) string
//   assetLabel "000-123" or "id-1a2b3c4d"; dupIndex 0→"", 1→"_b", 2→"_c"…; index 0→"", 1→"_2"…
//   → "Belege/{assetLabel}{dup}_Rechnung{suffix}{ext}"
type OpenFunc func(ReceiptFile) (io.ReadCloser, error)
func WriteBundle(w io.Writer, base string, modified time.Time, xlsx []byte, receipts []ReceiptFile, open OpenFunc) error
```

- [ ] **Step 1: Write the failing tests.**

```go
func TestSlug(t *testing.T) {
    // "Inventory list – Shipment A (Truck)" → "inventory-list-shipment-a-truck"
    // "Möbel & Bücher für Gärten" → "moebel-buecher-fuer-gaerten" ; "Straße" → "strasse"
    // "" → "export" ; "!!!" → "export" ; 80×"a b " → len ≤ 60 and no trailing "-"
}
func TestBaseName(t *testing.T) { /* ("Shipment A", 2026-10-31T23:59:00Z) → "Inventarliste_shipment-a_2026-10-31" */ }
func TestFormatAssetID(t *testing.T) { /* table above */ }
func TestFileExt(t *testing.T) {
    // ("rechnung.PDF","application/pdf")→".pdf"; ("scan","image/jpeg")→".jpg"; ("scan","image/png")→".png";
    // ("scan","image/heic")→".heic"; ("x","application/x-foo")→".bin"; ("a.verylongext","application/pdf")→".pdf"
}
func TestReceiptName(t *testing.T) {
    // ("000-123",0,0,"r.pdf","") → "Belege/000-123_Rechnung.pdf"
    // ("000-123",1,0,"r.pdf","") → "Belege/000-123_b_Rechnung.pdf"
    // ("000-123",0,1,"r.jpg","") → "Belege/000-123_Rechnung_2.jpg"
    // ("id-1a2b3c4d",0,0,"","application/pdf") → "Belege/id-1a2b3c4d_Rechnung.pdf"
}
func TestWriteBundle(t *testing.T) {
    // read back with archive/zip: entry names == [base+".xlsx", receipts…] in order; contents equal;
    // no name contains ".." or starts with "/"; open error → WriteBundle returns that error
}
```

- [ ] **Step 2: Run them and confirm they fail.** `task go:test -- ./internal/core/services/reporting/customsinventory/... -run 'TestSlug|TestBaseName|TestFormatAssetID|TestFileExt|TestReceiptName|TestWriteBundle'` → FAIL.
- [ ] **Step 3: Implement `files.go` and `bundle.go`.** Accept a title extension only if it matches `^[a-z0-9]{1,5}$` after lower-casing. `WriteBundle` streams each receipt with `io.Copy` (no full read into memory) and sets every entry's `Modified` to `modified`.
- [ ] **Step 4: Run them and confirm they pass.** Same command → PASS.
- [ ] **Step 5: Commit.** `git commit -m "feat(reporting): customs inventory file naming and zip bundle"`

---

### Task 6: API contract (stub handlers, routes, swagger, TS types)

**Files:**
- Create: `backend/app/api/handlers/v1/v1_ctrl_customs_inventory.go`
- Modify: `backend/app/api/routes.go` (next to line 286, `/reporting/bill-of-materials`)
- Modify: `backend/app/tools/typegen/main.go:51-57` (add a replace rule)
- Regenerate: `backend/app/api/static/docs/*`, `docs/public/api/*`, `frontend/lib/api/types/*`
- Test: the existing `backend/app/api/routes_auth_coverage_test.go`, which picks up new routes automatically.

**Interfaces:**
- Consumes: `customsinventory.Request`, `customsinventory.Preview` (Task 1).
- Produces:
  - routes `POST /v1/reporting/customs-inventory/preview` and `POST /v1/reporting/customs-inventory/export`;
  - controller methods `HandleCustomsInventoryPreview() errchain.HandlerFunc` and `HandleCustomsInventoryExport() errchain.HandlerFunc`;
  - TS types `CustomsInventoryRequest`, `CustomsInventoryScope`, `CustomsInventoryScopeQuery`, `CustomsInventoryOptions`, `CustomsInventoryHeader`, `CustomsInventoryPreview`, `CustomsInventoryIssue`, `CustomsInventoryIssueItem` in `frontend/lib/api/types/data-contracts.ts`.

- [ ] **Step 1: Add both handlers with full swaggo comments.**
  - **Preview:** `@Tags Reporting`, `@Accept json`, `@Produce json`, `@Param payload body customsinventory.Request true "Export request"`, `@Success 200 {object} customsinventory.Preview`, `@Security Bearer`. Built with `adapters.Action`.
  - **Export:** `@Produce application/zip,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`, `@Success 200 {file} binary`. A manual `errchain.HandlerFunc` using `adapters.DecodeBody[customsinventory.Request]`.
  - **For now** both return `validate.NewRequestError(errors.New("not implemented"), http.StatusNotImplemented)` after decoding.
- [ ] **Step 2: Mount the routes** with `chain.ToHandlerFunc(…, userMW...)`, like line 286.
- [ ] **Step 3: Add the typegen rule** `NewReReplace(\` Customsinventory\`, " CustomsInventory"),` to the `replaces` list.
- [ ] **Step 4: Regenerate.** Run `task swag && task typescript-types`. Expected: `grep -c '"customsinventory.Request"' backend/app/api/static/docs/swagger.json` ≥ 1 and `grep -c 'export interface CustomsInventoryRequest' frontend/lib/api/types/data-contracts.ts` = 1. If swag doesn't find the package, add `{{ .INTERNAL }}/core/services/reporting/customsinventory` to the `--dir` list in `Taskfile.yml:32` and rerun.
- [ ] **Step 5: Verify.** `task go:test -- ./app/api/...` → PASS (the anonymous-401 coverage includes both new routes). `task ui:check` → PASS.
- [ ] **Step 6: Commit.** `git commit -m "feat(api): customs inventory endpoints contract"`

---

### Task 7: Service and real handlers

**Files:**
- Create: `backend/internal/core/services/service_customs_inventory.go`
- Modify: `backend/internal/core/services/all.go` (field `CustomsInventory *CustomsInventoryService` in `AllServices`, constructed in `New`)
- Modify: `backend/app/api/handlers/v1/v1_ctrl_customs_inventory.go` (replace the stubs)
- Test: `backend/internal/core/services/service_customs_inventory_test.go` (uses the `main_test.go` bootstrap: `tRepos`, `tGroup`, `tUser`)

**Interfaces:**
- Consumes: Tasks 1, 3, 4 and 5. Repos:
  - `Entities.QueryByGroup(ctx, gid, repo.EntityQuery)` with `Page`/`PageSize` = -1;
  - `Entities.GetOneByGroup(ctx, gid, id)`;
  - `Entities.PathForEntity(ctx, gid, id)` (root first);
  - `Tags.GetAll(ctx, gid)`;
  - `Groups.GroupByID(ctx, gid)` (`.Currency`);
  - `blob.OpenBucket(ctx, repos.Attachments.GetConnString())` with `bucket.NewReader(ctx, repos.Attachments.GetFullPath(att.Path), nil)`.
- Produces:

```go
type CustomsInventoryService struct {
    repos    *repo.AllRepos
    now      func() time.Time // defaults to time.Now
    maxScope int              // defaults to customsinventory.MaxScope
}
type CustomsInventoryFile struct {
    Name        string // "<BaseName>.zip" or ".xlsx"
    ContentType string // "application/zip" | "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
    Write       func(w io.Writer) error
}
func (s *CustomsInventoryService) Preview(ctx Context, req customsinventory.Request) (customsinventory.Preview, error)
func (s *CustomsInventoryService) Export(ctx Context, req customsinventory.Request) (CustomsInventoryFile, error)
func toEntityQuery(q customsinventory.ScopeQuery) repo.EntityQuery
```

- [ ] **Step 1: Write the failing tests.** Create data in a fresh group (as `service_exports_test.go:51-90` does) with:
  - 2 locations;
  - 4 items, with purchase data set via `Entities.UpdateByGroup`;
  - tags "Shipment A" and "owned > 2 years";
  - one item with quantity 3 and a price per piece, so the service test also covers quantity × price;
  - one receipt via `Attachments.Create(ctx, itemID, repo.ItemCreateAttachment{Title: "rechnung.pdf", Content: bytes.NewReader(pdf)}, attachment.TypeReceipt, false)`.

```go
func TestCustomsInventory_PreviewByIDs(t *testing.T)           { /* totals and counts as computed from the 4 items */ }
func TestCustomsInventory_ScopeByQueryMatchesSearch(t *testing.T) { /* tag filter → same ID set as QueryByGroup; "#000-<n>" → only that item (Review Focus 4) */ }
func TestCustomsInventory_ForeignIDsExcluded(t *testing.T)     { /* item ID from another group → Issue excluded_not_found with empty Name; no other data */ }
func TestCustomsInventory_ForeignTagIgnored(t *testing.T)      { /* other group's tag in DutyFreeTagIDs → item without date stays dutiable */ }
func TestCustomsInventory_ExportZip(t *testing.T) {
    // Write into bytes.Buffer; zip has "<base>.xlsx" and "Belege/000-<n>_Rechnung.pdf" whose bytes == pdf; ContentType application/zip
    // two items sharing an asset ID (set via Entities.SetAssetID) → "_b" name present (Review Focus 3)
}
func TestCustomsInventory_ExportXLSXOnly(t *testing.T)         { /* IncludeReceipts=false → Name ends ".xlsx", ContentType xlsx, bytes open with excelize */ }
func TestCustomsInventory_ScopeTooLarge(t *testing.T)          { /* svc.maxScope = 2, 3 IDs → errors.Is(err, customsinventory.ErrScopeTooLarge) */ }
func BenchmarkCustomsInventoryPreview1000(b *testing.B)        { /* seeds 1000 items (b.StopTimer during seeding) in its own group; measures Preview; used for NFR-6 in Task 10 */ }
func TestToEntityQuery(t *testing.T) {
    // Page,PageSize == -1; Fields ["color=red","bad"] → [{Name:"color",Value:"red"}]; Q "#000-123" → Search "" and AssetID 123; Q "lamp" unchanged
}
```

- [ ] **Step 2: Run them and confirm they fail.** `task go:test -- ./internal/core/services/ -run 'TestCustomsInventory|TestToEntityQuery'` → FAIL.
- [ ] **Step 3: Implement the service.**
  - **`buildInput`:** resolve scope IDs (explicit, or via `toEntityQuery`, which mirrors `HandleEntitiesGetAll.extractQuery` including the `#` asset-id rule) and dedupe them. Then return `ErrScopeTooLarge` above `maxScope`.
  - **Loading:** load each ID with `GetOneByGroup`. A not-found ID stays in `ScopeIDs`, but not in `Items`.
  - **Mapping:** map `EntityOut` → `Item`:
    - `IsLocation` from `EntityType.IsLocation`;
    - `LocationPath` from `PathForEntity(Location.ID)` names, cached per location ID;
    - `Attachments` with `IsLink = MimeType == repo.MimeTypeLinkURL`.
  - **Tags and currency:** `Tags.GetAll` → `[]customsinventory.Tag`; currency from `GroupByID`.
  - **Preview:** `Build(in)`, then `.Preview()`.
  - **Export:** `Build`, then `WriteXLSX`. Return a `CustomsInventoryFile` whose `Write` either writes the xlsx bytes, or calls `WriteBundle` with an `OpenFunc` that opens the attachment by ID from the report's `ReceiptFile`. Keep a map from attachment ID to path built in `buildInput`.
  - **Spans:** add OTel spans like `ExportBillOfMaterialsCSV`, with counts only, no names or prices (NFR-8).
- [ ] **Step 4: Wire `all.go`** with `CustomsInventory: &CustomsInventoryService{repos: repos}`.
- [ ] **Step 5: Replace the handler stubs.**
  - `ctx := services.NewContext(r.Context())`.
  - Map errors with `errors.Is(err, customsinventory.ErrInvalidRequest) || errors.Is(err, customsinventory.ErrScopeTooLarge)` → `validate.NewRequestError(err, http.StatusBadRequest)`.
  - **Export:** set `Content-Type` and `Content-Disposition: attachment; filename="<Name>"` before calling `file.Write(w)`.
- [ ] **Step 6: Run the tests and confirm they pass.** `task go:test -- ./internal/core/services/ -run 'TestCustomsInventory|TestToEntityQuery'` → PASS. Then `task go:test` → PASS, and `task go:lint` → clean.
- [ ] **Step 7: Commit.** `git commit -m "feat(api): customs inventory preview and export"`

---

### Task 8: Frontend client, helpers, preferences, search state

**Files:**
- Modify: `frontend/lib/api/classes/reports.ts`
- Create: `frontend/lib/customs-inventory.ts`
- Create: `frontend/composables/use-customs-inventory-search.ts`
- Modify: `frontend/composables/use-preferences.ts` (add `customsInventory?: CustomsInventoryPrefs` to `LocationViewPreferences`; no default needed beyond `undefined`)
- Modify: `frontend/pages/search.vue` (around lines 348–385: write the state after a successful search; clear it in `onBeforeUnmount`)
- Test: `frontend/lib/customs-inventory.test.ts`

**Interfaces:**
- Consumes: the generated `CustomsInventory*` types (Task 6).
- Produces:

```ts
// lib/customs-inventory.ts
export type CustomsInventorySearchContext = { query: CustomsInventoryScopeQuery; total: number };
export type CustomsInventoryPrefs = {
  header: CustomsInventoryHeader; referenceDate: string; minMonths: number; groupBy: "location" | "entityType" | "tag";
  groupTagIds: string[]; dutyFreeTagIds: string[]; dutiableTagIds: string[];
  mergeIdentical: boolean; includeReceipts: boolean; receiptTypes: string[];
};
export const DEFAULT_CUSTOMS_INVENTORY_PREFS: CustomsInventoryPrefs; // groupBy "location", minMonths 6, merge true, receipts true, receiptTypes ["receipt"], empty strings/arrays
export function sanitizePrefs(p: Partial<CustomsInventoryPrefs> | undefined, knownTagIds: Set<string>): CustomsInventoryPrefs;
export function buildRequest(a: { scope: "selected" | "query"; itemIds: string[]; search: CustomsInventorySearchContext | null; prefs: CustomsInventoryPrefs }): CustomsInventoryRequest;
export function cutoffDate(referenceDate: string, months: number): string; // CALC-4, "" for invalid input
export function formatCents(cents: number, currency: string, locale: string): string;
export function filenameFromDisposition(h: string | null, fallback: string): string;

// composables/use-customs-inventory-search.ts
export function useCustomsInventorySearch(): Ref<CustomsInventorySearchContext | null>; // useState("customs-inventory-search", () => null)

// ReportsAPI additions
customsInventoryPreview(body: CustomsInventoryRequest): Promise<TResponse<CustomsInventoryPreview>>; // TResponse from lib/requests/requests.ts
customsInventoryExport(body: CustomsInventoryRequest): Promise<{ blob: Blob; filename: string }>; // throws on !ok, like renderBlob in label-templates.ts
```

- [ ] **Step 1: Write the failing tests.**

```ts
describe("customs-inventory", () => {
  test("buildRequest selected", () => { /* scope.itemIds == ids, scope.query undefined, options copied, header copied */ });
  test("buildRequest query", () => { /* scope.query == search.query, scope.itemIds == [] */ });
  test("buildRequest query without search throws", () => { /* expect(() => …).toThrow() */ });
  test("buildRequest keeps group tag order", () => { /* groupTagIds ["b","a"] stays ["b","a"] */ });
  test("sanitizePrefs drops unknown tags and fills defaults", () => { /* dutyFreeTagIds ["x","known"] → ["known"]; missing receiptTypes → ["receipt"]; missing or out-of-range minMonths (0, 61) → 6 */ });
  test("cutoffDate", () => {
    expect(cutoffDate("2026-08-31", 6)).toBe("2026-02-28");
    expect(cutoffDate("2028-08-31", 6)).toBe("2028-02-29");
    expect(cutoffDate("2026-11-15", 6)).toBe("2026-05-15");
    expect(cutoffDate("2026-11-15", 12)).toBe("2025-11-15");
    expect(cutoffDate("garbage", 6)).toBe("");
  });
  test("formatCents", () => { expect(formatCents(1924990, "EUR", "de-DE")).toBe(new Intl.NumberFormat("de-DE", { style: "currency", currency: "EUR" }).format(19249.9)); });
  test("filenameFromDisposition", () => { /* 'attachment; filename="Inventarliste_x_2026-10-31.zip"' → that name; null → fallback */ });
});
```

- [ ] **Step 2: Run them and confirm they fail.** `cd frontend && pnpm exec vitest run --config ./test/vitest.config.ts lib/customs-inventory.test.ts` → FAIL.
- [ ] **Step 3: Implement the helpers, the composable and the `ReportsAPI` methods.** For export, use `this.http.post` and read `resp.response.headers.get("Content-Disposition")`.
- [ ] **Step 4: Update `search.vue`.** After a successful search, set `useCustomsInventorySearch().value` to:
  - `query`: the same filter fields passed to `api.items.getAll` (`q`, `parentIds`, `entityTypeIds`, `tags`, `negateTags`, `matchAllTags`, `onlyWithPhoto`, `onlyWithoutPhoto`, `includeArchived`, `fields`), without `page`, `pageSize`, `orderBy` and `orderDirection`;
  - `total`: `data.total`.

  On error or an empty result, set `total: 0`. Set the state to `null` in `onBeforeUnmount`.
- [ ] **Step 5: Run the tests and confirm they pass.** Same vitest command → PASS. Then `task ui:check` → PASS.
- [ ] **Step 6: Commit.** `git commit -m "feat(ui): customs inventory client, helpers and search context"`

---

### Task 9: Dialog, menu entry, locales

**Files:**
- Create: `frontend/components/Reports/CustomsInventoryDialog.vue`
- Modify: `frontend/components/ui/dialog-provider/utils.ts`:
  - enum `CustomsInventoryExport = "customs-inventory-export"`;
  - `DialogParamsMap[DialogID.CustomsInventoryExport]: { itemIds: string[]; search?: CustomsInventorySearchContext | null }`.
- Modify: `frontend/layouts/default.vue` (mount `<ReportsCustomsInventoryDialog />` next to `LabelTemplatePrintDialog`)
- Modify: `frontend/components/Item/View/table/data-table-dropdown.vue` (new `DropdownMenuItem` right after the print-labels entry, for both `multi` and single `item`)
- Modify: `frontend/locales/en.json`, `frontend/locales/de.json` (under `components.reports.customs_inventory`)

**Interfaces:**
- Consumes:
  - Task 8: `buildRequest`, `sanitizePrefs`, `cutoffDate`, `formatCents`, `useCustomsInventorySearch`, `api.reports.customsInventoryPreview`/`customsInventoryExport`;
  - existing: `BaseModal`, `TagSelector` (v-model `string[]`, prop `tags`), `FormDatePicker`, `useViewPreferences`, `useDialog().registerOpenDialogCallback` (see `LabelTemplate/PrintDialog.vue:203,408`).
- Produces: the dialog. Nothing later depends on its internals.

- [ ] **Step 1: Add the dialog ID, params type and mount point.**
- [ ] **Step 2: Add the menu entry.** The `@click` opens:

```ts
openDialog(DialogID.CustomsInventoryExport, { params: {
  itemIds: multi ? multi.items.map(r => r.original.id) : [item!.id],
  search: useCustomsInventorySearch().value } })
```
- [ ] **Step 3: Build the dialog to match the mockup.**
  - **Left column:** scope radios (the query option is shown only when `search` is set), grouping segmented control plus ordered `TagSelector` for tag mode, merge checkbox, reference date, minimum months (number input, 1–60), read-only cut-off, duty-free tags, dutiable tags, header fields, receipts checkbox, receipt-type checkboxes (receipt/attachment/warranty/manual).
  - **Right column:** four KPI tiles (positions/items, duty-free, dutiable with count, total), formatted with `formatCents(…, preview.currency, locale)`, plus an issue list. Each issue item links via `NuxtLink` to `/item/{id}`.
  - **Behaviour:**
    - on open: load tags (`api.tags.getAll`) and run `sanitizePrefs(preferences.value.customsInventory, knownIds)`;
    - preview runs debounced (400 ms) on any option change, once a reference date is set;
    - while loading, show a spinner, and the export button is disabled;
    - if preview errors, show an error toast, keep the dialog open and disable the button;
    - the button label is `export_anyway` if any issue has severity `warning`, otherwise `export`;
    - on export success: trigger the download via an object-URL anchor (`a.download = filename`), save the prefs, `toast.success`, close the dialog;
    - on export failure: `toast.error`, and the dialog stays open.
- [ ] **Step 4: Add the locale keys.** Spec §8 keys, plus these issue strings (en / de):

| key | en | de |
|---|---|---|
| `issue_missing_purchase_date` | Purchase date missing → classified as dutiable | Kaufdatum fehlt → als „zu verzollen“ eingestuft |
| `issue_conflict_date_tag` | Conflict: duty-free tag, but bought after the cut-off (date wins) | Konflikt: Zollfrei-Tag, aber Kauf nach dem Stichtag (Datum gilt) |
| `issue_conflict_tags` | Conflict: duty-free and dutiable tag (dutiable wins) | Konflikt: Zollfrei- und Verzollen-Tag (Verzollen gilt) |
| `issue_purchased_after_reference` | Purchase date after the reference date | Kaufdatum nach dem Referenzdatum |
| `issue_missing_price` | No price recorded | Kein Preis erfasst |
| `issue_negative_value` | Negative price or quantity | Negativer Preis oder negative Menge |
| `issue_zero_quantity` | Quantity is 0 | Menge ist 0 |
| `issue_missing_asset_id` | Asset ID missing | Asset-ID fehlt |
| `issue_duplicate_asset_id` | Asset ID used more than once | Asset-ID mehrfach vergeben |
| `issue_dutiable_without_receipt` | Dutiable, but no receipt attached | Zu verzollen, aber kein Beleg angehängt |
| `issue_receipt_is_link` | Receipt is only a link and can't be included | Beleg ist nur ein Link und kann nicht beigelegt werden |
| `issue_excluded_location` / `_archived` / `_sold` / `_not_found` | Not exported: location / archived / sold / not found | Nicht exportiert: Lagerort / archiviert / verkauft / nicht gefunden |
| `kpi_positions` / `kpi_duty_free` / `kpi_dutiable` / `kpi_total` | Lines / items · Duty-free · Dutiable ({count} items) · Total value | Positionen / Artikel · Abgabenfrei · Zu verzollen ({count} Artikel) · Gesamtwert |
| `receipt_types` | Attachment types counted as receipts | Anhangtypen, die als Beleg zählen |
| `header_saved_hint` | Saved in your settings and prefilled next time. | Wird in Ihren Einstellungen gespeichert und beim nächsten Mal vorausgefüllt. |
| `toast_exported` / `toast_failed` | Inventory list downloaded / Export failed | Inventarliste heruntergeladen / Export fehlgeschlagen |

- [ ] **Step 5: Run the checks.** `task ui:fix && task ui:check && (cd frontend && pnpm run lint:ci)` → clean (≤ 1 warning).
- [ ] **Step 6: Manually check against a fresh demo account, not the local dev database.**
  1. Start the stack with `task go:run` and `task ui:dev`.
  2. Register a throwaway user, create 3 items (one with quantity > 1), a tag and a receipt.
  3. Search by the tag, open the dialog from the actions menu, and toggle every option. The KPIs and issues should update.
  4. Export with and without receipts, and open the `.xlsx` in LibreOffice or Numbers.
  5. Screenshot the dialog for the PR.
- [ ] **Step 7: Commit.** `git commit -m "feat(ui): customs inventory export dialog"`

---

### Task 10: End-to-end test and final sweep

**Files:**
- Create: `frontend/test/e2e/customs-inventory.browser.spec.ts`

**Interfaces:**
- Consumes: the whole feature. The demo account (`demo@example.com` / `demodemo`) is seeded because `HBOX_DEMO=true` in the e2e server tasks. Reuse the `login` helper pattern from `wipe-inventory.browser.spec.ts:30`.

- [ ] **Step 1: Write the e2e test.**

```ts
test("exports the customs inventory list from search", async ({ page }) => {
  // login; goto /search; switch to table view; select all on page via header checkbox;
  // open actions menu → click text "Export as Customs Inventory List" (en) ;
  // fill reference date; wait for the KPI "Total value" tile to show a currency amount;
  // uncheck "Include invoices"; const [dl] = await Promise.all([page.waitForEvent("download"), page.getByRole("button", { name: /Export/ }).click()]);
  // expect(dl.suggestedFilename()).toMatch(/^Inventarliste_[a-z0-9-]+_\d{4}-\d{2}-\d{2}\.xlsx$/);
  // re-check with invoices on → filename ends ".zip"
});
```

- [ ] **Step 2: Run it.** `task test:e2e -- customs-inventory` → PASS.
- [ ] **Step 3: Run the full sweep.** `task pr` → green.
- [ ] **Step 3b: Check performance (NFR-6).** Run `cd backend && go test ./internal/core/services/ -run '^$' -bench BenchmarkCustomsInventoryPreview1000 -benchtime 3x`. Expected: ≤ 3 s per op on dev hardware. Record the number in the PR. If it's slower, replace the per-item `GetOneByGroup` loop with a batched repo query in a follow-up PR, and don't block on it.
- [ ] **Step 4: Commit.** `git commit -m "test(e2e): customs inventory export"`
- [ ] **Step 5: Run the user acceptance checklist** (manual, with complete data, not the local dev database). Record the outcome in the PR description.
  - Export one shipment, filtered by its tag.
  - The dialog totals equal the totals in the file.
  - For a line with quantity > 1, the replacement value equals quantity × unit value.
  - Edit one unit price in Excel; the subtotal and totals update.
  - Every dutiable line shows a reason.
  - Every `Belege/` file opens and matches its line.
  - The file opens without a repair prompt in Excel, LibreOffice and Numbers.
