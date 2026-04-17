package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v3"
)

func cmdLogs() *cli.Command {
	return &cli.Command{
		Name:  "logs",
		Usage: "show daemon logs",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "follow",
				Aliases: []string{"f"},
				Usage:   "follow log output",
			},
			&cli.IntFlag{
				Name:    "lines",
				Aliases: []string{"n"},
				Usage:   "number of lines to show",
				Value:   50,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			fmt.Println("Showing daemon logs. Use Ctrl+C to stop.")
			args := []string{"--user", "-u", "gotimekpr.service", "-n", strconv.Itoa(c.Int("lines"))}
			if c.Bool("follow") {
				args = append(args, "-f")
			}
			return runCmd(ctx, "journalctl", args...)
		},
	}

}
