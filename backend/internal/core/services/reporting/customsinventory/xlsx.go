package customsinventory

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const SheetName = "Inventarliste"

// Sheet labels. They are fixed and don't depend on the UI language.
const (
	LabelSubtotal      = "Zwischensumme / subtotal"
	LabelDutyFreeTotal = "Total abgabenfrei / total duty-free"
	LabelDutiable      = "Zu verzollende Waren / goods subject to duty"
	LabelDutiableTotal = "Total zu verzollen / total subject to duty"
	LabelGrandTotal    = "Gesamtwert / total value"
	LabelNoLocation    = "Ohne Lagerort / no location"
	LabelOther         = "Übrige / other"
)

// ColumnHeaders returns the headers of columns A–K.
func ColumnHeaders(currency string) []string {
	return []string{
		"Pos.",
		"Menge / quantity",
		"Bezeichnung / description",
		"Typ / type",
		"Lagerort / location",
		"Inventar-Nr. / asset ID",
		"Kaufdatum / purchase date",
		"Einzelwert / unit value " + currency,
		"Wiederbeschaffungswert / replacement value " + currency,
		"Beleg / receipt",
		"Bemerkungen / comments",
	}
}

// maxFunctionArgs is the most arguments an Excel function takes.
const maxFunctionArgs = 255

var columnWidths = []float64{6, 10, 34, 14, 26, 18, 14, 14, 18, 34, 40}

// excelEpoch is day 0 of Excel's 1900 date system. Serials counted from it
// match Excel only from 1900-03-01 on (Excel counts a 29 Feb 1900), so
// earlier dates are written as text.
var (
	excelEpoch     = time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	firstSerialDay = time.Date(1900, 3, 1, 0, 0, 0, 0, time.UTC)
)

type sheetStyles struct {
	bold, boldQty, money, boldMoney, date int
}

// sheetWriter writes the sheet top to bottom. The stream writer keeps each
// formula's computed value next to the formula, as a number, and stores
// strings as inline text, so user text can never become a formula.
type sheetWriter struct {
	sw  *excelize.StreamWriter
	st  sheetStyles
	row int
	err error
}

func (w *sheetWriter) put(values ...any) int {
	r := w.row
	w.row++
	if w.err == nil && len(values) > 0 {
		w.err = w.sw.SetRow(fmt.Sprintf("A%d", r), values)
	}
	return r
}

func text(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (w *sheetWriter) money(cents int64, formula string, bold bool) excelize.Cell {
	style := w.st.money
	if bold {
		style = w.st.boldMoney
	}
	return excelize.Cell{StyleID: style, Formula: formula, Value: float64(cents) / 100}
}

// line writes one line and returns its row.
func (w *sheetWriter) line(l Line) int {
	r := w.row
	var date any
	switch {
	case l.PurchaseDateText != "":
		date = l.PurchaseDateText
	case l.PurchaseDate.Time().IsZero():
	case l.PurchaseDate.Time().Before(firstSerialDay):
		date = l.PurchaseDate.String()
	default:
		days := l.PurchaseDate.Time().UTC().Sub(excelEpoch).Hours() / 24
		date = excelize.Cell{StyleID: w.st.date, Value: days}
	}
	var unit any
	if l.HasPrice {
		unit = excelize.Cell{StyleID: w.st.money, Value: float64(l.UnitCents) / 100}
	}
	return w.put(
		l.Pos,
		l.Quantity,
		text(l.Name),
		text(l.EntityTypeName),
		text(strings.Join(l.Locations, ", ")),
		text(strings.Join(l.AssetIDs, ", ")),
		date,
		unit,
		w.money(l.LineCents, fmt.Sprintf("ROUND(B%d*H%d,2)", r, r), false),
		text(strings.Join(l.ReceiptNames, ", ")),
		text(strings.Join(l.Comments, "; ")),
	)
}

func (w *sheetWriter) bold(s string) excelize.Cell {
	return excelize.Cell{StyleID: w.st.bold, Value: s}
}

// total writes a bold label in A, a bold quantity sum in B and a bold value in I.
func (w *sheetWriter) total(label string, qty float64, qtyFormula string, cents int64, formula string) int {
	return w.put(
		w.bold(label),
		excelize.Cell{StyleID: w.st.boldQty, Formula: qtyFormula, Value: RoundQty(qty)},
		nil, nil, nil, nil, nil, nil,
		w.money(cents, formula, true),
	)
}

func (w *sheetWriter) headers(currency string) {
	hs := ColumnHeaders(currency)
	cells := make([]any, len(hs))
	for i, h := range hs {
		cells[i] = w.bold(h)
	}
	w.put(cells...)
}

// section writes the lines and their subtotal row, and returns the subtotal row.
func (w *sheetWriter) section(lines []Line, label string, qty float64, cents int64) int {
	first := w.row
	for _, l := range lines {
		w.line(l)
	}
	qtyFormula, formula := "0", "0"
	if len(lines) > 0 {
		qtyFormula = fmt.Sprintf("SUBTOTAL(9,B%d:B%d)", first, w.row-1)
		formula = fmt.Sprintf("SUBTOTAL(9,I%d:I%d)", first, w.row-1)
	}
	return w.total(label, qty, qtyFormula, cents, formula)
}

func newStyles(f *excelize.File) (sheetStyles, error) {
	var st sheetStyles
	moneyFmt := 4 // #,##0.00
	dateFmt := "dd.mm.yyyy"
	for _, s := range []struct {
		dst   *int
		style excelize.Style
	}{
		{&st.bold, excelize.Style{Font: &excelize.Font{Bold: true}}},
		{&st.boldQty, excelize.Style{Font: &excelize.Font{Bold: true}}}, // line quantities have no number format either
		{&st.money, excelize.Style{NumFmt: moneyFmt}},
		{&st.boldMoney, excelize.Style{NumFmt: moneyFmt, Font: &excelize.Font{Bold: true}}},
		{&st.date, excelize.Style{CustomNumFmt: &dateFmt}},
	} {
		id, err := f.NewStyle(&s.style)
		if err != nil {
			return st, err
		}
		*s.dst = id
	}
	return st, nil
}

func headerContact(h Header) string {
	var parts []string
	if h.Phone != "" {
		parts = append(parts, "Telefon / phone: "+h.Phone)
	}
	if h.Email != "" {
		parts = append(parts, "E-Mail / email: "+h.Email)
	}
	return strings.Join(parts, " · ")
}

func headerAddress(address string) string {
	var lines []string
	for _, l := range strings.Split(address, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return strings.Join(lines, ", ")
}

func lineCount(r Report) int {
	n := len(r.Dutiable)
	for _, g := range r.Groups {
		n += len(g.Lines)
	}
	return n
}

// WriteXLSX writes the report as a workbook with one sheet (FR-13 to FR-17).
func WriteXLSX(r Report) ([]byte, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	if err := f.SetSheetName("Sheet1", SheetName); err != nil {
		return nil, err
	}
	st, err := newStyles(f)
	if err != nil {
		return nil, err
	}
	sw, err := f.NewStreamWriter(SheetName)
	if err != nil {
		return nil, err
	}
	for i, width := range columnWidths {
		if err := sw.SetColWidth(i+1, i+1, width); err != nil {
			return nil, err
		}
	}

	w := &sheetWriter{sw: sw, st: st, row: 1}
	w.put(excelize.Cell{StyleID: st.bold, Value: r.Header.Title})
	w.put(text(r.Header.Name))
	w.put(text(headerAddress(r.Header.Address)))
	w.put(text(headerContact(r.Header)))
	w.put(strings.Join([]string{
		"Referenzdatum / reference date: " + FormatDE(r.ReferenceDate),
		fmt.Sprintf("Stichtag / cut-off (%d Monate / months): %s", r.MinMonths, FormatDE(r.Cutoff)),
		"Währung / currency: " + r.Currency,
		fmt.Sprintf("Erstellt / generated: %s · %d Positionen / lines · %s Stück / pieces",
			r.GeneratedAt.UTC().Format("02.01.2006"), lineCount(r), FormatQtyDE(r.TotalQty)),
	}, " · "))
	w.put()

	firstGroupRow := w.row
	subtotalRows := make([]int, 0, len(r.Groups))
	for _, g := range r.Groups {
		w.put(w.bold(g.Title))
		w.headers(r.Currency)
		subtotalRows = append(subtotalRows, w.section(g.Lines, LabelSubtotal, g.SubtotalQty, g.SubtotalCents))
		w.put()
	}
	// sumSubtotals adds up one column of the subtotal rows.
	sumSubtotals := func(col string) string {
		switch {
		case len(subtotalRows) == 0:
			return "0"
		case len(subtotalRows) <= maxFunctionArgs:
			refs := make([]string, len(subtotalRows))
			for i, row := range subtotalRows {
				refs[i] = fmt.Sprintf("%s%d", col, row)
			}
			return "SUM(" + strings.Join(refs, ",") + ")"
		default:
			// More groups than a function takes arguments: add up the subtotal rows by their label.
			return fmt.Sprintf(`SUMIF(A%d:A%d,"%s",%s%d:%s%d)`, firstGroupRow, w.row-1, LabelSubtotal, col, firstGroupRow, col, w.row-1)
		}
	}
	dutyFreeRow := w.total(LabelDutyFreeTotal, r.DutyFreeQty, sumSubtotals("B"), r.DutyFreeTotalCents, sumSubtotals("I"))
	w.put()

	w.put(w.bold(LabelDutiable))
	w.headers(r.Currency)
	dutiableRow := w.section(r.Dutiable, LabelDutiableTotal, r.DutiableQty, r.DutiableTotalCents)
	w.put()
	w.total(LabelGrandTotal, r.TotalQty, fmt.Sprintf("B%d+B%d", dutyFreeRow, dutiableRow),
		r.GrandTotalCents, fmt.Sprintf("I%d+I%d", dutyFreeRow, dutiableRow))

	if w.err != nil {
		return nil, w.err
	}
	if err := sw.MergeCell("A1", "K1"); err != nil {
		return nil, err
	}
	if err := sw.Flush(); err != nil {
		return nil, err
	}
	fullCalc := true
	if err := f.SetCalcProps(&excelize.CalcPropsOptions{FullCalcOnLoad: &fullCalc}); err != nil {
		return nil, err
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
