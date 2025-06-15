package config

import "path"

// Config describes the entire configuration of `standard-backups` across all config files.
type Config struct {
	Backends   []BackendManifestV1
	Apps       []AppManifestV1
	MainConfig MainConfig // TODO
}

// LoadConfig loads the entire `standard-backups` config from a given directory.
// Different parts of the config are loaded like so: Backends are loaded from
// `$dir/backends.d/*.yaml`. Apps are loaded from `$dir/apps.d/*.yaml`. Main
// Config is loaded from `$dir/config.yaml`.
func LoadConfig(dir string) (*Config, error) {
	backends, err := LoadBackendManifests(path.Join(dir, "backends.d"))
	if err != nil {
		return nil, err
	}
	apps, err := LoadAppManifests(path.Join(dir, "apps.d"))
	if err != nil {
		return nil, err
	}
	mainConfig, err := LoadMainConfig(path.Join(dir, "config.yaml"), backends, apps)
	if err != nil {
		return nil, err
	}
	return &Config{
		Backends:   backends,
		Apps:       apps,
		MainConfig: *mainConfig,
	}, nil
}
