package main

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Commands for dealing with the configuration",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
