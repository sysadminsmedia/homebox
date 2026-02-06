package repo

import (
	"time"

	"github.com/samber/lo"
)

func sqliteDateFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// orDefault returns the value of the pointer if it is not nil, otherwise it returns the default value
//
// This is used for nullable or potentially nullable fields (or aggregates) in the database when running
// queries. If the field is null, the pointer will be nil, so we return the default value instead.
func orDefault[T any](v *T, def T) T {
	return lo.FromPtrOr(v, def)
}
