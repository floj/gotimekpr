package main

import (
	"context"
	"fmt"

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

func printWeekdayLimits(ctx context.Context, qm *quota.QuotaManager) error {
	orderStartMonday := []int64{1, 2, 3, 4, 5, 6, 0}
	limit, err := qm.GetWeekdayLimits(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Weekly limits\n")
	fmt.Printf("----------------------\n")

	for _, idx := range orderStartMonday {
		l := limit[idx]
		fmt.Printf("%-10s | %s\n", l.WeekdayName(), quota.LimitToString(l.Duration))
	}
	return nil
}

func subcommandsWeek() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "set",
			Usage:     "Set the limits of one or more weekdays",
			ArgsUsage: "<duration|'unlimited'> <mon|tue|...|all|weekend|workdays> [<mon,tue>, ...]",
			Arguments: []cli.Argument{
				&cli.StringArg{
					Name: "duration",
					Config: cli.StringConfig{
						TrimSpace: true,
					},
				},
			},
			Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
				if c.Args().Len() < 2 {
					return ctx, fmt.Errorf("duration and at least one weekday must be provided")
				}
				return ctx, nil
			},
			Action: withQuotaManager(func(ctx context.Context, c *cli.Command, qm *quota.QuotaManager) error {
				d, err := quota.ParseLimit(c.StringArg("duration"))
				if err != nil {
					return err
				}
				days := c.Args().Slice()
				weekdays, err := quota.WeekdaysFromStrings(days)
				if err != nil {
					return err
				}

				if err := qm.SetWeekdayLimits(ctx, d, weekdays...); err != nil {
					return err
				}

				return printWeekdayLimits(ctx, qm)
			}),
		},
	}
}

func subcommandsToday() []*cli.Command {
	return []*cli.Command{
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
				d, err := quota.ParseLimit(c.StringArg("duration"))
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
				d, err := quota.ParseLimit(c.StringArg("duration"))
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
	}
}

func cmdLimits() *cli.Command {
	return &cli.Command{
		Name:    "limits",
		Aliases: []string{"l"},
		Usage:   "manage time limits",
		Commands: []*cli.Command{
			{
				Name:  "today",
				Usage: "show and manage today's limit",
				Action: withQuotaManager(func(ctx context.Context, c *cli.Command, qm *quota.QuotaManager) error {
					limit := qm.GetDateLimitToday(ctx)
					fmt.Printf("limit for today: %s\n", limit)
					return nil
				}),
				Commands: subcommandsToday(),
			},
			{
				Name:  "week",
				Usage: "show and manage the limits per weekday",
				Action: withQuotaManager(func(ctx context.Context, c *cli.Command, qm *quota.QuotaManager) error {
					return printWeekdayLimits(ctx, qm)
				}),
				Commands: subcommandsWeek(),
			},
		},
	}

}
