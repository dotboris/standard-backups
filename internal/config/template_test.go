package config

import (
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func TestConfigTemplateApply(t *testing.T) {
	tpl := configTemplate{
		Secrets: map[string]string{"s": "supersecret"},
	}
	var value map[string]any
	err := yaml.Unmarshal(
		// Using YAML here to ensure that this get converted into the right type.
		// The yaml lib is a little quirky about which type it chooses under the
		// hood when parsing and that has an effect on how we walk the tree.
		[]byte(testutils.DedentYaml(`
			simple: '{{ .Secrets.s }}'
			list: ['{{ .Secrets.s }}']
			nested:
				secret: '{{ .Secrets.s }}'
			listOfObjects:
				- secret: '{{ .Secrets.s }}'
			deepList:
				list: ['{{ .Secrets.s }}']
		`)),
		&value,
	)
	if !assert.NoError(t, err) {
		return
	}

	res, err := tpl.Apply("test", value)
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]any{
			"simple": "supersecret",
			"list":   []any{"supersecret"},
			"nested": map[string]any{"secret": "supersecret"},
			"listOfObjects": []any{
				map[string]any{"secret": "supersecret"},
			},
			"deepList": map[string]any{
				"list": []any{"supersecret"},
			},
		}, res)
	}
}
