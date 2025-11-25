package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreSetPrefix(t *testing.T) {
	store := NewStore()
	store.Prefix = "old"
	store.Issues = map[string]*Issue{
		"old-abc123": {ID: "old-abc123", Title: "First", Status: "open"},
		"old-def456": {ID: "old-def456", Title: "Second", Status: "open"},
	}

	err := store.SetPrefix("new")
	if err != nil {
		t.Fatalf("SetPrefix() failed: %v", err)
	}

	if store.Prefix != "new" {
		t.Errorf("expected prefix 'new', got '%s'", store.Prefix)
	}

	// Check all IDs were updated
	if len(store.Issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(store.Issues))
	}

	if _, exists := store.Issues["new-abc123"]; !exists {
		t.Error("expected 'new-abc123' to exist")
	}

	if _, exists := store.Issues["new-def456"]; !exists {
		t.Error("expected 'new-def456' to exist")
	}

	// Check issue IDs were updated internally
	if store.Issues["new-abc123"].ID != "new-abc123" {
		t.Errorf("expected issue ID 'new-abc123', got '%s'", store.Issues["new-abc123"].ID)
	}
}

func TestStoreSetPrefix_UpdatesReferences(t *testing.T) {
	store := NewStore()
	store.Prefix = "old"
	store.Issues = map[string]*Issue{
		"old-abc123": {
			ID:        "old-abc123",
			Title:     "First",
			Status:    "open",
			DependsOn: []string{"old-def456"},
		},
		"old-def456": {
			ID:     "old-def456",
			Title:  "Second",
			Status: "open",
			Blocks: []string{"old-abc123"},
		},
	}

	err := store.SetPrefix("new")
	if err != nil {
		t.Fatalf("SetPrefix() failed: %v", err)
	}

	// Check DependsOn was updated
	if len(store.Issues["new-abc123"].DependsOn) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(store.Issues["new-abc123"].DependsOn))
	}
	if store.Issues["new-abc123"].DependsOn[0] != "new-def456" {
		t.Errorf("expected dependency 'new-def456', got '%s'", store.Issues["new-abc123"].DependsOn[0])
	}

	// Check Blocks was updated
	if len(store.Issues["new-def456"].Blocks) != 1 {
		t.Fatalf("expected 1 blocker, got %d", len(store.Issues["new-def456"].Blocks))
	}
	if store.Issues["new-def456"].Blocks[0] != "new-abc123" {
		t.Errorf("expected blocker 'new-abc123', got '%s'", store.Issues["new-def456"].Blocks[0])
	}
}

func TestStoreSetPrefix_NormalizesHyphen(t *testing.T) {
	store := NewStore()
	store.Prefix = "old"
	store.Issues = map[string]*Issue{
		"old-abc123": {ID: "old-abc123", Title: "First", Status: "open"},
	}

	// Prefix without hyphen should be stored without hyphen
	err := store.SetPrefix("mint")
	if err != nil {
		t.Fatalf("SetPrefix() failed: %v", err)
	}

	if store.Prefix != "mint" {
		t.Errorf("expected prefix 'mint', got '%s'", store.Prefix)
	}

	// But IDs should have hyphen separator
	if _, exists := store.Issues["mint-abc123"]; !exists {
		t.Error("expected 'mint-abc123' to exist")
	}
}

func TestStoreSetPrefix_StripsTrailingHyphen(t *testing.T) {
	store := NewStore()
	store.Prefix = "old"
	store.Issues = map[string]*Issue{
		"old-abc123": {ID: "old-abc123", Title: "First", Status: "open"},
	}

	// Prefix with hyphen should have hyphen stripped
	err := store.SetPrefix("mint-")
	if err != nil {
		t.Fatalf("SetPrefix() failed: %v", err)
	}

	if store.Prefix != "mint" {
		t.Errorf("expected prefix 'mint', got '%s'", store.Prefix)
	}

	if _, exists := store.Issues["mint-abc123"]; !exists {
		t.Error("expected 'mint-abc123' to exist")
	}
}

func TestStoreSetPrefix_OldPrefixWithHyphen(t *testing.T) {
	store := NewStore()
	// Simulate old state where prefix was stored WITH trailing hyphen
	store.Prefix = "mt-"
	store.Issues = map[string]*Issue{
		"mt-0abc123": {
			ID:     "mt-0abc123",
			Title:  "First",
			Status: "open",
		},
		"mt-9def456": {
			ID:        "mt-9def456",
			Title:     "Second",
			Status:    "open",
			DependsOn: []string{"mt-0abc123"},
		},
	}

	err := store.SetPrefix("mint")
	if err != nil {
		t.Fatalf("SetPrefix() failed: %v", err)
	}

	if store.Prefix != "mint" {
		t.Errorf("expected prefix 'mint', got '%s'", store.Prefix)
	}

	// Verify nanoid is preserved (7 chars: 0abc123, NOT 6 chars: abc123)
	if _, exists := store.Issues["mint-0abc123"]; !exists {
		t.Error("expected 'mint-0abc123' to exist (nanoid should preserve leading '0')")
	}

	if _, exists := store.Issues["mint-9def456"]; !exists {
		t.Error("expected 'mint-9def456' to exist (nanoid should preserve leading '9')")
	}

	// Verify the nanoid length is preserved
	issue1 := store.Issues["mint-0abc123"]
	if issue1 != nil {
		expectedLength := len("mint-0abc123")
		actualLength := len(issue1.ID)
		if actualLength != expectedLength {
			t.Errorf("expected ID length %d, got %d (ID: %s)", expectedLength, actualLength, issue1.ID)
		}
	}

	// Verify dependencies were updated correctly
	issue2 := store.Issues["mint-9def456"]
	if issue2 != nil && len(issue2.DependsOn) > 0 {
		if issue2.DependsOn[0] != "mint-0abc123" {
			t.Errorf("expected dependency 'mint-0abc123', got '%s'", issue2.DependsOn[0])
		}
	}
}

func TestGetStoreFilePath_EnvVar(t *testing.T) {
	expectedPath := "/custom/path/to/mint-issues.yaml"
	t.Setenv("MINT_STORE_FILE", expectedPath)

	path, err := GetStoreFilePath()
	if err != nil {
		t.Fatalf("GetStoreFilePath() failed: %v", err)
	}

	if path != expectedPath {
		t.Errorf("expected path '%s', got '%s'", expectedPath, path)
	}
}

func TestGetStoreFilePath_GitInCurrentDir(t *testing.T) {
	t.Setenv("MINT_STORE_FILE", "")

	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.Mkdir(gitDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Save and restore cwd
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() { _ = os.Chdir(origCwd) }()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	path, err := GetStoreFilePath()
	if err != nil {
		t.Fatalf("GetStoreFilePath() failed: %v", err)
	}

	// Resolve symlinks on directory for comparison (macOS /var -> /private/var)
	resolvedTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}
	expectedPath := filepath.Join(resolvedTmpDir, "mint-issues.yaml")

	if path != expectedPath {
		t.Errorf("expected path '%s', got '%s'", expectedPath, path)
	}
}

func TestGetStoreFilePath_GitInParent(t *testing.T) {
	t.Setenv("MINT_STORE_FILE", "")

	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.Mkdir(gitDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	childDir := filepath.Join(tmpDir, "child")
	err = os.Mkdir(childDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create child dir: %v", err)
	}

	// Save and restore cwd
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() { _ = os.Chdir(origCwd) }()

	err = os.Chdir(childDir)
	if err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	path, err := GetStoreFilePath()
	if err != nil {
		t.Fatalf("GetStoreFilePath() failed: %v", err)
	}

	// Resolve symlinks on directory for comparison (macOS /var -> /private/var)
	resolvedTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}
	expectedPath := filepath.Join(resolvedTmpDir, "mint-issues.yaml")

	if path != expectedPath {
		t.Errorf("expected path '%s', got '%s'", expectedPath, path)
	}
}

func TestGetStoreFilePath_GitMultipleLevelsUp(t *testing.T) {
	t.Setenv("MINT_STORE_FILE", "")

	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.Mkdir(gitDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create nested directory structure: root/a/b/c
	deepDir := filepath.Join(tmpDir, "a", "b", "c")
	err = os.MkdirAll(deepDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create deep dir: %v", err)
	}

	// Save and restore cwd
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() { _ = os.Chdir(origCwd) }()

	err = os.Chdir(deepDir)
	if err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	path, err := GetStoreFilePath()
	if err != nil {
		t.Fatalf("GetStoreFilePath() failed: %v", err)
	}

	// Resolve symlinks on directory for comparison (macOS /var -> /private/var)
	resolvedTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}
	expectedPath := filepath.Join(resolvedTmpDir, "mint-issues.yaml")

	if path != expectedPath {
		t.Errorf("expected path '%s', got '%s'", expectedPath, path)
	}
}

func TestGetStoreFilePath_NoGitFallback(t *testing.T) {
	t.Setenv("MINT_STORE_FILE", "")

	tmpDir := t.TempDir()

	// Save and restore cwd
	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() { _ = os.Chdir(origCwd) }()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	path, err := GetStoreFilePath()
	if err != nil {
		t.Fatalf("GetStoreFilePath() failed: %v", err)
	}

	// Resolve symlinks on directory for comparison (macOS /var -> /private/var)
	resolvedTmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}
	expectedPath := filepath.Join(resolvedTmpDir, "mint-issues.yaml")

	if path != expectedPath {
		t.Errorf("expected path '%s', got '%s'", expectedPath, path)
	}
}
