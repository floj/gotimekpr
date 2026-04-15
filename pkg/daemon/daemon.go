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
	ctx  context.Context
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
		ctx:  ctx,
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

func (d *Daemon) track(lastRec db.Tracking) (db.Tracking, error) {
	if d.de.IsScreenLocked() {
		slog.Debug("not tracking this time")
		return db.Tracking{ID: -1}, nil
	}

	if lastRec.ID < 0 {
		rec, err := d.dbq.NewTrackingRecord(d.ctx)
		if err != nil {
			return lastRec, fmt.Errorf("error adding new tracking record: %w", err)
		}
		slog.Info("started tracking", "id", rec.ID, "created_at", rec.CreatedAt)
		return rec, nil
	}

	rec, err := d.dbq.UpdateTrackingDuration(d.ctx, lastRec.ID)
	if err != nil {
		return lastRec, fmt.Errorf("error updating tracking record: %w", err)
	}

	return rec, nil
}

func (d *Daemon) checkUsage() (usage, error) {
	limit := d.getTodaysLimit()

	usg := usage{
		exceeded:  false,
		used:      0,
		limit:     limit,
		remaining: limit,
	}

	dur, err := d.dbq.GetDurationForToday(d.ctx)
	if err != nil {
		return usg, fmt.Errorf("failed to get duration for today: %w", err)
	}

	if dur.Count == 0 {
		slog.Info("no tracking records for today")
		return usg, nil
	}

	usg.used = time.Duration(int64(dur.Total.Float64)) * time.Millisecond
	usg.remaining = max(limit-usg.used, 0)
	usg.exceeded = usg.remaining <= 0

	slog.Debug("usage today", "used", usg.used, "remaining", usg.remaining, "limit", limit)
	return usg, nil
}

func (d *Daemon) getTodaysLimit() time.Duration {
	dateLimit, err := d.dbq.GetDateLimit(d.ctx)
	if err == nil {
		return time.Duration(dateLimit.LimitMinutes) * time.Minute
	}
	if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to get date limit, falling back to weekday limit", "error", err)
	}

	weekdayLimit, err := d.dbq.GetWeekdayLimit(d.ctx)
	if err == nil {
		return time.Duration(weekdayLimit.LimitMinutes) * time.Minute
	}
	if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to get weekday limit, using no limit", "error", err)
		return -1
	}
	slog.Debug("no weekday limit set for today, using no limit")
	return -1
}

type usage struct {
	exceeded  bool
	used      time.Duration
	limit     time.Duration
	remaining time.Duration
}

func (d *Daemon) Run() error {
	ticker := time.NewTicker(d.conf.TrackingInterval)
	defer ticker.Stop()

	lastRec := db.Tracking{ID: -1}
	var err error
	for {
		select {
		case <-ticker.C:
			lastRec, err = d.track(lastRec)
			if err != nil {
				slog.Error("error tracking time", "error", err)
				continue
			}

			usg, err := d.checkUsage()
			if err != nil {
				slog.Error("error checking limit", "error", err)
				continue
			}

			if usg.exceeded {
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

			if usg.remaining < d.conf.NotifyBefore {
				slog.Info("limit approaching, sending notification", "remaining", usg.remaining)
				if err := d.de.SendNotification(fmt.Sprintf("%s remaining - you're close to your screen time limit for today.", usg.remaining)); err != nil {
					slog.Error("failed to send notification", "error", err)
				}
			}
		case <-d.ctx.Done():
			slog.Info("shutting down daemon")
			return nil
		}
	}
}
