package utils

func DerefPtrWithConv[T any, U any](v *T, convert func(v T) U, _default U) U {
	if v == nil {
		return _default
	}
	return convert(*v)
}

func DerefPtr[T any](v *T, _default T) T {
	if v == nil {
		return _default
	}
	return *v
}

func NewPtr[T any](v T) *T {
	return &v
}

func Deduplicate[T comparable](s []T) []T {
	inResult := make(map[T]bool)
	var ret []T
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			ret = append(ret, str)
		}
	}
	return ret
}
