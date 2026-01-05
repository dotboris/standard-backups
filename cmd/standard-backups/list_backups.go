package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dotboris/standard-backups/internal/redact"
	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

var (
	listBackupsJson    bool
	listBackupsColumns []string
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
						AutoWrap:   tw.WrapNormal,
						AutoFormat: tw.Off,
					},
				},
			}),
		)
		table.Header(listBackupsColumns)
		for _, backup := range res.Backups {
			row := make([]string, len(listBackupsColumns))
			for i, col := range listBackupsColumns {
				row[i] = formatColumn(col, backup)
			}
			err = table.Append(row)
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
	listBackupsCmd.Flags().StringSliceVarP(&listBackupsColumns,
		"columns", "C",
		[]string{"id", "time", "job", "destination", "size"},
		"Columns to include output",
	)
	listBackupsCmd.MarkFlagsMutuallyExclusive("json", "columns")

	rootCmd.AddCommand(listBackupsCmd)
}

func formatColumn(col string, backup proto.ListBackupsResponseItem) string {
	switch strings.ToLower(col) {
	case "id":
		return backup.Id
	case "time":
		return backup.Time
	case "job":
		return backup.Job
	case "destination":
		return backup.Destination
	case "size":
		unit := "B"
		size := float64(backup.Size)
		if size >= 1024 {
			size = size / 1024
			unit = "KB"
		}
		if size >= 1024 {
			size = size / 1024
			unit = "MB"
		}
		if size >= 1024 {
			size = size / 1024
			unit = "GB"
		}
		if size >= 1024 {
			size = size / 1024
			unit = "TB"
		}
		if size >= 1024 {
			size = size / 1024
			unit = "PB"
		}
		formatted := fmt.Sprintf("%.2f", size)
		formatted = strings.TrimRight(formatted, "0")
		formatted = strings.TrimRight(formatted, ".")
		return fmt.Sprintf("%s %s", formatted, unit)
	default:
		if col, ok := strings.CutPrefix(col, "extra."); ok {
			parts := strings.Split(col, ".")
			var value any = backup.Extra
			for _, part := range parts {
				if m, ok := value.(map[string]any); ok {
					value, ok = m[part]
					if !ok {
						break
					}
				}
			}
			if value == nil {
				return ""
			}
			return fmt.Sprintf("%v", value)
		}
		return ""
	}
}
