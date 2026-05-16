package e2e

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/dotboris/standard-backups/internal/testbackend"
	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testListBackupsConfigFull = testutils.DedentYaml(`
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

var testListBackupsConfigMinimal = testutils.DedentYaml(`
	version: 1
	destinations:
		my-dest:
			backend: test
`)

func TestListBackups(t *testing.T) {
	testCases := map[string]struct {
		config string
		dest   string
	}{
		"full_dest": {
			config: testListBackupsConfigFull,
			dest:   "my-dest",
		},
		"full_variant": {
			config: testListBackupsConfigFull,
			dest:   "my-dest/my-variant",
		},
		"minimal": {
			config: testListBackupsConfigMinimal,
			dest:   "my-dest",
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			tc := testutils.NewTestConfig(t)
			tb := testbackend.New(t, testbackend.Impl{
				ListBackups: testbackend.ListBackupsImpl{
					BaseImpl: testbackend.BaseImpl{Enable: true},
					Res: &proto.ListBackupsResponse{
						Backups: []proto.ListBackupsResponseItem{
							{
								Id:          "base",
								Time:        "2026-01-01T00:00:00Z",
								Job:         "whatever",
								Destination: "some-dest",
								Variant:     "some-variant",
								Size:        42069,
								Extra:       map[string]any{},
							},
							{
								Id: "nothing",
							},
							{
								Id: "extras",
								Extra: map[string]any{
									"key":    "value",
									"bool":   true,
									"number": 42.0,
									"array":  []any{1.0, 2.0, 3.0},
									"object": map[string]any{"foo": "bar"},
								},
							},
						},
					},
				},
			})
			tb.AddSelf(tc)
			tc.WriteConfig(testutils.DedentYaml(testCase.config))

			cmd := testutils.StandardBackups(t, "list-backups", testCase.dest, "--json")
			tc.Apply(cmd)
			tb.Apply(cmd)
			stdout := bytes.NewBufferString("")
			cmd.Stdout = stdout
			err := cmd.Run()
			require.NoError(t, err)

			trace := tb.RequireTrace("list-backups")
			snaps.MatchJSON(t, trace)
			snaps.MatchJSON(t, stdout.String())
		})
	}
}

func TestListBackupsError(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tb := testbackend.New(t, testbackend.Impl{
		ListBackups: testbackend.ListBackupsImpl{
			BaseImpl: testbackend.BaseImpl{
				Enable: true,
				Error:  "oops",
			},
		},
	})
	tb.AddSelf(tc)
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(testListBackupsConfigFull))

	cmd := testutils.StandardBackups(t, "list-backups", "my-dest")
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

func TestListBackupsNotImplemented(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tb := testbackend.New(t, testbackend.Impl{})
	tb.AddSelf(tc)
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(testListBackupsConfigFull))

	cmd := testutils.StandardBackups(t, "list-backups", "my-dest")
	tc.Apply(cmd)
	tb.Apply(cmd)
	stderr := bytes.NewBufferString("")
	cmd.Stderr = stderr
	err := cmd.Run()

	var exitError *exec.ExitError
	require.Error(t, err)
	assert.ErrorAs(t, err, &exitError)
	assert.Equal(t, 1, exitError.ExitCode())
	assert.Contains(t, stderr.String(), "Error: unhandled command list-backups\n")
}
