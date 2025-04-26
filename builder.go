package openapigen

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fmarmol/kin-openapi/openapi3"
	"github.com/fmarmol/openapigen/utils"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

var _enumImpl = reflect.TypeOf((*Enum)(nil)).Elem()
var _extensionsImpl = reflect.TypeOf((*ExtensionsI)(nil)).Elem()
var _selfExtentionsImpl = reflect.TypeOf((*SelfExtensionsI)(nil)).Elem()

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

type Property struct {
	name                 string
	_type                string
	format               string
	ref                  string
	additionalProperties *Property // ref to add object for the additional properties
	itemsProp            *Property
	required             bool
	description          string
	deprecated           bool
	_default             any
	nullable             bool
	minimum              *float64
	maximum              *float64
	enums                []any
	extensions           map[string]any
}

func (p Property) String() string {
	data := map[string]any{
		"name":                p.name,
		"_type":               p._type,
		"format":              p.format,
		"ref":                 p.ref,
		"items_props":         p.itemsProp,
		"addional_properties": p.additionalProperties,
		"required":            p.required,
		"description":         p.description,
		"deprecated":          p.deprecated,
		"_default":            p._default,
		"nullable":            p.nullable,
		"minimum":             p.minimum,
		"maximum":             p.maximum,
		"enums":               p.enums,
		"extensions":          p.extensions,
	}
	raw, _ := json.Marshal(data)
	return string(raw)
}

type Schema struct {
	object any
	owner  *Schema
	name   string
	enums  []any
	array  bool
}

func (s *Parameter) RefPath() string {
	if s.componentName == "" {
		panic("anonymous struct is not supported yet")
	}
	return fmt.Sprintf("#/components/parameters/%s", s.componentName)
}

func (s *Schema) RefPath() string {
	name := s.ObjectName()
	if name == "" {
		panic("anonymous struct is not supported yet")
	}
	return fmt.Sprintf("#/components/schemas/%s", s.ObjectName())
}

func (s *Schema) ObjectName() string {
	_type := reflect.TypeOf(s.object)
	name := _type.Name()
	if _type.Kind() == reflect.Slice {
		name = _type.Elem().Name() + "s" // TODO: a better pluralize function
	}
	if s.object == nil || s.name != "" {
		return s.name
	}

	if strings.Contains(name, "[") { // we assume we met a generic type, need to transform the name in something compatible with openapi
		name = strings.ReplaceAll(name, "[", "_")
		name = strings.ReplaceAll(name, "]", "_")
		name = strings.ReplaceAll(name, "/", "_")
		name = strings.ReplaceAll(name, "*", "_")
	}
	return name
}

func (s *Schema) Properties() ([]Property, []*Schema) {
	return Properties(s.object)
}

func tagFieldLookUp(tags []string, key string) (string, bool) {
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if strings.Contains(t, key+":") {
			value := t[len(key)+1:]
			return value, true
		}
	}
	return "", false

}

func setProperty(property *Property, newSchemas []*Schema, _type reflect.Type) ([]*Schema, *Schema) { // (all schemas, last schema added)
	kind := _type.Kind()
	var lastSchema *Schema

	switch kind {
	case reflect.Pointer:
		return setProperty(property, newSchemas, _type.Elem())
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
	case reflect.TypeOf(uuid.UUID{}).Kind():
		property._type = "string"
		property.format = "uuid"
	case reflect.Struct:
		newSchema := NewSchema(reflect.New(_type).Elem().Interface())
		property.ref = newSchema.RefPath()
		lastSchema = newSchema
		newSchemas = append(newSchemas, newSchema)
	case reflect.Slice:
		elemType := _type.Elem()
		property.itemsProp = &Property{}
		newSchemas, lastSchema = setProperty(property.itemsProp, newSchemas, elemType)
	case reflect.Map:
		keyType := _type.Key()
		if keyType != reflect.TypeOf("") {
			panic("key has to be string")
		}
		property.additionalProperties = &Property{}
		newSchemas, lastSchema = setProperty(property.additionalProperties, newSchemas, reflect.New(_type.Elem()).Elem().Type())
	default:
		panic(fmt.Errorf("kind %v not supported yet %v", _type.Kind(), property))
	}
	return newSchemas, lastSchema

}

func Properties(object any) ([]Property, []*Schema) {
	ret := []Property{}
	newSchemas := []*Schema{}

	_type := reflect.TypeOf(object)

	if _type.Kind() == reflect.Pointer {
		_type = _type.Elem()
	}

	if _type.Kind() == reflect.Slice {
		elemType := _type.Elem()
		if elemType.Kind() == reflect.Struct {
			newSchema := NewSchema(reflect.New(_type.Elem()).Elem().Interface())
			newSchema.array = true
			newSchema.owner = NewSchema(object)
			newSchemas = append(newSchemas, newSchema)
		}
		return ret, newSchemas
	}

	if _type.Kind() != reflect.Struct {
		panic(fmt.Errorf("object %v is not a struct", _type.Name()))
	}

	// Extensions
	var extensions map[FieldName]Extensions
	if _type.Implements(_extensionsImpl) {
		method, ok := _type.MethodByName("Extensions")
		if !ok {
			panic("not an extension")
		}
		dst := reflect.New(_type).Elem()
		values := method.Func.Call([]reflect.Value{dst})
		if len(values) != 1 {
			panic("Values() method should return a map, 0 found")
		}
		_extensions, ok := values[0].Interface().(map[string]map[string]any)
		if !ok {
			panic("extensions type cannot be converted into map[string]map[string]any")
		}
		extensions = _extensions
	}

	// Enum in parameter
	if _type.Implements(_enumImpl) {
		method, ok := _type.MethodByName("Values")
		if !ok {
			panic("not an enum")
		}
		dst := reflect.New(_type).Elem()
		values := method.Func.Call([]reflect.Value{dst})
		if len(values) != 1 {
			panic("Values() method should return 1 slice, 0 found")
		}
		enums, ok := values[0].Interface().([]any)
		if !ok {
			panic("enum values cannot be converted into []any")
		}
		newSchema := &Schema{enums: enums, object: dst.Interface()}
		newSchemas = append(newSchemas, newSchema)
	}

	for i := range _type.NumField() {
		field := _type.Field(i)
		fieldName := ToSnakeCase(field.Name)
		if !field.IsExported() {
			continue
		}
		var property Property

		tag := field.Tag.Get("oapi")
		if tag == "" {
			property.name = fieldName

		} else {
			tagValues := strings.Split(tag, ",")
			if len(tagValues) == 0 {
				panic("invalid oapi tag")
			}
			if value, ok := tagFieldLookUp(tagValues, "name"); ok {
				property.name = value
			} else {
				property.name = fieldName
			}
			if value, ok := tagFieldLookUp(tagValues, "format"); ok {
				property.format = value
			}
			if value, ok := tagFieldLookUp(tagValues, "description"); ok {
				property.description = value
			}
			if _, ok := tagFieldLookUp(tagValues, "deprecated"); ok {
				property.deprecated = true
			}
			if value, ok := tagFieldLookUp(tagValues, "default"); ok {
				property._default = parseString(value)
			}
			if value, ok := tagFieldLookUp(tagValues, "min"); ok {
				if val, err := strconv.ParseFloat(value, 64); err == nil {
					property.minimum = &val
				}
			}
			if value, ok := tagFieldLookUp(tagValues, "max"); ok {
				if val, err := strconv.ParseFloat(value, 64); err == nil {
					property.maximum = &val
				}
			}
			if slices.Contains(tagValues, "required:true") {
				property.required = true
			}
			if slices.Contains(tagValues, "nullable:true") {
				property.nullable = true
			}
		}
		exts, ok := extensions[field.Name]
		if ok {
			property.extensions = exts
		}

		if field.Type.Implements(_enumImpl) {
			method, ok := field.Type.MethodByName("Values")
			if !ok {
				panic("not an enum")
			}
			dst := reflect.New(field.Type).Elem()
			values := method.Func.Call([]reflect.Value{dst})
			if len(values) != 1 {
				panic("Values() method should return 1 slice, 0 found")
			}
			enums, ok := values[0].Interface().([]any)
			if !ok {
				panic("enum values cannot be converted into []any")
			}
			newSchema := &Schema{enums: enums, object: dst.Interface()}
			property.ref = newSchema.RefPath()
			ret = append(ret, property)
			newSchemas = append(newSchemas, newSchema)
			continue
		}

		switch field.Type {
		case reflect.TypeOf(uuid.UUID{}):
			property._type = "string"
			property.format = "uuid"
		case reflect.TypeOf(time.Time{}):
			property._type = "string"
			property.format = "date-time"
		default:
			newSchemas, _ = setProperty(&property, newSchemas, field.Type)
		}
		ret = append(ret, property)

	}
	return ret, newSchemas
}

func NewSchema(ref any) *Schema {
	return &Schema{object: ref}
}

type Response struct {
	code        int // -1 for default
	description string
	ref         *Schema
	content     string
	inline      []byte // WARNING: this only a temp fix to have a custom response inline, openapi3.Response (only json) (not a ref)
	headers     map[string]*Property
}

func NewResponse(code int) *Response {
	r := new(Response)
	r.code = code
	return r
}

// Header support only native types
// TODO find a better way to set header
func (r *Response) Header(key string, obj any, description ...string) *Response {

	property := new(Property)
	newSchemas := []*Schema{}
	_, _ = setProperty(property, newSchemas, reflect.TypeOf(obj))

	if len(description) > 0 {
		property.description = description[0]
	}

	if r.headers == nil {
		r.headers = make(map[string]*Property)
	}
	r.headers[key] = property
	return r
}

func (r *Response) Inline(data map[string]any) *Response {
	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	r.inline = raw
	return r
}

func (r *Response) Content(s string, obj any) *Response {
	r.content = s
	r.ref = NewSchema(obj)
	return r
}

func (r *Response) JSON(object any) *Response {
	r.content = "application/json"
	r.ref = NewSchema(object)
	return r
}

func (r *Response) Description(s string) *Response {
	r.description = s
	return r
}

type Tag struct {
	Name        string
	Description string
}

type Document struct {
	t               *openapi3.T
	paths           []*Path
	Version         string
	Title           string
	servers         []string
	bearerAuth      bool // only support bearer JWT for now
	tags            []Tag
	defaultResponse *Response
}

func (d *Document) SetDefaultResponse(r *Response) *Document {
	d.defaultResponse = r
	return d
}

func (d *Document) Tags(tags ...Tag) *Document {
	d.tags = append(d.tags, tags...)
	return d
}

func (d *Document) BearerAuth() *Document {
	d.bearerAuth = true
	return d
}

func (d *Document) Path(p *Path) *Document {
	p.defaultResponse = d.defaultResponse
	d.paths = append(d.paths, p)
	return d
}

func (d *Document) Paths(ps ...*Path) *Document {
	for _, p := range ps {
		d.Path(p)
	}
	return d
}

func (d *Document) Server(url string) *Document {
	d.servers = append(d.servers, url)
	return d
}

func setPathItemOperation(method string, pi *openapi3.PathItem, op *openapi3.Operation) error {
	switch method {
	case "get":
		pi.Get = op
	case "put":
		pi.Put = op
	case "post":
		pi.Post = op
	case "delete":
		pi.Delete = op
	case "options":
		pi.Options = op
	case "patch":
		pi.Patch = op
	case "connect":
		pi.Connect = op
	case "trace":
		pi.Trace = op
	default:
		return fmt.Errorf("method %s not supported", method)
	}
	return nil
}

func (d *Document) Write(w io.Writer, indent int) error {
	err := d.Build()
	if err != nil {
		return err
	}
	finalDoc := NewYamlDocument(d)
	enc := yaml.NewEncoder(w)
	enc.SetIndent(indent)
	return enc.Encode(finalDoc)
}

func (d *Document) Build() error {

	if d.t == nil {
		servers := utils.Map(d.servers, func(s string) *openapi3.Server {
			return &openapi3.Server{URL: s}
		})
		d.t = &openapi3.T{
			OpenAPI:    "3.0.0",
			Info:       &openapi3.Info{Version: d.Version, Title: d.Title},
			Servers:    openapi3.Servers(servers),
			Components: &openapi3.Components{},
		}
		if d.bearerAuth {
			d.t.Security = []openapi3.SecurityRequirement{
				map[string][]string{"bearerAuth": {}},
			}
			d.t.Components.SecuritySchemes = map[string]*openapi3.SecuritySchemeRef{
				"bearerAuth": {
					Value: &openapi3.SecurityScheme{
						Type:         "http",
						Scheme:       "bearer",
						BearerFormat: "JWT",
					},
				}}
		}

	}
	for _, t := range d.tags {
		d.t.Tags = append(d.t.Tags, &openapi3.Tag{Name: t.Name, Description: t.Description})
	}
	if d.t.Paths == nil {
		d.t.Paths = openapi3.NewPaths()
	}

	type OperationToRegister struct {
		method    string
		operation *openapi3.Operation
	}

	operationsToRegister := map[string][]OperationToRegister{}

	for _, path := range d.paths {
		path.SetDefaultResponse() // TODO try to find a better place to set
		responses := openapi3.NewResponses()

		if d.t.Components.Schemas == nil {
			d.t.Components.Schemas = make(openapi3.Schemas)
		}
		for code, r := range path.apiResponses {
			responses.Set(code, r)
			// if d.globalDefaultResponse != nil {
			// 	responses.Set("default", )
			// }
		}
		for name, schema := range path.apiSchemas {
			d.t.Components.Schemas[name] = schema
		}

		if d.t.Components.Parameters == nil {
			d.t.Components.Parameters = make(openapi3.ParametersMap)
		}
		for name, param := range path.componentParameters {
			d.t.Components.Parameters[name] = param
		}

		operation := &openapi3.Operation{
			Tags:        path.tags,
			Summary:     path.summary,
			Description: path.description,
			OperationID: path.operationID,
			Responses:   responses,
			Parameters:  path.parameters,
		}
		if path.description == "" {
			operation.Description = path.summary
		}
		if path.inline != nil {
			var openapiReq openapi3.RequestBody
			err := json.Unmarshal(path.inline, &openapiReq)
			if err != nil {
				panic(err)
			}
			operation.RequestBody = &openapi3.RequestBodyRef{
				Value: &openapiReq,
			}
		}
		if path.ref != nil {
			operation.RequestBody = &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Required: path.contentRequired,
					Content: openapi3.Content{
						path.content: &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Ref: path.ref.RefPath(),
							},
						},
					},
				},
			}
		}

		operationsToRegister[path.path] = append(operationsToRegister[path.path], OperationToRegister{method: path.method, operation: operation})

	}
	for path, operations := range operationsToRegister {
		newPathItem := new(openapi3.PathItem)
		for _, operation := range operations {
			setPathItemOperation(operation.method, newPathItem, operation.operation)
		}

		d.t.Paths.Set(path, newPathItem)
	}
	return nil
	// returns a map
	// return d.t.MarshalYAML()
}
