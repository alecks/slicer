package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const slicerVersion = "0.01"

func main() {
	config, err := readConfig()
	if err != nil {
		slog.Error("failed to read config, continuing anyway. specify config location with --config", "err", err)
	}
	slog.SetLogLoggerLevel(parseLogLevel(config.LogLevel))

	db, err := openDb(
		context.WithValue(context.Background(), ctxDbRetries, config.Db.Retries),
		config,
	)
	if err != nil {
		slog.Error("exiting. failed to open db", "err", err)
		os.Exit(1)
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	serve(config, db, sigint)
}
