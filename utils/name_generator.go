package utils

import (
	"math/rand"
)

// Adjectives and Nouns for generating random names
var adjectives = []string{"autumn", "hidden", "bitter", "misty", "silent", "empty", "dry", "dark", "summer", "icy", "delicate", "quiet", "white", "cool", "little", "morning", "thin", "dawn", "small", "sparkling"}
var nouns = []string{"world", "land", "year", "wind", "fire", "hill", "pond", "grove", "sky", "bird", "forest", "stream", "meadow", "sun", "tree", "sea", "flower", "lake", "river", "frost", "dream"}

//nolint:gosec
func GenerateRandomName() string {

	// Select a random adjective and noun
	adjective := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]

	return adjective + "-" + noun
}
