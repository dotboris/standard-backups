package main

import (
	"os"
	"path"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func setBogusConfig(t *testing.T) {
	t.Helper()
	confPath := path.Join(t.TempDir(), "config.yaml")
	err := os.WriteFile(confPath, []byte(testutils.DedentYaml(`
		version: 1
	`)), 0o644)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	old := configPath
	configPath = confPath
	t.Cleanup(func() {
		configPath = old
	})
}

func TestLoadConfigBackend(t *testing.T) {
	for _, envVar := range []string{
		"XDG_CONFIG_HOME",
		"XDG_CONFIG_DIRS",
		"XDG_DATA_HOME",
		"XDG_DATA_DIRS",
	} {
		t.Run(envVar, func(t *testing.T) {
			setBogusConfig(t)

			dir := t.TempDir()
			t.Setenv(envVar, dir)

			p := path.Join(dir, "standard-backups/backends/my-backend.yaml")
			err := os.MkdirAll(path.Dir(p), 0o755)
			if !assert.NoError(t, err) {
				return
			}
			err = os.WriteFile(p, []byte(testutils.DedentYaml(`
				version: 1
				protocol-version: 1
				name: my-backend
				bin: bogus
			`)), 0o644)
			if !assert.NoError(t, err) {
				return
			}

			c, err := loadConfig()
			if !assert.NoError(t, err) {
				return
			}
			b, err := c.GetBackendManifest("my-backend")
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, "my-backend", b.Name)
		})
	}
}

func TestLoadConfigRecipes(t *testing.T) {
	for _, envVar := range []string{
		"XDG_CONFIG_HOME",
		"XDG_CONFIG_DIRS",
		"XDG_DATA_HOME",
		"XDG_DATA_DIRS",
	} {
		t.Run(envVar, func(t *testing.T) {
			setBogusConfig(t)

			dir := t.TempDir()
			t.Setenv(envVar, dir)

			p := path.Join(dir, "standard-backups/recipes/my-recipe.yaml")
			err := os.MkdirAll(path.Dir(p), 0o755)
			if !assert.NoError(t, err) {
				return
			}
			err = os.WriteFile(p, []byte(testutils.DedentYaml(`
				version: 1
				name: my-recipe
				paths: [bogus]
			`)), 0o644)
			if !assert.NoError(t, err) {
				return
			}

			c, err := loadConfig()
			if !assert.NoError(t, err) {
				return
			}
			r, err := c.GetRecipeManifest("my-recipe")
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, "my-recipe", r.Name)
		})
	}
}
