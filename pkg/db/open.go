package db

import (
	"context"
	"database/sql"

	_ "embed"

	"github.com/floj/gotimekpr/pkg/config"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

func Open(ctx context.Context, conf config.Config) (*Queries, error) {
	db, err := sql.Open("sqlite", conf.DBURL)
	if err != nil {
		return nil, err
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return nil, err
	}

	return New(db), nil
}
