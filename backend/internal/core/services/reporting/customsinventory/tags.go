package customsinventory

import "github.com/google/uuid"

// TagIndex answers parent/child questions about the tags of one group.
type TagIndex struct {
	byID     map[uuid.UUID]Tag
	children map[uuid.UUID][]uuid.UUID
}

func NewTagIndex(tags []Tag) TagIndex {
	ti := TagIndex{
		byID:     make(map[uuid.UUID]Tag, len(tags)),
		children: make(map[uuid.UUID][]uuid.UUID),
	}
	for _, t := range tags {
		ti.byID[t.ID] = t
		if t.ParentID != uuid.Nil {
			ti.children[t.ParentID] = append(ti.children[t.ParentID], t.ID)
		}
	}
	return ti
}

// Expand returns the given tags plus all their descendants. IDs that aren't
// in the index are dropped, so tags from another group never match.
func (ti TagIndex) Expand(ids []uuid.UUID) map[uuid.UUID]struct{} {
	out := make(map[uuid.UUID]struct{})
	queue := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := ti.byID[id]; ok {
			queue = append(queue, id)
		}
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if _, seen := out[id]; seen {
			continue
		}
		out[id] = struct{}{}
		queue = append(queue, ti.children[id]...)
	}
	return out
}

func (ti TagIndex) Name(id uuid.UUID) string {
	return ti.byID[id].Name
}

// FirstGroup returns the first of groupTags, in that order, that the item
// carries directly or through a descendant tag.
func (ti TagIndex) FirstGroup(itemTags, groupTags []uuid.UUID) (uuid.UUID, bool) {
	for _, g := range groupTags {
		if matchesAny(itemTags, ti.Expand([]uuid.UUID{g})) {
			return g, true
		}
	}
	return uuid.Nil, false
}

func matchesAny(ids []uuid.UUID, set map[uuid.UUID]struct{}) bool {
	for _, id := range ids {
		if _, ok := set[id]; ok {
			return true
		}
	}
	return false
}
