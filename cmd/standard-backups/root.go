package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/phsym/console-slog"
	"github.com/spf13/cobra"
)

var (
	configPath   string
	backendDirs  []string
	recipeDirs   []string
	logLevelFlag string
	logJson      bool
	noColor      bool
	lockfilePath string
	lockTimeout  time.Duration
)

func setupLogging() error {
	var level slog.Level
	switch logLevelFlag {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return fmt.Errorf("unexpected value for --log-level. Got %s expected on of debug, info, warn, error", logLevelFlag)
	}

	var handler slog.Handler
	if logJson {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	} else {
		handler = console.NewHandler(os.Stderr, &console.HandlerOptions{
			Level:   level,
			NoColor: noColor,
		})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "standard-backups",
	Short: "Backup orchestrator with pluggable backends and recipes",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := setupLogging()
		if err != nil {
			return err
		}
		return nil
	},
	SilenceUsage: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath,
		"config", "c",
		"/etc/standard-backups/config.yaml",
		"Configuration file path",
	)
	rootCmd.PersistentFlags().StringSliceVarP(&backendDirs,
		"backend-dirs", "B",
		[]string{
			"/usr/local/share/standard-backups/backends",
			"/usr/share/standard-backups/backends",
			"/etc/standard-backups/backends.d",
		},
		"Directories where to search for backends",
	)
	rootCmd.PersistentFlags().StringSliceVarP(&recipeDirs,
		"recipe-dirs", "R",
		[]string{
			"/usr/local/share/standard-backups/recipes",
			"/usr/share/standard-backups/recipes",
			"/etc/standard-backups/recipes.d",
		},
		"Directories where to search for recipes",
	)

	rootCmd.PersistentFlags().BoolVarP(&logJson,
		"log-json", "j",
		false,
		"Enable json logging",
	)
	rootCmd.PersistentFlags().StringVarP(&logLevelFlag,
		"log-level", "l",
		"info",
		"Set logging level",
	)
	rootCmd.PersistentFlags().BoolVar(&noColor,
		"no-color",
		false,
		"Disable color output",
	)

	rootCmd.PersistentFlags().StringVarP(&lockfilePath,
		"lockfile", "L",
		"/var/run/standard-backups.pid",
		"Prevents multiple instances from running at once",
	)
	rootCmd.PersistentFlags().DurationVar(&lockTimeout,
		"lock-timeout",
		5*time.Minute,
		"How long to wait to acquire lock",
	)
}
