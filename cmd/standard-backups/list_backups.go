package main

import (
	"fmt"
	"log/slog"

	"github.com/dotboris/standard-backups/internal/redact"
	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/spf13/cobra"
)

var listBackupsCmd = &cobra.Command{
	Use:     "list-backups <destination>",
	Short:   "List available backups",
	Long:    `List all backups from a given destination.`,
	GroupID: "operations",
	Aliases: []string{"list", "ls", "l"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		destName := args[0]
		destination, ok := config.MainConfig.Destinations[destName]
		if !ok {
			return fmt.Errorf("could not find destination named %s", destName)
		}

		client, err := proto.NewBackendClient(*config, destination.Backend)
		if err != nil {
			return err
		}

		res, err := client.ListBackups(&proto.ListBackupsRequest{
			RawOptions:      destination.Options,
			DestinationName: destName,
		})
		if err != nil {
			return err
		}

		if len(res.Backups) == 0 {
			slog.Info("no backups found", slog.String("destination", destName))
			return nil
		}

		w := redact.Stdout
		fmt.Fprintf(w, "%+v", res)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listBackupsCmd)
}
