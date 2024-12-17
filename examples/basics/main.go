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
