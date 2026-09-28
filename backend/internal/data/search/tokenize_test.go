package search

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenize(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single word",
			input:    "hammer",
			expected: []string{"hammer"},
		},
		{
			name:     "multiple words",
			input:    "red tool box",
			expected: []string{"red", "tool", "box"},
		},
		{
			name:     "extra whitespace",
			input:    "  red \t tool\n",
			expected: []string{"red", "tool"},
		},
		{
			name:     "quoted phrase",
			input:    `red "tool box"`,
			expected: []string{"red", "tool box"},
		},
		{
			name:     "unterminated quote",
			input:    `red "tool box`,
			expected: []string{"red", "tool box"},
		},
		{
			name:     "empty quotes ignored",
			input:    `red ""`,
			expected: []string{"red"},
		},
		{
			name:     "duplicates removed",
			input:    "red red red",
			expected: []string{"red"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "unicode words",
			input:    "Тестовий Запис",
			expected: []string{"Тестовий", "Запис"},
		},
		{
			name:     "token count capped",
			input:    "t1 t2 t3 t4 t5 t6 t7 t8 t9 t10",
			expected: strings.Fields("t1 t2 t3 t4 t5 t6 t7 t8"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Tokenize(tc.input))
		})
	}
}
