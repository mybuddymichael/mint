package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewStore(t *testing.T) {
	store := NewStore()

	if store.Prefix != "mint" {
		t.Errorf("expected default prefix 'mint', got '%s'", store.Prefix)
	}

	if store.Issues == nil {
		t.Error("expected Issues map to be initialized")
	}
}

func TestStoreSaveAndLoad(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	// Create a store with some data
	store := NewStore()
	store.Prefix = "test-"
	store.Issues = map[string]*Issue{
		"test-abc1234": {
			ID:     "test-abc1234",
			Title:  "Test issue",
			Status: "open",
		},
	}

	// Save to file
	err := store.Save(filePath)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("expected file to exist after Save()")
	}

	// Load from file
	loadedStore, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("LoadStore() failed: %v", err)
	}

	// Verify prefix
	if loadedStore.Prefix != store.Prefix {
		t.Errorf("expected prefix '%s', got '%s'", store.Prefix, loadedStore.Prefix)
	}

	// Verify issues
	if len(loadedStore.Issues) != len(store.Issues) {
		t.Errorf("expected %d issues, got %d", len(store.Issues), len(loadedStore.Issues))
	}

	issue := loadedStore.Issues["test-abc1234"]
	if issue == nil {
		t.Fatal("expected issue 'test-abc1234' to exist")
		return
	}

	if issue.Title != "Test issue" {
		t.Errorf("expected title 'Test issue', got '%s'", issue.Title)
	}

	if issue.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", issue.Status)
	}
}

func TestLoadStoreNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.yaml")

	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("LoadStore() should not error on nonexistent file: %v", err)
	}

	// Should return new store with defaults
	if store.Prefix != "mint" {
		t.Errorf("expected default prefix 'mint', got '%s'", store.Prefix)
	}

	if len(store.Issues) != 0 {
		t.Errorf("expected empty issues map, got %d issues", len(store.Issues))
	}
}

func TestStoreSaveOrderDeterministic(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	// Create store with issues in non-alphabetical order
	store := NewStore()
	store.Issues = map[string]*Issue{
		"mint-zzz": {ID: "mint-zzz", Title: "Last", Status: "open"},
		"mint-aaa": {ID: "mint-aaa", Title: "First", Status: "open"},
		"mint-mmm": {ID: "mint-mmm", Title: "Middle", Status: "open"},
	}

	// Save multiple times
	err := store.Save(filePath)
	if err != nil {
		t.Fatalf("first Save() failed: %v", err)
	}

	// #nosec G304 -- filePath is constructed from t.TempDir() in test setup, safe test fixture path
	firstSave, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read first save: %v", err)
	}

	// Save again
	err = store.Save(filePath)
	if err != nil {
		t.Fatalf("second Save() failed: %v", err)
	}

	// #nosec G304 -- filePath is constructed from t.TempDir() in test setup, safe test fixture path
	secondSave, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read second save: %v", err)
	}

	// Both saves should be identical
	if !bytes.Equal(firstSave, secondSave) {
		t.Error("multiple saves produced different output")
		t.Logf("First:\n%s", firstSave)
		t.Logf("Second:\n%s", secondSave)
	}
}

func TestStoreSaveOrderSorted(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	// Create store with issues in non-alphabetical order
	store := NewStore()
	store.Issues = map[string]*Issue{
		"mint-zzz": {ID: "mint-zzz", Title: "Last", Status: "open"},
		"mint-aaa": {ID: "mint-aaa", Title: "First", Status: "open"},
		"mint-mmm": {ID: "mint-mmm", Title: "Middle", Status: "open"},
	}

	err := store.Save(filePath)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// #nosec G304 -- filePath is constructed from t.TempDir() in test setup, safe test fixture path
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)

	// Find positions of each issue ID in the file
	posAAA := strings.Index(contentStr, "mint-aaa")
	posMMM := strings.Index(contentStr, "mint-mmm")
	posZZZ := strings.Index(contentStr, "mint-zzz")

	if posAAA == -1 || posMMM == -1 || posZZZ == -1 {
		t.Fatal("not all issue IDs found in output")
	}

	// Verify they appear in alphabetical order
	if posAAA >= posMMM || posMMM >= posZZZ {
		t.Errorf("issues not in alphabetical order: aaa=%d, mmm=%d, zzz=%d", posAAA, posMMM, posZZZ)
		t.Logf("Content:\n%s", contentStr)
	}
}
