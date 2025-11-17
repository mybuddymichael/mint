package main

import (
	gonanoid "github.com/matoous/go-nanoid"
)

const (
	idLength       = 7
	customAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
)

// GenerateID generates a unique ID with the given prefix
// Example: GenerateID("mint") -> "mint-xgmx5l6"
func GenerateID(prefix string) string {
	id, err := gonanoid.Generate(customAlphabet, idLength)
	if err != nil {
		// Fallback to default if generation fails (should be rare)
		panic(err)
	}
	return prefix + "-" + id
}
