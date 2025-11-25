package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

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
	if !strings.Contains(output, "✔︎ Created issue") {
		t.Errorf("expected output to contain '✔︎ Created issue', got: %s", output)
	}
	if !strings.Contains(output, "ID      mint-") {
		t.Errorf("expected output to contain 'ID      mint-', got: %s", output)
	}
	if !strings.Contains(output, "Title   Test issue") {
		t.Errorf("expected output to contain 'Title   Test issue', got: %s", output)
	}
	if !strings.Contains(output, "Status  open") {
		t.Errorf("expected output to contain 'Status  open', got: %s", output)
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
	if !strings.Contains(output, "✔︎ Created issue") {
		t.Errorf("expected output to contain '✔︎ Created issue', got: %s", output)
	}
	if !strings.Contains(output, "Title   Test issue") {
		t.Errorf("expected output to contain 'Title   Test issue', got: %s", output)
	}
}

func TestCreateCommandAliasC(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mint", "c", "Test issue"})
	if err != nil {
		t.Fatalf("c alias command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "✔︎ Created issue") {
		t.Errorf("expected output to contain '✔︎ Created issue', got: %s", output)
	}
	if !strings.Contains(output, "Title   Test issue") {
		t.Errorf("expected output to contain 'Title   Test issue', got: %s", output)
	}
}

func TestCreateCommandAliasA(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	t.Setenv("MINT_STORE_FILE", filePath)

	err := cmd.Run(context.Background(), []string{"mint", "a", "Test issue"})
	if err != nil {
		t.Fatalf("a alias command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "✔︎ Created issue") {
		t.Errorf("expected output to contain '✔︎ Created issue', got: %s", output)
	}
	if !strings.Contains(output, "Title   Test issue") {
		t.Errorf("expected output to contain 'Title   Test issue', got: %s", output)
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

func TestCreateCommandWithDependsOn(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	// Setup: create dependency issue first
	store, _ := LoadStore(filePath)
	depIssue, _ := store.AddIssue("Dependency issue")
	_ = store.Save(filePath)

	// Execute: create with --depends-on
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf
	err := cmd.Run(context.Background(), []string{"mint", "create", "New issue", "--depends-on", depIssue.ID})
	if err != nil {
		t.Fatalf("create with --depends-on failed: %v", err)
	}

	// Verify: reload and check relationships
	store, _ = LoadStore(filePath)
	// Find newly created issue (will be the one that's not depIssue)
	var newIssue *Issue
	for _, iss := range store.Issues {
		if iss.ID != depIssue.ID {
			newIssue = iss
			break
		}
	}

	if newIssue == nil {
		t.Fatal("expected to find new issue")
	}

	if len(newIssue.DependsOn) != 1 || newIssue.DependsOn[0] != depIssue.ID {
		t.Errorf("expected DependsOn [%s], got %v", depIssue.ID, newIssue.DependsOn)
	}

	// Verify bidirectional relationship
	depIssueUpdated, _ := store.GetIssue(depIssue.ID)
	if len(depIssueUpdated.Blocks) != 1 || depIssueUpdated.Blocks[0] != newIssue.ID {
		t.Errorf("expected depIssue Blocks [%s], got %v", newIssue.ID, depIssueUpdated.Blocks)
	}
}

func TestCreateCommandWithDependsOnShortFlag(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	// Setup
	store, _ := LoadStore(filePath)
	depIssue, _ := store.AddIssue("Dependency issue")
	_ = store.Save(filePath)

	// Execute with -d
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf
	err := cmd.Run(context.Background(), []string{"mint", "create", "New issue", "-d", depIssue.ID})
	if err != nil {
		t.Fatalf("create with -d failed: %v", err)
	}

	// Verify
	store, _ = LoadStore(filePath)
	var newIssue *Issue
	for _, iss := range store.Issues {
		if iss.ID != depIssue.ID {
			newIssue = iss
			break
		}
	}

	if newIssue == nil {
		t.Fatal("expected to find new issue")
	}

	if len(newIssue.DependsOn) != 1 || newIssue.DependsOn[0] != depIssue.ID {
		t.Errorf("expected DependsOn [%s], got %v", depIssue.ID, newIssue.DependsOn)
	}
}

func TestCreateCommandWithMultipleDependsOn(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	// Setup: create two dependency issues
	store, _ := LoadStore(filePath)
	depIssue1, _ := store.AddIssue("Dependency 1")
	depIssue2, _ := store.AddIssue("Dependency 2")
	_ = store.Save(filePath)

	// Execute with multiple --depends-on
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf
	err := cmd.Run(context.Background(), []string{"mint", "create", "New issue", "--depends-on", depIssue1.ID, "--depends-on", depIssue2.ID})
	if err != nil {
		t.Fatalf("create with multiple --depends-on failed: %v", err)
	}

	// Verify
	store, _ = LoadStore(filePath)
	var newIssue *Issue
	for _, iss := range store.Issues {
		if iss.ID != depIssue1.ID && iss.ID != depIssue2.ID {
			newIssue = iss
			break
		}
	}

	if newIssue == nil {
		t.Fatal("expected to find new issue")
	}

	if len(newIssue.DependsOn) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(newIssue.DependsOn))
	}

	// Verify both dependencies are present
	hasDepIssue1 := false
	hasDepIssue2 := false
	for _, depID := range newIssue.DependsOn {
		if depID == depIssue1.ID {
			hasDepIssue1 = true
		}
		if depID == depIssue2.ID {
			hasDepIssue2 = true
		}
	}
	if !hasDepIssue1 || !hasDepIssue2 {
		t.Errorf("expected both dependencies, got %v", newIssue.DependsOn)
	}

	// Verify bidirectional relationships
	dep1Updated, _ := store.GetIssue(depIssue1.ID)
	if len(dep1Updated.Blocks) != 1 || dep1Updated.Blocks[0] != newIssue.ID {
		t.Errorf("expected dep1 Blocks [%s], got %v", newIssue.ID, dep1Updated.Blocks)
	}

	dep2Updated, _ := store.GetIssue(depIssue2.ID)
	if len(dep2Updated.Blocks) != 1 || dep2Updated.Blocks[0] != newIssue.ID {
		t.Errorf("expected dep2 Blocks [%s], got %v", newIssue.ID, dep2Updated.Blocks)
	}
}

func TestCreateCommandWithBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	// Setup: create issues to be blocked
	store, _ := LoadStore(filePath)
	blockedIssue1, _ := store.AddIssue("Blocked issue 1")
	blockedIssue2, _ := store.AddIssue("Blocked issue 2")
	_ = store.Save(filePath)

	// Execute with --blocks
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf
	err := cmd.Run(context.Background(), []string{"mint", "create", "Blocker issue", "--blocks", blockedIssue1.ID, "--blocks", blockedIssue2.ID})
	if err != nil {
		t.Fatalf("create with --blocks failed: %v", err)
	}

	// Verify
	store, _ = LoadStore(filePath)
	var newIssue *Issue
	for _, iss := range store.Issues {
		if iss.ID != blockedIssue1.ID && iss.ID != blockedIssue2.ID {
			newIssue = iss
			break
		}
	}

	if newIssue == nil {
		t.Fatal("expected to find new issue")
	}

	if len(newIssue.Blocks) != 2 {
		t.Errorf("expected 2 blocks, got %d", len(newIssue.Blocks))
	}

	// Verify both blocks are present
	hasBlocked1 := false
	hasBlocked2 := false
	for _, blockID := range newIssue.Blocks {
		if blockID == blockedIssue1.ID {
			hasBlocked1 = true
		}
		if blockID == blockedIssue2.ID {
			hasBlocked2 = true
		}
	}
	if !hasBlocked1 || !hasBlocked2 {
		t.Errorf("expected both blocks, got %v", newIssue.Blocks)
	}

	// Verify bidirectional relationships
	blocked1Updated, _ := store.GetIssue(blockedIssue1.ID)
	if len(blocked1Updated.DependsOn) != 1 || blocked1Updated.DependsOn[0] != newIssue.ID {
		t.Errorf("expected blocked1 DependsOn [%s], got %v", newIssue.ID, blocked1Updated.DependsOn)
	}

	blocked2Updated, _ := store.GetIssue(blockedIssue2.ID)
	if len(blocked2Updated.DependsOn) != 1 || blocked2Updated.DependsOn[0] != newIssue.ID {
		t.Errorf("expected blocked2 DependsOn [%s], got %v", newIssue.ID, blocked2Updated.DependsOn)
	}
}

func TestCreateCommandWithBlocksShortFlag(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	// Setup
	store, _ := LoadStore(filePath)
	blockedIssue, _ := store.AddIssue("Blocked issue")
	_ = store.Save(filePath)

	// Execute with -b
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf
	err := cmd.Run(context.Background(), []string{"mint", "create", "Blocker issue", "-b", blockedIssue.ID})
	if err != nil {
		t.Fatalf("create with -b failed: %v", err)
	}

	// Verify
	store, _ = LoadStore(filePath)
	var newIssue *Issue
	for _, iss := range store.Issues {
		if iss.ID != blockedIssue.ID {
			newIssue = iss
			break
		}
	}

	if newIssue == nil {
		t.Fatal("expected to find new issue")
	}

	if len(newIssue.Blocks) != 1 || newIssue.Blocks[0] != blockedIssue.ID {
		t.Errorf("expected Blocks [%s], got %v", blockedIssue.ID, newIssue.Blocks)
	}
}

func TestCreateCommandWithMixedRelationships(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	// Setup
	store, _ := LoadStore(filePath)
	depIssue, _ := store.AddIssue("Dependency issue")
	blockedIssue, _ := store.AddIssue("Blocked issue")
	_ = store.Save(filePath)

	// Execute with both --depends-on and --blocks and --description
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf
	err := cmd.Run(context.Background(), []string{"mint", "create", "Mixed issue", "--depends-on", depIssue.ID, "--blocks", blockedIssue.ID, "--description", "Test description"})
	if err != nil {
		t.Fatalf("create with mixed relationships failed: %v", err)
	}

	// Verify
	store, _ = LoadStore(filePath)
	var newIssue *Issue
	for _, iss := range store.Issues {
		if iss.ID != depIssue.ID && iss.ID != blockedIssue.ID {
			newIssue = iss
			break
		}
	}

	if newIssue == nil {
		t.Fatal("expected to find new issue")
	}

	// Check depends-on
	if len(newIssue.DependsOn) != 1 || newIssue.DependsOn[0] != depIssue.ID {
		t.Errorf("expected DependsOn [%s], got %v", depIssue.ID, newIssue.DependsOn)
	}

	// Check blocks
	if len(newIssue.Blocks) != 1 || newIssue.Blocks[0] != blockedIssue.ID {
		t.Errorf("expected Blocks [%s], got %v", blockedIssue.ID, newIssue.Blocks)
	}

	// Check description was added
	if len(newIssue.Comments) != 1 || newIssue.Comments[0] != "Test description" {
		t.Errorf("expected comment 'Test description', got %v", newIssue.Comments)
	}

	// Verify bidirectional relationships
	depUpdated, _ := store.GetIssue(depIssue.ID)
	if len(depUpdated.Blocks) != 1 || depUpdated.Blocks[0] != newIssue.ID {
		t.Errorf("expected dep Blocks [%s], got %v", newIssue.ID, depUpdated.Blocks)
	}

	blockedUpdated, _ := store.GetIssue(blockedIssue.ID)
	if len(blockedUpdated.DependsOn) != 1 || blockedUpdated.DependsOn[0] != newIssue.ID {
		t.Errorf("expected blocked DependsOn [%s], got %v", newIssue.ID, blockedUpdated.DependsOn)
	}
}

func TestCreateCommandWithInvalidDependency(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	// Setup: create empty store
	store, _ := LoadStore(filePath)
	_ = store.Save(filePath)

	// Execute with non-existent dependency
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf
	err := cmd.Run(context.Background(), []string{"mint", "create", "New issue", "--depends-on", "mint-nonexistent"})
	if err == nil {
		t.Error("expected error for invalid dependency")
	}

	// Verify no issue was created (atomic behavior)
	store, _ = LoadStore(filePath)
	if len(store.Issues) != 0 {
		t.Errorf("expected 0 issues after failed create, got %d", len(store.Issues))
	}
}

func TestCreateCommandWithInvalidBlockedIssue(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	// Setup: create empty store
	store, _ := LoadStore(filePath)
	_ = store.Save(filePath)

	// Execute with non-existent blocked issue
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf
	err := cmd.Run(context.Background(), []string{"mint", "create", "New issue", "--blocks", "mint-nonexistent"})
	if err == nil {
		t.Error("expected error for invalid blocked issue")
	}

	// Verify no issue was created (atomic behavior)
	store, _ = LoadStore(filePath)
	if len(store.Issues) != 0 {
		t.Errorf("expected 0 issues after failed create, got %d", len(store.Issues))
	}
}
