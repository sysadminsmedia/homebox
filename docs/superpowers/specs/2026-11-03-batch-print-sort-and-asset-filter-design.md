# Batch Print: Sort Order and Asset ID Filter: Design

**Date:** 2026-11-03 · **Status:** ready for review

## 1. What is wrong today

The batch print page (`/label-templates/batch`) lists items to print.

- **The order looks random, but isn't.** The page sends no sort, so the server falls back to name order. Asset IDs then come out jumbled, for example `000-004, 000-006, 000-005, 000-003`. The page has no way to change this.
- **Print order follows ticking order.** "Select all" takes rows in list order, and single rows in the order they were clicked. So a sorted list is the only easy way to a sorted print.
- **The list stops at 100 items without saying so.** The page asks for 100 rows and shows "100 results", even when more match.
- **Asset IDs can only be searched exactly** (`#000-123`). There is no way to say "from this ID on", which is what a person needs when labelling the next batch of items.

Upstream has no issue about this page. Related, but not overlapping: PR #1391 (order-by options on the search page), PR #1730 (`#` search in the location picker) and issue #1534 (asset ID ranges in the label generator).

## 2. Requirements

### R1 Sort order

- **FR-1** With "Items" chosen, the page shows an order-by select and a direction select next to the filters. With "Locations" chosen, both are hidden.
- **FR-2** Order-by options: Asset ID, Name, Location, Created at, Updated at, Quantity, Purchase price, Insured. This is the search page's list without "Archived", because the page never lists archived items.
- **FR-3** The default is Asset ID, ascending.
- **FR-4** The list reloads from the server in the chosen order, so sorting covers all matches, not only the rows on screen. Only the latest request counts: a slower, older response never replaces the list or the selection order.
- **FR-5** Changing the sort also re-orders the selection:
  - selected items in the list take their new list order;
  - selected items no longer in the list stay selected, after them, in their old order.

  This replaces any manual drag order. Changing a filter or search term leaves the selection order alone.
- **FR-6** When two items tie on the chosen field, the server orders them by ID, so a tie can't change between requests. This applies to every order-by value and to every caller of the entity query.

### R2 Result limit

- **FR-7** When more items match than the 100 shown, the page says so: `Showing the first {shown} of {total} items. Narrow the search to reach the others.`

### R3 Asset ID filter

- **FR-8** The search term may be an asset ID comparison. It starts with `#` or `asset-id` (also `asset id` and `assetid`, in any case), followed by one or two comparisons:

  | term | matches |
  |---|---|
  | `#>=100` | 000-100 and up |
  | `asset-id >= 000-100` | the same |
  | `#>100` | 000-101 and up |
  | `#<=200` | up to 000-200 |
  | `#<200` | up to 000-199 |
  | `#>=100 <=200` | 000-100 to 000-200 |
  | `#>=100<=200` | the same |

  Spaces around an operator and between the two comparisons are optional. A value is anything `#000-123` accepts today.
- **FR-9** Entities without an asset ID never match a comparison.
- **FR-10** A range that can't match (`#>=200 <=100`, or `#>` the highest possible ID) returns no results, not an error.
- **FR-11** A term that doesn't fit FR-8 stays a normal text search, as `#abc` does today. This includes text next to a comparison (`#>=100 books`).
- **FR-12** The filter works wherever the entity search runs: `GET /v1/entities`, so also the search page, and the customs export's "All items in the current search", which reruns the search on the server. Both must resolve a term the same way.
- **FR-13** `#000-123` keeps its exact-match meaning.
- **FR-14** The batch page shows a hint under the search box: `Filter by asset ID with #>=100, or #>=100 <=200 for a range.`
- **FR-15** The search page's "Query ID" line under the search box only appears for an exact `#000-123` term. Today it would show "invalid" for a valid comparison.

## 3. Exact copy

Key prefix `pages.label_templates.batch.`:

| key | en | de |
|---|---|---|
| `results_truncated` | Showing the first {shown} of {total} items. Narrow the search to reach the others. | Die ersten {shown} von {total} Artikeln werden angezeigt. Grenzen Sie die Suche ein, um die übrigen zu sehen. |
| `asset_filter_hint` | Filter by asset ID with #>=100, or #>=100 <=200 for a range. | Nach Asset-ID filtern mit #>=100, oder #>=100 <=200 für einen Bereich. |

`de.json` has no `pages.label_templates` block yet, so the rest of this page stays in English until it is translated. `items.asc` and `items.desc` also exist only in English; add them to `de.json` as `Aufsteigend` and `Absteigend`. `items.order_direction` exists in neither file, though the search page already uses it; add it as `Order Direction` and `Sortierrichtung`, and use it as the direction select's label.

## 4. Calls made while designing (the user can change them)

1. **Asset ID is the default sort.** Labels are usually printed in ID order. Name order stays one click away.
2. **The sort isn't remembered.** It resets to the default on each visit, like the page's other filters.
3. **A sort change overrides manual drag order (FR-5).** The person asked for an order, so the print follows it. Dragging afterwards still works.
4. **Items without an asset ID sort first when ascending,** as they do on the search page today. `#>=1` hides them.
5. **The search box carries the filter, not a new field.** `#` already means "asset ID" there, so the search page and the customs export get it too.
6. **The 100-row limit stays, with a message.** Lifting it needs paging on this page, which isn't needed to solve the problem: filters reach the rest.
7. **No location sorting.** The location list keeps its current order.

## 5. QA strategy

| layer | what | where |
|---|---|---|
| Go unit | FR-8 table, FR-10, FR-11, FR-13 | `repo/asset_id_type_test.go` |
| Repository | FR-6 tie-break; FR-9 and range filtering on a real test database | `repo/repo_entities_test.go` |
| Service | FR-12: the customs scope resolves a comparison | `services/service_customs_inventory_test.go` |
| Frontend unit | FR-5 selection order | `frontend/lib/label-batch.test.ts` |
| E2E | sort control, stale responses (FR-4), asset ID filter and hint on the batch page | `frontend/test/e2e/label-batch.browser.spec.ts` |
