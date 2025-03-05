package openapigen

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
	Street []int
	City   string
}

func (Addr) Extensions() map[FieldName]Extensions {
	return map[string]map[string]any{
		"Street": {"toto": "tata"},
	}
}

type Person struct {
	Name      string `oapi:"required:true,deprecated:true"`
	Addresses []Addr
	Age       float32 `oapi:"default:12.1,min:1,max:42,nullable:true"`
	Toto      *int
}

func (p Person) Extensions() map[FieldName]Extensions {
	return map[string]map[string]any{
		"Addresses": {
			"x-go-type": "uuid.UUID",
		},
	}
}

// func (p Person) SelfExtensions() Extensions {
// 	return map[string]any{
// 		"toto": "tata",
// 	}
// }

type Persons []Person

type OrderBy struct {
	Field string `oapi:"required:true"`
	Order string `oapi:"required:true"`
}

var OrderByQueryParam = NewComponentParameter("orderByQueryParam", Parameter{
	In:   "query",
	Name: "order",
	Ref:  []OrderBy{},
})

func TestBuilder(t *testing.T) {

	doc := &Document{Version: "0.0.1", Title: "awesome api"}
	// doc.SetDefaultResponse(NewResponse(-1).JSON(Person{}).Description("default response"))
	doc.Tags(Tag{Name: "one", Description: "one des"}, Tag{Name: "two", Description: "two"})
	doc.Server("/api").Server("/api/v3").BearerAuth().
		Paths(
			NewPath("/batches/").Delete().OperationID("listBatches").Summary("delete a batch").
				JSONBody(Person{}).
				// Content(Person{}, "image/*", true).
				// Inline(map[string]any{
				// 	"description": "OK",
				// 	"content": map[string]any{
				// 		"image/*": map[string]any{
				// 			"schema": map[string]any{
				// 				"type":   "string",
				// 				"format": "binary",
				// 			},
				// 		},
				// 	},
				// },
				// ).
				Responses(
					// NewResponse(204).Content("toto/titi", Person{}).Description("OK"),
					NewResponse(203).Inline(map[string]any{
						"description": "OK",
						"content": map[string]any{
							"image/*": map[string]any{
								"schema": map[string]any{
									"type":   "string",
									"format": "binary",
								},
							},
						},
					},
					).Description("OK"),
					// NewResponse(-1).JSON(Person{}).Description("DEFAULT"),
				),
			// NewPath("/batches/").Post().OperationID("createBatches").
			// 	Responses(
			// 		NewResponse(200).Description("OK"),
			// 	),
			// NewPath("/request/things").Get().Tags("things").OperationID("GetThings").
			// 	Response(
			// 		NewResponse(200).Description("OK").JSON(Persons{}),
			// 	),
			// NewPath("/requests/{id}/things").Get().Tags("things").OperationID("GetOneThing").
			// 	Parameter(Parameter{In: "path", Format: "uuid", Type: "string", Required: true, Name: "id"}).
			// 	Responses(
			// 		NewResponse(200).Description("OK").JSON(MyResponse{}),
			// 		NewResponse(500).Description("ERROR").JSON(Error{}),
			// 	),
			// NewPath("/requests/{id}/thing").Post().Tags("thing").OperationID("CreateOneThing").
			// 	Parameter(Parameter{In: "path", Format: "uuid", Type: "string", Required: true, Name: "id"}).
			// 	Parameter(Parameter{In: "query", Name: "type", Ref: MyEnum{}}).
			// 	Parameter(Parameter{In: "query", Name: "toto", Ref: []uuid.UUID{}}).
			// 	FormData(MyBody{}).
			// 	Responses(
			// 		NewResponse(200).Description("OK").JSON(MyResponse{}),
			// 		NewResponse(500).Description("ERROR").JSON(Error{}),
			// 	),
		)

	buffer := bytes.NewBuffer(nil)
	err := doc.Write(buffer, 2)
	fmt.Println(buffer.String())
	require.NoError(t, err)
}

func TestBuilderParameter(t *testing.T) {

	doc := &Document{}
	doc.
		Paths(
			NewPath("/items").Get().
				Parameter(OrderByQueryParam).
				Parameter(Parameter{In: "query", Name: "order_2", Ref: []OrderBy{}}),
		)

	buffer := bytes.NewBuffer(nil)
	err := doc.Write(buffer, 2)
	fmt.Println(buffer.String())
	require.NoError(t, err)
	assert.Equal(t, testBuilderParameterExpectedSpecs, strings.ReplaceAll(buffer.String(), "\n", ""))
}
