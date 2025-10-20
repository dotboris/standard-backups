package main

import (
	"log/slog"

	"github.com/dotboris/standard-backups/internal/config"
	"github.com/spf13/cobra"
)

var (
	configPath  string
	backendDirs []string
	recipeDirs  []string
)

func addConfigFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&configPath,
		"config", "c",
		"/etc/standard-backups/config.yaml",
		"Configuration file path",
	)
	cmd.PersistentFlags().StringSliceVarP(&backendDirs,
		"backend-dirs", "B",
		[]string{
			"/usr/local/share/standard-backups/backends",
			"/usr/share/standard-backups/backends",
			"/etc/standard-backups/backends.d",
		},
		"Directories where to search for backends",
	)
	cmd.PersistentFlags().StringSliceVarP(&recipeDirs,
		"recipe-dirs", "R",
		[]string{
			"/usr/local/share/standard-backups/recipes",
			"/usr/share/standard-backups/recipes",
			"/etc/standard-backups/recipes.d",
		},
		"Directories where to search for recipes",
	)
}

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
