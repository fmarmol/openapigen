package openapigen

type Enum interface {
	Values() []any
}

type enums struct{ v []any }

func (e enums) Values() []any {
	return e.v
}

func Enums(val ...any) Enum {
	return enums{val}
}
