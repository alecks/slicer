package main

import (
	"errors"
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
		Address string `yaml:"address,omitempty"`
	}
	Auth struct {
		SecretLocation string `yaml:"secret_location,omitempty"`
	}
	LogLevel string `yaml:"log_level,omitempty"`
}

var jwtSecret []byte

func readConfig() (*slicerConfig, error) {
	path := flag.String("config", slicerConfigPath, "location of the YAML config file")
	useDefaults := flag.Bool("use-defaults", true, "use the slicer_defaults.yml file")
	flag.Parse()

	conf := &slicerConfig{}
	if *useDefaults {
		if err := addDefaults(conf); err != nil {
			slog.Error("failed to read defaults file", "err", err)
		}
	}

	file, err := os.ReadFile(*path)
	if err != nil {
		slog.Error("failed to read config file", "err", err, "filepath", path)
		return nil, err
	}

	if err = yaml.Unmarshal(file, &conf); err != nil {
		slog.Error("failed to parse config file", "err", err)
		return nil, err
	}

	return conf, nil
}

func addDefaults(conf *slicerConfig) error {
	file, err := os.ReadFile(configDefaultsPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(file, conf)
}

func readJwtSecret(filepath string) (err error) {
	jwtSecret, err = os.ReadFile(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Info("SECRET BEING CREATED --- make sure to keep this file safe", "filepath", filepath)

			jwtSecret, err = generateRandomBytes(64)
			if err != nil {
				slog.Error("failed to generate random bytes for JWT secret", "err", err)
				os.Exit(1)
			}

			return os.WriteFile(filepath, jwtSecret, 0666)
		}

		slog.Error("failed to read JWT secret file", "filepath", filepath, "err", err)
		os.Exit(1)
	}
	return
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
