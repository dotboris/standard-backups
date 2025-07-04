package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/dotboris/standard-backups/internal/config"
	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func TestBackupSingleSimple(t *testing.T) {
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(
		&proto.BackupRequest{
			Paths: []string{"path1", "path2"},
			RawOptions: map[string]any{
				"foo": "bar",
				"biz": 42,
			},
		},
	).Return(nil)

	err := backupSingle(
		newTestLogger(),
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
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(false)

	err := backupSingle(
		newTestLogger(),
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
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(mock.Anything).Return(expectedErr)

	err := backupSingle(
		newTestLogger(),
		&config.RecipeManifestV1{},
		config.DestinationConfigV1{},
		b,
	)

	assert.ErrorIs(t, err, expectedErr)
}

func TestBackupSingleHooksSuccess(t *testing.T) {
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(mock.Anything).Return(nil)

	d := t.TempDir()
	hooksLog := path.Join(d, "hooks.log")

	err := backupSingle(
		newTestLogger(),
		&config.RecipeManifestV1{
			Hooks: config.HooksV1{
				Before: &config.HookV1{
					Shell:   "bash",
					Command: fmt.Sprintf("echo before >> %s", hooksLog),
				},
				After: &config.HookV1{
					Shell:   "bash",
					Command: fmt.Sprintf("echo after >> %s", hooksLog),
				},
				OnSuccess: &config.HookV1{
					Shell:   "bash",
					Command: fmt.Sprintf("echo on-success >> %s", hooksLog),
				},
				OnFailure: &config.HookV1{
					Shell:   "bash",
					Command: fmt.Sprintf("echo on-failure >> %s", hooksLog),
				},
			},
		},
		config.DestinationConfigV1{},
		b,
	)

	if assert.NoError(t, err) {
		log, err := os.ReadFile(hooksLog)
		if assert.NoError(t, err) {
			assert.Equal(t,
				strings.Trim(string(log), "\n"),
				testutils.Dedent(`
					before
					after
					on-success
				`),
			)
		}
	}
}

func TestBackupSingleHooksFailure(t *testing.T) {
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(mock.Anything).Return(errors.New("oops"))

	d := t.TempDir()
	hooksLog := path.Join(d, "hooks.log")

	err := backupSingle(
		newTestLogger(),
		&config.RecipeManifestV1{
			Hooks: config.HooksV1{
				Before: &config.HookV1{
					Shell:   "bash",
					Command: fmt.Sprintf("echo before >> %s", hooksLog),
				},
				After: &config.HookV1{
					Shell:   "bash",
					Command: fmt.Sprintf("echo after >> %s", hooksLog),
				},
				OnSuccess: &config.HookV1{
					Shell:   "bash",
					Command: fmt.Sprintf("echo on-success >> %s", hooksLog),
				},
				OnFailure: &config.HookV1{
					Shell:   "bash",
					Command: fmt.Sprintf("echo on-failure >> %s", hooksLog),
				},
			},
		},
		config.DestinationConfigV1{},
		b,
	)

	assert.Error(t, err)
	log, err := os.ReadFile(hooksLog)
	if assert.NoError(t, err) {
		assert.Equal(t,
			strings.Trim(string(log), "\n"),
			testutils.Dedent(`
				before
				after
				on-failure
			`),
		)
	}
}

func TestBackupSingleBeforeHookError(t *testing.T) {
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(true)

	err := backupSingle(
		newTestLogger(),
		&config.RecipeManifestV1{
			Hooks: config.HooksV1{
				Before: &config.HookV1{
					Shell:   "bash",
					Command: "exit 42",
				},
			},
		},
		config.DestinationConfigV1{},
		b,
	)

	var exitError *exec.ExitError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &exitError) {
		assert.Equal(t, exitError.ExitCode(), 42)
		b.AssertNotCalled(t, "Backup")
	}
}

func TestBackupSingleAfterHookError(t *testing.T) {
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(mock.Anything).Return(nil)

	err := backupSingle(
		newTestLogger(),
		&config.RecipeManifestV1{
			Hooks: config.HooksV1{
				After: &config.HookV1{
					Shell:   "bash",
					Command: "exit 42",
				},
			},
		},
		config.DestinationConfigV1{},
		b,
	)

	var exitError *exec.ExitError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &exitError) {
		assert.Equal(t, exitError.ExitCode(), 42)
	}
}

func TestBackupSingleOnSuccessHookError(t *testing.T) {
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(mock.Anything).Return(nil)

	err := backupSingle(
		newTestLogger(),
		&config.RecipeManifestV1{
			Hooks: config.HooksV1{
				OnSuccess: &config.HookV1{
					Shell:   "bash",
					Command: "exit 42",
				},
			},
		},
		config.DestinationConfigV1{},
		b,
	)

	var exitError *exec.ExitError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &exitError) {
		assert.Equal(t, exitError.ExitCode(), 42)
	}
}

func TestBackupSingleOnFailureHookError(t *testing.T) {
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(mock.Anything).Return(errors.New("oops"))

	err := backupSingle(
		newTestLogger(),
		&config.RecipeManifestV1{
			Hooks: config.HooksV1{
				OnFailure: &config.HookV1{
					Shell:   "bash",
					Command: "exit 42",
				},
			},
		},
		config.DestinationConfigV1{},
		b,
	)

	var exitError *exec.ExitError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &exitError) {
		assert.Equal(t, exitError.ExitCode(), 42)
	}
}

func TestBackupSingleBackupAndHooksError(t *testing.T) {
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(mock.Anything).Return(errors.New("oops"))

	err := backupSingle(
		newTestLogger(),
		&config.RecipeManifestV1{
			Hooks: config.HooksV1{
				After: &config.HookV1{
					Shell:   "bash",
					Command: "exit 43",
				},
				OnSuccess: &config.HookV1{
					Shell:   "bash",
					Command: "exit 44",
				},
				OnFailure: &config.HookV1{
					Shell:   "bash",
					Command: "exit 45",
				},
			},
		},
		config.DestinationConfigV1{},
		b,
	)

	assert.EqualError(t, err, testutils.Dedent(`
		backup failed: oops
		after hook failed: exit status 43
		on-failure hook failed: exit status 45
	`))
}

func TestBackupSingleOnlyHooksError(t *testing.T) {
	b := NewMockbackupClient(t)
	b.EXPECT().Enabled().Return(true)
	b.EXPECT().Backup(mock.Anything).Maybe().Return(nil)

	err := backupSingle(
		newTestLogger(),
		&config.RecipeManifestV1{
			Hooks: config.HooksV1{
				Before: &config.HookV1{
					Shell:   "bash",
					Command: "exit 42",
				},
				After: &config.HookV1{
					Shell:   "bash",
					Command: "exit 43",
				},
				OnSuccess: &config.HookV1{
					Shell:   "bash",
					Command: "exit 44",
				},
				OnFailure: &config.HookV1{
					Shell:   "bash",
					Command: "exit 45",
				},
			},
		},
		config.DestinationConfigV1{},
		b,
	)

	assert.EqualError(t, err, testutils.Dedent(`
		before hook failed: exit status 42
		on-failure hook failed: exit status 45
	`))
}

func TestBackupSingleOnFailureCalledOnError(t *testing.T) {
	tests := []struct {
		name  string
		hooks config.HooksV1
	}{
		{
			name: "before",
			hooks: config.HooksV1{
				Before: &config.HookV1{
					Shell:   "bash",
					Command: "exit 1",
				},
			},
		},
		{
			name: "after",
			hooks: config.HooksV1{
				After: &config.HookV1{
					Shell:   "bash",
					Command: "exit 1",
				},
			},
		},
		{
			name: "on-success",
			hooks: config.HooksV1{
				OnSuccess: &config.HookV1{
					Shell:   "bash",
					Command: "exit 1",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := NewMockbackupClient(t)
			b.EXPECT().Enabled().Return(true)
			b.EXPECT().Backup(mock.Anything).Maybe().Return(nil)

			d := t.TempDir()
			outFile := path.Join(d, "out.txt")
			test.hooks.OnFailure = &config.HookV1{
				Shell:   "bash",
				Command: fmt.Sprintf("echo hello from on-failure > %s", outFile),
			}

			err := backupSingle(
				newTestLogger(),
				&config.RecipeManifestV1{Hooks: test.hooks},
				config.DestinationConfigV1{},
				b,
			)

			assert.Error(t, err)
			output, err := os.ReadFile(outFile)
			if assert.NoError(t, err) {
				assert.Equal(t, string(output), "hello from on-failure\n")
			}
		})
	}
}
