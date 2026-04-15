package main

import (
	"context"
	"log/slog"

	"github.com/floj/gotimekpr/pkg/config"
	"github.com/floj/gotimekpr/pkg/daemon"
	"github.com/urfave/cli/v3"
)

func cmdDaemon() *cli.Command {
	return &cli.Command{
		Name:  "daemon",
		Usage: "run the daemon",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "no-logout",
				Usage: "if true, the daemon will not log out the user when the limit is exceeded, useful for testing and debugging.",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {

			conf, err := config.LoadConfig()
			if err != nil {
				return err
			}
			conf.NoLogout = c.Bool("no-logout")

			slog.Info("starting daemon", "db", conf.DBURL)
			d, err := daemon.NewDaemon(ctx, conf)
			if err != nil {
				return err
			}
			defer d.Close()
			return d.Run()
		},
	}
}
