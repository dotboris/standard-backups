package internal

import (
	"errors"
	"os"
	"os/exec"

	"github.com/dotboris/standard-backups/internal/config"
)

var errUnsupportedShell = errors.New("unsupported shell")

func runHook(hook config.HookV1) error {
	var (
		command string
		args    []string
	)
	switch hook.Shell {
	case "sh":
		command = "sh"
		args = []string{"-c", hook.Command}
	case "bash":
		command = "bash"
		args = []string{"-c", hook.Command}
	default:
		return errUnsupportedShell
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
