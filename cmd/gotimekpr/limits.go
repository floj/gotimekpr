package main

import (
	"context"
	"fmt"
	"time"

	"github.com/floj/gotimekpr/pkg/config"
	"github.com/floj/gotimekpr/pkg/db"
	"github.com/floj/gotimekpr/pkg/quota"
	"github.com/urfave/cli/v3"
)

func withQuotaManager(fn func(ctx context.Context, c *cli.Command, qm *quota.QuotaManager) error) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		conf, err := config.LoadConfig()
		if err != nil {
			return err
		}
		conf.NoLogout = c.Bool("no-logout")

		dbq, db, err := db.Open(ctx, conf)
		if err != nil {
			return err
		}
		defer db.Close()

		qm := quota.NewQuotaManager(dbq)
		return fn(ctx, c, qm)
	}
}

func cmdLimits() *cli.Command {
	return &cli.Command{
		Name:    "limits",
		Aliases: []string{"l"},
		Usage:   "retrieves today's limit",
		Action: withQuotaManager(func(ctx context.Context, c *cli.Command, qm *quota.QuotaManager) error {
			limit := qm.GetDateLimitToday(ctx)
			fmt.Printf("limit for today: %s\n", limit)
			return nil
		}),
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "adds additional time to todays limit",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "duration",
						UsageText: "duration to add to today's limit, e.g. 30m, 1h, etc.",
						Config: cli.StringConfig{
							TrimSpace: true,
						},
					},
				},
				Action: withQuotaManager(func(ctx context.Context, c *cli.Command, qm *quota.QuotaManager) error {
					d, err := time.ParseDuration(c.StringArg("duration"))
					if err != nil {
						return err
					}
					limit, err := qm.AddToDateLimitToday(ctx, d)
					if err != nil {
						return err
					}
					fmt.Printf("new limit for today: %s\n", limit)
					return nil
				}),
			},
			{
				Name:    "set",
				Aliases: []string{"s"},
				Usage:   "sets today's limit to a specific duration",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "duration",
						UsageText: "duration to set today's limit, e.g. 30m, 1h, etc.",
						Config: cli.StringConfig{
							TrimSpace: true,
						},
					},
				},
				Action: withQuotaManager(func(ctx context.Context, c *cli.Command, qm *quota.QuotaManager) error {
					d, err := time.ParseDuration(c.StringArg("duration"))
					if err != nil {
						return err
					}
					limit, err := qm.SetDateLimitToday(ctx, d)
					if err != nil {
						return err
					}
					fmt.Printf("new limit for today: %s\n", limit)
					return nil
				}),
			},
		}}

}
