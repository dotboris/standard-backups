package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dotboris/standard-backups/internal/backend"
	"github.com/dotboris/standard-backups/internal/config"
)

func Backup(cfg config.Config, jobName string) error {
	job, ok := cfg.MainConfig.Jobs[jobName]
	if !ok {
		return fmt.Errorf("could not find a job named %s", jobName)
	}

	recipe, err := cfg.GetRecipeManifest(job.Recipe)
	if err != nil {
		return err
	}

	var errs error
	errCount := 0
	for _, destName := range job.BackupTo {
		logger := slog.With(
			slog.String("recipe", recipe.Name),
			slog.String("destination", destName),
		)
		err := func() error {
			dest, ok := cfg.MainConfig.Destinations[destName]
			if !ok {
				return fmt.Errorf("could not find destination named %s", destName)
			}
			b, err := backend.NewBackend(cfg, dest.Backend)
			if err != nil {
				return err
			}
			return backupSingle(logger, recipe, dest, b)
		}()
		if err != nil {
			errCount += 1
			errs = errors.Join(err)
		}
	}

	if errs != nil {
		return fmt.Errorf("%d/%d backup operation failed: %w",
			errCount, len(job.BackupTo), errs)
	}

	return nil
}

type backupBackend interface {
	Enabled() bool
	Backup([]string, map[string]any) error
}

func backupSingle(
	logger *slog.Logger,
	recipe *config.RecipeManifestV1,
	destination config.DestinationConfigV1,
	b backupBackend,
) error {
	if !b.Enabled() {
		logger.Warn("skipping backup, backend is disabled")
		return nil
	}
	logger.Info("starting backup")
	startTime := time.Now()
	err := b.Backup(recipe.Paths, destination.Options)
	if err != nil {
		logger.Error("backup failed",
			slog.Duration("duration", time.Since(startTime)),
			slog.Any("error", err),
		)
		return err
	}
	logger.Info("completed backup",
		slog.Duration("duration", time.Since(startTime)))
	return nil
}
