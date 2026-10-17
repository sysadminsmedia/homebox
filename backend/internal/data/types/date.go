// Package types provides custom types for the application.
package types

import (
	"errors"
	"strings"
	"time"
)

// Date is a custom type that implements the MarshalJSON interface
// that applies date only formatting to the time.Time fields in order
// to avoid common time and timezone pitfalls when working with Times.
//
// Examples:
//
//	"2019-01-01" -> time.Time{2019-01-01 00:00:00 +0000 UTC}
//	"2019-01-01T21:10:30Z" -> time.Time{2019-01-01 00:00:00 +0000 UTC}
//	"2019-01-01T21:10:30+01:00" -> time.Time{2019-01-01 00:00:00 +0000 UTC}
type Date time.Time

func (d Date) Time() time.Time {
	return time.Time(d)
}

// DateFromTime returns a Date type from a time.Time type by stripping
// the time and timezone information.
func DateFromTime(t time.Time) Date {
	dateOnlyTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return Date(dateOnlyTime)
}

// DateFromDBTime returns a Date from a timestamp read out of a date-only
// database column.
//
// Date-only columns always hold UTC midnight of the intended calendar day —
// that is the invariant DateFromTime establishes on the way in. Drivers are
// not required to hand that instant back in UTC though: pgx decodes timestamptz
// into the process's local zone, so a server running with TZ=America/Chicago
// gets 2006-01-02T00:00:00Z back as 2006-01-01 18:00 CST. Reading the calendar
// components off that value yields the day before, and saving the entity again
// persists the shift. Normalize to UTC first so the stored day is recovered
// exactly, on every driver and in every server timezone.
func DateFromDBTime(t time.Time) Date {
	if t.IsZero() {
		return Date{}
	}

	return DateFromTime(t.UTC())
}

// DateFromString returns a Date type from a string by parsing the
// string into a time.Time type and then stripping the time and
// timezone information.
//
// Errors are ignored and an empty Date is returned.
func DateFromString(s string) Date {
	if s == "" {
		return Date{}
	}

	try := [...]string{
		"2006-01-02",
		"01/02/2006",
		"2006/01/02",
		time.RFC3339,
	}

	for _, format := range try {
		t, err := time.Parse(format, s)
		if err == nil {
			return DateFromTime(t)
		}
	}

	return Date{}
}

func (d Date) String() string {
	if time.Time(d).IsZero() {
		return ""
	}

	// Formatted in UTC, not in the value's own zone: every constructor here
	// normalizes to UTC midnight, so UTC is where the intended calendar day
	// lives. A Date produced by converting a driver-supplied time.Time
	// directly (types.Date(row.Field)) may still carry a local zone, and
	// formatting that in place would render the previous day west of
	// Greenwich. See DateFromDBTime.
	return time.Time(d).UTC().Format("2006-01-02")
}

func (d Date) MarshalJSON() ([]byte, error) {
	if time.Time(d).IsZero() {
		return []byte(`""`), nil
	}

	return []byte(`"` + d.String() + `"`), nil
}

func (d *Date) UnmarshalJSON(data []byte) (err error) {
	// unescape the string if necessary `\"` -> `"`
	str := strings.Trim(string(data), "\"")
	if str == "" || str == "null" || str == `""` {
		*d = Date{}
		return nil
	}

	try := [...]string{
		"2006-01-02",
		"01/02/2006",
		time.RFC3339,
	}

	set := false
	var t time.Time

	for _, format := range try {
		t, err = time.Parse(format, str)
		if err == nil {
			set = true
			break
		}
	}

	if !set {
		return errors.New("invalid date format")
	}

	// strip the time and timezone information
	*d = DateFromTime(t)

	return nil
}
