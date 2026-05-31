package main

import (
	"github.com/dotboris/standard-backups/internal/config"
	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/spf13/cobra"
)

var (
	execBackend     = ""
	execDestination = ""
	execCmd         = &cobra.Command{
		Use:   "exec {-b backend | -d destination} -- [args]...",
		Short: "Run backend specific commands",
		Long: `standard-backups exec runs a backend specific command.

The specific command behavior and arguments depend on the backend. Normally,
backends will expose their underlying backup tool. This lets you run commands
that are specific to that tool. Please refer to the backend's documentation for
an explanation of available commands.

It's important to always put -- before the backend specific arguments and flags.
If you fail to do this, some arguments or flags may be interpreted by
standard-backups and not the backend.

Note that you must specify either -b/--backend or -d/--destination. This
determines which backend the command will be sent to. Using -d/--destination
enables additional behavior. The backend will be fetched from the destination
config and the destination options will be passed to the backend. This lets it
automatically configure the command to work with the specified destination.
`,
		GroupID: "operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			var dest *config.DestinationConfigV1
			var ref *config.DestinationRef
			backend := execBackend
			if backend == "" {
				dest, ref, err = cfg.MainConfig.GetDestination(execDestination)
				if err != nil {
					return err
				}
				backend = dest.Backend
			}

			client, err := proto.NewBackendClient(*cfg, backend)
			if err != nil {
				return err
			}

			req := &proto.ExecRequest{
				Args: args,
			}
			if ref != nil {
				req.DestinationName = ref.Name
				req.VariantName = ref.Variant
			}
			if dest != nil {
				req.RawOptions = dest.Options
			}
			err = client.Exec(req)
			return err
		},
		DisableFlagsInUseLine: true,
	}
)

func init() {
	execCmd.Flags().StringVarP(&execBackend,
		"backend", "b", "",
		"Exec against backend",
	)
	execCmd.Flags().StringVarP(&execDestination,
		"destination", "d", "",
		"Exec against destination",
	)
	execCmd.MarkFlagsMutuallyExclusive("backend", "destination")
	execCmd.MarkFlagsOneRequired("backend", "destination")
	rootCmd.AddCommand(execCmd)
}
