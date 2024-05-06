package main

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type dbConfig struct {
	DSN         string `yaml:"dsn,omitempty"`
	PingTimeout int    `yaml:"ping_timeout,omitempty"`
	Retries     int    `yaml:"retries,omitempty"`
}

func openDb(ctx context.Context, conf *dbConfig) (*bun.DB, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*time.Duration(conf.PingTimeout))
	defer cancel()

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(conf.DSN)))
	db := bun.NewDB(sqldb, pgdialect.New())

	if err := db.PingContext(ctx); err != nil {
		slog.Error("failed to ping db", "err", err)
		db.Close()

		retries := ctx.Value(ctxDbRetries).(int)
		if retries == 0 {
			slog.Error("retry limit reached")
			return nil, err
		}
		slog.Info("retrying...", "db_retries_remaining", retries)

		return openDb(context.WithValue(context.Background(), ctxDbRetries, retries-1), conf)
	}

	slog.Info("connected to database")
	return db, nil
}
