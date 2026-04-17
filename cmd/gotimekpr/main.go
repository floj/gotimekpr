package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v3"
)

var (
	version   = "dev"
	buildDate = "unknown"
)

var runCmd = func(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	cmd := &cli.Command{
		Name:  "gotimekpr",
		Usage: "timekpr-next like app written in Go",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "shows debug logs",
				Sources: cli.EnvVars("GOTIMEKPR_DEBUG"),
			},
		},
		Commands: []*cli.Command{
			cmdDaemon(),
			cmdLimits(),
			cmdUsage(),
			cmdInstall(),
			cmdVersion(),
			cmdLogs(),
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			lvl := slog.LevelInfo
			if c.Bool("debug") {
				lvl = slog.LevelDebug
			}
			slog.SetDefault(slog.New(
				tint.NewHandler(os.Stderr, &tint.Options{
					Level:      lvl,
					TimeFormat: time.Kitchen,
				}),
			))
			return ctx, nil
		},
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.Run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
