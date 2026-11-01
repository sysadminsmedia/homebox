package customsinventory

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

var (
	tagP  = id(901) // "owned"
	tagP2 = id(902) // "owned > 2 years", child of P
	tagZ  = id(903) // "dutiable"
)

func classifyRules(minMonths int) Rules {
	ti := NewTagIndex([]Tag{
		{ID: tagP, Name: "owned"},
		{ID: tagP2, ParentID: tagP, Name: "owned > 2 years"},
		{ID: tagZ, Name: "dutiable"},
	})
	ref := d("2026-11-15")
	return Rules{
		Reference: ref,
		Cutoff:    Cutoff(ref, minMonths),
		MinMonths: minMonths,
		DutyFree:  ti.Expand([]uuid.UUID{tagP}),
		Dutiable:  ti.Expand([]uuid.UUID{tagZ}),
		Tags:      ti,
	}
}

func dateOrZero(s string) types.Date {
	if s == "" {
		return types.Date{}
	}
	return d(s)
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		date string
		tags []uuid.UUID
		want Classification
	}{
		{"dutiable tag wins", "2019-03-12", []uuid.UUID{tagZ}, Classification{StatusDutiable, ReasonTagDutiable, "dutiable", nil}},
		{"both tags", "", []uuid.UUID{tagZ, tagP2}, Classification{StatusDutiable, ReasonTagDutiable, "dutiable", []IssueCode{IssueConflictTags}}},
		{"after reference date", "2026-11-16", nil, Classification{StatusDutiable, ReasonPurchasedAfterReference, "", []IssueCode{IssuePurchasedAfterReference}}},
		{"reference day itself", "2026-11-15", nil, Classification{StatusDutiable, ReasonOwnedTooShort, "", nil}},
		{"day after cutoff", "2026-05-16", nil, Classification{StatusDutiable, ReasonOwnedTooShort, "", nil}},
		{"date beats tag", "2026-05-16", []uuid.UUID{tagP2}, Classification{StatusDutiable, ReasonOwnedTooShort, "", []IssueCode{IssueConflictDateTag}}},
		{"on cutoff", "2026-05-15", nil, Classification{StatusDutyFree, ReasonDate, "", nil}},
		{"old + tag, no conflict", "2019-03-12", []uuid.UUID{tagP2}, Classification{StatusDutyFree, ReasonDate, "", nil}},
		{"no date, child of duty-free tag", "", []uuid.UUID{tagP2}, Classification{StatusDutyFree, ReasonTagDutyFree, "owned > 2 years", nil}},
		{"no date, no tag", "", nil, Classification{StatusDutiable, ReasonMissingPurchaseDate, "", []IssueCode{IssueMissingPurchaseDate}}},
	}
	r := classifyRules(6)
	for _, tc := range cases {
		got := Classify(Item{PurchaseDate: dateOrZero(tc.date), TagIDs: tc.tags}, r)
		assert.Equal(t, tc.want, got, tc.name)
	}

	r12 := classifyRules(12)
	assert.Equal(t, ReasonOwnedTooShort, Classify(Item{PurchaseDate: d("2026-01-10")}, r12).Reason)
	assert.Equal(t, ReasonDate, Classify(Item{PurchaseDate: d("2025-11-15")}, r12).Reason)
}

func TestReasonComment(t *testing.T) {
	cases := []struct {
		c      Classification
		months int
		want   string
	}{
		{Classification{Reason: ReasonOwnedTooShort}, 6, "Kauf nach Stichtag (< 6 Monate)"},
		{Classification{Reason: ReasonOwnedTooShort}, 12, "Kauf nach Stichtag (< 12 Monate)"},
		{Classification{Reason: ReasonMissingPurchaseDate}, 6, "Kaufdatum fehlt"},
		{Classification{Reason: ReasonTagDutiable, ReasonTag: "dutiable"}, 6, "zu verzollen lt. Tag „dutiable“"},
		{Classification{Reason: ReasonPurchasedAfterReference}, 6, "Kauf nach Referenzdatum"},
		{Classification{Reason: ReasonTagDutyFree, ReasonTag: "owned > 2 years"}, 6, "zollfrei lt. Tag „owned > 2 years“"},
		{Classification{Reason: ReasonDate}, 6, ""},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, ReasonComment(tc.c, tc.months))
	}
}
