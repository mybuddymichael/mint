package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func createAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() == 0 {
		return fmt.Errorf("title is required")
	}

	title := cmd.Args().First()

	filePath, err := GetStoreFilePath()
	if err != nil {
		return err
	}

	store, err := LoadStore(filePath)
	if err != nil {
		return err
	}

	// Pre-validate relationship IDs exist
	dependsOnIDs := cmd.StringSlice("depends-on")
	for _, depID := range dependsOnIDs {
		if _, err := store.ResolveIssueID(depID); err != nil {
			return fmt.Errorf("dependency issue not found: %w", err)
		}
	}

	blocksIDs := cmd.StringSlice("blocks")
	for _, blockID := range blocksIDs {
		if _, err := store.ResolveIssueID(blockID); err != nil {
			return fmt.Errorf("blocked issue not found: %w", err)
		}
	}

	issue, err := store.AddIssue(title)
	if err != nil {
		return err
	}

	// Add dependencies
	for _, depID := range dependsOnIDs {
		if err := store.AddDependency(issue.ID, depID); err != nil {
			return err
		}
	}

	// Add blockers
	for _, blockID := range blocksIDs {
		if err := store.AddBlocker(issue.ID, blockID); err != nil {
			return err
		}
	}

	// Add description as first comment if provided
	if description := cmd.String("description"); description != "" {
		if err := store.AddComment(issue.ID, description); err != nil {
			return err
		}
	}

	// Add comment if provided
	if comment := cmd.String("comment"); comment != "" {
		if err := store.AddComment(issue.ID, comment); err != nil {
			return err
		}
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	w := cmd.Root().Writer
	if _, err := fmt.Fprintf(w, "\x1b[1;32m✔︎ Created issue\x1b[0m\n\n"); err != nil {
		return err
	}
	return PrintIssueDetails(w, issue, store)
}
