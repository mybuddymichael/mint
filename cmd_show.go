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

	return PrintIssueDetails(cmd.Root().Writer, issue, store)
}
