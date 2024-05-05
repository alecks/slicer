package main

import (
	"flag"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	slicerConfigPath   = "./slicer.yml"
	configDefaultsPath = "./slicer_defaults.yml"
)

type slicerConfig struct {
	Server struct {
		Address string `yaml:",omitempty"`
	}
	LogLevel string `yaml:"log_level,omitempty"`
}

func readConfig() (conf slicerConfig, err error) {
	path := flag.String("config", slicerConfigPath, "location of the YAML config file")
	useDefaults := flag.Bool("use-defaults", true, "use the slicer_defaults.yml file")
	flag.Parse()

	file, err := os.ReadFile(*path)
	if err != nil {
		return
	}

	if *useDefaults {
		addDefaults(&conf)
	}

	yaml.Unmarshal(file, &conf)
	return
}

func addDefaults(conf *slicerConfig) {
	file, err := os.ReadFile(configDefaultsPath)
	if err != nil {
		slog.Error("failed to read default config", "err", err)
		return
	}

	yaml.Unmarshal(file, conf)
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
