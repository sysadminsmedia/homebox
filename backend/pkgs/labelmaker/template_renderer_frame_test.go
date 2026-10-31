package labelmaker

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Text frames: a text element with a wrap flag (or autofit) keeps the size the
// editor gave it. These tests render at 96 dpi so canvas pixels map 1:1.

const (
	frameKeyText   = "text"
	frameKeyFont   = "fontSize"
	frameKeyWrap   = "wrap"
	frameKeyWidth  = "fixedWidth"
	frameKeyHeight = "fixedHeight"
	frameLongText  = "alpha beta gamma delta epsilon zeta eta theta iota kappa"
)

func renderTextFrame(t *testing.T, data map[string]interface{}) (image.Rectangle, bool) {
	t.Helper()
	r := newLimitsRenderer(t)
	img := whiteLabel(300, 160)
	frame := map[string]interface{}{frameKeyWidth: 80.0}
	for k, v := range data {
		frame[k] = v
	}
	elem := parseCanvasElement(canvasObject("textbox", 20, 10, 80, 20, map[string]interface{}{
		frameKeyText: frameLongText, frameKeyFont: 16.0, "fill": black, keyOriginX: topLeft, keyOriginY: "top",
		"data": frame,
	}))
	require.NoError(t, r.renderElement(img, elem, nil, nil, 96))
	return inkBounds(img)
}

func TestTextFrameWrapsInsideItsWidth(t *testing.T) {
	ink, inked := renderTextFrame(t, map[string]interface{}{frameKeyWrap: true})
	require.True(t, inked)
	assert.LessOrEqual(t, ink.Max.X, 20+80, "wrapped lines stay inside the frame width")
	assert.Greater(t, ink.Dy(), 40, "the long text needs several lines")
}

func TestTextFrameCutsLinesBelowItsHeight(t *testing.T) {
	ink, inked := renderTextFrame(t, map[string]interface{}{frameKeyWrap: true, frameKeyHeight: 24.0})
	require.True(t, inked)
	assert.LessOrEqual(t, ink.Max.Y, 10+24, "nothing prints below the frame")
	assert.LessOrEqual(t, ink.Max.X, 20+80)
}

func TestTextFrameWithoutWrapKeepsOneLineCutAtTheEdge(t *testing.T) {
	ink, inked := renderTextFrame(t, map[string]interface{}{frameKeyWrap: false})
	require.True(t, inked)
	assert.LessOrEqual(t, ink.Max.X, 20+80, "the single line is cut at the frame's right edge")
	assert.Less(t, ink.Dy(), 24, "one line only")
}

func TestLegacyTextBoxStillPrintsOneUnclippedLine(t *testing.T) {
	// No wrap key and no autofit: boxes made before frames existed keep
	// printing a single line that may run past their width.
	ink, inked := renderTextFrame(t, nil)
	require.True(t, inked)
	assert.Greater(t, ink.Max.X, 20+80, "legacy text is not clipped")
	assert.Less(t, ink.Dy(), 24, "one line only")
}

func TestTextFrameRotatedStaysInsideItsRotatedBox(t *testing.T) {
	// Rotated 90 degrees clockwise about its top-left corner the 80 x 24 frame
	// occupies x in [left-24, left] and y in [top, top+80].
	r := newLimitsRenderer(t)
	img := whiteLabel(300, 160)
	elem := parseCanvasElement(canvasObject("textbox", 120, 10, 80, 24, map[string]interface{}{
		frameKeyText: frameLongText, frameKeyFont: 16.0, "fill": black, keyOriginX: topLeft, keyOriginY: "top",
		"angle": 90.0,
		"data":  map[string]interface{}{frameKeyWrap: true, frameKeyWidth: 80.0, frameKeyHeight: 24.0},
	}))
	require.NoError(t, r.renderElement(img, elem, nil, nil, 96))
	ink, inked := inkBounds(img)
	require.True(t, inked)
	assert.LessOrEqual(t, ink.Dx(), 24+4, "the rotated frame is only as wide as its height")
	assert.LessOrEqual(t, ink.Dy(), 80+4, "and as tall as its width")
	assert.GreaterOrEqual(t, ink.Min.X, 120-24-4)
	assert.LessOrEqual(t, ink.Max.X, 120+4)
}
