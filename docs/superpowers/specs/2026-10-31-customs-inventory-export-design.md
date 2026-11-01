# Customs Inventory Export: Design

**Date:** 2026-10-31 · **Status:** agreed in conversation, ready for planning
**Mockups:** [`2026-10-31-customs-inventory-export-mockups.html`](2026-10-31-customs-inventory-export-mockups.html)
**Layout reference:** an inventory list template provided by the user. It is not committed, because it contains contact data.

## 1. Intent

Sometimes goods leave or enter a country, and someone asks for a list of them: a customs office, or a company that ships the goods. This feature creates that list from HomeBox. The list can be used for export and for the declaration of duties.

Customs rules often depend on how long the goods have been owned:

- Goods owned and used for long enough are often duty-free.
- Newer goods must be declared, with their value.

Many customs offices accept a list with this shape:

- Ordinary goods can be summarized, for example "12 boxes of books".
- Furniture, valuable items, larger devices and restricted goods are listed one by one.
- Goods that must be declared are listed at the end, in their own section, with their value.

HomeBox already has the data:

- locations (rooms, shelves, boxes);
- entity types (for example "Box", "Furniture", "Vehicle");
- tags (for example "Shipment A", or "owned > 2 years");
- the purchase price **per piece** and the purchase date;
- often an invoice as an attachment.

Today the only export is a raw CSV or JSON of the table, without attachments.

**This feature turns a HomeBox search into an inventory list as an Excel file. It can add the invoices as proof, in a ZIP file.** The layout follows one common template for such lists. It is not an official form. Each list usually covers one shipment, picked with the search filter.

**What success looks like:** the person who checks the goods can read the list and confirm that:

- every line matches the goods;
- the value of every line is quantity × price per piece;
- every number adds up;
- every item to declare is in its own section, with its value;
- every invoice can be found from its line.

A list that looks fine but has one wrong total, or one item listed twice, is a failure.

## 2. Goals, non-goals, anti-goals

**Goals**
- G1: From the item search, export the selected items, or all items that match the search, as one `.xlsx` inventory list. Optionally add the invoices in a ZIP file. Each line's value is **quantity × purchase price per piece**. HomeBox stores the price per piece.
- G2: Split goods into duty-free goods and «zu verzollende Waren» (goods to declare) automatically. The split uses purchase dates, tags, and a minimum ownership period (default 6 months).
- G3: Every number must agree everywhere:
  - line values, group subtotals and totals are the same in the dialog, in the file, and after Excel recalculates;
  - they still add up after someone edits a price in the file (for example, a later currency conversion).
- G4: Before the export, show every data problem that would make the list wrong or weak, with links to fix it in HomeBox.
- G5: Work like label printing: an entry in the item table's actions menu, then a dialog, then a download made by the server.

**Non-goals**
- Currency conversion. Values stay in the collection's currency. Conversion happens later, outside HomeBox.
- PDF output, a second sheet, or images in the sheet.
- Filling in official customs forms, or vehicle import papers.
- Other ways to value goods, such as market value or depreciation. The list's value is the replacement value, estimated as quantity × purchase price per piece.
- Keeping generated files on the server, or a history of exports.
- An entry on the Tools/Reports page. The flow starts from the search, because the selection lives there.

**Anti-goals** (failures, even if everything else works)
- An item is missing from the list, or counted twice, without a warning.
- An item that might need to be declared lands on the duty-free side without a warning.
- A total in the file differs from the dialog, or from Excel's own recalculation.
- A line value uses the price per piece without multiplying by the quantity.
- The flow refuses to export because data is incomplete. It warns, and the user decides.
- The export shows items or attachments from another collection, or a cell turns into a formula (injection).

## 3. Flow (see mockups)

1. **Select.** On `/search`, the user filters (for example by the tag "Shipment A") and selects rows. The header checkbox selects only the current page.
2. **Open.** The actions menu (`data-table-dropdown.vue`, both the multi-select and the single-item version) gets a new entry, **"Export as Customs Inventory List …"**. It sits right after "Print Labels for Selected Items" and opens `DialogID.CustomsInventoryExport`.
3. **Choose options** (left side of the dialog):
   - **Scope:** "Only selected items (n)" or "All items in the current search (m)". The second option only shows when the dialog was opened from `/search`, because only the search page knows the current query.
   - **Grouping:** location (default), entity type, or tag. For tag, the user picks tags in order.
   - **Merge identical items:** a checkbox, on by default.
   - **Customs rules:**
     - reference date (required; for example the day the goods cross the border; defaults to the last value used);
     - minimum ownership in months (default 6);
     - the cut-off date, shown read-only;
     - duty-free tags (several allowed);
     - optional "always dutiable" tags (several allowed).
   - **List header:** title, name, address (several lines), phone, email.
   - **Receipts:** an "Include invoices" checkbox, and which attachment types count as receipts (default: only `receipt`).
4. **Live check** (right side). Every option change sends a delayed (debounced) call to `POST /v1/reporting/customs-inventory/preview`. The answer has counts, totals and issues. Each issue lists the affected items, with links to `/item/{id}`.
5. **Export.** The button says "Export", or "Export anyway" when there are warnings. It calls `POST /v1/reporting/customs-inventory/export` and downloads the `.zip`, or just the `.xlsx` when receipts are off. The options are saved for next time.

## 4. Functional requirements

### Scope and inclusion
- **FR-1** The scope is either a list of item IDs, or a search query with the same fields as `GET /v1/entities`: `q`, `parentIds`, `entityTypeIds`, `tags`, `negateTags`, `matchAllTags`, `onlyWithPhoto`, `onlyWithoutPhoto`, `includeArchived`, `fields`. For a query, the server runs the search again, without pages.
- **FR-2** Every entity in scope ends up either on exactly one line, or in the *excluded* list with a reason:
  - `location`: the entity type is a location;
  - `archived`;
  - `sold`: a sold date is set;
  - `not_found`: the ID is not in this collection.
- **FR-3** The scope can have at most 5000 entities. Above that, the server answers 400 with a message.

### Grouping and lines
- **FR-4 Grouping modes:**
  - `location`: group by the item's location (`EntityOut.Location`). The group title is the full location path, joined with " › " (for example "Home › Living room"). Items without a location go under "Ohne Lagerort / no location".
  - `entityType`: group by the entity type name.
  - `tag`: the user gives an ordered list of tag IDs. An item goes into the first group whose tag it carries, directly or through a child tag. Items that match no group go under "Übrige / other".
- **FR-5** Group order:
  - `tag`: the order the user gave, then "Übrige / other".
  - otherwise: by title, A–Z, ignoring case, with "Ohne Lagerort / no location" last.

  Line order inside a group: name (ignoring case), then asset ID, then internal ID.
- **FR-6** Items to declare leave their group and go into a last section, «Zu verzollende Waren / goods subject to duty». This section has no groups, but each line shows its location. Its lines follow the group order, then the line order.
- **FR-7 Merge** (when on): items become one line when all of these match:
  - group;
  - entity type ID;
  - name, after trimming, collapsing inner spaces and ignoring case;
  - customs status;
  - price per piece, in cents.

  The merged line has:
  - quantity = the sum of the quantities;
  - asset IDs = all of them, ascending, separated by commas;
  - receipts = all of them;
  - purchase date = the date if all are the same; otherwise the text "YYYY-MM-DD – YYYY-MM-DD" (earliest to latest); empty if none has a date;
  - location = the different locations, joined by ", ";
  - comments = the different comments, joined by "; ".

  Its value is the summed quantity × the shared price per piece (CALC-2).
- **FR-8** Positions (`Pos.`) are numbered 1…n through the whole sheet, including the section of goods to declare.

### Customs classification (checked in this order)
- **FR-9** The cut-off is the reference date minus the minimum ownership period (CALC-4). For each item:
  1. It carries an "always dutiable" tag (or a child of one): **dutiable**, reason `tag_dutiable`. If it also carries a duty-free tag, add issue `conflict_tags`.
  2. Its purchase date is after the reference date: **dutiable**, reason `purchased_after_reference`, issue `purchased_after_reference`.
  3. Its purchase date is after the cut-off: **dutiable**, reason `owned_too_short`. If it carries a duty-free tag, add issue `conflict_date_tag`. The date wins.
  4. Its purchase date is on or before the cut-off: **duty-free**, reason `date`.
  5. It has no purchase date, but carries a duty-free tag (or a child of one): **duty-free**, reason `tag_duty_free`.
  6. It has no purchase date and no duty-free tag: **dutiable**, reason `missing_purchase_date`, issue `missing_purchase_date`.
- **FR-10** The comments column shows the reason for every dutiable line, and for duty-free lines that rely on a tag. The text is German (exact copy in §8):
  - `owned_too_short` → "Kauf nach Stichtag (< {n} Monate)"
  - `missing_purchase_date` → "Kaufdatum fehlt"
  - `tag_dutiable` → "zu verzollen lt. Tag „{tag}“"
  - `purchased_after_reference` → "Kauf nach Referenzdatum"
  - `tag_duty_free` → "zollfrei lt. Tag „{tag}“"
  - plus "; Wert fehlt" when the price is missing.

### Issues (the check before export)
- **FR-11** Issue codes. Each issue lists the affected items (ID, asset ID, name). All have severity `warning`, except the `excluded_*` codes, which are `info`.

  | code | when |
  |---|---|
  | `missing_purchase_date` | FR-9 step 6 |
  | `conflict_date_tag` | FR-9 step 3, with a duty-free tag |
  | `conflict_tags` | FR-9 step 1, with a duty-free tag |
  | `purchased_after_reference` | FR-9 step 2 |
  | `missing_price` | the price is 0, or rounds to 0 cents |
  | `negative_value` | price < 0, or quantity < 0; the line stays, with the stored value |
  | `zero_quantity` | quantity is 0 |
  | `missing_asset_id` | the asset ID is 0 |
  | `duplicate_asset_id` | the same non-zero asset ID is on more than one exported item |
  | `dutiable_without_receipt` | a dutiable item has no receipt attachment (of the selected types) |
  | `receipt_is_link` | a receipt attachment is an external link (`MimeTypeLinkURL`), so it can't go in the ZIP |
  | `excluded_location`, `excluded_archived`, `excluded_sold`, `excluded_not_found` | FR-2 |

- **FR-12** Issues never block the export. Only a failed self-check (NFR-1) does: HTTP 500, and no file.

### Output file
- **FR-13** One `.xlsx` workbook with one sheet, named `Inventarliste`. Layout:
  - **Rows 1–4:** title (bold); name; address (lines joined by ", "); "Telefon / phone: … · E-Mail / email: …". Empty parts are left out, but the rows stay.
  - **Row 5:** "Referenzdatum / reference date: dd.mm.yyyy" · "Stichtag / cut-off ({n} Monate / months): dd.mm.yyyy" · "Währung / currency: {ISO code, upper case}" · "Erstellt / generated: dd.mm.yyyy · {n} Artikel / items".
  - **Row 6:** empty. Then, for each duty-free group:
    - a group title row;
    - a column header row;
    - the lines;
    - a subtotal row;
    - an empty row.
  - **Then:**
    - "Total abgabenfrei / total duty-free";
    - an empty row;
    - the section of goods to declare (title row, column header row, lines, then "Total zu verzollen / total subject to duty");
    - an empty row;
    - "Gesamtwert / total value".

  The section of goods to declare is always there. With no lines, its total is 0.00. With no duty-free groups, the duty-free total is 0.00.
- **FR-14** Columns A–K, with the headers exactly as in §8:

  | col | content | cell type |
  |---|---|---|
  | A | Pos. | number |
  | B | Menge / quantity | number |
  | C | Bezeichnung / description | text |
  | D | Typ / type | text |
  | E | Lagerort / location | text |
  | F | Inventar-Nr. / asset ID | text, format `000-123` |
  | G | Kaufdatum / purchase date | date (format `dd.mm.yyyy`), or text for merged date ranges |
  | H | Einzelwert / unit value {CUR} | number (`#,##0.00`): the purchase price **per piece**; empty when the price is missing |
  | I | Wiederbeschaffungswert / replacement value {CUR} | formula `=ROUND(B{r}*H{r},2)`: quantity × price per piece |
  | J | Beleg / receipt | text, file names joined by ", " |
  | K | Bemerkungen / comments | text |

- **FR-15** Formulas:
  - line value: `=ROUND(B{r}*H{r},2)`;
  - group subtotal and total of goods to declare: `=SUBTOTAL(9,I{first}:I{last})`, or `=0` when that section has no lines;
  - duty-free total: `=SUM(I{sub1},I{sub2},…)`, or `=0` when there are no groups;
  - grand total: `=I{dutyFreeTotal}+I{dutiableTotal}`.

  Formulas use commas as separators (the file format doesn't depend on the locale). Every formula cell also stores its computed value (CALC rules). The workbook asks Excel to recalculate everything when it opens.
- **FR-16** Text cells are always written as text. A value that starts with `=`, `+`, `-`, `@`, tab or CR must never become a formula. Use `SetCellStr`, never `SetCellValue`, for user text.
- **FR-17** Formatting:
  - bold title, group title, header, subtotal and total rows;
  - column widths set;
  - no frozen panes (each group repeats its header row instead);
  - no colours, images or merged cells, except the merged title row A1:K1.

### Receipts and the ZIP
- **FR-18** Receipts are the item's attachments whose `type` is in the selected set (default `["receipt"]`; allowed: `receipt`, `attachment`, `warranty`, `manual`). `photo` and `thumbnail` are never included. External links are skipped, with issue `receipt_is_link`.
- **FR-19** Receipt file names:
  - pattern: `Belege/{assetID}_Rechnung{suffix}{ext}`;
  - assetID is `000-123`, or `id-{first 8 hex characters of the item UUID}` when the asset ID is 0;
  - suffix is empty for an item's first receipt, then `_2`, `_3` … in order of the attachment's `CreatedAt`, with the ID as a tie-break;
  - ext comes from the attachment title's extension if it has one; otherwise from the MIME type (`application/pdf`→`.pdf`, `image/jpeg`→`.jpg`, `image/png`→`.png`, `image/heic`→`.heic`); otherwise `.bin`.

  If several exported items share an asset ID, the second one (by internal ID) gets `_b` right after the asset ID, the third `_c`, and so on (`000-123_b_Rechnung.pdf`). So names never collide. Every name is ASCII and has no path separators.
- **FR-20** With receipts on, the answer is `application/zip`, named `Inventarliste_{slug}_{yyyy-mm-dd}.zip`. It holds `Inventarliste_{slug}_{yyyy-mm-dd}.xlsx` at the top and the `Belege/…` files. With receipts off, the answer is just the `.xlsx`, with the same base name.
  - The slug is the title in lower case, with ä→ae, ö→oe, ü→ue, ß→ss, other non-letters and non-digits collapsed to `-`, trimmed, and at most 60 characters. If nothing is left, it is `export`.
  - The date is the day of generation, in UTC.

### Saved options and UI
- **FR-21** The dialog remembers its last options in the synced view preferences (`useViewPreferences`, new key `customsInventory`). It saves: header fields, reference date, minimum months, grouping, group tag IDs, duty-free tag IDs, dutiable tag IDs, the merge flag, the receipts flag, and receipt types. When loading, it drops tag IDs it doesn't know (for example after switching collections).
- **FR-22** All UI text uses i18n keys in `frontend/locales/en.json` and `de.json`. The sheet's labels are fixed German/English constants in Go, and don't depend on the UI language.

## 5. Calculation rules

- **CALC-1 Cents.** A price `p float64` becomes `toCents(p) = round_half_away(round(p*100, 6 decimals))`, as an `int64`. The inner rounding removes tiny binary errors. Examples: 1.005 → 101 (like Excel's `ROUND(1.005,2)` = 1.01), 19.99 → 1999, 0.1+0.2 → 30, 2.675 → 268, −1.005 → −101.
- **CALC-2 Line value.** HomeBox stores the purchase price per piece. A line's value is `lineCents(qty, unitCents) = round_half_away(round(qty*unitCents, 6 decimals))`. Example: 4 × 89.50 → 35800 cents (358.00). A merged line uses the summed quantity, never a sum of rounded member values.
- **CALC-3 Sums.** Subtotals and totals are integer sums of `lineCents`. Excel shows `cents/100`. Duty-free total + dutiable total = grand total, exactly.
- **CALC-4 Cut-off.** `cutoff = subtractMonthsClamped(referenceDate, minMonths)`. If the day doesn't exist in the target month, use that month's last day. With 6 months: 2026-08-31 → 2026-02-28, 2028-08-31 → 2028-02-29, 2026-03-15 → 2025-09-15, 2026-12-31 → 2026-06-30. A purchase on the cut-off date is duty-free. Compare calendar dates only (`types.Date`, UTC midnight), never times. `minMonths` is a whole number from 1 to 60. If it's missing (0), use 6.
- **CALC-5 Quantities** are `float64` and written as they are. Fractions such as 2.5 (metres) are allowed. Summed quantities on merged lines are rounded to 6 decimals.
- **CALC-6 Missing price.** A price of 0 counts as missing: cell H is empty and the line value is 0.

**Note for reviewers:** the HomeBox dashboard's total value already uses price × quantity, so it also treats the price as per piece. Some HomeBox charts (value by location, by tag, and over time) add up prices without the quantity. Don't use those charts to check this list's totals.

## 6. Non-functional requirements

- **NFR-1 Self-check.** Before writing a file, the server checks these rules on the computed report. If one fails, it refuses with an error. The log has the group ID and counts, but no item data.
  - (a) every entity in scope is on exactly one line, or in exactly one excluded entry;
  - (b) for every line, the members' quantities add up to the line quantity, and recomputing `lineCents` gives the stored value;
  - (c) each group subtotal = the sum of its lines;
  - (d) duty-free total = the sum of the group subtotals;
  - (e) grand total = duty-free + dutiable;
  - (f) every receipt name used on a line is in the receipt list, and the other way round;
  - (g) receipt names are unique.
- **NFR-2 One calculation.** The preview and the export call the same `Build` function with the same inputs, so their totals and counts can't differ. The preview never writes files.
- **NFR-3 Same input, same output.** The same data and options give the same cell contents, except the "generated" date. Tests inject the clock.
- **NFR-4 Excel check.** A test opens the produced file again, recalculates every formula with excelize's calc engine, and checks that each result equals the stored value and the cents computed in Go (as a 2-decimal string). Manual check: the file opens without repair prompts in Excel, LibreOffice Calc and Apple Numbers.
- **NFR-5 Collection isolation.** Every entity, tag and attachment is loaded with the caller's group ID (`services.Context.GID`). IDs from another collection become `excluded_not_found`, never data.
- **NFR-6 Speed** (SQLite, developer machine), for 1000 items:
  - the preview answers in ≤ 3 s;
  - the export without receipts takes ≤ 5 s;
  - with receipts, the ZIP is streamed, so memory doesn't grow with attachment size.
- **NFR-7 Safety.**
  - text cells can't become formulas (FR-16);
  - ZIP entry names are generated, never taken from user input (FR-19);
  - the request size is capped by the existing body limit;
  - both endpoints need a login (the existing test that checks every non-public route rejects anonymous requests covers them).
- **NFR-8 Privacy.** Nothing is stored on the server. Header data lives only in the user's synced preferences. Logs and traces carry counts and IDs, never names, addresses or prices.
- **NFR-9 Lint and CI.** `task go:lint`, `task ui:fix` (≤ 1 warning in `lint:ci`), `task ui:check` and `task generate` must be clean. The swagger and TS types are regenerated, never edited by hand.

## 7. Technical shaping (decided)

**Backend**
- **New pure package** `backend/internal/core/services/reporting/customsinventory/`:
  - `model.go`: input and output types;
  - `money.go`: CALC-1, CALC-2, CALC-5;
  - `dates.go`: CALC-4;
  - `classify.go`: FR-9, FR-10;
  - `build.go`: grouping, merging, numbering, totals, issues, receipt names;
  - `verify.go`: NFR-1;
  - `xlsx.go`: FR-13 to FR-17;
  - `bundle.go`: the FR-20 ZIP writer.

  This package doesn't touch the database.
- **Dependency:** `github.com/xuri/excelize/v2 v2.11.0` (released 2026-07-06).
- **New service** `backend/internal/core/services/service_customs_inventory.go`, `CustomsInventoryService`:
  - finds the scope (`QueryByGroup` with `Page=-1, PageSize=-1`, or the given IDs);
  - loads each entity with `EntityRepository.GetOneByGroup`, which includes tags, entity type, attachments and the location;
  - gets location paths with `PathForEntity`, cached per location;
  - loads the collection's tags once with `TagRepository.GetAll(gid)`. The pure package finds child tags from `TagSummary.ParentID`. `TagRepository.GetDescendantTagIDs` is **not** used, because it isn't limited to one group. Tag IDs from outside the collection are dropped.
  - reads the currency from `GroupRepository.GroupByID`;
  - opens attachment files with `blob.OpenBucket(repos.Attachments.GetConnString())`, as `service_exports.go` does.

  It's wired in `services/all.go` as `AllServices.CustomsInventory`.
- **API types** (`Request`, `Preview`, `Issue`, …) live in the pure package, with JSON tags. Dates go over the API as `"YYYY-MM-DD"` strings, not `types.Date`, to keep the generated TS types simple. `backend/app/tools/typegen/main.go` gets one more replace rule (` Customsinventory` → ` CustomsInventory`), so the TS names read `CustomsInventoryRequest` and so on.
- **New handlers** in `backend/app/api/handlers/v1/v1_ctrl_customs_inventory.go`:
  - `HandleCustomsInventoryPreview` (adapter `adapters.Action`, JSON in and out);
  - `HandleCustomsInventoryExport` (a manual `errchain.HandlerFunc` with a binary answer, like `HandleLabelTemplatesRender`).

  Routes in `routes.go`, next to `/reporting/bill-of-materials`:
  - `POST /v1/reporting/customs-inventory/preview`
  - `POST /v1/reporting/customs-inventory/export`

**Frontend**
- **API:** `frontend/lib/api/classes/reports.ts` gets `customsInventoryPreview(body)` and `customsInventoryExport(body)`. The export follows the `renderBlob` pattern in `label-templates.ts`.
- **Dialog:** `frontend/components/Reports/CustomsInventoryDialog.vue`, registered as `DialogID.CustomsInventoryExport` with params `{ itemIds: string[] }`. Pure helpers (building the request, merging saved options) live in `frontend/lib/customs-inventory.ts`.
- **Search context:** dialogs are mounted in `layouts/default.vue`, outside the page, so the page can't pass data with `provide`/`inject`. Instead:
  - `frontend/composables/use-customs-inventory-search.ts` holds a `useState` value, `{ query, total } | null`;
  - `pages/search.vue` sets it after every successful search, and clears it when the page closes;
  - the menu entry reads it on click and passes it in the dialog params `{ itemIds, search? }`.

  This avoids passing props through `ItemViewSelectable` → table → columns → dropdown.
- **Menu entry:** in `data-table-dropdown.vue`, after the print-labels entry.

## 8. Exact copy

Sheet labels (Go constants):
- Section labels:
  - `Zwischensumme / subtotal`
  - `Total abgabenfrei / total duty-free`
  - `Zu verzollende Waren / goods subject to duty`
  - `Total zu verzollen / total subject to duty`
  - `Gesamtwert / total value`
  - `Ohne Lagerort / no location`
  - `Übrige / other`
- Column headers: as in FR-14, with `{CUR}` replaced by the upper-case ISO code.
- Rows 4 and 5: as in FR-13.

UI (en / de), key prefix `components.reports.customs_inventory.`:

| key | en | de |
|---|---|---|
| `menu_entry` | Export as Customs Inventory List … | Als Zoll-Inventarliste exportieren … |
| `title` | Export customs inventory list | Zoll-Inventarliste exportieren |
| `scope_selected` | Only selected items ({count}) | Nur ausgewählte Artikel ({count}) |
| `scope_query` | All items in the current search ({count}) | Alle Artikel der aktuellen Suche ({count}) |
| `group_by` | Grouping | Gliederung |
| `group_location` / `group_entity_type` / `group_tag` | Location / Entity type / Tag | Lagerort / Entitätstyp / Tag |
| `merge` | Merge identical items | Gleiche Positionen zusammenfassen |
| `reference_date` | Reference date | Referenzdatum |
| `reference_date_hint` | For example, the day the goods cross the border. | Zum Beispiel der Tag des Grenzübertritts. |
| `min_months` | Minimum ownership (months) | Mindestbesitzdauer (Monate) |
| `cutoff` | Cut-off (reference date minus {months} months) | Stichtag (Referenzdatum minus {months} Monate) |
| `duty_free_tags` | Duty-free tags (owned long enough) | Zollfrei-Tags (lange genug im Besitz) |
| `dutiable_tags` | Always dutiable (optional) | Immer verzollen (optional) |
| `header` | List header | Kopfbereich der Liste |
| `receipts` | Include invoices (download as ZIP) | Rechnungen beilegen (Download als ZIP) |
| `check` | Pre-export check | Prüfung vor dem Export |
| `export` / `export_anyway` | Export / Export anyway | Exportieren / Trotzdem exportieren |
| `issue_<code>` | one per FR-11 code | see the plan, Task 9 |

## 9. QA strategy

| layer | what | where |
|---|---|---|
| Pure unit | CALC-1 to CALC-5 tables; quantity × price per piece; FR-9 classification for every branch and boundary date; FR-7 merge keys; FR-19 name collisions; NFR-1 rejects hand-broken reports | `customsinventory/*_test.go` |
| Golden | a fixed set of about 15 items that covers every reason and issue, compared to an exact expected `Report` stored as JSON | `customsinventory/testdata/` |
| Excel check | NFR-4 recalculation test; FR-16 injection test; exact header and column labels | `customsinventory/xlsx_test.go` |
| Service | real test database (`services/main_test.go` setup): scope by IDs and by query; IDs from another group → `excluded_not_found`; receipts streamed into the ZIP with the right names | `services/service_customs_inventory_test.go` |
| Handler | 400 for a missing reference date, an empty scope, or > 5000 items; content types and file names; anonymous → 401 (the existing test picks up the new routes) | `handlers/v1/v1_ctrl_customs_inventory_test.go` |
| Frontend unit | request builder (scope switch, tag order); saved options drop unknown tags | `frontend/lib/customs-inventory.test.ts` |
| E2E | seed 3 items and 1 receipt, open the dialog from the search, export, check the download name, and check that the preview totals equal the totals in the downloaded `.xlsx` | `frontend/test/e2e/customs-inventory.browser.spec.ts` |

**Local data:** the local dev database is incomplete (for example, entity types are missing) and doesn't match the running dev instance. No test, golden file or acceptance check relies on it. Manual acceptance uses complete data.

## 10. Decisions log

**Decided by the user:**
- One list per shipment, picked with the search filter.
- Scope chosen in the dialog.
- Currency = the collection's currency, no conversion.
- Output is `.xlsx` with real numbers and formula sums.
- Invoices as separate files in a ZIP subfolder, named by asset ID.
- No second sheet.
- Grouping chosen in the dialog (location, entity type or tag).
- Duty-free tags chosen in the dialog; parent tags allowed.
- Merging identical items is a dialog option.
- The date beats the tag, and the conflict is shown before the export.
- No date and no duty-free tag = dutiable.
- Separate duty-free and dutiable totals.
- The server makes the file.
- HomeBox's price is per piece. A line's value (replacement value) is quantity × price per piece.

**Calls made during design** (the user can change them):
- The minimum ownership period is a dialog setting (1–60 months, default 6), so the feature isn't tied to one country's rule.
- Neutral names: "reference date" instead of a date tied to one situation; "Customs Inventory List" in the menu.
- German/English sheet labels, including "no location" and "other".
- Columns A–K. Column I is labelled "Wiederbeschaffungswert / replacement value", like the layout reference.
- Position numbers run through the whole sheet.
- Merged lines list all their asset IDs.
- Child items aren't rolled up into their container.
- Tag grouping uses the first matching tag in the user's order, with "other" for the rest.
- Archived, sold and location entities are left out and shown as info.
- Issues never block the export.
- At most 5000 items.
- Receipt types default to `receipt` only, with an option to add others (some invoices are filed as generic `attachment`).
- The header is saved in the synced preferences.
- The comments column holds the classification reasons, never item notes.
- A merged line's date is shown as a range.
- The generation date is in UTC.
