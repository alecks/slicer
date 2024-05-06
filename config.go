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
	slicerDefaultsPath = "./slicer_defaults.yml"
)

type ctxKeys int

const (
	ctxClaims ctxKeys = iota
	ctxDbRetries
	ctxDb
)

type slicerConfig struct {
	Server struct {
		Address string `yaml:"address,omitempty"`
	}
	Auth struct {
		SecretLocation string `yaml:"secret_location,omitempty"`
	}
	LogLevel string `yaml:"log_level,omitempty"`
	flags    slicerFlags
	Db       dbConfig
}

type slicerFlags struct {
	configPath       string
	useDefaultConfig bool
	doMigrate        bool
	doRollback       bool
}

var jwtSecret []byte

func readConfig() (*slicerConfig, error) {
	configPath := flag.String("config", slicerConfigPath, "location of the YAML config file")
	useDefaults := flag.Bool("use-defaults", true, "use the slicer_defaults.yml file")
	doMigrate := flag.Bool("migrate", false, "run migrations on startup")
	doRollback := flag.Bool("rollback-force", false, "rollback migrations on startup (dangerous)")
	flag.Parse()

	conf := &slicerConfig{}
	conf.flags = slicerFlags{
		configPath:       *configPath,
		useDefaultConfig: *useDefaults,
		doMigrate:        *doMigrate,
		doRollback:       *doRollback,
	}

	if *useDefaults {
		if err := addDefaults(conf); err != nil {
			slog.Error("failed to read defaults file", "err", err)
		}
	}

	file, err := os.ReadFile(*configPath)
	if err != nil {
		slog.Error("failed to read config file", "err", err, "filepath", configPath)
		return nil, err
	}

	if err = yaml.Unmarshal(file, &conf); err != nil {
		slog.Error("failed to parse config file", "err", err)
		return nil, err
	}

	jwtSecret, err = readJwtSecret(conf.Auth.SecretLocation)
	if err != nil {
		slog.Error("failed to create/read jwt secret, exiting", "err", err)
		os.Exit(1)
	}
	return conf, nil
}

func addDefaults(conf *slicerConfig) error {
	file, err := os.ReadFile(slicerDefaultsPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(file, conf)
}

func readJwtSecret(filepath string) ([]byte, error) {
	jwtSecret, err := os.ReadFile(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Info("SECRET BEING CREATED --- make sure to keep this file safe", "filepath", filepath)

			jwtSecret, err = generateRandomBytes(64)
			if err != nil {
				slog.Error("failed to generate random bytes for JWT secret", "err", err)
				return nil, err
			}

			return jwtSecret, os.WriteFile(filepath, jwtSecret, 0666)
		}

		slog.Error("failed to read JWT secret file", "filepath", filepath, "err", err)
		return nil, err
	}
	return jwtSecret, nil
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
