package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestCloseCommandAlias(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "cl", issue.ID})
	if err != nil {
		t.Fatalf("cl alias command failed: %v", err)
	}

	store, _ = LoadStore(filePath)
	closed, _ := store.GetIssue(issue.ID)

	if closed.Status != "closed" {
		t.Errorf("expected status 'closed', got '%s'", closed.Status)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Closed issue "+issue.ID) {
		t.Errorf("expected output to contain 'Closed issue %s', got: %s", issue.ID, output)
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

func TestOpenCommandAlias(t *testing.T) {
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

	err := cmd.Run(context.Background(), []string{"mint", "o", issue.ID})
	if err != nil {
		t.Fatalf("o alias command failed: %v", err)
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

func TestDeleteCommandAlias(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "d", issue.ID})
	if err != nil {
		t.Fatalf("d alias command failed: %v", err)
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
