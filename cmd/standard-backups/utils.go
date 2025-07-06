package main

import (
	"log/slog"

	"github.com/dotboris/standard-backups/internal/config"
)

func loadConfig() (*config.Config, error) {
	slog.Debug("loading config",
		slog.String("configPath", configPath),
		slog.Any("backendDirs", backendDirs),
		slog.Any("recipeDirs", recipeDirs),
	)
	return config.LoadConfig(
		configPath,
		backendDirs,
		recipeDirs,
	)
}
