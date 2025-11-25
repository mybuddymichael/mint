package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

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

	output := stripANSI(buf.String())
	if !strings.Contains(output, "✔︎ Issue updated") {
		t.Errorf("expected output to contain '✔︎ Issue updated', got: %s", output)
	}
	if !strings.Contains(output, "ID      "+issue.ID) {
		t.Errorf("expected output to contain 'ID      %s', got: %s", issue.ID, output)
	}
	if !strings.Contains(output, "Title   Test issue") {
		t.Errorf("expected output to contain 'Title   Test issue', got: %s", output)
	}
	if !strings.Contains(output, "Comments") {
		t.Errorf("expected output to contain 'Comments', got: %s", output)
	}
	if !strings.Contains(output, "Test comment") {
		t.Errorf("expected output to contain 'Test comment', got: %s", output)
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

	output := stripANSI(buf.String())
	if !strings.Contains(output, "✔︎ Issue updated") {
		t.Errorf("expected output to contain '✔︎ Issue updated', got: %s", output)
	}
	if !strings.Contains(output, "ID      "+issue1.ID) {
		t.Errorf("expected output to contain 'ID      %s', got: %s", issue1.ID, output)
	}
	if !strings.Contains(output, "Title   Updated") {
		t.Errorf("expected output to contain 'Title   Updated', got: %s", output)
	}
	if !strings.Contains(output, "Comments") {
		t.Errorf("expected output to contain 'Comments', got: %s", output)
	}
	if !strings.Contains(output, "Done") {
		t.Errorf("expected output to contain comment 'Done', got: %s", output)
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
	if !strings.Contains(output, "✔︎ Issue updated") {
		t.Errorf("expected output to contain '✔︎ Issue updated', got: %s", output)
	}
	if !strings.Contains(output, "ID      "+issue.ID) {
		t.Errorf("expected output to contain 'ID      %s' (full ID), got: %s", issue.ID, output)
	}
}

func TestUpdateCommandAlias(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Original title")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "u", issue.ID, "--title", "New title"})
	if err != nil {
		t.Fatalf("u alias command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ := store.GetIssue(issue.ID)

	if updated.Title != "New title" {
		t.Errorf("expected title 'New title', got '%s'", updated.Title)
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

func TestUpdateCommandRemoveDependsOn(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	issue3, _ := store.AddIssue("Issue 3")

	// Add dependencies
	_ = store.AddDependency(issue1.ID, issue2.ID)
	_ = store.AddDependency(issue1.ID, issue3.ID)
	_ = store.Save(filePath)

	// Test long flag
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "update", issue1.ID, "--remove-depends-on", issue2.ID})
	if err != nil {
		t.Fatalf("update command with --remove-depends-on failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ := store.GetIssue(issue1.ID)

	// Should have one dependency left (issue3)
	if len(updated.DependsOn) != 1 || updated.DependsOn[0] != issue3.ID {
		t.Errorf("expected DependsOn [%s], got %v", issue3.ID, updated.DependsOn)
	}

	// issue2 should no longer block issue1
	blocker, _ := store.GetIssue(issue2.ID)
	if len(blocker.Blocks) != 0 {
		t.Errorf("expected issue2 to have no blocks, got %v", blocker.Blocks)
	}

	// Test short flag
	cmd2 := newCommand()
	var buf2 bytes.Buffer
	cmd2.Writer = &buf2

	err = cmd2.Run(context.Background(), []string{"mint", "update", issue1.ID, "-rd", issue3.ID})
	if err != nil {
		t.Fatalf("update command with -rd failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ = store.GetIssue(issue1.ID)

	// Should have no dependencies left
	if len(updated.DependsOn) != 0 {
		t.Errorf("expected no dependencies, got %v", updated.DependsOn)
	}

	blocker3, _ := store.GetIssue(issue3.ID)
	if len(blocker3.Blocks) != 0 {
		t.Errorf("expected issue3 to have no blocks, got %v", blocker3.Blocks)
	}
}

func TestUpdateCommandRemoveBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Issue 1")
	issue2, _ := store.AddIssue("Issue 2")
	issue3, _ := store.AddIssue("Issue 3")

	// Add blockers
	_ = store.AddBlocker(issue1.ID, issue2.ID)
	_ = store.AddBlocker(issue1.ID, issue3.ID)
	_ = store.Save(filePath)

	// Test long flag
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "update", issue1.ID, "--remove-blocks", issue2.ID})
	if err != nil {
		t.Fatalf("update command with --remove-blocks failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ := store.GetIssue(issue1.ID)

	// Should have one block left (issue3)
	if len(updated.Blocks) != 1 || updated.Blocks[0] != issue3.ID {
		t.Errorf("expected Blocks [%s], got %v", issue3.ID, updated.Blocks)
	}

	// issue2 should no longer depend on issue1
	blocked, _ := store.GetIssue(issue2.ID)
	if len(blocked.DependsOn) != 0 {
		t.Errorf("expected issue2 to have no dependencies, got %v", blocked.DependsOn)
	}

	// Test short flag
	cmd2 := newCommand()
	var buf2 bytes.Buffer
	cmd2.Writer = &buf2

	err = cmd2.Run(context.Background(), []string{"mint", "update", issue1.ID, "-rb", issue3.ID})
	if err != nil {
		t.Fatalf("update command with -rb failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	updated, _ = store.GetIssue(issue1.ID)

	// Should have no blocks left
	if len(updated.Blocks) != 0 {
		t.Errorf("expected no blocks, got %v", updated.Blocks)
	}

	blocked3, _ := store.GetIssue(issue3.ID)
	if len(blocked3.DependsOn) != 0 {
		t.Errorf("expected issue3 to have no dependencies, got %v", blocked3.DependsOn)
	}
}
