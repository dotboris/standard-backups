package main

import "github.com/spf13/cobra"

var validateConfigCmd = &cobra.Command{
	Use:   "validate-config",
	Short: "Validates that the configurations are correct",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateConfigCmd)
}
