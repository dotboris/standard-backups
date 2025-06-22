package config

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestLoadRecipeManifestsSingleFile(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(path.Join(d, "example.yaml"), []byte(`version: 1
name: example 1
description: the first example
paths: [/app/to/backup/1]
pre-hook: echo before
post-hook: echo after
`), 0644)
	manifests, err := LoadRecipeManifests(d)
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{
			{
				Version:     1,
				Name:        "example 1",
				Description: "the first example",
				Paths:       []string{"/app/to/backup/1"},
				PreHook:     "echo before",
				PostHook:    "echo after",
			},
		}, manifests)
	}
}

func TestLoadRecipeManifestsMultipleFiles(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(path.Join(d, "app1.yaml"), []byte(`version: 1
name: app1
description: the app1
paths: [/app/to/backup/1]
pre-hook: echo before app1
post-hook: echo after app1
`), 0644)
	os.WriteFile(path.Join(d, "app2.yaml"), []byte(`version: 1
name: app2
description: the app2
paths: [/app/to/backup/2]
pre-hook: echo before app2
post-hook: echo after app2
`), 0644)
	manifests, err := LoadRecipeManifests(d)
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{
			{
				Version:     1,
				Name:        "app1",
				Description: "the app1",
				Paths:       []string{"/app/to/backup/1"},
				PreHook:     "echo before app1",
				PostHook:    "echo after app1",
			},
			{
				Version:     1,
				Name:        "app2",
				Description: "the app2",
				Paths:       []string{"/app/to/backup/2"},
				PreHook:     "echo before app2",
				PostHook:    "echo after app2",
			},
		}, manifests)
	}
}

func TestLoadRecipeManifestsIgnoreNonYaml(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(path.Join(d, "bogus.txt"), []byte("bogus"), 0644)
	manifests, err := LoadRecipeManifests(d)
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{}, manifests)
	}
}

func TestLoadRecipeManifestsEmptyDir(t *testing.T) {
	d := t.TempDir()
	manifests, err := LoadRecipeManifests(d)
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{}, manifests)
	}
}

func TestLoadRecipeManifestsMissingDir(t *testing.T) {
	d := t.TempDir()
	manifests, err := LoadRecipeManifests(path.Join(d, "does-not-exist"))
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{}, manifests)
	}
}

func TestLoadRecipeManifestsNoHooks(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(path.Join(d, "app.yaml"), []byte(`version: 1
name: app
description: app description
paths: [/app/to/backup]
`), 0644)
	manifests, err := LoadRecipeManifests(d)
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{
			{
				Version:     1,
				Name:        "app",
				Description: "app description",
				Paths:       []string{"/app/to/backup"},
				PreHook:     "",
				PostHook:    "",
			},
		}, manifests)
	}
}

func TestLoadRecipeManifestsInvalidEmptyFile(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(path.Join(d, "app.yaml"), []byte(""), 0644)
	_, err := LoadRecipeManifests(d)
	assert.Error(t, err)
}

func TestLoadRecipeManifestsInvalidBadVersion(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(path.Join(d, "app.yaml"), []byte(`version: -1
name: app
description: app description
paths: [/app/to/backup]
`), 0644)
	_, err := LoadRecipeManifests(d)
	assert.Error(t, err)
}

func TestLoadRecipeManifestsInvalidNoPaths(t *testing.T) {
	d := t.TempDir()
	p := path.Join(d, "app.yaml")
	os.WriteFile(
		p,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: app
			description: app description
		`)),
		0644,
	)
	_, err := LoadRecipeManifests(d)
	assert.Equal(
		t,
		testutils.Dedent(fmt.Sprintf(`
			recipe manifest %s is invalid: jsonschema validation failed with 'standard-backups://recipe-manifest-v1.schema.json#'
			- at '': missing property 'paths'
		`, p)),
		err.Error(),
	)
}

func TestLoadRecipeManifestsInvalidEmptyPaths(t *testing.T) {
	d := t.TempDir()
	p := path.Join(d, "app.yaml")
	os.WriteFile(
		p,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: app
			description: app description
			paths: []
		`)),
		0644,
	)
	_, err := LoadRecipeManifests(d)
	assert.Equal(
		t,
		testutils.Dedent(fmt.Sprintf(`
			recipe manifest %s is invalid: jsonschema validation failed with 'standard-backups://recipe-manifest-v1.schema.json#'
			- at '/paths': minItems: got 0, want 1
		`, p)),
		err.Error(),
	)
}
