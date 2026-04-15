package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

func cmdVersion() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Prints version info",
		Action: func(ctx context.Context, c *cli.Command) error {
			fmt.Printf("gotimekpr %s (built %s)\n", version, buildDate)
			return nil
		},
	}
}
