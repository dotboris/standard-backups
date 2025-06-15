package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/goccy/go-yaml"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var (
	backendManifestV1SchemaUrl = "standard-backups://backend-manifest-v1.schema.json"
	_backendManifestV1Schema   = map[string]any{
		"$schama": "https://json-schema.org/draft/2020-12/schema",
		"$id":     backendManifestV1SchemaUrl,
		"type":    "object",
		"required": []any{
			"version", "name", "bin", "protocol-version",
		},
		"properties": map[string]any{
			"version":          map[string]any{"const": 1},
			"name":             map[string]any{"type": "string"},
			"description":      map[string]any{"type": "string"},
			"bin":              map[string]any{"type": "string"},
			"protocol-version": map[string]any{"enum": []any{1}},
		},
	}
	backendManifestV1Schema jsonschema.Schema
)

func loadBackendManifestV1Schema() (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	err := compiler.AddResource(backendManifestV1SchemaUrl, _backendManifestV1Schema)
	if err != nil {
		return nil, err
	}
	schema, err := compiler.Compile(backendManifestV1SchemaUrl)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func init() {
	res, err := loadBackendManifestV1Schema()
	if err != nil {
		log.Panicf("[internal error] failed to load backend manifest v1 schema: %v", err)
	}
	backendManifestV1Schema = *res
}

type BackendManifestV1 struct {
	Version         int    `mapstructure:"version"`
	Name            string `mapstructure:"name"`
	Description     string `mapstructure:"description"`
	Bin             string `mapstructure:"bin"`
	ProtocolVersion int    `mapstructure:"protocol-version"`
}

func LoadBackendManifests(dir string) ([]BackendManifestV1, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list backend manifest in %s: %w", dir, err)
	}

	backendManifests := make([]BackendManifestV1, 0)
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".yaml") {
			fullPath := path.Join(dir, entry.Name())
			backendManifest, err := loadBackendManifest(fullPath)
			if err != nil {
				return nil, err
			}
			backendManifests = append(backendManifests, *backendManifest)
		}
	}

	return backendManifests, nil
}

func loadBackendManifest(path string) (*BackendManifestV1, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read app manifest %s: %w", path, err)
	}

	rawManifest := map[string]any{}
	err = yaml.Unmarshal(bytes, &rawManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to load backend manifest %s: %w", path, err)
	}

	err = backendManifestV1Schema.Validate(rawManifest)
	if err != nil {
		return nil, fmt.Errorf("backend manifest %s is invalid: %w", path, err)
	}

	var res BackendManifestV1
	err = mapstructure.Decode(rawManifest, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode backend manifest %s: %w", path, err)
	}

	return &res, nil
}
