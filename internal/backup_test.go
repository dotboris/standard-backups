package internal

import (
	"errors"
	"fmt"
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

func TestBackupSingleSimple(t *testing.T) {
	fac := NewMocknewBackendClienter(t)
	fac.EXPECT().NewBackendClient(mock.Anything, "the-backend").
		RunAndReturn(func(c config.Config, s string) (backuper, error) {
			client := NewMockbackuper(t)
			client.EXPECT().Backup(
				&proto.BackupRequest{
					Paths:           []string{"path1", "path2"},
					DestinationName: "dest",
					JobName:         "my-job",
					RawOptions: map[string]any{
						"foo": "bar",
						"biz": 42,
					},
				},
			).Return(nil)
			return client, nil
		})
	svc := backupService{backendClientFactory: fac}

	err := svc.Backup(
		config.Config{
			Recipes: []config.RecipeManifestV1{{
				Name:  "back-me-up",
				Paths: []string{"path1", "path2"},
			}},
			MainConfig: config.MainConfig{
				Destinations: map[string]config.DestinationConfigV1{
					"dest": {
						Backend: "the-backend",
						Options: map[string]any{
							"foo": "bar",
							"biz": 42,
						},
					},
				},
				Jobs: map[string]config.JobConfigV1{
					"my-job": {
						Recipe:   "back-me-up",
						BackupTo: []string{"dest"},
					},
				},
			},
		},
		"my-job",
	)
	assert.NoError(t, err)
}

func TestBackupSingleBackupError(t *testing.T) {
	expectedErr := errors.New("oops")
	fac := NewMocknewBackendClienter(t)
	fac.EXPECT().NewBackendClient(mock.Anything, "the-backend").
		RunAndReturn(func(c config.Config, s string) (backuper, error) {
			client := NewMockbackuper(t)
			client.EXPECT().Backup(mock.Anything).Return(expectedErr)
			return client, nil
		})
	svc := backupService{backendClientFactory: fac}

	err := svc.Backup(
		config.Config{
			Recipes: []config.RecipeManifestV1{{
				Name: "bogus",
			}},
			MainConfig: config.MainConfig{
				Destinations: map[string]config.DestinationConfigV1{
					"bogus": {
						Backend: "the-backend",
					},
				},
				Jobs: map[string]config.JobConfigV1{
					"do-it": {
						Recipe:   "bogus",
						BackupTo: []string{"bogus"},
					},
				},
			},
		},
		"do-it",
	)
	assert.ErrorIs(t, err, expectedErr)
}

func TestBackupSingleHooksSuccess(t *testing.T) {
	fac := NewMocknewBackendClienter(t)
	for _, name := range []string{"b1", "b2"} {
		fac.EXPECT().NewBackendClient(mock.Anything, name).
			RunAndReturn(func(c config.Config, s string) (backuper, error) {
				client := NewMockbackuper(t)
				client.EXPECT().Backup(mock.Anything).Return(nil)
				return client, nil
			})
	}
	svc := backupService{backendClientFactory: fac}

	d := t.TempDir()
	hooksLog := path.Join(d, "hooks.log")

	err := svc.Backup(
		config.Config{
			Recipes: []config.RecipeManifestV1{{
				Name: "r",
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
			}},
			MainConfig: config.MainConfig{
				Destinations: map[string]config.DestinationConfigV1{
					"d1": {
						Backend: "b1",
					},
					"d2": {
						Backend: "b2",
					},
				},
				Jobs: map[string]config.JobConfigV1{
					"do-it": {
						Recipe:   "r",
						BackupTo: []string{"d1", "d2"},
					},
				},
			},
		},
		"do-it",
	)

	if assert.NoError(t, err) {
		log, err := os.ReadFile(hooksLog)
		if assert.NoError(t, err) {
			assert.Equal(t,
				testutils.Dedent(`
					before
					after
					on-success
				`),
				strings.Trim(string(log), "\n"),
			)
		}
	}
}

func TestBackupSingleHooksFailure(t *testing.T) {
	fac := NewMocknewBackendClienter(t)
	for _, name := range []string{"b1", "b2"} {
		fac.EXPECT().NewBackendClient(mock.Anything, name).
			RunAndReturn(func(c config.Config, s string) (backuper, error) {
				client := NewMockbackuper(t)
				client.EXPECT().Backup(mock.Anything).Return(errors.New("oops"))
				return client, nil
			})
	}
	svc := backupService{backendClientFactory: fac}

	d := t.TempDir()
	hooksLog := path.Join(d, "hooks.log")

	err := svc.Backup(
		config.Config{
			Recipes: []config.RecipeManifestV1{{
				Name: "r",
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
			}},
			MainConfig: config.MainConfig{
				Destinations: map[string]config.DestinationConfigV1{
					"d1": {
						Backend: "b1",
					},
					"d2": {
						Backend: "b2",
					},
				},
				Jobs: map[string]config.JobConfigV1{
					"do-it": {
						Recipe:   "r",
						BackupTo: []string{"d1", "d2"},
					},
				},
			},
		},
		"do-it",
	)

	assert.Error(t, err)
	log, err := os.ReadFile(hooksLog)
	if assert.NoError(t, err) {
		assert.Equal(t,
			testutils.Dedent(`
				before
				after
				on-failure
			`),
			strings.Trim(string(log), "\n"),
		)
	}
}

func TestBackupSingleBeforeHookError(t *testing.T) {
	fac := NewMocknewBackendClienter(t)
	svc := backupService{backendClientFactory: fac}

	err := svc.Backup(
		config.Config{
			Recipes: []config.RecipeManifestV1{{
				Name: "r",
				Hooks: config.HooksV1{
					Before: &config.HookV1{
						Shell:   "bash",
						Command: "exit 42",
					},
				},
			}},
			MainConfig: config.MainConfig{
				Destinations: map[string]config.DestinationConfigV1{
					"bogus": {
						Backend: "the-backend",
					},
				},
				Jobs: map[string]config.JobConfigV1{
					"do-it": {
						Recipe:   "r",
						BackupTo: []string{"bogus"},
					},
				},
			},
		},
		"do-it",
	)

	var exitError *exec.ExitError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &exitError) {
		assert.Equal(t, exitError.ExitCode(), 42)
	}
}

func TestBackupSingleAfterHookError(t *testing.T) {
	fac := NewMocknewBackendClienter(t)
	fac.EXPECT().NewBackendClient(mock.Anything, "the-backend").
		RunAndReturn(func(c config.Config, s string) (backuper, error) {
			client := NewMockbackuper(t)
			client.EXPECT().Backup(mock.Anything).Return(nil)
			return client, nil
		})
	svc := backupService{backendClientFactory: fac}

	err := svc.Backup(
		config.Config{
			Recipes: []config.RecipeManifestV1{{
				Name: "r",
				Hooks: config.HooksV1{
					After: &config.HookV1{
						Shell:   "bash",
						Command: "exit 42",
					},
				},
			}},
			MainConfig: config.MainConfig{
				Destinations: map[string]config.DestinationConfigV1{
					"bogus": {
						Backend: "the-backend",
					},
				},
				Jobs: map[string]config.JobConfigV1{
					"do-it": {
						Recipe:   "r",
						BackupTo: []string{"bogus"},
					},
				},
			},
		},
		"do-it",
	)

	var exitError *exec.ExitError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &exitError) {
		assert.Equal(t, exitError.ExitCode(), 42)
	}
}

func TestBackupSingleOnSuccessHookError(t *testing.T) {
	fac := NewMocknewBackendClienter(t)
	fac.EXPECT().NewBackendClient(mock.Anything, "the-backend").
		RunAndReturn(func(c config.Config, s string) (backuper, error) {
			client := NewMockbackuper(t)
			client.EXPECT().Backup(mock.Anything).Return(nil)
			return client, nil
		})
	svc := backupService{backendClientFactory: fac}

	err := svc.Backup(
		config.Config{
			Recipes: []config.RecipeManifestV1{{
				Name: "r",
				Hooks: config.HooksV1{
					OnSuccess: &config.HookV1{
						Shell:   "bash",
						Command: "exit 42",
					},
				},
			}},
			MainConfig: config.MainConfig{
				Destinations: map[string]config.DestinationConfigV1{
					"bogus": {
						Backend: "the-backend",
					},
				},
				Jobs: map[string]config.JobConfigV1{
					"do-it": {
						Recipe:   "r",
						BackupTo: []string{"bogus"},
					},
				},
			},
		},
		"do-it",
	)

	var exitError *exec.ExitError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &exitError) {
		assert.Equal(t, exitError.ExitCode(), 42)
	}
}

func TestBackupSingleOnFailureHookError(t *testing.T) {
	fac := NewMocknewBackendClienter(t)
	fac.EXPECT().NewBackendClient(mock.Anything, "the-backend").
		RunAndReturn(func(c config.Config, s string) (backuper, error) {
			client := NewMockbackuper(t)
			client.EXPECT().Backup(mock.Anything).Return(errors.New("oops"))
			return client, nil
		})
	svc := backupService{backendClientFactory: fac}

	err := svc.Backup(
		config.Config{
			Recipes: []config.RecipeManifestV1{{
				Name: "r",
				Hooks: config.HooksV1{
					OnFailure: &config.HookV1{
						Shell:   "bash",
						Command: "exit 42",
					},
				},
			}},
			MainConfig: config.MainConfig{
				Destinations: map[string]config.DestinationConfigV1{
					"bogus": {
						Backend: "the-backend",
					},
				},
				Jobs: map[string]config.JobConfigV1{
					"do-it": {
						Recipe:   "r",
						BackupTo: []string{"bogus"},
					},
				},
			},
		},
		"do-it",
	)

	var exitError *exec.ExitError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &exitError) {
		assert.Equal(t, exitError.ExitCode(), 42)
	}
}

func TestBackupSingleBackupAndHooksError(t *testing.T) {
	fac := NewMocknewBackendClienter(t)
	fac.EXPECT().NewBackendClient(mock.Anything, "the-backend").
		RunAndReturn(func(c config.Config, s string) (backuper, error) {
			client := NewMockbackuper(t)
			client.EXPECT().Backup(mock.Anything).Return(errors.New("oops"))
			return client, nil
		})
	svc := backupService{backendClientFactory: fac}

	err := svc.Backup(
		config.Config{
			Recipes: []config.RecipeManifestV1{{
				Name: "r",
				Hooks: config.HooksV1{
					// We can't have Before fail otherwise, backup doesn't get performed
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
			}},
			MainConfig: config.MainConfig{
				Destinations: map[string]config.DestinationConfigV1{
					"bogus": {
						Backend: "the-backend",
					},
				},
				Jobs: map[string]config.JobConfigV1{
					"do-it": {
						Recipe:   "r",
						BackupTo: []string{"bogus"},
					},
				},
			},
		},
		"do-it",
	)

	assert.EqualError(t, err, testutils.Dedent(`
		backup failed: oops
		after hook failed: exit status 43
		on-failure hook failed: exit status 45
	`))
}

func TestBackupSingleOnlyHooksError(t *testing.T) {
	fac := NewMocknewBackendClienter(t)
	svc := backupService{backendClientFactory: fac}

	err := svc.Backup(
		config.Config{
			Recipes: []config.RecipeManifestV1{{
				Name: "r",
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
			}},
			MainConfig: config.MainConfig{
				Destinations: map[string]config.DestinationConfigV1{
					"bogus": {
						Backend: "the-backend",
					},
				},
				Jobs: map[string]config.JobConfigV1{
					"do-it": {
						Recipe:   "r",
						BackupTo: []string{"bogus"},
					},
				},
			},
		},
		"do-it",
	)

	assert.EqualError(t, err, testutils.Dedent(`
		before hook failed: exit status 42
		after hook failed: exit status 43
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
			fac := NewMocknewBackendClienter(t)
			fac.EXPECT().NewBackendClient(mock.Anything, "the-backend").
				RunAndReturn(func(c config.Config, s string) (backuper, error) {
					client := NewMockbackuper(t)
					client.EXPECT().Backup(mock.Anything).Maybe().Return(nil)
					return client, nil
				}).Maybe()
			svc := backupService{backendClientFactory: fac}

			d := t.TempDir()
			outFile := path.Join(d, "out.txt")
			test.hooks.OnFailure = &config.HookV1{
				Shell:   "bash",
				Command: fmt.Sprintf("echo hello from on-failure > %s", outFile),
			}

			err := svc.Backup(
				config.Config{
					Recipes: []config.RecipeManifestV1{{
						Name:  "r",
						Hooks: test.hooks,
					}},
					MainConfig: config.MainConfig{
						Destinations: map[string]config.DestinationConfigV1{
							"bogus": {
								Backend: "the-backend",
							},
						},
						Jobs: map[string]config.JobConfigV1{
							"do-it": {
								Recipe:   "r",
								BackupTo: []string{"bogus"},
							},
						},
					},
				},
				"do-it",
			)

			assert.Error(t, err)
			output, err := os.ReadFile(outFile)
			if assert.NoError(t, err) {
				assert.Equal(t, string(output), "hello from on-failure\n")
			}
		})
	}
}
