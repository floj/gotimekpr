package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/floj/gotimekpr/pkg/config"
	"github.com/floj/gotimekpr/pkg/db"
	"github.com/floj/gotimekpr/pkg/desktopenv"
	"github.com/godbus/dbus/v5"
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

			noLogout := c.Bool("no-logout")

			conf, err := config.LoadConfig()
			if err != nil {
				return err
			}

			slog.Info("starting daemon", "db", conf.DBURL)

			conn, err := dbus.ConnectSessionBus()
			if err != nil {
				return nil
			}
			defer conn.Close()

			dbq, err := db.Open(ctx, conf)
			if err != nil {
				return err
			}

			de, err := desktopenv.New()
			if err != nil {
				return err
			}
			defer de.Close()

			limits, err := dbq.GetLimits(ctx)
			if err != nil {
				return err
			}

			limitMap := map[int64]db.Limit{}
			for _, limit := range limits {
				limitMap[limit.Weekday] = limit
			}

			ticker := time.NewTicker(time.Duration(conf.IntervalSec) * time.Second)
			defer ticker.Stop()

			weekday := time.Now().Weekday()

			lastRec := db.Tracking{ID: -1}

			limitSecToday := int64(-1)
			if limit, ok := limitMap[int64(weekday)]; ok {
				limitSecToday = limit.LimitSec
			}

			for {
				select {
				case <-ticker.C:
					if de.IsScreenLocked() {
						slog.Debug("not tracking this time")
						lastRec = db.Tracking{ID: -1}
						continue
					}

					if lastRec.ID < 0 {
						rec, err := dbq.InsertTrackingRecord(ctx)
						slog.Debug("insert", "rec", rec)
						if err != nil {
							slog.Error("failed to insert tracking", "error", err)
							continue
						}
						lastRec = rec
						slog.Info("started tracking", "id", rec.ID, "created_at", rec.CreatedAt)
						continue
					}

					duration := time.Since(lastRec.UpdatedAt)
					rec, err := dbq.AddTrackingRecordDuration(ctx, db.AddTrackingRecordDurationParams{
						ID:         lastRec.ID,
						DurationMs: duration.Milliseconds(),
					})
					slog.Debug("update", "rec", rec)
					if err != nil {
						slog.Error("failed to update tracking", "error", err)
						continue
					}
					lastRec = rec
					slog.Debug("updated tracking", "id", rec.ID, "duration_ms", rec.DurationMs)

					if limitSecToday < 0 {
						slog.Debug("no limit for today")
						continue
					}

					dur, err := dbq.GetDurationForToday(ctx)
					if err != nil {
						slog.Error("failed to get duration for today", "error", err)
						continue
					}
					if dur.Count == 0 {
						slog.Info("no tracking records for today")
						continue
					}

					usedSec := int64(dur.Total.Float64 / 1000)
					remainingSec := max(limitSecToday-usedSec, 0)

					slog.Debug("duration for today", "limit", limitSecToday, "used", usedSec, "remaining", remainingSec)

					if remainingSec > conf.NotifyBeforeSec {
						slog.Debug("under limit, no notification")
						continue
					}

					slog.Info("limit exceeded, sending notification")
					if err := de.SendNotification(fmt.Sprintf("%d seconds remaining - you're close to your screen time limit for today.", remainingSec)); err != nil {
						slog.Error("failed to send notification", "error", err)
					}

					if remainingSec == 0 {
						de.SendNotification("You've exceeded your screen time limit for today!")
						if noLogout {
							slog.Info("no-logout flag is set, not logging out user")
							continue
						}
						slog.Info("logging out user")
						if err := de.Logout(); err != nil {
							slog.Error("failed to logout", "error", err)
						}
					}
				case <-ctx.Done():
					slog.Info("shutting down daemon")
					return nil
				}
			}
		},
	}
}
