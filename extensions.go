package openapigen

type Extensions = map[string]any
type FieldName = string

type ExtensionsI interface {
	Extensions() map[FieldName]Extensions
}

type SelfExtensionsI interface {
	SelfExtensions() Extensions
}
