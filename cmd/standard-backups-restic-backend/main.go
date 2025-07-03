package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/dotboris/standard-backups/pkg/backend"
	"github.com/go-viper/mapstructure/v2"
)

type Options struct {
	Repo string
	Env  map[string]string
}

var Backend = &backend.Backend{
	Backup: func(paths []string, rawOptions map[string]any) error {
		var options Options
		err := mapstructure.Decode(rawOptions, &options)
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

		backupArgs := []string{"backup"}
		backupArgs = append(backupArgs, paths...)
		err = restic(options.Repo, options.Env, backupArgs...)
		if err != nil {
			return fmt.Errorf("failed to backup %v to repo %s: %w",
				paths, options.Repo, err)
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
