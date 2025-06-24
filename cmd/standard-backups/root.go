package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/phsym/console-slog"
	"github.com/spf13/cobra"
)

var (
	configDir    string
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "standard-backups",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := setupLogging()
		if err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configDir,
		"config-dir", "c",
		"",
		"Directory where the configuration is stored",
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
	configDumpCmd.PersistentFlags().BoolVar(&noColor,
		"no-color",
		false,
		"Disable color output",
	)
}
