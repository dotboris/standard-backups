package main

import (
	"log/slog"

	"github.com/dotboris/standard-backups/internal/config"
	"github.com/k0kubun/pp/v3"
	"github.com/spf13/cobra"
)

var configDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Print out the contents of the configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("loading config",
			slog.String("configPath", configPath),
			slog.Any("backendDirs", backendDirs),
			slog.Any("recipeDirs", recipeDirs),
		)
		config, err := config.LoadConfig(configPath, backendDirs, recipeDirs)
		if err != nil {
			return err
		}
		pp := pp.New()
		pp.SetColoringEnabled(!noColor)
		pp.SetExportedOnly(true)
		pp.SetOmitEmpty(false)
		pp.Println(config)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configDumpCmd)
}
