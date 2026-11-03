package labelmaker

import "testing"

// TestSheetLayoutsFitPage guards against sheet presets that place labels
// beyond the page edge; such labels silently vanish from the rendered PDF.
func TestSheetLayoutsFitPage(t *testing.T) {
	const tolerance = 1.0 // mm, Avery templates round to 1/16"
	for _, p := range LabelPresets {
		l := p.SheetLayout
		if l == nil {
			continue
		}
		if l.Columns < 1 || l.Rows < 1 {
			t.Errorf("%s: layout needs at least one column and row", p.Key)
			continue
		}
		width := l.MarginLeft + float64(l.Columns)*p.Width + float64(l.Columns-1)*l.GutterH
		height := l.MarginTop + float64(l.Rows)*p.Height + float64(l.Rows-1)*l.GutterV
		if width > l.PageWidth+tolerance {
			t.Errorf("%s: %d columns need %.1f mm but the page is %.1f mm wide", p.Key, l.Columns, width, l.PageWidth)
		}
		if height > l.PageHeight+tolerance {
			t.Errorf("%s: %d rows need %.1f mm but the page is %.1f mm tall", p.Key, l.Rows, height, l.PageHeight)
		}
	}
}

// TestHerma5056IsCenteredOnA4 checks the A4 address label sheet: 2 × 6 labels
// of 97 × 42.3 mm with no gaps, centred, so equal margins on opposite sides.
func TestHerma5056IsCenteredOnA4(t *testing.T) {
	p := GetPresetByKey("herma_5056")
	if p == nil {
		t.Fatal("herma_5056 preset is missing")
	}
	l := p.SheetLayout
	if l == nil {
		t.Fatal("herma_5056 has no sheet layout")
	}
	if p.Brand != "Herma" || p.Width != 97 || p.Height != 42.3 {
		t.Errorf("got %s %.1f × %.1f mm, want Herma 97 × 42.3 mm", p.Brand, p.Width, p.Height)
	}
	if l.PageWidth != 210 || l.PageHeight != 297 || l.Columns != 2 || l.Rows != 6 || l.GutterH != 0 || l.GutterV != 0 {
		t.Errorf("layout %+v, want A4, 2 × 6, no gutters", *l)
	}
	right := l.PageWidth - l.MarginLeft - 2*p.Width
	bottom := l.PageHeight - l.MarginTop - 6*p.Height
	if d := right - l.MarginLeft; d > 0.05 || d < -0.05 {
		t.Errorf("left margin %.2f mm, right %.2f mm", l.MarginLeft, right)
	}
	if d := bottom - l.MarginTop; d > 0.05 || d < -0.05 {
		t.Errorf("top margin %.2f mm, bottom %.2f mm", l.MarginTop, bottom)
	}
}
