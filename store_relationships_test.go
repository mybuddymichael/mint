package main

import (
	"testing"
	"time"
)

func TestStoreRemoveDependency(t *testing.T) {
	store := NewStore()
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")

	// Add dependency: issue2 depends on issue1
	err := store.AddDependency(issue2.ID, issue1.ID)
	if err != nil {
		t.Fatalf("AddDependency() failed: %v", err)
	}

	// Verify it was added
	if len(issue2.DependsOn) != 1 || issue2.DependsOn[0] != issue1.ID {
		t.Fatal("dependency not added correctly")
	}
	if len(issue1.Blocks) != 1 || issue1.Blocks[0] != issue2.ID {
		t.Fatal("blocks relationship not added correctly")
	}

	// Remove the dependency
	err = store.RemoveDependency(issue2.ID, issue1.ID)
	if err != nil {
		t.Fatalf("RemoveDependency() failed: %v", err)
	}

	// Verify both sides were cleaned
	if len(issue2.DependsOn) != 0 {
		t.Errorf("expected issue2 to have no dependencies, got %d", len(issue2.DependsOn))
	}
	if len(issue1.Blocks) != 0 {
		t.Errorf("expected issue1 to have no blocks, got %d", len(issue1.Blocks))
	}
}

func TestStoreRemoveDependency_IssueNotFound(t *testing.T) {
	store := NewStore()
	issue, _ := store.AddIssue("Issue")

	err := store.RemoveDependency("mint-nonexistent", issue.ID)
	if err == nil {
		t.Fatal("expected error when removing dependency with nonexistent issue")
	}
}

func TestStoreRemoveDependency_DependencyNotFound(t *testing.T) {
	store := NewStore()
	issue, _ := store.AddIssue("Issue")

	err := store.RemoveDependency(issue.ID, "mint-nonexistent")
	if err == nil {
		t.Fatal("expected error when removing nonexistent dependency")
	}
}

func TestStoreRemoveDependency_MultipleRelationships(t *testing.T) {
	store := NewStore()
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	issue3, _ := store.AddIssue("Issue 3")

	// issue2 depends on both issue1 and issue3
	_ = store.AddDependency(issue2.ID, issue1.ID)
	_ = store.AddDependency(issue2.ID, issue3.ID)

	// Remove only the dependency on issue1
	err := store.RemoveDependency(issue2.ID, issue1.ID)
	if err != nil {
		t.Fatalf("RemoveDependency() failed: %v", err)
	}

	// Verify only issue1 relationship was removed
	if len(issue2.DependsOn) != 1 {
		t.Fatalf("expected issue2 to have 1 dependency, got %d", len(issue2.DependsOn))
	}
	if issue2.DependsOn[0] != issue3.ID {
		t.Errorf("expected remaining dependency to be %s, got %s", issue3.ID, issue2.DependsOn[0])
	}
	if len(issue1.Blocks) != 0 {
		t.Errorf("expected issue1 to have no blocks, got %d", len(issue1.Blocks))
	}
	if len(issue3.Blocks) != 1 {
		t.Errorf("expected issue3 to still have 1 block, got %d", len(issue3.Blocks))
	}
}

func TestStoreRemoveBlocker(t *testing.T) {
	store := NewStore()
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")

	// Add blocker: issue1 blocks issue2
	err := store.AddBlocker(issue1.ID, issue2.ID)
	if err != nil {
		t.Fatalf("AddBlocker() failed: %v", err)
	}

	// Verify it was added
	if len(issue1.Blocks) != 1 || issue1.Blocks[0] != issue2.ID {
		t.Fatal("blocker not added correctly")
	}
	if len(issue2.DependsOn) != 1 || issue2.DependsOn[0] != issue1.ID {
		t.Fatal("dependency relationship not added correctly")
	}

	// Remove the blocker
	err = store.RemoveBlocker(issue1.ID, issue2.ID)
	if err != nil {
		t.Fatalf("RemoveBlocker() failed: %v", err)
	}

	// Verify both sides were cleaned
	if len(issue1.Blocks) != 0 {
		t.Errorf("expected issue1 to have no blocks, got %d", len(issue1.Blocks))
	}
	if len(issue2.DependsOn) != 0 {
		t.Errorf("expected issue2 to have no dependencies, got %d", len(issue2.DependsOn))
	}
}

func TestStoreRemoveBlocker_IssueNotFound(t *testing.T) {
	store := NewStore()
	issue, _ := store.AddIssue("Issue")

	err := store.RemoveBlocker("mint-nonexistent", issue.ID)
	if err == nil {
		t.Fatal("expected error when removing blocker with nonexistent issue")
	}
}

func TestStoreRemoveBlocker_BlockedNotFound(t *testing.T) {
	store := NewStore()
	issue, _ := store.AddIssue("Issue")

	err := store.RemoveBlocker(issue.ID, "mint-nonexistent")
	if err == nil {
		t.Fatal("expected error when removing nonexistent blocked issue")
	}
}

func TestStoreRemoveBlocker_MultipleRelationships(t *testing.T) {
	store := NewStore()
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	issue3, _ := store.AddIssue("Issue 3")

	// issue1 blocks both issue2 and issue3
	_ = store.AddBlocker(issue1.ID, issue2.ID)
	_ = store.AddBlocker(issue1.ID, issue3.ID)

	// Remove only the block on issue2
	err := store.RemoveBlocker(issue1.ID, issue2.ID)
	if err != nil {
		t.Fatalf("RemoveBlocker() failed: %v", err)
	}

	// Verify only issue2 relationship was removed
	if len(issue1.Blocks) != 1 {
		t.Fatalf("expected issue1 to have 1 block, got %d", len(issue1.Blocks))
	}
	if issue1.Blocks[0] != issue3.ID {
		t.Errorf("expected remaining block to be %s, got %s", issue3.ID, issue1.Blocks[0])
	}
	if len(issue2.DependsOn) != 0 {
		t.Errorf("expected issue2 to have no dependencies, got %d", len(issue2.DependsOn))
	}
	if len(issue3.DependsOn) != 1 {
		t.Errorf("expected issue3 to still have 1 dependency, got %d", len(issue3.DependsOn))
	}
}

func TestStoreAddDependency_UpdatesBothTimestamps(t *testing.T) {
	store := NewStore()
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")

	original1 := issue1.UpdatedAt
	original2 := issue2.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	// issue2 depends on issue1
	err := store.AddDependency(issue2.ID, issue1.ID)
	if err != nil {
		t.Fatalf("AddDependency() failed: %v", err)
	}

	if !issue1.UpdatedAt.After(original1) {
		t.Error("issue1 (blocker) UpdatedAt should be updated")
	}

	if !issue2.UpdatedAt.After(original2) {
		t.Error("issue2 (dependent) UpdatedAt should be updated")
	}
}

func TestStoreAddBlocker_UpdatesBothTimestamps(t *testing.T) {
	store := NewStore()
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")

	original1 := issue1.UpdatedAt
	original2 := issue2.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	// issue1 blocks issue2
	err := store.AddBlocker(issue1.ID, issue2.ID)
	if err != nil {
		t.Fatalf("AddBlocker() failed: %v", err)
	}

	if !issue1.UpdatedAt.After(original1) {
		t.Error("issue1 (blocker) UpdatedAt should be updated")
	}

	if !issue2.UpdatedAt.After(original2) {
		t.Error("issue2 (blocked) UpdatedAt should be updated")
	}
}

func TestStoreRemoveDependency_UpdatesBothTimestamps(t *testing.T) {
	store := NewStore()
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	_ = store.AddDependency(issue2.ID, issue1.ID)

	original1 := issue1.UpdatedAt
	original2 := issue2.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	err := store.RemoveDependency(issue2.ID, issue1.ID)
	if err != nil {
		t.Fatalf("RemoveDependency() failed: %v", err)
	}

	if !issue1.UpdatedAt.After(original1) {
		t.Error("issue1 UpdatedAt should be updated")
	}

	if !issue2.UpdatedAt.After(original2) {
		t.Error("issue2 UpdatedAt should be updated")
	}
}

func TestStoreRemoveBlocker_UpdatesBothTimestamps(t *testing.T) {
	store := NewStore()
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	_ = store.AddBlocker(issue1.ID, issue2.ID)

	original1 := issue1.UpdatedAt
	original2 := issue2.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	err := store.RemoveBlocker(issue1.ID, issue2.ID)
	if err != nil {
		t.Fatalf("RemoveBlocker() failed: %v", err)
	}

	if !issue1.UpdatedAt.After(original1) {
		t.Error("issue1 UpdatedAt should be updated")
	}

	if !issue2.UpdatedAt.After(original2) {
		t.Error("issue2 UpdatedAt should be updated")
	}
}
