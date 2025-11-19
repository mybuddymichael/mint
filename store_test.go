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

	firstSave, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read first save: %v", err)
	}

	// Save again
	err = store.Save(filePath)
	if err != nil {
		t.Fatalf("second Save() failed: %v", err)
	}

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

func TestStoreAddIssue(t *testing.T) {
	store := NewStore()

	issue, err := store.AddIssue("Test issue")
	if err != nil {
		t.Fatalf("AddIssue() failed: %v", err)
	}

	if issue == nil {
		t.Fatal("AddIssue() returned nil issue")
		return
	}

	if issue.Title != "Test issue" {
		t.Errorf("expected title 'Test issue', got '%s'", issue.Title)
	}

	if issue.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", issue.Status)
	}

	if !strings.HasPrefix(issue.ID, store.Prefix) {
		t.Errorf("expected ID to have prefix '%s', got '%s'", store.Prefix, issue.ID)
	}

	// Verify issue was added to store
	if store.Issues[issue.ID] != issue {
		t.Error("issue not found in store")
	}
}

func TestStoreAddIssueUnique(t *testing.T) {
	store := NewStore()

	// Add multiple issues and verify IDs are unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		issue, err := store.AddIssue("Test issue")
		if err != nil {
			t.Fatalf("AddIssue() failed on iteration %d: %v", i, err)
		}

		if ids[issue.ID] {
			t.Errorf("duplicate ID generated: %s", issue.ID)
		}
		ids[issue.ID] = true
	}

	if len(store.Issues) != 100 {
		t.Errorf("expected 100 issues in store, got %d", len(store.Issues))
	}
}

func TestStoreCloseIssue(t *testing.T) {
	store := NewStore()
	issue, err := store.AddIssue("Test issue")
	if err != nil {
		t.Fatalf("AddIssue() failed: %v", err)
	}

	err = store.CloseIssue(issue.ID, "")
	if err != nil {
		t.Fatalf("CloseIssue() failed: %v", err)
	}

	if issue.Status != "closed" {
		t.Errorf("expected status 'closed', got '%s'", issue.Status)
	}

	if len(issue.Comments) != 0 {
		t.Errorf("expected no comments when closing without reason, got %d", len(issue.Comments))
	}
}

func TestStoreCloseIssue_WithReason(t *testing.T) {
	store := NewStore()
	issue, err := store.AddIssue("Test issue")
	if err != nil {
		t.Fatalf("AddIssue() failed: %v", err)
	}

	err = store.CloseIssue(issue.ID, "Done")
	if err != nil {
		t.Fatalf("CloseIssue() failed: %v", err)
	}

	if issue.Status != "closed" {
		t.Errorf("expected status 'closed', got '%s'", issue.Status)
	}

	if len(issue.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(issue.Comments))
	}

	expectedComment := "Closed with reason: Done"
	if issue.Comments[0] != expectedComment {
		t.Errorf("expected comment '%s', got '%s'", expectedComment, issue.Comments[0])
	}
}

func TestStoreCloseIssue_NotFound(t *testing.T) {
	store := NewStore()

	err := store.CloseIssue("mint-nonexistent", "")
	if err == nil {
		t.Fatal("expected error when closing nonexistent issue")
	}

	expectedErr := "issue not found: mint-nonexistent"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestStoreReopenIssue(t *testing.T) {
	store := NewStore()
	issue, err := store.AddIssue("Test issue")
	if err != nil {
		t.Fatalf("AddIssue() failed: %v", err)
	}

	// Close it first
	err = store.CloseIssue(issue.ID, "")
	if err != nil {
		t.Fatalf("CloseIssue() failed: %v", err)
	}

	// Now reopen
	err = store.ReopenIssue(issue.ID)
	if err != nil {
		t.Fatalf("ReopenIssue() failed: %v", err)
	}

	if issue.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", issue.Status)
	}
}

func TestStoreReopenIssue_NotFound(t *testing.T) {
	store := NewStore()

	err := store.ReopenIssue("mint-nonexistent")
	if err == nil {
		t.Fatal("expected error when reopening nonexistent issue")
	}

	expectedErr := "issue not found: mint-nonexistent"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestStoreResolveIssueID_ExactMatch(t *testing.T) {
	store := NewStore()
	store.Issues = map[string]*Issue{
		"mint-abc123": {ID: "mint-abc123", Title: "First", Status: "open"},
		"mint-def456": {ID: "mint-def456", Title: "Second", Status: "open"},
	}

	id, err := store.ResolveIssueID("mint-abc123")
	if err != nil {
		t.Fatalf("ResolveIssueID() failed: %v", err)
	}

	if id != "mint-abc123" {
		t.Errorf("expected 'mint-abc123', got '%s'", id)
	}
}

func TestStoreResolveIssueID_UniquePrefix(t *testing.T) {
	store := NewStore()
	store.Issues = map[string]*Issue{
		"mint-abc123": {ID: "mint-abc123", Title: "First", Status: "open"},
		"mint-def456": {ID: "mint-def456", Title: "Second", Status: "open"},
	}

	id, err := store.ResolveIssueID("mint-a")
	if err != nil {
		t.Fatalf("ResolveIssueID() failed: %v", err)
	}

	if id != "mint-abc123" {
		t.Errorf("expected 'mint-abc123', got '%s'", id)
	}
}

func TestStoreResolveIssueID_AmbiguousPrefix(t *testing.T) {
	store := NewStore()
	store.Issues = map[string]*Issue{
		"mint-abc123": {ID: "mint-abc123", Title: "First", Status: "open"},
		"mint-abc456": {ID: "mint-abc456", Title: "Second", Status: "open"},
	}

	_, err := store.ResolveIssueID("mint-abc")
	if err == nil {
		t.Fatal("expected error for ambiguous prefix")
	}

	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("expected error to mention 'ambiguous', got '%s'", err.Error())
	}
}

func TestStoreResolveIssueID_NotFound(t *testing.T) {
	store := NewStore()
	store.Issues = map[string]*Issue{
		"mint-abc123": {ID: "mint-abc123", Title: "First", Status: "open"},
	}

	_, err := store.ResolveIssueID("mint-xyz")
	if err == nil {
		t.Fatal("expected error for not found")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to mention 'not found', got '%s'", err.Error())
	}
}

func TestStoreGetIssue_WithPrefix(t *testing.T) {
	store := NewStore()
	store.Issues = map[string]*Issue{
		"mint-abc123": {ID: "mint-abc123", Title: "First", Status: "open"},
		"mint-def456": {ID: "mint-def456", Title: "Second", Status: "open"},
	}

	issue, err := store.GetIssue("mint-a")
	if err != nil {
		t.Fatalf("GetIssue() failed: %v", err)
	}

	if issue.ID != "mint-abc123" {
		t.Errorf("expected ID 'mint-abc123', got '%s'", issue.ID)
	}
}

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

func TestStoreDeleteIssue(t *testing.T) {
	store := NewStore()
	issue, err := store.AddIssue("Test issue")
	if err != nil {
		t.Fatalf("AddIssue() failed: %v", err)
	}

	err = store.DeleteIssue(issue.ID)
	if err != nil {
		t.Fatalf("DeleteIssue() failed: %v", err)
	}

	if _, exists := store.Issues[issue.ID]; exists {
		t.Error("issue should be deleted from store")
	}
}

func TestStoreDeleteIssue_CleansReferences(t *testing.T) {
	store := NewStore()
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	issue3, _ := store.AddIssue("Issue 3")

	// Set up dependencies: issue2 depends on issue1, issue1 blocks issue3
	_ = store.AddDependency(issue2.ID, issue1.ID)
	_ = store.AddBlocker(issue1.ID, issue3.ID)

	// Delete issue1
	err := store.DeleteIssue(issue1.ID)
	if err != nil {
		t.Fatalf("DeleteIssue() failed: %v", err)
	}

	// issue1 should be gone
	if _, exists := store.Issues[issue1.ID]; exists {
		t.Error("issue1 should be deleted from store")
	}

	// issue2's DependsOn should be cleaned
	if len(issue2.DependsOn) != 0 {
		t.Errorf("issue2 should have no dependencies, got %d", len(issue2.DependsOn))
	}

	// issue3's Blocks should be cleaned
	if len(issue3.Blocks) != 0 {
		t.Errorf("issue3 should have no blockers, got %d", len(issue3.Blocks))
	}
}

func TestStoreDeleteIssue_NotFound(t *testing.T) {
	store := NewStore()

	err := store.DeleteIssue("mint-nonexistent")
	if err == nil {
		t.Fatal("expected error when deleting nonexistent issue")
	}

	expectedErr := "issue not found: mint-nonexistent"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}
