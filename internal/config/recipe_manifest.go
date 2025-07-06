package config

import (
	"errors"
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
	recipeManifestV1SchemaUrl = "standard-backups://recipe-manifest-v1.schema.json"
	_recipeManifestV1Schema   = map[string]any{
		"$schama": "https://json-schema.org/draft/2020-12/schema",
		"$id":     recipeManifestV1SchemaUrl,
		"type":    "object",
		"required": []any{
			"version", "name", "paths",
		},
		"properties": map[string]any{
			"version":     map[string]any{"const": 1},
			"name":        map[string]any{"type": "string"},
			"description": map[string]any{"type": "string"},
			"paths": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
				"minItems": 1,
			},
			"hooks": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"before":     map[string]any{"$ref": "#/$defs/hook"},
					"after":      map[string]any{"$ref": "#/$defs/hook"},
					"on-success": map[string]any{"$ref": "#/$defs/hook"},
					"on-failure": map[string]any{"$ref": "#/$defs/hook"},
				},
			},
		},
		"$defs": map[string]any{
			"hook": map[string]any{
				"type":     "object",
				"required": []any{"shell", "command"},
				"properties": map[string]any{
					"shell":   map[string]any{"enum": []any{"bash", "sh"}},
					"command": map[string]any{"type": "string"},
				},
			},
		},
	}
	recipeManifestV1Schema jsonschema.Schema
)

func loadRecipeManifestV1Schema() (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	err := compiler.AddResource(recipeManifestV1SchemaUrl, _recipeManifestV1Schema)
	if err != nil {
		return nil, err
	}
	schema, err := compiler.Compile(recipeManifestV1SchemaUrl)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func init() {
	res, err := loadRecipeManifestV1Schema()
	if err != nil {
		log.Panicf("[internal error] failed to load recipe manifest v1 schema: %v", err)
	}
	recipeManifestV1Schema = *res
}

type HookV1 struct {
	Shell   string `mapstructure:"shell"`
	Command string `mapstructure:"command"`
}

type HooksV1 struct {
	Before    *HookV1 `mapstructure:"before"`
	After     *HookV1 `mapstructure:"after"`
	OnSuccess *HookV1 `mapstructure:"on-success"`
	OnFailure *HookV1 `mapstructure:"on-failure"`
}

type RecipeManifestV1 struct {
	Version     int      `mapstructure:"version"`
	Name        string   `mapstructure:"name"`
	Description string   `mapstructure:"description"`
	Paths       []string `mapstructure:"paths"`
	Hooks       HooksV1  `mapstructure:"hooks"`
}

func LoadRecipeManifests(dirs []string) ([]RecipeManifestV1, error) {
	manifests := []RecipeManifestV1{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("failed to list recipe manifest in %s: %w", dir, err)
		}

		for _, entry := range entries {
			if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".yaml") {
				fullPath := path.Join(dir, entry.Name())
				manifest, err := loadRecipeManifest(fullPath)
				if err != nil {
					return nil, err
				}
				manifests = append(manifests, *manifest)
			}
		}
	}
	return manifests, nil
}

func loadRecipeManifest(path string) (*RecipeManifestV1, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read recipe manifest %s: %w", path, err)
	}

	rawManifest := map[string]any{}
	err = yaml.Unmarshal(bytes, &rawManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to load recipe manifest %s: %w", path, err)
	}

	err = recipeManifestV1Schema.Validate(rawManifest)
	if err != nil {
		return nil, fmt.Errorf("recipe manifest %s is invalid: %w", path, err)
	}

	var res RecipeManifestV1
	err = mapstructure.Decode(rawManifest, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode recipe manifest %s: %w", path, err)
	}

	return &res, nil
}
