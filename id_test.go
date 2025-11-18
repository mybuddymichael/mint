package main

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	prefix := "mint"
	length := 7
	id := GenerateID(prefix, length)

	// Check format: prefix + hyphen + specified length
	if !strings.HasPrefix(id, prefix+"-") {
		t.Errorf("expected ID to start with '%s-', got '%s'", prefix, id)
	}

	expectedLen := len(prefix) + 1 + length // prefix + hyphen + nanoid
	if len(id) != expectedLen {
		t.Errorf("expected ID length %d, got %d (ID: %s)", expectedLen, len(id), id)
	}
}

func TestGenerateIDCustomAlphabet(t *testing.T) {
	prefix := "mint"
	length := 7
	id := GenerateID(prefix, length)

	// Extract the nano ID part (after prefix and hyphen)
	nanoID := strings.TrimPrefix(id, prefix+"-")

	// Verify only lowercase letters and numbers
	allowedChars := "abcdefghijklmnopqrstuvwxyz0123456789"
	for _, char := range nanoID {
		if !strings.ContainsRune(allowedChars, char) {
			t.Errorf("ID contains invalid character '%c', ID: %s", char, id)
		}
	}
}

func TestGenerateIDUniqueness(t *testing.T) {
	prefix := "mint"
	length := 7
	seen := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id := GenerateID(prefix, length)
		if seen[id] {
			t.Errorf("generated duplicate ID: %s", id)
		}
		seen[id] = true
	}

	if len(seen) != iterations {
		t.Errorf("expected %d unique IDs, got %d", iterations, len(seen))
	}
}

func TestGenerateIDDifferentPrefix(t *testing.T) {
	prefix := "test"
	length := 7
	id := GenerateID(prefix, length)

	if !strings.HasPrefix(id, prefix+"-") {
		t.Errorf("expected ID to start with '%s-', got '%s'", prefix, id)
	}

	expectedLen := len(prefix) + 1 + length // prefix + hyphen + nanoid
	if len(id) != expectedLen {
		t.Errorf("expected ID length %d, got %d", expectedLen, len(id))
	}
}

func TestCalculateIDLength(t *testing.T) {
	tests := []struct {
		name        string
		issueCount  int
		expected    int
		description string
	}{
		{
			name:        "zero issues",
			issueCount:  0,
			expected:    3,
			description: "minimum length for empty project",
		},
		{
			name:        "single issue",
			issueCount:  1,
			expected:    3,
			description: "minimum length",
		},
		{
			name:        "9 issues",
			issueCount:  9,
			expected:    3,
			description: "still at minimum length",
		},
		{
			name:        "10 issues",
			issueCount:  10,
			expected:    4,
			description: "crosses threshold to 4 chars",
		},
		{
			name:        "50 issues",
			issueCount:  50,
			expected:    4,
			description: "stays at 4 chars",
		},
		{
			name:        "100 issues",
			issueCount:  100,
			expected:    5,
			description: "medium project needs 5 chars",
		},
		{
			name:        "1000 issues",
			issueCount:  1000,
			expected:    6,
			description: "large project needs 6 chars",
		},
		{
			name:        "10000 issues",
			issueCount:  10000,
			expected:    7,
			description: "very large project needs 7 chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateIDLength(tt.issueCount)
			if result != tt.expected {
				t.Errorf("CalculateIDLength(%d) = %d, want %d (%s)",
					tt.issueCount, result, tt.expected, tt.description)
			}
		})
	}
}

func TestFormatID(t *testing.T) {
	tests := []struct {
		name            string
		id              string
		uniquePrefixLen int
		expected        string
	}{
		{
			name:            "normal case with extraneous chars",
			id:              "mint-abc1234",
			uniquePrefixLen: 6,
			expected:        "\033[4mmint-a\033[24m\033[38;5;8mbc1234\033[0m",
		},
		{
			name:            "unique prefix is full length",
			id:              "mint-xyz",
			uniquePrefixLen: 8,
			expected:        "\033[4mmint-xyz\033[0m",
		},
		{
			name:            "unique prefix exceeds length",
			id:              "mint-x",
			uniquePrefixLen: 10,
			expected:        "\033[4mmint-x\033[0m",
		},
		{
			name:            "unique prefix is 1",
			id:              "mint-abc",
			uniquePrefixLen: 1,
			expected:        "\033[4mm\033[24m\033[38;5;8mint-abc\033[0m",
		},
		{
			name:            "empty string",
			id:              "",
			uniquePrefixLen: 0,
			expected:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatID(tt.id, tt.uniquePrefixLen)
			if result != tt.expected {
				t.Errorf("FormatID(%q, %d) = %q, want %q", tt.id, tt.uniquePrefixLen, result, tt.expected)
			}
		})
	}
}
