# Customs Inventory Export: Items with a Sale Price: Design

**Date:** 2026-11-08 · **Status:** ready, every question in §4 is decided
**Builds on:** [`2026-10-31-customs-inventory-export-design.md`](2026-10-31-customs-inventory-export-design.md) (the "first spec") and [`2026-11-02-customs-inventory-export-round-2-design.md`](2026-11-02-customs-inventory-export-round-2-design.md) ("round 2"). Where they differ, this one wins.

**What this changes:**
- the meaning of "sold" in FR-2 (`excluded_sold`) and FR-28 (no contents loaded below a sold entity). It now depends on the reference date;
- FR-11: one more issue code, `sold_to_without_price`.

## 1. What is wrong today

A user who sells an item usually keeps it in HomeBox and fills in the sale fields, rather than deleting it. A sold item has left the collection, so it must not appear on the inventory list.

Today an item counts as sold only when its **sold date** is set:

- it is left out of the sheet, with the info issue `excluded_sold` ("Not exported: sold") in the pre-export check (FR-2, FR-11);
- with "Roll contents into their container" on, nothing below it is loaded (FR-28).

An item with a **sold price** but no sold date is still exported, as if it were still there. Its purchase price is added to the totals.

The other way round, an item sold **after** the reference date is left out today, although it was still there on that date.

## 2. Requirements

- **FR-43 Sold.** Whether an entity counts as sold depends on the reference date:
  - **sold date set:** sold when the sold date is on or before the reference date. A sold date after the reference date means the entity was still there on that date, so it isn't sold, whatever the sold price;
  - **no sold date:** sold when the sold price is above 0 after rounding to cents (CALC-1): `ToCents(soldPrice) > 0`. A negative sold price, or one that rounds to 0 cents, doesn't count.

  "Sold to" and the sold notes only document the sale. They never make an entity sold.
- **FR-44** FR-2 (`sold`), FR-11 (`excluded_sold`) and FR-28 ("never descends below a sold entity") use FR-43. This covers scope entities and content entities alike: a content entity with only a sold price is left out with `excluded_sold`, and its value isn't rolled into its container.
- **FR-45 One rule.** The pure package (`customsinventory`) owns FR-43 as one exported method that takes the reference date. The service uses that method for FR-28, with the request's reference date, and never repeats the rule.
- **FR-46 Sold to without a price.** New issue code `sold_to_without_price`, severity `warning`. It is raised for an entity whose "sold to" isn't empty after trimming and whose sold price doesn't count: it rounds to 0 cents or is negative (`ToCents(soldPrice) <= 0`). It applies to:
  - every entity on a line, scope or content (this includes one sold after the reference date);
  - every entity left out as `excluded_sold`, which here means one with a sold date on or before the reference date (Q3).

  It is never raised for entities left out as `location`, `archived` or `not_found`. The issue has no detail text. In `IssueOrder` it comes right after `content_split` and before `container_quantity`, so it is the last warning, before every info code (the dialog shows issues in this order).
- **Unchanged:**
  - NFR-1 (a): every sold entity ends in exactly one excluded entry, even when both the date and the price are set;
  - the API request and the response shape. The only API change is the new `IssueCode` value, which reaches the swagger docs and the TS enum `CustomsInventoryIssueCode` through codegen;
  - the database schema (the sold fields are already stored and loaded with the entity);
  - **a sold container doesn't make its contents sold** (Q1). Contents that are in scope themselves keep their own lines.

## 3. Exact copy

UI, key prefix `components.reports.customs_inventory.`:

| key | en | de |
|---|---|---|
| `issue_sold_to_without_price` | “Sold To” is filled in, but there is no sold price above 0 | „Verkauft an“ ist ausgefüllt, aber es fehlt ein Verkaufspreis über 0 |

`issue_excluded_sold` stays "Not exported: sold" / "Nicht exportiert: verkauft". The sheet doesn't change.

## 4. Decisions and calls

**Decided by the user:**
- **Q1 Contents of a sold container:** they are not sold because their container is. In-scope contents keep their own lines.
- **Q2 Sold to:** it only documents the sale. Only the sold date and the sold price count. A "sold to" without a sold price is a warning (FR-46).
- **Q3 Warning on items already left out:** yes. An item left out by its sold date still gets `sold_to_without_price` in the dialog.
- **Call 1, sold price:** a negative sold price doesn't count. Only a price above 0 after rounding to cents does.
- **Call 3, sold date:** a sold date only counts up to the reference date.

**Calls made while completing the spec (the user can change them):**
- **4. On the reference date counts.** "Up to the reference date" includes the day itself: an entity sold on the reference date is sold. This follows how `purchased_after_reference` treats a purchase on that date.
- **5. A later sold date wins over the price.** An entity with a sold price and a sold date after the reference date isn't sold. The date says when the sale happened, so the entity was still there on the reference date.
- **6. The warning covers a negative price.** A "sold to" with a negative price gets `sold_to_without_price` too, because the entity stays on the list. That's why the copy says "no sold price above 0" rather than "missing".

## 5. QA strategy

| layer | what | where |
|---|---|---|
| Pure unit | FR-43 table (date before, on and after the reference date, with and without a price; price only; price rounding to 0; negative price; neither); price-only item and price-only content excluded; an item sold after the reference date stays on a line; nothing below a price-only container; FR-46 for a line member, a content entity, a sold-dated item, a negative price, a "sold to" of spaces only, a "sold to" with a price | `customsinventory/build_test.go`, `contents_test.go`, `model_test.go` |
| Service | real test database: the sold price and "sold to" reach the pure package; no contents loaded below a price-only container; contents loaded below a container sold after the reference date | `services/service_customs_inventory_test.go` |
| Codegen | the TS enum has `IssueSoldToWithoutPrice = "sold_to_without_price"` | `frontend/lib/api/types/data-contracts.ts` |
