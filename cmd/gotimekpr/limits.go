package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/floj/gotimekpr/pkg/config"
	"github.com/floj/gotimekpr/pkg/db"
	"github.com/urfave/cli/v3"
)

func cmdLimits() *cli.Command {
	return &cli.Command{
		Name:  "limits",
		Usage: "manage limits",
		Commands: []*cli.Command{
			{
				Name:  "get",
				Usage: "retrieves today's limit",
				Action: func(ctx context.Context, c *cli.Command) error {
					conf, err := config.LoadConfig()
					if err != nil {
						return err
					}
					dbq, db, err := db.Open(ctx, conf)
					if err != nil {
						return err
					}
					defer db.Close()

					limit, err := dbq.GetDateLimitToday(ctx)

					if err == nil {
						fmt.Printf("limit for today: %s\n", time.Duration(limit.LimitMinutes)*time.Minute)
						return nil
					}
					if err == sql.ErrNoRows {
						fmt.Println("no limit set for today")
						return nil
					}
					return err
				},
			},
			{
				Name:  "add",
				Usage: "adds additional time to todays limit",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "duration",
						UsageText: "duration to add to today's limit, e.g. 30m, 1h, etc.",
						Config: cli.StringConfig{
							TrimSpace: true,
						},
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					d, err := time.ParseDuration(c.StringArg("duration"))
					if err != nil {
						return err
					}
					conf, err := config.LoadConfig()
					if err != nil {
						return err
					}
					dbq, db, err := db.Open(ctx, conf)
					if err != nil {
						return err
					}
					defer db.Close()

					limit, err := dbq.AddToDateLimitToday(ctx, int64(d.Minutes()))
					if err != nil {
						return err
					}
					fmt.Printf("new limit for today: %s\n", time.Duration(limit.LimitMinutes)*time.Minute)
					return nil
				},
			},
			{
				Name:  "set",
				Usage: "sets today's limit to a specific duration",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "duration",
						UsageText: "duration to set today's limit, e.g. 30m, 1h, etc.",
						Config: cli.StringConfig{
							TrimSpace: true,
						},
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					d, err := time.ParseDuration(c.StringArg("duration"))
					if err != nil {
						return err
					}
					conf, err := config.LoadConfig()
					if err != nil {
						return err
					}
					dbq, db, err := db.Open(ctx, conf)
					if err != nil {
						return err
					}
					defer db.Close()

					limit, err := dbq.SetDateLimitToday(ctx, int64(d.Minutes()))
					if err != nil {
						return err
					}
					fmt.Printf("new limit for today: %s\n", time.Duration(limit.LimitMinutes)*time.Minute)
					return nil
				},
			},
		}}

}
