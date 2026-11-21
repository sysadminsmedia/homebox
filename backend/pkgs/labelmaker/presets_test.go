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
