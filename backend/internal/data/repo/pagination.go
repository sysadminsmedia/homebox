package repo

type PaginationResult[T any] struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
	Items    []T `json:"items"`
}

func calculateOffset(page, pageSize int) int {
	offset := (page - 1) * pageSize
	if offset < 0 {
		return 0
	}

	return offset
}
