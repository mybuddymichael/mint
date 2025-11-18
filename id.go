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

// FormatID formats an ID with underlined unique prefix and color 8 extraneous suffix
// Example: FormatID("mint-abc1234", 6) -> underlined "mint-a" + gray "bc1234"
func FormatID(id string, uniquePrefixLen int) string {
	if id == "" {
		return ""
	}

	// If uniquePrefixLen >= length, entire ID is unique (underline only)
	if uniquePrefixLen >= len(id) {
		return "\033[4m" + id + "\033[0m"
	}

	// Split into unique prefix and extraneous suffix
	uniquePart := id[:uniquePrefixLen]
	extraneousPart := id[uniquePrefixLen:]

	// Format: underline unique, then turn off underline and add color 8 for extraneous
	return "\033[4m" + uniquePart + "\033[24m\033[38;5;8m" + extraneousPart + "\033[0m"
}
