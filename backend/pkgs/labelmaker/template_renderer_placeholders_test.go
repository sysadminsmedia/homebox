package labelmaker

import "testing"

func TestReplacePlaceholdersBlanksUnknownFields(t *testing.T) {
	r := &TemplateRenderer{}

	item := &ItemData{Name: "Drill", AssetID: "000-042", CustomFields: map[string]string{"colour": "red"}}
	got := r.replaceDataPlaceholders("{{item_name}} #{{asset_id}} {{custom_colour}} [{{location_name}}] {{item_count}}", item)
	want := "Drill #000-042 red [] "
	if got != want {
		t.Fatalf("item placeholders: got %q, want %q", got, want)
	}

	loc := &LocationData{Name: "Garage", FullPath: "Home > Garage", ItemCount: 3}
	got = r.replaceLocationPlaceholders("{{location_name}} ({{item_count}}) #{{asset_id}} SN {{serial_number}}", loc)
	want = "Garage (3) # SN "
	if got != want {
		t.Fatalf("location placeholders: got %q, want %q", got, want)
	}

	// Text without placeholders, and braces that are not placeholders, pass through unchanged.
	for _, text := range []string{"plain text", "{{ spaced }}", "{{}}", "{not-a-placeholder}"} {
		if got := r.replaceLocationPlaceholders(text, loc); got != text {
			t.Fatalf("%q changed to %q", text, got)
		}
	}
}
