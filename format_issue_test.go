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
	if !strings.Contains(output, "ID      "+issue.ID) {
		t.Errorf("expected output to contain 'ID      %s', got: %s", issue.ID, output)
	}
	if !strings.Contains(output, "Title   Test issue") {
		t.Errorf("expected output to contain 'Title   Test issue', got: %s", output)
	}
	if !strings.Contains(output, "Status  open") {
		t.Errorf("expected output to contain 'Status  open', got: %s", output)
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
	if !strings.Contains(output, "Depends on") {
		t.Errorf("expected output to contain 'Depends on', got: %s", output)
	}
	if !strings.Contains(output, "  "+issue2.ID+" open Dependency") {
		t.Errorf("expected output to contain '  %s open Dependency', got: %s", issue2.ID, output)
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
	if !strings.Contains(output, "Blocks") {
		t.Errorf("expected output to contain 'Blocks', got: %s", output)
	}
	if !strings.Contains(output, "  "+issue2.ID+" open Blocked issue") {
		t.Errorf("expected output to contain '  %s open Blocked issue', got: %s", issue2.ID, output)
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

	output := stripANSI(buf.String())
	if strings.Contains(output, "Depends on") {
		t.Errorf("expected output NOT to contain 'Depends on', got: %s", output)
	}
	if strings.Contains(output, "Blocks") {
		t.Errorf("expected output NOT to contain 'Blocks', got: %s", output)
	}
	if strings.Contains(output, "Comments") {
		t.Errorf("expected output NOT to contain 'Comments', got: %s", output)
	}
}

func TestPrintIssueDetails_KeyFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	dep, _ := store.AddIssue("Dependency")
	block, _ := store.AddIssue("Blocked")
	_ = store.AddDependency(issue.ID, dep.ID)
	_ = store.AddBlocker(issue.ID, block.ID)
	_ = store.AddComment(issue.ID, "Test comment")
	_ = store.Save(filePath)

	var buf bytes.Buffer
	err := PrintIssueDetails(&buf, issue, store)
	if err != nil {
		t.Fatalf("PrintIssueDetails failed: %v", err)
	}

	output := buf.String()

	// Check that keys are bold (1m) and color 5 (38;5;5m)
	if !strings.Contains(output, "\033[1m\033[38;5;5mID\033[0m") {
		t.Errorf("expected 'ID' key to be bold and color 5, got: %s", output)
	}
	if !strings.Contains(output, "\033[1m\033[38;5;5mTitle\033[0m") {
		t.Errorf("expected 'Title' key to be bold and color 5, got: %s", output)
	}
	if !strings.Contains(output, "\033[1m\033[38;5;5mStatus\033[0m") {
		t.Errorf("expected 'Status' key to be bold and color 5, got: %s", output)
	}
	if !strings.Contains(output, "\033[1m\033[38;5;5mDepends on\033[0m") {
		t.Errorf("expected 'Depends on' key to be bold and color 5, got: %s", output)
	}
	if !strings.Contains(output, "\033[1m\033[38;5;5mBlocks\033[0m") {
		t.Errorf("expected 'Blocks' key to be bold and color 5, got: %s", output)
	}
	if !strings.Contains(output, "\033[1m\033[38;5;5mComments\033[0m") {
		t.Errorf("expected 'Comments' key to be bold and color 5, got: %s", output)
	}

	// Check that there are no colons after the keys
	stripped := stripANSI(output)
	if strings.Contains(stripped, "ID:") {
		t.Errorf("expected no colon after 'ID', got: %s", stripped)
	}
	if strings.Contains(stripped, "Title:") {
		t.Errorf("expected no colon after 'Title', got: %s", stripped)
	}
	if strings.Contains(stripped, "Status:") {
		t.Errorf("expected no colon after 'Status', got: %s", stripped)
	}
	if strings.Contains(stripped, "Depends on:") {
		t.Errorf("expected no colon after 'Depends on', got: %s", stripped)
	}
	if strings.Contains(stripped, "Blocks:") {
		t.Errorf("expected no colon after 'Blocks', got: %s", stripped)
	}
	if strings.Contains(stripped, "Comments:") {
		t.Errorf("expected no colon after 'Comments', got: %s", stripped)
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

func TestPrintIssueDetails_WithStaleDepends(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue with stale dependency")
	// Manually inject stale reference
	issue.DependsOn = []string{"nonexistent-id"}
	_ = store.Save(filePath)

	var buf bytes.Buffer
	err := PrintIssueDetails(&buf, issue, store)
	if err != nil {
		t.Fatalf("PrintIssueDetails should not fail on stale refs: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Depends on") {
		t.Errorf("expected output to contain 'Depends on', got: %s", output)
	}
	if !strings.Contains(output, "nonexistent-id (not found)") {
		t.Errorf("expected output to contain 'nonexistent-id (not found)', got: %s", output)
	}
}

func TestPrintIssueDetails_WithStaleBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue with stale blocker")
	// Manually inject stale reference
	issue.Blocks = []string{"nonexistent-block"}
	_ = store.Save(filePath)

	var buf bytes.Buffer
	err := PrintIssueDetails(&buf, issue, store)
	if err != nil {
		t.Fatalf("PrintIssueDetails should not fail on stale refs: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, "Blocks") {
		t.Errorf("expected output to contain 'Blocks', got: %s", output)
	}
	if !strings.Contains(output, "nonexistent-block (not found)") {
		t.Errorf("expected output to contain 'nonexistent-block (not found)', got: %s", output)
	}
}

func TestPrintIssueDetails_WithMixedValidAndStaleRefs(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	main, _ := store.AddIssue("Main issue")
	validDep, _ := store.AddIssue("Valid dependency")
	validBlock, _ := store.AddIssue("Valid blocker")
	_ = store.AddDependency(main.ID, validDep.ID)
	_ = store.AddBlocker(main.ID, validBlock.ID)
	// Manually inject stale references
	main.DependsOn = append(main.DependsOn, "stale-dep")
	main.Blocks = append(main.Blocks, "stale-block")
	_ = store.Save(filePath)

	var buf bytes.Buffer
	err := PrintIssueDetails(&buf, main, store)
	if err != nil {
		t.Fatalf("PrintIssueDetails should not fail on stale refs: %v", err)
	}

	output := stripANSI(buf.String())
	// Check valid refs are shown normally
	if !strings.Contains(output, validDep.ID+" open Valid dependency") {
		t.Errorf("expected output to contain valid dependency, got: %s", output)
	}
	if !strings.Contains(output, validBlock.ID+" open Valid blocker") {
		t.Errorf("expected output to contain valid blocker, got: %s", output)
	}
	// Check stale refs are shown with (not found)
	if !strings.Contains(output, "stale-dep (not found)") {
		t.Errorf("expected output to contain 'stale-dep (not found)', got: %s", output)
	}
	if !strings.Contains(output, "stale-block (not found)") {
		t.Errorf("expected output to contain 'stale-block (not found)', got: %s", output)
	}
}

func TestPrintIssueDetails_Whitespace(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	issue, _ := store.AddIssue("Test issue")
	dep, _ := store.AddIssue("Dependency")
	_ = store.AddDependency(issue.ID, dep.ID)
	_ = store.AddComment(issue.ID, "Test comment")
	_ = store.Save(filePath)

	var buf bytes.Buffer
	err := PrintIssueDetails(&buf, issue, store)
	if err != nil {
		t.Fatalf("PrintIssueDetails failed: %v", err)
	}

	output := buf.String()
	stripped := stripANSI(output)

	// Should start with newline
	if !strings.HasPrefix(output, "\n") {
		t.Errorf("expected output to start with newline")
	}

	// Should have blank line between Status and Depends on
	if !strings.Contains(stripped, "Status  open\n\nDepends on") {
		t.Errorf("expected blank line between Status and Depends on, got: %s", stripped)
	}

	// Should have blank line between Depends on and Comments
	if !strings.Contains(stripped, "Dependency\n\nComments") {
		t.Errorf("expected blank line between Depends on and Comments, got: %s", stripped)
	}

	// Should end with newline
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("expected output to end with newline")
	}
}
