package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dotboris/standard-backups/internal/backend"
	"github.com/dotboris/standard-backups/internal/config"
)

type backupInvocation struct {
	Backend         *backend.Backend
	BackendName     string
	Destination     config.DestinationConfigV1
	DestinationName string
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
			Backend:         b,
			BackendName:     dest.Backend,
			Destination:     dest,
			DestinationName: destName,
		})
	}

	var errs error
	errCount := 0
	for _, invocation := range invocations {
		logger := slog.With(
			slog.String("job", jobName),
			slog.String("recipe", job.Recipe),
			slog.String("destination", invocation.DestinationName),
			slog.String("backend", invocation.BackendName),
		)
		if !invocation.Backend.Enabled() {
			logger.Warn("skipping backup, backend is disabled")
			continue
		}
		logger.Info("starting backup")
		startTime := time.Now()
		err = invocation.Backend.Backup(recipe.Paths, invocation.Destination.Options)
		logger.Info("completed backup", slog.Duration("duration", time.Since(startTime)))
		if err != nil {
			logger.Error("backup failed",
				slog.Duration("duration", time.Since(startTime)),
				slog.Any("error", err),
			)
			errCount += 1
			errs = errors.Join(err)
		}
	}

	if errs != nil {
		return fmt.Errorf("backup operation failed for %d backends: %w", errCount, errs)
	}

	return nil
}
