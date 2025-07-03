package backend

import (
	"errors"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestBackendCallBackup(t *testing.T) {
	t.Setenv("STANDARD_BACKUPS_COMMAND", "backup")
	t.Setenv("STANDARD_BACKUPS_PATHS", "/foo:/bar")
	t.Setenv("STANDARD_BACKUPS_OPTIONS", testutils.Dedent(`
		{
			"key": "value",
			"bool": true,
			"number": 42,
			"array": [1,2,3],
			"object": { "foo": "bar" }
		}
	`))
	var (
		paths   []string
		options map[string]any
		called  bool
	)
	b := &Backend{
		Backup: func(p []string, o map[string]any) error {
			paths = p
			options = o
			called = true
			return nil
		},
	}
	err := b.execute()
	if assert.NoError(t, err) {
		assert.True(t, called, "Backup func was not called")
		assert.Equal(t, []string{"/foo", "/bar"}, paths)
		assert.Equal(t, map[string]any{
			"key":    "value",
			"bool":   true,
			"number": 42.0,
			"array":  []any{1.0, 2.0, 3.0},
			"object": map[string]any{"foo": "bar"},
		}, options)
	}
}

func TestBackendBackupError(t *testing.T) {
	t.Setenv("STANDARD_BACKUPS_COMMAND", "backup")
	t.Setenv("STANDARD_BACKUPS_PATHS", "/bogus")
	t.Setenv("STANDARD_BACKUPS_OPTIONS", "{}")
	expectedErr := errors.New("oops")
	b := &Backend{
		Backup: func(p []string, o map[string]any) error {
			return expectedErr
		},
	}
	err := b.execute()
	assert.ErrorIs(t, err, expectedErr)
}

func TestBackendBackupNotImplemented(t *testing.T) {
	t.Setenv("STANDARD_BACKUPS_COMMAND", "backup")
	b := &Backend{}
	err := b.execute()
	assert.EqualError(t, err, "unhandled command backup")
}

func TestBackendUnknownCommand(t *testing.T) {
	t.Setenv("STANDARD_BACKUPS_COMMAND", "bogus")
	b := &Backend{}
	err := b.execute()
	assert.EqualError(t, err, "unknown command bogus")
}
