package config

import (
	"fmt"
	"os"
)

type ValidationError struct {
	File      string
	FieldPath string
	Err       error
}

func (c *Config) Validate() []ValidationError {
	res := []ValidationError{}

	for _, b := range c.Backends {
		err := func() *ValidationError {
			f, err := os.Stat(b.Bin)
			if err != nil {
				return &ValidationError{
					File:      b.Path,
					FieldPath: "/bin",
					Err:       err,
				}
			}
			if f.IsDir() {
				return &ValidationError{
					File:      b.Path,
					FieldPath: "/bin",
					Err:       fmt.Errorf("%s is a directory", b.Bin),
				}
			}
			return nil
		}()
		if err != nil {
			res = append(res, *err)
		}
	}

	for destName, dest := range c.MainConfig.Destinations {
		if dest.DefaultVariant == "" {
			continue
		}
		_, ok := dest.Variants[dest.DefaultVariant]
		if !ok {
			res = append(res, ValidationError{
				File:      c.MainConfig.path,
				FieldPath: fmt.Sprintf("/destinations/%s/default-variant", destName),
				Err:       fmt.Errorf("unknown variant %s for destination %s", dest.DefaultVariant, destName),
			})
		}
	}

	for jobName, job := range c.MainConfig.Jobs {
		for destIndex, destName := range job.BackupTo {
			// TODO: validate refs with variants
			// TODO: crash with no default ref
			_, ok := c.MainConfig.Destinations[destName]
			if !ok {
				res = append(res, ValidationError{
					File:      c.MainConfig.path,
					FieldPath: fmt.Sprintf("/jobs/%s/backup-to/%d", jobName, destIndex),
					Err:       fmt.Errorf("unknown destination %s", destName),
				})
			}
		}
	}

	return res
}
