# openapigen

:warning: This package is in early development. Breaking changes can occur in next phases of development.

## Goal
Use DSL like written in go to build your openapi yaml files.

## TOC
- [Installation](#installation)
- [Getting started](#getting-started)
- [Routing](#routing)
- [Parameters](#parameters)
- [Request Body description](#request-body)
- [Response Body description](#response-body)
- [Fields description](#fields)
- [Enums](#enums)
- [Extensions](#extensions)
- [Additional properties](#additional-properties)
- [Generics](#generics)

### Installation
```sh
go get github.com/fmarmol/openapigen@latest
```

### Getting started

```go
package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/fmarmol/openapigen"
)

type Movie struct {
	Title string `json:"title"`
	Year  int    `json:"year"`
}

type Movies []Movie

func generateDoc() error {
	doc := openapigen.Document{Title: "my api", Version: "1.0"}
	doc.Paths(
		openapigen.NewPath("/movies").Get().
			Description("return a list of movies").
			Responses(
				openapigen.NewResponse(200).JSON(Movies{}).Description("success"),
			),
	)
	return doc.Write(os.Stdout, 2)
}

func main() {
	_ = generateDoc()
	http.HandleFunc("GET /movies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		movies := []Movie{
			{Title: "star wars", Year: 1977},
			{Title: "matrix", Year: 1999},
		}
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(movies)
	})
	_ = http.ListenAndServe(":8080", nil)
}
```

We'll give the following result:

```sh
openapi: 3.0.0
info:
  title: my api
  version: "1.0"
security: null
tags: null
paths:
  /movies:
    get:
      description: return a list of movies
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Movies'
          description: success
        default:
          description: ""
components:
  schemas:
    Movie:
      properties:
        title:
          type: string
        year:
          type: integer
      type: object
    Movies:
      items:
        $ref: '#/components/schemas/Movie'
      type: array
```

## Routing
In every rest API you have to choose an HTTP method for each of your route. In openapigen you write the same by using one the following methods:

```go
  Path.Get()
  Path.Post()
  Path.Put()
  Path.Patch()
  Path.Delete()
  Path.Options()
  Path.Connect()
  Path.Trace()
```

## Parameters
What would be an api without path or query parameters, you can easily describe these parameters using the following methods.

### path parameters
For example if you have an endpoint using an `uuid` to get information about a user
like `/api/users/{id}`

```go
Parameter(NewParameter().InPath().Name("id").Type("string").Format("uuid").Required())
```

### query parameters
You can also do the same using query parameters like `/api/users?id={id}`

```go
Parameter(NewParameter().InQuery().Name("id").Type("string").Format("uuid").Required())
```

In this case the `Required()` can be omitted, if the parameter is optional

## Request Body
For almost anything which is not a `GET` request you need to specify the body of your request.

For now the formats supported are:
- JSON with the method `JsonBody`
- Multipart/form-data with the method `FormData`


example:

```go
JsonBody(Movie{}, true) // true is optional and means the body is required
```

## Response Body
The same way for the response body which returns you api call, you can use your types using the method `JSON`


```go
Responses(
  NewResponse(200).JSON(Movie{}).Description("return a movie object"),
)
```

Only `JSON` is currently supported

## Fields

Using go structures, allow you to specify the fields in the request and response body. All exported fields will be translated into openapi components schemas.
The default behaviour will transform the fields name into optional `snake_case` openapi fields.

To have a better control on what you want to express, here a list of tags you can use.

- `name`
- `format`
- `description`
- `deprecated`
- `default`
- `min`
- `max`
- `required`
- `nullable`


Go natives types are turned into:

| types | openapi |
|-------|---------|
|int8, int16, int | `type:"integer"`              |
|int32            | `type:"int32"`                |
|int64            | `type:"int,format:in64"`      |
|float32          | `type:"number,format:double"` |
|float64          | `type:"number,format:float"`  |
|bool             | `type:boolean`                |

Few non primitive types are automatically preconfigured like:
- time.Time which is equivalent to `format:date-time`
- uuid.UUID wich is equivalent to `type:string,format:uuid` from `github.com/google/uuid` 

### Enums

Its quite common to have fields which can have only a set of values. They are enums, in order to express it into openapi you have to write a custom type
for the enum fields which implement the Enum interface

```go
type Enum interface{
  Value() []any
}  
```

for example:

```go

  type Gender string

  func (Gender) Values() []any{
    return []any{"male", "female"}
  }

  type User struct {
    Gender Gender
  }
```

#### Enums in parameters
You can also express enums in parameters using the method `Enum`

```go
Parameter(NewParameter().InQuery().Name("gender").Enum(Gender{}))
```

## Extensions

### Self extensions

## Additional properties

## Generics (beta)
