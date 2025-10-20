package main

import (
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/dotboris/standard-backups/internal/config"
	"github.com/spf13/cobra"
)

var configPath string

func addConfigFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&configPath,
		"config", "c",
		"/etc/standard-backups/config.yaml",
		"Configuration file path",
	)
}

func loadConfig() (*config.Config, error) {
	home, _ := os.UserHomeDir()
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" && home != "" {
		xdgConfigHome = path.Join(home, ".config")
	}
	xdgConfigDirs := os.Getenv("XDG_CONFIG_DIRS")
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" && home != "" {
		xdgDataHome = path.Join(home, ".local/share")
	}
	xdgDataDirs := os.Getenv("XDG_DATA_DIRS")

	dirs := []string{}
	dirs = append(dirs, xdgConfigHome)
	dirs = append(dirs, "/etc")
	dirs = append(dirs, strings.Split(xdgConfigDirs, ":")...)
	dirs = append(dirs, xdgDataHome)
	dirs = append(dirs, strings.Split(xdgDataDirs, ":")...)

	backendDirs := []string{}
	recipeDirs := []string{}
	for _, dir := range dirs {
		if dir != "" && path.IsAbs(dir) {
			backendDir := path.Join(dir, "standard-backups/backends")
			backendDirStat, err := os.Stat(backendDir)
			if err == nil && backendDirStat.IsDir() {
				backendDirs = append(backendDirs, backendDir)
			}

			recipeDir := path.Join(dir, "standard-backups/recipes")
			recipeDirStat, err := os.Stat(recipeDir)
			if err == nil && recipeDirStat.IsDir() {
				recipeDirs = append(recipeDirs, recipeDir)
			}
		}
	}

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
