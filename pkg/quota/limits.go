package quota

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/floj/gotimekpr/pkg/db"
)

func ParseLimit(s string) (time.Duration, error) {
	if s == "unlimited" {
		return -1, nil
	}
	return time.ParseDuration(s)
}

func LimitToString(d time.Duration) string {
	if d < 0 {
		return "unlimited"
	}
	return d.String()
}

func limitToMinutes(d time.Duration) int64 {
	if d < 0 {
		return -1
	}
	return int64(d.Minutes())
}

func isValidWeekday(wd int64) bool {
	return wd >= 0 && wd <= 6
}

type WeeklyLimits []WeeklyLimit

type WeeklyLimit struct {
	Weekday  int64
	Duration time.Duration
}

func (l WeeklyLimit) WeekdayName() string {
	if !isValidWeekday(l.Weekday) {
		return fmt.Sprintf("invalid weekday index: %d", l.Weekday)
	}
	return WeekdayToString(l.Weekday)
}

func (qm *QuotaManager) GetDateLimitToday(ctx context.Context) time.Duration {
	dateLimit, err := qm.dbq.GetDateLimitToday(ctx)
	if err == nil {
		if dateLimit.LimitMinutes < 0 {
			return -1
		}
		return time.Duration(dateLimit.LimitMinutes) * time.Minute
	}
	if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to get date limit, falling back to weekday limit", "error", err)
	}

	weekdayLimit, err := qm.dbq.GetWeekdayLimitToday(ctx)
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

func (qm *QuotaManager) AddToDateLimitToday(ctx context.Context, d time.Duration) (time.Duration, error) {
	limit, err := qm.dbq.AddToDateLimitToday(ctx, limitToMinutes(d))
	if err != nil {
		return -1, fmt.Errorf("failed to add to date limit: %w", err)
	}
	return time.Duration(limit.LimitMinutes) * time.Minute, nil
}

func (qm *QuotaManager) SetDateLimitToday(ctx context.Context, d time.Duration) (time.Duration, error) {
	limit, err := qm.dbq.SetDateLimitToday(ctx, limitToMinutes(d))
	if err != nil {
		return -1, fmt.Errorf("failed to set date limit: %w", err)
	}
	return time.Duration(limit.LimitMinutes) * time.Minute, nil
}

func (qm *QuotaManager) SetWeekdayLimits(ctx context.Context, d time.Duration, weekdays ...int64) error {
	for _, wd := range weekdays {
		if wd < 0 || wd > 6 {
			return fmt.Errorf("invalid weekday: %d", wd)
		}
	}
	return qm.dbq.SetWeekdayLimits(ctx, db.SetWeekdayLimitsParams{
		LimitMinutes: limitToMinutes(d),
		Weekdays:     weekdays,
	})
}

func (qm *QuotaManager) GetWeekdayLimits(ctx context.Context) (WeeklyLimits, error) {
	ll, err := qm.dbq.GetWeekdayLimits(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekday limits: %w", err)
	}
	limits := make(WeeklyLimits, 7)
	// fill with unlimited by default
	for i := range limits {
		limits[i] = WeeklyLimit{Weekday: int64(i), Duration: -1}
	}
	for _, l := range ll {
		if !isValidWeekday(l.Weekday) {
			slog.Warn("invalid weekday index in database, skipping", "index", l.Weekday)
			continue
		}
		limits[l.Weekday] = WeeklyLimit{
			Weekday:  l.Weekday,
			Duration: time.Duration(l.LimitMinutes) * time.Minute,
		}
	}
	return limits, nil
}
