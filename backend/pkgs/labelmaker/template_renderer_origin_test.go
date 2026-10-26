package labelmaker

import (
	"math"
	"testing"
)

const keyOriginX = "originX"

// canvasObject builds a Fabric-style object map for parser tests.
func canvasObject(kind string, left, top, width, height float64, extra map[string]interface{}) map[string]interface{} {
	m := map[string]interface{}{"type": kind, "left": left, "top": top, "width": width, "height": height}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func TestNormalizeOrigin(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 0.01 }
	centered := map[string]interface{}{keyOriginX: "center", "originY": "center"}

	// Fabric 7 default: left/top point at the centre of the object.
	elem := parseCanvasElement(canvasObject("textbox", 100, 50, 40, 20, map[string]interface{}{
		"scaleX": 2.0, keyOriginX: "center", "originY": "center",
	}))
	if !near(elem.Left, 60) || !near(elem.Top, 40) {
		t.Fatalf("center origin: got left=%.2f top=%.2f, want 60, 40", elem.Left, elem.Top)
	}
	if elem.OriginX != "left" || elem.OriginY != "top" {
		t.Fatalf("origin not normalised: %s/%s", elem.OriginX, elem.OriginY)
	}

	// Explicit left/top origins are untouched.
	elem = parseCanvasElement(canvasObject("rect", 10, 20, 30, 30, map[string]interface{}{keyOriginX: "left", "originY": "top"}))
	if !near(elem.Left, 10) || !near(elem.Top, 20) {
		t.Fatalf("left/top origin moved: left=%.2f top=%.2f", elem.Left, elem.Top)
	}

	// Rotation happens about the origin point: a centred 40x20 box rotated 90
	// degrees has its top-left corner at (cx+10, cy-20).
	centered["angle"] = 90.0
	elem = parseCanvasElement(canvasObject("rect", 100, 100, 40, 20, centered))
	if !near(elem.Left, 110) || !near(elem.Top, 80) {
		t.Fatalf("rotated center origin: got left=%.2f top=%.2f, want 110, 80", elem.Left, elem.Top)
	}

	// Numeric origins (0..1) are accepted too.
	elem = parseCanvasElement(canvasObject("rect", 100, 100, 40, 20, map[string]interface{}{keyOriginX: 1.0, "originY": 0.5}))
	if !near(elem.Left, 60) || !near(elem.Top, 90) {
		t.Fatalf("numeric origin: got left=%.2f top=%.2f, want 60, 90", elem.Left, elem.Top)
	}
}
