package quota

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

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
	limit, err := qm.dbq.AddToDateLimitToday(ctx, int64(d.Minutes()))
	if err != nil {
		return -1, fmt.Errorf("failed to add to date limit: %w", err)
	}
	return time.Duration(limit.LimitMinutes) * time.Minute, nil
}

func (qm *QuotaManager) SetDateLimitToday(ctx context.Context, d time.Duration) (time.Duration, error) {
	limit, err := qm.dbq.SetDateLimitToday(ctx, int64(d.Minutes()))
	if err != nil {
		return -1, fmt.Errorf("failed to set date limit: %w", err)
	}
	return time.Duration(limit.LimitMinutes) * time.Minute, nil
}
