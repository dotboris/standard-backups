package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resticCmd(t *testing.T, repo string, secret string, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command("restic", "-r", repo)
	cmd.Args = append(cmd.Args, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("RESTIC_PASSWORD=%s", secret))
	cmd.Dir = testutils.GetRepoRoot(t)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func resticGetRepoId(t *testing.T, repo string, secret string) (string, error) {
	t.Helper()
	cmd := resticCmd(t, repo, secret, "cat", "config")
	cmd.Stdout = nil
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	var output struct {
		Id string `json:"id"`
	}
	err = json.Unmarshal(stdout, &output)
	if err != nil {
		return "", err
	}
	return output.Id, nil
}

func resticListSnapshots(t *testing.T, repo string, secret string) []map[string]any {
	t.Helper()
	cmd := resticCmd(t, repo, secret, "snapshots", "--json")
	cmd.Stdout = nil
	stdout, err := cmd.Output()
	require.NoError(t, err)
	var output []map[string]any
	err = json.Unmarshal(stdout, &output)
	require.NoError(t, err)
	return output
}

func TestResticBackupBase(t *testing.T) {
	repoDir := t.TempDir()

	tc := testutils.NewTestConfig(t)
	tc.AddBackend("restic", "dist/standard-backups-restic-backend")
	sourceDir := t.TempDir()
	err := os.WriteFile(
		path.Join(sourceDir, "back-me-up.txt"),
		[]byte("back me up"),
		0o644,
	)
	require.NoError(t, err)
	err = os.WriteFile(
		path.Join(sourceDir, "not-me.txt"),
		[]byte("can't touch this"),
		0o644,
	)
	require.NoError(t, err)
	tc.AddRecipe("bogus", testutils.DedentYaml(fmt.Sprintf(`
		version: 1
		name: bogus
		paths:
			- %s
		exclude:
			- 'not-me.txt'
	`, sourceDir)))
	tc.WriteConfig(testutils.DedentYaml(fmt.Sprintf(`
		version: 1
		secrets:
			pass:
				literal: supersecret
		destinations:
			my-dest:
				backend: restic
				options:
					repo: %s
					env:
						RESTIC_PASSWORD: '{{ .Secrets.pass }}'
		jobs:
			my-job:
				recipe: bogus
				backup-to: [my-dest]
	`, repoDir)))

	cmd := testutils.StandardBackups(t, "backup", "my-job")
	tc.Apply(cmd)
	err = cmd.Run()
	require.NoError(t, err)

	// Check that repo has been initialized
	err = resticCmd(t, repoDir, "supersecret", "cat", "config").Run()
	require.NoErrorf(t, err, "restic repo %s has not been initialized", repoDir)

	// Check that we can list the backup
	cmd = testutils.StandardBackups(t, "list-backups", "my-dest", "--json")
	tc.Apply(cmd)
	cmd.Stdout = nil
	stdout, err := cmd.Output()
	require.NoError(t, err, "failed to list backups")
	var output []proto.ListBackupsResponseItem
	err = json.Unmarshal(stdout, &output)
	require.NoError(t, err)
	assert.Len(t, output, 1)

	// Test restore
	restoreDir := t.TempDir()
	cmd = testutils.StandardBackups(t, "restore", "my-dest", output[0].Id, restoreDir)
	tc.Apply(cmd)
	err = cmd.Run()
	require.NoError(t, err)
	restoredFile, err := os.ReadFile(path.Join(restoreDir, sourceDir, "back-me-up.txt"))
	require.NoError(t, err)
	assert.Equal(t, "back me up", string(restoredFile))
	_, err = os.Stat(path.Join(restoreDir, "not-me.txt"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestResticBackupPreservesExistingRepo(t *testing.T) {
	repoDir := t.TempDir()

	// Create a repo ourselves
	err := resticCmd(t, repoDir, "supersecret", "-v", "init").Run()
	require.NoError(t, err)

	expectedRepoId, err := resticGetRepoId(t, repoDir, "supersecret")
	require.NoError(t, err)
	assert.NotZero(t, expectedRepoId)

	tc := testutils.NewTestConfig(t)
	tc.AddBackend("restic", "dist/standard-backups-restic-backend")
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(fmt.Sprintf(`
		version: 1
		secrets:
			pass:
				literal: supersecret
		destinations:
			my-dest:
				backend: restic
				options:
					repo: %s
					env:
						RESTIC_PASSWORD: '{{ .Secrets.pass }}'
		jobs:
			my-job:
				recipe: bogus
				backup-to: [my-dest]
	`, repoDir)))

	cmd := testutils.StandardBackups(t, "backup", "my-job")
	tc.Apply(cmd)
	err = cmd.Run()
	require.NoError(t, err)

	actualRepoId, err := resticGetRepoId(t, repoDir, "supersecret")
	require.NoError(t, err)
	assert.NotZero(t, actualRepoId)
	assert.Equal(t, expectedRepoId, actualRepoId)
}

func TestResticBackupForget(t *testing.T) {
	simpleRepoDir := t.TempDir()
	varsRepoDir := t.TempDir()

	tc := testutils.NewTestConfig(t)
	tc.AddBackend("restic", "dist/standard-backups-restic-backend")
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(fmt.Sprintf(`
		version: 1
		secrets:
			pass:
				literal: supersecret
		destinations:
			simple:
				backend: restic
				options:
					repo: %s
					forget:
						enable: true
						options:
							group-by: ''
							keep-last: 1
					env:
						RESTIC_PASSWORD: '{{ .Secrets.pass }}'
			vars:
				backend: restic
				options:
					repo: %s
					forget:
						enable: true
						options:
							group-by: ''
					env:
						RESTIC_PASSWORD: '{{ .Secrets.pass }}'
				variants:
					a:
						forget:
							options:
								keep-last: 1
					b:
						forget:
							options:
								keep-last: 2
		jobs:
			my-job:
				recipe: bogus
				backup-to: [simple, vars/a, vars/b]
	`, simpleRepoDir, varsRepoDir)))

	for range 3 {
		cmd := testutils.StandardBackups(t, "backup", "my-job")
		tc.Apply(cmd)
		err := cmd.Run()
		require.NoError(t, err)
	}

	snapshots := resticListSnapshots(t, simpleRepoDir, "supersecret")
	assert.Len(t, snapshots, 1)

	snapshots = resticListSnapshots(t, varsRepoDir, "supersecret")
	assert.Len(t, snapshots, 3)

	var aCount, bCount int
	for _, snapshot := range snapshots {
		tags := snapshot["tags"].([]any)
		fmt.Printf("snapshot %v tags %v\n", snapshot["id"], tags)
		for _, rawTag := range tags {
			tag := rawTag.(string)
			if tag == "sb:variant:a" {
				aCount += 1
			}
			if tag == "sb:variant:b" {
				bCount += 1
			}
		}
	}
	assert.Equal(t, 1, aCount)
	assert.Equal(t, 2, bCount)
}

func TestResticExec(t *testing.T) {
	repoDir := t.TempDir()

	tc := testutils.NewTestConfig(t)
	tc.AddBackend("restic", "dist/standard-backups-restic-backend")
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(fmt.Sprintf(`
		version: 1
		secrets:
			pass:
				literal: supersecret
		destinations:
			my-dest:
				backend: restic
				options:
					repo: %s
					env:
						RESTIC_PASSWORD: '{{ .Secrets.pass }}'
		jobs:
			my-job:
				recipe: bogus
				backup-to: [my-dest]
	`, repoDir)))

	cmd := testutils.StandardBackups(t, "backup", "my-job")
	tc.Apply(cmd)
	err := cmd.Run()
	require.NoError(t, err, "failed to backup")

	expectedSnapshots := resticListSnapshots(t, repoDir, "supersecret")

	// Check that exec returns the same thing
	cmd = testutils.StandardBackups(t, "exec", "-d", "my-dest")
	tc.Apply(cmd)
	cmd.Args = append(cmd.Args, "--", "snapshots", "--json")
	cmd.Stdout = nil
	stdout, err := cmd.Output()
	require.NoError(t, err, "failed to list expected snapshots with exec in %s", repoDir)
	var actualSnapshots []map[string]any
	err = json.Unmarshal(stdout, &actualSnapshots)
	require.NoError(t, err)
	assert.Equal(t, expectedSnapshots, actualSnapshots)
}

func TestResticListBackups(t *testing.T) {
	tc := testutils.NewTestConfig(t)
	tc.AddBackend("restic", "dist/standard-backups-restic-backend")
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(fmt.Sprintf(`
		version: 1
		secrets:
			pass:
				literal: supersecret
		destinations:
			simple:
				backend: restic
				options:
					repo: %s
					env:
						RESTIC_PASSWORD: '{{ .Secrets.pass }}'
			vars:
				backend: restic
				options:
					repo: %s
					env:
						RESTIC_PASSWORD: '{{ .Secrets.pass }}'
				variants:
					a: {}
					b: {}
		jobs:
			my-job:
				recipe: bogus
				backup-to: [simple, vars/a, vars/b]
	`, t.TempDir(), t.TempDir())))

	for range 2 {
		cmd := testutils.StandardBackups(t, "backup", "my-job")
		tc.Apply(cmd)
		err := cmd.Run()
		require.NoError(t, err, "failed to backup")
	}

	checks := []struct {
		destination string
		variant     string
		ref         string
	}{
		{
			destination: "simple",
			variant:     "",
			ref:         "simple",
		},
		{
			destination: "vars",
			variant:     "a",
			ref:         "vars/a",
		},
		{
			destination: "vars",
			variant:     "b",
			ref:         "vars/b",
		},
	}
	for _, check := range checks {
		t.Run(check.ref, func(t *testing.T) {
			cmd := testutils.StandardBackups(t, "list-backups", check.ref, "--json")
			tc.Apply(cmd)
			cmd.Stdout = nil
			stdout, err := cmd.Output()
			require.NoError(t, err)
			var output []proto.ListBackupsResponseItem
			err = json.Unmarshal(stdout, &output)
			require.NoError(t, err)

			assert.Len(t, output, 2)
			for i := range 2 {
				backupTime, err := time.Parse(time.RFC3339, output[i].Time)
				assert.NoError(t, err, "failed to parse output[%d].Time in %s", i, check.ref)
				assert.WithinRange(t, backupTime,
					backupTime.Add(time.Minute*-2),
					backupTime.Add(time.Minute*2))

				assert.NotEmpty(t, output[i].Id, i)
				assert.NotEmpty(t, output[i].Extra, i)
				assert.Greater(t, output[i].Size, 0, i)
				assert.Equal(t, "my-job", output[i].Job, i)
				assert.Equal(t, check.destination, output[i].Destination, i)
				assert.Equal(t, check.variant, output[i].Variant, i)
			}

			assert.NotEqual(t, output[0].Id, output[1].Id)
		})
	}
}
