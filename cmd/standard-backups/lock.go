package main

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nightlyone/lockfile"
)

func acquireLock() (func(), error) {
	logger := slog.With(
		slog.String("lockfilePath", lockfilePath),
		slog.Duration("lockTimeout", lockTimeout),
	)

	lock, err := lockfile.New(lockfilePath)
	if err != nil {
		return nil, err
	}

	logger.Debug("attempting to acquire lock")
	startTime := time.Now()
	for {
		err := lock.TryLock()
		if err == nil {
			logger.Debug("acquired lock")
			return func() {
				err := lock.Unlock()
				if err == nil {
					logger.Debug("released lock")
				} else {
					logger.Warn("failed to release lock",
						slog.Any("error", err))
				}
			}, nil
		} else if errors.Is(err, lockfile.ErrBusy) || errors.Is(err, lockfile.ErrNotExist) {
			if time.Since(startTime) >= lockTimeout {
				return nil, fmt.Errorf("failed to acquire lock %s after %d", lockfilePath, lockTimeout)
			}
			logger.Debug("waiting for other instance to finish")
			time.Sleep(1 * time.Second)
		} else {
			return nil, err
		}
	}
}
