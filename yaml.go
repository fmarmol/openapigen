package openapigen

type YamlDocument struct {
	Openapi    any `yaml:"openapi,omitempty"`
	Info       any `yaml:"info,omitempty"`
	Servers    any `yaml:"servers,omitempty"`
	Security   any `yaml:"security"`
	Tags       any `yaml:"tags"`
	Paths      any `yaml:"paths,omitempty"`
	Components any `yaml:"components,omitempty"`
}

func NewYamlDocument(d *Document) YamlDocument {

	doc, err := d.t.MarshalYAML()
	if err != nil {
		panic(err)
	}
	dict := doc.(map[string]any)

	var ret YamlDocument
	ret.Openapi = dict["openapi"]
	ret.Info = dict["info"]
	ret.Servers = dict["servers"]
	ret.Security = dict["security"]
	ret.Tags = dict["tags"]
	ret.Paths = dict["paths"]
	ret.Components = dict["components"]
	return ret
}
