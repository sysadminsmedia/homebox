package repo

type PaginationResult[T any] struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
	Items    []T `json:"items"`
	TotalPrice float64 `json:"totalPrice"`
}

func calculateOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}
