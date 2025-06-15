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
	os.WriteFile(configPath, []byte(`version: 1`), 0644)

	mainConfig, err := LoadMainConfig(configPath, []BackendManifestV1{}, []AppManifestV1{})
	if assert.NoError(t, err) {
		assert.Equal(t, &MainConfig{
			Version: 1,
		}, mainConfig)
	}
}

func TestLoadMainConfigBadVersion(t *testing.T) {
	d := t.TempDir()
	configPath := path.Join(d, "config.yaml")
	os.WriteFile(configPath, []byte(`version: -1`), 0644)

	_, err := LoadMainConfig(configPath, []BackendManifestV1{}, []AppManifestV1{})
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
	os.WriteFile(configPath, []byte(``), 0644)

	_, err := LoadMainConfig(configPath, []BackendManifestV1{}, []AppManifestV1{})
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

func TestLoadMainConfigBadTargetKey(t *testing.T) {
	for _, key := range []string{"42", "-42", "-nope", "no/slash", "no:colon"} {
		t.Run(key, func(t *testing.T) {

			d := t.TempDir()
			configPath := path.Join(d, "config.yaml")
			os.WriteFile(
				configPath,
				[]byte(testutils.DedentYaml(fmt.Sprintf(`
					version: 1
					targets:
						%s:
							backend: bogus
				`, key))),
				0644,
			)

			_, err := LoadMainConfig(
				configPath,
				[]BackendManifestV1{
					{Version: 1, Name: "bogus", ProtocolVersion: 1, Bin: "bogus"},
				},
				[]AppManifestV1{},
			)
			var validationErr *jsonschema.ValidationError
			if assert.Error(t, err) && assert.ErrorAs(t, err, &validationErr) {
				assert.Equal(t,
					testutils.Dedent(fmt.Sprintf(`
						jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
						- at '/targets': additional properties '%s' not allowed
					`, key)),
					validationErr.Error(),
				)
			}
		})
	}
}

func TestLoadMainConfigTargetBadBackend(t *testing.T) {
	d := t.TempDir()
	configPath := path.Join(d, "config.yaml")
	os.WriteFile(
		configPath,
		[]byte(testutils.DedentYaml(`
			version: 1
			targets:
				test:
					backend: nope
		`)),
		0644,
	)
	_, err := LoadMainConfig(
		configPath,
		[]BackendManifestV1{
			{Version: 1, Name: "bogus", ProtocolVersion: 1, Bin: "bogus"},
			{Version: 1, Name: "other", ProtocolVersion: 1, Bin: "other"},
		},
		[]AppManifestV1{},
	)
	var validationErr *jsonschema.ValidationError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &validationErr) {
		assert.Equal(t,
			testutils.Dedent(`
				jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
				- at '/targets/test/backend': value must be one of 'bogus', 'other'
			`),
			validationErr.Error(),
		)
	}
}

func TestLoadMainConfigBadSourceKey(t *testing.T) {
	for _, key := range []string{"42", "-42", "-nope", "no/slash", "no:colon"} {
		t.Run(key, func(t *testing.T) {

			d := t.TempDir()
			configPath := path.Join(d, "config.yaml")
			os.WriteFile(
				configPath,
				[]byte(testutils.DedentYaml(fmt.Sprintf(`
					version: 1
					sources:
						%s:
							app: bogus
							backup-to: []
				`, key))),
				0644,
			)

			_, err := LoadMainConfig(
				configPath,
				[]BackendManifestV1{},
				[]AppManifestV1{
					{Version: 1, Name: "bogus"},
				},
			)
			var validationErr *jsonschema.ValidationError
			if assert.Error(t, err) && assert.ErrorAs(t, err, &validationErr) {
				assert.Equal(t,
					testutils.Dedent(fmt.Sprintf(`
						jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
						- at '/sources': additional properties '%s' not allowed
					`, key)),
					validationErr.Error(),
				)
			}
		})
	}
}

func TestLoadMainConfigTargetBadApp(t *testing.T) {
	d := t.TempDir()
	configPath := path.Join(d, "config.yaml")
	os.WriteFile(
		configPath,
		[]byte(testutils.DedentYaml(`
			version: 1
			sources:
				test:
					app: nope
					backup-to: []
		`)),
		0644,
	)
	_, err := LoadMainConfig(
		configPath,
		[]BackendManifestV1{},
		[]AppManifestV1{
			{Version: 1, Name: "bogus"},
			{Version: 1, Name: "other"},
		},
	)
	var validationErr *jsonschema.ValidationError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &validationErr) {
		assert.Equal(t,
			testutils.Dedent(`
				jsonschema validation failed with 'standard-backups://main-config-v1.schema.json#'
				- at '/sources/test/app': value must be one of 'bogus', 'other'
			`),
			validationErr.Error(),
		)
	}
}
