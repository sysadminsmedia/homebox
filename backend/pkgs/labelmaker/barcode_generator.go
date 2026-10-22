// Package labelmaker provides functionality for generating and printing labels
package labelmaker

import (
	"fmt"
	"image"
	"image/color"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/datamatrix"
	"github.com/boombuler/barcode/ean"
	"github.com/boombuler/barcode/qr"
)

// BarcodeFormat represents supported barcode types
type BarcodeFormat string

const (
	BarcodeQR         BarcodeFormat = "qr"
	BarcodeCode128    BarcodeFormat = "code128"
	BarcodeCode39     BarcodeFormat = "code39"
	BarcodeDataMatrix BarcodeFormat = "datamatrix"
	BarcodeEAN13      BarcodeFormat = "ean13"
	BarcodeEAN8       BarcodeFormat = "ean8"
	BarcodeUPCA       BarcodeFormat = "upca"
)

// BarcodeOptions contains configuration for barcode generation
type BarcodeOptions struct {
	Width           int                     // Target width in pixels
	Height          int                     // Target height in pixels
	ErrorCorrection qr.ErrorCorrectionLevel // For QR codes
	IncludeText     bool                    // Include human-readable text below barcode
}

// DefaultBarcodeOptions returns sensible defaults for barcode generation
func DefaultBarcodeOptions() BarcodeOptions {
	return BarcodeOptions{
		Width:           200,
		Height:          200,
		ErrorCorrection: qr.M,
		IncludeText:     false,
	}
}

// GenerateBarcode creates a barcode image of the specified format
func GenerateBarcode(format BarcodeFormat, content string, opts BarcodeOptions) (image.Image, error) {
	if content == "" {
		return nil, fmt.Errorf("barcode content cannot be empty")
	}

	var bc barcode.Barcode
	var err error

	switch format {
	case BarcodeQR:
		bc, err = generateQR(content, opts)
	case BarcodeCode128:
		bc, err = generateCode128(content)
	case BarcodeCode39:
		bc, err = generateCode39(content)
	case BarcodeDataMatrix:
		bc, err = generateDataMatrix(content)
	case BarcodeEAN13:
		bc, err = generateEAN13(content)
	case BarcodeEAN8:
		bc, err = generateEAN8(content)
	case BarcodeUPCA:
		bc, err = generateUPCA(content)
	default:
		return nil, fmt.Errorf("unsupported barcode format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate %s barcode: %w", format, err)
	}

	// Scale to target size
	if opts.Width > 0 && opts.Height > 0 {
		bc, err = barcode.Scale(bc, opts.Width, opts.Height)
		if err != nil {
			return nil, fmt.Errorf("failed to scale barcode: %w", err)
		}
	}

	return bc, nil
}

// generateQR creates a QR code
func generateQR(content string, opts BarcodeOptions) (barcode.Barcode, error) {
	return qr.Encode(content, opts.ErrorCorrection, qr.Auto)
}

// generateCode128 creates a Code 128 barcode
func generateCode128(content string) (barcode.Barcode, error) {
	return code128.Encode(content)
}

// generateCode39 creates a Code 39 barcode
func generateCode39(content string) (barcode.Barcode, error) {
	return code39.Encode(content, true, true)
}

// generateDataMatrix creates a DataMatrix barcode
func generateDataMatrix(content string) (barcode.Barcode, error) {
	return datamatrix.Encode(content)
}

// generateEAN13 creates an EAN-13 barcode
func generateEAN13(content string) (barcode.Barcode, error) {
	// EAN-13 requires exactly 12 or 13 digits
	if len(content) < 12 || len(content) > 13 {
		return nil, fmt.Errorf("EAN-13 requires 12 or 13 digits, got %d", len(content))
	}
	return ean.Encode(content)
}

// generateEAN8 creates an EAN-8 barcode
func generateEAN8(content string) (barcode.Barcode, error) {
	// EAN-8 requires exactly 7 or 8 digits
	if len(content) < 7 || len(content) > 8 {
		return nil, fmt.Errorf("EAN-8 requires 7 or 8 digits, got %d", len(content))
	}
	return ean.Encode(content)
}

// generateUPCA creates a UPC-A barcode (via EAN-13 with leading zero)
func generateUPCA(content string) (barcode.Barcode, error) {
	// UPC-A requires exactly 11 or 12 digits
	if len(content) < 11 || len(content) > 12 {
		return nil, fmt.Errorf("UPC-A requires 11 or 12 digits, got %d", len(content))
	}
	// UPC-A is EAN-13 with a leading zero
	if len(content) == 11 {
		content = "0" + content
	} else if len(content) == 12 {
		content = "0" + content[:11] // Will recalculate check digit
	}
	return ean.Encode(content)
}

// GetSupportedFormats returns a list of all supported barcode formats
func GetSupportedFormats() []BarcodeFormat {
	return []BarcodeFormat{
		BarcodeQR,
		BarcodeCode128,
		BarcodeCode39,
		BarcodeDataMatrix,
		BarcodeEAN13,
		BarcodeEAN8,
		BarcodeUPCA,
	}
}

// ContentType indicates what kind of content a barcode format can encode
type ContentType string

const (
	ContentTypeAny          ContentType = "any"          // Can encode any text (URLs, names, etc.)
	ContentTypeAlphanumeric ContentType = "alphanumeric" // Letters, numbers, limited symbols
	ContentTypeNumeric      ContentType = "numeric"      // Digits only
)

// BarcodeFormatInfo provides metadata about a barcode format
type BarcodeFormatInfo struct {
	Format      BarcodeFormat `json:"format"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Is2D        bool          `json:"is2D"`
	MaxLength   int           `json:"maxLength"`   // 0 means variable/unlimited
	ContentType ContentType   `json:"contentType"` // What kind of content this format supports
}

// GetBarcodeFormatInfo returns metadata for all supported formats
func GetBarcodeFormatInfo() []BarcodeFormatInfo {
	return []BarcodeFormatInfo{
		{
			Format:      BarcodeQR,
			Name:        "QR Code",
			Description: "2D barcode that can encode URLs, text, and other data",
			Is2D:        true,
			MaxLength:   4296,
			ContentType: ContentTypeAny,
		},
		{
			Format:      BarcodeCode128,
			Name:        "Code 128",
			Description: "High-density 1D barcode for alphanumeric data",
			Is2D:        false,
			MaxLength:   0,
			ContentType: ContentTypeAny, // Supports full ASCII
		},
		{
			Format:      BarcodeCode39,
			Name:        "Code 39",
			Description: "Variable length 1D barcode (A-Z, 0-9, limited symbols)",
			Is2D:        false,
			MaxLength:   0,
			ContentType: ContentTypeAlphanumeric, // Limited charset
		},
		{
			Format:      BarcodeDataMatrix,
			Name:        "DataMatrix",
			Description: "2D barcode for small items and industrial marking",
			Is2D:        true,
			MaxLength:   2335,
			ContentType: ContentTypeAny,
		},
		{
			Format:      BarcodeEAN13,
			Name:        "EAN-13",
			Description: "European Article Number, 13-digit numeric code",
			Is2D:        false,
			MaxLength:   13,
			ContentType: ContentTypeNumeric,
		},
		{
			Format:      BarcodeEAN8,
			Name:        "EAN-8",
			Description: "Compact 8-digit numeric code for small packages",
			Is2D:        false,
			MaxLength:   8,
			ContentType: ContentTypeNumeric,
		},
		{
			Format:      BarcodeUPCA,
			Name:        "UPC-A",
			Description: "Universal Product Code, 12-digit numeric barcode",
			Is2D:        false,
			MaxLength:   12,
			ContentType: ContentTypeNumeric,
		},
	}
}

// CreatePlaceholderBarcode creates a placeholder image for barcode preview
func CreatePlaceholderBarcode(width, height int, format BarcodeFormat) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with light gray
	gray := color.RGBA{R: 238, G: 238, B: 238, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, gray)
		}
	}

	// Draw border
	borderColor := color.RGBA{R: 204, G: 204, B: 204, A: 255}
	for x := 0; x < width; x++ {
		img.Set(x, 0, borderColor)
		img.Set(x, height-1, borderColor)
	}
	for y := 0; y < height; y++ {
		img.Set(0, y, borderColor)
		img.Set(width-1, y, borderColor)
	}

	return img
}
