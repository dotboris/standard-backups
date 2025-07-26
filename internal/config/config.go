package config

import (
	"fmt"
)

// Config describes the entire configuration of `standard-backups` across all config files.
type Config struct {
	Backends   []BackendManifestV1
	Recipes    []RecipeManifestV1
	MainConfig MainConfig
	Secrets    map[string]string
}

func LoadConfig(
	configPath string,
	backendsSearchDirs []string,
	recipesSearchDirs []string,
) (*Config, error) {
	backends, err := LoadBackendManifests(backendsSearchDirs)
	if err != nil {
		return nil, err
	}
	recipes, err := LoadRecipeManifests(recipesSearchDirs)
	if err != nil {
		return nil, err
	}
	mainConfig, err := LoadMainConfig(configPath, backends, recipes)
	if err != nil {
		return nil, err
	}
	secrets, err := loadSecrets(mainConfig.Secrets)
	if err != nil {
		return nil, err
	}
	return &Config{
		Backends:   backends,
		Recipes:    recipes,
		MainConfig: *mainConfig,
		Secrets:    secrets,
	}, nil
}

func (c *Config) GetBackendManifest(name string) (*BackendManifestV1, error) {
	for _, m := range c.Backends {
		if m.Name == name {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("could not find backend named %s", name)
}

func (c *Config) GetRecipeManifest(name string) (*RecipeManifestV1, error) {
	for _, m := range c.Recipes {
		if m.Name == name {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("could not find recipe named %s", name)
}
