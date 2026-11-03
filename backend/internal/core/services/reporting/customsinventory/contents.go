package customsinventory

import (
	"maps"
	"slices"
	"strings"

	"github.com/google/uuid"
)

// universe returns the scope IDs, deduplicated and in order. With
// RollUpContents on, it appends, in item order, every item outside the scope
// whose parent is an included item that is in scope or is itself content
// (FR-28). An excluded item is listed, but nothing below it is.
func universe(in Input, byID map[uuid.UUID]Item) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(in.ScopeIDs))
	listed := make(map[uuid.UUID]bool, len(in.ScopeIDs))
	for _, id := range in.ScopeIDs {
		if !listed[id] {
			listed[id] = true
			ids = append(ids, id)
		}
	}
	if !in.Options.RollUpContents {
		return ids
	}

	scope := maps.Clone(listed)
	memo := map[uuid.UUID]bool{}
	visiting := map[uuid.UUID]bool{}
	// holds reports whether the entity id is included and in scope or content.
	var holds func(id uuid.UUID) bool
	holds = func(id uuid.UUID) bool {
		it, ok := byID[id]
		if _, excluded := exclusion(it, ok); excluded {
			return false
		}
		if scope[id] {
			return true
		}
		if v, ok := memo[id]; ok {
			return v
		}
		if visiting[id] {
			return false // a cycle that never reaches the scope
		}
		visiting[id] = true
		memo[id] = holds(it.ParentID)
		return memo[id]
	}
	for _, it := range in.Items {
		if !listed[it.ID] && holds(it.ParentID) {
			listed[it.ID] = true
			ids = append(ids, it.ID)
		}
	}
	return ids
}

// linkParents points every entry at the entry of its parent, when the parent
// is included in this export, and cuts parent cycles (broken data) at the
// cycle member with the smallest ID, so every chain ends.
func linkParents(entries []*entry, rollUp bool) {
	if !rollUp {
		return
	}
	byID := make(map[uuid.UUID]*entry, len(entries))
	for _, e := range entries {
		byID[e.it.ID] = e
	}
	for _, e := range entries {
		if p := byID[e.it.ParentID]; p != nil {
			e.parent = p
			p.isParent = true
		}
	}

	const onPath, done = 1, 2
	state := make(map[*entry]int, len(entries))
	for _, e := range entries {
		var path []*entry
		cur := e
		for cur != nil && state[cur] == 0 {
			state[cur] = onPath
			path = append(path, cur)
			cur = cur.parent
		}
		if cur != nil && state[cur] == onPath {
			cycle := path[slices.Index(path, cur):]
			cut := slices.MinFunc(cycle, func(a, b *entry) int { return strings.Compare(a.it.ID.String(), b.it.ID.String()) })
			cut.parent.isParent = slices.ContainsFunc(entries, func(x *entry) bool { return x != cut && x.parent == cut.parent })
			cut.parent = nil
		}
		for _, p := range path {
			state[p] = done
		}
	}
}

// effectiveTags returns the item's own tags followed by the tags of every
// ancestor along its included parents, deduplicated (FR-35). Without
// RollUpContents no entry has a parent, so these are the item's own tags.
func effectiveTags(e *entry) []uuid.UUID {
	tags := slices.Clone(e.it.TagIDs)
	seen := map[*entry]bool{e: true}
	for a := e.parent; a != nil && !seen[a]; a = a.parent {
		seen[a] = true
		for _, t := range a.it.TagIDs {
			if !slices.Contains(tags, t) {
				tags = append(tags, t)
			}
		}
	}
	return tags
}

// assignOwners rolls each entry into the outermost container it reaches by
// stepping to parents with the same customs status and quantity exactly 1,
// never skipping one (FR-30). A container with another quantity is never
// absorbed itself.
func assignOwners(entries []*entry) {
	for _, e := range entries {
		if e.isParent && e.it.Quantity != 1 {
			continue
		}
		var owner *entry
		seen := map[*entry]bool{e: true}
		for a := e.parent; a != nil && !seen[a] && a.class.Status == e.class.Status && a.it.Quantity == 1; a = a.parent {
			seen[a] = true
			owner = a
		}
		if owner != nil {
			e.owner = owner
			owner.contents = append(owner.contents, e)
		}
	}
	for _, e := range entries {
		slices.SortFunc(e.contents, compareEntries)
	}
}

// computeValues sets each entry's line unit value: its own price per piece
// plus the value of its contents (FR-31).
func computeValues(entries []*entry) {
	for _, e := range entries {
		e.lineUnitCents = e.unitCents
		for _, c := range e.contents {
			e.lineUnitCents += LineCents(c.it.Quantity, c.unitCents)
		}
		if e.owner == nil {
			e.hasPrice = e.lineUnitCents != 0
		} else {
			e.hasPrice = e.unitCents != 0
		}
	}
}

func contentsLabel(qty float64) string {
	return "inkl. Inhalt (" + FormatQtyDE(qty) + " Stk.)"
}
