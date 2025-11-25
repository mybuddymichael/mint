package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func setPrefixAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() == 0 {
		return fmt.Errorf("new prefix is required")
	}

	newPrefix := cmd.Args().First()

	filePath, err := GetStoreFilePath()
	if err != nil {
		return err
	}

	store, err := LoadStore(filePath)
	if err != nil {
		return err
	}

	if err := store.SetPrefix(newPrefix); err != nil {
		return err
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	w := cmd.Root().Writer
	_, err = fmt.Fprintf(w, "Prefix set to \"%s\" and all issues updated\n", store.Prefix)
	return err
}
