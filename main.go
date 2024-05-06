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

	_, err = openDb(context.WithValue(context.Background(), ctxDbRetries, config.Db.Retries), &config.Db)
	if err != nil {
		slog.Error("exiting. failed to open db", "err", err)
		os.Exit(1)
	}

	if err := readJwtSecret(config.Auth.SecretLocation); err != nil {
		slog.Error("failed to read jwt secret; check config/auth.secret_location", "err", err)
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	serve(config.Server.Address, sigint)
}
