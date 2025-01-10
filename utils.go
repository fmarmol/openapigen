package openapigen

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/google/uuid"
)

func isStruct(obj any) bool {
	return reflect.TypeOf(obj).Kind() == reflect.Struct

}

func isSlice(obj any) (reflect.Type, bool) {
	_type := reflect.TypeOf(obj)
	if _type.Kind() != reflect.Slice {
		return nil, false
	}
	elem := _type.Elem()
	return elem, true
}

func typeToProp(_type reflect.Type) Property {
	var property Property
	switch {
	case _type == reflect.TypeOf(uuid.UUID{}):
		property._type = "string"
		property.format = "uuid"
		return property
	}
	kind := _type.Kind()

	switch kind {
	case reflect.Int8, reflect.Int16, reflect.Int:
		property._type = "integer"
	case reflect.Int32:
		property._type = "integer"
		property.format = "int32"
	case reflect.Int64:
		property._type = "integer"
		property.format = "int64"
	case reflect.Float64:
		property._type = "number"
		property.format = "double"
	case reflect.Float32:
		property._type = "number"
		property.format = "float"
	case reflect.Bool:
		property._type = "boolean"
	case reflect.String:
		property._type = "string"
	default:
		panic(fmt.Errorf("kind %v not supported in kindToProp", kind))
	}
	return property
}

func parseString(value string) any {
	if bool, err := strconv.ParseBool(value); err == nil {
		return bool
	} else if val, err := strconv.ParseInt(value, 10, 64); err == nil {
		return val
	} else if val, err := strconv.ParseFloat(value, 64); err == nil {
		return val
	} else {
		return value
	}
}
