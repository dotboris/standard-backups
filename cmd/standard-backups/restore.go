package main

import (
	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:     "restore destination backup-id output-dir",
	Short:   "Restore a backup from a given destination",
	GroupID: "operations",
	Args:    cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		destName := args[0]
		backupId := args[1]
		outputDir := args[2]

		config, err := loadConfig()
		if err != nil {
			return err
		}

		destination, ref, err := config.MainConfig.GetDestination(destName)
		if err != nil {
			return err
		}

		client, err := proto.NewBackendClient(*config, destination.Backend)
		if err != nil {
			return err
		}

		err = client.Restore(&proto.RestoreRequest{
			RawOptions:      destination.Options,
			DestinationName: ref.Name, // TODO: split dest name from variant
			BackupId:        backupId,
			OutputDir:       outputDir,
		})
		return err
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
