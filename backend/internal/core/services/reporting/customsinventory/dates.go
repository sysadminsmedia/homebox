package customsinventory

import (
	"time"

	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

// ParseDate parses a strict YYYY-MM-DD date.
func ParseDate(s string) (types.Date, error) {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return types.Date{}, err
	}
	return types.DateFromTime(t), nil
}

// SubtractMonthsClamped goes back the given number of months. If the day
// doesn't exist in the target month, it uses that month's last day.
func SubtractMonthsClamped(d types.Date, months int) types.Date {
	t := d.Time().UTC()
	idx := t.Year()*12 + int(t.Month()) - 1 - months
	year, month := idx/12, time.Month(idx%12+1)
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	return types.Date(time.Date(year, month, min(t.Day(), lastDay), 0, 0, 0, 0, time.UTC))
}

// Cutoff is the reference date minus the minimum ownership period (CALC-4).
// months 0 means DefaultMinMonths.
func Cutoff(ref types.Date, months int) types.Date {
	if months == 0 {
		months = DefaultMinMonths
	}
	return SubtractMonthsClamped(ref, months)
}

// FormatDE renders a date as dd.mm.yyyy, or "" for the zero date.
func FormatDE(d types.Date) string {
	if d.Time().IsZero() {
		return ""
	}
	return d.Time().UTC().Format("02.01.2006")
}
