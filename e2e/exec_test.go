package e2e

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestExecWithBackend(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tc.AddBogusRecipe(t, "bogus")
	dumpBackend := testutils.NewDumpBackend(t)
	tc.AddBackend("test", dumpBackend.Path)
	tc.WriteConfig(testutils.DedentYaml(`
		version: 1
		destinations:
			my-dest:
				backend: test
				options:
					number: 42
					string: forty two
					bool: true
					obj: { yay: true }
					array: [1, 2, 3]
		jobs:
			my-job:
				recipe: bogus
				backup-to: [my-dest]
	`))

	cmd := testutils.StandardBackups(t, "exec", "-b", "test")
	cmd.Args = append(cmd.Args, tc.Args()...)
	cmd.Args = append(cmd.Args, "--", "foo", "bar", "--baz", "--qux=2", "-x")
	err := cmd.Run()
	if !assert.NoError(t, err) {
		return
	}

	// Options don't get passed around with `-b`
	assert.Nil(t, dumpBackend.ReadJsonMap("STANDARD_BACKUPS_OPTIONS"))
	assert.Equal(
		t,
		[]any{"foo", "bar", "--baz", "--qux=2", "-x"},
		dumpBackend.ReadJsonArray("STANDARD_BACKUPS_ARGS"),
	)
	assert.Equal(t, "", dumpBackend.ReadString("STANDARD_BACKUPS_DESTINATION_NAME"))
}

func TestExecWithDestination(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tc.AddBogusRecipe(t, "bogus")
	dumpBackend := testutils.NewDumpBackend(t)
	tc.AddBackend("test", dumpBackend.Path)
	tc.WriteConfig(testutils.DedentYaml(`
		version: 1
		destinations:
			my-dest:
				backend: test
				options:
					number: 42
					string: forty two
					bool: true
					obj: { yay: true }
					array: [1, 2, 3]
		jobs:
			my-job:
				recipe: bogus
				backup-to: [my-dest]
	`))

	cmd := testutils.StandardBackups(t, "exec", "-d", "my-dest")
	cmd.Args = append(cmd.Args, tc.Args()...)
	cmd.Args = append(cmd.Args, "--", "foo", "bar", "--baz", "--qux=2", "-x")
	err := cmd.Run()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, map[string]any{
		"number": 42.0,
		"string": "forty two",
		"bool":   true,
		"obj":    map[string]any{"yay": true},
		"array":  []any{1.0, 2.0, 3.0},
	}, dumpBackend.ReadJsonMap("STANDARD_BACKUPS_OPTIONS"))
	assert.Equal(
		t,
		[]any{"foo", "bar", "--baz", "--qux=2", "-x"},
		dumpBackend.ReadJsonArray("STANDARD_BACKUPS_ARGS"),
	)
	assert.Equal(t, "my-dest", dumpBackend.ReadString("STANDARD_BACKUPS_DESTINATION_NAME"))
}

func TestExecBackendNotFound(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tc.AddBogusRecipe(t, "bogus")
	dumpBackend := testutils.NewDumpBackend(t)
	tc.AddBackend("test", dumpBackend.Path)
	tc.WriteConfig(testutils.DedentYaml(`
		version: 1
	`))

	cmd := testutils.StandardBackups(t, "exec", "-b", "does-not-exist")
	cmd.Args = append(cmd.Args, tc.Args()...)
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	err := cmd.Run()
	assert.EqualError(t, err, "exit status 1")
	assert.Equal(t, "Error: could not find backend named does-not-exist\n", stderr.String())
}

func TestExecDestinationNotFound(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tc.AddBogusRecipe(t, "bogus")
	dumpBackend := testutils.NewDumpBackend(t)
	tc.AddBackend("test", dumpBackend.Path)
	tc.WriteConfig(testutils.DedentYaml(`
		version: 1
	`))

	cmd := testutils.StandardBackups(t, "exec", "-d", "does-not-exist")
	cmd.Args = append(cmd.Args, tc.Args()...)
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	err := cmd.Run()
	assert.EqualError(t, err, "exit status 1")
	assert.Equal(t, "Error: could not find destination named does-not-exist\n", stderr.String())
}

func TestExecInnerError(t *testing.T) {
	crashBackend := path.Join(t.TempDir(), "crash-backend.sh")
	err := os.WriteFile(crashBackend, []byte(testutils.Dedent(`
		#!/bin/bash
		echo oops
		exit 42
	`)), 0o755)
	if !assert.NoError(t, err) {
		return
	}

	for _, testCase := range []struct {
		name string
		args []string
	}{
		{"backend", []string{"-b", "test"}},
		{"destination", []string{"-d", "my-dest"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			tc := testutils.NewTestConfig(t)
			tc.AddBogusRecipe(t, "bogus")
			tc.AddBackend("test", crashBackend)
			tc.WriteConfig(testutils.DedentYaml(`
				version: 1
				destinations:
					my-dest:
						backend: test
				jobs:
					my-job:
						recipe: bogus
						backup-to: [my-dest]
			`))

			cmd := testutils.StandardBackups(t, "exec")
			cmd.Args = append(cmd.Args, tc.Args()...)
			cmd.Args = append(cmd.Args, testCase.args...)
			stdout := bytes.Buffer{}
			cmd.Stdout = &stdout
			stderr := bytes.Buffer{}
			cmd.Stderr = &stderr
			err = cmd.Run()
			assert.EqualError(t, err, "exit status 1")
			assert.Equal(t, "oops\n", stdout.String())
			assert.Equal(t, "Error: exit status 42\n", stderr.String())
		})
	}
}
