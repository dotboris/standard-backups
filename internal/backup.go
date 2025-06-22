package internal

import (
	"errors"
	"fmt"

	"github.com/dotboris/standard-backups/internal/backend"
	"github.com/dotboris/standard-backups/internal/config"
)

type backupInvocation struct {
	Backend     *backend.Backend
	BackendName string
	Destination config.DestinationConfigV1
}

func Backup(cfg config.Config, jobName string) error {
	job, ok := cfg.MainConfig.Jobs[jobName]
	if !ok {
		return fmt.Errorf("could not find a job named %s", jobName)
	}

	recipe, err := cfg.GetRecipeManifest(job.Recipe)
	if err != nil {
		return err
	}

	invocations := []backupInvocation{}
	for _, destName := range job.BackupTo {
		dest, ok := cfg.MainConfig.Destinations[destName]
		if !ok {
			return fmt.Errorf("could not find destination named %s", destName)
		}
		b, err := backend.NewBackend(cfg, dest.Backend)
		if err != nil {
			return err
		}
		invocations = append(invocations, backupInvocation{
			Backend:     b,
			BackendName: dest.Backend,
			Destination: dest,
		})
	}

	var errs error
	errCount := 0
	for _, invocation := range invocations {
		if !invocation.Backend.Enabled() {
			fmt.Printf("skipping backend %s, it's disabled", invocation.BackendName)
			continue
		}
		err = invocation.Backend.Backup(recipe.Paths, invocation.Destination.Options)
		if err != nil {
			errCount += 1
			errs = errors.Join(err)
		}
	}

	if errs != nil {
		return fmt.Errorf("backup operation failed for %d backends: %w", errCount, errs)
	}

	return nil
}
