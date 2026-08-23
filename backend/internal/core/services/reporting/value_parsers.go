package reporting

import (
	"strconv"
	"strings"

	"github.com/samber/lo"
)

func parseSeparatedString(s string, sep string) ([]string, error) {
	list := strings.Split(s, sep)

	trimmed := lo.Map(list, func(s string, _ int) string {
		return strings.TrimSpace(s)
	})

	return lo.Filter(trimmed, func(s string, _ int) bool {
		return s != ""
	}), nil
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func parseNillableFloat(s string) (*float64, error) {
	if s == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}

	return &f, nil
}
