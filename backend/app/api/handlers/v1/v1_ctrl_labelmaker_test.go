package v1

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Repeated across the table below; extracted to satisfy goconst.
const (
	fixtureLocationParam    = "locationLabel"
	fixtureLocationFallback = "Location"
)

func TestLabelText(t *testing.T) {
	cases := []struct {
		name     string
		query    string
		param    string
		fallback string
		expect   string
	}{
		{
			name:     "missing parameter falls back to English",
			query:    "/labelmaker/entity/1",
			param:    fixtureLocationParam,
			fallback: fixtureLocationFallback,
			expect:   fixtureLocationFallback,
		},
		{
			name:     "translation is used when supplied",
			query:    "/labelmaker/entity/1?locationLabel=Emplacement",
			param:    fixtureLocationParam,
			fallback: fixtureLocationFallback,
			expect:   "Emplacement",
		},
		{
			name:     "empty value falls back to English",
			query:    "/labelmaker/entity/1?locationLabel=",
			param:    fixtureLocationParam,
			fallback: fixtureLocationFallback,
			expect:   fixtureLocationFallback,
		},
		{
			name:     "whitespace-only value falls back to English",
			query:    "/labelmaker/entity/1?locationLabel=%20%20",
			param:    fixtureLocationParam,
			fallback: fixtureLocationFallback,
			expect:   fixtureLocationFallback,
		},
		{
			name:     "surrounding whitespace is trimmed",
			query:    "/labelmaker/entity/1?locationLabel=%20Ubicaci%C3%B3n%20",
			param:    fixtureLocationParam,
			fallback: fixtureLocationFallback,
			expect:   "Ubicación",
		},
		{
			name:     "a different parameter is unaffected",
			query:    "/labelmaker/location/1?locationDescription=Emplacement%20Homebox",
			param:    "locationDescription",
			fallback: "Homebox Location",
			expect:   "Emplacement Homebox",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.query, nil)
			assert.Equal(t, tt.expect, labelText(r, tt.param, tt.fallback))
		})
	}
}
