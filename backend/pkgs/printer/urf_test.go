package printer

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestConvertPNGToURF(t *testing.T) {
	// Create a small test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 5))
	// Fill with solid color
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 128, 64, 255})
		}
	}

	// Encode as PNG
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}

	// Convert to URF
	urfData, err := ConvertPNGToURF(pngBuf.Bytes(), 300)
	if err != nil {
		t.Fatalf("ConvertPNGToURF failed: %v", err)
	}

	// Verify header
	if len(urfData) < 56 { // 12 byte file header + 44 byte page header
		t.Fatalf("URF data too short: %d bytes", len(urfData))
	}

	// Check magic
	magic := string(urfData[0:8])
	if magic != "UNIRAST\x00" {
		t.Errorf("incorrect magic: got %q, want %q", magic, "UNIRAST\x00")
	}

	// Check page count (at offset 8)
	pageCount := binary.BigEndian.Uint32(urfData[8:12])
	if pageCount != 1 {
		t.Errorf("incorrect page count: got %d, want 1", pageCount)
	}

	// Check page header (starts at offset 12)
	bpp := urfData[12]
	if bpp != 24 {
		t.Errorf("incorrect bpp: got %d, want 24", bpp)
	}

	colorspace := urfData[13]
	if colorspace != 3 {
		t.Errorf("incorrect colorspace: got %d, want 3 (sRGB)", colorspace)
	}

	// Check width at offset 12 + 12 = 24
	width := binary.BigEndian.Uint32(urfData[24:28])
	if width != 10 {
		t.Errorf("incorrect width: got %d, want 10", width)
	}

	// Check height at offset 12 + 16 = 28
	height := binary.BigEndian.Uint32(urfData[28:32])
	if height != 5 {
		t.Errorf("incorrect height: got %d, want 5", height)
	}

	// Check DPI at offset 12 + 20 = 32
	dpi := binary.BigEndian.Uint32(urfData[32:36])
	if dpi != 300 {
		t.Errorf("incorrect DPI: got %d, want 300", dpi)
	}

	// Verify the total page header is 44 bytes
	// So raster data starts at offset 12 + 44 = 56
	rasterStart := 56

	// First byte of raster should be line repeat count (0 = print once)
	if urfData[rasterStart] != 0 {
		t.Errorf("incorrect line repeat count: got %d, want 0", urfData[rasterStart])
	}

	t.Logf("URF file size: %d bytes", len(urfData))
	t.Logf("First 60 bytes: % x", urfData[:min(60, len(urfData))])
}

func TestConvertPNGToGrayscaleURF(t *testing.T) {
	// Create a small test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 5))
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255})
		}
	}

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}

	urfData, err := ConvertPNGToGrayscaleURF(pngBuf.Bytes(), 300)
	if err != nil {
		t.Fatalf("ConvertPNGToGrayscaleURF failed: %v", err)
	}

	// Check bpp (should be 8 for grayscale)
	bpp := urfData[12]
	if bpp != 8 {
		t.Errorf("incorrect bpp: got %d, want 8", bpp)
	}

	// Check colorspace (should be 1 for sGray)
	colorspace := urfData[13]
	if colorspace != 1 {
		t.Errorf("incorrect colorspace: got %d, want 1 (sGray)", colorspace)
	}

	t.Logf("Grayscale URF file size: %d bytes", len(urfData))
}

func TestPackBitsCompression(t *testing.T) {
	// Test with a run of identical pixels
	img := image.NewRGBA(image.Rect(0, 0, 100, 1))
	// Fill with solid white
	for x := 0; x < 100; x++ {
		img.Set(x, 0, color.RGBA{255, 255, 255, 255})
	}

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}

	urfData, err := ConvertPNGToURF(pngBuf.Bytes(), 300)
	if err != nil {
		t.Fatalf("ConvertPNGToURF failed: %v", err)
	}

	// A 100-pixel run of identical pixels should compress well
	// 12 (file header) + 44 (page header) + 1 (line repeat) + compressed data
	// Uncompressed would be: 12 + 44 + 1 + 100*3 = 357 bytes
	// With RLE compression: 12 + 44 + 1 + (code + 3 bytes) * ceil(100/128) = ~62 bytes
	t.Logf("URF file size for 100 identical pixels: %d bytes (uncompressed would be 357)", len(urfData))

	if len(urfData) > 100 {
		t.Logf("Compression ratio: %.1f%%", float64(len(urfData))/357*100)
	}
}
