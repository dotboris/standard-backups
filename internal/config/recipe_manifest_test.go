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
	p := path.Join(d, "example.yaml")
	err := os.WriteFile(p,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: example 1
			description: the first example
			paths: [/app/to/backup/1]
			hooks:
				before:
					shell: bash
					command: echo before
				after:
					shell: sh
					command: echo after
				on-success:
					shell: bash
					command: echo success
				on-failure:
					shell: bash
					command: echo failure
		`)),
		0o644)
	if !assert.NoError(t, err) {
		return
	}
	manifests, err := LoadRecipeManifests([]string{d})
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{
			{
				path:        p,
				Version:     1,
				Name:        "example 1",
				Description: "the first example",
				Paths:       []string{"/app/to/backup/1"},
				Hooks: HooksV1{
					Before: &HookV1{
						Shell:   "bash",
						Command: "echo before",
					},
					After: &HookV1{
						Shell:   "sh",
						Command: "echo after",
					},
					OnSuccess: &HookV1{
						Shell:   "bash",
						Command: "echo success",
					},
					OnFailure: &HookV1{
						Shell:   "bash",
						Command: "echo failure",
					},
				},
			},
		}, manifests)
	}
}

func TestLoadRecipeManifestsMultipleFiles(t *testing.T) {
	d1 := t.TempDir()
	p1 := path.Join(d1, "app1.yaml")
	err := os.WriteFile(p1,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: app1
			description: the app1
			paths: [/app/to/backup/1]
			hooks:
				before:
					shell: bash
					command: echo before 1
				after:
					shell: sh
					command: echo after 1
				on-success:
					shell: bash
					command: echo success 1
				on-failure:
					shell: bash
					command: echo failure 1
		`)),
		0o644)
	if !assert.NoError(t, err) {
		return
	}
	d2 := t.TempDir()
	p2 := path.Join(d2, "app2.yaml")
	err = os.WriteFile(p2,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: app2
			description: the app2
			paths: [/app/to/backup/2]
			hooks:
				before:
					shell: bash
					command: echo before 2
				after:
					shell: sh
					command: echo after 2
				on-success:
					shell: bash
					command: echo success 2
				on-failure:
					shell: bash
					command: echo failure 2
		`)),
		0o644)
	if !assert.NoError(t, err) {
		return
	}
	manifests, err := LoadRecipeManifests([]string{d1, d2})
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{
			{
				path:        p1,
				Version:     1,
				Name:        "app1",
				Description: "the app1",
				Paths:       []string{"/app/to/backup/1"},
				Hooks: HooksV1{
					Before: &HookV1{
						Shell:   "bash",
						Command: "echo before 1",
					},
					After: &HookV1{
						Shell:   "sh",
						Command: "echo after 1",
					},
					OnSuccess: &HookV1{
						Shell:   "bash",
						Command: "echo success 1",
					},
					OnFailure: &HookV1{
						Shell:   "bash",
						Command: "echo failure 1",
					},
				},
			},
			{
				path:        p2,
				Version:     1,
				Name:        "app2",
				Description: "the app2",
				Paths:       []string{"/app/to/backup/2"},
				Hooks: HooksV1{
					Before: &HookV1{
						Shell:   "bash",
						Command: "echo before 2",
					},
					After: &HookV1{
						Shell:   "sh",
						Command: "echo after 2",
					},
					OnSuccess: &HookV1{
						Shell:   "bash",
						Command: "echo success 2",
					},
					OnFailure: &HookV1{
						Shell:   "bash",
						Command: "echo failure 2",
					},
				},
			},
		}, manifests)
	}
}

func TestLoadRecipeManifestsIgnoreNonYaml(t *testing.T) {
	d := t.TempDir()
	err := os.WriteFile(path.Join(d, "bogus.txt"), []byte("bogus"), 0o644)
	if !assert.NoError(t, err) {
		return
	}
	manifests, err := LoadRecipeManifests([]string{d})
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{}, manifests)
	}
}

func TestLoadRecipeManifestsEmptyDir(t *testing.T) {
	d := t.TempDir()
	manifests, err := LoadRecipeManifests([]string{d})
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{}, manifests)
	}
}

func TestLoadRecipeManifestsMissingDir(t *testing.T) {
	d := t.TempDir()
	manifests, err := LoadRecipeManifests([]string{path.Join(d, "does-not-exist")})
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{}, manifests)
	}
}

func TestLoadRecipeManifestsNoHooks(t *testing.T) {
	d := t.TempDir()
	p := path.Join(d, "app.yaml")
	err := os.WriteFile(p,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: app
			description: app description
			paths: [/app/to/backup]
		`)),
		0o644)
	if !assert.NoError(t, err) {
		return
	}
	manifests, err := LoadRecipeManifests([]string{d})
	if assert.NoError(t, err) {
		assert.Equal(t, []RecipeManifestV1{
			{
				path:        p,
				Version:     1,
				Name:        "app",
				Description: "app description",
				Paths:       []string{"/app/to/backup"},
			},
		}, manifests)
	}
}

func TestLoadRecipeManifestsInvalidEmptyFile(t *testing.T) {
	d := t.TempDir()
	err := os.WriteFile(path.Join(d, "app.yaml"), []byte(""), 0o644)
	if !assert.NoError(t, err) {
		return
	}
	_, err = LoadRecipeManifests([]string{d})
	assert.Error(t, err)
}

func TestLoadRecipeManifestsInvalidBadVersion(t *testing.T) {
	d := t.TempDir()
	err := os.WriteFile(
		path.Join(d, "app.yaml"),
		[]byte(testutils.DedentYaml(`
			version: -1
			name: app
			description: app description
			paths: [/app/to/backup]
		`)),
		0o644)
	if !assert.NoError(t, err) {
		return
	}
	_, err = LoadRecipeManifests([]string{d})
	assert.Error(t, err)
}

func TestLoadRecipeManifestsInvalidNoPaths(t *testing.T) {
	d := t.TempDir()
	p := path.Join(d, "app.yaml")
	err := os.WriteFile(
		p,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: app
			description: app description
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	_, err = LoadRecipeManifests([]string{d})
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
	err := os.WriteFile(
		p,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: app
			description: app description
			paths: []
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	_, err = LoadRecipeManifests([]string{d})
	assert.Equal(
		t,
		testutils.Dedent(fmt.Sprintf(`
			recipe manifest %s is invalid: jsonschema validation failed with 'standard-backups://recipe-manifest-v1.schema.json#'
			- at '/paths': minItems: got 0, want 1
		`, p)),
		err.Error(),
	)
}

func TestLoadRecipeManifestsInvalidHooks(t *testing.T) {
	for _, hook := range []string{"before", "after", "on-success", "on-failure"} {
		t.Run(fmt.Sprintf("%s/bad_shell", hook), func(t *testing.T) {
			d := t.TempDir()
			p := path.Join(d, "app.yaml")
			err := os.WriteFile(
				p,
				[]byte(testutils.DedentYaml(fmt.Sprintf(`
					version: 1
					name: app
					description: app description
					paths: [bogus]
					hooks:
						%s:
							shell: nope
							command: echo test
				`, hook))),
				0o644,
			)
			if !assert.NoError(t, err) {
				return
			}
			_, err = LoadRecipeManifests([]string{d})
			if assert.Error(t, err) {
				assert.Equal(t,
					testutils.Dedent(fmt.Sprintf(`
						recipe manifest %s is invalid: jsonschema validation failed with 'standard-backups://recipe-manifest-v1.schema.json#'
						- at '/hooks/%s/shell': value must be one of 'bash', 'sh'
					`, p, hook)),
					err.Error(),
				)
			}
		})
		t.Run(fmt.Sprintf("%s/no_shell", hook), func(t *testing.T) {
			d := t.TempDir()
			p := path.Join(d, "app.yaml")
			err := os.WriteFile(
				p,
				[]byte(testutils.DedentYaml(fmt.Sprintf(`
					version: 1
					name: app
					description: app description
					paths: [bogus]
					hooks:
						%s:
							command: echo test
				`, hook))),
				0o644,
			)
			if !assert.NoError(t, err) {
				return
			}
			_, err = LoadRecipeManifests([]string{d})
			if assert.Error(t, err) {
				assert.Equal(t,
					testutils.Dedent(fmt.Sprintf(`
						recipe manifest %s is invalid: jsonschema validation failed with 'standard-backups://recipe-manifest-v1.schema.json#'
						- at '/hooks/%s': missing property 'shell'
					`, p, hook)),
					err.Error(),
				)
			}
		})
		t.Run(fmt.Sprintf("%s/no_command", hook), func(t *testing.T) {
			d := t.TempDir()
			p := path.Join(d, "app.yaml")
			err := os.WriteFile(
				p,
				[]byte(testutils.DedentYaml(fmt.Sprintf(`
					version: 1
					name: app
					description: app description
					paths: [bogus]
					hooks:
						%s:
							shell: sh
				`, hook))),
				0o644,
			)
			if !assert.NoError(t, err) {
				return
			}
			_, err = LoadRecipeManifests([]string{d})
			if assert.Error(t, err) {
				assert.Equal(t,
					testutils.Dedent(fmt.Sprintf(`
						recipe manifest %s is invalid: jsonschema validation failed with 'standard-backups://recipe-manifest-v1.schema.json#'
						- at '/hooks/%s': missing property 'command'
					`, p, hook)),
					err.Error(),
				)
			}
		})
	}
}
