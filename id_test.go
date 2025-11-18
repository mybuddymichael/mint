package main

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	prefix := "mint"
	id := GenerateID(prefix)

	// Check format: prefix + hyphen + 7 characters
	if !strings.HasPrefix(id, prefix+"-") {
		t.Errorf("expected ID to start with '%s-', got '%s'", prefix, id)
	}

	expectedLen := len(prefix) + 1 + 7 // prefix + hyphen + nanoid
	if len(id) != expectedLen {
		t.Errorf("expected ID length %d, got %d (ID: %s)", expectedLen, len(id), id)
	}
}

func TestGenerateIDCustomAlphabet(t *testing.T) {
	prefix := "mint"
	id := GenerateID(prefix)

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
	seen := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id := GenerateID(prefix)
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
	id := GenerateID(prefix)

	if !strings.HasPrefix(id, prefix+"-") {
		t.Errorf("expected ID to start with '%s-', got '%s'", prefix, id)
	}

	expectedLen := len(prefix) + 1 + 7 // prefix + hyphen + nanoid
	if len(id) != expectedLen {
		t.Errorf("expected ID length %d, got %d", expectedLen, len(id))
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
