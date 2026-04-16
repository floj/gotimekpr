package quota

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Usage struct {
	Exceeded  bool
	Used      time.Duration
	Limit     time.Duration
	Remaining time.Duration
}

func (qm *QuotaManager) GetUsage(ctx context.Context) (Usage, error) {
	limit := qm.GetDateLimitToday(ctx)

	usg := Usage{
		Exceeded:  false,
		Limit:     limit,
		Used:      0,
		Remaining: -1,
	}

	dur, err := qm.dbq.GetDurationForToday(ctx)
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
