package labelmaker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateBarcode(t *testing.T) {
	opts := BarcodeOptions{Width: 469, Height: 156}

	tests := []struct {
		name    string
		format  BarcodeFormat
		content string
		wantErr bool
	}{
		{"code128 short", BarcodeCode128, "SN123", false},
		// Needs more modules than the box has pixels; must shrink, not fail
		{"code128 long url", BarcodeCode128, "https://homebox.example.com/item/7aefa6d0-909b-48c6-b625-545a792ec07a", false},
		{"qr url", BarcodeQR, "https://homebox.example.com/item/7aefa6d0-909b-48c6-b625-545a792ec07a", false},
		// Asset IDs are formatted with a dash and are shorter than EAN/UPC data
		{"ean13 asset id", BarcodeEAN13, "000-001", false},
		{"ean8 asset id", BarcodeEAN8, "000-001", false},
		{"upca asset id", BarcodeUPCA, "000-001", false},
		{"ean13 full", BarcodeEAN13, "123456789012", false},
		{"ean13 not numeric", BarcodeEAN13, "SN123", true},
		{"ean8 too long", BarcodeEAN8, "123456789", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			img, err := GenerateBarcode(tc.format, tc.content, opts)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, opts.Width, img.Bounds().Dx())
			assert.Equal(t, opts.Height, img.Bounds().Dy())
		})
	}
}

func TestNormalizeNumeric(t *testing.T) {
	got, err := normalizeNumeric("000-123", 7)
	require.NoError(t, err)
	assert.Equal(t, "0000123", got)

	got, err = normalizeNumeric("12345678", 7)
	require.NoError(t, err)
	assert.Equal(t, "12345678", got)

	_, err = normalizeNumeric("12a", 7)
	require.Error(t, err)
}
