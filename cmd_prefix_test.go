package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetPrefixCommand_NoPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("First issue")
	oldID1 := issue1.ID
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	// No prefix arg means set to empty string
	err := cmd.Run(context.Background(), []string{"mint", "set-prefix"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	if store.Prefix != "" {
		t.Errorf("expected empty prefix, got '%s'", store.Prefix)
	}

	// Verify issue ID was updated to remove prefix
	expectedNewID := oldID1[len("mint")+1:]
	if _, exists := store.Issues[expectedNewID]; !exists {
		t.Errorf("expected new ID %s to exist", expectedNewID)
	}
}

func TestSetPrefixCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("First issue")
	issue2, _ := store.AddIssue("Second issue")
	oldID1 := issue1.ID
	oldID2 := issue2.ID
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", "app"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	if store.Prefix != "app" {
		t.Errorf("expected prefix 'app', got '%s'", store.Prefix)
	}

	// Compute new IDs
	newID1 := "app-" + oldID1[len("mint")+1:]
	newID2 := "app-" + oldID2[len("mint")+1:]

	if _, exists := store.Issues[oldID1]; exists {
		t.Error("old issue ID should not exist after prefix change")
	}
	if _, exists := store.Issues[oldID2]; exists {
		t.Error("old issue ID should not exist after prefix change")
	}
	if _, exists := store.Issues[newID1]; !exists {
		t.Errorf("expected new ID %s to exist", newID1)
	}
	if _, exists := store.Issues[newID2]; !exists {
		t.Errorf("expected new ID %s to exist", newID2)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Prefix set to \"app\" and all issues updated") {
		t.Errorf("expected success message, got: %s", output)
	}
}

func TestSetPrefixCommand_EmptyStore(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", "newprefix"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	if store.Prefix != "newprefix" {
		t.Errorf("expected prefix 'newprefix', got '%s'", store.Prefix)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Prefix set to \"newprefix\" and all issues updated") {
		t.Errorf("expected success message, got: %s", output)
	}
}

func TestSetPrefixCommand_OutputFormat(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	_, _ = store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", "myprefix"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	output := stripANSI(buf.String())
	expected := "Prefix set to \"myprefix\" and all issues updated\n"
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
}

func TestSetPrefixCommand_StripsTrailingHyphen(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	_, _ = store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", "app-"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	if store.Prefix != "app" {
		t.Errorf("expected prefix 'app' (normalized), got '%s'", store.Prefix)
	}

	for id := range store.Issues {
		if !strings.HasPrefix(id, "app-") {
			t.Errorf("expected issue ID to start with 'app-', got '%s'", id)
		}
		if strings.HasPrefix(id, "app--") {
			t.Errorf("expected single hyphen separator, got double hyphen in '%s'", id)
		}
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Prefix set to \"app\"") {
		t.Errorf("expected normalized prefix in output, got: %s", output)
	}
}

func TestSetPrefixCommand_MultipleIssuesVariousStatuses(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	_, _ = store.AddIssue("Open issue")
	issue2, _ := store.AddIssue("Closed issue")
	_, _ = store.AddIssue("Ready issue")
	_ = store.CloseIssue(issue2.ID, "Done")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", "new"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	if len(store.Issues) != 3 {
		t.Errorf("expected 3 issues, got %d", len(store.Issues))
	}

	for id := range store.Issues {
		if !strings.HasPrefix(id, "new-") {
			t.Errorf("expected all IDs to have 'new-' prefix, got '%s'", id)
		}
	}

	openCount := 0
	closedCount := 0
	for _, issue := range store.Issues {
		switch issue.Status {
		case "open":
			openCount++
		case "closed":
			closedCount++
		}
	}
	if openCount != 2 {
		t.Errorf("expected 2 open issues, got %d", openCount)
	}
	if closedCount != 1 {
		t.Errorf("expected 1 closed issue, got %d", closedCount)
	}
}

func TestSetPrefixCommand_PreserveRelationships(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	oldID1 := issue1.ID
	oldID2 := issue2.ID
	_ = store.AddDependency(issue1.ID, issue2.ID)
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", "new"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)

	newID1 := "new-" + oldID1[len("mint")+1:]
	newID2 := "new-" + oldID2[len("mint")+1:]

	newIssue1 := store.Issues[newID1]
	newIssue2 := store.Issues[newID2]

	if newIssue1 == nil || newIssue2 == nil {
		t.Fatal("could not find issues after prefix change")
	}

	if len(newIssue1.DependsOn) != 1 {
		t.Fatalf("expected issue1 to have 1 dependency, got %d", len(newIssue1.DependsOn))
	}
	if newIssue1.DependsOn[0] != newID2 {
		t.Errorf("expected issue1 to depend on %s, got %s", newID2, newIssue1.DependsOn[0])
	}

	if len(newIssue2.Blocks) != 1 {
		t.Fatalf("expected issue2 to block 1 issue, got %d", len(newIssue2.Blocks))
	}
	if newIssue2.Blocks[0] != newID1 {
		t.Errorf("expected issue2 to block %s, got %s", newID1, newIssue2.Blocks[0])
	}
}

func TestSetPrefixCommand_PreserveComplexRelationships(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	issue3, _ := store.AddIssue("Issue 3")
	oldID1 := issue1.ID
	oldID2 := issue2.ID
	oldID3 := issue3.ID
	_ = store.AddDependency(issue1.ID, issue2.ID)
	_ = store.AddDependency(issue1.ID, issue3.ID)
	_ = store.AddBlocker(issue2.ID, issue3.ID)
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", "new"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)

	newID1 := "new-" + oldID1[len("mint")+1:]
	newID2 := "new-" + oldID2[len("mint")+1:]
	newID3 := "new-" + oldID3[len("mint")+1:]

	newIssue1 := store.Issues[newID1]
	newIssue2 := store.Issues[newID2]
	newIssue3 := store.Issues[newID3]

	if newIssue1 == nil || newIssue2 == nil || newIssue3 == nil {
		t.Fatal("could not find all issues after prefix change")
	}

	if len(newIssue1.DependsOn) != 2 {
		t.Errorf("expected issue1 to have 2 dependencies, got %d", len(newIssue1.DependsOn))
	}

	// issue2 blocks both issue1 (from dependency) and issue3 (from explicit blocker)
	if len(newIssue2.Blocks) != 2 {
		t.Errorf("expected issue2 to block 2 issues, got %d", len(newIssue2.Blocks))
	}

	if len(newIssue3.DependsOn) != 1 || newIssue3.DependsOn[0] != newID2 {
		t.Errorf("expected issue3 to depend on issue2")
	}
}

func TestSetPrefixCommand_PreservesIssueProperties(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	oldID := issue.ID
	_ = store.AddComment(issue.ID, "First comment")
	_ = store.AddComment(issue.ID, "Second comment")
	_ = store.CloseIssue(issue.ID, "Done")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", "new"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)

	newID := "new-" + oldID[len("mint")+1:]
	newIssue := store.Issues[newID]

	if newIssue == nil {
		t.Fatal("could not find issue after prefix change")
	}

	if newIssue.Title != "Test issue" {
		t.Errorf("expected title 'Test issue', got '%s'", newIssue.Title)
	}
	if newIssue.Status != "closed" {
		t.Errorf("expected status 'closed', got '%s'", newIssue.Status)
	}
	if len(newIssue.Comments) != 3 {
		t.Errorf("expected 3 comments, got %d", len(newIssue.Comments))
	}
	if newIssue.Comments[0] != "First comment" {
		t.Errorf("expected first comment preserved, got '%s'", newIssue.Comments[0])
	}
}

func TestSetPrefixCommand_CreatesFileIfNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	// Don't create file - it shouldn't exist
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatal("file should not exist before test")
	}

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", "newapp"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("file should be created by set-prefix command")
	}

	store, _ := LoadStore(filePath)
	if store.Prefix != "newapp" {
		t.Errorf("expected prefix 'newapp', got '%s'", store.Prefix)
	}
}

func TestSetPrefixCommand_ChangeMultipleTimes(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	oldID := issue.ID
	_ = store.Save(filePath)

	// First change: mint -> app1
	cmd1 := newCommand()
	var buf1 bytes.Buffer
	cmd1.Writer = &buf1

	err := cmd1.Run(context.Background(), []string{"mint", "set-prefix", "app1"})
	if err != nil {
		t.Fatalf("first set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	if store.Prefix != "app1" {
		t.Errorf("expected prefix 'app1', got '%s'", store.Prefix)
	}

	intermediateID := "app1-" + oldID[len("mint")+1:]
	if _, exists := store.Issues[intermediateID]; !exists {
		t.Errorf("expected issue with ID %s after first change", intermediateID)
	}

	// Second change: app1 -> app2
	cmd2 := newCommand()
	var buf2 bytes.Buffer
	cmd2.Writer = &buf2

	err = cmd2.Run(context.Background(), []string{"mint", "set-prefix", "app2"})
	if err != nil {
		t.Fatalf("second set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	if store.Prefix != "app2" {
		t.Errorf("expected prefix 'app2', got '%s'", store.Prefix)
	}

	finalID := "app2-" + oldID[len("mint")+1:]
	if _, exists := store.Issues[finalID]; !exists {
		t.Errorf("expected issue with ID %s after second change", finalID)
	}

	if _, exists := store.Issues[oldID]; exists {
		t.Errorf("old ID %s should not exist", oldID)
	}
	if _, exists := store.Issues[intermediateID]; exists {
		t.Errorf("intermediate ID %s should not exist", intermediateID)
	}
}

func TestSetPrefixCommand_SetToEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("First issue")
	oldID1 := issue1.ID // Will be "mint-xyz"
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", ""})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	if store.Prefix != "" {
		t.Errorf("expected empty prefix, got '%s'", store.Prefix)
	}

	// Extract nanoid from old ID (after "mint-")
	expectedNewID := oldID1[len("mint")+1:]

	if _, exists := store.Issues[oldID1]; exists {
		t.Error("old issue ID should not exist after prefix change to empty")
	}
	if _, exists := store.Issues[expectedNewID]; !exists {
		t.Errorf("expected new ID %s to exist", expectedNewID)
	}

	// Verify no hyphens in new ID
	for id := range store.Issues {
		if strings.Contains(id, "-") {
			t.Errorf("expected no hyphens in IDs with empty prefix, got '%s'", id)
		}
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Prefix set to \"\" and all issues updated") {
		t.Errorf("expected success message for empty prefix, got: %s", output)
	}
}

func TestSetPrefixCommand_FromEmptyToNonEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	// Create store with empty prefix
	store := NewStore()
	store.Prefix = ""
	_ = store.Save(filePath)

	// Manually add issue with no prefix
	store.Issues = map[string]*Issue{
		"abc123": {ID: "abc123", Title: "Test", Status: "open"},
	}
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix", "app"})
	if err != nil {
		t.Fatalf("set-prefix command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	if store.Prefix != "app" {
		t.Errorf("expected prefix 'app', got '%s'", store.Prefix)
	}

	if _, exists := store.Issues["abc123"]; exists {
		t.Error("old ID 'abc123' should not exist")
	}
	if _, exists := store.Issues["app-abc123"]; !exists {
		t.Error("expected 'app-abc123' to exist")
	}
}
