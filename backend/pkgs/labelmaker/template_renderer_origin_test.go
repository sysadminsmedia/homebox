package labelmaker

import (
	"math"
	"testing"
)

func TestNormalizeOrigin(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 0.01 }

	// Fabric 7 default: left/top point at the centre of the object.
	elem := parseCanvasElement(map[string]interface{}{
		"type": "textbox", "left": 100.0, "top": 50.0, "width": 40.0, "height": 20.0,
		"scaleX": 2.0, "originX": "center", "originY": "center",
	})
	if !near(elem.Left, 60) || !near(elem.Top, 40) {
		t.Fatalf("center origin: got left=%.2f top=%.2f, want 60, 40", elem.Left, elem.Top)
	}
	if elem.OriginX != "left" || elem.OriginY != "top" {
		t.Fatalf("origin not normalised: %s/%s", elem.OriginX, elem.OriginY)
	}

	// Explicit left/top origins are untouched.
	elem = parseCanvasElement(map[string]interface{}{
		"type": "rect", "left": 10.0, "top": 20.0, "width": 30.0, "height": 30.0,
		"originX": "left", "originY": "top",
	})
	if !near(elem.Left, 10) || !near(elem.Top, 20) {
		t.Fatalf("left/top origin moved: left=%.2f top=%.2f", elem.Left, elem.Top)
	}

	// Rotation happens about the origin point: a centred 40x20 box rotated 90
	// degrees has its top-left corner at (cx+10, cy-20).
	elem = parseCanvasElement(map[string]interface{}{
		"type": "rect", "left": 100.0, "top": 100.0, "width": 40.0, "height": 20.0,
		"angle": 90.0, "originX": "center", "originY": "center",
	})
	if !near(elem.Left, 110) || !near(elem.Top, 80) {
		t.Fatalf("rotated center origin: got left=%.2f top=%.2f, want 110, 80", elem.Left, elem.Top)
	}

	// Numeric origins (0..1) are accepted too.
	elem = parseCanvasElement(map[string]interface{}{
		"type": "rect", "left": 100.0, "top": 100.0, "width": 40.0, "height": 20.0,
		"originX": 1.0, "originY": 0.5,
	})
	if !near(elem.Left, 60) || !near(elem.Top, 90) {
		t.Fatalf("numeric origin: got left=%.2f top=%.2f, want 60, 90", elem.Left, elem.Top)
	}
}
