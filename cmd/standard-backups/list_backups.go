package main

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dotboris/standard-backups/internal/redact"
	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

var listBackupsJson bool

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

		if listBackupsJson {
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			err = enc.Encode(res.Backups)
			if err != nil {
				return err
			}
			return nil
		}

		table := tablewriter.NewTable(w,
			tablewriter.WithRendition(tw.Rendition{
				Borders: tw.BorderNone,
			}),
			tablewriter.WithConfig(tablewriter.Config{
				Header: tw.CellConfig{
					Formatting: tw.CellFormatting{
						// Destination gets truncated sometimes, we want to keep it in full
						AutoWrap: tw.WrapNormal,
					},
				},
			}),
		)
		table.Header("Id", "Time", "Job", "Destination", "Size")
		for _, backup := range res.Backups {
			err = table.Append([]string{
				backup.Id,
				backup.Time,
				backup.Job,
				backup.Destination,
				fmt.Sprintf("%d B", backup.Bytes),
			})
			if err != nil {
				return err
			}
		}

		fmt.Fprintln(w)
		err = table.Render()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	listBackupsCmd.Flags().BoolVar(&listBackupsJson,
		"json", false,
		"Print backups to stdout as JSON",
	)

	rootCmd.AddCommand(listBackupsCmd)
}
