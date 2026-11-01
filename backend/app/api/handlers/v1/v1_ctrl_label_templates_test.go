package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/labelmaker"
)

func TestSheetLayoutFor(t *testing.T) {
	presetKey := "avery_5160"
	preset := labelmaker.GetPresetByKey(presetKey)
	require.NotNil(t, preset)
	require.NotNil(t, preset.SheetLayout)
	presetMarginLeft := preset.SheetLayout.MarginLeft

	t.Run("preset layout with offset", func(t *testing.T) {
		template := repo.LabelTemplateOut{Preset: &presetKey, Width: 10, Height: 10, OffsetX: 1.5, OffsetY: -0.5}
		data := &labelmaker.TemplateData{Width: 10, Height: 10}

		layout := sheetLayoutFor(template, data)
		require.NotNil(t, layout)
		assert.InDelta(t, presetMarginLeft+1.5, layout.MarginLeft, 1e-9)
		assert.InDelta(t, preset.SheetLayout.MarginTop-0.5, layout.MarginTop, 1e-9)
		// The preset fixes the label size and must not be modified
		assert.InDelta(t, preset.Width, data.Width, 1e-9)
		assert.InDelta(t, presetMarginLeft, labelmaker.GetPresetByKey(presetKey).SheetLayout.MarginLeft, 1e-9)
	})

	t.Run("custom layout overrides preset", func(t *testing.T) {
		custom := &labelmaker.SheetLayout{PageWidth: 210, PageHeight: 297, Columns: 2, Rows: 5, MarginTop: 8, MarginLeft: 6}
		template := repo.LabelTemplateOut{Preset: &presetKey, SheetLayout: custom}

		layout := sheetLayoutFor(template, &labelmaker.TemplateData{})
		require.NotNil(t, layout)
		assert.Equal(t, 2, layout.Columns)
		assert.InDelta(t, 6.0, layout.MarginLeft, 1e-9)
	})

	t.Run("no layout", func(t *testing.T) {
		assert.Nil(t, sheetLayoutFor(repo.LabelTemplateOut{OffsetX: 2}, &labelmaker.TemplateData{}))
	})
}
