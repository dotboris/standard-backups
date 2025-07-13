package main

import (
	"github.com/k0kubun/pp/v3"
	"github.com/spf13/cobra"
)

var printConfigCmd = &cobra.Command{
	Use:   "print-config",
	Short: "Print out the contents of the configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}
		pp := pp.New()
		pp.SetColoringEnabled(!noColor)
		pp.SetExportedOnly(true)
		pp.SetOmitEmpty(false)
		_, err = pp.Println(config)
		return err
	},
}

func init() {
	rootCmd.AddCommand(printConfigCmd)
}
