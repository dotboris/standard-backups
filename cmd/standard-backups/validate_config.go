package main

import (
	"fmt"
	"strings"

	"github.com/dotboris/standard-backups/internal/config"
	"github.com/spf13/cobra"
)

var validateConfigCmd = &cobra.Command{
	Use:   "validate-config",
	Short: "Validates that the configurations are correct",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		errs := c.Validate()
		if len(errs) > 0 {
			byFile := map[string][]config.ValidationError{}
			for _, err := range errs {
				byFile[err.File] = append(byFile[err.File], err)
			}
			message := strings.Builder{}
			for file, errs := range byFile {
				message.WriteString(fmt.Sprintf("%s:\n", file))
				for _, err := range errs {
					message.WriteString(fmt.Sprintf("- at '%s': %s\n", err.FieldPath, err.Err))
				}
			}
			return fmt.Errorf("configuration is not valid:\n%s", message.String())
		}

		fmt.Println("configuration is valid")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateConfigCmd)
}
