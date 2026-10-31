package labelmaker

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

const alignCenter = "center"

// inkBounds returns the bounding box of non-white pixels.
func inkBounds(img image.Image) (image.Rectangle, bool) {
	b := img.Bounds()
	var box image.Rectangle
	found := false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			if r < 0x8000 || g < 0x8000 || bl < 0x8000 {
				p := image.Rect(x, y, x+1, y+1)
				if !found {
					box, found = p, true
				} else {
					box = box.Union(p)
				}
			}
		}
	}
	return box, found
}

func TestRotatedTextKeepsOriginAndIsNotClipped(t *testing.T) {
	r, err := NewTemplateRenderer()
	if err != nil {
		t.Fatal(err)
	}
	const left, top = 150.0, 20.0
	tpl := &TemplateData{
		Width:  60,
		Height: 60,
		DPI:    96, // 1 canvas px = 1 image px (canvas coordinates are at screenDPI)
		CanvasData: map[string]interface{}{
			"objects": []interface{}{
				canvasObject("textbox", left, top, 60, 24, map[string]interface{}{
					"text": "HELLO", "fontSize": 20.0, "fill": "#000000", "angle": 90.0,
				}),
			},
		},
	}
	out, err := r.RenderTemplate(RenderContext{Template: tpl})
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	box, ok := inkBounds(img)
	if !ok {
		t.Fatal("rotated text rendered nothing")
	}
	// Rotated 90 degrees clockwise about its top-left origin, the text runs
	// downwards from (left, top) and extends to the left of the origin.
	if box.Dy() <= box.Dx() {
		t.Errorf("expected vertical text, got bounds %v", box)
	}
	if box.Max.X > int(left)+3 || box.Max.X < int(left)-12 {
		t.Errorf("right edge %d should sit at the origin x=%d", box.Max.X, int(left))
	}
	if box.Min.Y < int(top)-3 || box.Min.Y > int(top)+12 {
		t.Errorf("top edge %d should sit at the origin y=%d", box.Min.Y, int(top))
	}
	if box.Dy() < 40 {
		t.Errorf("text looks clipped: height %d px for a 5-letter word", box.Dy())
	}
}

func TestRotatedTextRejectsOversizedBuffers(t *testing.T) {
	r, err := NewTemplateRenderer()
	if err != nil {
		t.Fatal(err)
	}
	// A centre-aligned box 100k px wide with a 2000 px font asks for a temporary
	// buffer far beyond maxRenderPixels; it must be refused before allocation.
	tpl := &TemplateData{
		Width:  60,
		Height: 60,
		DPI:    96,
		CanvasData: map[string]interface{}{
			"objects": []interface{}{
				canvasObject("textbox", 0, 0, 100000, 3000, map[string]interface{}{
					"text": "HUGE", "fontSize": 2000.0, "fill": "#000000", "angle": 45.0, "textAlign": alignCenter,
				}),
			},
		},
	}
	// RenderTemplate skips elements that fail, so check the element renderer itself.
	elem := parseCanvasElement(tpl.CanvasData["objects"].([]interface{})[0].(map[string]interface{}))
	img := image.NewRGBA(image.Rect(0, 0, 60, 60))
	if err := r.renderElement(img, elem, nil, nil, 96); err == nil {
		t.Fatal("expected an error for an oversized rotated text element")
	}
}
