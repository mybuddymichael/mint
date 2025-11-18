package main

import (
	"bytes"
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func TestCommandName(t *testing.T) {
	cmd := newCommand()
	if cmd.Name != "mint" {
		t.Errorf("expected command name 'mint', got '%s'", cmd.Name)
	}
}

func TestCommandRuns(t *testing.T) {
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint"})
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

	err := cmd.Run(context.Background(), []string{"mint", "--help"})
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

	err := cmd.Run(context.Background(), []string{"mint", "create", "Test issue"})
	if err != nil {
		t.Fatalf("create command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Created issue mint-") {
		t.Errorf("expected output to contain 'Created issue mint-', got: %s", output)
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

	if !strings.HasPrefix(issue.ID, "mint-") {
		t.Errorf("expected ID to start with 'mint-', got '%s'", issue.ID)
	}
}

func TestCreateCommandNoTitle(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mint", "create"})
	if err == nil {
		t.Error("expected error when no title provided")
	}
}

func TestAddCommandAlias(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mint", "add", "Test issue"})
	if err != nil {
		t.Fatalf("add alias command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Created issue mint-") {
		t.Errorf("expected output to contain 'Created issue mint-', got: %s", output)
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

	err = cmd.Run(context.Background(), []string{"mint", "show", issue.ID})
	if err != nil {
		t.Fatalf("show command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "ID:      "+issue.ID) {
		t.Errorf("expected output to contain 'ID:      %s', got: %s", issue.ID, output)
	}
	if !strings.Contains(output, "Title:   Test show issue") {
		t.Errorf("expected output to contain 'Title:   Test show issue', got: %s", output)
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

	err := cmd.Run(context.Background(), []string{"mint", "show", "mint-nonexistent"})
	if err == nil {
		t.Error("expected error when showing non-existent issue")
	}
}

func TestShowCommandWithComments(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue with comments")
	_ = store.AddComment(issue.ID, "First comment")
	_ = store.AddComment(issue.ID, "Second comment")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "show", issue.ID})
	if err != nil {
		t.Fatalf("show command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Comments:") {
		t.Errorf("expected output to contain 'Comments:', got: %s", output)
	}
	if !strings.Contains(output, "  First comment") {
		t.Errorf("expected output to contain '  First comment', got: %s", output)
	}
	if !strings.Contains(output, "  Second comment") {
		t.Errorf("expected output to contain '  Second comment', got: %s", output)
	}
}

func TestShowCommandNoComments(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue without comments")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "show", issue.ID})
	if err != nil {
		t.Fatalf("show command failed: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Comments:") {
		t.Errorf("expected output NOT to contain 'Comments:' when no comments, got: %s", output)
	}
}

func TestShowCommandWithRelationships(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Main issue")
	issue2, _ := store.AddIssue("Dependency issue")
	issue3, _ := store.AddIssue("Blocked issue")
	_ = store.AddDependency(issue1.ID, issue2.ID)
	_ = store.AddBlocker(issue1.ID, issue3.ID)
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "show", issue1.ID})
	if err != nil {
		t.Fatalf("show command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Depends on:") {
		t.Errorf("expected output to contain 'Depends on:', got: %s", output)
	}
	if !strings.Contains(output, "  "+issue2.ID+" Dependency issue") {
		t.Errorf("expected output to contain '  %s Dependency issue', got: %s", issue2.ID, output)
	}
	if !strings.Contains(output, "Blocks:") {
		t.Errorf("expected output to contain 'Blocks:', got: %s", output)
	}
	if !strings.Contains(output, "  "+issue3.ID+" Blocked issue") {
		t.Errorf("expected output to contain '  %s Blocked issue', got: %s", issue3.ID, output)
	}
}

func TestShowCommandNoRelationships(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Standalone issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "show", issue.ID})
	if err != nil {
		t.Fatalf("show command failed: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Depends on:") {
		t.Errorf("expected output NOT to contain 'Depends on:' when no dependencies, got: %s", output)
	}
	if strings.Contains(output, "Blocks:") {
		t.Errorf("expected output NOT to contain 'Blocks:' when no blocks, got: %s", output)
	}
}

func TestShowCommandWithRelationshipsAndComments(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Main issue")
	issue2, _ := store.AddIssue("Dependency")
	issue3, _ := store.AddIssue("Blocked")
	_ = store.AddDependency(issue1.ID, issue2.ID)
	_ = store.AddBlocker(issue1.ID, issue3.ID)
	_ = store.AddComment(issue1.ID, "Test comment")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "show", issue1.ID})
	if err != nil {
		t.Fatalf("show command failed: %v", err)
	}

	output := buf.String()

	dependsIdx := strings.Index(output, "Depends on:")
	blocksIdx := strings.Index(output, "Blocks:")
	commentsIdx := strings.Index(output, "Comments:")

	if dependsIdx == -1 {
		t.Error("expected output to contain 'Depends on:'")
	}
	if blocksIdx == -1 {
		t.Error("expected output to contain 'Blocks:'")
	}
	if commentsIdx == -1 {
		t.Error("expected output to contain 'Comments:'")
	}

	if commentsIdx != -1 && dependsIdx != -1 && commentsIdx < dependsIdx {
		t.Errorf("Comments section should appear after Depends on section")
	}
	if commentsIdx != -1 && blocksIdx != -1 && commentsIdx < blocksIdx {
		t.Errorf("Comments section should appear after Blocks section")
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

	err = cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := stripANSI(buf.String())

	if !strings.Contains(output, "   "+issue1.ID+" open First issue") {
		t.Errorf("expected output to contain issue1 with open status and 3-space indent, got: %s", output)
	}
	if !strings.Contains(output, "   "+issue2.ID+" closed Second issue") {
		t.Errorf("expected output to contain issue2 with closed status and 3-space indent, got: %s", output)
	}
	if !strings.Contains(output, "   "+issue3.ID+" open Third issue") {
		t.Errorf("expected output to contain issue3 with open status and 3-space indent, got: %s", output)
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

	err := cmd.Run(context.Background(), []string{"mint", "list"})
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

	err := cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed when no file exists: %v", err)
	}

	output := buf.String()
	expected := "No issues file found.\n"
	if output != expected {
		t.Errorf("expected %q, got: %q", expected, output)
	}
}

func TestListCommandWithOpenFlag(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Open issue 1")
	issue2, _ := store.AddIssue("Closed issue")
	issue3, _ := store.AddIssue("Open issue 2")
	_ = store.CloseIssue(issue2.ID, "")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--open"})
	if err != nil {
		t.Fatalf("list --open command failed: %v", err)
	}

	output := buf.String()
	strippedOutput := stripANSI(output)

	if !strings.Contains(output, " OPEN ") {
		t.Errorf("expected output to contain ' OPEN ' header, got: %s", output)
	}
	if !strings.Contains(strippedOutput, "   "+issue1.ID) {
		t.Errorf("expected output to contain open issue1 %s with 3-space indent, got: %s", issue1.ID, strippedOutput)
	}
	if strings.Contains(strippedOutput, issue2.ID) {
		t.Errorf("expected output NOT to contain closed issue2 %s, got: %s", issue2.ID, strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   "+issue3.ID) {
		t.Errorf("expected output to contain open issue3 %s with 3-space indent, got: %s", issue3.ID, strippedOutput)
	}
}

func TestListCommandWithOpenFlagEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Closed issue")
	_ = store.CloseIssue(issue.ID, "")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--open"})
	if err != nil {
		t.Fatalf("list --open command failed: %v", err)
	}

	output := buf.String()
	strippedOutput := stripANSI(output)

	if !strings.Contains(output, " OPEN ") {
		t.Errorf("expected output to contain ' OPEN ' header, got: %s", output)
	}
	if !strings.Contains(strippedOutput, "   (No open issues.)") {
		t.Errorf("expected output to contain '   (No open issues.)', got: %s", strippedOutput)
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

	err := cmd.Run(context.Background(), []string{"mint", "update", issue.ID, "--title", "New title"})
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

	err := cmd.Run(context.Background(), []string{"mint", "update", issue1.ID, "--depends-on", issue2.ID})
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

	err := cmd.Run(context.Background(), []string{"mint", "update", issue1.ID, "--blocks", issue2.ID, "--blocks", issue3.ID})
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

	err := cmd.Run(context.Background(), []string{"mint", "update", issue.ID, "--comment", "Test comment"})
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

	err := cmd.Run(context.Background(), []string{"mint", "update", issue1.ID, "--title", "Updated", "--depends-on", issue2.ID, "--comment", "Done"})
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

	err := cmd.Run(context.Background(), []string{"mint", "update", "--title", "New title"})
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

	err := cmd.Run(context.Background(), []string{"mint", "update", "mint-invalid", "--title", "New title"})
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

	err := cmd.Run(context.Background(), []string{"mint", "close", issue.ID})
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

	output := stripANSI(buf.String())
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

	err := cmd.Run(context.Background(), []string{"mint", "close", issue.ID, "--reason", "Done"})
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

	output := stripANSI(buf.String())
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

	err := cmd.Run(context.Background(), []string{"mint", "close"})
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

	err := cmd.Run(context.Background(), []string{"mint", "close", "mint-invalid"})
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

	err := cmd.Run(context.Background(), []string{"mint", "open", issue.ID})
	if err != nil {
		t.Fatalf("open command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	reopened, _ := store.GetIssue(issue.ID)

	if reopened.Status != "open" {
		t.Errorf("expected status 'open', got '%s'", reopened.Status)
	}

	output := stripANSI(buf.String())
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

	err := cmd.Run(context.Background(), []string{"mint", "open"})
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

	err := cmd.Run(context.Background(), []string{"mint", "open", "mint-invalid"})
	if err == nil {
		t.Error("expected error when opening non-existent issue")
	}
}

func TestSetPrefixCommand_NoPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "set-prefix"})
	if err == nil {
		t.Error("expected error when no prefix provided")
	}
}

func TestUpdateCommand_PartialID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	// Use partial ID (first 8 chars)
	partialID := issue.ID[:8]
	err := cmd.Run(context.Background(), []string{"mint", "update", partialID, "--title", "Updated"})
	if err != nil {
		t.Fatalf("update command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Updated "+issue.ID) {
		t.Errorf("expected output to contain 'Updated %s' (full ID), got: %s", issue.ID, output)
	}
}

func TestCloseCommand_PartialID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	// Use partial ID (first 8 chars)
	partialID := issue.ID[:8]
	err := cmd.Run(context.Background(), []string{"mint", "close", partialID})
	if err != nil {
		t.Fatalf("close command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Closed issue "+issue.ID) {
		t.Errorf("expected output to contain 'Closed issue %s' (full ID), got: %s", issue.ID, output)
	}
}

func TestOpenCommand_PartialID(t *testing.T) {
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

	// Use partial ID (first 8 chars)
	partialID := issue.ID[:8]
	err := cmd.Run(context.Background(), []string{"mint", "open", partialID})
	if err != nil {
		t.Fatalf("open command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Re-opened issue "+issue.ID) {
		t.Errorf("expected output to contain 'Re-opened issue %s' (full ID), got: %s", issue.ID, output)
	}
}

func TestListCommandGroupedByStatus(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Open issue 1")
	issue2, _ := store.AddIssue("Closed issue 1")
	issue3, _ := store.AddIssue("Open issue 2")
	issue4, _ := store.AddIssue("Closed issue 2")
	_ = store.CloseIssue(issue2.ID, "")
	_ = store.CloseIssue(issue4.ID, "")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := buf.String()
	strippedOutput := stripANSI(output)

	// Check for OPEN header
	if !strings.Contains(output, " OPEN ") {
		t.Error("expected output to contain ' OPEN ' header")
	}

	// Check for CLOSED header
	if !strings.Contains(output, " CLOSED ") {
		t.Error("expected output to contain ' CLOSED ' header")
	}

	// Check that open issues are under OPEN section with 3-space indent
	if !strings.Contains(strippedOutput, "   "+issue1.ID) {
		t.Errorf("expected open issue1 to be indented with 3 spaces, got: %s", strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   "+issue3.ID) {
		t.Errorf("expected open issue3 to be indented with 3 spaces, got: %s", strippedOutput)
	}

	// Check that closed issues are under CLOSED section with 3-space indent
	if !strings.Contains(strippedOutput, "   "+issue2.ID) {
		t.Errorf("expected closed issue2 to be indented with 3 spaces, got: %s", strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   "+issue4.ID) {
		t.Errorf("expected closed issue4 to be indented with 3 spaces, got: %s", strippedOutput)
	}

	// Verify order: OPEN section comes before CLOSED section
	openIdx := strings.Index(strippedOutput, "OPEN")
	closedIdx := strings.Index(strippedOutput, "CLOSED")
	if openIdx == -1 || closedIdx == -1 || openIdx >= closedIdx {
		t.Error("expected OPEN section to appear before CLOSED section")
	}
}

func TestListCommandEmptyOpenSection(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Closed issue")
	_ = store.CloseIssue(issue.ID, "")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := buf.String()
	strippedOutput := stripANSI(output)

	// Check for OPEN header
	if !strings.Contains(output, " OPEN ") {
		t.Error("expected output to contain ' OPEN ' header even when empty")
	}

	// Check for "(No open issues.)" message with 3-space indent
	if !strings.Contains(strippedOutput, "   (No open issues.)") {
		t.Errorf("expected '   (No open issues.)' when no open issues, got: %s", strippedOutput)
	}

	// Check that closed issue is shown
	if !strings.Contains(strippedOutput, "   "+issue.ID) {
		t.Errorf("expected closed issue to be shown, got: %s", strippedOutput)
	}
}

func TestListCommandEmptyClosedSection(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Open issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := buf.String()
	strippedOutput := stripANSI(output)

	// Check for CLOSED header
	if !strings.Contains(output, " CLOSED ") {
		t.Error("expected output to contain ' CLOSED ' header even when empty")
	}

	// Check for "(No closed issues.)" message with 3-space indent
	if !strings.Contains(strippedOutput, "   (No closed issues.)") {
		t.Errorf("expected '   (No closed issues.)' when no closed issues, got: %s", strippedOutput)
	}

	// Check that open issue is shown
	if !strings.Contains(strippedOutput, "   "+issue.ID) {
		t.Errorf("expected open issue to be shown, got: %s", strippedOutput)
	}
}

func TestListCommandGroupedWithOpenFlag(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Open issue")
	issue2, _ := store.AddIssue("Closed issue")
	_ = store.CloseIssue(issue2.ID, "")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--open"})
	if err != nil {
		t.Fatalf("list --open command failed: %v", err)
	}

	output := buf.String()
	strippedOutput := stripANSI(output)

	// Should show OPEN header
	if !strings.Contains(output, " OPEN ") {
		t.Error("expected output to contain ' OPEN ' header")
	}

	// Should NOT show CLOSED header
	if strings.Contains(output, " CLOSED ") {
		t.Error("expected output NOT to contain ' CLOSED ' header when using --open flag")
	}

	// Should show open issue
	if !strings.Contains(strippedOutput, "   "+issue1.ID) {
		t.Errorf("expected open issue to be shown with 3-space indent, got: %s", strippedOutput)
	}

	// Should NOT show closed issue
	if strings.Contains(strippedOutput, issue2.ID) {
		t.Errorf("expected closed issue NOT to be shown, got: %s", strippedOutput)
	}
}

func TestListCommandGroupedWithOpenFlagEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Closed issue")
	_ = store.CloseIssue(issue.ID, "")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--open"})
	if err != nil {
		t.Fatalf("list --open command failed: %v", err)
	}

	output := buf.String()
	strippedOutput := stripANSI(output)

	// Should show OPEN header
	if !strings.Contains(output, " OPEN ") {
		t.Error("expected output to contain ' OPEN ' header")
	}

	// Should show "(No open issues.)"
	if !strings.Contains(strippedOutput, "   (No open issues.)") {
		t.Errorf("expected '   (No open issues.)' when --open flag used with no open issues, got: %s", strippedOutput)
	}
}

func TestListCommandGroupedSorting(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	// Create multiple open and closed issues
	_, _ = store.AddIssue("Open issue 1")
	issue2, _ := store.AddIssue("Closed issue 1")
	_, _ = store.AddIssue("Open issue 2")
	issue4, _ := store.AddIssue("Closed issue 2")
	_, _ = store.AddIssue("Open issue 3")
	_ = store.CloseIssue(issue2.ID, "")
	_ = store.CloseIssue(issue4.ID, "")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := stripANSI(buf.String())
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Find the OPEN and CLOSED section boundaries
	openIdx := -1
	closedIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "OPEN") {
			openIdx = i
		}
		if strings.Contains(line, "CLOSED") {
			closedIdx = i
			break
		}
	}

	if openIdx == -1 || closedIdx == -1 {
		t.Fatalf("could not find OPEN or CLOSED headers in output")
	}

	// Extract open issue IDs (between OPEN header and CLOSED header)
	var openIDs []string
	for i := openIdx + 1; i < closedIdx; i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strings.HasPrefix(line, "(") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				openIDs = append(openIDs, parts[0])
			}
		}
	}

	// Extract closed issue IDs (after CLOSED header)
	var closedIDs []string
	for i := closedIdx + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strings.HasPrefix(line, "(") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				closedIDs = append(closedIDs, parts[0])
			}
		}
	}

	// Verify open issues are sorted
	for i := 1; i < len(openIDs); i++ {
		if openIDs[i-1] >= openIDs[i] {
			t.Errorf("open issues not sorted: %s >= %s", openIDs[i-1], openIDs[i])
		}
	}

	// Verify closed issues are sorted
	for i := 1; i < len(closedIDs); i++ {
		if closedIDs[i-1] >= closedIDs[i] {
			t.Errorf("closed issues not sorted: %s >= %s", closedIDs[i-1], closedIDs[i])
		}
	}
}

func TestListCommandAlignment(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	// Manually create issues with different ID lengths to test alignment
	store.Issues["mint-a"] = &Issue{ID: "mint-a", Title: "Short ID", Status: "open"}
	store.Issues["mint-abc123"] = &Issue{ID: "mint-abc123", Title: "Medium ID", Status: "open"}
	store.Issues["mint-xyz"] = &Issue{ID: "mint-xyz", Title: "Another short", Status: "closed"}
	store.Issues["mint-longer123"] = &Issue{ID: "mint-longer123", Title: "Long ID", Status: "closed"}
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := stripANSI(buf.String())
	lines := strings.Split(output, "\n")

	// Find all issue lines and track where status words appear
	var statusColumns []int
	for _, line := range lines {
		if strings.HasPrefix(line, "   mint-") {
			openIdx := strings.Index(line, "open ")
			closedIdx := strings.Index(line, "closed ")
			if openIdx != -1 {
				statusColumns = append(statusColumns, openIdx)
			} else if closedIdx != -1 {
				statusColumns = append(statusColumns, closedIdx)
			}
		}
	}

	if len(statusColumns) == 0 {
		t.Fatal("no status columns found in output")
	}

	// Verify all status words start at the same column
	firstCol := statusColumns[0]
	for i, col := range statusColumns {
		if col != firstCol {
			t.Errorf("status word at index %d starts at column %d, expected %d\nOutput:\n%s", i, col, firstCol, output)
		}
	}
}
