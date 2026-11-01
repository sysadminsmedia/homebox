package customsinventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCents(t *testing.T) {
	cases := map[float64]int64{
		0: 0, 19.99: 1999, 1.005: 101, 2.675: 268, 0.1 + 0.2: 30, -1.005: -101,
		1234567.891: 123456789, 0.004: 0, 0.005: 1,
	}
	for in, want := range cases {
		assert.Equal(t, want, ToCents(in), "ToCents(%v)", in)
	}
}

func TestLineCents(t *testing.T) {
	type c struct {
		q       float64
		u, want int64
	}
	for _, tc := range []c{
		{1, 1999, 1999}, {4, 8950, 35800}, {2.5, 33, 83}, {0.1, 33, 3},
		{3, 333, 999}, {14, 0, 0}, {1.5, 101, 152}, {-1, 500, -500},
	} {
		assert.Equal(t, tc.want, LineCents(tc.q, tc.u), "%v×%v", tc.q, tc.u)
	}
}

func TestFormatCents(t *testing.T) {
	cases := map[int64]string{0: "0.00", 5: "0.05", -101: "-1.01", 123456789: "1234567.89", -5: "-0.05"}
	for in, want := range cases {
		assert.Equal(t, want, FormatCents(in), "FormatCents(%v)", in)
	}
}

func TestRoundQty(t *testing.T) {
	assert.InDelta(t, 0.3, RoundQty(0.1+0.2), 0)
	assert.InDelta(t, 1.5, RoundQty(0.5+0.5+0.5), 0)
}
