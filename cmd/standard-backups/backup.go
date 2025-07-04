package main

import (
	"github.com/dotboris/standard-backups/internal"
	"github.com/dotboris/standard-backups/internal/config"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup <job>",
	Short: "Perform a backup for the given job",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		unlock, err := acquireLock()
		if err != nil {
			return err
		}
		defer unlock()
		jobName := args[0]
		cfg, err := config.LoadConfig(configDir)
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
