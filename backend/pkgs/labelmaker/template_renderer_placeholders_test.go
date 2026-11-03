package labelmaker

import "testing"

func TestReplacePlaceholdersBlanksUnknownFields(t *testing.T) {
	r := &TemplateRenderer{}

	item := &EntityData{Name: "Drill", AssetID: "000-042", CustomFields: map[string]string{"colour": "red"}}
	got := r.replaceDataPlaceholders("{{item_name}} #{{asset_id}} {{custom_colour}} [{{location_name}}] {{custom_missing}}", item)
	want := "Drill #000-042 red [] "
	if got != want {
		t.Fatalf("item placeholders: got %q, want %q", got, want)
	}

	// Text without placeholders, and braces that are not placeholders, pass through unchanged.
	for _, text := range []string{"plain text", "{{ spaced }}", "{{}}", "{not-a-placeholder}"} {
		if got := r.replaceDataPlaceholders(text, item); got != text {
			t.Fatalf("%q changed to %q", text, got)
		}
	}
}

// Items and locations are both entities: every field resolves for either,
// and the generic url follows the entity's kind.
func TestEntityFieldsShared(t *testing.T) {
	r := &TemplateRenderer{}

	loc := &EntityData{
		Name:         "Garage",
		AssetID:      "000-007",
		IsLocation:   true,
		LocationName: "Garage",
		LocationPath: []string{"Home", "Garage"},
		ItemCount:    3,
		ItemURL:      "https://hb.test/item/x",
		LocationURL:  "https://hb.test/location/x",
	}
	got := r.replaceDataPlaceholders("{{name}} ({{item_count}}) #{{asset_id}} {{location_path}} {{url}}", loc)
	want := "Garage (3) #000-007 Home > Garage https://hb.test/location/x"
	if got != want {
		t.Fatalf("location placeholders: got %q, want %q", got, want)
	}

	// Legacy location-only names still resolve
	if got := r.getFieldValue("path", loc); got != "Home > Garage" {
		t.Fatalf("path alias: got %q", got)
	}
	if got := r.getBarcodeContent("location_url", loc); got != loc.LocationURL {
		t.Fatalf("location_url barcode: got %q", got)
	}
	// A fixed value is encoded verbatim for locations too
	if got := r.getBarcodeContent("012345678905", loc); got != "012345678905" {
		t.Fatalf("fixed barcode value: got %q", got)
	}

	item := &EntityData{Name: "Drill", ItemURL: "https://hb.test/item/y", LocationURL: "https://hb.test/location/y"}
	if got := r.getFieldValue("url", item); got != item.ItemURL {
		t.Fatalf("item url: got %q", got)
	}
}
