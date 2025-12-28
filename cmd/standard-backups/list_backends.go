package main

import (
	"fmt"

	"github.com/dotboris/standard-backups/internal/redact"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var listBackendsCmd = &cobra.Command{
	Use:     "list-backends",
	Short:   "List all discovered backends",
	GroupID: "config",
	RunE: func(cmd *cobra.Command, args []string) error {
		if noColor {
			color.Disable()
		}

		config, err := loadConfig()
		if err != nil {
			return err
		}

		w := redact.Stdout
		for i, backend := range config.Backends {
			description := "(no description)"
			if backend.Description != "" {
				description = backend.Description
			}

			if i > 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "%s (%s)\n",
				color.Bold.Text(backend.Name),
				color.Cyan.Text(backend.Path))
			fmt.Fprintf(w, "  %s: %s\n",
				color.Magenta.Text("description"),
				description)
			fmt.Fprintf(w, "  %s: %s\n",
				color.Magenta.Text("bin"),
				backend.Bin)
			fmt.Fprintf(w, "  %s: %d\n",
				color.Magenta.Text("version"),
				backend.Version)
			fmt.Fprintf(w, "  %s: %d\n",
				color.Magenta.Text("protocol version"),
				backend.ProtocolVersion)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listBackendsCmd)
}
