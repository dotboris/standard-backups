package config

import (
	"fmt"
	"os"

	"github.com/go-viper/mapstructure/v2"
	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var (
	mainConfigV1SchemaUrl = "standard-backups://main-config-v1.schema.json"
)

type BackendConfigV1 struct {
	Enable bool
}

type TargetConfigV1 struct {
	Backend string
}

type SourceConfigV1 struct {
	App      string
	BackupTo []string `mapstructure:"backup-to"`
}

// MainConfig is the configuration file that system administrators are expected
// to write. In other words, it's `$dir/config.yaml`.
type MainConfig struct {
	Version  int
	Backends map[string]BackendConfigV1
	Targets  map[string]TargetConfigV1
	Sources  map[string]SourceConfigV1
}

func makeMainConfigSchema(backends []BackendManifestV1, apps []AppManifestV1) (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()

	backendsProperties := map[string]any{}
	backendNames := []any{}
	for _, backend := range backends {
		backendsProperties[backend.Name] = map[string]any{"$ref": "#/$defs/backend"}
		backendNames = append(backendNames, backend.Name)
	}

	appNames := []any{}
	for _, app := range apps {
		appNames = append(appNames, app.Name)
	}

	err := compiler.AddResource(mainConfigV1SchemaUrl, map[string]any{
		"$schama": "https://json-schema.org/draft/2020-12/schema",
		"$id":     mainConfigV1SchemaUrl,
		"type":    "object",
		"required": []any{
			"version",
		},
		"properties": map[string]any{
			"version": map[string]any{"const": 1},
			"backends": map[string]any{
				"type":       "object",
				"properties": backendsProperties,
			},
			"targets": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"patternProperties": map[string]any{
					"^[a-zA-Z][a-zA-Z0-9_-]*$": map[string]any{
						"type":     "object",
						"required": []any{"backend"},
						"properties": map[string]any{
							"backend": map[string]any{"enum": backendNames},
						},
					},
				},
			},
			"sources": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"patternProperties": map[string]any{
					"^[a-zA-Z][a-zA-Z0-9_-]*$": map[string]any{
						"type":     "object",
						"required": []any{"app", "backup-to"},
						"properties": map[string]any{
							"app": map[string]any{"enum": appNames},
							"backup-to": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "string",
								},
							},
						},
					},
				},
			},
		},
		"$defs": map[string]any{
			"backend": map[string]any{},
		},
	})
	if err != nil {
		return nil, err
	}
	schema, err := compiler.Compile(mainConfigV1SchemaUrl)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func LoadMainConfig(path string, backends []BackendManifestV1, apps []AppManifestV1) (*MainConfig, error) {
	schema, err := makeMainConfigSchema(backends, apps)
	if err != nil {
		return nil, fmt.Errorf("[internal error] failed to build main config schema: %w", err)
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to main config %s: %w", path, err)
	}

	rawConfig := map[string]any{}
	err = yaml.Unmarshal(bytes, &rawConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load main config %s: %w", path, err)
	}

	err = schema.Validate(rawConfig)
	if err != nil {
		return nil, fmt.Errorf("main config %s is invalid: %w", path, err)
	}

	var res MainConfig
	err = mapstructure.Decode(rawConfig, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode main config %s: %w", path, err)
	}

	return &res, nil
}
