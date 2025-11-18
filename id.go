package main

import (
	"math"

	gonanoid "github.com/matoous/go-nanoid"
)

const (
	customAlphabet       = "abcdefghijklmnopqrstuvwxyz0123456789"
	alphabetSize         = 36
	minIDLength          = 3
	collisionProbability = 0.001 // 0.1%
)

// CalculateIDLength calculates required ID length based on issue count
// to maintain 0.1% collision probability using birthday paradox formula:
// L = max(minIDLength, ceil(log_base(n² / (2 × ln(1/(1-P))))))
func CalculateIDLength(issueCount int) int {
	if issueCount <= 0 {
		return minIDLength
	}

	// Calculate denominator: -2 × ln(1-P) = -2 × ln(0.999) ≈ 0.002001
	denominator := -2.0 * math.Log(1.0-collisionProbability)

	// Calculate required capacity: n² / denominator
	n := float64(issueCount)
	requiredCapacity := (n * n) / denominator

	// Calculate required length: log_base(requiredCapacity)
	length := math.Log(requiredCapacity) / math.Log(alphabetSize)

	// Round up and apply minimum
	calculatedLength := int(math.Ceil(length))
	if calculatedLength < minIDLength {
		return minIDLength
	}
	return calculatedLength
}

// GenerateID generates a unique ID with the given prefix and length
// Example: GenerateID("mint", 7) -> "mint-xgmx5l6"
func GenerateID(prefix string, length int) string {
	id, err := gonanoid.Generate(customAlphabet, length)
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
