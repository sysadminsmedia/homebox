# Customs Export: Items with a Sale Price Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Leave an item out of the customs inventory list when its sold price is set, not only when its sold date is.

**Architecture:** The pure package gets the sold price on `Item` and one method, `Item.Sold()`, that holds the rule (FR-43). `exclusion()` and the rollup's `universe()` already go through `exclusion()`, so they follow. The service maps the sold price and uses `Item.Sold()` in `descends()`.

**Tech Stack:** Go (testify).

**Spec:** [`docs/superpowers/specs/2026-11-08-customs-export-sold-price-design.md`](../specs/2026-11-08-customs-export-sold-price-design.md)

## Global Constraints

- No schema, migration, API, swagger, TS type or locale change. If `task --force generate` changes anything except `backend/go.sum`, stop: the plan is wrong.
- `golden_report.json` must not change: new cases go into their own tests, not into `goldenInput()`.
- For a scoped Go run use `cd backend && go test <pkg>`; `task go:test` always runs every package.
- `task go:lint` clean before the PR.
- Tests never use the local dev database.
- **PR slicing.** One PR with Tasks 1–2, against `main`. The body follows the repo's PR conventions (persona first line, signature).

## Review Focus

1. **A sold price that rounds to 0 cents (0.004).** Expected: not sold, the item stays on the sheet. Test: Task 1, `TestItem_Sold`.
2. **A negative sold price.** Expected: sold (spec call 1). Test: Task 1, `TestItem_Sold`.
3. **Both the sold date and the sold price set.** Expected: one `excluded_sold` entry, not two; NFR-1 (a) holds. Test: Task 1, `TestBuild_ExcludesSoldByPrice`.
4. **A price-only sold content entity inside a box, rollup on.** Expected: `excluded_sold`, and the box's value doesn't include it. Test: Task 1, `TestBuild_RollUp_SoldByPriceContentExcluded`.
5. **A price-only sold container, rollup on.** Expected: the container is `excluded_sold`, its contents aren't loaded and don't count toward the 5000 limit. Test: Task 2, `TestCustomsInventory_RollUpSkipsSoldByPrice`.

---

### Task 1: The sold rule in the pure package (FR-43, FR-44, FR-45)

**Files:**
- Modify: `backend/internal/core/services/reporting/customsinventory/model.go` (`Item`)
- Modify: `backend/internal/core/services/reporting/customsinventory/build.go` (`exclusion`)
- Test: `backend/internal/core/services/reporting/customsinventory/model_test.go`, `build_test.go`, `contents_test.go`

**Interfaces:**
- Produces: field `Item.SoldPrice float64` (right after `SoldDate`); method `func (it Item) Sold() bool`.

- [ ] **Step 1: Write the failing tests**

`TestItem_Sold` in `model_test.go`, table-driven over `Item{SoldDate, SoldPrice}`:

| case | SoldDate | SoldPrice | want |
|---|---|---|---|
| neither | zero | 0 | false |
| date only | `d("2026-01-10")` | 0 | true |
| price only | zero | 45.00 | true |
| both | `d("2026-01-10")` | 45.00 | true |
| rounds to 0 cents | zero | 0.004 | false |
| rounds to 1 cent | zero | 0.005 | true |
| negative | zero | -5.00 | true |

`TestBuild_ExcludesSoldByPrice` in `build_test.go`: `in := goldenInput()`; `setItem(&in, id(171), …SoldPrice = 300)` (Tent, price only); `setItem(&in, id(182), …SoldPrice = 80)` (Bookshelf, date and price). `r := mustBuild(t, in)`. Assert:
- `issueNames(r, IssueExcludedSold)` equals `[]string{"Bookshelf", "Tent"}`;
- `r.Excluded` holds exactly one entry with ID `id(182)`;
- `issueNames(r, IssuePurchasedAfterReference)` is empty (Tent was its only member).

`TestBuild_RollUp_SoldByPriceContentExcluded` in `contents_test.go`, next to `TestBuild_RollUp_SoldContentExcluded`, with the same assertions: `setItem(&in, bookID, …SoldPrice = 25)` → one excluded entry, `bookID`, `IssueExcludedSold`; `lineByName(t, r, "Box books").UnitCents` is `3300`.

`TestBuild_RollUp_ContentUnderSoldByPriceIgnored` in `contents_test.go`, next to `TestBuild_RollUp_ContentUnderArchivedIgnored`, with the same assertions: `setItem(&in, pouchID, …SoldPrice = 1)` → one excluded entry, `pouchID`, `IssueExcludedSold`; `penID` not in `r.ScopeIDs`; `"Box books"` unit cents `5550`.

- [ ] **Step 2: Run them and see them fail**

Run: `cd backend && go test ./internal/core/services/reporting/customsinventory/ -run 'Sold'`
Expected: build failure, `Item` has no field `SoldPrice` and no method `Sold`.

- [ ] **Step 3: Implement**

Add `SoldPrice float64` to `Item`, after `SoldDate`. Add `Sold()` in `model.go`: `!it.SoldDate.Time().IsZero() || ToCents(it.SoldPrice) != 0`, with a doc comment that names FR-43. In `exclusion()`, replace the sold-date case with `case it.Sold():`.

- [ ] **Step 4: Run the package tests and see them pass**

Run: `cd backend && go test ./internal/core/services/reporting/customsinventory/`
Expected: PASS, including the golden test with `golden_report.json` unchanged.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/core/services/reporting/customsinventory/
git commit -m "feat(customs): treat items with a sold price as sold"
```

### Task 2: The service passes the sold price on (FR-44, FR-45)

**Files:**
- Modify: `backend/internal/core/services/service_customs_inventory.go` (`toCustomsItem`, `descends`)
- Test: `backend/internal/core/services/service_customs_inventory_test.go`

**Interfaces:**
- Consumes: `Item.SoldPrice`, `Item.Sold()` from Task 1.

- [ ] **Step 1: Write the failing tests**

Both set the sold price with `tClient.Entity.UpdateOneID(id).SetSoldPrice(…).Save(f.ctx)`, and nothing else on the sale fields.

`TestCustomsInventory_SoldPriceExcluded`: `f := newCustomsFixture(t)`; set `f.lamp`'s sold price to 15.00; preview `f.request(f.clock, f.chairs, f.espresso, f.lamp)`. Assert `p.LineCount == 3` and `issueByCode(p, customsinventory.IssueExcludedSold)` lists exactly the lamp.

`TestCustomsInventory_RollUpSkipsSoldByPrice`: `b := withBox(t, f)`; set `b.pouch`'s sold price to 1.00; request `f.request(b.box)` with `RollUpContents = true` and `svc.maxScope = 3` (box, book, pouch: the pen must not be loaded). Assert no error, `p.ContentCount == 1` (the book), `p.GrandTotalCents == 3000`, and the `excluded_sold` issue lists exactly `"Pouch"`.

- [ ] **Step 2: Run them and see them fail**

Run: `cd backend && go test ./internal/core/services/ -run 'TestCustomsInventory_(SoldPriceExcluded|RollUpSkipsSoldByPrice)'`
Expected: FAIL. The first counts 4 lines. The second fails with `ErrScopeTooLarge`, because the pen is loaded below the pouch.

- [ ] **Step 3: Implement**

In `toCustomsItem`, map `SoldPrice: e.SoldPrice`. In `descends`, return `!it.Archived && !it.Sold()`, and keep the comment pointing at FR-28.

- [ ] **Step 4: Run the service tests and see them pass**

Run: `cd backend && go test ./internal/core/services/`
Expected: PASS, including every existing `TestCustomsInventory_*`.

- [ ] **Step 5: Lint and check codegen**

Run: `task go:lint`, then `task --force generate`, then `git status --short`.
Expected: lint clean; after `git checkout backend/go.sum`, no file changed by codegen.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/core/services/service_customs_inventory.go backend/internal/core/services/service_customs_inventory_test.go
git commit -m "feat(customs): pass the sold price to the inventory list"
```
