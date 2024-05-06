package main

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/alecks/slicer/migrations"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
)

type dbConfig struct {
	DSN         string `yaml:"dsn,omitempty"`
	PingTimeout int    `yaml:"ping_timeout,omitempty"`
	Retries     int    `yaml:"retries,omitempty"`
}

func openDb(ctx context.Context, conf *slicerConfig) (*bun.DB, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*time.Duration(conf.Db.PingTimeout))
	defer cancel()

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(conf.Db.DSN)))
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

	if conf.flags.doMigrate && conf.flags.doRollback {
		slog.Error("can't rollback and migrate! run these individually")
		return db, nil
	} else if conf.flags.doMigrate {
		if err := migrateDb(db); err != nil {
			slog.Error("failed to migrate db, continuing anyway", "err", err)
			return db, nil
		}
	} else if conf.flags.doRollback {
		if err := rollbackDb(db); err != nil {
			slog.Error("failed to rollback db", "err", err)
		}
	}
	return db, nil
}

func migrateDb(db *bun.DB) error {
	m := migrate.NewMigrator(db, migrations.Migrations)
	if err := m.Init(context.Background()); err != nil {
		slog.Error("failed to initialise migrations table", "err", err)
		return err
	}

	if err := m.Lock(context.Background()); err != nil {
		return err
	}
	defer m.Unlock(context.Background())

	group, err := m.Migrate(context.Background())
	if err != nil {
		return err
	}
	if group.IsZero() {
		slog.Info("no migrations to run, db is up to date")
	} else {
		slog.Info("migrations complete", "group", group)
	}
	return nil
}

func rollbackDb(db *bun.DB) error {
	m := migrate.NewMigrator(db, migrations.Migrations)

	if err := m.Lock(context.Background()); err != nil {
		return err
	}
	defer m.Unlock(context.Background())

	group, err := m.Rollback(context.Background())
	if err != nil {
		return err
	}
	slog.Info("rolled back migrations", "group", group)
	return nil
}
