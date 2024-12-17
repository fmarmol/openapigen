package utils

import "fmt"

func MapToString[T any](t T) string {
	return fmt.Sprint(t)
}

func Map[T any, U ~[]T, K any, V []K](list U, transform func(t T) K) V {
	ret := make(V, len(list))
	for index, elem := range list {
		ret[index] = transform(elem)
	}
	return ret
}
