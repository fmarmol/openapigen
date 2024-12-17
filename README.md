# openapigen

:warning: This package is in early development. Breaking changes can occur in next phases of development.

## Goal
Use DSL like written in go to build your openapi yaml files.

## TOC
- [Installation](#installation)
- [Getting started](#getting-started)
- [Routing](#routing)
- [Parameters](#parameters)

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

### path parameters

### query parameters

