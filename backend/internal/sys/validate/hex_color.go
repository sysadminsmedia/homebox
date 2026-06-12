package validate

import (
	"fmt"
	"regexp"
	"strings"
)

var hexColorPattern = regexp.MustCompile(`^#?(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// NormalizeHexColor validates and canonicalizes a nullable hex color.
// Accepted inputs are #RGB, #RRGGBB, RGB, and RRGGBB.
func NormalizeHexColor(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	input := strings.TrimSpace(*value)
	if !hexColorPattern.MatchString(input) {
		return nil, fmt.Errorf("invalid color %q: expected #RGB or #RRGGBB", *value)
	}

	input = strings.TrimPrefix(input, "#")
	if len(input) == 3 {
		input = strings.Repeat(string(input[0]), 2) +
			strings.Repeat(string(input[1]), 2) +
			strings.Repeat(string(input[2]), 2)
	}

	normalized := "#" + strings.ToUpper(input)
	return &normalized, nil
}