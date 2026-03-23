package repo

import "github.com/samber/lo"

type MapFunc[T any, U any] func(T) U

func (a MapFunc[T, U]) Map(v T) U {
	return a(v)
}

func (a MapFunc[T, U]) MapEach(v []T) []U {
	return lo.Map(v, func(item T, _ int) U {
		return a(item)
	})
}

func (a MapFunc[T, U]) MapErr(v T, err error) (U, error) {
	if err != nil {
		var zero U
		return zero, err
	}

	return a(v), nil
}

func (a MapFunc[T, U]) MapEachErr(v []T, err error) ([]U, error) {
	if err != nil {
		return nil, err
	}

	return a.MapEach(v), nil
}
