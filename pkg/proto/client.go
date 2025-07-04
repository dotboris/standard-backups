package proto

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dotboris/standard-backups/internal/config"
)

type BackendClient struct {
	Manifest config.BackendManifestV1
	Config   config.BackendConfigV1
}

func NewBackendClient(cfg config.Config, name string) (*BackendClient, error) {
	manifest, err := cfg.GetBackendManifest(name)
	if err != nil {
		return nil, err
	}

	backendCfg, ok := cfg.MainConfig.Backends[name]
	if !ok {
		return nil, fmt.Errorf("could not find a configuration for backend %s", name)
	}

	return &BackendClient{
		Manifest: *manifest,
		Config:   backendCfg,
	}, nil
}

func (bc *BackendClient) Enabled() bool {
	return bc.Config.Enable
}

func (bc *BackendClient) cmd(command string, env []string) *exec.Cmd {
	cmd := exec.Command(bc.Manifest.Bin)
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("STANDARD_BACKUPS_COMMAND=%s", command),
	)
	cmd.Env = append(cmd.Env, env...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd
}
