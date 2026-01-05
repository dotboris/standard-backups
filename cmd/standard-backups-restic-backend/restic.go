package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func resticCmd(repo string, env map[string]string, args ...string) *exec.Cmd {
	finalArgs := []string{"--repo", repo}
	finalArgs = append(finalArgs, args...)
	cmd := exec.Command("restic", finalArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	for key, value := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	return cmd
}

func restic(repo string, env map[string]string, args ...string) error {
	cmd := resticCmd(repo, env, args...)
	fmt.Fprintf(os.Stderr, "running restic: %s\n", cmd.String())
	err := cmd.Run()
	return err
}

func resticOutput(repo string, env map[string]string, args ...string) ([]byte, error) {
	cmd := resticCmd(repo, env, args...)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	fmt.Fprintf(os.Stderr, "running restic: %s\n", cmd.String())
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return output.Bytes(), nil
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
		if b, ok := value.(bool); ok {
			if b {
				res = append(res, flag)
			}
		} else if s, ok := value.(string); ok {
			res = append(res, flag, s)
		} else if i, ok := value.(int); ok {
			res = append(res, flag, fmt.Sprint(i))
		} else if f, ok := value.(float64); ok {
			res = append(res, flag, fmt.Sprint(f))
		} else {
			return nil, fmt.Errorf("could not convert option %s: %v to restic flags", key, value)
		}
	}
	return res, nil
}
