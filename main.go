package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := newCommand()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func newCommand() *cli.Command {
	return &cli.Command{
		Name:  "mt",
		Usage: "Simple command line tool to create and track work on a software project",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_, err := fmt.Fprintln(cmd.Writer, "Mint issue tracker")
			return err
		},
		Commands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new issue",
				ArgsUsage: "<title>",
				Action:    createAction,
			},
		},
	}
}

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

	if err := store.Save(filePath); err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.Root().Writer, "Created issue %s\n", issue.ID)
	return err
}
