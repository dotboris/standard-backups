package e2e

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/dotboris/standard-backups/internal/testbackend"
	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBackupConfigFull = testutils.DedentYaml(`
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
			variants:
				my-variant:
					number: 69
	jobs:
		my-job:
			recipe: bogus
			backup-to: [my-dest/my-variant]
`)

func TestBackup(t *testing.T) {
	testCases := map[string]struct {
		config string
		recipe string
	}{
		"full": {
			config: testBackupConfigFull,
			recipe: testutils.DedentYaml(`
				version: 1
				name: bogus
				paths: [/path/to/backup, /path/to/also/backup]
				exclude: [exclude/me, also/me]
			`),
		},
		"minimal": {
			config: testutils.DedentYaml(`
				version: 1
				destinations:
					my-dest:
						backend: test
				jobs:
					my-job:
						recipe: bogus
						backup-to: [my-dest]
			`),
			recipe: testutils.DedentYaml(`
				version: 1
				name: bogus
				paths: [/path/to/backup]
			`),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			tc := testutils.NewTestConfig(t)
			tb := testbackend.New(t, testbackend.Impl{
				Backup: testbackend.BaseImpl{
					Enable: true,
				},
			})
			tb.AddSelf(tc)
			tc.AddRecipe("bogus", testCase.recipe)
			tc.WriteConfig(testutils.DedentYaml(testCase.config))

			cmd := testutils.StandardBackups(t, "backup", "my-job")
			tc.Apply(cmd)
			tb.Apply(cmd)
			err := cmd.Run()
			require.NoError(t, err)

			trace := tb.RequireTrace("backup")
			snaps.MatchJSON(t, trace)
		})
	}
}

func TestBackupError(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tb := testbackend.New(t, testbackend.Impl{
		Backup: testbackend.BaseImpl{
			Enable: true,
			Error:  "oops",
		},
	})
	tb.AddSelf(tc)
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(testBackupConfigFull))

	cmd := testutils.StandardBackups(t, "backup", "my-job")
	tc.Apply(cmd)
	tb.Apply(cmd)
	stderr := bytes.NewBufferString("")
	cmd.Stderr = stderr
	err := cmd.Run()

	var exitError *exec.ExitError
	require.Error(t, err)
	assert.ErrorAs(t, err, &exitError)
	assert.Equal(t, 1, exitError.ExitCode())
	assert.Contains(t, stderr.String(), "Error: oops\n")
}

func TestBackupNotImplemented(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tb := testbackend.New(t, testbackend.Impl{})
	tb.AddSelf(tc)
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(testBackupConfigFull))

	cmd := testutils.StandardBackups(t, "backup", "my-job")
	tc.Apply(cmd)
	tb.Apply(cmd)
	stderr := bytes.NewBufferString("")
	cmd.Stderr = stderr
	err := cmd.Run()

	var exitError *exec.ExitError
	require.Error(t, err)
	assert.ErrorAs(t, err, &exitError)
	assert.Equal(t, 1, exitError.ExitCode())
	assert.Contains(t, stderr.String(), "Error: unhandled command backup\n")
}
