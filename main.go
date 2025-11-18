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
		Name:  "mint",
		Usage: "A simple command line tool to create and track work.",
		Commands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new issue",
				ArgsUsage: "<title>",
				Action:    createAction,
			},
			{
				Name:  "list",
				Usage: "List all issues",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "open",
						Usage: "Only show open issues",
					},
				},
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
			{
				Name:      "set-prefix",
				Usage:     "Change the issue ID prefix",
				ArgsUsage: "<new-prefix>",
				Action:    setPrefixAction,
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

	_, err = fmt.Fprintf(cmd.Root().Writer, "Created issue %s\n", store.FormatID(issue.ID))
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

	// Filter by status if --open flag is set
	openOnly := cmd.Bool("open")
	if openOnly {
		filtered := make([]*Issue, 0)
		for _, issue := range issues {
			if issue.Status == "open" {
				filtered = append(filtered, issue)
			}
		}
		issues = filtered
	}

	// Handle empty results
	if len(issues) == 0 {
		if openOnly {
			_, err := fmt.Fprintln(w, "No open issues found.")
			return err
		}
		_, err := fmt.Fprintln(w, "No issues found.")
		return err
	}

	header := "All issues:"
	if openOnly {
		header = "Open issues:"
	}
	if _, err := fmt.Fprintln(w, header); err != nil {
		return err
	}

	for _, issue := range issues {
		formattedID := store.FormatID(issue.ID)
		if _, err := fmt.Fprintf(w, "%s %s %s\n", formattedID, issue.Status, issue.Title); err != nil {
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

	// Add comment
	if comment := cmd.String("comment"); comment != "" {
		if err := store.AddComment(fullID, comment); err != nil {
			return err
		}
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	w := cmd.Root().Writer
	_, err = fmt.Fprintf(w, "Updated %s\n", store.FormatID(fullID))
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

	// Resolve partial ID to full ID
	fullID, err := store.ResolveIssueID(id)
	if err != nil {
		return err
	}

	reason := cmd.String("reason")
	if err := store.CloseIssue(fullID, reason); err != nil {
		return err
	}

	if err := store.Save(filePath); err != nil {
		return err
	}

	w := cmd.Root().Writer
	_, err = fmt.Fprintf(w, "Closed issue %s\n", store.FormatID(fullID))
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
	_, err = fmt.Fprintf(w, "Prefix set to \"%s\" and all issues updated\n", newPrefix)
	return err
}
