package proto

import (
	"errors"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var testBackupReqFull = &BackupRequest{
	Paths:           []string{"/foo", "/bar"},
	Exclude:         []string{"biz", "baz"},
	DestinationName: "my-dest",
	VariantName:     "my-variant",
	JobName:         "my-job",
	RawOptions: map[string]any{
		"key":    "value",
		"bool":   true,
		"number": 42.0,
		"array":  []any{1.0, 2.0, 3.0},
		"object": map[string]any{"foo": "bar"},
	},
}

var testBackupReqMinimal = &BackupRequest{
	Paths:           []string{"/foo", "/bar"},
	DestinationName: "my-dest",
	JobName:         "my-job",
}

func TestBackupRequestE2E(t *testing.T) {
	testCases := map[string]*BackupRequest{
		"full":    testBackupReqFull,
		"minimal": testBackupReqMinimal,
	}
	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			testutils.SetEnvFromToEnv(t, req)
			var (
				gotReq *BackupRequest
				called bool
			)
			b := &BackendImpl{
				Backup: func(r *BackupRequest) error {
					called = true
					gotReq = r
					return nil
				},
			}
			t.Setenv(COMMAND_ENV, "backup")
			err := b.execute()
			if assert.NoError(t, err) {
				assert.True(t, called, "Backup func was not called")
				assert.Equal(t, req, gotReq)
			}
		})
	}
}

func TestBackupError(t *testing.T) {
	t.Setenv(COMMAND_ENV, "backup")
	testutils.SetEnvFromToEnv(t, testBackupReqFull)
	expectedErr := errors.New("oops")
	b := &BackendImpl{
		Backup: func(req *BackupRequest) error {
			return expectedErr
		},
	}
	err := b.execute()
	assert.ErrorIs(t, err, expectedErr)
}

func TestBackupNotImplemented(t *testing.T) {
	t.Setenv(COMMAND_ENV, "backup")
	b := &BackendImpl{}
	err := b.execute()
	assert.EqualError(t, err, "unhandled command backup")
}
