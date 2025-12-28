package e2e

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/dotboris/standard-backups/internal/redact"
	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRedactSecretsBackend(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tc.AddBogusRecipe(t, "bogus")

	stdoutBackend := path.Join(t.TempDir(), "opts-stdout.sh")
	err := os.WriteFile(stdoutBackend, []byte(testutils.Dedent(`
		#!/usr/bin/env bash
		echo hello from $0
		echo "$STANDARD_BACKUPS_OPTIONS"
	`)), 0x755)
	if !assert.NoError(t, err) {
		return
	}
	tc.AddBackend("opts-stdout", stdoutBackend)

	stderrBackend := path.Join(t.TempDir(), "opts-stderr.sh")
	err = os.WriteFile(stderrBackend, []byte(testutils.Dedent(`
		#!/usr/bin/env bash
		echo hello from $0
		echo >&2 "$STANDARD_BACKUPS_OPTIONS"
	`)), 0x755)
	if !assert.NoError(t, err) {
		return
	}
	tc.AddBackend("opts-stderr", stderrBackend)

	tc.WriteConfig(testutils.DedentYaml(`
		version: 1
		secrets:
			literal:
				literal: supersecret
		destinations:
			stdout:
				backend: opts-stdout
				options:
					literal: '{{ .Secrets.literal }}'
			stderr:
				backend: opts-stderr
				options:
					literal: '{{ .Secrets.literal }}'
		jobs:
			my-job:
				recipe: bogus
				backup-to: [stdout, stderr]
	`))

	cmd := testutils.StandardBackups(t, "backup", "my-job")
	tc.Apply(cmd)
	stdout := bytes.NewBuffer(nil)
	cmd.Stdout = stdout
	stderr := bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	err = cmd.Run()
	if !assert.NoError(t, err,
		"stdout: %s\nstderr: %s\n", stdout.String(), stderr.String()) {
		return
	}
	assert.NotContains(t, stdout.String(), "supersecret")
	assert.Contains(t, stdout.String(), redact.REPLACE)
	assert.NotContains(t, stderr.String(), "supersecret")
	assert.Contains(t, stderr.String(), redact.REPLACE)
}
