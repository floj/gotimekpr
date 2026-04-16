package quota

import (
	"context"

	"github.com/floj/gotimekpr/pkg/db"
)

func (qm *QuotaManager) TrackNew(ctx context.Context) (db.Tracking, error) {
	return qm.dbq.NewTrackingRecord(ctx)
}

func (qm *QuotaManager) TrackUpdate(ctx context.Context, rec db.Tracking) (db.Tracking, error) {
	return qm.dbq.UpdateTrackingDuration(ctx, rec.ID)
}
