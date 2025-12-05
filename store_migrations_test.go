package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Framework tests - test migration system mechanics

func TestRunMigrations_IncrementsVersion(t *testing.T) {
	store := &Store{
		SchemaVersion: 0,
		Prefix:        "test",
		Issues:        make(map[string]*Issue),
	}

	migrated, err := RunMigrations(store)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	if !migrated {
		t.Error("expected migrated=true when starting from v0")
	}

	if store.SchemaVersion <= 0 {
		t.Errorf("version should have incremented from 0, got %d", store.SchemaVersion)
	}
}

func TestRunMigrations_AlreadyCurrent(t *testing.T) {
	store := &Store{
		SchemaVersion: CurrentSchemaVersion,
		Prefix:        "test",
		Issues:        make(map[string]*Issue),
	}

	migrated, err := RunMigrations(store)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	if migrated {
		t.Error("expected migrated=false for current version")
	}

	if store.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("version should remain %d", CurrentSchemaVersion)
	}
}

func TestNewStore_HasCurrentVersion(t *testing.T) {
	store := NewStore()

	if store.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("new store should have version %d, got %d", CurrentSchemaVersion, store.SchemaVersion)
	}
}

func TestLoadStore_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.yaml")

	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("LoadStore failed: %v", err)
	}

	if store.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("expected version %d, got %d", CurrentSchemaVersion, store.SchemaVersion)
	}
}

func TestLoadStore_CurrentVersionNoMigration(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	// Create file with current version
	store := NewStore()
	if _, err := store.AddIssue("Test issue"); err != nil {
		t.Fatalf("Failed to add issue: %v", err)
	}
	if err := store.Save(filePath); err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}

	// Load should not re-save
	origData, _ := os.ReadFile(filePath)

	loadedStore, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("LoadStore failed: %v", err)
	}

	if loadedStore.SchemaVersion != CurrentSchemaVersion {
		t.Errorf("expected version %d, got %d", CurrentSchemaVersion, loadedStore.SchemaVersion)
	}

	// File should be unchanged
	newData, _ := os.ReadFile(filePath)
	if string(origData) != string(newData) {
		t.Error("file should not be modified when no migration needed")
	}
}

// Migration-specific tests - test each migration with hardcoded versions

func TestMigration_V0ToV1_AddsSchemaVersion(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	// Create v0 file (no schema_version field)
	v0Data := `prefix: mint
issues:
  mint-abc:
    id: mint-abc
    title: Test issue
    status: open
  mint-xyz:
    id: mint-xyz
    title: Another issue
    status: closed
    comments:
      - "This is a comment"
`
	if err := os.WriteFile(filePath, []byte(v0Data), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load should migrate from v0 to at least v1
	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("LoadStore failed: %v", err)
	}

	if store.SchemaVersion < 1 {
		t.Errorf("expected version >= 1 after migrating from v0, got %d", store.SchemaVersion)
	}

	// Verify data preserved
	if len(store.Issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(store.Issues))
	}

	if store.Prefix != "mint" {
		t.Errorf("expected prefix 'mint', got '%s'", store.Prefix)
	}

	// Verify issue data intact
	issue := store.Issues["mint-abc"]
	if issue == nil {
		t.Fatal("issue mint-abc not found")
	}
	if issue.Title != "Test issue" {
		t.Errorf("title not preserved")
	}
	if issue.Status != "open" {
		t.Errorf("status not preserved")
	}

	issueXyz := store.Issues["mint-xyz"]
	if issueXyz == nil {
		t.Fatal("issue mint-xyz not found")
	}
	if len(issueXyz.Comments) != 1 {
		t.Errorf("comments not preserved")
	}

	// Verify file was updated on disk with schema_version
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(data), "schema_version:") {
		t.Error("schema_version not found in saved file")
	}

	// Second load should not migrate
	store2, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("second LoadStore failed: %v", err)
	}

	if store2.SchemaVersion != store.SchemaVersion {
		t.Error("version changed on second load")
	}

	if len(store2.Issues) != 2 {
		t.Errorf("issues lost after migration")
	}
}

func TestMigration_V0ToV1_PreservesRelationships(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	// Create v0 file with dependencies and blocks
	v0Data := `prefix: test
issues:
  test-foo:
    id: test-foo
    title: Foo
    status: open
    blocks:
      - test-bar
  test-bar:
    id: test-bar
    title: Bar
    status: open
    depends_on:
      - test-foo
`
	if err := os.WriteFile(filePath, []byte(v0Data), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("LoadStore failed: %v", err)
	}

	if store.SchemaVersion < 1 {
		t.Errorf("expected version >= 1, got %d", store.SchemaVersion)
	}

	// Verify relationships preserved
	foo := store.Issues["test-foo"]
	if foo == nil {
		t.Fatal("issue test-foo not found")
	}
	if len(foo.Blocks) != 1 || foo.Blocks[0] != "test-bar" {
		t.Error("blocks relationship not preserved")
	}

	bar := store.Issues["test-bar"]
	if bar == nil {
		t.Fatal("issue test-bar not found")
	}
	if len(bar.DependsOn) != 1 || bar.DependsOn[0] != "test-foo" {
		t.Error("depends_on relationship not preserved")
	}
}
