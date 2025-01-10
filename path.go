package openapigen

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

type Parameter struct {
	Name     string
	In       string
	Type     string
	Format   string
	Required bool
	Ref      any
}

type Path struct {
	path         string
	method       string
	tags         []string
	summary      string
	operationID  string
	parameters   []*openapi3.ParameterRef
	responses    []*Response
	apiResponses map[string]*openapi3.ResponseRef
	apiSchemas   map[string]*openapi3.SchemaRef
	jsonBody     *Schema
	formData     *Schema
}

func NewPath(path string) *Path {
	p := new(Path)
	p.path = path
	p.apiResponses = make(map[string]*openapi3.ResponseRef)
	p.apiSchemas = make(map[string]*openapi3.SchemaRef)
	// //parse path

	// KEEP FOR LATER
	// elems := strings.Split(path, "/")
	// for _, elem := range elems {
	// 	if elem[0] == '{' && elem[len(elem)-1] == '}' {
	// 		param := elem[1 : len(elem)-1]

	// 		p.parameters = append(p.parameters, Parameter{
	// 			in:       "path",
	// 			_type:    "string",
	// 			format:   "uuid",
	// 			required: true,
	// 			name:     param,
	// 		})

	// 	}
	// }

	return p
}

func (p *Path) JSONBody(obj any) *Path {
	p.jsonBody = NewSchema(obj)
	p.registerSchema(p.jsonBody)
	return p
}

func (p *Path) FormData(obj any) *Path {
	p.formData = NewSchema(obj)
	p.registerSchema(p.formData)
	return p
}

func (p *Path) Parameter(param Parameter) *Path {

	if param.Ref != nil {
		if t, ok := isSlice(param.Ref); ok {
			prop := typeToProp(t)
			p.parameters = append(p.parameters, &openapi3.ParameterRef{
				Value: &openapi3.Parameter{
					In:       param.In,
					Name:     param.Name,
					Required: param.Required,
					Schema: &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"array"},
							Items: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:   &openapi3.Types{prop._type},
									Format: prop.format,
								},
							},
						},
					},
				},
			})
			return p
		}

		schema := NewSchema(param.Ref)
		p.parameters = append(p.parameters, &openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				In:       param.In,
				Name:     param.Name,
				Required: param.Required,
				Schema: &openapi3.SchemaRef{
					Ref: schema.RefPath(),
				},
			},
		})
		p.registerSchema(schema)
		return p
	}
	_type := &openapi3.Types{param.Type}

	p.parameters = append(p.parameters, &openapi3.ParameterRef{
		Value: &openapi3.Parameter{
			In:       param.In,
			Name:     param.Name,
			Required: param.Required,
			Schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   _type,
					Format: param.Format,
				},
			},
		},
	})
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

	if s.enums != nil {
		value.Enum = s.enums
		value.Type = &openapi3.Types{"string"} // TODO support other type
		p.apiSchemas[s.ObjectName()] = openapi3.NewSchemaRef("", value)
		return
	}
	if s.array {
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

		if property.items {
			value.Properties[property.name] = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"array"},
					Items: &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{property._type},
							Format:      property.format,
							Description: property.description,
							Deprecated:  property.deprecated,
							Default:     property._default,
							Min:         property.minimum,
							Max:         property.maximum,
						},
					},
				},
			}
			continue

		}
		if property.itemsRef != "" {
			value.Properties[property.name] = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"array"},
					Items: &openapi3.SchemaRef{
						Ref: property.itemsRef,
					},
				},
			}
			continue
		}
		if property.ref != "" {
			value.Properties[property.name] = &openapi3.SchemaRef{
				Ref: property.ref,
			}
		} else {
			value.Properties[property.name] = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:        &openapi3.Types{property._type},
					Format:      property.format,
					Description: property.description,
					Deprecated:  property.deprecated,
					Default:     property._default,
					Min:         property.minimum,
					Max:         property.maximum,
				},
			}
		}
	}
	p.apiSchemas[s.ObjectName()] = openapi3.NewSchemaRef("", value)
	for _, s := range newSchemas {
		p.registerSchema(s)
	}

}

func (p *Path) Response(r *Response) *Path {
	p.responses = append(p.responses, r)
	if r.json == nil {
		return p
	}
	codeStr := fmt.Sprint(r.code)
	if r.code == -1 {
		codeStr = "default"
	}

	p.apiResponses[codeStr] = &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: &r.description,
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Ref: r.json.RefPath(),
					},
				},
			},
		},
	}
	p.registerSchema(r.json)
	return p
}
