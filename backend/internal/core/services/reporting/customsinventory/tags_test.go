package customsinventory

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// id builds a readable, deterministic UUID such as 00000000-0000-0000-0000-000000000163.
func id(n int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
}

func keys(m map[uuid.UUID]struct{}) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestTagIndexExpand(t *testing.T) {
	root, child, grandchild, other := id(1), id(2), id(3), id(4)
	ti := NewTagIndex([]Tag{
		{ID: root, Name: "root"},
		{ID: child, ParentID: root, Name: "child"},
		{ID: grandchild, ParentID: child, Name: "grandchild"},
		{ID: other, Name: "other"},
	})

	assert.ElementsMatch(t, []uuid.UUID{root, child, grandchild}, keys(ti.Expand([]uuid.UUID{root})))
	assert.ElementsMatch(t, []uuid.UUID{child, grandchild}, keys(ti.Expand([]uuid.UUID{child})))
	assert.Empty(t, ti.Expand([]uuid.UUID{id(99)}), "unknown id is dropped")
	assert.Equal(t, "grandchild", ti.Name(grandchild))

	a, b := id(10), id(11)
	cyclic := NewTagIndex([]Tag{{ID: a, ParentID: b, Name: "a"}, {ID: b, ParentID: a, Name: "b"}})
	assert.ElementsMatch(t, []uuid.UUID{a, b}, keys(cyclic.Expand([]uuid.UUID{a})))
}

func TestTagIndexFirstGroup(t *testing.T) {
	a, b, childOfA, x := id(1), id(2), id(3), id(4)
	ti := NewTagIndex([]Tag{
		{ID: a, Name: "A"}, {ID: b, Name: "B"}, {ID: childOfA, ParentID: a, Name: "A child"}, {ID: x, Name: "X"},
	})

	got, ok := ti.FirstGroup([]uuid.UUID{b, childOfA}, []uuid.UUID{a, b})
	assert.True(t, ok)
	assert.Equal(t, a, got)

	got, ok = ti.FirstGroup([]uuid.UUID{x}, []uuid.UUID{a, b})
	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, got)
}
