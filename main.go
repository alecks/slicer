package main

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
)

func main() {
	config, err := readConfig()
	if err != nil {
		slog.Error("failed to read config, continuing anyway", "err", err)
	}
	slog.SetLogLoggerLevel(parseLogLevel(config.LogLevel))

	e := echo.New()

	e.Use(slogecho.New(slog.Default()))
	e.Use(middleware.Recover())

	e.Logger.Fatal(e.Start(config.Server.Address))
}
