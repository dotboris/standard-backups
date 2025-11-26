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
			before:
				shell: bash
				command: echo before
			after:
				shell: sh
				command: echo after
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
				Before: &HookV1{
					Shell:   "bash",
					Command: "echo before",
				},
				After: &HookV1{
					Shell:   "sh",
					Command: "echo after",
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
			before:
				shell: bash
				command: echo before 1
			after:
				shell: sh
				command: echo after 1
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
			before:
				shell: bash
				command: echo before 2
			after:
				shell: sh
				command: echo after 2
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
				Before: &HookV1{
					Shell:   "bash",
					Command: "echo before 1",
				},
				After: &HookV1{
					Shell:   "sh",
					Command: "echo after 1",
				},
			},
			{
				path:        p2,
				Version:     1,
				Name:        "app2",
				Description: "the app2",
				Paths:       []string{"/app/to/backup/2"},
				Before: &HookV1{
					Shell:   "bash",
					Command: "echo before 2",
				},
				After: &HookV1{
					Shell:   "sh",
					Command: "echo after 2",
				},
			},
		}, manifests)
	}
}

func TestLoadRecipeManifestsExclude(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "not-set",
			content: testutils.DedentYaml(`
				version: 1
				name: example 1
				description: the first example
				paths: [/app/to/backup/1]
			`),
			expected: nil,
		},
		{
			name: "empty",
			content: testutils.DedentYaml(`
				version: 1
				name: example 1
				description: the first example
				paths: [/app/to/backup/1]
				exclude: []
			`),
			expected: []string{},
		},
		{
			name: "one-value",
			content: testutils.DedentYaml(`
				version: 1
				name: example 1
				description: the first example
				paths: [/app/to/backup/1]
				exclude:
					- value 1
			`),
			expected: []string{"value 1"},
		},
		{
			name: "two-values",
			content: testutils.DedentYaml(`
				version: 1
				name: example 1
				description: the first example
				paths: [/app/to/backup/1]
				exclude:
					- value 1
					- value 2
			`),
			expected: []string{"value 1", "value 2"},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			d := t.TempDir()
			p := path.Join(d, "example.yaml")
			err := os.WriteFile(p, []byte(testCase.content), 0o644)
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
						Exclude:     testCase.expected,
					},
				}, manifests)
			}
		})
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
	for _, hook := range []string{"before", "after"} {
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
						- at '/%s/shell': value must be one of 'bash', 'sh'
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
						- at '/%s': missing property 'shell'
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
						- at '/%s': missing property 'command'
					`, p, hook)),
					err.Error(),
				)
			}
		})
	}
}
