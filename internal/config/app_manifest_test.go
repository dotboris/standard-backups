package config

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAppManifests(t *testing.T) {
	for _, ext := range []string{".yaml", ".yml"} {
		t.Run(fmt.Sprintf("single_file%s", ext), func(t *testing.T) {
			d := t.TempDir()
			os.WriteFile(path.Join(d, fmt.Sprintf("example%s", ext)), []byte(`version: 1
name: example 1
description: the first example
directory: /app/to/backup/1
pre-hook: echo before
post-hook: echo after
`), 0644)
			appManifests, err := LoadAppManifests(d)
			if assert.NoError(t, err) {
				assert.Equal(t, []AppManifest{
					{
						Version:     1,
						Name:        "example 1",
						Description: "the first example",
						Directory:   "/app/to/backup/1",
						PreHook:     "echo before",
						PostHook:    "echo after",
					},
				}, appManifests)
			}
		})
	}

	t.Run("multiple_files", func(t *testing.T) {
		d := t.TempDir()
		os.WriteFile(path.Join(d, "app1.yaml"), []byte(`version: 1
name: app1
description: the app1
directory: /app/to/backup/1
pre-hook: echo before app1
post-hook: echo after app1
`), 0644)
		os.WriteFile(path.Join(d, "app2.yaml"), []byte(`version: 1
name: app2
description: the app2
directory: /app/to/backup/2
pre-hook: echo before app2
post-hook: echo after app2
`), 0644)
		appManifests, err := LoadAppManifests(d)
		if assert.NoError(t, err) {
			assert.Equal(t, []AppManifest{
				{
					Version:     1,
					Name:        "app1",
					Description: "the app1",
					Directory:   "/app/to/backup/1",
					PreHook:     "echo before app1",
					PostHook:    "echo after app1",
				},
				{
					Version:     1,
					Name:        "app2",
					Description: "the app2",
					Directory:   "/app/to/backup/2",
					PreHook:     "echo before app2",
					PostHook:    "echo after app2",
				},
			}, appManifests)
		}
	})

	t.Run("ignore_non_yaml", func(t *testing.T) {
		d := t.TempDir()
		os.WriteFile(path.Join(d, "bogus.txt"), []byte("bogus"), 0644)
		appManifests, err := LoadAppManifests(d)
		if assert.NoError(t, err) {
			assert.Equal(t, []AppManifest{}, appManifests)
		}
	})

	t.Run("empty_dir", func(t *testing.T) {
		d := t.TempDir()
		appManifests, err := LoadAppManifests(d)
		if assert.NoError(t, err) {
			assert.Equal(t, []AppManifest{}, appManifests)
		}
	})

	t.Run("no_hooks", func(t *testing.T) {
		d := t.TempDir()
		os.WriteFile(path.Join(d, "app.yaml"), []byte(`version: 1
name: app
description: app description
directory: /app/to/backup
`), 0644)
		appManifests, err := LoadAppManifests(d)
		if assert.NoError(t, err) {
			assert.Equal(t, []AppManifest{
				{
					Version:     1,
					Name:        "app",
					Description: "app description",
					Directory:   "/app/to/backup",
					PreHook:     "",
					PostHook:    "",
				},
			}, appManifests)
		}
	})

	t.Run("invalid_empty_file", func(t *testing.T) {
		d := t.TempDir()
		os.WriteFile(path.Join(d, "app.yaml"), []byte(""), 0644)
		_, err := LoadAppManifests(d)
		assert.Error(t, err)
	})

	t.Run("invalid_bad_version", func(t *testing.T) {
		d := t.TempDir()
		os.WriteFile(path.Join(d, "app.yaml"), []byte(`version: -1
name: app
description: app description
directory: /app/to/backup
`), 0644)
		_, err := LoadAppManifests(d)
		assert.Error(t, err)
	})
}
