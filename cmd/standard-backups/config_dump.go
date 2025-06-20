package main

import (
	"fmt"

	"github.com/dotboris/standard-backups/internal/config"
	"github.com/k0kubun/pp/v3"
	"github.com/spf13/cobra"
)

var configDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Print out the contents of the configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Using config in %s\n", ConfigDir)
		config, err := config.LoadConfig(ConfigDir)
		if err != nil {
			return err
		}
		pp := pp.New()
		pp.SetColoringEnabled(true)
		pp.SetExportedOnly(true)
		pp.SetOmitEmpty(false)
		pp.Println(config)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configDumpCmd)
}
