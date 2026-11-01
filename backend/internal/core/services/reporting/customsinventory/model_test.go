package customsinventory

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validRequest() Request {
	return Request{
		Scope:   Scope{ItemIDs: []uuid.UUID{uuid.New()}},
		Options: Options{GroupBy: GroupByLocation, ReferenceDate: "2026-11-15"},
	}
}

func TestRequestValidate(t *testing.T) {
	require.NoError(t, validRequest().Validate())

	ok := map[string]func(r *Request){
		"nil receipt types": func(r *Request) { r.Options.ReceiptTypes = nil },
		"min months 0":      func(r *Request) { r.Options.MinMonths = 0 },
		"min months 60":     func(r *Request) { r.Options.MinMonths = 60 },
		"query scope":       func(r *Request) { r.Scope = Scope{Query: &ScopeQuery{}} },
		"tag grouping":      func(r *Request) { r.Options.GroupBy = GroupByTag; r.Options.GroupTagIDs = []uuid.UUID{uuid.New()} },
		"all receipt types": func(r *Request) { r.Options.ReceiptTypes = AllowedReceiptTypes },
	}
	for name, mutate := range ok {
		r := validRequest()
		mutate(&r)
		require.NoError(t, r.Validate(), name)
	}

	bad := map[string]func(r *Request){
		"empty reference date":   func(r *Request) { r.Options.ReferenceDate = "" },
		"invalid reference date": func(r *Request) { r.Options.ReferenceDate = "2026-13-01" },
		"min months -1":          func(r *Request) { r.Options.MinMonths = -1 },
		"min months 61":          func(r *Request) { r.Options.MinMonths = 61 },
		"ids and query":          func(r *Request) { r.Scope.Query = &ScopeQuery{} },
		"neither ids nor query":  func(r *Request) { r.Scope = Scope{} },
		"unknown group by":       func(r *Request) { r.Options.GroupBy = "room" },
		"tag without group tags": func(r *Request) { r.Options.GroupBy = GroupByTag },
		"photo receipt type":     func(r *Request) { r.Options.ReceiptTypes = []string{"photo"} },
	}
	for name, mutate := range bad {
		r := validRequest()
		mutate(&r)
		assert.ErrorIs(t, r.Validate(), ErrInvalidRequest, name)
	}
}
