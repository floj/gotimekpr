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

			if usg.Limit < 0 {
				fmt.Printf("Limit     | unlimited\n")
			} else {
				fmt.Printf("Limit     | %s\n", usg.Limit)
			}

			fmt.Printf("Used      | %s\n", usg.Used)

			if usg.Remaining >= 0 {
				fmt.Printf("Remaining | %s\n", usg.Remaining)
			} else {
				fmt.Printf("Remaining | N/A\n")
			}

			if usg.Exceeded {
				fmt.Printf("Exceeded  | yes\n")
			} else {
				fmt.Printf("Exceeded  | no\n")
			}

			return nil
		}),
	}
}
