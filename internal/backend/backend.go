package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dotboris/standard-backups/internal/config"
)

type Backend struct {
	Manifest config.BackendManifestV1
	Config   config.BackendConfigV1
}

func NewBackend(cfg config.Config, name string) (*Backend, error) {
	manifest, err := cfg.GetBackendManifest(name)
	if err != nil {
		return nil, err
	}

	backendCfg, ok := cfg.MainConfig.Backends[name]
	if !ok {
		return nil, fmt.Errorf("could not find a configuration for backend %s", name)
	}

	return &Backend{
		Manifest: *manifest,
		Config:   backendCfg,
	}, nil
}

func (b *Backend) Enabled() bool {
	return b.Config.Enable
}

func (b *Backend) Backup(paths []string, options map[string]any) error {
	jsonOptions, err := json.Marshal(options)
	if err != nil {
		return err
	}

	cmd := exec.Command(b.Manifest.Bin)
	cmd.Env = append(
		os.Environ(),
		"STANDARD_BACKUPS_COMMAND=backup",
		fmt.Sprintf("STANDARD_BACKUPS_PATHS=%s", strings.Join(paths, ":")),
		fmt.Sprintf("STANDARD_BACKUPS_OPTIONS=%s", jsonOptions),
	)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	return err
}
