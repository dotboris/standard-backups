package main

import (
	"errors"

	"github.com/dotboris/standard-backups/internal"
	"github.com/dotboris/standard-backups/internal/config"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup <job>",
	Short: "Perform a backup for the given job",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires at least one argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		jobName := args[0]
		cfg, err := config.LoadConfig(ConfigDir)
		if err != nil {
			return err
		}
		err = internal.Backup(*cfg, jobName)
		return err
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
