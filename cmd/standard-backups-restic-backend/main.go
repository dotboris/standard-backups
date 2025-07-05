package main

import (
	"fmt"

	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/go-viper/mapstructure/v2"
)

type Forget struct {
	Enable  bool
	Options map[string]any
}

type Options struct {
	Repo   string
	Forget Forget
	Env    map[string]string
}

var Backend = &proto.BackendImpl{
	Backup: func(req *proto.BackupRequest) error {
		var options Options
		err := mapstructure.Decode(req.RawOptions, &options)
		if err != nil {
			return err
		}

		exists, err := checkRepoExists(options.Repo, options.Env)
		if err != nil {
			return err
		}
		if !exists {
			fmt.Printf("repo %s does not exist, creating it", options.Repo)
			err := restic(options.Repo, options.Env, "init")
			if err != nil {
				return fmt.Errorf("failed to initialize repository %s: %w",
					options.Repo, err)
			}
		}

		tagArgs := []string{
			"--tag", fmt.Sprintf("sb:dest:%s", req.DestinationName),
			"--tag", fmt.Sprintf("sb:job:%s", req.JobName),
		}

		backupArgs := []string{"backup"}
		backupArgs = append(backupArgs, tagArgs...)
		backupArgs = append(backupArgs, req.Paths...)
		err = restic(options.Repo, options.Env, backupArgs...)
		if err != nil {
			return fmt.Errorf("failed to backup %v to repo %s: %w",
				req.Paths, options.Repo, err)
		}

		if options.Forget.Enable {
			forgetArgs := []string{"forget"}
			forgetArgs = append(forgetArgs, tagArgs...)
			forgetOptionArgs, err := optionsToArgs(options.Forget.Options)
			if err != nil {
				return err
			}
			forgetArgs = append(forgetArgs, forgetOptionArgs...)
			err = restic(options.Repo, options.Env, forgetArgs...)
			if err != nil {
				return fmt.Errorf("failed to forget %v to repo %s: %w",
					req.Paths, options.Repo, err)
			}
		}

		return nil
	},
}

func main() {
	Backend.Execute()
}
