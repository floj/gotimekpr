package main

import (
	"context"
	"fmt"

	"github.com/floj/gotimekpr/pkg/config"
	"github.com/floj/gotimekpr/pkg/db"
	"github.com/floj/gotimekpr/pkg/quota"
	"github.com/urfave/cli/v3"
)

func cmdUsage() *cli.Command {
	return &cli.Command{
		Name:    "usage",
		Aliases: []string{"u"},
		Usage:   "shows today's usage and limit",
		Action: func(ctx context.Context, c *cli.Command) error {
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

			usg, err := qm.GetUsage(ctx)
			if err != nil {
				return err
			}

			if usg.Limit < 0 {
				fmt.Printf("Limit: unlimited\n")
			} else {
				fmt.Printf("Limit: %s\n", usg.Limit)
			}
			fmt.Printf("Used: %s\n", usg.Used)
			if usg.Remaining >= 0 {
				fmt.Printf("Remaining: %s\n", usg.Remaining)
			} else {
				fmt.Printf("Remaining: N/A\n")
			}
			if usg.Exceeded {
				fmt.Printf("Limit exceeded!\n")
			}

			return nil
		},
	}
}
