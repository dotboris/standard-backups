package config

import (
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

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

func TestConfigTemplateError(t *testing.T) {
	tpl := configTemplate{}
	var value map[string]any
	err := yaml.Unmarshal(
		// Using YAML here to ensure that this get converted into the right type.
		// The yaml lib is a little quirky about which type it chooses under the
		// hood when parsing and that has an effect on how we walk the tree.
		[]byte(testutils.DedentYaml(`
			listOfObjects:
				- secret: ...
				- secret: '{{ crashHere }}'
		`)),
		&value,
	)
	require.NoError(t, err)

	_, err = tpl.Apply("test", value)
	assert.EqualError(
		t,
		err,
		`template: test.listOfObjects.1.secret:1: function "crashHere" not defined`,
	)
}
