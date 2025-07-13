package config

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSuccess(t *testing.T) {
	bin := path.Join(t.TempDir(), "backend-bin")
	_, err := os.Create(bin)
	if !assert.NoError(t, err) {
		return
	}
	err = os.Chmod(bin, 0o755)
	if !assert.NoError(t, err) {
		return
	}

	c := Config{
		Backends: []BackendManifestV1{
			{
				path:            "bogus/backends.d/b.yaml",
				Version:         1,
				Name:            "b",
				Bin:             bin,
				ProtocolVersion: 0,
			},
		},
		Recipes: []RecipeManifestV1{
			{
				path:    "bogus/recipes.d/r.yaml",
				Version: 0,
				Name:    "r",
				Paths:   []string{"bogus/back-me-up"},
			},
		},
		MainConfig: MainConfig{
			path:    "bogus/config.yaml",
			Version: 1,
			Destinations: map[string]DestinationConfigV1{
				"d": {
					Backend: "b",
					Options: map[string]any{},
				},
			},
			Jobs: map[string]JobConfigV1{
				"j": {
					Recipe:   "r",
					BackupTo: []string{"d"},
				},
			},
		},
	}

	res := c.Validate()
	assert.Empty(t, res)
}

func TestValidateBackendBinNotFound(t *testing.T) {
	c := Config{
		Backends: []BackendManifestV1{
			{
				path:            "path/to/backend.yaml",
				Version:         1,
				Name:            "busted-backend",
				Bin:             "bogus/does-not-exist",
				ProtocolVersion: 1,
			},
		},
	}
	res := c.Validate()
	assert.Len(t, res, 1)
	assert.Equal(t, "path/to/backend.yaml", res[0].File)
	assert.Equal(t, "/bin", res[0].FieldPath)
	assert.EqualError(t, res[0].Err,
		"stat bogus/does-not-exist: no such file or directory")
}

func TestValidateBackendBinIsDir(t *testing.T) {
	bin := path.Join(t.TempDir(), "some-dir")
	err := os.Mkdir(bin, 0o755)
	if !assert.NoError(t, err) {
		return
	}
	c := Config{
		Backends: []BackendManifestV1{
			{
				path:            "path/to/backend.yaml",
				Version:         1,
				Name:            "busted-backend",
				Bin:             bin,
				ProtocolVersion: 1,
			},
		},
	}
	res := c.Validate()
	assert.Len(t, res, 1)
	assert.Equal(t, "path/to/backend.yaml", res[0].File)
	assert.Equal(t, "/bin", res[0].FieldPath)
	assert.EqualError(t, res[0].Err,
		fmt.Sprintf("%s is a directory", bin))
}

func TestValidateMainConfigUnknownDestination(t *testing.T) {
	c := Config{
		MainConfig: MainConfig{
			path: "bogus/config.yaml",
			Jobs: map[string]JobConfigV1{
				"my-job": {
					BackupTo: []string{"nope"},
				},
			},
		},
	}
	res := c.Validate()
	assert.Len(t, res, 1)
	assert.Equal(t, "bogus/config.yaml", res[0].File)
	assert.Equal(t, "/jobs/my-job/backup-to/0", res[0].FieldPath)
	assert.EqualError(t, res[0].Err, "unknown destination nope")
}
