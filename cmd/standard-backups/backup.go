package main

import (
	"github.com/dotboris/standard-backups/internal"
	"github.com/spf13/cobra"
)

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
		cfg, err := loadConfig()
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
