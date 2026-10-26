package labelmaker

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression tests for the renderer's defences against untrusted template and
// canvas input: oversized output, extreme element scales, far-off coordinates
// and image decompression bombs must all be refused or clipped before any
// large allocation or long loop, and must do so quickly.

const (
	limitTestBudget = 3 * time.Second
	black           = "#000000"
	topLeft         = "left" // originX value; originY uses "top"
)

// whiteLabel returns a label-sized image with the white background RenderTemplate uses.
func whiteLabel(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 0xff
	}
	return img
}

func newLimitsRenderer(t *testing.T) *TemplateRenderer {
	t.Helper()
	r, err := NewTemplateRenderer()
	require.NoError(t, err)
	return r
}

// withinBudget fails the test when fn takes longer than limitTestBudget, which
// is how an unbounded loop or allocation would show up.
func withinBudget(t *testing.T, name string, fn func()) {
	t.Helper()
	start := time.Now()
	fn()
	assert.Lessf(t, time.Since(start), limitTestBudget, "%s took too long", name)
}

func TestRenderTemplateRejectsOversizedOutput(t *testing.T) {
	r := newLimitsRenderer(t)

	// 1000 x 1000 mm at 1200 dpi is about 2.2 gigapixels.
	withinBudget(t, "oversized template", func() {
		_, err := r.RenderTemplate(RenderContext{Template: &TemplateData{Width: 1000, Height: 1000, DPI: 1200}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds the renderer limit")
	})

	// A large but legitimate label still renders.
	out, err := r.RenderTemplate(RenderContext{Template: &TemplateData{Width: 100, Height: 150, DPI: 600}})
	require.NoError(t, err)
	cfg, err := png.DecodeConfig(bytes.NewReader(out))
	require.NoError(t, err)
	assert.Equal(t, mmToPixels(100, 600), cfg.Width)
}

func TestRenderElementsWithExtremeScaleAreBounded(t *testing.T) {
	r := newLimitsRenderer(t)
	img := whiteLabel(400, 200)
	item := &ItemData{ID: uuid.New(), Name: "x", ItemURL: "https://example.test/item/1"}
	huge := map[string]interface{}{"scaleX": 1e5, "scaleY": 1e5, keyOriginX: topLeft, "originY": "top"}

	cases := map[string]map[string]interface{}{
		"rect":    canvasObject("rect", 0, 0, 100, 100, mergeMaps(huge, map[string]interface{}{"fill": black, "stroke": "#ff0000", "strokeWidth": 2.0})),
		"image":   canvasObject("image", 0, 0, 100, 100, mergeMaps(huge, map[string]interface{}{"src": smallPNGDataURL(t)})),
		"barcode": canvasObject("group", 0, 0, 100, 100, mergeMaps(huge, map[string]interface{}{"data": map[string]interface{}{"type": "barcode", "format": "qr", "contentSource": "item_url"}})),
	}
	for name, obj := range cases {
		withinBudget(t, name, func() {
			elem := parseCanvasElement(obj)
			require.NoError(t, r.renderElement(img, elem, item, nil, 300))
		})
	}
	// The rect filled the whole label: its clipped draw still landed.
	assert.Equal(t, color.RGBA{0, 0, 0, 255}, img.RGBAAt(200, 100))
}

func TestRenderElementsFarOutsideLabelAreSkippedQuickly(t *testing.T) {
	r := newLimitsRenderer(t)
	item := &ItemData{ID: uuid.New(), Name: "x", ItemURL: "https://example.test/item/1"}
	far := map[string]interface{}{keyOriginX: topLeft, "originY": "top"}

	cases := map[string]map[string]interface{}{
		"rect negative":    canvasObject("rect", -1e9, -1e9, 100, 100, mergeMaps(far, map[string]interface{}{"fill": black})),
		"rect positive":    canvasObject("rect", 1e9, 1e9, 100, 100, mergeMaps(far, map[string]interface{}{"fill": black})),
		"image negative":   canvasObject("image", -1e9, -1e9, 100, 100, mergeMaps(far, map[string]interface{}{"src": smallPNGDataURL(t)})),
		"text negative":    canvasObject("textbox", -1e9, -1e9, 100, 20, mergeMaps(far, map[string]interface{}{"text": "far", "fontSize": 12.0, "fill": black})),
		"barcode negative": canvasObject("group", -1e9, -1e9, 100, 100, mergeMaps(far, map[string]interface{}{"data": map[string]interface{}{"type": "barcode", "format": "qr", "contentSource": "item_url"}})),
	}
	for name, obj := range cases {
		withinBudget(t, name, func() {
			img := whiteLabel(400, 200)
			elem := parseCanvasElement(obj)
			require.NoError(t, r.renderElement(img, elem, item, nil, 300))
			// Nothing may land on the label.
			box, inked := inkBounds(img)
			assert.Falsef(t, inked, "%s drew at %v although it lies outside the label", name, box)
		})
	}
}

func TestDecodeImageSourceRejectsBombsAndOversizedPayloads(t *testing.T) {
	// A 44-byte PNG that declares a 50000 x 50000 canvas; image.Decode would
	// allocate 10 GB for it. DecodeConfig sees the declared size first.
	bomb := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngHeaderDeclaring(50000, 50000))
	withinBudget(t, "decompression bomb", func() {
		_, err := decodeImageSource(bomb)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "limit is")
	})

	// Payload larger than the byte cap is refused before any decoding.
	big := "data:image/png;base64," + base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0}, maxImageSourceBytes+1))
	_, err := decodeImageSource(big)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bytes")

	// Non-data sources are never fetched.
	_, err = decodeImageSource("https://example.test/logo.png")
	require.Error(t, err)

	// A small embedded image still decodes.
	decoded, err := decodeImageSource(smallPNGDataURL(t))
	require.NoError(t, err)
	assert.Equal(t, image.Rect(0, 0, 4, 4), decoded.Bounds())
}

// pngHeaderDeclaring builds the PNG signature plus a valid IHDR chunk for the
// given size and nothing else, which is all DecodeConfig needs to read.
func pngHeaderDeclaring(width, height uint32) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:4], width)
	binary.BigEndian.PutUint32(ihdr[4:8], height)
	ihdr[8] = 8 // bit depth
	ihdr[9] = 6 // RGBA
	chunk := append([]byte("IHDR"), ihdr...)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(ihdr)))
	buf.Write(chunk)
	_ = binary.Write(&buf, binary.BigEndian, crc32.ChecksumIEEE(chunk))
	return buf.Bytes()
}

func smallPNGDataURL(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}
