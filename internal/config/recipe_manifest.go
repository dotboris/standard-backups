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

type RecipeManifestV1 struct {
	Version     int    `mapstructure:"version"`
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
	Directory   string `mapstructure:"directory"`
	PreHook     string `mapstructure:"pre-hook"`
	PostHook    string `mapstructure:"post-hook"`
}

func LoadRecipeManifests(dir string) ([]RecipeManifestV1, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []RecipeManifestV1{}, nil
		}
		return nil, fmt.Errorf("failed to list recipe manifest in %s: %w", dir, err)
	}

	manifests := make([]RecipeManifestV1, 0)
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
