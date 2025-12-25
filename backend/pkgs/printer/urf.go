package printer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
)

// URF (Unified Raster Format) encoder for AirPrint-compatible printers
// URF is Apple's raster format used by AirPrint printers
// Format based on reverse engineering: https://github.com/mbevand/urf2image

// ConvertPNGToURF converts a PNG image to URF format for AirPrint printers
func ConvertPNGToURF(pngData []byte, dpi int) ([]byte, error) {
	// Decode PNG
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if dpi <= 0 {
		dpi = 300
	}

	var buf bytes.Buffer

	// Write URF file header (12 bytes)
	// "UNIRAST\0" magic (8 bytes) + page count (4 bytes big-endian)
	buf.WriteString("UNIRAST\x00")
	_ = binary.Write(&buf, binary.BigEndian, uint32(1)) // 1 page

	// Write page header (44 bytes total)
	// Bytes 0-3: bpp, colorspace, duplex, quality
	buf.WriteByte(24) // bits per pixel (24 = RGB)
	buf.WriteByte(3)  // colorspace (3 = sRGB)
	buf.WriteByte(0)  // duplex (0 = simplex)
	buf.WriteByte(4)  // quality (4 = normal)
	// Bytes 4-7: unknown0
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))
	// Bytes 8-11: unknown1
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))
	// Bytes 12-15: width
	_ = binary.Write(&buf, binary.BigEndian, uint32(width))
	// Bytes 16-19: height
	_ = binary.Write(&buf, binary.BigEndian, uint32(height))
	// Bytes 20-23: DPI
	_ = binary.Write(&buf, binary.BigEndian, uint32(dpi))
	// Bytes 24-27: unknown2
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))
	// Bytes 28-31: unknown3
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))
	// Bytes 32-43: padding (3 more uint32s to reach 44 bytes)
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))

	// Write raster data
	// Each line starts with a line repeat count byte (0 = print once, n = repeat n+1 times)
	// Then PackBits-encoded pixel data follows
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// Collect pixels for this line (as BGR - bytes are reversed in URF)
		pixels := make([][]byte, width)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert from 16-bit to 8-bit and store as BGR (reversed)
			pixels[x-bounds.Min.X] = []byte{byte(b >> 8), byte(g >> 8), byte(r >> 8)}
		}

		// Write line repeat count (0 = print this line once)
		buf.WriteByte(0)

		// Write PackBits-encoded pixels
		// PackBits for URF works on whole pixels:
		// - Codes 0-127: repeat next pixel n+1 times
		// - Codes 128 (-128 signed): fill rest of line with white (0xFF)
		// - Codes 129-255 (-127 to -1 signed): copy next n+1 pixels literally (n = 256 - code)
		encodeLinePackBits(&buf, pixels)
	}

	return buf.Bytes(), nil
}

// encodeLinePackBits encodes a line of pixels using URF PackBits compression
func encodeLinePackBits(buf *bytes.Buffer, pixels [][]byte) {
	i := 0
	bytesPerPixel := len(pixels[0])

	for i < len(pixels) {
		// Check for a run of identical pixels
		runLen := 1
		for i+runLen < len(pixels) && runLen < 128 {
			if !bytesEqual(pixels[i], pixels[i+runLen]) {
				break
			}
			runLen++
		}

		if runLen >= 2 {
			// Run-length encode: code 0-127 means repeat pixel n+1 times
			buf.WriteByte(byte(runLen - 1))
			buf.Write(pixels[i])
			i += runLen
		} else {
			// Find literal run (non-repeating pixels)
			literalLen := 1
			for i+literalLen < len(pixels) && literalLen < 128 {
				// Check if next pixel starts a run of 2+ identical
				if i+literalLen+1 < len(pixels) &&
					bytesEqual(pixels[i+literalLen], pixels[i+literalLen+1]) {
					break
				}
				literalLen++
			}

			// Literal run: code 129-255 (or -127 to -1 as signed)
			// n = 256 - code, so code = 256 - n = 257 - literalLen
			code := byte(257 - literalLen)
			buf.WriteByte(code)
			for j := 0; j < literalLen; j++ {
				buf.Write(pixels[i+j])
			}
			i += literalLen
		}
	}

	// If we haven't filled the line, we could use code 128 to fill with white
	// But since we've written all pixels, we're done
	_ = bytesPerPixel // prevent unused variable warning
}

// bytesEqual compares two byte slices for equality
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ConvertPNGToGrayscaleURF converts a PNG to grayscale URF (for monochrome printers)
func ConvertPNGToGrayscaleURF(pngData []byte, dpi int) ([]byte, error) {
	// Decode PNG
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if dpi <= 0 {
		dpi = 300
	}

	var buf bytes.Buffer

	// Write URF file header (12 bytes)
	buf.WriteString("UNIRAST\x00")
	_ = binary.Write(&buf, binary.BigEndian, uint32(1))

	// Write page header (44 bytes total)
	buf.WriteByte(8)                                         // bits per pixel (8 = grayscale)
	buf.WriteByte(1)                                         // colorspace (1 = sGray)
	buf.WriteByte(0)                                         // duplex
	buf.WriteByte(4)                                         // quality
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))      // unknown0
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))      // unknown1
	_ = binary.Write(&buf, binary.BigEndian, uint32(width))  // width
	_ = binary.Write(&buf, binary.BigEndian, uint32(height)) // height
	_ = binary.Write(&buf, binary.BigEndian, uint32(dpi))    // DPI
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))      // unknown2
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))      // unknown3
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))      // padding
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))      // padding
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))      // padding

	// Write raster data
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		pixels := make([][]byte, width)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to grayscale using luminosity method
			gray := uint8((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256)
			pixels[x-bounds.Min.X] = []byte{gray}
		}

		// Write line repeat count
		buf.WriteByte(0)

		// Encode with PackBits
		encodeLinePackBits(&buf, pixels)
	}

	return buf.Bytes(), nil
}

// Ensure img implements the image.Image interface check
var _ image.Image = (*image.RGBA)(nil)
