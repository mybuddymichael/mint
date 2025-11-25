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

	issue, err := store.AddIssue(title)
	if err != nil {
		return err
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
