package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCheckRepoExists(t *testing.T) {
	tests := map[string]string{
		"system": "restic",
	}
	restic016, ok := os.LookupEnv("RESTIC_0_16")
	if ok {
		tests["0.16"] = fmt.Sprintf("%s/bin/restic", restic016)
	}
	for name, restic := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("RESTIC", restic)
			t.Run("exists", func(t *testing.T) {
				repo := t.TempDir()
				cmd := exec.Command(restic, "init", "--repo", repo)
				cmd.Env = append(cmd.Env, "RESTIC_PASSWORD=supersecret")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				require.NoError(t, err)

				res, err := checkRepoExists(repo, map[string]string{
					"RESTIC_PASSWORD": "supersecret",
				})
				assert.NoError(t, err)
				assert.True(t, res)
			})
			t.Run("not-exists", func(t *testing.T) {
				res, err := checkRepoExists(t.TempDir(), map[string]string{})
				assert.NoError(t, err)
				assert.False(t, res)
			})
		})
	}
}
