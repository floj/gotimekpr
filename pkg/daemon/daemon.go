package daemon

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/floj/gotimekpr/pkg/config"
	"github.com/floj/gotimekpr/pkg/db"
	"github.com/floj/gotimekpr/pkg/desktopenv"
	"github.com/godbus/dbus/v5"
)

type Daemon struct {
	conf config.Config
	conn *dbus.Conn
	db   *sql.DB
	dbq  *db.Queries
	de   *desktopenv.DesktopEnv
}

func NewDaemon(ctx context.Context, conf config.Config) (*Daemon, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, err
	}

	de, err := desktopenv.New(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	dbq, db, err := db.Open(ctx, conf)
	if err != nil {
		return nil, err
	}

	return &Daemon{
		conf: conf,
		conn: conn,
		db:   db,
		dbq:  dbq,
		de:   de,
	}, nil
}

func (d *Daemon) Close() error {
	errs := []error{}

	if err := d.db.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing database: %w", err))
	}

	if err := d.conn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("error closing dbus connection: %w", err))
	}

	return errors.Join(errs...)
}

func (d *Daemon) track(ctx context.Context, lastRec db.Tracking) (db.Tracking, error) {
	if d.de.IsScreenLocked() {
		slog.Debug("not tracking this time")
		return db.Tracking{ID: -1}, nil
	}

	if lastRec.ID < 0 {
		rec, err := d.dbq.InsertTrackingRecord(ctx)
		if err != nil {
			return lastRec, fmt.Errorf("error adding new tracking record: %w", err)
		}
		slog.Info("started tracking", "id", rec.ID, "created_at", rec.CreatedAt)
		return rec, nil
	}

	duration := time.Since(lastRec.UpdatedAt)
	rec, err := d.dbq.AddTrackingRecordDuration(ctx, db.AddTrackingRecordDurationParams{
		ID:         lastRec.ID,
		DurationMs: duration.Milliseconds(),
	})
	if err != nil {
		return lastRec, fmt.Errorf("error updating tracking record: %w", err)
	}

	return rec, nil
}

func (d *Daemon) checkLimitExceeded(ctx context.Context, limit time.Duration) (bool, time.Duration, error) {
	if limit < 0 {
		slog.Debug("no limit for today")
		return false, -1, nil
	}

	dur, err := d.dbq.GetDurationForToday(ctx)
	if err != nil {
		return false, -1, fmt.Errorf("failed to get duration for today: %w", err)
	}
	if dur.Count == 0 {
		slog.Info("no tracking records for today")
		return false, limit, nil
	}

	used := time.Duration(int64(dur.Total.Float64)) * time.Millisecond
	remaining := limit - used

	slog.Debug("duration for today", "limit", limit, "used", used, "remaining", remaining)
	return remaining <= 0, max(remaining, 0), nil
}

func (d *Daemon) Run(ctx context.Context) error {

	limits, err := d.dbq.GetLimits(ctx)
	if err != nil {
		return err
	}

	limitMap := map[int64]db.Limit{}
	for _, limit := range limits {
		limitMap[limit.Weekday] = limit
	}

	ticker := time.NewTicker(d.conf.TrackingInterval)
	defer ticker.Stop()

	weekday := time.Now().Weekday()

	var limitToday time.Duration = -1
	if limit, ok := limitMap[int64(weekday)]; ok {
		limitToday = time.Duration(limit.LimitSec) * time.Second
	}

	lastRec := db.Tracking{ID: -1}
	for {
		select {
		case <-ticker.C:
			lastRec, err = d.track(ctx, lastRec)
			if err != nil {
				slog.Error("error tracking time", "error", err)
				continue
			}

			exceeded, remaining, err := d.checkLimitExceeded(ctx, limitToday)
			if err != nil {
				slog.Error("error checking limit", "error", err)
				continue
			}

			if exceeded {
				d.de.SendNotification("You've exceeded your screen time limit for today!")
				if d.conf.NoLogout {
					slog.Info("no-logout flag is set, not logging out user")
					continue
				}
				slog.Info("logging out user")
				if err := d.de.Logout(); err != nil {
					slog.Error("failed to logout", "error", err)
					continue
				}
				continue
			}

			if remaining < d.conf.NotifyBefore {
				slog.Info("limit approaching, sending notification", "remaining", remaining)
				if err := d.de.SendNotification(fmt.Sprintf("%s remaining - you're close to your screen time limit for today.", remaining)); err != nil {
					slog.Error("failed to send notification", "error", err)
				}
			}
		case <-ctx.Done():
			slog.Info("shutting down daemon")
			return nil
		}
	}
}
