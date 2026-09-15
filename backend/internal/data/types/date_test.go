package types

import (
	"encoding/json"
	"testing"
	"time"
)

// zones spanning both sides of Greenwich; the western ones are where the
// off-by-one used to show up.
var zones = []string{
	"UTC",
	"America/Los_Angeles",
	"America/Chicago",
	"America/New_York",
	"Europe/Berlin",
	"Pacific/Auckland",
}

func loadZones(t *testing.T) []*time.Location {
	t.Helper()

	locs := make([]*time.Location, 0, len(zones))
	for _, name := range zones {
		loc, err := time.LoadLocation(name)
		if err != nil {
			t.Skipf("timezone database unavailable: %v", err)
		}
		locs = append(locs, loc)
	}
	return locs
}

// A date-only column holds UTC midnight of the intended day. pgx hands that
// instant back in the process's local zone, so the calendar components must be
// read in UTC or the day slips backwards west of Greenwich.
func TestDateFromDBTimePreservesDayInEveryZone(t *testing.T) {
	const want = "2026-04-18"
	stored := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)

	for _, loc := range loadZones(t) {
		asDriverReturnsIt := stored.In(loc)

		if got := DateFromDBTime(asDriverReturnsIt).String(); got != want {
			t.Errorf("DateFromDBTime in %s = %q, want %q", loc, got, want)
		}
	}
}

func TestDateFromDBTimeZeroStaysZero(t *testing.T) {
	if got := DateFromDBTime(time.Time{}).String(); got != "" {
		t.Errorf("DateFromDBTime(zero) = %q, want empty", got)
	}
}

// Date.String is the last line of defence: a Date built by converting a
// driver value directly (types.Date(row.Field)) still carries the local zone.
func TestDateStringRendersUTCDay(t *testing.T) {
	const want = "2026-04-18"
	stored := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)

	for _, loc := range loadZones(t) {
		d := Date(stored.In(loc))

		if got := d.String(); got != want {
			t.Errorf("Date(%v).String() = %q, want %q", loc, got, want)
		}

		b, err := json.Marshal(d)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(b) != `"`+want+`"` {
			t.Errorf("Date(%v) marshalled to %s, want %q", loc, b, want)
		}
	}
}

// The full API round-trip: the browser sends YYYY-MM-DD, we persist UTC
// midnight, the driver reads it back in the server's zone, and the response
// must carry the same day the user picked — and keep doing so when the
// unchanged value is saved again.
func TestDateRoundTripIsStableAcrossSaves(t *testing.T) {
	const want = "2026-08-27"

	for _, loc := range loadZones(t) {
		var body struct {
			PurchaseDate Date `json:"purchaseDate"`
		}
		if err := json.Unmarshal([]byte(`{"purchaseDate":"`+want+`"}`), &body); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		date := body.PurchaseDate
		for hop := 1; hop <= 3; hop++ {
			// Persist, then read back the way the driver would hand it over.
			date = DateFromDBTime(date.Time().In(loc))

			if got := date.String(); got != want {
				t.Fatalf("save %d in %s produced %q, want %q", hop, loc, got, want)
			}
		}
	}
}

func TestDateFromStringHonoursTheSendersOffset(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"2026-04-18", "2026-04-18"},
		// An offset-bearing timestamp means the day its own offset shows,
		// not the day it lands on in UTC.
		{"2026-04-18T00:00:00+02:00", "2026-04-18"},
		{"2026-04-18T00:00:00Z", "2026-04-18"},
		// Why clients must not send Date.toISOString(): a UTC-normalized
		// timestamp has already lost the picked day before it reaches us,
		// and nothing here can recover it. This is the shape the pre-v0.26
		// frontend sent, and the reason those rows are a day early at rest.
		{"2026-04-17T22:00:00Z", "2026-04-17"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := DateFromString(tt.in).String(); got != tt.want {
			t.Errorf("DateFromString(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
