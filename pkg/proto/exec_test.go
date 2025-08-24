package proto

import (
	"errors"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var testExecReq = &ExecRequest{
	Args:            []string{"foo", "bar", "--baz", "-b"},
	DestinationName: "my-dest",
	RawOptions: map[string]any{
		"key":    "value",
		"bool":   true,
		"number": 42.0,
		"array":  []any{1.0, 2.0, 3.0},
		"object": map[string]any{"foo": "bar"},
	},
}

func TestExecRequestE2E(t *testing.T) {
	testutils.SetEnvFromToEnv(t, testExecReq)
	var (
		gotReq *ExecRequest
		called bool
	)
	b := &BackendImpl{
		Exec: func(r *ExecRequest) error {
			called = true
			gotReq = r
			return nil
		},
	}
	t.Setenv("STANDARD_BACKUPS_COMMAND", "exec")
	err := b.execute()
	if assert.NoError(t, err) {
		assert.True(t, called, "Exec func was not called")
		assert.Equal(t, testExecReq, gotReq)
	}
}

func TestExecError(t *testing.T) {
	t.Setenv("STANDARD_BACKUPS_COMMAND", "exec")
	testutils.SetEnvFromToEnv(t, testExecReq)
	expectedErr := errors.New("oops")
	b := &BackendImpl{
		Exec: func(req *ExecRequest) error {
			return expectedErr
		},
	}
	err := b.execute()
	assert.ErrorIs(t, err, expectedErr)
}

func TestExecNotImplemented(t *testing.T) {
	t.Setenv("STANDARD_BACKUPS_COMMAND", "exec")
	b := &BackendImpl{}
	err := b.execute()
	assert.EqualError(t, err, "unhandled command exec")
}
