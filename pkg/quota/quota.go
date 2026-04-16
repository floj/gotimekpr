package quota

import (
	"github.com/floj/gotimekpr/pkg/db"
)

type QuotaManager struct {
	dbq *db.Queries
}

func NewQuotaManager(dbq *db.Queries) *QuotaManager {
	return &QuotaManager{
		dbq: dbq,
	}
}
