package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	maxIDLen := max(len(issue1.ID), len(issue2.ID))

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

	// Should have timestamps followed by blank line before Depends on
	if !strings.Contains(stripped, "Updated") {
		t.Errorf("expected Updated timestamp, got: %s", stripped)
	}
	// Check for blank line before Depends on (after timestamps)
	if !strings.Contains(stripped, "\n\nDepends on") {
		t.Errorf("expected blank line before Depends on, got: %s", stripped)
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

func TestFormatRelativeTime_Seconds(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"4 seconds ago", 4 * time.Second, "4s ago"},
		{"1 second ago", 1 * time.Second, "1s ago"},
		{"30 seconds ago", 30 * time.Second, "30s ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			past := now.Add(-tt.duration)
			result := formatRelativeTime(past)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatRelativeTime_Minutes(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"1 minute ago", 1 * time.Minute, "1m ago"},
		{"5 minutes ago", 5 * time.Minute, "5m ago"},
		{"59 minutes ago", 59 * time.Minute, "59m ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			past := now.Add(-tt.duration)
			result := formatRelativeTime(past)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatRelativeTime_Hours(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"1 hour ago", 1 * time.Hour, "1h ago"},
		{"3 hours ago", 3 * time.Hour, "3h ago"},
		{"23 hours ago", 23 * time.Hour, "23h ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			past := now.Add(-tt.duration)
			result := formatRelativeTime(past)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatRelativeTime_Days(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"1 day ago", 24 * time.Hour, "1d ago"},
		{"2 days ago", 48 * time.Hour, "2d ago"},
		{"5 days ago", 120 * time.Hour, "5d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			past := now.Add(-tt.duration)
			result := formatRelativeTime(past)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatRelativeTime_MixedUnits(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		// Mixed durations should show only the largest unit
		{"5 hours 31 minutes 4 seconds", 5*time.Hour + 31*time.Minute + 4*time.Second, "5h ago"},
		{"2 days 5 hours 31 minutes", 2*24*time.Hour + 5*time.Hour + 31*time.Minute, "2d ago"},
		{"31 minutes 45 seconds", 31*time.Minute + 45*time.Second, "31m ago"},
		{"1 hour 59 minutes 59 seconds", 1*time.Hour + 59*time.Minute + 59*time.Second, "1h ago"},
		{"23 hours 59 minutes 59 seconds", 23*time.Hour + 59*time.Minute + 59*time.Second, "23h ago"},
		{"10 days 12 hours 30 minutes", 10*24*time.Hour + 12*time.Hour + 30*time.Minute, "10d ago"},
		{"59 seconds", 59 * time.Second, "59s ago"},
		{"1 minute 30 seconds", 1*time.Minute + 30*time.Second, "1m ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			past := now.Add(-tt.duration)
			result := formatRelativeTime(past)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrintIssueDetails_WithRelativeTimestamps(t *testing.T) {
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

	// Check that Created timestamp has relative time in parentheses
	if !strings.Contains(output, "Created") {
		t.Errorf("expected 'Created' in output, got: %s", output)
	}
	// Look for pattern like "2025-12-10 11:01:25 (0s ago)" or similar
	if !strings.Contains(output, "ago)") {
		t.Errorf("expected relative time in Created field, got: %s", output)
	}

	// Check that Updated timestamp has relative time in parentheses
	if !strings.Contains(output, "Updated") {
		t.Errorf("expected 'Updated' in output, got: %s", output)
	}
}
