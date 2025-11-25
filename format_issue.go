package main

import (
	"fmt"
	"io"
	"strings"
)

// PrintIssueDetails prints full issue details including ID, Title, Status,
// Dependencies, Blocks, and Comments
func PrintIssueDetails(w io.Writer, issue *Issue, store *Store) error {
	if _, err := fmt.Fprintf(w, "\033[1m\033[38;5;5mID\033[0m      %s\n", store.FormatID(issue.ID)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\033[1m\033[38;5;5mTitle\033[0m   %s\n", issue.Title); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\033[1m\033[38;5;5mStatus\033[0m  %s\n", issue.Status); err != nil {
		return err
	}
	if len(issue.DependsOn) > 0 {
		if _, err := fmt.Fprintln(w, "\033[1m\033[38;5;5mDepends on\033[0m"); err != nil {
			return err
		}
		for _, depID := range issue.DependsOn {
			dep, err := store.GetIssue(depID)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "  %s %s\n", store.FormatID(dep.ID), dep.Title); err != nil {
				return err
			}
		}
	}
	if len(issue.Blocks) > 0 {
		if _, err := fmt.Fprintln(w, "\033[1m\033[38;5;5mBlocks\033[0m"); err != nil {
			return err
		}
		for _, blockID := range issue.Blocks {
			blocked, err := store.GetIssue(blockID)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "  %s %s\n", store.FormatID(blocked.ID), blocked.Title); err != nil {
				return err
			}
		}
	}
	if len(issue.Comments) > 0 {
		if _, err := fmt.Fprintln(w, "\033[1m\033[38;5;5mComments\033[0m"); err != nil {
			return err
		}
		for _, comment := range issue.Comments {
			if _, err := fmt.Fprintf(w, "  %s\n", comment); err != nil {
				return err
			}
		}
	}
	return nil
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
