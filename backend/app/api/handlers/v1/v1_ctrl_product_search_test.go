package v1

import (
	"encoding/json"
	"testing"
)

func TestUPCITEMDBResponseUnmarshalNumericListPrice(t *testing.T) {
	body := []byte(`{
		"code": "OK",
		"total": 1,
		"offset": 0,
		"items": [{
			"title": "Example",
			"offers": [{
				"merchant": "ExampleStore",
				"list_price": 19.99,
				"price": 14.5,
				"shipping": 4.25
			}]
		}]
	}`)

	var result UPCITEMDBResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(result.Items) != 1 || len(result.Items[0].Offers) != 1 {
		t.Fatalf("expected one item with one offer, got items=%d", len(result.Items))
	}

	offer := result.Items[0].Offers[0]
	if offer.ListPrice != "19.99" {
		t.Fatalf("expected list_price %q, got %q", "19.99", offer.ListPrice)
	}
	if offer.Shipping != "4.25" {
		t.Fatalf("expected shipping %q, got %q", "4.25", offer.Shipping)
	}
}

func TestUPCITEMDBResponseUnmarshalStringListPrice(t *testing.T) {
	body := []byte(`{
		"items": [{
			"offers": [{
				"list_price": "19.99",
				"shipping": "Free"
			}]
		}]
	}`)

	var result UPCITEMDBResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(result.Items) == 0 || len(result.Items[0].Offers) == 0 {
		t.Fatalf("expected at least one item with one offer, got items=%d", len(result.Items))
	}

	offer := result.Items[0].Offers[0]
	if offer.ListPrice != "19.99" {
		t.Fatalf("expected list_price %q, got %q", "19.99", offer.ListPrice)
	}
	if offer.Shipping != "Free" {
		t.Fatalf("expected shipping %q, got %q", "Free", offer.Shipping)
	}
}

func TestUPCITEMDBResponseUnmarshalNullListPrice(t *testing.T) {
	body := []byte(`{
		"items": [{
			"offers": [{
				"list_price": null,
				"shipping": null
			}]
		}]
	}`)

	var result UPCITEMDBResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(result.Items) == 0 || len(result.Items[0].Offers) == 0 {
		t.Fatalf("expected at least one item with one offer, got items=%d", len(result.Items))
	}

	offer := result.Items[0].Offers[0]
	if offer.ListPrice != "" {
		t.Fatalf("expected empty list_price, got %q", offer.ListPrice)
	}
	if offer.Shipping != "" {
		t.Fatalf("expected empty shipping, got %q", offer.Shipping)
	}
}

func TestFlexibleStringRejectsCompositeTypes(t *testing.T) {
	cases := map[string]string{
		"object": `{"foo":"bar"}`,
		"array":  `[1,2,3]`,
		"bool":   `true`,
	}

	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			var f flexibleString
			if err := f.UnmarshalJSON([]byte(payload)); err == nil {
				t.Fatalf("expected error for %s payload, got nil (value=%q)", name, f)
			}
		})
	}
}

func TestFlexibleStringHandlesLeadingWhitespace(t *testing.T) {
	var f flexibleString
	if err := f.UnmarshalJSON([]byte("   12.50")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != "12.50" {
		t.Fatalf("expected %q, got %q", "12.50", f)
	}
}

func TestBuildOpenFactsBarcodeProduct(t *testing.T) {
	product, ok := buildOpenFactsBarcodeProduct("openbeautyfacts.org", "3600522058124", openFactsProduct{
		ProductName:   "Shower Gel",
		Brands:        "Example Brand",
		GenericName:   "Body wash",
		Categories:    "Hygiene, Shower",
		Quantity:      "250 ml",
		ImageFrontURL: "http://images.openbeautyfacts.org/images/products/360/052/205/8124/front_en.1.400.jpg",
	})

	if !ok {
		t.Fatal("expected product to be built")
	}
	if product.Barcode != "3600522058124" {
		t.Fatalf("unexpected barcode: %s", product.Barcode)
	}
	if product.SearchEngineName != "openbeautyfacts.org" {
		t.Fatalf("unexpected source name: %s", product.SearchEngineName)
	}
	if product.Item.Name != "Shower Gel" {
		t.Fatalf("unexpected product name: %s", product.Item.Name)
	}
	if product.Manufacturer != "Example Brand" {
		t.Fatalf("unexpected manufacturer: %s", product.Manufacturer)
	}
	if product.Item.Description != "Body wash | Hygiene, Shower | 250 ml" {
		t.Fatalf("unexpected description: %s", product.Item.Description)
	}
	if product.ImageURL != "https://images.openbeautyfacts.org/images/products/360/052/205/8124/front_en.1.400.jpg" {
		t.Fatalf("unexpected image URL: %s", product.ImageURL)
	}
}

func TestBuildOpenFactsBarcodeProductFallsBackToGenericName(t *testing.T) {
	product, ok := buildOpenFactsBarcodeProduct("openproductsfacts.org", "1234567890123", openFactsProduct{
		GenericName: "Replacement filter",
		Categories:  "Appliance parts",
	})

	if !ok {
		t.Fatal("expected product to be built")
	}
	if product.Item.Name != "Replacement filter" {
		t.Fatalf("unexpected product name: %s", product.Item.Name)
	}
	if product.Item.Description != "Appliance parts" {
		t.Fatalf("unexpected description: %s", product.Item.Description)
	}
}

func TestBuildOpenFactsBarcodeProductRequiresName(t *testing.T) {
	_, ok := buildOpenFactsBarcodeProduct("openfoodfacts.org", "1234567890123", openFactsProduct{})
	if ok {
		t.Fatal("expected empty product to be ignored")
	}
}

func TestBuildOpenFactsBarcodeProductRejectsUntrustedImageHost(t *testing.T) {
	product, ok := buildOpenFactsBarcodeProduct("openfoodfacts.org", "1234567890123", openFactsProduct{
		ProductName: "Example Product",
		ImageURL:    "https://example.com/image.jpg",
	})

	if !ok {
		t.Fatal("expected product to be built")
	}
	if product.ImageURL != "" {
		t.Fatalf("expected untrusted image URL to be cleared, got %q", product.ImageURL)
	}
}

func TestBuildOpenFactsBarcodeProductRejectsUnsupportedImageScheme(t *testing.T) {
	product, ok := buildOpenFactsBarcodeProduct("openfoodfacts.org", "1234567890123", openFactsProduct{
		ProductName: "Example Product",
		ImageURL:    "ftp://images.openfoodfacts.org/image.jpg",
	})

	if !ok {
		t.Fatal("expected product to be built")
	}
	if product.ImageURL != "" {
		t.Fatalf("expected unsupported image URL to be cleared, got %q", product.ImageURL)
	}
}

func TestSanitizeHeaderRemovesControlCharacters(t *testing.T) {
	got := sanitizeHeader("owner@example.com\r\nInjected: value\t")
	if got != "owner@example.comInjected: value" {
		t.Fatalf("unexpected sanitized header: %q", got)
	}
}
