package main

import (
	"testing"
	"time"
)

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

func TestStoreAddComment_UpdatesTimestamp(t *testing.T) {
	store := NewStore()
	issue, _ := store.AddIssue("Test")
	originalUpdated := issue.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	err := store.AddComment(issue.ID, "Comment")
	if err != nil {
		t.Fatalf("AddComment() failed: %v", err)
	}

	if !issue.UpdatedAt.After(originalUpdated) {
		t.Errorf("UpdatedAt should be after original")
	}
}

func TestStoreCloseIssue_UpdatesTimestamp(t *testing.T) {
	store := NewStore()
	issue, _ := store.AddIssue("Test")
	originalUpdated := issue.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	err := store.CloseIssue(issue.ID, "")
	if err != nil {
		t.Fatalf("CloseIssue() failed: %v", err)
	}

	if !issue.UpdatedAt.After(originalUpdated) {
		t.Errorf("UpdatedAt should be after original")
	}
}

func TestStoreReopenIssue_UpdatesTimestamp(t *testing.T) {
	store := NewStore()
	issue, _ := store.AddIssue("Test")
	_ = store.CloseIssue(issue.ID, "")
	originalUpdated := issue.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	err := store.ReopenIssue(issue.ID)
	if err != nil {
		t.Fatalf("ReopenIssue() failed: %v", err)
	}

	if !issue.UpdatedAt.After(originalUpdated) {
		t.Errorf("UpdatedAt should be after original")
	}
}
