package proto

import (
	"os"
	"os/exec"

	"github.com/dotboris/standard-backups/internal/config"
	"github.com/dotboris/standard-backups/internal/redact"
)

type BackendClient struct {
	Manifest config.BackendManifestV1
}

func NewBackendClient(cfg config.Config, name string) (*BackendClient, error) {
	manifest, err := cfg.GetBackendManifest(name)
	if err != nil {
		return nil, err
	}

	return &BackendClient{
		Manifest: *manifest,
	}, nil
}

func (bc *BackendClient) cmd(command string, env []string) *exec.Cmd {
	cmd := exec.Command(bc.Manifest.Bin)
	cmd.Env = append(
		os.Environ(),
		toEnvStr(COMMAND_ENV, command),
	)
	cmd.Env = append(cmd.Env, env...)
	cmd.Stdout = redact.Stdout
	cmd.Stderr = redact.Stderr
	return cmd
}
