package v1

import (
	"net/url"
	"strconv"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

func queryUUIDList(params url.Values, key string) []uuid.UUID {
	return lo.FilterMap(params[key], func(id string, _ int) (uuid.UUID, bool) {
		uid, err := uuid.Parse(id)
		return uid, err == nil
	})
}

func queryIntOrNegativeOne(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return i
}

func queryBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}
