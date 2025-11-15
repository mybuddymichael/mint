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
			{
				Name:      "update",
				Usage:     "Update an issue",
				ArgsUsage: "<issue-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "title",
						Usage: "New title for the issue",
					},
					&cli.StringSliceFlag{
						Name:  "depends-on",
						Usage: "Add dependency (can be repeated)",
					},
					&cli.StringSliceFlag{
						Name:  "blocks",
						Usage: "Add blocked issues (can be repeated)",
					},
					&cli.StringFlag{
						Name:  "comment",
						Usage: "Add a comment",
					},
				},
				Action: updateAction,
			},
			{
				Name:      "close",
				Usage:     "Close an issue",
				ArgsUsage: "<issue-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "reason",
						Usage: "Reason for closing",
					},
				},
				Action: closeAction,
			},
			{
				Name:      "open",
				Usage:     "Re-open a closed issue",
				ArgsUsage: "<issue-id>",
				Action:    openAction,
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

	// Verify issue exists
	if _, err := store.GetIssue(id); err != nil {
		return err
	}

	// Update title
	if title := cmd.String("title"); title != "" {
		if err := store.UpdateIssueTitle(id, title); err != nil {
			return err
		}
	}

	// Add dependencies
	if dependsOn := cmd.StringSlice("depends-on"); len(dependsOn) > 0 {
		for _, depID := range dependsOn {
			if err := store.AddDependency(id, depID); err != nil {
				return err
			}
		}
	}

	// Add blockers
	if blocks := cmd.StringSlice("blocks"); len(blocks) > 0 {
		for _, blockID := range blocks {
			if err := store.AddBlocker(id, blockID); err != nil {
				return err
			}
		}
	}

	// Add comment
	if comment := cmd.String("comment"); comment != "" {
		if err := store.AddComment(id, comment); err != nil {
			return err
		}
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	w := cmd.Root().Writer
	_, err = fmt.Fprintf(w, "Updated %s\n", id)
	return err
}

func closeAction(ctx context.Context, cmd *cli.Command) error {
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

	reason := cmd.String("reason")
	if err := store.CloseIssue(id, reason); err != nil {
		return err
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	w := cmd.Root().Writer
	_, err = fmt.Fprintf(w, "Closed issue %s\n", id)
	return err
}

func openAction(ctx context.Context, cmd *cli.Command) error {
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

	if err := store.ReopenIssue(id); err != nil {
		return err
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	w := cmd.Root().Writer
	_, err = fmt.Fprintf(w, "Re-opened issue %s\n", id)
	return err
}
