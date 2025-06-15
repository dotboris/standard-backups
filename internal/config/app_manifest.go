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
	appManifestV1SchemaUrl = "standard-backups://app-manifest-v1.schema.json"
	_appManifestV1Schema   = map[string]any{
		"$schama": "https://json-schema.org/draft/2020-12/schema",
		"$id":     appManifestV1SchemaUrl,
		"type":    "object",
		"required": []any{
			"version", "name", "directory",
		},
		"properties": map[string]any{
			"version":     map[string]any{"const": 1},
			"name":        map[string]any{"type": "string"},
			"description": map[string]any{"type": "string"},
			"directory":   map[string]any{"type": "string"},
			"pre-hook":    map[string]any{"type": "string"},
			"post-hook":   map[string]any{"type": "string"},
		},
	}
	appManifestV1Schema jsonschema.Schema
)

func loadAppManifestV1Schema() (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	err := compiler.AddResource(appManifestV1SchemaUrl, _appManifestV1Schema)
	if err != nil {
		return nil, err
	}
	schema, err := compiler.Compile(appManifestV1SchemaUrl)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func init() {
	res, err := loadAppManifestV1Schema()
	if err != nil {
		log.Panicf("[internal error] failed to load app manifest v1 schema: %v", err)
	}
	appManifestV1Schema = *res
}

type AppManifestV1 struct {
	Version     int    `mapstructure:"version"`
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
	Directory   string `mapstructure:"directory"`
	PreHook     string `mapstructure:"pre-hook"`
	PostHook    string `mapstructure:"post-hook"`
}

func LoadAppManifests(dir string) ([]AppManifestV1, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list app manifest in %s: %w", dir, err)
	}

	appManifests := make([]AppManifestV1, 0)
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".yaml") {
			fullPath := path.Join(dir, entry.Name())
			appManifest, err := loadAppManifest(fullPath)
			if err != nil {
				return nil, err
			}
			appManifests = append(appManifests, *appManifest)
		}
	}

	return appManifests, nil
}

func loadAppManifest(path string) (*AppManifestV1, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read app manifest %s: %w", path, err)
	}

	rawManifest := map[string]any{}
	err = yaml.Unmarshal(bytes, &rawManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to load app manifest %s: %w", path, err)
	}

	err = appManifestV1Schema.Validate(rawManifest)
	if err != nil {
		return nil, fmt.Errorf("app manifest %s is invalid: %w", path, err)
	}

	var res AppManifestV1
	err = mapstructure.Decode(rawManifest, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode app manifest %s: %w", path, err)
	}

	return &res, nil
}
