package utils

func Filter[T any, U []T](list U, filter func(v T) bool) U {
	ret := make(U, 0, len(list))
	for _, elem := range list {
		if filter(elem) {
			ret = append(ret, elem)
		}
	}
	return ret
}
