package textutils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sysadminsmedia/homebox/backend/pkgs/textutils"
)

func TestTokenizeSearchQuery(t *testing.T) {
	testCases := []struct {
		name  string
		query string
		want  []string
	}{
		{
			name:  "empty string",
			query: "",
			want:  nil,
		},
		{
			name:  "whitespace only",
			query: "   \t\n ",
			want:  nil,
		},
		{
			name:  "single token",
			query: "item",
			want:  []string{"item"},
		},
		{
			name:  "multiple tokens",
			query: "item blue",
			want:  []string{"item", "blue"},
		},
		{
			name:  "extra whitespace between tokens",
			query: "  item \t blue  ",
			want:  []string{"item", "blue"},
		},
		{
			name:  "quoted phrase",
			query: `"blue box"`,
			want:  []string{"blue box"},
		},
		{
			name:  "quoted phrase mixed with tokens",
			query: `red "blue box" large`,
			want:  []string{"red", "blue box", "large"},
		},
		{
			name:  "quoted phrase preserves inner punctuation",
			query: `"long description, blue"`,
			want:  []string{"long description, blue"},
		},
		{
			name:  "unbalanced quote consumes remainder",
			query: `red "blue box`,
			want:  []string{"red", "blue box"},
		},
		{
			name:  "empty quotes are dropped",
			query: `red "" blue`,
			want:  []string{"red", "blue"},
		},
		{
			name:  "adjacent quoted phrases",
			query: `"red box""blue box"`,
			want:  []string{"red box", "blue box"},
		},
		{
			name:  "unicode tokens",
			query: "electrónica café",
			want:  []string{"electrónica", "café"},
		},
		{
			name:  "unicode whitespace separator",
			query: "item\u00a0blue",
			want:  []string{"item", "blue"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, textutils.TokenizeSearchQuery(tc.query))
		})
	}
}
