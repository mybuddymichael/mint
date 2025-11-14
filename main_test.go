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
