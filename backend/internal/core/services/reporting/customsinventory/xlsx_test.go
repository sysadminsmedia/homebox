package customsinventory

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// testReport is a hand-built report, independent of Build.
func testReport() Report {
	return Report{
		Header: Header{
			Title:   "Möbel – Übersicht",
			Name:    `=HYPERLINK("http://x")`,
			Address: "1 Example Street\n12345 Sample Town",
			Phone:   "+1 555 0100",
			Email:   "jane@example.com",
		},
		Currency:      "EUR",
		ReferenceDate: d("2026-11-15"),
		Cutoff:        d("2026-05-15"),
		MinMonths:     6,
		GeneratedAt:   time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC),
		ItemCount:     20,
		Groups: []Group{
			{
				Title: "Home › Living room",
				Lines: []Line{
					{
						Pos: 1, Quantity: 4, Name: "Dining chair", EntityTypeName: typeNameFurniture,
						Locations: []string{"Home › Living room"}, AssetIDs: []string{"000-103"},
						PurchaseDate: d("2020-05-01"), UnitCents: 8950, HasPrice: true, LineCents: 35800,
						Status: StatusDutyFree, Reason: ReasonDate,
					},
					{
						Pos: 2, Quantity: 14, Name: nameFiles, EntityTypeName: "Box",
						Locations: []string{"Home › Living room"}, AssetIDs: []string{"000-131", "000-132"},
						PurchaseDateText: "2019-01-01 – 2021-06-30",
						Comments:         []string{"zollfrei lt. Tag „owned > 2 years“; Wert fehlt"},
						Status:           StatusDutyFree, Reason: ReasonTagDutyFree,
					},
				},
				SubtotalCents: 35800,
				SubtotalQty:   18,
			},
			{
				Title: "Home › Office",
				Lines: []Line{
					{
						Pos: 3, Quantity: 2.5, Name: "Shelf", EntityTypeName: typeNameFurniture,
						Locations: []string{"-"}, AssetIDs: []string{"000-110"},
						PurchaseDate: d("2020-01-01"), UnitCents: 33, HasPrice: true, LineCents: 83,
						Status: StatusDutyFree, Reason: ReasonDate,
					},
				},
				SubtotalCents: 83,
				SubtotalQty:   2.5,
			},
		},
		Dutiable: []Line{
			{
				Pos: 4, Quantity: 1, Name: nameEspresso, EntityTypeName: "Appliance",
				Locations: []string{"Home › Kitchen"}, AssetIDs: []string{"000-146"},
				PurchaseDate: d("2026-09-02"), UnitCents: 64990, HasPrice: true, LineCents: 64990,
				ReceiptNames: []string{"Belege/000-146_Rechnung.pdf", "Belege/000-146_Rechnung_2.jpg"},
				Comments:     []string{"Kauf nach Stichtag (< 6 Monate)"},
				Status:       StatusDutiable, Reason: ReasonOwnedTooShort,
			},
			{
				Pos: 5, Quantity: 1, Name: "Office chair", EntityTypeName: typeNameFurniture,
				Locations: []string{"Home › Office"}, AssetIDs: []string{"000-150"},
				Comments: []string{"Kaufdatum fehlt; Wert fehlt"},
				Status:   StatusDutiable, Reason: ReasonMissingPurchaseDate,
			},
		},
		DutyFreeTotalCents: 35883,
		DutiableTotalCents: 64990,
		GrandTotalCents:    100873,
		DutyFreeQty:        20.5,
		DutiableQty:        2,
		TotalQty:           22.5,
	}
}

func openXLSX(t *testing.T, r Report) *excelize.File {
	t.Helper()
	b, err := WriteXLSX(r)
	require.NoError(t, err)
	f, err := excelize.OpenReader(bytes.NewReader(b))
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func cell(t *testing.T, f *excelize.File, ref string) string {
	t.Helper()
	v, err := f.GetCellValue(SheetName, ref)
	require.NoError(t, err)
	return v
}

func rawCell(t *testing.T, f *excelize.File, ref string) string {
	t.Helper()
	v, err := f.GetCellValue(SheetName, ref, excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	return v
}

func formula(t *testing.T, f *excelize.File, ref string) string {
	t.Helper()
	v, err := f.GetCellFormula(SheetName, ref)
	require.NoError(t, err)
	return v
}

// findRow returns the 1-based row whose column col equals value, or 0.
func findRow(t *testing.T, f *excelize.File, col, value string) int {
	t.Helper()
	rows, err := f.GetRows(SheetName)
	require.NoError(t, err)
	idx := int(col[0] - 'A')
	for i, row := range rows {
		if idx < len(row) && row[idx] == value {
			return i + 1
		}
	}
	return 0
}

func TestWriteXLSX_LayoutAndLabels(t *testing.T) {
	f := openXLSX(t, testReport())

	assert.Equal(t, "Möbel – Übersicht", cell(t, f, "A1"))
	assert.Equal(t, `=HYPERLINK("http://x")`, cell(t, f, "A2"))
	assert.Equal(t, "1 Example Street, 12345 Sample Town", cell(t, f, "A3"))
	assert.Equal(t, "Telefon / phone: +1 555 0100 · E-Mail / email: jane@example.com", cell(t, f, "A4"))
	a5 := cell(t, f, "A5")
	assert.Contains(t, a5, "Referenzdatum / reference date: 15.11.2026")
	assert.Contains(t, a5, "Stichtag / cut-off (6 Monate / months): 15.05.2026")
	assert.Contains(t, a5, "Währung / currency: EUR")
	assert.Contains(t, a5, "Erstellt / generated: 31.10.2026 · 5 Positionen / lines · 22,5 Stück / pieces")
	assert.NotContains(t, a5, "Artikel / items")
	assert.Empty(t, cell(t, f, "A6"))
	assert.Equal(t, "Home › Living room", cell(t, f, "A7"))

	rows, err := f.GetRows(SheetName)
	require.NoError(t, err)
	assert.Equal(t, ColumnHeaders("EUR"), rows[7])
	assert.Equal(t, "Einzelwert / unit value EUR", ColumnHeaders("EUR")[7])
	assert.Equal(t, "Wiederbeschaffungswert / replacement value EUR", ColumnHeaders("EUR")[8])

	assert.NotZero(t, findRow(t, f, "A", LabelSubtotal))
	assert.NotZero(t, findRow(t, f, "A", LabelDutyFreeTotal))
	assert.NotZero(t, findRow(t, f, "A", LabelDutiable))
	assert.NotZero(t, findRow(t, f, "A", LabelDutiableTotal))
	assert.Equal(t, LabelGrandTotal, rows[len(rows)-1][0])

	merged, err := f.GetMergeCells(SheetName)
	require.NoError(t, err)
	require.Len(t, merged, 1)
	assert.Equal(t, "A1", merged[0].GetStartAxis())
	assert.Equal(t, "K1", merged[0].GetEndAxis())
}

func TestWriteXLSX_HeaderLeavesOutEmptyParts(t *testing.T) {
	r := testReport()
	r.Header = Header{Title: "List", Email: "jane@example.com"}
	f := openXLSX(t, r)
	assert.Empty(t, cell(t, f, "A2"))
	assert.Empty(t, cell(t, f, "A3"))
	assert.Equal(t, "E-Mail / email: jane@example.com", cell(t, f, "A4"))
	assert.Contains(t, cell(t, f, "A5"), "Referenzdatum")
	assert.Equal(t, "Home › Living room", cell(t, f, "A7"), "header rows stay even when empty")
}

func TestWriteXLSX_LineValueFormula(t *testing.T) {
	f := openXLSX(t, testReport())
	r := findRow(t, f, "C", "Dining chair")
	require.NotZero(t, r)

	assert.Equal(t, "4", rawCell(t, f, fmt.Sprintf("B%d", r)))
	assert.Equal(t, "89.5", rawCell(t, f, fmt.Sprintf("H%d", r)))
	assert.Equal(t, fmt.Sprintf("ROUND(B%d*H%d,2)", r, r), formula(t, f, fmt.Sprintf("I%d", r)))
	calc, err := f.CalcCellValue(SheetName, fmt.Sprintf("I%d", r), excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	assert.Equal(t, "358", calc)
}

func TestWriteXLSX_FormulasRecalculateToCachedValues(t *testing.T) {
	rep := testReport()
	f := openXLSX(t, rep)

	// Expected cents of every formula cell in column I, in row order.
	var want []int64
	for _, g := range rep.Groups {
		for _, l := range g.Lines {
			want = append(want, l.LineCents)
		}
		want = append(want, g.SubtotalCents)
	}
	want = append(want, rep.DutyFreeTotalCents)
	for _, l := range rep.Dutiable {
		want = append(want, l.LineCents)
	}
	want = append(want, rep.DutiableTotalCents, rep.GrandTotalCents)

	rows, err := f.GetRows(SheetName)
	require.NoError(t, err)
	var got []int64
	var formulas []string
	for i := range rows {
		ref := fmt.Sprintf("I%d", i+1)
		fx := formula(t, f, ref)
		if fx == "" {
			continue
		}
		formulas = append(formulas, fx)
		cached, err := strconv.ParseFloat(rawCell(t, f, ref), 64)
		require.NoError(t, err, "cached value of %s", ref)
		calcStr, err := f.CalcCellValue(SheetName, ref, excelize.Options{RawCellValue: true})
		require.NoError(t, err)
		calc, err := strconv.ParseFloat(calcStr, 64)
		require.NoError(t, err, "calc value of %s", ref)
		assert.Equal(t, ToCents(cached), ToCents(calc), "cached vs recalculated %s (%s)", ref, fx)
		got = append(got, ToCents(cached))
	}
	assert.Equal(t, want, got)

	lineRow := findRow(t, f, "C", "Shelf")
	subRows := []int{}
	for i, row := range rows {
		if len(row) > 0 && row[0] == LabelSubtotal {
			subRows = append(subRows, i+1)
		}
	}
	require.Len(t, subRows, 2)
	assert.Equal(t, fmt.Sprintf("SUBTOTAL(9,I%d:I%d)", lineRow, lineRow), formula(t, f, fmt.Sprintf("I%d", subRows[1])))
	dutyFreeRow := findRow(t, f, "A", LabelDutyFreeTotal)
	assert.Equal(t, fmt.Sprintf("SUM(I%d,I%d)", subRows[0], subRows[1]), formula(t, f, fmt.Sprintf("I%d", dutyFreeRow)))
	dutiableRow := findRow(t, f, "A", LabelDutiableTotal)
	grandRow := findRow(t, f, "A", LabelGrandTotal)
	assert.Equal(t, fmt.Sprintf("I%d+I%d", dutyFreeRow, dutiableRow), formula(t, f, fmt.Sprintf("I%d", grandRow)))
	assert.Len(t, formulas, len(want))

	// Every formula cell in column B recalculates to its stored quantity.
	nB := 0
	for i := range rows {
		ref := fmt.Sprintf("B%d", i+1)
		fx := formula(t, f, ref)
		if fx == "" {
			continue
		}
		nB++
		cached, err := strconv.ParseFloat(rawCell(t, f, ref), 64)
		require.NoError(t, err, "cached value of %s", ref)
		calcStr, err := f.CalcCellValue(SheetName, ref, excelize.Options{RawCellValue: true})
		require.NoError(t, err)
		calc, err := strconv.ParseFloat(calcStr, 64)
		require.NoError(t, err, "calc value of %s", ref)
		assert.InDelta(t, RoundQty(cached), RoundQty(calc), 0, "cached vs recalculated %s (%s)", ref, fx)
	}
	assert.Equal(t, 5, nB, "two subtotals, duty-free total, dutiable total, grand total")
}

func TestWriteXLSX_QuantityTotals(t *testing.T) {
	rep := testReport()
	f := openXLSX(t, rep)

	chair := findRow(t, f, "C", "Dining chair")
	var subRows []int
	rows, err := f.GetRows(SheetName)
	require.NoError(t, err)
	for i, row := range rows {
		if len(row) > 0 && row[0] == LabelSubtotal {
			subRows = append(subRows, i+1)
		}
	}
	require.Len(t, subRows, 2)
	first := fmt.Sprintf("B%d", subRows[0])
	assert.Equal(t, fmt.Sprintf("SUBTOTAL(9,B%d:B%d)", chair, subRows[0]-1), formula(t, f, first))
	assert.Equal(t, strconv.FormatFloat(rep.Groups[0].SubtotalQty, 'f', -1, 64), rawCell(t, f, first))

	dutyFreeRow := findRow(t, f, "A", LabelDutyFreeTotal)
	dutiableRow := findRow(t, f, "A", LabelDutiableTotal)
	grandRow := findRow(t, f, "A", LabelGrandTotal)
	assert.Equal(t, fmt.Sprintf("SUM(B%d,B%d)", subRows[0], subRows[1]), formula(t, f, fmt.Sprintf("B%d", dutyFreeRow)))
	assert.Equal(t, fmt.Sprintf("B%d+B%d", dutyFreeRow, dutiableRow), formula(t, f, fmt.Sprintf("B%d", grandRow)))
	assert.Equal(t, "22.5", rawCell(t, f, fmt.Sprintf("B%d", grandRow)))

	style, err := f.GetCellStyle(SheetName, fmt.Sprintf("B%d", grandRow))
	require.NoError(t, err)
	st, err := f.GetStyle(style)
	require.NoError(t, err)
	require.NotNil(t, st.Font)
	assert.True(t, st.Font.Bold, "quantity totals are bold")

	rep.Dutiable = nil
	rep.DutiableTotalCents, rep.DutiableQty = 0, 0
	rep.GrandTotalCents, rep.TotalQty = rep.DutyFreeTotalCents, rep.DutyFreeQty
	f = openXLSX(t, rep)
	ref := fmt.Sprintf("B%d", findRow(t, f, "A", LabelDutiableTotal))
	assert.Equal(t, "0", formula(t, f, ref))
	assert.Equal(t, "0", rawCell(t, f, ref))
}

func TestWriteXLSX_TextNeverFormula(t *testing.T) {
	f := openXLSX(t, testReport())
	shelf := findRow(t, f, "C", "Shelf")
	for ref, want := range map[string]string{
		"A2":                      `=HYPERLINK("http://x")`,
		fmt.Sprintf("E%d", shelf): "-",
	} {
		assert.Empty(t, formula(t, f, ref), ref)
		assert.Equal(t, want, cell(t, f, ref), ref)
		typ, err := f.GetCellType(SheetName, ref)
		require.NoError(t, err)
		assert.Equal(t, excelize.CellTypeInlineString, typ, ref)
	}
	assert.True(t, strings.HasPrefix(cell(t, f, "A4"), "Telefon / phone: +1 555 0100"))
}

func TestWriteXLSX_CellTypes(t *testing.T) {
	f := openXLSX(t, testReport())
	chair := findRow(t, f, "C", "Dining chair")
	files := findRow(t, f, "C", nameFiles)

	for _, ref := range []string{fmt.Sprintf("B%d", chair), fmt.Sprintf("H%d", chair), fmt.Sprintf("G%d", chair)} {
		_, err := strconv.ParseFloat(rawCell(t, f, ref), 64)
		require.NoError(t, err, "%s is numeric", ref)
		typ, err := f.GetCellType(SheetName, ref)
		require.NoError(t, err)
		assert.NotEqual(t, excelize.CellTypeInlineString, typ, ref)
	}
	assert.Equal(t, "01.05.2020", cell(t, f, fmt.Sprintf("G%d", chair)))
	assert.Equal(t, "2019-01-01 – 2021-06-30", cell(t, f, fmt.Sprintf("G%d", files)))
	assert.Empty(t, cell(t, f, fmt.Sprintf("H%d", files)), "no price → empty unit value")
	assert.Equal(t, "000-131, 000-132", cell(t, f, fmt.Sprintf("F%d", files)))

	espresso := findRow(t, f, "C", nameEspresso)
	assert.Equal(t, "Belege/000-146_Rechnung.pdf, Belege/000-146_Rechnung_2.jpg", cell(t, f, fmt.Sprintf("J%d", espresso)))
	assert.Equal(t, "Kauf nach Stichtag (< 6 Monate)", cell(t, f, fmt.Sprintf("K%d", espresso)))
}

func TestWriteXLSX_EmptyDutiableSection(t *testing.T) {
	r := testReport()
	r.Dutiable = nil
	r.DutiableTotalCents = 0
	r.GrandTotalCents = r.DutyFreeTotalCents
	f := openXLSX(t, r)

	dutiableRow := findRow(t, f, "A", LabelDutiableTotal)
	require.NotZero(t, dutiableRow)
	ref := fmt.Sprintf("I%d", dutiableRow)
	assert.Equal(t, "0", formula(t, f, ref))
	assert.Equal(t, "0", rawCell(t, f, ref))

	grand := fmt.Sprintf("I%d", findRow(t, f, "A", LabelGrandTotal))
	calc, err := f.CalcCellValue(SheetName, grand, excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	v, err := strconv.ParseFloat(calc, 64)
	require.NoError(t, err)
	assert.Equal(t, r.DutyFreeTotalCents, ToCents(v))
}

func TestWriteXLSX_NoGroups(t *testing.T) {
	r := testReport()
	r.Groups = nil
	r.DutyFreeTotalCents = 0
	r.GrandTotalCents = r.DutiableTotalCents
	f := openXLSX(t, r)

	ref := fmt.Sprintf("I%d", findRow(t, f, "A", LabelDutyFreeTotal))
	assert.Equal(t, "0", formula(t, f, ref))
	assert.Equal(t, "0", rawCell(t, f, ref))
}

func TestWriteXLSX_FullCalcOnLoad(t *testing.T) {
	f := openXLSX(t, testReport())
	props, err := f.GetCalcProps()
	require.NoError(t, err)
	require.NotNil(t, props.FullCalcOnLoad)
	assert.True(t, *props.FullCalcOnLoad)
}

func TestWriteXLSX_Deterministic(t *testing.T) {
	a := openXLSX(t, testReport())
	b := openXLSX(t, testReport())
	rowsA, err := a.GetRows(SheetName)
	require.NoError(t, err)
	rowsB, err := b.GetRows(SheetName)
	require.NoError(t, err)
	assert.Equal(t, rowsA, rowsB)
}

func TestWriteXLSX_DatesBefore1900AreText(t *testing.T) {
	r := testReport()
	r.Groups[1].Lines[0].PurchaseDate = d("0202-09-30")
	r.Groups[0].Lines[0].PurchaseDate = d("1900-02-28")
	f := openXLSX(t, r)

	for name, want := range map[string]string{"Shelf": "0202-09-30", "Dining chair": "1900-02-28"} {
		ref := fmt.Sprintf("G%d", findRow(t, f, "C", name))
		assert.Equal(t, want, cell(t, f, ref), name)
		typ, err := f.GetCellType(SheetName, ref)
		require.NoError(t, err)
		assert.Equal(t, excelize.CellTypeInlineString, typ, "%s: Excel's date serials start in 1900", name)
	}
}

// Excel allows at most 255 arguments per function, so the duty-free total
// can't list one subtotal per group when there are more groups than that.
func TestWriteXLSX_ManyGroupsStayWithinExcelLimits(t *testing.T) {
	r := testReport()
	r.Groups = nil
	r.DutyFreeTotalCents = 0
	for i := range 300 {
		r.Groups = append(r.Groups, Group{
			Title:         fmt.Sprintf("Box %03d", i),
			Lines:         []Line{{Pos: i + 1, Quantity: 1, Name: "Books", UnitCents: 100, HasPrice: true, LineCents: 100}},
			SubtotalCents: 100,
			SubtotalQty:   1,
		})
		r.DutyFreeTotalCents += 100
	}
	r.DutyFreeQty = 300
	r.GrandTotalCents = r.DutyFreeTotalCents + r.DutiableTotalCents
	f := openXLSX(t, r)

	ref := fmt.Sprintf("I%d", findRow(t, f, "A", LabelDutyFreeTotal))
	fx := formula(t, f, ref)
	assert.LessOrEqual(t, strings.Count(fx, ",")+1, 255, "argument count of %.40s…", fx)
	assert.Less(t, len(fx), 8192)
	calc, err := f.CalcCellValue(SheetName, ref, excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	v, err := strconv.ParseFloat(calc, 64)
	require.NoError(t, err)
	assert.Equal(t, int64(30000), ToCents(v))
	assert.Equal(t, "300", rawCell(t, f, ref))

	qref := fmt.Sprintf("B%d", findRow(t, f, "A", LabelDutyFreeTotal))
	qfx := formula(t, f, qref)
	assert.LessOrEqual(t, strings.Count(qfx, ",")+1, 255, "argument count of %.40s…", qfx)
	qcalc, err := f.CalcCellValue(SheetName, qref, excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	assert.Equal(t, "300", qcalc)
}
