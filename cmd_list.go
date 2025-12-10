package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/urfave/cli/v3"
)

func listAction(_ context.Context, cmd *cli.Command) error {
	filePath, err := GetStoreFilePath()
	if err != nil {
		return err
	}

	w := cmd.Root().Writer

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_, err := fmt.Fprintln(w, "No issues file found.")
		return err
	}

	store, err := LoadStore(filePath)
	if err != nil {
		return err
	}

	issues := store.ListIssues()

	// Handle completely empty store
	if len(issues) == 0 {
		_, err := fmt.Fprintln(w, "No issues found.")
		return err
	}

	// Calculate max ID length and separate issues into ready, blocked, and closed
	maxIDLen := 0
	readyIssues := make([]*Issue, 0)
	blockedIssues := make([]*Issue, 0)
	closedIssues := make([]*Issue, 0)
	for _, issue := range issues {
		if len(issue.ID) > maxIDLen {
			maxIDLen = len(issue.ID)
		}
		if issue.Status == "open" {
			if store.IsReady(issue) {
				readyIssues = append(readyIssues, issue)
			} else {
				blockedIssues = append(blockedIssues, issue)
			}
		} else {
			closedIssues = append(closedIssues, issue)
		}
	}

	openOnly := cmd.Bool("open")
	readyOnly := cmd.Bool("ready")
	limit := cmd.Int("limit")

	// Sort issues by timestamps (only what we'll display)
	sortByCreatedAt(readyIssues)
	if !readyOnly {
		sortByCreatedAt(blockedIssues)
	}
	if !openOnly && !readyOnly {
		sortByUpdatedAt(closedIssues)
	}

	// Track original counts before applying limit
	readyTotalCount := len(readyIssues)
	blockedTotalCount := len(blockedIssues)
	closedTotalCount := len(closedIssues)

	// Apply limit to each section (if limit > 0)
	if limit > 0 {
		if len(readyIssues) > limit {
			readyIssues = readyIssues[:limit]
		}
		if len(blockedIssues) > limit {
			blockedIssues = blockedIssues[:limit]
		}
		if len(closedIssues) > limit {
			closedIssues = closedIssues[:limit]
		}
	}

	// Display READY section
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	readyHeader := "\033[48;5;2m\033[38;5;0m READY \033[0m"
	// Add limit indicator if section is actually limited
	if limit > 0 && len(readyIssues) < readyTotalCount {
		readyHeader += fmt.Sprintf(" \033[38;5;8m(%d of %d)\033[0m", len(readyIssues), readyTotalCount)
	}
	if _, err := fmt.Fprintln(w, readyHeader); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if len(readyIssues) == 0 {
		if _, err := fmt.Fprintln(w, "   (No ready issues.)"); err != nil {
			return err
		}
	} else {
		if err := printIssueList(w, readyIssues, maxIDLen, store); err != nil {
			return err
		}
	}

	// Display BLOCKED section (skip if --ready flag is set)
	if !readyOnly {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}

		blockedHeader := "\033[48;5;1m\033[38;5;0m BLOCKED \033[0m"
		// Add limit indicator if section is actually limited
		if limit > 0 && len(blockedIssues) < blockedTotalCount {
			blockedHeader += fmt.Sprintf(" \033[38;5;8m(%d of %d)\033[0m", len(blockedIssues), blockedTotalCount)
		}
		if _, err := fmt.Fprintln(w, blockedHeader); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}

		if len(blockedIssues) == 0 {
			if _, err := fmt.Fprintln(w, "   (No blocked issues.)"); err != nil {
				return err
			}
		} else {
			if err := printIssueList(w, blockedIssues, maxIDLen, store); err != nil {
				return err
			}
		}
	}

	// Display CLOSED section (skip if --open or --ready flag is set)
	if !openOnly && !readyOnly {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}

		closedHeader := "\033[48;5;0m\033[38;5;15m CLOSED \033[0m"
		// Add limit indicator if section is actually limited
		if limit > 0 && len(closedIssues) < closedTotalCount {
			closedHeader += fmt.Sprintf(" \033[38;5;8m(%d of %d)\033[0m", len(closedIssues), closedTotalCount)
		}
		if _, err := fmt.Fprintln(w, closedHeader); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}

		if len(closedIssues) == 0 {
			if _, err := fmt.Fprintln(w, "   (No closed issues.)"); err != nil {
				return err
			}
		} else {
			if err := printIssueList(w, closedIssues, maxIDLen, store); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	return nil
}

func sortByCreatedAt(issues []*Issue) {
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].CreatedAt.After(issues[j].CreatedAt)
	})
}

func sortByUpdatedAt(issues []*Issue) {
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].UpdatedAt.After(issues[j].UpdatedAt)
	})
}
