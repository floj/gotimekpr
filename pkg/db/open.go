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

func Open(ctx context.Context, conf config.Config) (*Queries, *sql.DB, error) {
	db, err := sql.Open("sqlite", conf.DBURL)
	if err != nil {
		return nil, nil, err
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		db.Close()
		return nil, nil, err
	}

	return New(db), db, nil
}
