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
