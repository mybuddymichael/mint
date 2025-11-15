package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandName(t *testing.T) {
	cmd := newCommand()
	if cmd.Name != "mt" {
		t.Errorf("expected command name 'mt', got '%s'", cmd.Name)
	}
}

func TestCommandRuns(t *testing.T) {
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt"})
	if err != nil {
		t.Errorf("cmd.Run() returned error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected output, got none")
	}
}

func TestCommandHelp(t *testing.T) {
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "--help"})
	if err != nil {
		t.Errorf("cmd.Run() with --help returned error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected help output, got none")
	}
}

func TestCreateCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	// Set the store file path for this test
	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mt", "create", "Test issue"})
	if err != nil {
		t.Fatalf("create command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created issue mt-") {
		t.Errorf("expected output to contain 'Created issue mt-', got: %s", output)
	}

	// Verify the issue was saved
	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	if len(store.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(store.Issues))
	}

	// Find the issue
	var issue *Issue
	for _, iss := range store.Issues {
		issue = iss
		break
	}

	if issue == nil {
		t.Fatal("expected to find an issue")
		return
	}

	if issue.Title != "Test issue" {
		t.Errorf("expected title 'Test issue', got '%s'", issue.Title)
	}

	if issue.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", issue.Status)
	}

	if !strings.HasPrefix(issue.ID, "mt-") {
		t.Errorf("expected ID to start with 'mt-', got '%s'", issue.ID)
	}
}

func TestCreateCommandNoTitle(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mt", "create"})
	if err == nil {
		t.Error("expected error when no title provided")
	}
}

func TestShowCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	t.Setenv("MINT_STORE_FILE", filePath)

	// Create issue directly via Store
	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	issue, err := store.AddIssue("Test show issue")
	if err != nil {
		t.Fatalf("failed to add issue: %v", err)
	}

	if err := store.Save(filePath); err != nil {
		t.Fatalf("failed to save store: %v", err)
	}

	// Test show command
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"mt", "show", issue.ID})
	if err != nil {
		t.Fatalf("show command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ID:      "+issue.ID) {
		t.Errorf("expected output to contain 'ID:      %s', got: %s", issue.ID, output)
	}
	if !strings.Contains(output, "Title:   \"Test show issue\"") {
		t.Errorf("expected output to contain 'Title:   \"Test show issue\"', got: %s", output)
	}
	if !strings.Contains(output, "Status:  open") {
		t.Errorf("expected output to contain 'Status:  open', got: %s", output)
	}
}

func TestShowCommandNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mt", "show", "mt-nonexistent"})
	if err == nil {
		t.Error("expected error when showing non-existent issue")
	}
}

func TestListCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	t.Setenv("MINT_STORE_FILE", filePath)

	// Create issues directly via Store
	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	issue1, err := store.AddIssue("First issue")
	if err != nil {
		t.Fatalf("failed to add issue: %v", err)
	}

	issue2, err := store.AddIssue("Second issue")
	if err != nil {
		t.Fatalf("failed to add issue: %v", err)
	}

	issue3, err := store.AddIssue("Third issue")
	if err != nil {
		t.Fatalf("failed to add issue: %v", err)
	}

	// Close issue2 to test both open and closed statuses
	if err := store.CloseIssue(issue2.ID, ""); err != nil {
		t.Fatalf("failed to close issue: %v", err)
	}

	if err := store.Save(filePath); err != nil {
		t.Fatalf("failed to save store: %v", err)
	}

	// Test list command
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"mt", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "All issues:") {
		t.Errorf("expected output to contain 'All issues:', got: %s", output)
	}

	if !strings.Contains(output, issue1.ID+" open \"First issue\"") {
		t.Errorf("expected output to contain issue1 with open status, got: %s", output)
	}
	if !strings.Contains(output, issue2.ID+" closed \"Second issue\"") {
		t.Errorf("expected output to contain issue2 with closed status, got: %s", output)
	}
	if !strings.Contains(output, issue3.ID+" open \"Third issue\"") {
		t.Errorf("expected output to contain issue3 with open status, got: %s", output)
	}

	// Verify issues are sorted by ID
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines of output, got %d", len(lines))
	}

	// Extract IDs from output lines (skip "All issues:" header)
	var outputIDs []string
	for i := 1; i < len(lines); i++ {
		parts := strings.SplitN(lines[i], " ", 2)
		if len(parts) > 0 {
			outputIDs = append(outputIDs, parts[0])
		}
	}

	// Verify sorted
	for i := 1; i < len(outputIDs); i++ {
		if outputIDs[i-1] >= outputIDs[i] {
			t.Errorf("IDs not sorted: %s >= %s", outputIDs[i-1], outputIDs[i])
		}
	}
}

func TestListCommandEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	t.Setenv("MINT_STORE_FILE", filePath)

	// Create empty store
	store := NewStore()
	if err := store.Save(filePath); err != nil {
		t.Fatalf("failed to save empty store: %v", err)
	}

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "list"})
	if err != nil {
		t.Fatalf("list command failed on empty store: %v", err)
	}

	output := buf.String()
	expected := "No issues found.\n"
	if output != expected {
		t.Errorf("expected %q, got: %q", expected, output)
	}
}

func TestListCommandNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	t.Setenv("MINT_STORE_FILE", filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "list"})
	if err != nil {
		t.Fatalf("list command failed when no file exists: %v", err)
	}

	output := buf.String()
	expected := "No issues file found.\n"
	if output != expected {
		t.Errorf("expected %q, got: %q", expected, output)
	}
}

func TestUpdateCommandTitle(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Original title")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "update", issue.ID, "--title", "New title"})
	if err != nil {
		t.Fatalf("update command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ := store.GetIssue(issue.ID)

	if updated.Title != "New title" {
		t.Errorf("expected title 'New title', got '%s'", updated.Title)
	}
}

func TestUpdateCommandDependsOn(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "update", issue1.ID, "--depends-on", issue2.ID})
	if err != nil {
		t.Fatalf("update command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ := store.GetIssue(issue1.ID)

	if len(updated.DependsOn) != 1 || updated.DependsOn[0] != issue2.ID {
		t.Errorf("expected DependsOn [%s], got %v", issue2.ID, updated.DependsOn)
	}

	blocker, _ := store.GetIssue(issue2.ID)
	if len(blocker.Blocks) != 1 || blocker.Blocks[0] != issue1.ID {
		t.Errorf("expected Blocks [%s], got %v", issue1.ID, blocker.Blocks)
	}
}

func TestUpdateCommandBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	issue3, _ := store.AddIssue("Issue 3")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "update", issue1.ID, "--blocks", issue2.ID, "--blocks", issue3.ID})
	if err != nil {
		t.Fatalf("update command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ := store.GetIssue(issue1.ID)

	if len(updated.Blocks) != 2 {
		t.Errorf("expected 2 blocks, got %d", len(updated.Blocks))
	}

	blocked2, _ := store.GetIssue(issue2.ID)
	if len(blocked2.DependsOn) != 1 || blocked2.DependsOn[0] != issue1.ID {
		t.Errorf("expected issue2 DependsOn [%s], got %v", issue1.ID, blocked2.DependsOn)
	}

	blocked3, _ := store.GetIssue(issue3.ID)
	if len(blocked3.DependsOn) != 1 || blocked3.DependsOn[0] != issue1.ID {
		t.Errorf("expected issue3 DependsOn [%s], got %v", issue1.ID, blocked3.DependsOn)
	}
}

func TestUpdateCommandComment(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "update", issue.ID, "--comment", "Test comment"})
	if err != nil {
		t.Fatalf("update command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ := store.GetIssue(issue.ID)

	if len(updated.Comments) != 1 || updated.Comments[0] != "Test comment" {
		t.Errorf("expected Comments ['Test comment'], got %v", updated.Comments)
	}
}

func TestUpdateCommandMultipleFlags(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "update", issue1.ID, "--title", "Updated", "--depends-on", issue2.ID, "--comment", "Done"})
	if err != nil {
		t.Fatalf("update command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ := store.GetIssue(issue1.ID)

	if updated.Title != "Updated" {
		t.Errorf("expected title 'Updated', got '%s'", updated.Title)
	}
	if len(updated.DependsOn) != 1 || updated.DependsOn[0] != issue2.ID {
		t.Errorf("expected DependsOn [%s], got %v", issue2.ID, updated.DependsOn)
	}
	if len(updated.Comments) != 1 || updated.Comments[0] != "Done" {
		t.Errorf("expected Comments ['Done'], got %v", updated.Comments)
	}
}

func TestUpdateCommandNoID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "update", "--title", "New title"})
	if err == nil {
		t.Error("expected error when no issue ID provided")
	}
}

func TestUpdateCommandInvalidID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "update", "mt-invalid", "--title", "New title"})
	if err == nil {
		t.Error("expected error when updating non-existent issue")
	}
}

func TestCloseCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "close", issue.ID})
	if err != nil {
		t.Fatalf("close command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	closed, _ := store.GetIssue(issue.ID)

	if closed.Status != "closed" {
		t.Errorf("expected status 'closed', got '%s'", closed.Status)
	}

	if len(closed.Comments) != 0 {
		t.Errorf("expected no comments when closing without reason, got %d", len(closed.Comments))
	}

	output := buf.String()
	if !strings.Contains(output, "Closed issue "+issue.ID) {
		t.Errorf("expected output to contain 'Closed issue %s', got: %s", issue.ID, output)
	}
}

func TestCloseCommand_WithReason(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "close", issue.ID, "--reason", "Done"})
	if err != nil {
		t.Fatalf("close command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	closed, _ := store.GetIssue(issue.ID)

	if closed.Status != "closed" {
		t.Errorf("expected status 'closed', got '%s'", closed.Status)
	}

	if len(closed.Comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(closed.Comments))
	}

	expectedComment := "Closed with reason: Done"
	if closed.Comments[0] != expectedComment {
		t.Errorf("expected comment '%s', got '%s'", expectedComment, closed.Comments[0])
	}

	output := buf.String()
	if !strings.Contains(output, "Closed issue "+issue.ID) {
		t.Errorf("expected output to contain 'Closed issue %s', got: %s", issue.ID, output)
	}
}

func TestCloseCommand_NoID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "close"})
	if err == nil {
		t.Error("expected error when no issue ID provided")
	}
}

func TestCloseCommand_InvalidID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "close", "mt-invalid"})
	if err == nil {
		t.Error("expected error when closing non-existent issue")
	}
}

func TestOpenCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.CloseIssue(issue.ID, "")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "open", issue.ID})
	if err != nil {
		t.Fatalf("open command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	reopened, _ := store.GetIssue(issue.ID)

	if reopened.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", reopened.Status)
	}

	output := buf.String()
	if !strings.Contains(output, "Re-opened issue "+issue.ID) {
		t.Errorf("expected output to contain 'Re-opened issue %s', got: %s", issue.ID, output)
	}
}

func TestOpenCommand_NoID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "open"})
	if err == nil {
		t.Error("expected error when no issue ID provided")
	}
}

func TestOpenCommand_InvalidID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt", "open", "mt-invalid"})
	if err == nil {
		t.Error("expected error when opening non-existent issue")
	}
}
