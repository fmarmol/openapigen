package openapigen

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/fmarmol/openapigen/utils"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

var _enumImpl = reflect.TypeOf((*Enum)(nil)).Elem()

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

type Property struct {
	name        string
	_type       string
	format      string
	ref         string
	items       bool
	itemsRef    string
	required    bool
	description string
}

type Schema struct {
	object any
	name   string
	enums  []any
}

func (s *Schema) RefPath() string {
	if s.ObjectName() == "" {
		panic("anonymous struct is not supported yet")
	}
	return fmt.Sprintf("#/components/schemas/%s", s.ObjectName())
}

func (s *Schema) ObjectName() string {
	if s.object == nil {
		return s.name
	}
	_type := reflect.TypeOf(s.object)
	return _type.Name()
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

func Properties(object any) ([]Property, []*Schema) {
	ret := []Property{}
	newSchemas := []*Schema{}

	_type := reflect.TypeOf(object)

	if _type.Kind() != reflect.Struct {
		panic(fmt.Errorf("object %v is not a struct", _type.Name()))
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
		if !field.IsExported() {
			continue
		}
		var property Property

		tag := field.Tag.Get("oapi")
		if tag == "" {
			fieldName := ToSnakeCase(field.Name)
			property.name = fieldName

		} else {
			tagValues := strings.Split(tag, ",")
			if len(tagValues) == 0 {
				panic("invalid oapi tag")
			}
			if value, ok := tagFieldLookUp(tagValues, "name"); ok {
				property.name = value
			}
			if value, ok := tagFieldLookUp(tagValues, "format"); ok {
				property.format = value
			}
			if value, ok := tagFieldLookUp(tagValues, "description"); ok {
				property.description = value
			}
			if slices.Contains(tagValues, "required:true") {
				property.required = true
			}

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
		default:
			//nolint:all
			switch field.Type.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
				property._type = "integer"
			case reflect.Bool:
				property._type = "boolean"
			case reflect.String:
				property._type = "string"
			case reflect.Struct:
				newSchema := NewSchema(reflect.New(field.Type).Elem().Interface())
				property.ref = newSchema.RefPath()
				newSchemas = append(newSchemas, newSchema)
			case reflect.Slice:
				elemType := field.Type.Elem()
				if elemType.Kind() == reflect.Struct {
					newSchema := NewSchema(reflect.New(field.Type.Elem()).Elem().Interface())
					property.itemsRef = newSchema.RefPath()
					newSchemas = append(newSchemas, newSchema)
				} else {
					property.items = true
					switch elemType {
					case reflect.TypeOf(uuid.UUID{}):
						property._type = "string"
						property.format = "uuid"
					default:
						switch elemType.Kind() {
						case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
							property._type = "integer"
						case reflect.String:
							property._type = "string"
						}
					}
				}
			default:
				panic(fmt.Errorf("kind %v not supported yet", field.Type.Kind()))
			}
		}
		ret = append(ret, property)

	}
	return ret, newSchemas
}

func NewSchema(ref any) *Schema {
	return &Schema{object: ref}
}

type Response struct {
	code        int
	description string
	json        *Schema
}

func NewResponse(code int) *Response {
	r := new(Response)
	r.code = code
	return r
}

func (r *Response) JSON(object any) *Response {
	r.json = NewSchema(object)
	return r
}

func (r *Response) Description(s string) *Response {
	r.description = s
	return r
}

type Document struct {
	t          *openapi3.T
	paths      []*Path
	Version    string
	Title      string
	servers    []string
	bearerAuth bool // only support bearer JWT for now
}

func (d *Document) BearerAuth() *Document {
	d.bearerAuth = true
	return d
}

func (d *Document) Path(p *Path) *Document {
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

func (d *Document) Write(w io.Writer) error {
	specs, err := d.Build()
	if err != nil {
		return err
	}
	n, err := w.Write(specs)
	if err != nil {
		return err
	}
	if n != len(specs) {
		return fmt.Errorf("invalid write: %d byte written instead of %d bytes", n, len(specs))
	}
	return nil
}

func (d *Document) Build() ([]byte, error) {

	if d.t == nil {
		servers := utils.Map(d.servers, func(s string) *openapi3.Server {
			return &openapi3.Server{URL: s}
		})
		d.t = &openapi3.T{OpenAPI: "3.0.0", Info: &openapi3.Info{Version: d.Version, Title: d.Title}, Servers: openapi3.Servers(servers)}
		if d.bearerAuth {
			d.t.Security = []openapi3.SecurityRequirement{
				map[string][]string{"bearerAuth": {}},
			}
			if d.t.Components == nil {
				d.t.Components = &openapi3.Components{}
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
	if d.t.Paths == nil {
		d.t.Paths = openapi3.NewPaths()
	}
	for _, path := range d.paths {
		pathItem := &openapi3.PathItem{}
		responses := openapi3.NewResponses()

		if d.t.Components.Schemas == nil {
			d.t.Components.Schemas = make(openapi3.Schemas)
		}
		for code, r := range path.apiResponses {
			responses.Set(code, r)
		}
		for name, schema := range path.apiSchemas {
			d.t.Components.Schemas[name] = schema
		}

		operation := &openapi3.Operation{
			Tags:        path.tags,
			Summary:     path.summary,
			OperationID: path.operationID,
			Responses:   responses,
			Parameters:  path.parameters,
		}
		if path.jsonBody != nil {
			operation.RequestBody = &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Ref: path.jsonBody.RefPath(),
							},
						},
					},
				},
			}
		}
		if path.formData != nil {
			operation.RequestBody = &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Content: openapi3.Content{
						"multipart/form-data": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Ref: path.formData.RefPath(),
							},
						},
					},
				},
			}
		}

		err := setPathItemOperation(path.method, pathItem, operation)
		if err != nil {
			return nil, err
		}
		d.t.Paths.Set(path.path, pathItem)

	}
	// returns a map
	m, err := d.t.MarshalYAML()
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(m)
}
