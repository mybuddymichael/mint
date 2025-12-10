package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

func TestListCommandReadyWithClosedDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)

	// Create blockers
	closedBlocker1, _ := store.AddIssue("Closed blocker 1")
	closedBlocker2, _ := store.AddIssue("Closed blocker 2")
	openBlocker, _ := store.AddIssue("Open blocker")

	// Create blocked issues
	issueWithAllClosedDeps, _ := store.AddIssue("Issue with all closed deps")
	issueWithMixedDeps, _ := store.AddIssue("Issue with mixed deps")
	issueWithAllOpenDeps, _ := store.AddIssue("Issue with all open deps")

	// Set up dependencies
	_ = store.AddDependency(issueWithAllClosedDeps.ID, closedBlocker1.ID)
	_ = store.AddDependency(issueWithAllClosedDeps.ID, closedBlocker2.ID)

	_ = store.AddDependency(issueWithMixedDeps.ID, closedBlocker1.ID)
	_ = store.AddDependency(issueWithMixedDeps.ID, openBlocker.ID)

	_ = store.AddDependency(issueWithAllOpenDeps.ID, openBlocker.ID)

	// Close some blockers
	_ = store.CloseIssue(closedBlocker1.ID, "")
	_ = store.CloseIssue(closedBlocker2.ID, "")

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

	// Issue with all closed dependencies should show as READY
	if !strings.Contains(output, " READY ") {
		t.Error("expected output to contain ' READY ' header")
	}

	// Split output into sections to check where each issue appears
	readyIdx := strings.Index(strippedOutput, " READY ")
	blockedIdx := strings.Index(strippedOutput, " BLOCKED ")

	issueIdx := strings.Index(strippedOutput, issueWithAllClosedDeps.ID)
	if issueIdx == -1 {
		t.Fatalf("issue %s not found in output", issueWithAllClosedDeps.ID)
	}

	// Check that issue appears after READY but before BLOCKED
	if issueIdx < readyIdx || issueIdx > blockedIdx {
		t.Errorf("expected issue %s with all closed deps to show in READY section (between %d and %d), but found at %d",
			issueWithAllClosedDeps.ID, readyIdx, blockedIdx, issueIdx)
	}

	// Issues with any open dependencies should show as BLOCKED
	if !strings.Contains(output, " BLOCKED ") {
		t.Error("expected output to contain ' BLOCKED ' header")
	}
	if !strings.Contains(strippedOutput, "   "+issueWithMixedDeps.ID) {
		t.Errorf("expected issue %s with mixed deps to show in BLOCKED section, got: %s", issueWithMixedDeps.ID, strippedOutput)
	}
	if !strings.Contains(strippedOutput, "   "+issueWithAllOpenDeps.ID) {
		t.Errorf("expected issue %s with all open deps to show in BLOCKED section, got: %s", issueWithAllOpenDeps.ID, strippedOutput)
	}

	// Closed blockers should show in CLOSED section
	if !strings.Contains(output, " CLOSED ") {
		t.Error("expected output to contain ' CLOSED ' header")
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

func TestListCommandAlias(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")

	t.Setenv("MINT_STORE_FILE", filePath)

	store, err := LoadStore(filePath)
	if err != nil {
		t.Fatalf("failed to load store: %v", err)
	}

	issue, err := store.AddIssue("Test issue")
	if err != nil {
		t.Fatalf("failed to add issue: %v", err)
	}

	if err := store.Save(filePath); err != nil {
		t.Fatalf("failed to save store: %v", err)
	}

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"mint", "l"})
	if err != nil {
		t.Fatalf("l alias command failed: %v", err)
	}

	output := stripANSI(buf.String())
	if !strings.Contains(output, issue.ID) {
		t.Errorf("expected output to contain issue ID '%s', got: %s", issue.ID, output)
	}
}

func TestListCommandNewlines(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	_, _ = store.AddIssue("Test issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := buf.String()

	// Should start with a blank line before READY header
	if !strings.HasPrefix(output, "\n") {
		t.Error("expected output to start with blank line before READY header")
	}

	// Should end with a blank line after all content
	if !strings.HasSuffix(output, "\n\n") {
		t.Errorf("expected output to end with double newline (content newline + final blank), got: %q", output[len(output)-10:])
	}
}

func TestListSortsReadyByCreatedAt(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	now := time.Now()

	// Create 5 ready issues with different creation times, added in shuffled order
	store.Issues["issue3"] = &Issue{ID: "issue3", Title: "Third oldest", Status: "open", CreatedAt: now.Add(-3 * time.Hour)}
	store.Issues["issue1"] = &Issue{ID: "issue1", Title: "Oldest", Status: "open", CreatedAt: now.Add(-5 * time.Hour)}
	store.Issues["issue5"] = &Issue{ID: "issue5", Title: "Newest", Status: "open", CreatedAt: now}
	store.Issues["issue2"] = &Issue{ID: "issue2", Title: "Second oldest", Status: "open", CreatedAt: now.Add(-4 * time.Hour)}
	store.Issues["issue4"] = &Issue{ID: "issue4", Title: "Second newest", Status: "open", CreatedAt: now.Add(-1 * time.Hour)}

	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := stripANSI(buf.String())

	// Find positions of each issue ID in output
	idx1 := strings.Index(output, "issue1")
	idx2 := strings.Index(output, "issue2")
	idx3 := strings.Index(output, "issue3")
	idx4 := strings.Index(output, "issue4")
	idx5 := strings.Index(output, "issue5")

	// Verify all issues are present
	if idx1 == -1 || idx2 == -1 || idx3 == -1 || idx4 == -1 || idx5 == -1 {
		t.Fatalf("expected all issues in output, got: %s", output)
	}

	// Verify order: newest to oldest (issue5, issue4, issue3, issue2, issue1)
	if idx5 >= idx4 || idx4 >= idx3 || idx3 >= idx2 || idx2 >= idx1 {
		t.Errorf("expected ready issues sorted by CreatedAt (newest first): issue5(%d), issue4(%d), issue3(%d), issue2(%d), issue1(%d)\noutput:\n%s",
			idx5, idx4, idx3, idx2, idx1, output)
	}
}

func TestListSortsBlockedByCreatedAt(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	now := time.Now()

	// Create a blocker issue
	store.Issues["blocker"] = &Issue{ID: "blocker", Title: "Blocker", Status: "open", CreatedAt: now}

	// Create 5 blocked issues with different creation times, added in shuffled order
	store.Issues["blocked3"] = &Issue{ID: "blocked3", Title: "Third oldest blocked", Status: "open", CreatedAt: now.Add(-3 * time.Hour), DependsOn: []string{"blocker"}}
	store.Issues["blocked1"] = &Issue{ID: "blocked1", Title: "Oldest blocked", Status: "open", CreatedAt: now.Add(-5 * time.Hour), DependsOn: []string{"blocker"}}
	store.Issues["blocked5"] = &Issue{ID: "blocked5", Title: "Newest blocked", Status: "open", CreatedAt: now.Add(-30 * time.Minute), DependsOn: []string{"blocker"}}
	store.Issues["blocked2"] = &Issue{ID: "blocked2", Title: "Second oldest blocked", Status: "open", CreatedAt: now.Add(-4 * time.Hour), DependsOn: []string{"blocker"}}
	store.Issues["blocked4"] = &Issue{ID: "blocked4", Title: "Second newest blocked", Status: "open", CreatedAt: now.Add(-1 * time.Hour), DependsOn: []string{"blocker"}}

	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := stripANSI(buf.String())

	// Find BLOCKED section
	blockedIdx := strings.Index(output, "BLOCKED")
	closedIdx := strings.Index(output, "CLOSED")
	if blockedIdx == -1 || closedIdx == -1 {
		t.Fatalf("expected BLOCKED and CLOSED sections in output")
	}

	// Extract just the BLOCKED section
	blockedSection := output[blockedIdx:closedIdx]

	// Find positions of each blocked issue in the BLOCKED section
	idx1 := strings.Index(blockedSection, "blocked1")
	idx2 := strings.Index(blockedSection, "blocked2")
	idx3 := strings.Index(blockedSection, "blocked3")
	idx4 := strings.Index(blockedSection, "blocked4")
	idx5 := strings.Index(blockedSection, "blocked5")

	// Verify all blocked issues are present
	if idx1 == -1 || idx2 == -1 || idx3 == -1 || idx4 == -1 || idx5 == -1 {
		t.Fatalf("expected all blocked issues in BLOCKED section, got: %s", blockedSection)
	}

	// Verify order: newest to oldest (blocked5, blocked4, blocked3, blocked2, blocked1)
	if idx5 >= idx4 || idx4 >= idx3 || idx3 >= idx2 || idx2 >= idx1 {
		t.Errorf("expected blocked issues sorted by CreatedAt (newest first): blocked5(%d), blocked4(%d), blocked3(%d), blocked2(%d), blocked1(%d)\nblocked section:\n%s",
			idx5, idx4, idx3, idx2, idx1, blockedSection)
	}
}

func TestListSortsClosedByUpdatedAt(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	now := time.Now()

	// Create 5 closed issues with different update times, added in shuffled order
	store.Issues["closed3"] = &Issue{ID: "closed3", Title: "Third oldest update", Status: "closed", CreatedAt: now.Add(-10 * time.Hour), UpdatedAt: now.Add(-3 * time.Hour)}
	store.Issues["closed1"] = &Issue{ID: "closed1", Title: "Oldest update", Status: "closed", CreatedAt: now.Add(-10 * time.Hour), UpdatedAt: now.Add(-5 * time.Hour)}
	store.Issues["closed5"] = &Issue{ID: "closed5", Title: "Newest update", Status: "closed", CreatedAt: now.Add(-10 * time.Hour), UpdatedAt: now}
	store.Issues["closed2"] = &Issue{ID: "closed2", Title: "Second oldest update", Status: "closed", CreatedAt: now.Add(-10 * time.Hour), UpdatedAt: now.Add(-4 * time.Hour)}
	store.Issues["closed4"] = &Issue{ID: "closed4", Title: "Second newest update", Status: "closed", CreatedAt: now.Add(-10 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour)}

	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list"})
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	output := stripANSI(buf.String())

	// Find CLOSED section
	closedIdx := strings.Index(output, "CLOSED")
	if closedIdx == -1 {
		t.Fatalf("expected CLOSED section in output")
	}

	// Extract just the CLOSED section (from CLOSED to end)
	closedSection := output[closedIdx:]

	// Find positions of each closed issue in the CLOSED section
	idx1 := strings.Index(closedSection, "closed1")
	idx2 := strings.Index(closedSection, "closed2")
	idx3 := strings.Index(closedSection, "closed3")
	idx4 := strings.Index(closedSection, "closed4")
	idx5 := strings.Index(closedSection, "closed5")

	// Verify all closed issues are present
	if idx1 == -1 || idx2 == -1 || idx3 == -1 || idx4 == -1 || idx5 == -1 {
		t.Fatalf("expected all closed issues in CLOSED section, got: %s", closedSection)
	}

	// Verify order: newest to oldest by UpdatedAt (closed5, closed4, closed3, closed2, closed1)
	if idx5 >= idx4 || idx4 >= idx3 || idx3 >= idx2 || idx2 >= idx1 {
		t.Errorf("expected closed issues sorted by UpdatedAt (newest first): closed5(%d), closed4(%d), closed3(%d), closed2(%d), closed1(%d)\nclosed section:\n%s",
			idx5, idx4, idx3, idx2, idx1, closedSection)
	}
}

func TestListCommandWithLimitFlag(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	now := time.Now()

	// Create 5 ready issues
	for i := 1; i <= 5; i++ {
		store.Issues[fmt.Sprintf("ready%d", i)] = &Issue{
			ID:        fmt.Sprintf("ready%d", i),
			Title:     fmt.Sprintf("Ready issue %d", i),
			Status:    "open",
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	// Create a blocker and 5 blocked issues
	store.Issues["blocker"] = &Issue{
		ID:        "blocker",
		Title:     "Blocker issue",
		Status:    "open",
		CreatedAt: now.Add(-10 * time.Hour),
	}
	for i := 1; i <= 5; i++ {
		store.Issues[fmt.Sprintf("blocked%d", i)] = &Issue{
			ID:        fmt.Sprintf("blocked%d", i),
			Title:     fmt.Sprintf("Blocked issue %d", i),
			Status:    "open",
			CreatedAt: now.Add(-time.Duration(5+i) * time.Hour),
			DependsOn: []string{"blocker"},
		}
	}

	// Create 5 closed issues
	for i := 1; i <= 5; i++ {
		store.Issues[fmt.Sprintf("closed%d", i)] = &Issue{
			ID:        fmt.Sprintf("closed%d", i),
			Title:     fmt.Sprintf("Closed issue %d", i),
			Status:    "closed",
			CreatedAt: now.Add(-time.Duration(15+i) * time.Hour),
			UpdatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--limit", "2"})
	if err != nil {
		t.Fatalf("list --limit command failed: %v", err)
	}

	output := stripANSI(buf.String())

	// Count issues in each section
	readyIdx := strings.Index(output, "READY")
	blockedIdx := strings.Index(output, "BLOCKED")
	closedIdx := strings.Index(output, "CLOSED")

	readySection := output[readyIdx:blockedIdx]
	blockedSection := output[blockedIdx:closedIdx]
	closedSection := output[closedIdx:]

	// Count ready issues (should be max 2)
	readyCount := 0
	for i := 1; i <= 5; i++ {
		if strings.Contains(readySection, fmt.Sprintf("ready%d", i)) {
			readyCount++
		}
	}
	if readyCount != 2 {
		t.Errorf("expected 2 ready issues with --limit 2, got %d in section:\n%s", readyCount, readySection)
	}

	// Count blocked issues (should be max 2)
	blockedCount := 0
	for i := 1; i <= 5; i++ {
		if strings.Contains(blockedSection, fmt.Sprintf("blocked%d", i)) {
			blockedCount++
		}
	}
	if blockedCount != 2 {
		t.Errorf("expected 2 blocked issues with --limit 2, got %d in section:\n%s", blockedCount, blockedSection)
	}

	// Count closed issues (should be max 2)
	closedIssueCount := 0
	for i := 1; i <= 5; i++ {
		if strings.Contains(closedSection, fmt.Sprintf("closed%d", i)) {
			closedIssueCount++
		}
	}
	if closedIssueCount != 2 {
		t.Errorf("expected 2 closed issues with --limit 2, got %d in section:\n%s", closedIssueCount, closedSection)
	}
}

func TestListCommandWithLimitZero(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	_, _ = store.AddIssue("Ready issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--limit", "0"})
	if err != nil {
		t.Fatalf("list --limit 0 command failed: %v", err)
	}

	output := stripANSI(buf.String())

	// With limit 0 (ignored), all issues should be shown
	if !strings.Contains(output, "Ready issue") {
		t.Errorf("expected issue to be shown when limit is 0 (ignored), got: %s", output)
	}
}

func TestListCommandWithNegativeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	_, _ = store.AddIssue("Ready issue")
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--limit", "-5"})
	if err != nil {
		t.Fatalf("list --limit -5 command failed: %v", err)
	}

	output := stripANSI(buf.String())

	// With negative limit (ignored), all issues should be shown
	if !strings.Contains(output, "Ready issue") {
		t.Errorf("expected issue to be shown when limit is negative (ignored), got: %s", output)
	}
}

func TestListCommandWithLimitAndReady(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store, _ := LoadStore(filePath)
	now := time.Now()

	// Create 5 ready issues
	for i := 1; i <= 5; i++ {
		store.Issues[fmt.Sprintf("ready%d", i)] = &Issue{
			ID:        fmt.Sprintf("ready%d", i),
			Title:     fmt.Sprintf("Ready issue %d", i),
			Status:    "open",
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}
	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--limit", "2", "--ready"})
	if err != nil {
		t.Fatalf("list --limit --ready command failed: %v", err)
	}

	output := stripANSI(buf.String())

	// Count ready issues (should be max 2)
	readyCount := 0
	for i := 1; i <= 5; i++ {
		if strings.Contains(output, fmt.Sprintf("ready%d", i)) {
			readyCount++
		}
	}
	if readyCount != 2 {
		t.Errorf("expected 2 ready issues with --limit 2 --ready, got %d\noutput: %s", readyCount, output)
	}

	// Should not show BLOCKED or CLOSED sections
	if strings.Contains(output, "BLOCKED") {
		t.Error("expected no BLOCKED section with --ready flag")
	}
	if strings.Contains(output, "CLOSED") {
		t.Error("expected no CLOSED section with --ready flag")
	}
}

func TestListCommandLimitIndicator(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	now := time.Now()

	// Create 5 ready issues
	for i := 1; i <= 5; i++ {
		store.Issues[fmt.Sprintf("ready%d", i)] = &Issue{
			ID:        fmt.Sprintf("ready%d", i),
			Title:     fmt.Sprintf("Ready issue %d", i),
			Status:    "open",
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	// Create a blocker and 3 blocked issues
	store.Issues["blocker"] = &Issue{
		ID:        "blocker",
		Title:     "Blocker issue",
		Status:    "open",
		CreatedAt: now.Add(-10 * time.Hour),
	}
	for i := 1; i <= 3; i++ {
		store.Issues[fmt.Sprintf("blocked%d", i)] = &Issue{
			ID:        fmt.Sprintf("blocked%d", i),
			Title:     fmt.Sprintf("Blocked issue %d", i),
			Status:    "open",
			CreatedAt: now.Add(-time.Duration(5+i) * time.Hour),
			DependsOn: []string{"blocker"},
		}
	}

	// Create 7 closed issues
	for i := 1; i <= 7; i++ {
		store.Issues[fmt.Sprintf("closed%d", i)] = &Issue{
			ID:        fmt.Sprintf("closed%d", i),
			Title:     fmt.Sprintf("Closed issue %d", i),
			Status:    "closed",
			CreatedAt: now.Add(-time.Duration(15+i) * time.Hour),
			UpdatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--limit", "2"})
	if err != nil {
		t.Fatalf("list --limit command failed: %v", err)
	}

	output := buf.String()
	strippedOutput := stripANSI(output)

	// READY section: 5 ready + 1 blocker (no deps) = 6 total, limited to 2 - should show indicator
	if !strings.Contains(strippedOutput, "READY  (2 of 6)") {
		t.Errorf("expected 'READY  (2 of 6)' in output, got:\n%s", strippedOutput)
	}

	// BLOCKED section: 3 blocked issues (all depend on blocker), limited to 2 - should show indicator
	if !strings.Contains(strippedOutput, "BLOCKED  (2 of 3)") {
		t.Errorf("expected 'BLOCKED  (2 of 3)' in output, got:\n%s", strippedOutput)
	}

	// CLOSED section: 7 total, limited to 2 - should show indicator
	if !strings.Contains(strippedOutput, "CLOSED  (2 of 7)") {
		t.Errorf("expected 'CLOSED  (2 of 7)' in output, got:\n%s", strippedOutput)
	}
}

func TestListCommandLimitIndicatorNotShown(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "mint-issues.yaml")
	t.Setenv("MINT_STORE_FILE", filePath)

	store := NewStore()
	now := time.Now()

	// Create 2 ready issues (limit is 5, so no indicator should show)
	for i := 1; i <= 2; i++ {
		store.Issues[fmt.Sprintf("ready%d", i)] = &Issue{
			ID:        fmt.Sprintf("ready%d", i),
			Title:     fmt.Sprintf("Ready issue %d", i),
			Status:    "open",
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	_ = store.Save(filePath)

	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "list", "--limit", "5"})
	if err != nil {
		t.Fatalf("list --limit command failed: %v", err)
	}

	output := buf.String()
	strippedOutput := stripANSI(output)

	// READY section: 2 total, limit 5 (not limited) - should NOT show indicator
	if strings.Contains(strippedOutput, "READY (") {
		t.Errorf("expected no limit indicator for READY when limit is not reached, got:\n%s", strippedOutput)
	}
}
