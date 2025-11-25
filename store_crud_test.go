package main

import (
	"strings"
	"testing"
)

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
