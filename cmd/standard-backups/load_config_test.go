package main

import (
	"fmt"
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

func createTestBackendManifest(t *testing.T, dir string, name string, bin string) {
	t.Helper()
	p := path.Join(dir, fmt.Sprintf("standard-backups/backends/%s.yaml", name))
	err := os.MkdirAll(path.Dir(p), 0o755)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = os.WriteFile(p, []byte(testutils.DedentYaml(fmt.Sprintf(`
		version: 1
		protocol-version: 1
		name: %s
		bin: %s
	`, name, bin))), 0o644)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

func createTestRecipeManifest(t *testing.T, dir string, name string, description string) {
	t.Helper()
	p := path.Join(dir, fmt.Sprintf("standard-backups/recipes/%s.yaml", name))
	err := os.MkdirAll(path.Dir(p), 0o755)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = os.WriteFile(p, []byte(testutils.DedentYaml(fmt.Sprintf(`
		version: 1
		name: %s
		description: %s
		paths: [bogus]
	`, name, description))), 0o644)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
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
			createTestBackendManifest(t, dir, "my-backend", "bogus")

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
			createTestRecipeManifest(t, dir, "my-recipe", "bogus")

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

func TestLoadConfigBackendLoadOrder(t *testing.T) {
	getBackendBin := func() (string, error) {
		t.Helper()
		c, err := loadConfig()
		if err != nil {
			return "", err
		}
		b, err := c.GetBackendManifest("my-backend")
		if err != nil {
			return "", err
		}
		return b.Bin, nil
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CONFIG_DIRS", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_DATA_DIRS", "")

	setBogusConfig(t)

	bin, err := getBackendBin()
	assert.Empty(t, bin)
	assert.Error(t, err)

	dataDir1 := t.TempDir()
	createTestBackendManifest(t, dataDir1, "my-backend", "data dir 1")
	t.Setenv("XDG_DATA_DIRS", dataDir1)
	bin, err = getBackendBin()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "data dir 1", bin)

	dataDir2 := t.TempDir()
	createTestBackendManifest(t, dataDir2, "my-backend", "data dir 2")
	t.Setenv("XDG_DATA_DIRS", fmt.Sprintf("%s:%s", dataDir2, dataDir1))
	bin, err = getBackendBin()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "data dir 2", bin)

	dataHome := t.TempDir()
	createTestBackendManifest(t, dataHome, "my-backend", "data home")
	t.Setenv("XDG_DATA_HOME", dataHome)
	bin, err = getBackendBin()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "data home", bin)

	configDir1 := t.TempDir()
	createTestBackendManifest(t, configDir1, "my-backend", "config dir 1")
	t.Setenv("XDG_CONFIG_DIRS", configDir1)
	bin, err = getBackendBin()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "config dir 1", bin)

	configDir2 := t.TempDir()
	createTestBackendManifest(t, configDir2, "my-backend", "config dir 2")
	t.Setenv("XDG_CONFIG_DIRS", fmt.Sprintf("%s:%s", configDir2, configDir1))
	bin, err = getBackendBin()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "config dir 2", bin)

	configHome := t.TempDir()
	createTestBackendManifest(t, configHome, "my-backend", "config home")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	bin, err = getBackendBin()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "config home", bin)
}

func TestLoadConfigRecipesLoadOrder(t *testing.T) {
	getRecipeDescription := func() (string, error) {
		t.Helper()
		c, err := loadConfig()
		if err != nil {
			return "", err
		}
		r, err := c.GetRecipeManifest("my-recipe")
		if err != nil {
			return "", err
		}
		return r.Description, nil
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CONFIG_DIRS", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_DATA_DIRS", "")

	setBogusConfig(t)

	description, err := getRecipeDescription()
	assert.Empty(t, description)
	assert.Error(t, err)

	dataDir1 := t.TempDir()
	createTestRecipeManifest(t, dataDir1, "my-recipe", "data dir 1")
	t.Setenv("XDG_DATA_DIRS", dataDir1)
	description, err = getRecipeDescription()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "data dir 1", description)

	dataDir2 := t.TempDir()
	createTestRecipeManifest(t, dataDir2, "my-recipe", "data dir 2")
	t.Setenv("XDG_DATA_DIRS", fmt.Sprintf("%s:%s", dataDir2, dataDir1))
	description, err = getRecipeDescription()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "data dir 2", description)

	dataHome := t.TempDir()
	createTestRecipeManifest(t, dataHome, "my-recipe", "data home")
	t.Setenv("XDG_DATA_HOME", dataHome)
	description, err = getRecipeDescription()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "data home", description)

	configDir1 := t.TempDir()
	createTestRecipeManifest(t, configDir1, "my-recipe", "config dir 1")
	t.Setenv("XDG_CONFIG_DIRS", configDir1)
	description, err = getRecipeDescription()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "config dir 1", description)

	configDir2 := t.TempDir()
	createTestRecipeManifest(t, configDir2, "my-recipe", "config dir 2")
	t.Setenv("XDG_CONFIG_DIRS", fmt.Sprintf("%s:%s", configDir2, configDir1))
	description, err = getRecipeDescription()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "config dir 2", description)

	configHome := t.TempDir()
	createTestRecipeManifest(t, configHome, "my-recipe", "config home")
	t.Setenv("XDG_CONFIG_HOME", configHome)
	description, err = getRecipeDescription()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "config home", description)
}
