package openapigen

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type MyEnum struct{}

func (p MyEnum) Values() []any {
	return []any{"FOO", "BAR"}
}
func (p MyEnum) Kind() string { return "string" } // TODO not used yet

type MyBody struct {
	MyList []string
	MyEnum MyEnum // enum example
	IsTrue bool
	File   string `oapi:"name:file,format:binary,description:binary file"`
}

type MyResponse struct {
	ID   uuid.UUID `oapi:"name:custom_id,required:true"`
	Name string
}
type Error struct {
	Code    int
	Message string
}

type Addr struct {
	Street int
	City   string
}

type Person struct {
	Name      string
	Addresses []Addr
	Age       float32
}

type Persons []Person

func TestBuilder(t *testing.T) {
	doc := &Document{Version: "0.0.1", Title: "awesome api"}
	doc.Server("/api").Server("/api/v3").BearerAuth().
		Paths(
			NewPath("/request/things").Get().Tags("things").OperationID("GetThings").
				Response(
					NewResponse(200).Description("OK").JSON(Persons{}),
				),
			NewPath("/requests/{id}/things").Get().Tags("things").OperationID("GetOneThing").
				Parameter(Parameter{In: "path", Format: "uuid", Type: "string", Required: true, Name: "id"}).
				Responses(
					NewResponse(200).Description("OK").JSON(MyResponse{}),
					NewResponse(500).Description("ERROR").JSON(Error{}),
				),
			NewPath("/requests/{id}/thing").Post().Tags("thing").OperationID("CreateOneThing").
				Parameter(Parameter{In: "path", Format: "uuid", Type: "string", Required: true, Name: "id"}).
				Parameter(Parameter{In: "query", Name: "type", Ref: MyEnum{}}).
				Parameter(Parameter{In: "query", Name: "toto", Ref: []uuid.UUID{}}).
				FormData(MyBody{}).
				Responses(
					NewResponse(200).Description("OK").JSON(MyResponse{}),
					NewResponse(500).Description("ERROR").JSON(Error{}),
				),
		)

	buffer := bytes.NewBuffer(nil)
	err := doc.Write(buffer, 2)
	fmt.Println(buffer.String())
	require.NoError(t, err)
}
