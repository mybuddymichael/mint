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
			fmt.Fprintln(cmd.Writer, "Mint issue tracker")
			return nil
		},
	}
}
