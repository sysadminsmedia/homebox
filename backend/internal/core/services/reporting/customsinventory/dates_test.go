package customsinventory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

func d(s string) types.Date {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return types.DateFromTime(t)
}

func TestSubtractMonthsClamped(t *testing.T) {
	cases := map[string]string{
		"2026-08-31": "2026-02-28",
		"2028-08-31": "2028-02-29",
		"2026-03-15": "2025-09-15",
		"2026-12-31": "2026-06-30",
		"2026-01-31": "2025-07-31",
	}
	for in, want := range cases {
		assert.Equal(t, want, SubtractMonthsClamped(d(in), 6).String(), "SubtractMonthsClamped(%s, 6)", in)
	}
}

func TestCutoff(t *testing.T) {
	assert.Equal(t, "2026-05-15", Cutoff(d("2026-11-15"), 6).String())
	assert.Equal(t, "2026-05-15", Cutoff(d("2026-11-15"), 0).String())
	assert.Equal(t, "2025-11-15", Cutoff(d("2026-11-15"), 12).String())
}

func TestParseDate(t *testing.T) {
	got, err := ParseDate("2026-11-15")
	require.NoError(t, err)
	assert.Equal(t, d("2026-11-15"), got)

	for _, bad := range []string{"15.11.2026", "", "2026-13-01", "2026-11-15T10:00:00Z"} {
		_, err := ParseDate(bad)
		assert.Error(t, err, "ParseDate(%q)", bad)
	}
}

func TestFormatDE(t *testing.T) {
	assert.Equal(t, "15.11.2026", FormatDE(d("2026-11-15")))
	assert.Empty(t, FormatDE(types.Date{}))
}
