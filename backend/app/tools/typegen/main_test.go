package main

import "testing"

func TestPostProcessTypesPreservesOptionalLocationPath(t *testing.T) {
	raw := `export interface EntitySummary {
  locationPath?: string;
  name?: string;
  parent?: EntitySummary | null;
}
export interface EntityOut {
  locationPath?: string;
  description?: string;
}
`
	want := `export interface EntitySummary {
  locationPath?: string;
  name: string;
  parent?: EntitySummary | null;
}
export interface EntityOut {
  locationPath?: string;
  description: string;
}
`
	got := postProcessTypes(raw)
	if got != want {
		t.Fatalf("post-processed API types differ:\nGot:\n%s\nWant:\n%s", got, want)
	}
}
