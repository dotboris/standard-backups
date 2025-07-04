package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

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

func resticCmd(repo string, env map[string]string, args ...string) *exec.Cmd {
	finalArgs := []string{"--repo", repo}
	finalArgs = append(finalArgs, args...)
	cmd := exec.Command("restic", finalArgs...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	for key, value := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	return cmd
}

func restic(repo string, env map[string]string, args ...string) error {
	cmd := resticCmd(repo, env, args...)
	fmt.Printf("running restic: %s\n", cmd.String())
	err := cmd.Run()
	return err
}

func checkRepoExists(repo string, env map[string]string) (bool, error) {
	cmd := resticCmd(repo, env, "cat", "config")
	cmd.Stderr = nil
	cmd.Stdout = nil

	err := cmd.Run()
	var exitError *exec.ExitError
	if errors.As(err, &exitError) &&
		exitError.ExitCode() == 10 /*repo does not exist*/ {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check if repo %s exists: %w", repo, err)
	}

	return true, nil
}

func optionsToArgs(options map[string]any) ([]string, error) {
	res := []string{}
	for key, value := range options {
		flag := fmt.Sprintf("--%s", key)
		if b, ok := value.(bool); ok && b {
			res = append(res, flag)
		} else if s, ok := value.(string); ok {
			res = append(res, flag, s)
		} else if i, ok := value.(int); ok {
			res = append(res, flag, fmt.Sprint(i))
		} else if f, ok := value.(float64); ok {
			res = append(res, flag, fmt.Sprint(f))
		} else {
			return nil, fmt.Errorf("could not convert option %s: %s to restic flags", key, value)
		}
	}
	return res, nil
}
