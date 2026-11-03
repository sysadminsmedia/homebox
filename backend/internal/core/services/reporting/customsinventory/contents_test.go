package customsinventory

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	boxID, bookID, vaseID, pouchID, penID = id(1001), id(1002), id(1003), id(1004), id(1005)
)

func contentItem(n int, name string, qty, price float64, parent uuid.UUID) Item {
	it := furniture(n, int64(n-1000), name, qty, price, "2019-01-01", locLiving, pathLiving)
	it.ParentID = parent
	return it
}

// containerInput: a duty-free box (own price 0) that holds books, a vase and a pouch with pens.
// Only the box is in scope. Contents value: 3×10.00 + 25.50 + 5.00 + 2×1.25 = 63.00, 7 pieces.
func containerInput() Input {
	box := contentItem(1001, "Box books", 1, 0, locLiving)
	in := smallInput(box)
	in.Items = append(in.Items,
		contentItem(1002, "Book", 3, 10, boxID), contentItem(1003, "Vase", 1, 25.50, boxID),
		contentItem(1004, "Pouch", 1, 5, boxID), contentItem(1005, "Pen", 2, 1.25, pouchID))
	in.Options.RollUpContents = true
	return in
}

func hasComment(l Line, s string) bool {
	return slices.ContainsFunc(l.Comments, func(c string) bool { return strings.Contains(c, s) })
}

func TestBuild_RollUp_ListsOnlyTopContainer(t *testing.T) {
	r := mustBuild(t, containerInput())
	lines := allLines(r)
	require.Len(t, lines, 1)
	l := lines[0]
	assert.Equal(t, "Box books", l.Name)
	assert.Equal(t, int64(6300), l.UnitCents)
	assert.Equal(t, int64(6300), l.LineCents)
	assert.True(t, l.HasPrice)
	assert.Equal(t, 5, r.ItemCount)
	assert.Equal(t, 4, r.Preview().ContentCount)
	assert.Contains(t, l.Comments, "inkl. Inhalt (7 Stk.)")
	assert.ElementsMatch(t, []uuid.UUID{boxID, bookID, vaseID, pouchID, penID}, r.ScopeIDs)
	assert.Equal(t, []int64{0}, l.MemberOwnCents)
}

func TestBuild_RollUp_ParentPriceAdded(t *testing.T) {
	in := containerInput()
	setItem(&in, boxID, func(it *Item) { it.PurchasePrice = 12 })
	assert.Equal(t, int64(7500), allLines(mustBuild(t, in))[0].UnitCents)
}

func TestBuild_RollUp_ChildrenInScopeNotListedTwice(t *testing.T) {
	in := containerInput()
	in.ScopeIDs = []uuid.UUID{boxID, bookID, penID}
	r := mustBuild(t, in)
	assert.Len(t, allLines(r), 1)
	assert.NoError(t, Verify(r))
}

func TestBuild_RollUp_Off(t *testing.T) {
	in := containerInput()
	in.Options.RollUpContents = false
	in.ScopeIDs = []uuid.UUID{boxID, bookID}
	r := mustBuild(t, in)
	lines := allLines(r)
	assert.Len(t, lines, 2)
	assert.Len(t, r.ScopeIDs, 2)
	for _, l := range lines {
		assert.False(t, hasComment(l, "inkl. Inhalt"), l.Name)
		assert.False(t, hasComment(l, "Inhalt von"), l.Name)
	}
}

func TestBuild_RollUp_DutiableChildSplitOut(t *testing.T) {
	in := containerInput()
	setItem(&in, vaseID, func(it *Item) { it.PurchaseDate = d("2026-10-01") })
	r := mustBuild(t, in)
	assert.Equal(t, int64(3750), lineByName(t, r, "Box books").UnitCents)
	require.Len(t, r.Dutiable, 1)
	assert.Equal(t, "Vase", r.Dutiable[0].Name)
	assert.True(t, hasComment(r.Dutiable[0], "Inhalt von 000-001"))
	assert.Equal(t, []string{"Vase"}, issueNames(r, IssueContentSplit))
}

func TestBuild_RollUp_ContainerQuantityNotOne(t *testing.T) {
	in := containerInput()
	setItem(&in, boxID, func(it *Item) { it.Quantity = 2 })
	r := mustBuild(t, in)
	lines := allLines(r)
	require.Len(t, lines, 4)
	assert.Equal(t, int64(750), lineByName(t, r, "Pouch").UnitCents)
	assert.True(t, hasComment(lineByName(t, r, "Book"), "Inhalt von 000-001"))
	assert.Equal(t, []string{"Box books"}, issueNames(r, IssueContainerQuantity))
	assert.Empty(t, issueNames(r, IssueContentSplit))
}

func TestBuild_RollUp_QuantityParentStopsWalk(t *testing.T) {
	in := containerInput()
	setItem(&in, pouchID, func(it *Item) { it.Quantity = 2 })
	r := mustBuild(t, in)
	require.Len(t, allLines(r), 3)
	assert.Equal(t, int64(5550), lineByName(t, r, "Box books").UnitCents)
	assert.True(t, hasComment(lineByName(t, r, "Pouch"), "Inhalt von 000-001"))
	assert.True(t, hasComment(lineByName(t, r, "Pen"), "Inhalt von 000-004"))
	assert.Equal(t, []string{"Pouch"}, issueNames(r, IssueContainerQuantity))
	assert.Empty(t, issueNames(r, IssueContentSplit))
}

func TestBuild_RollUp_DifferentStatusParentStopsWalk(t *testing.T) {
	in := containerInput()
	bag := contentItem(1006, "Bag", 1, 40, boxID)
	bag.PurchaseDate = d("2026-10-01")
	scarf := contentItem(1007, "Scarf", 1, 15, bag.ID)
	in.Items = append(in.Items, bag, scarf)
	r := mustBuild(t, in)

	assert.Equal(t, int64(6300), lineByName(t, r, "Box books").UnitCents)
	bagLine := lineByName(t, r, "Bag")
	assert.Equal(t, StatusDutiable, bagLine.Status)
	assert.True(t, hasComment(bagLine, "Inhalt von 000-001"))
	scarfLine := lineByName(t, r, "Scarf")
	assert.Equal(t, StatusDutyFree, scarfLine.Status)
	assert.Equal(t, int64(1500), scarfLine.UnitCents)
	assert.True(t, hasComment(scarfLine, "Inhalt von 000-006"))
	assert.Equal(t, []string{"Bag", "Scarf"}, issueNames(r, IssueContentSplit))
	assert.NoError(t, Verify(r))
}

func TestBuild_RollUp_MissingPrice(t *testing.T) {
	in := containerInput()
	setItem(&in, bookID, func(it *Item) { it.PurchasePrice = 0 })
	r := mustBuild(t, in)
	assert.Equal(t, []string{"Book"}, issueNames(r, IssueMissingPrice))
	assert.True(t, hasComment(lineByName(t, r, "Box books"), "Wert fehlt"))
}

func TestBuild_RollUp_ReceiptsOfContents(t *testing.T) {
	in := containerInput()
	in.Options.IncludeReceipts = true
	setItem(&in, vaseID, func(it *Item) {
		it.Attachments = []Attachment{receipt(1, "r.pdf", "application/pdf", "2024-01-01T00:00:00Z")}
	})
	r := mustBuild(t, in)
	want := ReceiptName("000-003", 0, 0, "r.pdf", "application/pdf")
	assert.Equal(t, []string{want}, lineByName(t, r, "Box books").ReceiptNames)
	require.Len(t, r.Receipts, 1)
	assert.Equal(t, want, r.Receipts[0].Name)
	assert.Equal(t, vaseID, r.Receipts[0].ItemID)
}

func TestBuild_RollUp_SoldContentExcluded(t *testing.T) {
	in := containerInput()
	setItem(&in, bookID, func(it *Item) { it.SoldDate = d("2025-01-01") })
	r := mustBuild(t, in)
	require.Len(t, r.Excluded, 1)
	assert.Equal(t, bookID, r.Excluded[0].ID)
	assert.Equal(t, IssueExcludedSold, r.Excluded[0].Code)
	assert.Equal(t, int64(3300), lineByName(t, r, "Box books").UnitCents)
}

func TestBuild_RollUp_ContentUnderArchivedIgnored(t *testing.T) {
	in := containerInput()
	setItem(&in, pouchID, func(it *Item) { it.Archived = true })
	r := mustBuild(t, in)
	require.Len(t, r.Excluded, 1)
	assert.Equal(t, pouchID, r.Excluded[0].ID)
	assert.Equal(t, IssueExcludedArchived, r.Excluded[0].Code)
	assert.NotContains(t, r.ScopeIDs, penID)
	assert.Equal(t, int64(5550), lineByName(t, r, "Box books").UnitCents)
}

func TestBuild_RollUp_ParentCycleTerminates(t *testing.T) {
	in := containerInput()
	setItem(&in, boxID, func(it *Item) { it.ParentID = penID })
	r, err := Build(in)
	require.NoError(t, err)
	lines := allLines(r)
	require.Len(t, lines, 1)
	assert.Equal(t, "Box books", lines[0].Name)
	assert.Equal(t, int64(6300), lines[0].UnitCents)
}

func TestBuild_RollUp_ScopeChildSplitOnce(t *testing.T) {
	in := containerInput()
	setItem(&in, vaseID, func(it *Item) { it.PurchaseDate = d("2026-10-01") })
	in.ScopeIDs = append(in.ScopeIDs, vaseID)
	r := mustBuild(t, in)
	n := 0
	for _, l := range allLines(r) {
		if slices.Contains(l.MemberIDs, vaseID) {
			n++
		}
	}
	assert.Equal(t, 1, n)
}

func TestBuild_RollUp_MergedContainersKeepTheirContents(t *testing.T) {
	in := containerInput()
	box2 := contentItem(1009, "Box books", 1, 0, locLiving)
	in.Items = append(in.Items, box2, contentItem(1010, "Plate", 1, 63, box2.ID))
	in.ScopeIDs = append(in.ScopeIDs, box2.ID)
	r := mustBuild(t, in)
	lines := allLines(r)
	require.Len(t, lines, 1)
	l := lines[0]
	assert.InDelta(t, 2.0, l.Quantity, 0)
	assert.Equal(t, int64(6300), l.UnitCents)
	assert.Equal(t, int64(12600), l.LineCents)
	owners := map[uuid.UUID]bool{}
	for _, c := range l.Contents {
		owners[c.OwnerID] = true
	}
	assert.Equal(t, map[uuid.UUID]bool{boxID: true, box2.ID: true}, owners)
	assert.NoError(t, Verify(r))
}
