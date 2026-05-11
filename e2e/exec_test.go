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

var testExecConfigFull = testutils.DedentYaml(`
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

var testExecConfigMinimal = testutils.DedentYaml(`
	version: 1
	destinations:
		my-dest:
			backend: test
`)

var (
	testExecArgsBackend     = []string{"-b", "test"}
	testExecArgsDestination = []string{"-d", "my-dest"}
	testExecArgsVariant     = []string{"-d", "my-dest/my-variant"}
)

func TestExec(t *testing.T) {
	testCases := map[string]struct {
		config string
		args   []string
	}{
		"full_backend": {
			config: testExecConfigFull,
			args:   testExecArgsBackend,
		},
		"full_destination": {
			config: testExecConfigFull,
			args:   testExecArgsDestination,
		},
		"full_variant": {
			config: testExecConfigFull,
			args:   testExecArgsVariant,
		},
		"minimal_backend": {
			config: testExecConfigMinimal,
			args:   testExecArgsBackend,
		},
		"minimal_destination": {
			config: testExecConfigMinimal,
			args:   testExecArgsDestination,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			tc := testutils.NewTestConfig(t)
			tb := testbackend.New(t, testbackend.Impl{
				Exec: testbackend.BaseImpl{
					Enable: true,
				},
			})
			tb.AddSelf(tc)
			tc.WriteConfig(testutils.DedentYaml(testCase.config))

			cmd := testutils.StandardBackups(t, "exec")
			cmd.Args = append(cmd.Args, testCase.args...)
			tc.Apply(cmd)
			tb.Apply(cmd)
			cmd.Args = append(cmd.Args, "--", "foo", "bar", "--baz", "--qux=2", "-x")
			err := cmd.Run()
			require.NoError(t, err)

			trace := tb.RequireTrace("exec")
			snaps.MatchJSON(t, trace)
		})
	}
}

func TestExecError(t *testing.T) {
	testCases := map[string][]string{
		"backend":     testExecArgsBackend,
		"destination": testExecArgsDestination,
	}
	for name, args := range testCases {
		t.Run(name, func(t *testing.T) {
			tc := testutils.NewTestConfig(t)
			tb := testbackend.New(t, testbackend.Impl{
				Exec: testbackend.BaseImpl{
					Enable: true,
					Error:  "oops",
				},
			})
			tb.AddSelf(tc)
			tc.WriteConfig(testutils.DedentYaml(testExecConfigFull))

			cmd := testutils.StandardBackups(t, "exec")
			cmd.Args = append(cmd.Args, args...)
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
		})
	}
}

func TestExecNotImplemented(t *testing.T) {
	testCases := map[string][]string{
		"backend":     testExecArgsBackend,
		"destination": testExecArgsDestination,
	}
	for name, args := range testCases {
		t.Run(name, func(t *testing.T) {
			tc := testutils.NewTestConfig(t)
			tb := testbackend.New(t, testbackend.Impl{})
			tb.AddSelf(tc)
			tc.WriteConfig(testutils.DedentYaml(testExecConfigFull))

			cmd := testutils.StandardBackups(t, "exec")
			cmd.Args = append(cmd.Args, args...)
			tc.Apply(cmd)
			tb.Apply(cmd)
			stderr := bytes.NewBufferString("")
			cmd.Stderr = stderr
			err := cmd.Run()

			var exitError *exec.ExitError
			require.Error(t, err)
			assert.ErrorAs(t, err, &exitError)
			assert.Equal(t, 1, exitError.ExitCode())
			assert.Contains(t, stderr.String(), "Error: unhandled command exec\n")
		})
	}
}

func TestExecBackendNotFound(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tb := testbackend.New(t, testbackend.Impl{})
	tb.AddSelf(tc)
	tc.WriteConfig(testExecConfigFull)

	cmd := testutils.StandardBackups(t, "exec", "-b", "does-not-exist")
	tc.Apply(cmd)
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	err := cmd.Run()
	assert.EqualError(t, err, "exit status 1")
	assert.Equal(t, "Error: could not find backend named does-not-exist\n", stderr.String())
}

func TestExecDestinationNotFound(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tb := testbackend.New(t, testbackend.Impl{})
	tb.AddSelf(tc)
	tc.WriteConfig(testExecConfigFull)

	cmd := testutils.StandardBackups(t, "exec", "-d", "does-not-exist")
	tc.Apply(cmd)
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	err := cmd.Run()
	assert.EqualError(t, err, "exit status 1")
	assert.Equal(
		t,
		"Error: unknown destination does-not-exist: destinations.does-not-exist not in main config\n",
		stderr.String(),
	)
}
