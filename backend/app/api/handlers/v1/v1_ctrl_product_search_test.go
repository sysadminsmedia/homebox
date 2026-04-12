package v1

import "testing"

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

func TestSanitizeHeaderRemovesControlCharacters(t *testing.T) {
	got := sanitizeHeader("owner@example.com\r\nInjected: value\t")
	if got != "owner@example.comInjected: value" {
		t.Fatalf("unexpected sanitized header: %q", got)
	}
}
