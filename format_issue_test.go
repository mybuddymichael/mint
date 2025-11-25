package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintIssueDetails_BasicIssue(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.Save(filePath)

	var buf bytes.Buffer
	err := PrintIssueDetails(&buf, issue, store)
	if err != nil {
		t.Fatalf("PrintIssueDetails failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "ID:      "+issue.ID) {
		t.Errorf("expected output to contain 'ID:      %s', got: %s", issue.ID, output)
	}
	if !strings.Contains(output, "Title:   Test issue") {
		t.Errorf("expected output to contain 'Title:   Test issue', got: %s", output)
	}
	if !strings.Contains(output, "Status:  open") {
		t.Errorf("expected output to contain 'Status:  open', got: %s", output)
	}
}

func TestPrintIssueDetails_WithDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Main issue")
	issue2, _ := store.AddIssue("Dependency")
	_ = store.AddDependency(issue1.ID, issue2.ID)
	_ = store.Save(filePath)

	var buf bytes.Buffer
	err := PrintIssueDetails(&buf, issue1, store)
	if err != nil {
		t.Fatalf("PrintIssueDetails failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Depends on:") {
		t.Errorf("expected output to contain 'Depends on:', got: %s", output)
	}
	if !strings.Contains(output, "  "+issue2.ID+" Dependency") {
		t.Errorf("expected output to contain '  %s Dependency', got: %s", issue2.ID, output)
	}
}

func TestPrintIssueDetails_WithBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("Main issue")
	issue2, _ := store.AddIssue("Blocked issue")
	_ = store.AddBlocker(issue1.ID, issue2.ID)
	_ = store.Save(filePath)

	var buf bytes.Buffer
	err := PrintIssueDetails(&buf, issue1, store)
	if err != nil {
		t.Fatalf("PrintIssueDetails failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Blocks:") {
		t.Errorf("expected output to contain 'Blocks:', got: %s", output)
	}
	if !strings.Contains(output, "  "+issue2.ID+" Blocked issue") {
		t.Errorf("expected output to contain '  %s Blocked issue', got: %s", issue2.ID, output)
	}
}

func TestPrintIssueDetails_WithComments(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	_ = store.AddComment(issue.ID, "First comment")
	_ = store.AddComment(issue.ID, "Second comment")
	_ = store.Save(filePath)

	var buf bytes.Buffer
	err := PrintIssueDetails(&buf, issue, store)
	if err != nil {
		t.Fatalf("PrintIssueDetails failed: %v", err)
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

func TestPrintIssueDetails_NoOptionalSections(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Standalone issue")
	_ = store.Save(filePath)

	var buf bytes.Buffer
	err := PrintIssueDetails(&buf, issue, store)
	if err != nil {
		t.Fatalf("PrintIssueDetails failed: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Depends on:") {
		t.Errorf("expected output NOT to contain 'Depends on:', got: %s", output)
	}
	if strings.Contains(output, "Blocks:") {
		t.Errorf("expected output NOT to contain 'Blocks:', got: %s", output)
	}
	if strings.Contains(output, "Comments:") {
		t.Errorf("expected output NOT to contain 'Comments:', got: %s", output)
	}
}

func TestPrintIssueList_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue1, _ := store.AddIssue("First issue")
	issue2, _ := store.AddIssue("Second issue")
	_ = store.CloseIssue(issue2.ID, "")
	_ = store.Save(filePath)

	issues := []*Issue{issue1, issue2}
	maxIDLen := len(issue1.ID)
	if len(issue2.ID) > maxIDLen {
		maxIDLen = len(issue2.ID)
	}

	var buf bytes.Buffer
	err := printIssueList(&buf, issues, maxIDLen, store)
	if err != nil {
		t.Fatalf("printIssueList failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, issue1.ID) {
		t.Errorf("expected output to contain '%s', got: %s", issue1.ID, output)
	}
	if !strings.Contains(output, "First issue") {
		t.Errorf("expected output to contain 'First issue', got: %s", output)
	}
	if !strings.Contains(output, issue2.ID) {
		t.Errorf("expected output to contain '%s', got: %s", issue2.ID, output)
	}
	if !strings.Contains(output, "Second issue") {
		t.Errorf("expected output to contain 'Second issue', got: %s", output)
	}
}
