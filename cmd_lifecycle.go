package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func closeAction(_ context.Context, cmd *cli.Command) error {
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

	reason := cmd.String("reason")
	if err := store.CloseIssue(fullID, reason); err != nil {
		return err
	}

	issue, err := store.GetIssue(fullID)
	if err != nil {
		return err
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	w := cmd.Root().Writer
	if _, err := fmt.Fprintf(w, "\x1b[1;32m✔︎ Issue closed\x1b[0m\n"); err != nil {
		return err
	}
	return PrintIssueDetails(w, issue, store)
}

func openAction(_ context.Context, cmd *cli.Command) error {
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

	if err := store.ReopenIssue(fullID); err != nil {
		return err
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	w := cmd.Root().Writer
	_, err = fmt.Fprintf(w, "Re-opened issue %s\n", store.FormatID(fullID))
	return err
}

func deleteAction(_ context.Context, cmd *cli.Command) error {
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

	if err := store.DeleteIssue(fullID); err != nil {
		return err
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	w := cmd.Root().Writer
	_, err = fmt.Fprintf(w, "Deleted issue %s\n", store.FormatID(fullID))
	return err
}
