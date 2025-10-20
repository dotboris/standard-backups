package testutils

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func StandardBackups(t *testing.T, args ...string) *exec.Cmd {
	cmd := exec.Command(
		"./dist/standard-backups",
		args...,
	)
	cmd.Dir = GetRepoRoot(t)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func ApplyExampleConfig(t *testing.T, cmd *exec.Cmd) {
	root := GetRepoRoot(t)
	cmd.Args = append(
		cmd.Args,
		"--config",
		fmt.Sprintf("%s/examples/config/etc/standard-backups/config.yaml", root),
	)
	if cmd.Env == nil {
		cmd.Env = append(cmd.Env, os.Environ()...)
	}
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("XDG_DATA_DIRS=%s/examples/config/share", root),
	)
}
