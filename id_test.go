package main

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	prefix := "mt-"
	id := GenerateID(prefix)

	// Check format: prefix + 7 characters
	if !strings.HasPrefix(id, prefix) {
		t.Errorf("expected ID to start with '%s', got '%s'", prefix, id)
	}

	expectedLen := len(prefix) + 7
	if len(id) != expectedLen {
		t.Errorf("expected ID length %d, got %d (ID: %s)", expectedLen, len(id), id)
	}
}

func TestGenerateIDCustomAlphabet(t *testing.T) {
	prefix := "mt-"
	id := GenerateID(prefix)

	// Extract the nano ID part (after prefix)
	nanoID := strings.TrimPrefix(id, prefix)

	// Verify only lowercase letters and numbers
	allowedChars := "abcdefghijklmnopqrstuvwxyz0123456789"
	for _, char := range nanoID {
		if !strings.ContainsRune(allowedChars, char) {
			t.Errorf("ID contains invalid character '%c', ID: %s", char, id)
		}
	}
}

func TestGenerateIDUniqueness(t *testing.T) {
	prefix := "mt-"
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
	prefix := "test-"
	id := GenerateID(prefix)

	if !strings.HasPrefix(id, prefix) {
		t.Errorf("expected ID to start with '%s', got '%s'", prefix, id)
	}

	expectedLen := len(prefix) + 7
	if len(id) != expectedLen {
		t.Errorf("expected ID length %d, got %d", expectedLen, len(id))
	}
}
