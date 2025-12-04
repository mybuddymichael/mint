package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
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
