package repo

import "github.com/samber/lo"

type PaginationResult[T any] struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
	Items    []T `json:"items"`
}

// EntityListResult is the response type for paginated entity queries,
// extending PaginationResult with a total price summary.
type EntityListResult struct {
	PaginationResult[EntitySummary]
	TotalPrice float64 `json:"totalPrice"`
}

func calculateOffset(page, pageSize int) int {
	offset := (page - 1) * pageSize
	return lo.Ternary(offset < 0, 0, offset)
}
