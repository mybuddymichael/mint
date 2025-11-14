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
		Usage: "A simple command line tool to create and track work.",
		Commands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new issue",
				ArgsUsage: "<title>",
				Action:    createAction,
			},
			{
				Name:   "list",
				Usage:  "List all issues",
				Action: listAction,
			},
			{
				Name:      "show",
				Usage:     "Show an issue and its details",
				ArgsUsage: "<issue-id>",
				Action:    showAction,
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

func listAction(ctx context.Context, cmd *cli.Command) error {
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

	// Handle empty store
	if len(issues) == 0 {
		_, err := fmt.Fprintln(w, "No issues found.")
		return err
	}

	if _, err := fmt.Fprintln(w, "All issues:"); err != nil {
		return err
	}

	for _, issue := range issues {
		if _, err := fmt.Fprintf(w, "%s \"%s\"\n", issue.ID, issue.Title); err != nil {
			return err
		}
	}

	return nil
}

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
	if _, err := fmt.Fprintf(w, "ID:      %s\n", issue.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Title:   \"%s\"\n", issue.Title); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "Status:  %s\n", issue.Status)
	return err
}
