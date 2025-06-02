package openapigen

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fmarmol/kin-openapi/openapi3"
)

func NewComponentParameter(name string, param Parameter) Parameter {
	param.isComponent = true
	param.componentName = name
	return param
}

type Path struct {
	path                string
	method              string
	tags                []string
	summary             string
	description         string
	operationID         string
	parameters          []*openapi3.ParameterRef
	responses           []*Response
	apiResponses        map[string]*openapi3.ResponseRef
	apiSchemas          map[string]*openapi3.SchemaRef
	componentParameters map[string]*openapi3.ParameterRef
	content             string
	ref                 *Schema
	inline              []byte // WARNING: this only a temp fix to have a custom request body inline, openapi3.Response (only json) (not a ref)
	defaultResponse     *Response
	contentRequired     bool
}

func NewPath(path string) *Path {
	p := new(Path)
	p.path = path
	p.apiResponses = make(map[string]*openapi3.ResponseRef)
	p.apiSchemas = make(map[string]*openapi3.SchemaRef)
	p.componentParameters = make(map[string]*openapi3.ParameterRef)
	return p
}

func (p *Path) Content(obj any, content string, required ...bool) *Path {
	p.ref = NewSchema(obj)
	p.content = content

	if len(required) > 0 && required[0] {
		p.contentRequired = true
	}
	p.registerSchema(p.ref)
	return p
}

func (p *Path) Inline(data map[string]any) *Path {
	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	p.inline = raw
	return p
}

func (p *Path) JSONBody(obj any, required ...bool) *Path {
	return p.Content(obj, "application/json", required...)
}

func (p *Path) FormData(obj any, required ...bool) *Path {
	return p.Content(obj, "multipart/form-data", required...)
}

func (p *Path) Parameter(param *Parameter) *Path {
	var schemaRef *openapi3.SchemaRef

	if param.ref != nil {
		if t, ok := isSlice(param.ref); ok {
			var itemsSchema *openapi3.SchemaRef
			prop, isStruct := typeToProp(t)
			if !isStruct {
				itemsSchema = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:   &openapi3.Types{prop._type},
						Format: prop.format,
						Enum:   prop.enums,
					},
				}
			} else {
				schema := NewSchema(reflect.Zero(t).Interface())
				p.registerSchema(schema)
				itemsSchema = &openapi3.SchemaRef{Ref: schema.RefPath()}
			}
			schemaRef = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:  &openapi3.Types{"array"},
					Items: itemsSchema,
				},
			}
		} else {
			schema := NewSchema(param.ref)
			p.registerSchema(schema)
			schemaRef = &openapi3.SchemaRef{Ref: schema.RefPath()}
		}
	} else {
		schemaRef = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:   &openapi3.Types{param._type},
				Format: param.format,
				Min:    param.min,
				Max:    param.max,
			},
		}
		if param.enums != nil {
			schemaRef.Value.Enum = param.enums.Values()
		}
	}

	paramRef := &openapi3.ParameterRef{}
	oapiParam := &openapi3.Parameter{
		In:       string(param.in),
		Name:     param.name,
		Required: param.required,
		Schema:   schemaRef,
	}
	if param.isComponent {
		p.registerParameter(param, oapiParam)
		paramRef.Ref = param.RefPath()
	} else {
		paramRef.Value = oapiParam
	}
	p.parameters = append(p.parameters, paramRef)

	return p
}

func (p *Path) Method(m string) *Path {
	p.method = m
	return p
}

func (p *Path) Get() *Path {
	return p.Method("get")
}
func (p *Path) Post() *Path {
	return p.Method("post")
}
func (p *Path) Put() *Path {
	return p.Method("put")
}

func (p *Path) Delete() *Path {
	return p.Method("delete")
}

func (p *Path) Options() *Path {
	return p.Method("options")
}

func (p *Path) Patch() *Path {
	return p.Method("patch")
}

func (p *Path) Connect() *Path {
	return p.Method("connect")
}
func (p *Path) Trace() *Path {
	return p.Method("trace")
}

func (p *Path) Tags(tags ...string) *Path {
	p.tags = tags
	return p
}
func (p *Path) Summary(s string) *Path {
	p.summary = s
	return p
}

func (p *Path) Description(s string) *Path {
	p.description = s
	return p
}

func (p *Path) OperationID(o string) *Path {
	p.operationID = o
	return p
}

func (p *Path) Responses(rs ...*Response) *Path {
	for _, r := range rs {
		_ = p.Response(r)
	}
	return p
}

func (p *Path) registerSchema(s *Schema) {
	value := openapi3.NewObjectSchema()

	reflectValue := reflect.ValueOf(s.object)

	if reflectValue.Type().Implements(_selfExtentionsImpl) {
		method, ok := reflectValue.Type().MethodByName("SelfExtensions")
		if !ok {
			panic("no")
		}
		dst := reflect.New(reflectValue.Type()).Elem()
		values := method.Func.Call([]reflect.Value{dst})
		if len(values) != 1 {
			panic("Values() method should return a map, 0 found")
		}
		_extensions, ok := values[0].Interface().(map[string]any)
		if !ok {
			panic("extensions type cannot be converted into map[string]any")
		}
		value.Extensions = _extensions
	}

	if s.enums != nil {
		value.Enum = s.enums
		value.Type = &openapi3.Types{"string"} // TODO support other type
		p.apiSchemas[s.ObjectName()] = openapi3.NewSchemaRef("", value)
		return
	}
	if s.array {
		// fmt.Printf("NEW OBJ: %+#v\n", s.object)

		value.Type = &openapi3.Types{"array"}
		value.Items = &openapi3.SchemaRef{
			Ref: s.RefPath(),
		}
		p.apiSchemas[s.owner.ObjectName()] = openapi3.NewSchemaRef("", value)
		p.registerSchema(NewSchema(s.object)) // need to register the child
		return
	}

	properties, newSchemas := s.Properties()

	value.Properties = make(openapi3.Schemas)
	for _, property := range properties {
		if property.required {
			value.Required = append(value.Required, property.name)
		}

		value.Properties[property.name] = oapiSchemaFromProperty(&property)

	}
	p.apiSchemas[s.ObjectName()] = openapi3.NewSchemaRef("", value)
	for _, s := range newSchemas {
		p.registerSchema(s)
	}

}

func oapiSchemaFromProperty(property *Property) *openapi3.SchemaRef {
	if property == nil {
		return nil
	}

	if property.ref != "" {
		return &openapi3.SchemaRef{
			Ref: property.ref,
		}
	}

	var pType *openapi3.Types
	switch {
	case property.itemsProp != nil:
		pType = &openapi3.Types{"array"}
	case property._type != "":
		pType = &openapi3.Types{property._type}
	}

	return &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        pType,
			Format:      property.format,
			Description: property.description,
			Deprecated:  property.deprecated,
			Default:     property._default,
			Min:         property.minimum,
			Max:         property.maximum,
			Enum:        property.enums,
			Nullable:    property.nullable,
			Extensions:  property.extensions,
			Items:       oapiSchemaFromProperty(property.itemsProp),
			AdditionalProperties: openapi3.AdditionalProperties{
				Schema: oapiSchemaFromProperty(property.additionalProperties),
			},
		},
	}
}

func (p *Path) registerParameter(param *Parameter, oapiParam *openapi3.Parameter) {
	value := openapi3.NewObjectSchema()

	properties, _ := Properties(struct {
		In     string
		Name   string
		Schema *openapi3.Schema
	}{
		In:     string(param.in),
		Name:   string(param.name),
		Schema: oapiParam.Schema.Value,
	})

	value.Properties = make(openapi3.Schemas)
	for _, property := range properties {
		if property.required {
			value.Required = append(value.Required, property.name)
		}
		value.Properties[property.name] = oapiSchemaFromProperty(&property)
	}
	p.componentParameters[param.componentName] = &openapi3.ParameterRef{
		Value: oapiParam,
	}
}

// after initial build
func (p *Path) SetDefaultResponse() {
	if p.defaultResponse != nil {
		p.apiResponses["default"] = &openapi3.ResponseRef{
			Value: &openapi3.Response{
				Description: &p.defaultResponse.description,
			},
		}
		if p.defaultResponse.ref != nil {

			p.apiResponses["default"].Value.Content = openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Ref: p.defaultResponse.ref.RefPath(),
					},
				},
			}
			p.registerSchema(p.defaultResponse.ref)
		}
	}
}

func (p *Path) Response(r *Response) *Path {

	p.responses = append(p.responses, r)
	codeStr := fmt.Sprint(r.code)
	if r.code == -1 {
		codeStr = "default"
	}
	if r.inline != nil {
		var openapiResp openapi3.Response
		err := json.Unmarshal(r.inline, &openapiResp)
		if err != nil {
			panic(err)
		}
		p.apiResponses[codeStr] = &openapi3.ResponseRef{
			Value: &openapiResp,
		}

	} else {

		if r.ref == nil {
			p.apiResponses[codeStr] = &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: &r.description,
				},
			}
		} else {
			p.apiResponses[codeStr] = &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: &r.description,
					Content: openapi3.Content{
						r.content: &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Ref: r.ref.RefPath(),
							},
						},
					},
				},
			}
			p.registerSchema(r.ref)

		}
		if r.headers != nil {
			p.apiResponses[codeStr].Value.Headers = make(openapi3.Headers)
			for key, prop := range r.headers {

				ptype := openapi3.Types([]string{prop._type})
				p.apiResponses[codeStr].Value.Headers[key] = &openapi3.HeaderRef{
					Value: &openapi3.Header{
						Parameter: openapi3.Parameter{
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:        &ptype,
									Description: prop.description,
								},
							},
						},
					},
				}
			}
		}
	}
	return p
}
