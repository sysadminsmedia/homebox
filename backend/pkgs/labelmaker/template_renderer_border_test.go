package labelmaker

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The label border is a template-level frame stored in canvasData; these
// tests pin where it lands and that bad values cannot break the render.

const (
	borderEnabled = "enabled"
	borderWidth   = "width"
	borderInset   = "inset"
)

func renderBorderTemplate(t *testing.T, border map[string]interface{}) image.Image {
	t.Helper()
	r, err := NewTemplateRenderer()
	require.NoError(t, err)
	canvas := map[string]interface{}{"objects": []interface{}{}}
	if border != nil {
		canvas["labelBorder"] = border
	}
	out, err := r.RenderTemplate(RenderContext{Template: &TemplateData{Width: 50, Height: 25, DPI: 96, CanvasData: canvas}})
	require.NoError(t, err)
	img, err := png.Decode(bytes.NewReader(out))
	require.NoError(t, err)
	return img
}

func isBlackAt(img image.Image, x, y int) bool {
	c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
	return c.R == 0 && c.G == 0 && c.B == 0 && c.A == 255
}

func TestLabelBorderIsDrawnInsetFromTheEdges(t *testing.T) {
	img := renderBorderTemplate(t, map[string]interface{}{borderEnabled: true, borderWidth: 1.0, borderInset: 2.0, "color": "#000000"})
	b := img.Bounds()
	inset := mmToPixels(2, 96) // 7 px
	width := mmToPixels(1, 96) // 3 px
	midY := b.Dy() / 2
	midX := b.Dx() / 2

	assert.False(t, isBlackAt(img, 0, midY), "label edge must stay white")
	assert.False(t, isBlackAt(img, inset-1, midY), "pixel just outside the frame must stay white")
	assert.True(t, isBlackAt(img, inset, midY), "left band")
	assert.True(t, isBlackAt(img, inset+width-1, midY), "left band inner edge")
	assert.False(t, isBlackAt(img, inset+width, midY), "inside the frame must stay white")
	assert.True(t, isBlackAt(img, b.Dx()-1-inset, midY), "right band")
	assert.True(t, isBlackAt(img, midX, inset), "top band")
	assert.True(t, isBlackAt(img, midX, b.Dy()-1-inset), "bottom band")
	assert.False(t, isBlackAt(img, midX, midY), "centre must stay white")
}

func TestLabelBorderIsOffByDefault(t *testing.T) {
	for name, border := range map[string]map[string]interface{}{
		"absent":   nil,
		"disabled": {borderEnabled: false, borderWidth: 1.0, borderInset: 0.0},
		"garbage":  {borderEnabled: "yes"},
	} {
		img := renderBorderTemplate(t, border)
		b := img.Bounds()
		for _, p := range []image.Point{{0, 0}, {b.Dx() / 2, 0}, {3, b.Dy() / 2}, {b.Dx() / 2, b.Dy() / 2}} {
			assert.Falsef(t, isBlackAt(img, p.X, p.Y), "%s: pixel %v should be white", name, p)
		}
	}
}

func TestLabelBorderClampsOversizedValues(t *testing.T) {
	// 500 mm line on a 25 mm label: width is capped at 10 mm and the inset is
	// pulled in until the frame fits, so the render succeeds and stays bounded.
	img := renderBorderTemplate(t, map[string]interface{}{borderEnabled: true, borderWidth: 500.0, borderInset: 500.0})
	b := img.Bounds()
	assert.True(t, isBlackAt(img, b.Dx()/2, b.Dy()/2), "a frame wider than the label fills it")

	// A zero-size line still draws a one pixel frame at the edge.
	img = renderBorderTemplate(t, map[string]interface{}{borderEnabled: true, borderWidth: 0.0, borderInset: 0.0})
	assert.True(t, isBlackAt(img, 0, b.Dy()/2))
	assert.False(t, isBlackAt(img, 2, b.Dy()/2))
}
