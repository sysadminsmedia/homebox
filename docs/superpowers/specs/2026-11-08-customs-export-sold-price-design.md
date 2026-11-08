# Customs Inventory Export: Items with a Sale Price: Design

**Date:** 2026-11-08 · **Status:** draft, open questions in §4
**Builds on:** [`2026-10-31-customs-inventory-export-design.md`](2026-10-31-customs-inventory-export-design.md) (the "first spec") and [`2026-11-02-customs-inventory-export-round-2-design.md`](2026-11-02-customs-inventory-export-round-2-design.md) ("round 2"). Where they differ, this one wins.

**What this changes:**
- the meaning of "sold" in FR-2 (`excluded_sold`) and FR-28 (no contents loaded below a sold entity);
- FR-11: one more issue code, `sold_to_without_price`.

## 1. What is wrong today

A user who sells an item usually keeps it in HomeBox and fills in the sale fields, rather than deleting it. A sold item has left the collection, so it must not appear on the inventory list.

Today an item counts as sold only when its **sold date** is set:

- it is left out of the sheet, with the info issue `excluded_sold` ("Not exported: sold") in the pre-export check (FR-2, FR-11);
- with "Roll contents into their container" on, nothing below it is loaded (FR-28).

An item with a **sold price** but no sold date is still exported, as if it were still there. Its purchase price is added to the totals.

## 2. Requirements

- **FR-43 Sold.** An entity counts as sold when either:
  - its sold date is set; or
  - its sold price is not 0 after rounding to cents (CALC-1): `ToCents(soldPrice) != 0`. A negative sold price counts as sold.

  "Sold to" and the sold notes only document the sale. They never make an entity sold.
- **FR-44** FR-2 (`sold`), FR-11 (`excluded_sold`) and FR-28 ("never descends below a sold entity") use FR-43. This covers scope entities and content entities alike: a content entity with only a sold price is left out with `excluded_sold`, and its value isn't rolled into its container.
- **FR-45 One rule.** The pure package (`customsinventory`) owns FR-43 as one exported method. The service uses that method for FR-28 and never repeats the rule.
- **FR-46 Sold to without a price.** New issue code `sold_to_without_price`, severity `warning`. It is raised for an entity whose "sold to" isn't empty after trimming and whose sold price is 0 after rounding to cents. It applies to:
  - every entity on a line, scope or content;
  - every entity left out as `excluded_sold` (that is, one with a sold date). See Q3.

  It is never raised for entities left out as `location`, `archived` or `not_found`. The issue has no detail text. In `IssueOrder` it comes right after `container_quantity`, so it is the last warning, before the `excluded_*` info codes.
- **Unchanged:**
  - NFR-1 (a): every sold entity ends in exactly one excluded entry, even when both the date and the price are set;
  - the API request and the response shape. The only API change is the new `IssueCode` value, which reaches the swagger docs and the TS enum `CustomsInventoryIssueCode` through codegen;
  - the database schema (the sold fields are already stored and loaded with the entity);
  - **a sold container doesn't make its contents sold** (Q1). Contents that are in scope themselves keep their own lines.

## 3. Exact copy

UI, key prefix `components.reports.customs_inventory.`:

| key | en | de |
|---|---|---|
| `issue_sold_to_without_price` | “Sold To” is filled in, but the sold price is missing | „Verkauft an“ ist ausgefüllt, aber der Verkaufspreis fehlt |

`issue_excluded_sold` stays "Not exported: sold" / "Nicht exportiert: verkauft". The sheet doesn't change.

## 4. Calls, decisions and open questions

**Calls made while designing (the user can change them):**

1. **Any non-zero sold price counts**, negative too (for example, a fee paid to have something taken away). A price that rounds to 0 cents doesn't, matching `missing_price`.
2. *(Call 2, "sold to alone doesn't count", is now the user's decision Q2 below.)*
3. **Any sold date still counts, even one after the reference date** (unchanged from the first spec).

**Decided by the user:**
- **Q1 Contents of a sold container:** they are not sold because their container is. In-scope contents keep their own lines.
- **Q2 Sold to:** it only documents the sale. Only the sold date and the sold price count. A "sold to" without a sold price is a warning (FR-46).

**Open questions:**
- **Q3 Warning on items already left out.** An item with a sold date and a "sold to", but no sold price, is already left out as sold. Should it still get `sold_to_without_price`? This draft follows the user's answer word for word, so yes. The other option is to raise it only for entities on a line.
- **Calls 1 and 3:** confirm them, or name the ones to change.

## 5. QA strategy

| layer | what | where |
|---|---|---|
| Pure unit | FR-43 table (date only, price only, both, neither, price rounding to 0, negative price); price-only item and price-only content excluded; nothing below a price-only container; FR-46 for a line member, a content entity, a sold-dated item, a "sold to" of spaces only, a "sold to" with a price | `customsinventory/build_test.go`, `contents_test.go`, `model_test.go` |
| Service | real test database: the sold price and "sold to" reach the pure package; no contents loaded below a price-only container | `services/service_customs_inventory_test.go` |
| Codegen | the TS enum has `IssueSoldToWithoutPrice = "sold_to_without_price"` | `frontend/lib/api/types/data-contracts.ts` |
