package customsinventory

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerify_AcceptsBuiltReport(t *testing.T) {
	assert.NoError(t, Verify(mustBuild(t, goldenInput())))
}

func TestVerify_RejectsCorruption(t *testing.T) {
	cases := map[string]func(r *Report){
		"invariant (a)": func(r *Report) {
			l := &r.Groups[0].Lines[1]
			l.MemberIDs = append(l.MemberIDs, r.Groups[0].Lines[0].MemberIDs[0])
			l.MemberQuantities = append(l.MemberQuantities, 0)
		},
		"invariant (b)": func(r *Report) { r.Groups[0].Lines[0].LineCents++ },
		"invariant (c)": func(r *Report) { r.Groups[0].SubtotalCents++ },
		"invariant (d)": func(r *Report) { r.DutyFreeTotalCents++ },
		"invariant (e)": func(r *Report) { r.GrandTotalCents++ },
		"invariant (f)": func(r *Report) { r.Receipts = r.Receipts[1:] },
		"invariant (g)": func(r *Report) { r.Receipts = append(r.Receipts, r.Receipts[0]) },
		"invariant (h)": func(r *Report) { r.Groups[0].SubtotalQty++ },
	}
	for want, mutate := range cases {
		r := mustBuild(t, goldenInput())
		mutate(&r)
		err := Verify(r)
		require.Error(t, err, want)
		assert.Contains(t, err.Error(), want)
	}
}

func TestVerify_RejectsWrongTotalQuantity(t *testing.T) {
	r := mustBuild(t, goldenInput())
	r.TotalQty++
	err := Verify(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invariant (h)")
}

func TestVerify_RejectsBrokenContents(t *testing.T) {
	cases := map[string]func(r *Report){
		"invariant (i)": func(r *Report) {
			l := &r.Groups[0].Lines[0]
			l.UnitCents++
			l.LineCents = LineCents(l.Quantity, l.UnitCents)
		},
		"invariant (a)": func(r *Report) {
			l := &r.Groups[0].Lines[0]
			l.Contents = append(l.Contents, l.Contents[0])
		},
	}
	for want, mutate := range cases {
		r := mustBuild(t, containerInput())
		mutate(&r)
		err := Verify(r)
		require.Error(t, err, want)
		assert.Contains(t, err.Error(), want)
	}

	r := mustBuild(t, containerInput())
	r.Groups[0].Lines[0].Contents[0].OwnerID = uuid.New()
	err := Verify(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invariant (i)")
}
