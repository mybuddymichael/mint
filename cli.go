package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var version = "dev-?"

func main() {
	cmd := newCommand()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func newCommand() *cli.Command {
	return &cli.Command{
		Name:                  "mint",
		Usage:                 "A simple command line tool to create and track work.",
		Version:               version,
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:      "create",
				Aliases:   []string{"add", "c", "a"},
				Usage:     "Create a new issue",
				ArgsUsage: "<title>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "description",
						Usage: "Description for the issue (added as first comment)",
					},
					&cli.StringFlag{
						Name:  "comment",
						Usage: "Add a comment to the issue",
					},
				},
				Action: createAction,
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List all issues",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "open",
						Usage: "Only show open issues",
					},
					&cli.BoolFlag{
						Name:  "ready",
						Usage: "Only show ready issues",
					},
				},
				Action: listAction,
			},
			{
				Name:      "show",
				Aliases:   []string{"s"},
				Usage:     "Show an issue and its details",
				ArgsUsage: "<issue-id>",
				Action:    showAction,
			},
			{
				Name:      "update",
				Aliases:   []string{"u"},
				Usage:     "Update an issue",
				ArgsUsage: "<issue-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "title",
						Aliases: []string{"t"},
						Usage:   "New title for the issue",
					},
					&cli.StringSliceFlag{
						Name:    "depends-on",
						Aliases: []string{"d"},
						Usage:   "Add dependency (can be repeated)",
					},
					&cli.StringSliceFlag{
						Name:    "blocks",
						Aliases: []string{"b"},
						Usage:   "Add blocked issues (can be repeated)",
					},
					&cli.StringSliceFlag{
						Name:    "remove-depends-on",
						Aliases: []string{"rd"},
						Usage:   "Remove dependency (can be repeated)",
					},
					&cli.StringSliceFlag{
						Name:    "remove-blocks",
						Aliases: []string{"rb"},
						Usage:   "Remove blocked issues (can be repeated)",
					},
					&cli.StringFlag{
						Name:    "comment",
						Aliases: []string{"c"},
						Usage:   "Add a comment",
					},
				},
				Action: updateAction,
			},
			{
				Name:      "close",
				Aliases:   []string{"cl"},
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
				Aliases:   []string{"o"},
				Usage:     "Re-open a closed issue",
				ArgsUsage: "<issue-id>",
				Action:    openAction,
			},
			{
				Name:      "delete",
				Aliases:   []string{"d"},
				Usage:     "Delete an issue",
				ArgsUsage: "<issue-id>",
				Action:    deleteAction,
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
