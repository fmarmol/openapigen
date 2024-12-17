package openapigen

type Pin string

const (
	PATH  Pin = "path"
	QUERY Pin = "query"
)

type Parameter struct {
	isComponent   bool
	componentName string
	name          string
	in            Pin
	_type         string
	format        string
	required      bool
	ref           any
	enums         Enum
	min, max      *float64
}

func NewParameter(name string) *Parameter {
	ret := new(Parameter)
	ret.name = name
	return ret
}

func (p *Parameter) AsComponent(name string) *Parameter {
	p.isComponent = true
	p.componentName = name
	return p
}

func (p *Parameter) Name(name string) *Parameter {
	p.name = name
	return p
}

func (p *Parameter) InPath() *Parameter {
	return p.In(PATH)
}

func (p *Parameter) InQuery() *Parameter {
	return p.In(QUERY)
}

func (p *Parameter) In(v Pin) *Parameter {
	p.in = v
	return p
}

func (p *Parameter) Type(v string) *Parameter {
	p._type = v
	return p
}

func (p *Parameter) Format(v string) *Parameter {
	p.format = v
	return p
}

func (p *Parameter) Required() *Parameter {
	p.required = true
	return p
}

func (p *Parameter) Ref(v any) *Parameter {
	p.ref = v
	return p
}

func (p *Parameter) Enum(v Enum) *Parameter {
	p.enums = v
	return p
}

func (p *Parameter) Min(v float64) *Parameter {
	p.min = &v
	return p
}
func (p *Parameter) Max(v float64) *Parameter {
	p.max = &v
	return p
}
