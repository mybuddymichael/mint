package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

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
	if !strings.Contains(output, "ID      "+issue.ID) {
		t.Errorf("expected output to contain 'ID      %s', got: %s", issue.ID, output)
	}
	if !strings.Contains(output, "Title   Test show issue") {
		t.Errorf("expected output to contain 'Title   Test show issue', got: %s", output)
	}
	if !strings.Contains(output, "Status  open") {
		t.Errorf("expected output to contain 'Status  open', got: %s", output)
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

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Comments") {
		t.Errorf("expected output to contain 'Comments', got: %s", output)
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

	output := stripANSI(buf.String())
	if strings.Contains(output, "Comments") {
		t.Errorf("expected output NOT to contain 'Comments' when no comments, got: %s", output)
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
	if !strings.Contains(output, "Depends on") {
		t.Errorf("expected output to contain 'Depends on', got: %s", output)
	}
	if !strings.Contains(output, "  "+issue2.ID+" open Dependency issue") {
		t.Errorf("expected output to contain '  %s open Dependency issue', got: %s", issue2.ID, output)
	}
	if !strings.Contains(output, "Blocks") {
		t.Errorf("expected output to contain 'Blocks', got: %s", output)
	}
	if !strings.Contains(output, "  "+issue3.ID+" open Blocked issue") {
		t.Errorf("expected output to contain '  %s open Blocked issue', got: %s", issue3.ID, output)
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

	output := stripANSI(buf.String())
	if strings.Contains(output, "Depends on") {
		t.Errorf("expected output NOT to contain 'Depends on' when no dependencies, got: %s", output)
	}
	if strings.Contains(output, "Blocks") {
		t.Errorf("expected output NOT to contain 'Blocks' when no blocks, got: %s", output)
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

	output := stripANSI(buf.String())

	dependsIdx := strings.Index(output, "Depends on")
	blocksIdx := strings.Index(output, "Blocks")
	commentsIdx := strings.Index(output, "Comments")

	if dependsIdx == -1 {
		t.Error("expected output to contain 'Depends on'")
	}
	if blocksIdx == -1 {
		t.Error("expected output to contain 'Blocks'")
	}
	if commentsIdx == -1 {
		t.Error("expected output to contain 'Comments'")
	}

	if commentsIdx != -1 && dependsIdx != -1 && commentsIdx < dependsIdx {
		t.Errorf("Comments section should appear after Depends on section")
	}
	if commentsIdx != -1 && blocksIdx != -1 && commentsIdx < blocksIdx {
		t.Errorf("Comments section should appear after Blocks section")
	}
}

func TestShowCommandAlias(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	t.Setenv("MINT_STORE_FILE", filePath)

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

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"mint", "s", issue.ID})
	if err != nil {
		t.Fatalf("s alias command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "ID      "+issue.ID) {
		t.Errorf("expected output to contain 'ID      %s', got: %s", issue.ID, output)
	}
	if !strings.Contains(output, "Title   Test show issue") {
		t.Errorf("expected output to contain 'Title   Test show issue', got: %s", output)
	}
}
