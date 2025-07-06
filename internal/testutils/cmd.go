package testutils

import (
	"os"
	"os/exec"
	"testing"
)

var ExampleConfigArgs = []string{
	"--config", "examples/config/etc/standard-backups/config.yaml",
	"--backend-dirs", "examples/config/etc/standard-backups/backends.d",
	"--recipe-dirs", "examples/config/etc/standard-backups/recipes.d",
}

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
