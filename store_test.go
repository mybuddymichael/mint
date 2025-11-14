package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore(t *testing.T) {
	store := NewStore()

	if store.Prefix != "mt-" {
		t.Errorf("expected default prefix 'mt-', got '%s'", store.Prefix)
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
	if store.Prefix != "mt-" {
		t.Errorf("expected default prefix 'mt-', got '%s'", store.Prefix)
	}

	if len(store.Issues) != 0 {
		t.Errorf("expected empty issues map, got %d issues", len(store.Issues))
	}
}
