package config

import "github.com/santhosh-tekuri/jsonschema/v6"

const hookSchemaUrl = "standard-backups://hook.schema.json"

var (
	hookSchemaDoc = map[string]any{
		"type":     "object",
		"required": []any{"shell", "command"},
		"properties": map[string]any{
			"shell":   map[string]any{"enum": []any{"bash", "sh"}},
			"command": map[string]any{"type": "string"},
		},
	}
	hookSchemaRef = map[string]any{"$ref": hookSchemaUrl}
)

func addHookSchema(compiler *jsonschema.Compiler) error {
	return compiler.AddResource(hookSchemaUrl, hookSchemaDoc)
}

type HookV1 struct {
	Shell   string `mapstructure:"shell"`
	Command string `mapstructure:"command"`
}
