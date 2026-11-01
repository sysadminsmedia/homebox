package customsinventory

import (
	"testing"

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
	}
	for want, mutate := range cases {
		r := mustBuild(t, goldenInput())
		mutate(&r)
		err := Verify(r)
		require.Error(t, err, want)
		assert.Contains(t, err.Error(), want)
	}
}
