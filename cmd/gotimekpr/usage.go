package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/floj/gotimekpr/pkg/config"
	"github.com/floj/gotimekpr/pkg/daemon"
	"github.com/floj/gotimekpr/pkg/db"
	"github.com/urfave/cli/v3"
)

func cmdUsage() *cli.Command {
	// BAD: just copied here from daemon.go as a quick win, should be extracted to a common package or so
	getTodaysLimit := func(ctx context.Context, dbq *db.Queries) time.Duration {
		dateLimit, err := dbq.GetDateLimitToday(ctx)
		if err == nil {
			if dateLimit.LimitMinutes < 0 {
				return -1
			}
			return time.Duration(dateLimit.LimitMinutes) * time.Minute
		}
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("failed to get date limit, falling back to weekday limit", "error", err)
		}

		weekdayLimit, err := dbq.GetWeekdayLimitToday(ctx)
		if err == nil {
			if weekdayLimit.LimitMinutes < 0 {
				return -1
			}
			return time.Duration(weekdayLimit.LimitMinutes) * time.Minute
		}
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("failed to get weekday limit, using no limit", "error", err)
			return -1
		}
		slog.Debug("no weekday limit set for today, using no limit")
		return -1
	}

	checkUsage := func(ctx context.Context, dbq *db.Queries) (daemon.Usage, error) {
		limit := getTodaysLimit(ctx, dbq)

		usg := daemon.Usage{
			Exceeded:  false,
			Limit:     limit,
			Used:      0,
			Remaining: -1,
		}

		dur, err := dbq.GetDurationForToday(ctx)
		if err != nil {
			return usg, fmt.Errorf("failed to get duration for today: %w", err)
		}

		if dur.Count == 0 {
			slog.Info("no tracking records for today")
			return usg, nil
		}

		usg.Used = time.Duration(int64(dur.Total.Float64)) * time.Second
		if limit < 0 {
			return usg, nil
		}
		usg.Remaining = max(limit-usg.Used, 0)
		usg.Exceeded = usg.Remaining <= 0

		return usg, nil
	}

	return &cli.Command{
		Name:    "usage",
		Aliases: []string{"u"},
		Usage:   "shows today's usage and limit",
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

			usg, err := checkUsage(ctx, dbq)
			if err != nil {
				return err
			}

			if usg.Limit < 0 {
				fmt.Printf("Today's limit: unlimited\n")
			} else {
				fmt.Printf("Today's limit: %s\n", usg.Limit)
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
