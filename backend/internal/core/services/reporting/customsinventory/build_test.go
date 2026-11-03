package customsinventory

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "rewrite testdata/golden_report.json")

func TestBuild_EveryScopeItemOnceOrExcluded(t *testing.T) {
	in := goldenInput()
	r := mustBuild(t, in)

	seen := map[uuid.UUID]int{}
	for _, l := range allLines(r) {
		for _, m := range l.MemberIDs {
			seen[m]++
		}
	}
	for _, e := range r.Excluded {
		seen[e.ID]++
	}
	assert.Len(t, seen, len(in.ScopeIDs))
	for _, sid := range in.ScopeIDs {
		assert.Equal(t, 1, seen[sid], "scope ID %s", sid)
	}

	codes := map[uuid.UUID]IssueCode{}
	for _, e := range r.Excluded {
		codes[e.ID] = e.Code
	}
	assert.Equal(t, map[uuid.UUID]IssueCode{
		id(180):   IssueExcludedLocation,
		id(181):   IssueExcludedArchived,
		id(182):   IssueExcludedSold,
		missingID: IssueExcludedNotFound,
	}, codes)
}

func TestBuild_Totals(t *testing.T) {
	r := mustBuild(t, goldenInput())
	// Duty-free: 120000 + 89000 + 35800 + 83 + 0 + 500 + 3000 = 248383
	// Dutiable: 64990 + 0 + 7900 + 4550 + 0 + 0 = 77440
	assert.Equal(t, int64(248383), r.DutyFreeTotalCents)
	assert.Equal(t, int64(77440), r.DutiableTotalCents)
	assert.Equal(t, int64(325823), r.GrandTotalCents)
	assert.Equal(t, 26, r.ItemCount) // 30 in scope, 4 excluded
}

func TestBuild_LineValueIsQuantityTimesUnitPrice(t *testing.T) {
	r := mustBuild(t, goldenInput())
	chair := lineByName(t, r, "Dining chair")
	assert.InDelta(t, 4.0, chair.Quantity, 0)
	assert.Equal(t, int64(8950), chair.UnitCents)
	assert.Equal(t, int64(35800), chair.LineCents)

	require.NotEmpty(t, r.Groups)
	assert.Equal(t, "Home › Living room", r.Groups[0].Title)
	assert.Equal(t, int64(120000+89000+35800), r.Groups[0].SubtotalCents)
}

func TestBuild_MergeIdentical(t *testing.T) {
	r := mustBuild(t, goldenInput())
	var files []Line
	for _, l := range allLines(r) {
		if l.Name == nameFiles {
			files = append(files, l)
		}
	}
	require.Len(t, files, 2)

	merged := files[0]
	assert.InDelta(t, 14.0, merged.Quantity, 0)
	assert.False(t, merged.HasPrice)
	assert.Len(t, merged.MemberIDs, 14)
	want := make([]string, 0, 14)
	for n := int64(131); n <= 144; n++ {
		want = append(want, FormatAssetID(n))
	}
	assert.Equal(t, want, merged.AssetIDs)
	assert.Equal(t, []string{"zollfrei lt. Tag „owned > 2 years“; Wert fehlt"}, merged.Comments)

	assert.Equal(t, []string{"000-145"}, files[1].AssetIDs)
	assert.Equal(t, int64(500), files[1].UnitCents)
	assert.True(t, files[1].HasPrice)
}

func TestBuild_MergeOffKeepsLines(t *testing.T) {
	in := goldenInput()
	in.Options.MergeIdentical = false
	r := mustBuild(t, in)
	n := 0
	for _, l := range allLines(r) {
		if l.Name == nameFiles || l.Name == "files" {
			n++
		}
	}
	assert.Equal(t, 15, n)
}

func TestBuild_MergedDateRange(t *testing.T) {
	a := furniture(1, 1, "Chair", 1, 10, "2021-06-30", locLiving, pathLiving)
	b := furniture(2, 2, "chair ", 1, 10, "2019-01-01", locLiving, pathLiving)
	r := mustBuild(t, smallInput(a, b))
	require.Len(t, r.Groups, 1)
	require.Len(t, r.Groups[0].Lines, 1)
	l := r.Groups[0].Lines[0]
	assert.Equal(t, "2019-01-01 – 2021-06-30", l.PurchaseDateText)
	assert.True(t, l.PurchaseDate.Time().IsZero())

	c := furniture(3, 3, "Chair", 1, 10, "2019-01-01", locLiving, pathLiving)
	same := mustBuild(t, smallInput(b, c)).Groups[0].Lines[0]
	assert.Empty(t, same.PurchaseDateText)
	assert.Equal(t, "2019-01-01", same.PurchaseDate.String())
}

func TestBuild_MergedValueUsesSummedQuantity(t *testing.T) {
	var items []Item
	for i := 1; i <= 3; i++ {
		items = append(items, furniture(i, int64(i), "Cable", 0.5, 0.33, "2020-01-01", locOffice, pathOffice))
	}
	r := mustBuild(t, smallInput(items...))
	require.Len(t, r.Groups, 1)
	require.Len(t, r.Groups[0].Lines, 1)
	l := r.Groups[0].Lines[0]
	assert.InDelta(t, 1.5, l.Quantity, 0)
	assert.Equal(t, int64(33), l.UnitCents)
	assert.Equal(t, int64(50), l.LineCents, "summed quantity × unit price, not a sum of rounded member values (51)")
	assert.Equal(t, int64(50), r.GrandTotalCents)
}

func groupTitles(r Report) []string {
	titles := make([]string, len(r.Groups))
	for i, g := range r.Groups {
		titles[i] = g.Title
	}
	return titles
}

func TestBuild_GroupOrder(t *testing.T) {
	r := mustBuild(t, goldenInput())
	assert.Equal(t, []string{"Home › Living room", "Home › Office", LabelNoLocation}, groupTitles(r))

	in := goldenInput()
	in.Options.GroupBy = GroupByTag
	in.Options.GroupTagIDs = []uuid.UUID{tagB, tagA}
	setItem(&in, id(101), func(it *Item) { it.TagIDs = []uuid.UUID{tagA} })
	setItem(&in, id(102), func(it *Item) { it.TagIDs = []uuid.UUID{tagB} })
	assert.Equal(t, []string{"Group B", "Group A", LabelOther}, groupTitles(mustBuild(t, in)))

	in = goldenInput()
	in.Options.GroupBy = GroupByEntityType
	assert.Equal(t, []string{"Box", typeNameFurniture, "Tool"}, groupTitles(mustBuild(t, in)))
}

func TestBuild_TagGroupingFirstMatchWins(t *testing.T) {
	in := goldenInput()
	in.Options.GroupBy = GroupByTag
	in.Options.GroupTagIDs = []uuid.UUID{tagA, tagB}
	setItem(&in, id(103), func(it *Item) { it.TagIDs = []uuid.UUID{tagB, tagA1} })
	r := mustBuild(t, in)

	var hits []string
	for _, g := range r.Groups {
		for _, l := range g.Lines {
			if l.Name == "Dining chair" {
				hits = append(hits, g.Title)
			}
		}
	}
	assert.Equal(t, []string{"Group A"}, hits)
}

func TestBuild_PositionsContinuous(t *testing.T) {
	r := mustBuild(t, goldenInput())
	lines := allLines(r)
	for i, l := range lines {
		assert.Equal(t, i+1, l.Pos, l.Name)
	}
	assert.Len(t, lines, 13)
}

func TestBuild_DutiableSection(t *testing.T) {
	r := mustBuild(t, goldenInput())
	names := make([]string, len(r.Dutiable))
	for i, l := range r.Dutiable {
		names[i] = l.Name
		assert.Equal(t, StatusDutiable, l.Status, l.Name)
	}
	// group order (Garage, Kitchen, Living room, Office), then line order
	assert.Equal(t, []string{"Bicycle", "Tent", nameEspresso, "Desk lamp", "Floor lamp", "Office chair"}, names)

	espresso := lineByName(t, r, nameEspresso)
	assert.Equal(t, []string{"Kauf nach Stichtag (< 6 Monate)"}, espresso.Comments)
	assert.Equal(t, []string{"Home › Kitchen"}, espresso.Locations)
	assert.Equal(t, []string{"Kaufdatum fehlt; Wert fehlt"}, lineByName(t, r, "Office chair").Comments)
	assert.Equal(t, []string{"zu verzollen lt. Tag „dutiable“; Wert fehlt"}, lineByName(t, r, "Bicycle").Comments)
	assert.Equal(t, []string{"Kauf nach Referenzdatum; Wert fehlt"}, lineByName(t, r, "Tent").Comments)
}

func TestBuild_MinMonthsChangesCutoff(t *testing.T) {
	in := goldenInput()
	in.Options.MinMonths = 1
	r := mustBuild(t, in)
	assert.Equal(t, "2026-10-15", r.Cutoff.String())
	for _, name := range []string{nameEspresso, "Floor lamp", "Desk lamp"} {
		assert.Equal(t, StatusDutyFree, lineByName(t, r, name).Status, name)
	}
}

func TestBuild_ReceiptNames(t *testing.T) {
	r := mustBuild(t, goldenInput())
	assert.Equal(t, []string{"Belege/000-101_Rechnung.pdf"}, lineByName(t, r, "Grandfather clock").ReceiptNames)
	assert.Equal(t, []string{"Belege/000-146_Rechnung.pdf", "Belege/000-146_Rechnung_2.jpg"}, lineByName(t, r, nameEspresso).ReceiptNames)
	assert.Equal(t, []string{"Belege/000-163_Rechnung.pdf"}, lineByName(t, r, "Floor lamp").ReceiptNames, "lower internal ID")
	assert.Equal(t, []string{"Belege/000-163_b_Rechnung.pdf"}, lineByName(t, r, "Desk lamp").ReceiptNames)
	assert.Empty(t, lineByName(t, r, "Sofa").ReceiptNames, "links and photos are never bundled")

	var names []string
	for _, rf := range r.Receipts {
		names = append(names, rf.Name)
		assert.NotEqual(t, id(5102), rf.AttachmentID)
		assert.NotEqual(t, id(5103), rf.AttachmentID)
	}
	assert.Len(t, names, 5)
	assert.Contains(t, issueNames(r, IssueReceiptIsLink), "Sofa")

	in := goldenInput()
	in.Options.IncludeReceipts = false
	off := mustBuild(t, in)
	assert.Empty(t, off.Receipts)
	for _, l := range allLines(off) {
		assert.Empty(t, l.ReceiptNames, l.Name)
	}
}

func TestBuild_DefaultReceiptTypes(t *testing.T) {
	in := goldenInput()
	in.Options.ReceiptTypes = nil
	r := mustBuild(t, in)
	assert.Equal(t, []string{"Belege/000-146_Rechnung.pdf"}, lineByName(t, r, nameEspresso).ReceiptNames)
}

func TestBuild_Issues(t *testing.T) {
	r := mustBuild(t, goldenInput())

	var codes []IssueCode
	for _, is := range r.Issues {
		codes = append(codes, is.Code)
		assert.Equal(t, SeverityOf(is.Code), is.Severity, is.Code)
	}
	var want []IssueCode
	for _, c := range IssueOrder {
		if c != IssueNegativeValue && c != IssueZeroQuantity {
			want = append(want, c)
		}
	}
	assert.Equal(t, want, codes, "every code except negative_value and zero_quantity, in order")

	assert.Equal(t, []string{"Office chair"}, issueNames(r, IssueMissingPurchaseDate))
	assert.Equal(t, []string{nameEspresso}, issueNames(r, IssueConflictDateTag))
	assert.Equal(t, []string{"Bicycle"}, issueNames(r, IssueConflictTags))
	assert.Equal(t, []string{"Tent"}, issueNames(r, IssuePurchasedAfterReference))
	assert.Equal(t, []string{"Toolbox"}, issueNames(r, IssueMissingAssetID))
	assert.Equal(t, []string{"Desk lamp", "Floor lamp"}, issueNames(r, IssueDuplicateAssetID))
	assert.Equal(t, []string{"Bicycle", "Office chair", "Tent"}, issueNames(r, IssueDutiableWithoutReceipt))
	assert.Equal(t, []string{"Sofa"}, issueNames(r, IssueReceiptIsLink))
	assert.Equal(t, []string{"Basement"}, issueNames(r, IssueExcludedLocation))
	assert.Equal(t, []string{"Old radio"}, issueNames(r, IssueExcludedArchived))
	assert.Equal(t, []string{"Bookshelf"}, issueNames(r, IssueExcludedSold))
	assert.Equal(t, []string{""}, issueNames(r, IssueExcludedNotFound), "not found → no data")

	missingPrice := issueNames(r, IssueMissingPrice)
	assert.Len(t, missingPrice, 17) // 14 files + office chair + bicycle + tent
	assert.True(t, slices.IsSortedFunc(missingPrice, compareFold))
	for _, is := range r.Issues {
		if is.Code == IssueMissingPrice {
			assert.Equal(t, "000-131", is.Items[1].AssetID, "same name → asset ID order")
		}
	}
}

func TestBuild_NegativeAndZero(t *testing.T) {
	zero := furniture(1, 1, "Lamp", 0, 10.00, "2020-01-01", locLiving, pathLiving)
	negative := furniture(2, 2, "Refund", 1, -10.00, "2020-01-01", locLiving, pathLiving)
	r := mustBuild(t, smallInput(zero, negative))

	assert.Equal(t, []string{"Lamp"}, issueNames(r, IssueZeroQuantity))
	assert.Equal(t, []string{"Refund"}, issueNames(r, IssueNegativeValue))
	assert.Equal(t, int64(0), lineByName(t, r, "Lamp").LineCents)
	assert.Equal(t, int64(-1000), lineByName(t, r, "Refund").LineCents)
	assert.Equal(t, int64(-1000), r.GrandTotalCents)
}

func TestBuild_UnknownTagIDsIgnored(t *testing.T) {
	unknown := id(12345)

	in := goldenInput()
	in.Options.DutyFreeTagIDs = []uuid.UUID{tagP, unknown}
	in.Options.DutiableTagIDs = []uuid.UUID{tagZ, unknown}
	same := mustBuild(t, in)
	golden := mustBuild(t, goldenInput())
	assert.Equal(t, golden.DutyFreeTotalCents, same.DutyFreeTotalCents)
	assert.Equal(t, golden.DutiableTotalCents, same.DutiableTotalCents)

	in = goldenInput()
	in.Options.DutyFreeTagIDs = []uuid.UUID{unknown}
	r := mustBuild(t, in)
	assert.Equal(t, StatusDutiable, lineByName(t, r, nameFiles).Status)
	assert.Contains(t, issueNames(r, IssueMissingPurchaseDate), nameFiles)
}

func TestBuild_DeduplicatesScopeIDs(t *testing.T) {
	in := goldenInput()
	in.ScopeIDs = append(in.ScopeIDs, id(101), id(101))
	r := mustBuild(t, in)
	assert.Equal(t, int64(325823), r.GrandTotalCents)
	assert.Len(t, r.ScopeIDs, len(goldenInput().ScopeIDs))
}

func TestBuild_InvalidReferenceDate(t *testing.T) {
	in := goldenInput()
	in.Options.ReferenceDate = "15.11.2026"
	_, err := Build(in)
	assert.ErrorIs(t, err, ErrInvalidRequest)
}

func TestBuild_Golden(t *testing.T) {
	got, err := json.MarshalIndent(mustBuild(t, goldenInput()), "", "  ")
	require.NoError(t, err)
	got = append(got, '\n')

	path := filepath.Join("testdata", "golden_report.json")
	if *update {
		require.NoError(t, os.MkdirAll("testdata", 0o755))
		require.NoError(t, os.WriteFile(path, got, 0o600))
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err, "run with -update to create the golden file")
	assert.JSONEq(t, string(want), string(got))
}

func TestPreview_MatchesReport(t *testing.T) {
	r := mustBuild(t, goldenInput())
	p := r.Preview()

	assert.Equal(t, "EUR", p.Currency)
	assert.Equal(t, "2026-11-15", p.ReferenceDate)
	assert.Equal(t, 6, p.MinMonths)
	assert.Equal(t, "2026-05-15", p.Cutoff)
	assert.Equal(t, r.ItemCount, p.ItemCount)
	assert.Equal(t, 13, p.LineCount)
	assert.Equal(t, 6, p.DutiableItemCount)
	assert.Equal(t, 4, p.ExcludedCount)
	assert.Equal(t, 5, p.ReceiptCount)
	assert.Equal(t, r.DutyFreeTotalCents, p.DutyFreeTotalCents)
	assert.Equal(t, r.DutiableTotalCents, p.DutiableTotalCents)
	assert.Equal(t, r.GrandTotalCents, p.GrandTotalCents)
	assert.Equal(t, r.Issues, p.Issues)
}

func TestBuild_PieceCounts(t *testing.T) {
	in := smallInput(
		furniture(1, 1, "Box books", 3, 10, "2019-01-01", locLiving, pathLiving),
		furniture(2, 2, "Cable", 2.5, 1, "2019-01-01", locLiving, pathLiving),
		furniture(3, 3, "Rope", 0.3, 1, "2019-01-01", locLiving, pathLiving),
		furniture(4, 4, "New lamp", 4, 20, "2026-10-01", locLiving, pathLiving), // dutiable
	)
	r := mustBuild(t, in)
	assert.InDelta(t, 5.8, r.DutyFreeQty, 0)
	assert.InDelta(t, 5.8, r.Groups[0].SubtotalQty, 0)
	assert.InDelta(t, 4.0, r.DutiableQty, 0)
	assert.InDelta(t, 9.8, r.TotalQty, 0)
	p := r.Preview()
	assert.InDelta(t, 9.8, p.PieceCount, 0)
	assert.InDelta(t, 4.0, p.DutiablePieceCount, 0)
}
