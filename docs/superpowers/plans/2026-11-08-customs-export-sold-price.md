# Customs Export: Items with a Sale Price Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Leave an item out of the customs inventory list when it was sold by the reference date: by a sold date on or before that date, or, without a sold date, by a sold price above 0. Warn about a "sold to" without such a price.

**Architecture:** The pure package gets the sold price and "sold to" on `Item`, one method `Item.SoldBy(ref)` that holds the rule (FR-43), and one new warning code (FR-46). `exclusion()` and the rollup's `universe()` take the reference date and use `SoldBy`. The service maps both fields, parses the reference date once and uses `SoldBy` in `descends()`.

**Tech Stack:** Go (testify); swag + typegen codegen; i18n JSON.

**Spec:** [`docs/superpowers/specs/2026-11-08-customs-export-sold-price-design.md`](../specs/2026-11-08-customs-export-sold-price-design.md)

## Global Constraints

- No schema or migration change.
- Codegen: after adding the issue code, run `task --force generate` (plain `task generate` can report "up to date") and revert `backend/go.sum` with `git checkout backend/go.sum`. Only the swagger docs (`backend/app/api/static/docs/`, `docs/public/api/`) and `frontend/lib/api/types/data-contracts.ts` may change, and only by the new enum value. Anything else: stop, the plan is wrong. Never hand-edit generated files.
- UI copy goes into `frontend/locales/en.json` and `de.json` only, exactly as in spec §3.
- `golden_report.json` must not change: new cases go into their own tests, not into `goldenInput()`. (The golden Bookshelf is sold on 2026-01-10, before the reference date 2026-11-15, so it stays excluded.)
- Dates compare the way `classify.go` compares the purchase date: `t.After(ref.Time())` on `types.Date.Time()`.
- For a scoped Go run use `cd backend && go test <pkg>`; `task go:test` always runs every package.
- `task go:lint`, `task ui:fix` (at most 1 warning in `lint:ci`) and `task ui:check` clean before the PR.
- Tests never use the local dev database.
- **PR slicing.** One PR with Tasks 1–3, against `main`. The body follows the repo's PR conventions (persona first line, signature).

## Review Focus

1. **Sold on the reference date itself.** Expected: sold, left out (spec call 4). Test: Task 1, `TestItem_SoldBy` ("date on reference").
2. **A sold price together with a sold date after the reference date.** Expected: not sold, the item stays on the sheet (spec call 5). Tests: Task 1, `TestItem_SoldBy` ("date after, with price") and `TestBuild_SoldAfterReferenceListed`.
3. **A sold price that rounds to 0 cents (0.004), or a negative one.** Expected: not sold, the item stays on the sheet; with a "sold to" it also gets the warning. Tests: Task 1, `TestItem_SoldBy`; Task 2, `TestBuild_SoldToWithoutPrice` (Shelf, Dining chair).
4. **Both the sold date and the sold price set.** Expected: one `excluded_sold` entry, not two; NFR-1 (a) holds. Test: Task 1, `TestBuild_ExcludesSoldByPrice`.
5. **A container sold after the reference date, rollup on.** Expected: its contents are loaded and rolled in like any other container's. Test: Task 3, `TestCustomsInventory_RollUpBelowSoldAfterReference`.

---

### Task 1: The sold rule in the pure package (FR-43, FR-44, FR-45)

**Files:**
- Modify: `backend/internal/core/services/reporting/customsinventory/model.go` (`Item`)
- Modify: `backend/internal/core/services/reporting/customsinventory/build.go` (`exclusion`, its call in `Build`)
- Modify: `backend/internal/core/services/reporting/customsinventory/contents.go` (`universe`)
- Test: `backend/internal/core/services/reporting/customsinventory/model_test.go`, `build_test.go`, `contents_test.go`

**Interfaces:**
- Produces: field `Item.SoldPrice float64` (right after `SoldDate`); method `func (it Item) SoldBy(ref types.Date) bool`. Unexported: `exclusion(it Item, found bool, ref types.Date) (IssueCode, bool)` and `universe(in Input, byID map[uuid.UUID]Item, ref types.Date) []uuid.UUID`.

- [ ] **Step 1: Write the failing tests**

`TestItem_SoldBy` in `model_test.go`, table-driven over `Item{SoldDate, SoldPrice}`, all with `ref := d("2026-11-15")`:

| case | SoldDate | SoldPrice | want |
|---|---|---|---|
| neither | zero | 0 | false |
| date before | `d("2026-01-10")` | 0 | true |
| date on reference | `d("2026-11-15")` | 0 | true |
| date after | `d("2026-11-16")` | 0 | false |
| date before, with price | `d("2026-01-10")` | 45.00 | true |
| date after, with price | `d("2026-11-16")` | 45.00 | false |
| price only | zero | 45.00 | true |
| rounds to 0 cents | zero | 0.004 | false |
| rounds to 1 cent | zero | 0.005 | true |
| negative | zero | -5.00 | false |

`TestBuild_ExcludesSoldByPrice` in `build_test.go`: `in := goldenInput()`; `setItem(&in, id(171), …SoldPrice = 300)` (Tent, price only); `setItem(&in, id(182), …SoldPrice = 80)` (Bookshelf, date and price). `r := mustBuild(t, in)`. Assert:
- `issueNames(r, IssueExcludedSold)` equals `[]string{"Bookshelf", "Tent"}`;
- exactly one `r.Excluded` entry has ID `id(182)` (NFR-1 a: date and price together give one entry, not two);
- `issueNames(r, IssuePurchasedAfterReference)` is empty (Tent was its only member).

`TestBuild_SoldAfterReferenceListed` in `build_test.go`: `in := goldenInput()`; `setItem(&in, id(182), …SoldDate = d("2026-11-16"), SoldPrice = 80)` (Bookshelf, sold the day after); `setItem(&in, id(102), …SoldPrice = -20)` (Sofa). `r := mustBuild(t, in)`. Assert `issueNames(r, IssueExcludedSold)` is empty, and no `r.Excluded` entry has ID `id(182)` or `id(102)`.

`TestBuild_RollUp_SoldByPriceContentExcluded` in `contents_test.go`, next to `TestBuild_RollUp_SoldContentExcluded`, with the same assertions: `setItem(&in, bookID, …SoldPrice = 25)` → one excluded entry, `bookID`, `IssueExcludedSold`; `lineByName(t, r, "Box books").UnitCents` is `3300`.

`TestBuild_RollUp_ContentSoldAfterReferenceRolledIn` in `contents_test.go`, next to it: `setItem(&in, bookID, …SoldDate = d("2026-11-16"))` → `r.Excluded` is empty and `"Box books"` unit cents `6300`.

`TestBuild_RollUp_ContentUnderSoldByPriceIgnored` in `contents_test.go`, next to `TestBuild_RollUp_ContentUnderArchivedIgnored`, with the same assertions: `setItem(&in, pouchID, …SoldPrice = 1)` → one excluded entry, `pouchID`, `IssueExcludedSold`; `penID` not in `r.ScopeIDs`; `"Box books"` unit cents `5550`.

- [ ] **Step 2: Run them and see them fail**

Run: `cd backend && go test ./internal/core/services/reporting/customsinventory/ -run 'Sold'`
Expected: build failure, `Item` has no field `SoldPrice` and no method `SoldBy`.

- [ ] **Step 3: Implement**

- `model.go`: add `SoldPrice float64` to `Item`, after `SoldDate`. Add `SoldBy(ref)` with a doc comment that names FR-43: with a sold date, `!soldDate.After(ref.Time())`; without one, `ToCents(it.SoldPrice) > 0`.
- `build.go`: `exclusion` takes `ref` and its sold case becomes `case it.SoldBy(ref):`. `Build` passes the `ref` it already parsed, to `exclusion` and to `universe`.
- `contents.go`: `universe` takes `ref` and passes it to `exclusion`.

- [ ] **Step 4: Run the package tests and see them pass**

Run: `cd backend && go test ./internal/core/services/reporting/customsinventory/`
Expected: PASS, including the golden test with `golden_report.json` unchanged.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/core/services/reporting/customsinventory/
git commit -m "feat(customs): count an item as sold by the reference date, by date or by price"
```

### Task 2: The "sold to without price" warning (FR-46)

**Files:**
- Modify: `backend/internal/core/services/reporting/customsinventory/model.go` (`Item`, `IssueCode` consts, `IssueOrder`)
- Modify: `backend/internal/core/services/reporting/customsinventory/build.go` (`Build`'s exclusion loop, `itemIssues`)
- Modify: `frontend/locales/en.json`, `frontend/locales/de.json`
- Generated: `backend/app/api/static/docs/*`, `docs/public/api/*`, `frontend/lib/api/types/data-contracts.ts`
- Test: `backend/internal/core/services/reporting/customsinventory/build_test.go`, `contents_test.go`

**Interfaces:**
- Consumes: `Item.SoldPrice` from Task 1.
- Produces: field `Item.SoldTo string` (right after `SoldPrice`); `IssueSoldToWithoutPrice IssueCode = "sold_to_without_price"`, in `IssueOrder` right after `IssueContentSplit` (the last warning; `IssueContainerQuantity` is info); unexported `soldToWithoutPrice(it Item) bool` and `addExcludedIssues(issues *issueSet, code IssueCode, it Item, item IssueItem)` in `build.go`.

- [ ] **Step 1: Write the failing tests**

`TestBuild_SoldToWithoutPrice` in `build_test.go`: `in := goldenInput()`, then `setItem` with `SoldTo`:

| item | SoldTo | SoldPrice | warning? | why |
|---|---|---|---|---|
| `id(102)` Sofa | `"Jane Doe"` | 0 | yes | still on a line |
| `id(110)` Shelf | `"Jane Doe"` | 0.004 | yes | rounds to 0 cents, still on a line |
| `id(103)` Dining chair | `"Jane Doe"` | -5.00 | yes | negative, still on a line |
| `id(182)` Bookshelf | `"Jane Doe"` | 0 | yes | left out by its sold date (spec Q3) |
| `id(171)` Tent | `"Jane Doe"` | 300 | no | sold by price |
| `id(101)` Grandfather clock | `"   "` | 0 | no | spaces only |
| `id(181)` Old radio | `"Jane Doe"` | 0 | no | archived |

`r := mustBuild(t, in)`. Assert:
- `issueNames(r, IssueSoldToWithoutPrice)` equals `[]string{"Bookshelf", "Dining chair", "Shelf", "Sofa"}`;
- that issue's `Severity` is `SeverityWarning`, and every item's `Detail` is `""`;
- no `r.Excluded` entry has ID `id(102)`, `id(103)` or `id(110)`;
- `slices.Index(IssueOrder, IssueSoldToWithoutPrice)` is `slices.Index(IssueOrder, IssueContentSplit) + 1`, and every code after it in `IssueOrder` has `SeverityOf` = `SeverityInfo`.

`TestBuild_RollUp_SoldToWithoutPriceContent` in `contents_test.go`: `in := containerInput()`; `setItem(&in, bookID, …SoldTo = "Jane Doe")`. Assert `issueNames(r, IssueSoldToWithoutPrice)` is `[]string{"Book"}`, `r.Excluded` is empty and `lineByName(t, r, "Box books").UnitCents` is `6300` (the book is still rolled in).

In `TestBuild_Issues`, add `IssueSoldToWithoutPrice` to the list of codes the golden data doesn't raise, and update the message to "every code except negative_value, zero_quantity, the rollup codes and sold_to_without_price, in order".

- [ ] **Step 2: Run them and see them fail**

Run: `cd backend && go test ./internal/core/services/reporting/customsinventory/ -run 'SoldTo|TestBuild_Issues'`
Expected: build failure, no field `SoldTo` and no `IssueSoldToWithoutPrice`.

- [ ] **Step 3: Implement the code and the warning**

- `model.go`: add `SoldTo string` after `SoldPrice`; add the const after `IssueContentSplit`, and put it right after `IssueContentSplit` in `IssueOrder`, before `IssueContainerQuantity`. `SeverityOf` already returns `warning` for it.
- `build.go`: `soldToWithoutPrice` is `strings.TrimSpace(it.SoldTo) != "" && ToCents(it.SoldPrice) <= 0`.
  - `addExcludedIssues` adds `code` for `item`, plus `IssueSoldToWithoutPrice` when `code == IssueExcludedSold` and `soldToWithoutPrice(it)`. It replaces the `issues.add(code, …)` call in `Build`'s exclusion loop, so `Build` doesn't grow (gocyclo).
  - In `itemIssues`, add `IssueSoldToWithoutPrice` with `ref` when `soldToWithoutPrice(e.it)`.

- [ ] **Step 4: Run the package tests and see them pass**

Run: `cd backend && go test ./internal/core/services/reporting/customsinventory/`
Expected: PASS, `golden_report.json` unchanged.

- [ ] **Step 5: Add the copy and regenerate**

Add `issue_sold_to_without_price` to `components.reports.customs_inventory` in both locale files, right after `issue_content_split`, with the exact en and de text from spec §3. Then run `task --force generate` and `git checkout backend/go.sum`.
Expected: `git diff --stat` shows only the two locale files, `backend/app/api/static/docs/`, `docs/public/api/` and `data-contracts.ts`. `CustomsInventoryIssueCode` has `IssueSoldToWithoutPrice = "sold_to_without_price"` between `IssueContentSplit` and `IssueContainerQuantity`.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/core/services/reporting/customsinventory/ frontend/locales/en.json frontend/locales/de.json backend/app/api/static/docs/ docs/public/api/ frontend/lib/api/types/
git commit -m "feat(customs): warn about a sold-to without a sold price"
```

### Task 3: The service passes the sale fields and the reference date on (FR-44, FR-45, FR-46)

**Files:**
- Modify: `backend/internal/core/services/service_customs_inventory.go` (`buildInput`, `loadContents`, `descends`, `toCustomsItem`)
- Test: `backend/internal/core/services/service_customs_inventory_test.go`

**Interfaces:**
- Consumes: `Item.SoldPrice`, `Item.SoldTo`, `Item.SoldBy(ref types.Date)`, `IssueSoldToWithoutPrice` from Tasks 1–2; `customsinventory.ParseDate`.
- Produces (unexported): `descends(it customsinventory.Item, ref types.Date) bool`; `loadContents` gets a `ref types.Date` parameter right after `frontier`.

- [ ] **Step 1: Write the failing tests**

All four set the sale fields with `tClient.Entity.UpdateOneID(id).SetSoldPrice(…)`, `.SetSoldTo(…)` or `.SetSoldDate(…)`, then `.Save(f.ctx)`. The request's reference date is the fixture's `2026-11-15`.

`TestCustomsInventory_SoldPriceExcluded`: `f := newCustomsFixture(t)`; set `f.lamp`'s sold price to 15.00; preview `f.request(f.clock, f.chairs, f.espresso, f.lamp)`. Assert `p.LineCount == 3` and `issueByCode(p, customsinventory.IssueExcludedSold)` lists exactly the lamp.

`TestCustomsInventory_SoldToWithoutPrice`: set `f.lamp`'s sold to `"Jane Doe"`; same preview. Assert `p.LineCount == 4` and `issueByCode(p, customsinventory.IssueSoldToWithoutPrice)` lists exactly `"Old lamp"`.

`TestCustomsInventory_RollUpSkipsSoldByPrice`: `b := withBox(t, f)`; set `b.pouch`'s sold price to 1.00; request `f.request(b.box)` with `RollUpContents = true` and `svc.maxScope = 3` (box, book, pouch: the pen must not be loaded). Assert no error, `p.ContentCount == 1` (the book), `p.GrandTotalCents == 3000`, and the `excluded_sold` issue lists exactly `"Pouch"`.

`TestCustomsInventory_RollUpBelowSoldAfterReference`: `b := withBox(t, f)`; set `b.pouch`'s sold date to 2026-12-01 (UTC) and its sold price to 1.00; request `f.request(b.box)` with `RollUpContents = true`. Assert no error, `p.ContentCount == 3`, `p.GrandTotalCents == 3450` (as in the plain rollup test), and `issueByCode(p, customsinventory.IssueExcludedSold)` is nil.

- [ ] **Step 2: Run them and see them fail**

Run: `cd backend && go test ./internal/core/services/ -run 'TestCustomsInventory_(SoldPriceExcluded|SoldToWithoutPrice|RollUpSkipsSoldByPrice|RollUpBelowSoldAfterReference)'`
Expected: FAIL. The first counts 4 lines. The second finds no `sold_to_without_price` issue. The third fails with `ErrScopeTooLarge`, because the pen is loaded below the pouch. The fourth counts 2 contents, because the pen isn't loaded below the pouch's sold date.

- [ ] **Step 3: Implement**

- `buildInput`: right after the scope-size check, parse the reference date with `customsinventory.ParseDate(req.Options.ReferenceDate)`. On error return `fmt.Errorf("%w: reference date must be YYYY-MM-DD", customsinventory.ErrInvalidRequest)`, the same text as `Build`. Pass `ref` to `descends` and to `loadContents`.
- `descends` returns `!it.Archived && !it.SoldBy(ref)`; keep the comment pointing at FR-28.
- `toCustomsItem` maps `SoldPrice: e.SoldPrice` and `SoldTo: e.SoldTo`.

- [ ] **Step 4: Run the service tests and see them pass**

Run: `cd backend && go test ./internal/core/services/`
Expected: PASS, including every existing `TestCustomsInventory_*`.

- [ ] **Step 5: Lint**

Run: `task go:lint`, `task ui:fix`, `task ui:check`.
Expected: all clean; `lint:ci` at most 1 warning.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/core/services/service_customs_inventory.go backend/internal/core/services/service_customs_inventory_test.go
git commit -m "feat(customs): pass the sale fields and the reference date to the inventory list"
```
