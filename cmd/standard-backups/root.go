package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/phsym/console-slog"
	"github.com/spf13/cobra"
)

var (
	logLevelFlag string
	logJson      bool
	noColor      bool
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
		return fmt.Errorf(
			"unexpected value for --log-level. Got %s expected on of debug, info, warn, error",
			logLevelFlag,
		)
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
	rootCmd.AddGroup(
		&cobra.Group{ID: "operations", Title: "Operation Commands"},
		&cobra.Group{ID: "config", Title: "Configuration Commands"},
	)

	addConfigFlags(rootCmd)

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
}
