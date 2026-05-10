package proto

import (
	"errors"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var testExecReqFull = &ExecRequest{
	Args:            []string{"foo", "bar", "--baz", "-b"},
	DestinationName: "my-dest",
	VariantName:     "my-variant",
	RawOptions: map[string]any{
		"key":    "value",
		"bool":   true,
		"number": 42.0,
		"array":  []any{1.0, 2.0, 3.0},
		"object": map[string]any{"foo": "bar"},
	},
}
var testExecReqMinimal = &ExecRequest{}

func TestExecRequestE2E(t *testing.T) {
	testCases := map[string]*ExecRequest{
		"full":    testExecReqFull,
		"minimal": testExecReqMinimal,
	}
	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			testutils.SetEnvFromToEnv(t, req)
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
			t.Setenv(COMMAND_ENV, "exec")
			err := b.execute()
			if assert.NoError(t, err) {
				assert.True(t, called, "Exec func was not called")
				assert.Equal(t, req, gotReq)
			}
		})
	}
}

func TestExecError(t *testing.T) {
	t.Setenv(COMMAND_ENV, "exec")
	testutils.SetEnvFromToEnv(t, testExecReqFull)
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
	t.Setenv(COMMAND_ENV, "exec")
	b := &BackendImpl{}
	err := b.execute()
	assert.EqualError(t, err, "unhandled command exec")
}
