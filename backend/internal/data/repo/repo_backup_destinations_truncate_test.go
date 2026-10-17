package repo

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestTruncateUTF8(t *testing.T) {
	assert.Equal(t, "abc", truncateUTF8("abc", 10))
	// "é" is two bytes; cutting at byte 3 would split the second one.
	out := truncateUTF8("aéé", 4)
	assert.Equal(t, "aé", out)
	assert.True(t, utf8.ValidString(out))
}
