package customsinventory

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Inventory list – Shipment A (Truck)": "inventory-list-shipment-a-truck",
		"Möbel & Bücher für Gärten":           "moebel-buecher-fuer-gaerten",
		"Straße":                              "strasse",
		"ÄÖÜ":                                 "aeoeue",
		"":                                    "export",
		"!!!":                                 "export",
	}
	for in, want := range cases {
		assert.Equal(t, want, Slug(in), "Slug(%q)", in)
	}

	long := Slug(strings.Repeat("a b ", 80))
	assert.LessOrEqual(t, len(long), 60)
	assert.False(t, strings.HasSuffix(long, "-"))
}

func TestBaseName(t *testing.T) {
	assert.Equal(t, "Inventarliste_shipment-a_2026-10-31", BaseName("Shipment A", time.Date(2026, 10, 31, 23, 59, 0, 0, time.UTC)))
	berlin := time.FixedZone("CET", 3600)
	assert.Equal(t, "Inventarliste_export_2026-10-31",
		BaseName("", time.Date(2026, 11, 1, 0, 30, 0, 0, berlin)), "the date is taken in UTC")
}

func TestFormatAssetID(t *testing.T) {
	cases := map[int64]string{123: "000-123", 123456: "123-456", 1234567: "123-4567", 0: "", -1: ""}
	for in, want := range cases {
		assert.Equal(t, want, FormatAssetID(in), "FormatAssetID(%d)", in)
	}
}

func TestFileExt(t *testing.T) {
	cases := []struct{ title, mime, want string }{
		{"rechnung.PDF", "application/pdf", ".pdf"},
		{"scan", "image/jpeg", ".jpg"},
		{"scan", "image/png", ".png"},
		{"scan", "image/heic", ".heic"},
		{"x", "application/x-foo", ".bin"},
		{"a.verylongext", "application/pdf", ".pdf"},
		{"scan", "image/jpeg; charset=binary", ".jpg"},
		{"../../etc/passwd", "", ".bin"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, FileExt(tc.title, tc.mime), "FileExt(%q, %q)", tc.title, tc.mime)
	}
}

func TestReceiptName(t *testing.T) {
	assert.Equal(t, "Belege/000-123_Rechnung.pdf", ReceiptName("000-123", 0, 0, "r.pdf", ""))
	assert.Equal(t, "Belege/000-123_b_Rechnung.pdf", ReceiptName("000-123", 1, 0, "r.pdf", ""))
	assert.Equal(t, "Belege/000-123_c_Rechnung.pdf", ReceiptName("000-123", 2, 0, "r.pdf", ""))
	assert.Equal(t, "Belege/000-123_Rechnung_2.jpg", ReceiptName("000-123", 0, 1, "r.jpg", ""))
	assert.Equal(t, "Belege/id-1a2b3c4d_Rechnung.pdf", ReceiptName("id-1a2b3c4d", 0, 0, "", "application/pdf"))
}
