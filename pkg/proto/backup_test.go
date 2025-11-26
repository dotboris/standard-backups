package proto

import (
	"errors"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var testBackupReq = &BackupRequest{
	Paths:           []string{"/foo", "/bar"},
	Exclude:         []string{"biz", "baz"},
	DestinationName: "my-dest",
	JobName:         "my-job",
	RawOptions: map[string]any{
		"key":    "value",
		"bool":   true,
		"number": 42.0,
		"array":  []any{1.0, 2.0, 3.0},
		"object": map[string]any{"foo": "bar"},
	},
}

func TestBackupRequestE2E(t *testing.T) {
	testutils.SetEnvFromToEnv(t, testBackupReq)
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
	t.Setenv("STANDARD_BACKUPS_COMMAND", "backup")
	err := b.execute()
	if assert.NoError(t, err) {
		assert.True(t, called, "Backup func was not called")
		assert.Equal(t, testBackupReq, gotReq)
	}
}

func TestBackupError(t *testing.T) {
	t.Setenv("STANDARD_BACKUPS_COMMAND", "backup")
	testutils.SetEnvFromToEnv(t, testBackupReq)
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
	t.Setenv("STANDARD_BACKUPS_COMMAND", "backup")
	b := &BackendImpl{}
	err := b.execute()
	assert.EqualError(t, err, "unhandled command backup")
}
