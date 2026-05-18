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

var testRestoreConfigFull = testutils.DedentYaml(`
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
			default-variant: default
			variants:
				my-variant:
					number: 69
				default:
					default: true
`)

var testRestoreConfigMinimal = testutils.DedentYaml(`
	version: 1
	destinations:
		my-dest:
			backend: test
`)

func TestRestore(t *testing.T) {
	testCases := map[string]struct {
		config string
		dest   string
	}{
		"full_dest": {
			config: testRestoreConfigFull,
			dest:   "my-dest",
		},
		"full_variant": {
			config: testRestoreConfigFull,
			dest:   "my-dest/my-variant",
		},
		"minimal": {
			config: testRestoreConfigMinimal,
			dest:   "my-dest",
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			tc := testutils.NewTestConfig(t)
			tb := testbackend.New(t, testbackend.Impl{
				Restore: testbackend.BaseImpl{
					Enable: true,
				},
			})
			tb.AddSelf(tc)
			tc.WriteConfig(testutils.DedentYaml(testCase.config))

			cmd := testutils.StandardBackups(
				t,
				"restore",
				testCase.dest,
				"my-backup-id",
				"path/to/restore/dir",
			)
			tc.Apply(cmd)
			tb.Apply(cmd)
			err := cmd.Run()
			require.NoError(t, err)

			trace := tb.RequireTrace("restore")
			snaps.MatchJSON(t, trace)
		})
	}
}

func TestRestoreError(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tb := testbackend.New(t, testbackend.Impl{
		Restore: testbackend.BaseImpl{
			Enable: true,
			Error:  "oops",
		},
	})
	tb.AddSelf(tc)
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(testRestoreConfigFull))

	cmd := testutils.StandardBackups(t, "restore", "my-dest", "my-backup-id", "path/to/restore/dir")
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

func TestRestoreNotImplemented(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tb := testbackend.New(t, testbackend.Impl{})
	tb.AddSelf(tc)
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(testRestoreConfigFull))

	cmd := testutils.StandardBackups(t, "restore", "my-dest", "my-backup-id", "path/to/restore/dir")
	tc.Apply(cmd)
	tb.Apply(cmd)
	stderr := bytes.NewBufferString("")
	cmd.Stderr = stderr
	err := cmd.Run()

	var exitError *exec.ExitError
	require.Error(t, err)
	assert.ErrorAs(t, err, &exitError)
	assert.Equal(t, 1, exitError.ExitCode())
	assert.Contains(t, stderr.String(), "Error: unhandled command restore\n")
}
