package config

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
)

func TestLoadMainConfigMinimalConfig(t *testing.T) {
	d := t.TempDir()
	configPath := path.Join(d, "config.yaml")
	err := os.WriteFile(configPath, []byte(`version: 1`), 0o644)
	if !assert.NoError(t, err) {
		return
	}

	mainConfig, err := LoadMainConfig(configPath, []BackendManifestV1{}, []RecipeManifestV1{})
	if assert.NoError(t, err) {
		assert.Equal(t, &MainConfig{
			path:    configPath,
			Version: 1,
		}, mainConfig)
	}
}

func TestLoadMainConfigBadVersion(t *testing.T) {
	d := t.TempDir()
	configPath := path.Join(d, "config.yaml")
	err := os.WriteFile(configPath, []byte(`version: -1`), 0o644)
	if !assert.NoError(t, err) {
		return
	}

	_, err = LoadMainConfig(configPath, []BackendManifestV1{}, []RecipeManifestV1{})
	assert.Error(t, err)
	var validationErr *jsonschema.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t,
		testutils.Dedent(`
			jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
			- at '/version': value must be 1
		`),
		validationErr.Error(),
	)
}

func TestLoadMainConfigEmptyConfig(t *testing.T) {
	d := t.TempDir()
	configPath := path.Join(d, "config.yaml")
	err := os.WriteFile(configPath, []byte(``), 0o644)
	if !assert.NoError(t, err) {
		return
	}

	_, err = LoadMainConfig(configPath, []BackendManifestV1{}, []RecipeManifestV1{})
	assert.Error(t, err)
	var validationErr *jsonschema.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t,
		testutils.Dedent(`
			jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
			- at '': missing property 'version'
		`),
		validationErr.Error(),
	)
}

func TestLoadMainConfigBadDestinationKey(t *testing.T) {
	for _, key := range []string{"42", "-42", "-nope", "no/slash", "no:colon"} {
		t.Run(key, func(t *testing.T) {
			d := t.TempDir()
			configPath := path.Join(d, "config.yaml")
			err := os.WriteFile(
				configPath,
				[]byte(testutils.DedentYaml(fmt.Sprintf(`
					version: 1
					destinations:
						%s:
							backend: bogus
				`, key))),
				0o644,
			)
			if !assert.NoError(t, err) {
				return
			}

			_, err = LoadMainConfig(
				configPath,
				[]BackendManifestV1{
					{Version: 1, Name: "bogus", ProtocolVersion: 1, Bin: "bogus"},
				},
				[]RecipeManifestV1{},
			)
			var validationErr *jsonschema.ValidationError
			if assert.Error(t, err) && assert.ErrorAs(t, err, &validationErr) {
				assert.Equal(t,
					testutils.Dedent(fmt.Sprintf(`
						jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
						- at '/destinations': additional properties '%s' not allowed
					`, key)),
					validationErr.Error(),
				)
			}
		})
	}
}

func TestLoadMainConfigDestinationBadBackend(t *testing.T) {
	d := t.TempDir()
	configPath := path.Join(d, "config.yaml")
	err := os.WriteFile(
		configPath,
		[]byte(testutils.DedentYaml(`
			version: 1
			destinations:
				test:
					backend: nope
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	_, err = LoadMainConfig(
		configPath,
		[]BackendManifestV1{
			{Version: 1, Name: "bogus", ProtocolVersion: 1, Bin: "bogus"},
			{Version: 1, Name: "other", ProtocolVersion: 1, Bin: "other"},
		},
		[]RecipeManifestV1{},
	)
	var validationErr *jsonschema.ValidationError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &validationErr) {
		assert.Equal(t,
			testutils.Dedent(`
				jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
				- at '/destinations/test/backend': value must be one of 'bogus', 'other'
			`),
			validationErr.Error(),
		)
	}
}

func TestLoadMainConfigBadJobKey(t *testing.T) {
	for _, key := range []string{"42", "-42", "-nope", "no/slash", "no:colon"} {
		t.Run(key, func(t *testing.T) {
			d := t.TempDir()
			configPath := path.Join(d, "config.yaml")
			err := os.WriteFile(
				configPath,
				[]byte(testutils.DedentYaml(fmt.Sprintf(`
					version: 1
					jobs:
						%s:
							recipe: bogus
							backup-to: []
				`, key))),
				0o644,
			)
			if !assert.NoError(t, err) {
				return
			}

			_, err = LoadMainConfig(
				configPath,
				[]BackendManifestV1{},
				[]RecipeManifestV1{
					{Version: 1, Name: "bogus"},
				},
			)
			var validationErr *jsonschema.ValidationError
			if assert.Error(t, err) && assert.ErrorAs(t, err, &validationErr) {
				assert.Equal(t,
					testutils.Dedent(fmt.Sprintf(`
						jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
						- at '/jobs': additional properties '%s' not allowed
					`, key)),
					validationErr.Error(),
				)
			}
		})
	}
}

func TestLoadMainConfigTargetBadRecipe(t *testing.T) {
	d := t.TempDir()
	configPath := path.Join(d, "config.yaml")
	err := os.WriteFile(
		configPath,
		[]byte(testutils.DedentYaml(`
			version: 1
			jobs:
				test:
					recipe: nope
					backup-to: []
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	_, err = LoadMainConfig(
		configPath,
		[]BackendManifestV1{},
		[]RecipeManifestV1{
			{Version: 1, Name: "bogus"},
			{Version: 1, Name: "other"},
		},
	)
	var validationErr *jsonschema.ValidationError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &validationErr) {
		assert.Equal(t,
			testutils.Dedent(`
				jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
				- at '/jobs/test/recipe': value must be one of 'bogus', 'other'
			`),
			validationErr.Error(),
		)
	}
}

func TestLoadMainConfigSecretDefinition(t *testing.T) {
	testCases := []struct {
		name           string
		config         string
		expectedSecret SecretConfigV1
	}{
		{
			name: "from-file",
			config: testutils.DedentYaml(`
				version: 1
				secrets:
					mySecret:
						from-file: /path/to/secret.txt
			`),
			expectedSecret: SecretConfigV1{
				FromFile: "/path/to/secret.txt",
			},
		},
		{
			name: "literal",
			config: testutils.DedentYaml(`
				version: 1
				secrets:
					mySecret:
						literal: what could possibly go wrong
			`),
			expectedSecret: SecretConfigV1{
				Literal: "what could possibly go wrong",
			},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			d := t.TempDir()
			configPath := path.Join(d, "config.yaml")
			err := os.WriteFile(configPath, []byte(test.config), 0o644)
			if !assert.NoError(t, err) {
				return
			}
			c, err := LoadMainConfig(
				configPath,
				[]BackendManifestV1{},
				[]RecipeManifestV1{},
			)
			if assert.NoError(t, err) {
				assert.Equal(t, test.expectedSecret, c.Secrets["mySecret"])
			}
		})
	}
}

func TestLoadMainConfigSecretDefinitionOneProperty(t *testing.T) {
	d := t.TempDir()
	configPath := path.Join(d, "config.yaml")
	err := os.WriteFile(
		configPath,
		[]byte(testutils.DedentYaml(`
			version: 1
			secrets:
				mySecret:
					from-file: /path/to/secret.txt
					literal: supersecret
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	_, err = LoadMainConfig(
		configPath,
		[]BackendManifestV1{},
		[]RecipeManifestV1{},
	)
	var validationErr *jsonschema.ValidationError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &validationErr) {
		assert.Equal(t,
			testutils.Dedent(`
				jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
				- at '/secrets/mySecret': maxProperties: got 2, want 1
			`),
			validationErr.Error(),
		)
	}
}
