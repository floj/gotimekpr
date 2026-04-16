package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/floj/gotimekpr/pkg/config"
	"github.com/floj/gotimekpr/pkg/db"
	"github.com/floj/gotimekpr/pkg/desktopenv"
	"github.com/floj/gotimekpr/pkg/quota"
	"github.com/godbus/dbus/v5"
)

type Daemon struct {
	conf config.Config
	ctx  context.Context
	conn *dbus.Conn
	de   desktopenv.DesktopEnv
	qm   *quota.QuotaManager
}

func NewDaemon(ctx context.Context, conf config.Config, qm *quota.QuotaManager) (*Daemon, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, err
	}

	de, err := desktopenv.New(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Daemon{
		conf: conf,
		ctx:  ctx,
		conn: conn,
		de:   de,
		qm:   qm,
	}, nil
}

func (d *Daemon) Close() error {
	errs := []error{}

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
		rec, err := d.qm.TrackNew(d.ctx)
		if err != nil {
			return lastRec, fmt.Errorf("error adding new tracking record: %w", err)
		}
		slog.Info("started tracking", "id", rec.ID, "created_at", rec.CreatedAt)
		return rec, nil
	}

	rec, err := d.qm.TrackUpdate(d.ctx, lastRec)
	if err != nil {
		return lastRec, fmt.Errorf("error updating tracking record: %w", err)
	}

	return rec, nil
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

			usg, err := d.qm.GetUsage(d.ctx)
			if err != nil {
				slog.Error("error checking limit", "error", err)
				continue
			}
			slog.Debug("current usage", "used", usg.Used, "remaining", usg.Remaining, "limit", usg.Limit)

			if usg.Limit < 0 {
				continue
			}

			if usg.Exceeded {
				d.de.SendNotification("Logout", "You've exceeded your screen time limit for today!")
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

			if usg.Remaining < d.conf.NotifyBefore {
				slog.Info("limit approaching, sending notification", "remaining", usg.Remaining)
				if err := d.de.SendNotification("Screentime Alert", fmt.Sprintf("%s remaining - you're close to your screen time limit for today.", usg.Remaining)); err != nil {
					slog.Error("failed to send notification", "error", err)
				}
			}
		case <-d.ctx.Done():
			if lastRec.ID > 0 {
				_, err := d.qm.TrackUpdate(context.Background(), lastRec) // final update
				if err != nil {
					slog.Error("failed to update final tracking record", "error", err)
				}
			}
			slog.Info("shutting down daemon")
			return nil
		}
	}
}
