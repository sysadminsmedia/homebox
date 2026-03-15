package repo

import "github.com/samber/lo"

// mapTErrFunc is a factory function that returns a mapper function that
// wraps the given mapper function but first will check for an error and
// return the error if present.
//
// Helpful for wrapping database calls that return both a value and an error
func mapTErrFunc[T any, Y any](fn func(T) Y) func(T, error) (Y, error) {
	return func(t T, err error) (Y, error) {
		if err != nil {
			var zero Y
			return zero, err
		}

		return fn(t), nil
	}
}

func mapTEachFunc[T any, Y any](fn func(T) Y) func([]T) []Y {
	return func(items []T) []Y {
		return lo.Map(items, func(item T, _ int) Y {
			return fn(item)
		})
	}
}

func mapTEachErrFunc[T any, Y any](fn func(T) Y) func([]T, error) ([]Y, error) {
	return func(items []T, err error) ([]Y, error) {
		if err != nil {
			return nil, err
		}

		return lo.Map(items, func(item T, _ int) Y {
			return fn(item)
		}), nil
	}
}

func mapEach[T any, U any](items []T, fn func(T) U) []U {
	return lo.Map(items, func(item T, _ int) U {
		return fn(item)
	})
}
