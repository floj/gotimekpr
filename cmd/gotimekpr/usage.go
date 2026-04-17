package main

import (
	"context"
	"fmt"

	"github.com/floj/gotimekpr/pkg/quota"
	"github.com/urfave/cli/v3"
)

func cmdUsage() *cli.Command {
	return &cli.Command{
		Name:    "usage",
		Aliases: []string{"u"},
		Usage:   "shows today's usage and limit",
		Action: withQuotaManager(func(ctx context.Context, c *cli.Command, qm *quota.QuotaManager) error {
			usg, err := qm.GetUsage(ctx)
			if err != nil {
				return err
			}
			remaining := "N/A"
			if usg.Remaining >= 0 {
				remaining = usg.Remaining.String()
			}

			exceeded := "no"
			if usg.Exceeded {
				exceeded = "yes"
			}

			fmt.Printf("Limit     | %s\n", quota.LimitToString(usg.Limit))
			fmt.Printf("Used      | %s\n", usg.Used)
			fmt.Printf("Remaining | %s\n", remaining)
			fmt.Printf("Exceeded  | %s\n", exceeded)
			return nil
		}),
	}
}
