package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const slicerVersion = "0.01"

func main() {
	config, err := readConfig()
	if err != nil {
		slog.Error("failed to read config, continuing anyway", "err", err)
	}
	slog.SetLogLoggerLevel(parseLogLevel(config.LogLevel))

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	serve(config.Server.Address, sigint)
}
