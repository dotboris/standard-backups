package main

import (
	"fmt"
	"strings"

	"github.com/dotboris/standard-backups/internal/redact"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var listRecipesCmd = &cobra.Command{
	Use:     "list-recipes",
	Short:   "List all discovered recipes",
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
		for i, recipe := range config.Recipes {
			description := "(no description)"
			if recipe.Description != "" {
				description = recipe.Description
			}

			if i > 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "%s (%s)\n",
				color.Bold.Text(recipe.Name),
				color.Cyan.Text(recipe.Path))
			fmt.Fprintf(w, "  %s: v%d\n",
				color.Magenta.Text("version"),
				recipe.Version)
			fmt.Fprintf(w, "  %s: %s\n",
				color.Magenta.Text("description"),
				description)

			fmt.Fprintf(w, "  %s: \n", color.Magenta.Text("paths"))
			for _, p := range recipe.Paths {
				fmt.Fprintf(w, "    - %s\n", p)
			}

			if len(recipe.Exclude) > 0 {
				fmt.Fprintf(w, "  %s: \n",
					color.Magenta.Text("exclude"))
				for _, e := range recipe.Exclude {
					fmt.Fprintf(w, "    - %s\n", e)
				}
			}

			if recipe.Before != nil {
				fmt.Fprintf(w, "  %s: (%s)\n",
					color.Magenta.Text("before"), recipe.Before.Shell)
				for line := range strings.Lines(recipe.Before.Command) {
					fmt.Fprintf(w, "    %s\n", strings.TrimRight(line, "\n"))
				}
			}

			if recipe.After != nil {
				fmt.Fprintf(w, "  %s: (%s)\n",
					color.Magenta.Text("after"), recipe.After.Shell)
				for line := range strings.Lines(recipe.After.Command) {
					fmt.Fprintf(w, "    %s\n", strings.TrimRight(line, "\n"))
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listRecipesCmd)
}
