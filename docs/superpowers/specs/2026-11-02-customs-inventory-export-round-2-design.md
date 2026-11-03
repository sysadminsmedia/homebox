# Customs Inventory Export, Round 2: Design

**Date:** 2026-11-02 · **Status:** written from live-test feedback, ready for planning
**Builds on:** [`2026-10-31-customs-inventory-export-design.md`](2026-10-31-customs-inventory-export-design.md) (the "first spec"). Where the two differ, this one wins.

**What this changes in the first spec:**
- G5 and §3 step 2: the export no longer starts from the item table's actions menu (R3).
- FR-7: merging can match on the first word of the name (R1).
- FR-13 row 5 and the total rows: they show piece counts (R4).
- FR-21: two more saved options (R1, R2).
- §10, the call "Child items aren't rolled up into their container": reversed (R2).

## 1. What live testing showed

- **R1 Similar names.** Merging works, but boxes named "Books & deco", "books and deco" and "Books & material" stay separate lines. A fuzzy match would be too complex. Matching on the first word ("Books") is enough.
- **R2 Nested items.** Items can sit inside other items, for example a box that holds books. The list should show only the box. The box's value must include everything inside it, plus the box's own price if it has one.
- **R3 Entry point.** The export entry shows up in every item table, including the table of sub-items on an item page. That isn't where a list is made.
- **R4 Counts.** The dialog shows "Lines / items" and "Dutiable (n items)", and the sheet says "n Artikel / items". All three count database entries, not pieces. A box entry with quantity 3 counts as 1. The values were already right: every line is quantity × price per piece.

## 2. Requirements

### R1 Merge by first word
- **FR-23** New option `mergeFirstWord` (bool, default false). It only has an effect when `mergeIdentical` is on.
- **FR-24** The first word of a name is the first run of Unicode letters and digits in the normalized name (trimmed, inner spaces collapsed). Leading characters that aren't letters or digits are skipped. If the name has no letter or digit, the whole normalized name is used. Examples:

  | name | first word |
  |---|---|
  | `Books & deco` | `Books` |
  | `  books   and deco` | `books` |
  | `Bücher/Deko` | `Bücher` |
  | `3D-Drucker` | `3D` |
  | `& more` | `more` |
  | `—` | `—` |

- **FR-25** With the option on, the merge key uses the lower-case first word instead of the lower-case full name. The other key parts stay: group, entity type, customs status and unit value. Items with different unit values stay on separate lines, because column H holds one price per piece.
- **FR-26** Line name: the full name of the line's first member when all members share the same full name (ignoring case); otherwise that member's first word. A line with one member keeps its full name. With the option on, entries are sorted by first word (ignoring case) before the existing order (FR-5), so lines that share a first word sit together.

### R2 Containers and their contents
- **FR-27** New option `rollUpContents` (bool). The API default is false, which keeps the first spec's behaviour. The dialog default is true.
- **FR-28 Contents.** With the option on, the server also loads the descendants of every scope entity, through the parent links, level by level. It walks only through item children:
  - it never descends into a location entity;
  - it never descends below an archived or sold entity;
  - archived children aren't loaded, as in the search.

  Scope and contents together count toward the 5000 limit (FR-3).
- **FR-29 Accounting.** Every scope entity and every loaded content entity ends up in exactly one place:
  - on a line, as a member;
  - absorbed into a line, as content;
  - or in the excluded list.

  This extends NFR-1 (a) to content entities.
- **FR-30 Owner.** For each included entity, walk up its parent chain one step at a time. Step to the parent only while the parent is an included entity of this export, has **the same customs status** and has **quantity exactly 1**. The owner is the last entity reached. If the first step already fails, the entity gets its own line. The walk never skips a container: a duty-free scarf in a dutiable bag inside a duty-free box gets its own line next to the bag, and isn't rolled into the box. If the parent links form a cycle (broken data), the link is cut at the cycle member with the smallest ID, so every chain ends.
- **FR-31 Value.** A line member's unit value is its own price per piece plus the sum of its absorbed contents' values. Each content's value is its quantity × its price per piece (CALC-2). The owner's quantity is 1, so line value = unit value, and the formula `=ROUND(B*H,2)` still holds. The merge key (FR-7, FR-25) uses this rolled-up unit value. A price is missing (CALC-6) when the rolled-up unit value is 0.
- **FR-32 Comments.**
  - A line with contents adds the contents' comments, each once. It then adds `inkl. Inhalt ({n} Stk.)`, where n is the sum of the absorbed contents' quantities, written with a German decimal comma.
  - An entity on its own line whose direct parent is an included entity of this export gets `Inhalt von {asset label of the parent}`.
  - Asset labels are formatted as in FR-19 (`000-123`, or `id-1a2b3c4d`).
- **FR-33 Receipts.** Absorbed contents' receipts appear on their owner's line and in the ZIP. FR-18 and FR-19 apply unchanged.
- **FR-34 Issues.** Every per-entity issue applies to content entities exactly as to scope entities, with one exception: `missing_price` is not raised for an entity whose own price is 0 when it absorbs contents and its rolled-up value isn't 0. Two new codes:

  | code | severity | when | detail |
  |---|---|---|---|
  | `content_split` | warning | an entity's direct parent is an included entity of this export with a different customs status, so the entity has its own line | the parent's asset label |
  | `container_quantity` | info | an entity with quantity ≠ 1 has included children, so nothing is rolled into it | its quantity |

- **FR-35 Inherited tags.** With the option on, an entity is classified (FR-9) with its own tags plus the tags of every ancestor on its chain of included parents. This applies to both duty-free and always-dutiable tags. The date still beats a duty-free tag (FR-9 step 3).
- **FR-36** The preview adds `contentCount`, the number of absorbed entities.

### R3 Entry point
- **FR-37** No item table's actions menu has the export entry any more.
- **FR-38** `/search` shows a button, "Export inventory list …", in the header of the results. It opens the dialog with:
  - the rows selected in the results table (possibly none);
  - the current search.

  The scope defaults to "Only selected items" when rows are selected, otherwise to "All items in the current search". The button is disabled until the first search has finished.
- **FR-39** No other page opens the dialog.

### R4 Pieces
- **FR-40** The preview adds:
  - `pieceCount`: the sum of all line quantities, which equals the sum of column B;
  - `dutiablePieceCount`: the same sum over the dutiable lines.
- **FR-41** Dialog KPIs:
  - "Lines / pieces" shows `lineCount / pieceCount`;
  - the dutiable KPI counts pieces.
- **FR-42** Sheet:
  - Row 5 ends with `Erstellt / generated: dd.mm.yyyy · {lines} Positionen / lines · {pieces} Stück / pieces`.
  - Column B of every subtotal and total row holds the quantity sum, in bold:
    - subtotals and the dutiable total: `=SUBTOTAL(9,B{first}:B{last})`, or `=0` when the section is empty;
    - duty-free total: `=SUM(B{sub1},…)`, using the same SUMIF fallback as column I when there are too many groups;
    - grand total: `=B{dutyFreeTotal}+B{dutiableTotal}`.
  - The stored values are the Go sums, rounded to 6 decimals (CALC-5).
- **NFR-1 additions:**
  - (h) the quantity sums of the groups, of both sections and of the whole sheet match their lines;
  - (i) each line member's unit value = its own unit value + its contents' values; every member with contents has quantity 1; every content's owner is a member of that line.

## 3. Exact copy

Sheet:
- `Positionen / lines`
- `Stück / pieces`
- `inkl. Inhalt ({n} Stk.)`
- `Inhalt von {label}`

UI, key prefix `components.reports.customs_inventory.`:

| key | en | de |
|---|---|---|
| `search_button` | Export inventory list … | Inventarliste exportieren … |
| `kpi_positions` | Lines / pieces | Positionen / Stück |
| `kpi_dutiable` | Dutiable ({count} pieces) | Zu verzollen ({count} Stück) |
| `kpi_contents` | + {count} items inside containers | + {count} Artikel in Behältern |
| `merge_first_word` | Match by first word only | Nur nach erstem Wort zusammenfassen |
| `merge_first_word_hint` | “Books & deco” and “books and deco” become one line “Books”. The unit value must still match. | „Bücher & Deko“ und „Bücher und Deko“ werden eine Zeile „Bücher“. Der Einzelwert muss gleich sein. |
| `roll_up` | Roll contents into their container | Inhalt in den Behälter einrechnen |
| `roll_up_hint` | A box is listed once, and its value includes everything inside it. Items with a different customs status keep their own line. | Eine Kiste erscheint einmal, ihr Wert enthält alles darin. Artikel mit anderem Zollstatus bleiben eigene Zeilen. |
| `issue_content_split` | Own line: customs status differs from its container | Eigene Zeile: Zollstatus weicht vom Behälter ab |
| `issue_container_quantity` | Contents listed separately: the container's quantity isn't 1 | Inhalt einzeln aufgeführt: Menge des Behälters ist nicht 1 |

`menu_entry` is removed.

## 4. Calls made while designing (the user can change them)

1. **Rollup is a dialog option, on by default.** The API default stays off, so old requests behave as before.
2. **Only quantity-1 containers absorb contents.** That keeps "unit value × quantity = line value" exact. A container with another quantity lists its contents separately (`container_quantity`). One entry with quantity 3 doesn't say which of the three boxes holds the contents, and every way of adding them to the line goes wrong. Example: "Moving box", quantity 3, 2.00 each, holding Books 3 × 10.00 = 30.00; the true total is 36.00.

   | approach | column H | column I | problem |
   |---|---|---|---|
   | add the contents to the price per piece | 32.00 | 96.00 | the books count three times |
   | add the contents to the line value only | 2.00 | 36.00 | B × H = 6.00 ≠ 36.00, so the formula and the check disagree |
   | spread the contents over the pieces | 12.00 | 36.00 | claims a third of the contents per box; uneven amounts lose cents (25.00 / 3 → 10.33 × 3 = 30.99) |
   | **chosen: no rollup** | 2.00 | 6.00 | none: the books get their own line, 3 × 10.00 = 30.00, comment `Inhalt von 000-005`; total 36.00 |

   To roll contents up, record each physical box that holds something as its own entry with quantity 1.
3. **Contents with a different customs status get their own line** (`content_split`), rather than being hidden in a container with another status. This protects the first spec's anti-goal: a dutiable item must never sit on the duty-free side. The walk stops at that container, so nothing below it skips past it into an outer box (FR-30).
4. **Contents inherit their containers' customs tags.** Giving a box the duty-free tag covers its undated contents. A purchase date after the cut-off still makes an item dutiable.
5. **A container is classified like any item.** A box without a purchase date and without a duty-free tag is dutiable (FR-9 step 6). If it holds duty-free contents, they split out of it, and the pre-export check says so.
6. **First-word merge keeps the unit value in the key.** Lines whose members all share one full name keep that name.
7. **The entry point is a button on the search page, not a separate "Export lists" page.** The search already holds the filters and the selection; a separate page would repeat the filter UI.
8. **The sheet's total rows also show quantity sums in column B,** so the checker can see "12 boxes" as well as the value.
9. **A status difference between an item and its container is a warning, not an error.** Both end up on their correct side of the sheet, each on its own line, so nothing is hidden. An error would have to block the export until the data changes, although an old scarf in a new bag is ordinary data. The pre-export check names every such item.

## 5. QA strategy

| layer | what | where |
|---|---|---|
| Pure unit | FR-24 table; FR-25/26 merge; FR-28–36 rollup (owner, value, split, quantity rule, receipts, inherited tags, exclusions); FR-40 sums; NFR-1 (h), (i) reject hand-broken reports | `customsinventory/*_test.go` |
| Excel check | column B formulas recalculate to the stored sums; row 5 text | `customsinventory/xlsx_test.go` |
| Service | real test database: contents loaded through parent links, counted toward the limit, off by default | `services/service_customs_inventory_test.go` |
| Frontend unit | saved options and request for both new flags; quantity formatting | `frontend/lib/customs-inventory.test.ts` |
| E2E | open the dialog from the search page button, with and without a selection; no export entry in the actions menu | `frontend/test/e2e/customs-inventory.browser.spec.ts` |
