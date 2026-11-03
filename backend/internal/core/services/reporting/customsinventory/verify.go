package customsinventory

import (
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// ErrSelfCheck marks a report that failed its own invariants (NFR-1).
var ErrSelfCheck = errors.New("customs inventory: self-check failed")

func invariant(letter, format string, args ...any) error {
	return fmt.Errorf("invariant (%s): %s: %w", letter, fmt.Sprintf(format, args...), ErrSelfCheck)
}

// Verify checks the report's invariants (NFR-1 (a)–(i)). The messages carry
// counts and IDs only, never names or prices.
func Verify(r Report) error {
	lines := allReportLines(r)

	// (a) every scope entity is on exactly one line or in exactly one excluded entry
	count := make(map[uuid.UUID]int, len(r.ScopeIDs))
	for _, id := range r.ScopeIDs {
		count[id] = 0
	}
	for _, id := range memberAndExcludedIDs(r, lines) {
		n, inScope := count[id]
		if !inScope {
			return invariant("a", "entity %s is not in scope", id)
		}
		if n > 0 {
			return invariant("a", "entity %s appears more than once", id)
		}
		count[id] = 1
	}
	for id, n := range count {
		if n == 0 {
			return invariant("a", "entity %s is missing", id)
		}
	}

	// (b) member quantities add up and the line value recomputes
	for _, l := range lines {
		if len(l.MemberIDs) != len(l.MemberQuantities) || len(l.MemberIDs) == 0 {
			return invariant("b", "line %d has %d members and %d quantities", l.Pos, len(l.MemberIDs), len(l.MemberQuantities))
		}
		var qty float64
		for _, q := range l.MemberQuantities {
			qty += q
		}
		if RoundQty(qty) != RoundQty(l.Quantity) {
			return invariant("b", "line %d quantity doesn't match its members", l.Pos)
		}
		if l.LineCents != LineCents(l.Quantity, l.UnitCents) {
			return invariant("b", "line %d value doesn't recompute", l.Pos)
		}
	}

	// (i) is checked before the sums, which a wrong rolled-up value would also break.
	if err := verifyContents(lines); err != nil {
		return err
	}

	// (c) group subtotals
	var dutyFree int64
	for i, g := range r.Groups {
		var sum int64
		for _, l := range g.Lines {
			sum += l.LineCents
		}
		if sum != g.SubtotalCents {
			return invariant("c", "group %d subtotal doesn't match its lines", i+1)
		}
		dutyFree += g.SubtotalCents
	}

	// (d) duty-free total
	if dutyFree != r.DutyFreeTotalCents {
		return invariant("d", "duty-free total doesn't match the group subtotals")
	}

	// (e) dutiable total and grand total
	var dutiable int64
	for _, l := range r.Dutiable {
		dutiable += l.LineCents
	}
	if dutiable != r.DutiableTotalCents {
		return invariant("e", "dutiable total doesn't match its lines")
	}
	if r.GrandTotalCents != r.DutyFreeTotalCents+r.DutiableTotalCents {
		return invariant("e", "grand total isn't duty-free + dutiable")
	}

	// (f) receipt names on lines and in the receipt list match
	listed := make(map[string]bool, len(r.Receipts))
	for _, rf := range r.Receipts {
		listed[rf.Name] = true
	}
	used := map[string]bool{}
	for _, l := range lines {
		for _, n := range l.ReceiptNames {
			if !listed[n] {
				return invariant("f", "line %d names a receipt that isn't bundled", l.Pos)
			}
			used[n] = true
		}
	}
	for n := range listed {
		if !used[n] {
			return invariant("f", "a bundled receipt isn't named on any line")
		}
	}

	// (g) receipt names are unique
	unique := make(map[string]bool, len(r.Receipts))
	for _, rf := range r.Receipts {
		if unique[rf.Name] {
			return invariant("g", "receipt name used twice")
		}
		unique[rf.Name] = true
	}

	return verifyQuantities(r)
}

// verifyQuantities checks invariant (h): the quantity sums of the groups, of
// both sections and of the whole sheet match their lines.
func verifyQuantities(r Report) error {
	var dutyFreeQty float64
	for i, g := range r.Groups {
		var sum float64
		for _, l := range g.Lines {
			sum += l.Quantity
		}
		if RoundQty(sum) != RoundQty(g.SubtotalQty) {
			return invariant("h", "group %d quantity doesn't match its lines", i+1)
		}
		dutyFreeQty += g.SubtotalQty
	}
	if RoundQty(dutyFreeQty) != RoundQty(r.DutyFreeQty) {
		return invariant("h", "duty-free quantity doesn't match the groups")
	}
	var dutiableQty float64
	for _, l := range r.Dutiable {
		dutiableQty += l.Quantity
	}
	if RoundQty(dutiableQty) != RoundQty(r.DutiableQty) {
		return invariant("h", "dutiable quantity doesn't match its lines")
	}
	if RoundQty(r.TotalQty) != RoundQty(r.DutyFreeQty+r.DutiableQty) {
		return invariant("h", "total quantity isn't duty-free + dutiable")
	}
	return nil
}

// verifyContents checks invariant (i): each member's unit value is its own
// price per piece plus its contents' values, every member with contents has
// quantity 1, and every content's owner is a member of that line.
func verifyContents(lines []Line) error {
	for _, l := range lines {
		if len(l.MemberOwnCents) != len(l.MemberIDs) {
			return invariant("i", "line %d has %d members and %d own values", l.Pos, len(l.MemberIDs), len(l.MemberOwnCents))
		}
		contentCents := make(map[uuid.UUID]int64, len(l.MemberIDs))
		hasContents := make(map[uuid.UUID]bool, len(l.MemberIDs))
		for _, c := range l.Contents {
			if !slices.Contains(l.MemberIDs, c.OwnerID) {
				return invariant("i", "line %d has content whose owner isn't a member", l.Pos)
			}
			contentCents[c.OwnerID] += LineCents(c.Quantity, c.UnitCents)
			hasContents[c.OwnerID] = true
		}
		for k, id := range l.MemberIDs {
			if l.UnitCents != l.MemberOwnCents[k]+contentCents[id] {
				return invariant("i", "line %d member %d unit value isn't its own value plus its contents", l.Pos, k+1)
			}
			if hasContents[id] && l.MemberQuantities[k] != 1 {
				return invariant("i", "line %d member %d has contents but quantity isn't 1", l.Pos, k+1)
			}
		}
	}
	return nil
}

func memberAndExcludedIDs(r Report, lines []Line) []uuid.UUID {
	var ids []uuid.UUID
	for _, l := range lines {
		ids = append(ids, l.MemberIDs...)
		for _, c := range l.Contents {
			ids = append(ids, c.ID)
		}
	}
	for _, e := range r.Excluded {
		ids = append(ids, e.ID)
	}
	return ids
}
