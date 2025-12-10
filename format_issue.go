package main

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// formatRelativeTime returns a relative time string like "2d ago" or "5m ago"
// for the given time. It does not include parentheses.
func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)

	// Handle seconds
	if duration < time.Minute {
		seconds := int(duration.Seconds())
		return fmt.Sprintf("%ds ago", seconds)
	}

	// Handle minutes
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	}

	// Handle hours
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}

	// Handle days
	days := int(duration.Hours() / 24)
	return fmt.Sprintf("%dd ago", days)
}

// PrintIssueDetails prints full issue details including ID, Title, Status,
// Dependencies, Blocks, and Comments
func PrintIssueDetails(w io.Writer, issue *Issue, store *Store) error {
	var b strings.Builder
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "\033[1m\033[38;5;5mID\033[0m      %s\n", store.FormatID(issue.ID))
	fmt.Fprintf(&b, "\033[1m\033[38;5;5mTitle\033[0m   %s\n", issue.Title)
	fmt.Fprintf(&b, "\033[1m\033[38;5;5mStatus\033[0m  %s\n", issue.Status)
	if !issue.CreatedAt.IsZero() {
		fmt.Fprintf(&b, "\033[1m\033[38;5;5mCreated\033[0m %s (%s)\n", issue.CreatedAt.Format(time.DateTime), formatRelativeTime(issue.CreatedAt))
	}
	if !issue.UpdatedAt.IsZero() {
		fmt.Fprintf(&b, "\033[1m\033[38;5;5mUpdated\033[0m %s (%s)\n", issue.UpdatedAt.Format(time.DateTime), formatRelativeTime(issue.UpdatedAt))
	}
	if len(issue.DependsOn) > 0 || len(issue.Blocks) > 0 || len(issue.Comments) > 0 {
		fmt.Fprintln(&b)
	}
	if len(issue.DependsOn) > 0 {
		fmt.Fprintln(&b, "\033[1m\033[38;5;5mDepends on\033[0m")
		for _, depID := range issue.DependsOn {
			dep, err := store.GetIssue(depID)
			if err != nil {
				fmt.Fprintf(&b, "  %s (not found)\n", store.FormatID(depID))
			} else {
				fmt.Fprintf(&b, "  %s %s %s\n", store.FormatID(dep.ID), dep.Status, dep.Title)
			}
		}
	}
	if len(issue.Blocks) > 0 {
		// Add blank line between Blocks and Depends on sections
		if len(issue.DependsOn) > 0 {
			fmt.Fprintln(&b)
		}
		fmt.Fprintln(&b, "\033[1m\033[38;5;5mBlocks\033[0m")
		for _, blockID := range issue.Blocks {
			blocked, err := store.GetIssue(blockID)
			if err != nil {
				fmt.Fprintf(&b, "  %s (not found)\n", store.FormatID(blockID))
			} else {
				fmt.Fprintf(&b, "  %s %s %s\n", store.FormatID(blocked.ID), blocked.Status, blocked.Title)
			}
		}
	}
	if len(issue.Comments) > 0 {
		// Add blank line between Comments and relationship sections (Depends on/Blocks)
		if len(issue.DependsOn) > 0 || len(issue.Blocks) > 0 {
			fmt.Fprintln(&b)
		}
		fmt.Fprintln(&b, "\033[1m\033[38;5;5mComments\033[0m")
		for _, comment := range issue.Comments {
			fmt.Fprintf(&b, "  %s\n", comment)
		}
	}
	fmt.Fprintln(&b)
	_, err := fmt.Fprint(w, b.String())
	return err
}

// printIssueList prints a list of issues with aligned formatting
func printIssueList(w io.Writer, issues []*Issue, maxIDLen int, store *Store) error {
	for _, issue := range issues {
		formattedID := store.FormatID(issue.ID)
		// Pad shorter IDs so status words align across all issues
		padding := strings.Repeat(" ", 1+maxIDLen-len(issue.ID))
		if _, err := fmt.Fprintf(w, "   %s%s%s %s\n", formattedID, padding, issue.Status, issue.Title); err != nil {
			return err
		}
	}
	return nil
}
