package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func updateAction(ctx context.Context, cmd *cli.Command) error {
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

	// Resolve partial ID to full ID
	fullID, err := store.ResolveIssueID(id)
	if err != nil {
		return err
	}

	// Update title
	if title := cmd.String("title"); title != "" {
		if err := store.UpdateIssueTitle(fullID, title); err != nil {
			return err
		}
	}

	// Add dependencies
	if dependsOn := cmd.StringSlice("depends-on"); len(dependsOn) > 0 {
		for _, depID := range dependsOn {
			if err := store.AddDependency(fullID, depID); err != nil {
				return err
			}
		}
	}

	// Add blockers
	if blocks := cmd.StringSlice("blocks"); len(blocks) > 0 {
		for _, blockID := range blocks {
			if err := store.AddBlocker(fullID, blockID); err != nil {
				return err
			}
		}
	}

	// Remove dependencies
	if removeDeps := cmd.StringSlice("remove-depends-on"); len(removeDeps) > 0 {
		for _, depID := range removeDeps {
			if err := store.RemoveDependency(fullID, depID); err != nil {
				return err
			}
		}
	}

	// Remove blockers
	if removeBlocks := cmd.StringSlice("remove-blocks"); len(removeBlocks) > 0 {
		for _, blockID := range removeBlocks {
			if err := store.RemoveBlocker(fullID, blockID); err != nil {
				return err
			}
		}
	}

	// Add comment
	if comment := cmd.String("comment"); comment != "" {
		if err := store.AddComment(fullID, comment); err != nil {
			return err
		}
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	issue, err := store.GetIssue(fullID)
	if err != nil {
		return err
	}

	w := cmd.Root().Writer
	if _, err := fmt.Fprintf(w, "\x1b[1;32m✔︎ Issue updated\x1b[0m\n"); err != nil {
		return err
	}
	return PrintIssueDetails(w, issue, store)
}
