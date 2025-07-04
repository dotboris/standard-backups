package proto

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackupRequestE2E(t *testing.T) {
	req := &BackupRequest{
		Paths: []string{"/foo", "/bar"},
		RawOptions: map[string]any{
			"key":    "value",
			"bool":   true,
			"number": 42.0,
			"array":  []any{1.0, 2.0, 3.0},
			"object": map[string]any{"foo": "bar"},
		},
	}
	env, err := req.ToEnv()
	if !assert.NoError(t, err) {
		return
	}
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		t.Setenv(key, value)
	}
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
	err = b.execute()
	if assert.NoError(t, err) {
		assert.True(t, called, "Backup func was not called")
		assert.Equal(t, req, gotReq)
	}
}

func TestBackupError(t *testing.T) {
	t.Setenv("STANDARD_BACKUPS_COMMAND", "backup")
	t.Setenv("STANDARD_BACKUPS_PATHS", "/bogus")
	t.Setenv("STANDARD_BACKUPS_OPTIONS", "{}")
	expectedErr := errors.New("oops")
	b := &BackendImpl{
		Backup: func(req *BackupRequest) error {
			return expectedErr
		},
	}
	err := b.execute()
	assert.ErrorIs(t, err, expectedErr)
}

func TestBackendBackupNotImplemented(t *testing.T) {
	t.Setenv("STANDARD_BACKUPS_COMMAND", "backup")
	b := &BackendImpl{}
	err := b.execute()
	assert.EqualError(t, err, "unhandled command backup")
}
