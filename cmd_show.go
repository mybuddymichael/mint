package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func showAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() == 0 {
		return fmt.Errorf("issue ID is required")
	}

	id := cmd.Args().First()

	filePath, err := GetStoreFilePath()
	if err != nil {
		return err
	}

	store, err := LoadStore(filePath)
	if err != nil {
		return err
	}

	issue, err := store.GetIssue(id)
	if err != nil {
		return err
	}

	w := cmd.Root().Writer
	if _, err := fmt.Fprintf(w, "ID:      %s\n", store.FormatID(issue.ID)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Title:   %s\n", issue.Title); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Status:  %s\n", issue.Status); err != nil {
		return err
	}
	if len(issue.DependsOn) > 0 {
		if _, err := fmt.Fprintln(w, "Depends on:"); err != nil {
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
		if _, err := fmt.Fprintln(w, "Blocks:"); err != nil {
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
		if _, err := fmt.Fprintln(w, "Comments:"); err != nil {
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
