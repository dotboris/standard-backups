package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dotboris/standard-backups/internal/config"
	"github.com/dotboris/standard-backups/pkg/proto"
)

type (
	backuper interface {
		Backup(req *proto.BackupRequest) error
	}
	backendClientFactory struct{}
	newBackendClienter   interface {
		NewBackendClient(cfg config.Config, name string) (backuper, error)
	}
	backupService struct {
		backendClientFactory newBackendClienter
	}
)

func (f *backendClientFactory) NewBackendClient(cfg config.Config, name string) (backuper, error) {
	return proto.NewBackendClient(cfg, name)
}

func NewBackupService() backupService {
	return backupService{
		backendClientFactory: &backendClientFactory{},
	}
}

func (s *backupService) Backup(cfg config.Config, jobName string) error {
	startTime := time.Now()
	job, ok := cfg.MainConfig.Jobs[jobName]
	if !ok {
		return fmt.Errorf("could not find a job named %s", jobName)
	}

	recipe, err := cfg.GetRecipeManifest(job.Recipe)
	if err != nil {
		return err
	}

	logger := slog.With(
		slog.String("job", jobName),
		slog.String("recipe", recipe.Name),
	)

	var errs error

	if recipe.Hooks.Before != nil {
		logger.Info("running before hook", slog.Any("hook", recipe.Hooks.Before))
		err := runHook(*recipe.Hooks.Before)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("before hook failed: %w", err))
		}
	}

	if errs == nil {
		for _, destName := range job.BackupTo {
			dest, ok := cfg.MainConfig.Destinations[destName]
			if !ok {
				errs = errors.Join(errs,
					fmt.Errorf("could not find destination named %s", destName))
				continue
			}
			client, err := s.backendClientFactory.NewBackendClient(cfg, dest.Backend)
			if err != nil {
				errs = errors.Join(errs,
					fmt.Errorf("failed to create backup client for destination named %s: %w", destName, err))
				continue
			}
			logger.Info("performing backup",
				slog.String("destination", destName),
				slog.String("backend", dest.Backend))
			err = client.Backup(&proto.BackupRequest{Paths: recipe.Paths, DestinationName: destName, JobName: jobName, RawOptions: dest.Options})
			if err != nil {
				errs = errors.Join(errs,
					fmt.Errorf("failed to backup destination named %s: %w", destName, err))
				continue
			}
		}
	}

	if recipe.Hooks.After != nil {
		logger.Info("running after hook", slog.Any("hook", recipe.Hooks.After))
		err := runHook(*recipe.Hooks.After)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("after hook failed: %w", err))
		}
	}

	if errs == nil {
		logger.Info("completed backup", slog.Duration("duration", time.Since(startTime)))
		if recipe.Hooks.OnSuccess != nil {
			logger.Info("running on-success hook", slog.Any("hook", recipe.Hooks.OnSuccess))
			err := runHook(*recipe.Hooks.OnSuccess)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("on-success hook failed: %w", err))
			}
		}
	}

	if errs != nil {
		logger.Error("backup failed", slog.Duration("duration", time.Since(startTime)), slog.Any("error", errs))
		if recipe.Hooks.OnFailure != nil {
			logger.Info("running on-failure hook", slog.Any("hook", recipe.Hooks.OnFailure))
			err := runHook(*recipe.Hooks.OnFailure)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("on-failure hook failed: %w", err))
			}
		}
	}

	return errs
}
