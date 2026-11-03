package customsinventory

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// snap rounds x to 6 decimals to remove binary noise before the final rounding.
func snap(x float64) float64 {
	return math.Round(x*1e6) / 1e6
}

// ToCents converts a price to cents, rounding half away from zero (CALC-1).
func ToCents(p float64) int64 {
	return int64(math.Round(snap(p * 100)))
}

// LineCents is the value of a line: quantity × price per piece, in cents (CALC-2).
func LineCents(qty float64, unitCents int64) int64 {
	return int64(math.Round(snap(qty * float64(unitCents))))
}

// RoundQty rounds a quantity to 6 decimals (CALC-5).
func RoundQty(q float64) float64 {
	return snap(q)
}

// FormatQtyDE renders a quantity with a German decimal comma and no
// trailing zeros, for example "2,5".
func FormatQtyDE(q float64) string {
	return strings.Replace(strconv.FormatFloat(RoundQty(q), 'f', -1, 64), ".", ",", 1)
}

// FormatCents renders cents as a plain decimal string, for example "1234.50".
func FormatCents(c int64) string {
	sign := ""
	if c < 0 {
		sign, c = "-", -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, c/100, c%100)
}
