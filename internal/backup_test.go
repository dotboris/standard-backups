package internal

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/dotboris/standard-backups/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() (bytes.Buffer, *slog.Logger) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	return buf, logger
}

func TestBackupSingleSimple(t *testing.T) {
	b := NewMockbackupBackend(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(
		[]string{"path1", "path2"},
		map[string]any{
			"foo": "bar",
			"biz": 42,
		},
	).Return(nil)

	_, logger := newTestLogger()
	err := backupSingle(
		logger,
		&config.RecipeManifestV1{
			Paths: []string{"path1", "path2"},
		},
		config.DestinationConfigV1{
			Options: map[string]any{
				"foo": "bar",
				"biz": 42,
			},
		},
		b,
	)

	assert.NoError(t, err)
}

func TestBackupSingleSkip(t *testing.T) {
	b := NewMockbackupBackend(t)
	b.EXPECT().Enabled().Return(false)

	_, logger := newTestLogger()
	err := backupSingle(
		logger,
		&config.RecipeManifestV1{
			Paths: []string{"path1", "path2"},
		},
		config.DestinationConfigV1{
			Options: map[string]any{
				"foo": "bar",
				"biz": 42,
			},
		},
		b,
	)

	if assert.NoError(t, err) {
		b.AssertNotCalled(t, "Backup")
	}
}

func TestBackupSingleBackupError(t *testing.T) {
	expectedErr := errors.New("oops")
	b := NewMockbackupBackend(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(mock.Anything, mock.Anything).Return(expectedErr)

	_, logger := newTestLogger()
	err := backupSingle(
		logger,
		&config.RecipeManifestV1{},
		config.DestinationConfigV1{},
		b,
	)

	assert.ErrorIs(t, err, expectedErr)
}
