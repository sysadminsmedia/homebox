package labelmaker

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderPDFLeavesNilSlotsEmpty(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	require.NoError(t, err)

	template := &TemplateData{
		Name:   "asset tag",
		Width:  66.675,
		Height: 25.4,
		DPI:    150,
		CanvasData: map[string]interface{}{
			"objects": []interface{}{
				map[string]interface{}{"type": "Textbox", "left": 4.0, "top": 4.0, "width": 200.0, "height": 30.0, "text": "{{asset_id}}", "fontSize": 16.0},
			},
		},
	}

	// Two used slots, then two labels
	entities := []*EntityData{nil, nil, {AssetID: "000-001"}, {AssetID: "000-002"}}

	layout := &SheetLayout{PageWidth: 215.9, PageHeight: 279.4, Columns: 3, Rows: 10, MarginTop: 12.7, MarginLeft: 4.8, GutterH: 3.2}
	pdf, err := renderer.RenderToPDFWithSheetLayout(template, entities, DefaultPDFOptions(), layout)
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(pdf, []byte("%PDF")))

	pdf, err = renderer.RenderToPDF(template, entities, DefaultPDFOptions())
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(pdf, []byte("%PDF")))
}
