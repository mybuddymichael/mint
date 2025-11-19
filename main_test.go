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
	if !strings.Contains(output, `"Test issue"`) {
		t.Errorf("expected output to contain issue title in quotes, got: %s", output)
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
	if !strings.Contains(output, `"Test issue"`) {
		t.Errorf("expected output to contain issue title in quotes, got: %s", output)
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
	readyIssue, _ := store.AddIssue("Ready issue")
	blockedIssue, _ := store.AddIssue("Blocked issue")
	blocker, _ := store.AddIssue("Blocker issue")
	closedIssue, _ := store.AddIssue("Closed issue")

	_ = store.AddDependency(blockedIssue.ID, blocker.ID)
	_ = store.CloseIssue(closedIssue.ID, "")
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

	// Should show READY and BLOCKED headers
	if !strings.Contains(output, " READY ") {
		t.Error("expected output to contain ' READY ' header")
	}
	if !strings.Contains(output, " BLOCKED ") {
		t.Error("expected output to contain ' BLOCKED ' header")
	}

	// Should NOT show CLOSED header
	if strings.Contains(output, " CLOSED ") {
		t.Error("expected output NOT to contain ' CLOSED ' header with --open flag")
	}

	// Should show ready issues
	if !strings.Contains(strippedOutput, "   "+readyIssue.ID) {
		t.Errorf("expected ready issue %s with 3-space indent, got: %s", readyIssue.ID, strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   "+blocker.ID) {
		t.Errorf("expected blocker (ready) issue %s with 3-space indent, got: %s", blocker.ID, strippedOutput)
	}

	// Should show blocked issue
	if !strings.Contains(strippedOutput, "   "+blockedIssue.ID) {
		t.Errorf("expected blocked issue %s with 3-space indent, got: %s", blockedIssue.ID, strippedOutput)
	}

	// Should NOT show closed issue
	if strings.Contains(strippedOutput, closedIssue.ID) {
		t.Errorf("expected output NOT to contain closed issue %s, got: %s", closedIssue.ID, strippedOutput)
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

	// Should show READY and BLOCKED headers
	if !strings.Contains(output, " READY ") {
		t.Error("expected output to contain ' READY ' header")
	}
	if !strings.Contains(output, " BLOCKED ") {
		t.Error("expected output to contain ' BLOCKED ' header")
	}

	// Should NOT show CLOSED header
	if strings.Contains(output, " CLOSED ") {
		t.Error("expected output NOT to contain ' CLOSED ' header with --open flag")
	}

	// Should show empty messages
	if !strings.Contains(strippedOutput, "   (No ready issues.)") {
		t.Errorf("expected '   (No ready issues.)', got: %s", strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   (No blocked issues.)") {
		t.Errorf("expected '   (No blocked issues.)', got: %s", strippedOutput)
	}
}

func TestUpdateCommandTitle(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Original title")
	_ = store.Save(filePath)

	// Test long flag
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "update", issue.ID, "--title", "New title"})
	if err != nil {
		t.Fatalf("update command with --title failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ := store.GetIssue(issue.ID)

	if updated.Title != "New title" {
		t.Errorf("expected title 'New title', got '%s'", updated.Title)
	}

	// Test short flag
	cmd2 := newCommand()
	var buf2 bytes.Buffer
	cmd2.Writer = &buf2

	err = cmd2.Run(context.Background(), []string{"mint", "update", issue.ID, "-t", "Newer title"})
	if err != nil {
		t.Fatalf("update command with -t failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ = store.GetIssue(issue.ID)

	if updated.Title != "Newer title" {
		t.Errorf("expected title 'Newer title', got '%s'", updated.Title)
	}
}

func TestUpdateCommandDependsOn(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	issue3, _ := store.AddIssue("Issue 3")
	_ = store.Save(filePath)

	// Test long flag
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "update", issue1.ID, "--depends-on", issue2.ID})
	if err != nil {
		t.Fatalf("update command with --depends-on failed: %v", err)
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

	// Test short flag
	cmd2 := newCommand()
	var buf2 bytes.Buffer
	cmd2.Writer = &buf2

	err = cmd2.Run(context.Background(), []string{"mint", "update", issue1.ID, "-d", issue3.ID})
	if err != nil {
		t.Fatalf("update command with -d failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ = store.GetIssue(issue1.ID)

	if len(updated.DependsOn) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(updated.DependsOn))
	}

	blocker3, _ := store.GetIssue(issue3.ID)
	if len(blocker3.Blocks) != 1 || blocker3.Blocks[0] != issue1.ID {
		t.Errorf("expected issue3 Blocks [%s], got %v", issue1.ID, blocker3.Blocks)
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
	issue4, _ := store.AddIssue("Issue 4")
	issue5, _ := store.AddIssue("Issue 5")
	_ = store.Save(filePath)

	// Test long flag
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "update", issue1.ID, "--blocks", issue2.ID, "--blocks", issue3.ID})
	if err != nil {
		t.Fatalf("update command with --blocks failed: %v", err)
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

	// Test short flag
	cmd2 := newCommand()
	var buf2 bytes.Buffer
	cmd2.Writer = &buf2

	err = cmd2.Run(context.Background(), []string{"mint", "update", issue1.ID, "-b", issue4.ID, "-b", issue5.ID})
	if err != nil {
		t.Fatalf("update command with -b failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ = store.GetIssue(issue1.ID)

	if len(updated.Blocks) != 4 {
		t.Errorf("expected 4 blocks total, got %d", len(updated.Blocks))
	}

	blocked4, _ := store.GetIssue(issue4.ID)
	if len(blocked4.DependsOn) != 1 || blocked4.DependsOn[0] != issue1.ID {
		t.Errorf("expected issue4 DependsOn [%s], got %v", issue1.ID, blocked4.DependsOn)
	}

	blocked5, _ := store.GetIssue(issue5.ID)
	if len(blocked5.DependsOn) != 1 || blocked5.DependsOn[0] != issue1.ID {
		t.Errorf("expected issue5 DependsOn [%s], got %v", issue1.ID, blocked5.DependsOn)
	}
}

func TestUpdateCommandComment(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.Save(filePath)

	// Test long flag
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "update", issue.ID, "--comment", "Test comment"})
	if err != nil {
		t.Fatalf("update command with --comment failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ := store.GetIssue(issue.ID)

	if len(updated.Comments) != 1 || updated.Comments[0] != "Test comment" {
		t.Errorf("expected Comments ['Test comment'], got %v", updated.Comments)
	}

	// Test short flag
	cmd2 := newCommand()
	var buf2 bytes.Buffer
	cmd2.Writer = &buf2

	err = cmd2.Run(context.Background(), []string{"mint", "update", issue.ID, "-c", "Another comment"})
	if err != nil {
		t.Fatalf("update command with -c failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ = store.GetIssue(issue.ID)

	if len(updated.Comments) != 2 || updated.Comments[1] != "Another comment" {
		t.Errorf("expected 2 comments with second being 'Another comment', got %v", updated.Comments)
	}
}

func TestUpdateCommandMultipleFlags(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	issue3, _ := store.AddIssue("Issue 3")
	_ = store.Save(filePath)

	// Test long flags
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "update", issue1.ID, "--title", "Updated", "--depends-on", issue2.ID, "--comment", "Done"})
	if err != nil {
		t.Fatalf("update command with long flags failed: %v", err)
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

	// Test short flags
	cmd2 := newCommand()
	var buf2 bytes.Buffer
	cmd2.Writer = &buf2

	err = cmd2.Run(context.Background(), []string{"mint", "update", issue1.ID, "-t", "Revised", "-b", issue3.ID, "-c", "Also done"})
	if err != nil {
		t.Fatalf("update command with short flags failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ = store.GetIssue(issue1.ID)

	if updated.Title != "Revised" {
		t.Errorf("expected title 'Revised', got '%s'", updated.Title)
	}
	if len(updated.Blocks) != 1 || updated.Blocks[0] != issue3.ID {
		t.Errorf("expected Blocks [%s], got %v", issue3.ID, updated.Blocks)
	}
	if len(updated.Comments) != 2 || updated.Comments[1] != "Also done" {
		t.Errorf("expected 2 comments with second being 'Also done', got %v", updated.Comments)
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
	readyIssue1, _ := store.AddIssue("Ready issue 1")
	blockedIssue, _ := store.AddIssue("Blocked issue")
	blocker, _ := store.AddIssue("Blocker issue")
	closedIssue1, _ := store.AddIssue("Closed issue 1")
	closedIssue2, _ := store.AddIssue("Closed issue 2")

	_ = store.AddDependency(blockedIssue.ID, blocker.ID)
	_ = store.CloseIssue(closedIssue1.ID, "")
	_ = store.CloseIssue(closedIssue2.ID, "")
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

	// Check for READY header
	if !strings.Contains(output, " READY ") {
		t.Error("expected output to contain ' READY ' header")
	}

	// Check for BLOCKED header
	if !strings.Contains(output, " BLOCKED ") {
		t.Error("expected output to contain ' BLOCKED ' header")
	}

	// Check for CLOSED header
	if !strings.Contains(output, " CLOSED ") {
		t.Error("expected output to contain ' CLOSED ' header")
	}

	// Check that ready issues appear (open with no dependencies)
	if !strings.Contains(strippedOutput, "   "+readyIssue1.ID) {
		t.Errorf("expected ready issue to be indented with 3 spaces, got: %s", strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   "+blocker.ID) {
		t.Errorf("expected blocker issue (also ready) to be indented with 3 spaces, got: %s", strippedOutput)
	}

	// Check that blocked issue appears (open with dependencies)
	if !strings.Contains(strippedOutput, "   "+blockedIssue.ID) {
		t.Errorf("expected blocked issue to be indented with 3 spaces, got: %s", strippedOutput)
	}

	// Check that closed issues appear
	if !strings.Contains(strippedOutput, "   "+closedIssue1.ID) {
		t.Errorf("expected closed issue1 to be indented with 3 spaces, got: %s", strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   "+closedIssue2.ID) {
		t.Errorf("expected closed issue2 to be indented with 3 spaces, got: %s", strippedOutput)
	}

	// Verify order: READY, BLOCKED, CLOSED
	readyIdx := strings.Index(strippedOutput, "READY")
	blockedIdx := strings.Index(strippedOutput, "BLOCKED")
	closedIdx := strings.Index(strippedOutput, "CLOSED")
	if readyIdx == -1 || blockedIdx == -1 || closedIdx == -1 {
		t.Error("expected all three section headers")
	}
	if readyIdx >= blockedIdx || blockedIdx >= closedIdx {
		t.Error("expected section order: READY, BLOCKED, CLOSED")
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

	// Check for all three headers
	if !strings.Contains(output, " READY ") {
		t.Error("expected output to contain ' READY ' header even when empty")
	}
	if !strings.Contains(output, " BLOCKED ") {
		t.Error("expected output to contain ' BLOCKED ' header even when empty")
	}
	if !strings.Contains(output, " CLOSED ") {
		t.Error("expected output to contain ' CLOSED ' header")
	}

	// Check for empty messages with 3-space indent
	if !strings.Contains(strippedOutput, "   (No ready issues.)") {
		t.Errorf("expected '   (No ready issues.)' when no ready issues, got: %s", strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   (No blocked issues.)") {
		t.Errorf("expected '   (No blocked issues.)' when no blocked issues, got: %s", strippedOutput)
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

	// Check for all three headers
	if !strings.Contains(output, " READY ") {
		t.Error("expected output to contain ' READY ' header")
	}
	if !strings.Contains(output, " BLOCKED ") {
		t.Error("expected output to contain ' BLOCKED ' header even when empty")
	}
	if !strings.Contains(output, " CLOSED ") {
		t.Error("expected output to contain ' CLOSED ' header even when empty")
	}

	// Check for empty messages with 3-space indent
	if !strings.Contains(strippedOutput, "   (No blocked issues.)") {
		t.Errorf("expected '   (No blocked issues.)' when no blocked issues, got: %s", strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   (No closed issues.)") {
		t.Errorf("expected '   (No closed issues.)' when no closed issues, got: %s", strippedOutput)
	}

	// Check that ready issue is shown
	if !strings.Contains(strippedOutput, "   "+issue.ID) {
		t.Errorf("expected ready issue to be shown, got: %s", strippedOutput)
	}
}

func TestListCommandWithReadyFlag(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	readyIssue, _ := store.AddIssue("Ready issue")
	blockedIssue, _ := store.AddIssue("Blocked issue")
	blocker, _ := store.AddIssue("Blocker issue")
	closedIssue, _ := store.AddIssue("Closed issue")

	_ = store.AddDependency(blockedIssue.ID, blocker.ID)
	_ = store.CloseIssue(closedIssue.ID, "")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--ready"})
	if err != nil {
		t.Fatalf("list --ready command failed: %v", err)
	}

	output := buf.String()
	strippedOutput := stripANSI(output)

	// Should show READY header only
	if !strings.Contains(output, " READY ") {
		t.Error("expected output to contain ' READY ' header")
	}

	// Should NOT show BLOCKED or CLOSED headers
	if strings.Contains(output, " BLOCKED ") {
		t.Error("expected output NOT to contain ' BLOCKED ' header with --ready flag")
	}
	if strings.Contains(output, " CLOSED ") {
		t.Error("expected output NOT to contain ' CLOSED ' header with --ready flag")
	}

	// Should show ready issues
	if !strings.Contains(strippedOutput, "   "+readyIssue.ID) {
		t.Errorf("expected ready issue %s with 3-space indent, got: %s", readyIssue.ID, strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   "+blocker.ID) {
		t.Errorf("expected blocker (ready) issue %s with 3-space indent, got: %s", blocker.ID, strippedOutput)
	}

	// Should NOT show blocked or closed issues
	if strings.Contains(strippedOutput, blockedIssue.ID) {
		t.Errorf("expected output NOT to contain blocked issue %s, got: %s", blockedIssue.ID, strippedOutput)
	}
	if strings.Contains(strippedOutput, closedIssue.ID) {
		t.Errorf("expected output NOT to contain closed issue %s, got: %s", closedIssue.ID, strippedOutput)
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

func TestCreateCommandWithDescription(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mint", "create", "Test issue", "--description", "This is the description"})
	if err != nil {
		t.Fatalf("create command with description failed: %v", err)
	}

	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	if len(store.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(store.Issues))
	}

	var issue *Issue
	for _, iss := range store.Issues {
		issue = iss
		break
	}

	if issue == nil {
		t.Fatal("expected to find an issue")
	}

	if len(issue.Comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(issue.Comments))
	}

	if len(issue.Comments) > 0 && issue.Comments[0] != "This is the description" {
		t.Errorf("expected comment 'This is the description', got '%s'", issue.Comments[0])
	}
}

func TestCreateCommandWithComment(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mint", "create", "Test issue", "--comment", "This is a comment"})
	if err != nil {
		t.Fatalf("create command with comment failed: %v", err)
	}

	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	var issue *Issue
	for _, iss := range store.Issues {
		issue = iss
		break
	}

	if issue == nil {
		t.Fatal("expected to find an issue")
	}

	if len(issue.Comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(issue.Comments))
	}

	if len(issue.Comments) > 0 && issue.Comments[0] != "This is a comment" {
		t.Errorf("expected comment 'This is a comment', got '%s'", issue.Comments[0])
	}
}

func TestCreateCommandWithBothDescriptionAndComment(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mint", "create", "Test issue", "--description", "Description text", "--comment", "Comment text"})
	if err != nil {
		t.Fatalf("create command with both flags failed: %v", err)
	}

	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	var issue *Issue
	for _, iss := range store.Issues {
		issue = iss
		break
	}

	if issue == nil {
		t.Fatal("expected to find an issue")
	}

	if len(issue.Comments) != 2 {
		t.Errorf("expected 2 comments, got %d", len(issue.Comments))
	}

	if len(issue.Comments) > 0 && issue.Comments[0] != "Description text" {
		t.Errorf("expected first comment 'Description text', got '%s'", issue.Comments[0])
	}

	if len(issue.Comments) > 1 && issue.Comments[1] != "Comment text" {
		t.Errorf("expected second comment 'Comment text', got '%s'", issue.Comments[1])
	}
}

func TestCreateCommandWithEmptyDescription(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mint", "create", "Test issue", "--description", ""})
	if err != nil {
		t.Fatalf("create command with empty description failed: %v", err)
	}

	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	var issue *Issue
	for _, iss := range store.Issues {
		issue = iss
		break
	}

	if issue == nil {
		t.Fatal("expected to find an issue")
	}

	if len(issue.Comments) != 0 {
		t.Errorf("expected 0 comments for empty description, got %d", len(issue.Comments))
	}
}

func TestCreateCommandWithEmptyComment(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mint", "create", "Test issue", "--comment", ""})
	if err != nil {
		t.Fatalf("create command with empty comment failed: %v", err)
	}

	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	var issue *Issue
	for _, iss := range store.Issues {
		issue = iss
		break
	}

	if issue == nil {
		t.Fatal("expected to find an issue")
	}

	if len(issue.Comments) != 0 {
		t.Errorf("expected 0 comments for empty comment, got %d", len(issue.Comments))
	}
}

func TestDeleteCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "delete", issue.ID})
	if err != nil {
		t.Fatalf("delete command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	_, err = store.GetIssue(issue.ID)
	if err == nil {
		t.Error("expected issue to be deleted")
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Deleted issue "+issue.ID) {
		t.Errorf("expected output to contain 'Deleted issue %s', got: %s", issue.ID, output)
	}
}

func TestDeleteCommand_NoID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "delete"})
	if err == nil {
		t.Error("expected error when no issue ID provided")
	}
}

func TestDeleteCommand_InvalidID(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "delete", "mint-nonexistent"})
	if err == nil {
		t.Error("expected error when deleting non-existent issue")
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "--version"})
	if err != nil {
		t.Fatalf("--version flag failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, version) {
		t.Errorf("expected output to contain version '%s', got: %s", version, output)
	}
}

func TestVersionFlagShort(t *testing.T) {
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "-v"})
	if err != nil {
		t.Fatalf("-v flag failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, version) {
		t.Errorf("expected output to contain version '%s', got: %s", version, output)
	}
}

func TestShellCompletionEnabled(t *testing.T) {
	cmd := newCommand()
	if !cmd.EnableShellCompletion {
		t.Error("expected EnableShellCompletion to be true")
	}
}
