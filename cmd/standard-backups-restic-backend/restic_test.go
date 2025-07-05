package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionsToArgs(t *testing.T) {
	cases := []struct {
		name    string
		options map[string]any
		args    []string
	}{
		{
			name:    "string",
			options: map[string]any{"foo": "bar"},
			args:    []string{"--foo", "bar"},
		},
		{
			name:    "int",
			options: map[string]any{"foo": 42},
			args:    []string{"--foo", "42"},
		},
		{
			name:    "float-no-digits",
			options: map[string]any{"foo": 42.0},
			args:    []string{"--foo", "42"},
		},
		{
			name:    "float-with-digits",
			options: map[string]any{"foo": 42.5},
			args:    []string{"--foo", "42.5"},
		},
		{
			name:    "bool-true",
			options: map[string]any{"foo": true},
			args:    []string{"--foo"},
		},
		{
			name:    "bool-false",
			options: map[string]any{"foo": false},
			args:    []string{},
		},
		{
			name:    "empty",
			options: map[string]any{},
			args:    []string{},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			res, err := optionsToArgs(test.options)
			if assert.NoError(t, err) {
				assert.Equal(t, test.args, res)
			}
		})
	}
}
