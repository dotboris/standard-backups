package proto

import (
	"errors"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var (
	testRestoreReqFull = &RestoreRequest{
		RawOptions: map[string]any{
			"key":    "value",
			"bool":   true,
			"number": 42.0,
			"array":  []any{1.0, 2.0, 3.0},
			"object": map[string]any{"foo": "bar"},
		},
		DestinationName: "my-dest",
		VariantName:     "my-variant",
		BackupId:        "backup-1234",
		OutputDir:       "/path/to/something",
	}
	testRestoreReqMinimal = &RestoreRequest{
		DestinationName: "my-dest",
		BackupId:        "backup-1234",
		OutputDir:       "/path/to/something",
	}
)

func TestRestoreE2E(t *testing.T) {
	testCases := map[string]*RestoreRequest{
		"full":    testRestoreReqFull,
		"minimal": testRestoreReqMinimal,
	}
	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			testutils.SetEnvFromToEnv(t, req)
			var (
				gotReq *RestoreRequest
				called bool
			)
			b := &BackendImpl{
				Restore: func(r *RestoreRequest) error {
					called = true
					gotReq = r
					return nil
				},
			}
			t.Setenv(COMMAND_ENV, "restore")
			err := b.execute()
			if assert.NoError(t, err) {
				assert.True(t, called, "Restore func was not called")
				assert.Equal(t, req, gotReq)
			}
		})
	}
}

func TestRestoreError(t *testing.T) {
	t.Setenv(COMMAND_ENV, "restore")
	testutils.SetEnvFromToEnv(t, testRestoreReqFull)
	expectedErr := errors.New("oops")
	b := &BackendImpl{
		Restore: func(r *RestoreRequest) error {
			return expectedErr
		},
	}
	err := b.execute()
	assert.ErrorIs(t, err, expectedErr)
}

func TestRestoreNotImplemented(t *testing.T) {
	t.Setenv(COMMAND_ENV, "restore")
	testutils.SetEnvFromToEnv(t, testRestoreReqFull)
	b := &BackendImpl{}
	err := b.execute()
	assert.EqualError(t, err, "unhandled command restore")
}
