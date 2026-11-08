# Customs Inventory Export: Items with a Sale Price: Design

**Date:** 2026-11-08 · **Status:** draft, open questions in §4
**Builds on:** [`2026-10-31-customs-inventory-export-design.md`](2026-10-31-customs-inventory-export-design.md) (the "first spec") and [`2026-11-02-customs-inventory-export-round-2-design.md`](2026-11-02-customs-inventory-export-round-2-design.md) ("round 2"). Where they differ, this one wins.

**What this changes:** the meaning of "sold" in FR-2 (`excluded_sold`) and FR-28 (no contents loaded below a sold entity).

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

  Sold to and sold notes don't count on their own.
- **FR-44** FR-2 (`sold`), FR-11 (`excluded_sold`) and FR-28 ("never descends below a sold entity") use FR-43. This covers scope entities and content entities alike: a content entity with only a sold price is left out with `excluded_sold`, and its value isn't rolled into its container.
- **FR-45 One rule.** The pure package (`customsinventory`) owns FR-43 as one exported method. The service uses that method for FR-28 and never repeats the rule.
- **Unchanged:**
  - NFR-1 (a): every sold entity ends in exactly one excluded entry, even when both the date and the price are set;
  - the UI copy, the API request and response, the swagger docs and the TS types;
  - the database schema (the sold price is already stored and loaded with the entity).

## 3. Exact copy

None changes. `issue_excluded_sold` stays "Not exported: sold" / "Nicht exportiert: verkauft".

## 4. Calls made while designing, and open questions

**Calls (the user can change them):**

1. **Any non-zero sold price counts**, negative too (for example, a fee paid to have something taken away). A price that rounds to 0 cents doesn't, matching `missing_price`.
2. **"Sold to" alone doesn't count.** It may be filled in for a planned sale that hasn't happened.
3. **Any sold date still counts, even one after the reference date** (unchanged from the first spec).

**Open questions:**

- **Q1 Contents of a sold container that are in scope themselves.** With rollup on, a sold box's contents aren't loaded. But when the contents match the search on their own (for example, they carry the shipment tag too), each gets its own line today, as if the box had been sold empty. Should they also be left out? This draft keeps today's behaviour.
- **Q2 Sold to.** Should a filled-in "sold to" also count as sold (call 2)?

## 5. QA strategy

| layer | what | where |
|---|---|---|
| Pure unit | FR-43 table (date only, price only, both, neither, price rounding to 0, negative price); price-only item and price-only content excluded; nothing below a price-only container | `customsinventory/build_test.go`, `contents_test.go`, `model_test.go` |
| Service | real test database: the sold price reaches the pure package; no contents loaded below a price-only container | `services/service_customs_inventory_test.go` |
