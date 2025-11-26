package main

import (
	"fmt"
	"io"
	"strings"
)

// PrintIssueDetails prints full issue details including ID, Title, Status,
// Dependencies, Blocks, and Comments
func PrintIssueDetails(w io.Writer, issue *Issue, store *Store) error {
	var b strings.Builder
	fmt.Fprintf(&b, "\033[1m\033[38;5;5mID\033[0m      %s\n", store.FormatID(issue.ID))
	fmt.Fprintf(&b, "\033[1m\033[38;5;5mTitle\033[0m   %s\n", issue.Title)
	fmt.Fprintf(&b, "\033[1m\033[38;5;5mStatus\033[0m  %s\n", issue.Status)
	if len(issue.DependsOn) > 0 {
		fmt.Fprintln(&b, "\033[1m\033[38;5;5mDepends on\033[0m")
		for _, depID := range issue.DependsOn {
			dep, err := store.GetIssue(depID)
			if err != nil {
				return err
			}
			fmt.Fprintf(&b, "  %s %s %s\n", store.FormatID(dep.ID), dep.Status, dep.Title)
		}
	}
	if len(issue.Blocks) > 0 {
		fmt.Fprintln(&b, "\033[1m\033[38;5;5mBlocks\033[0m")
		for _, blockID := range issue.Blocks {
			blocked, err := store.GetIssue(blockID)
			if err != nil {
				return err
			}
			fmt.Fprintf(&b, "  %s %s %s\n", store.FormatID(blocked.ID), blocked.Status, blocked.Title)
		}
	}
	if len(issue.Comments) > 0 {
		fmt.Fprintln(&b, "\033[1m\033[38;5;5mComments\033[0m")
		for _, comment := range issue.Comments {
			fmt.Fprintf(&b, "  %s\n", comment)
		}
	}
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
