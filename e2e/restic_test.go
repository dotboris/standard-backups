package e2e

import (
	"bytes"
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
)

func resticGetRepoId(t *testing.T, repo string, secret string) (string, error) {
	cmd := exec.Command("restic", "-r", repo, "cat", "config")
	cmd.Env = append(os.Environ(), fmt.Sprintf("RESTIC_PASSWORD=%s", secret))
	cmd.Dir = testutils.GetRepoRoot(t)
	cmd.Stderr = os.Stderr
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
	if !assert.NoError(t, err) {
		return
	}
	err = os.WriteFile(
		path.Join(sourceDir, "not-me.txt"),
		[]byte("can't touch this"),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
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
	if !assert.NoError(t, err) {
		return
	}

	// Check that repo has been initialized
	cmd = exec.Command("restic", "-r", repoDir, "cat", "config")
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD=supersecret")
	cmd.Dir = testutils.GetRepoRoot(t)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if !assert.NoError(t, err, "restic repo %s has not been initialized", repoDir) {
		return
	}

	// Check that we can list the backup
	cmd = testutils.StandardBackups(t, "list-backups", "my-dest", "--json")
	tc.Apply(cmd)
	stdout := bytes.NewBuffer(nil)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if !assert.NoError(t, err, "failed to list backups") {
		return
	}
	var output []map[string]any
	err = json.Unmarshal(stdout.Bytes(), &output)
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, output, 1)
	id, ok := output[0]["id"]
	if !assert.True(t, ok) {
		return
	}

	// Test restore
	restoreDir := t.TempDir()
	cmd = exec.Command("restic",
		"-v",
		"-r", repoDir,
		"restore", fmt.Sprintf("%s:%s", id, sourceDir),
		"--target", restoreDir,
	)
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD=supersecret")
	cmd.Dir = testutils.GetRepoRoot(t)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if !assert.NoError(t, err) {
		return
	}
	restoredFile, err := os.ReadFile(path.Join(restoreDir, "back-me-up.txt"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "back me up", string(restoredFile))
	_, err = os.Stat(path.Join(restoreDir, "not-me.txt"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestResticBackupPreservesExistingRepo(t *testing.T) {
	repoDir := t.TempDir()

	// Create a repo ourselves
	cmd := exec.Command("restic", "-v", "-r", repoDir, "init")
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD=supersecret")
	cmd.Dir = testutils.GetRepoRoot(t)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if !assert.NoError(t, err) {
		return
	}

	expectedRepoId, err := resticGetRepoId(t, repoDir, "supersecret")
	if !assert.NoError(t, err) {
		return
	}
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

	cmd = testutils.StandardBackups(t, "backup", "my-job")
	tc.Apply(cmd)
	err = cmd.Run()
	if !assert.NoError(t, err) {
		return
	}

	actualRepoId, err := resticGetRepoId(t, repoDir, "supersecret")
	if !assert.NoError(t, err) {
		return
	}
	assert.NotZero(t, actualRepoId)
	assert.Equal(t, expectedRepoId, actualRepoId)
}

func TestResticBackupForget(t *testing.T) {
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
					forget:
						enable: true
						options:
							keep-last: 1
					env:
						RESTIC_PASSWORD: '{{ .Secrets.pass }}'
		jobs:
			my-job:
				recipe: bogus
				backup-to: [my-dest]
	`, repoDir)))

	for range 2 {
		cmd := testutils.StandardBackups(t, "backup", "my-job")
		tc.Apply(cmd)
		err := cmd.Run()
		if !assert.NoError(t, err) {
			return
		}

		// Count number of snapshots
		cmd = exec.Command("restic", "-r", repoDir, "snapshots", "--json")
		cmd.Env = append(os.Environ(), "RESTIC_PASSWORD=supersecret")
		cmd.Dir = testutils.GetRepoRoot(t)
		stdout := bytes.NewBuffer(nil)
		cmd.Stdout = stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if !assert.NoError(t, err) {
			return
		}
		var output []map[string]any
		err = json.Unmarshal(stdout.Bytes(), &output)
		if !assert.NoError(t, err) {
			return
		}
		assert.Len(t, output, 1)
	}
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
	if !assert.NoError(t, err, "failed to backup") {
		return
	}

	cmd = exec.Command("restic", "-r", repoDir, "snapshots", "--json")
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD=supersecret")
	cmd.Dir = testutils.GetRepoRoot(t)
	stdout := bytes.NewBufferString("")
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if !assert.NoError(t, err, "failed to list expected snapshots in %s", repoDir) {
		return
	}
	var expectedSnapshots []map[string]any
	err = json.Unmarshal(stdout.Bytes(), &expectedSnapshots)
	if !assert.NoError(t, err) {
		return
	}

	// Check that exec returns the same thing
	cmd = testutils.StandardBackups(t, "exec", "-d", "my-dest")
	tc.Apply(cmd)
	cmd.Args = append(cmd.Args, "--", "snapshots", "--json")
	stdout = bytes.NewBufferString("")
	cmd.Stdout = stdout
	err = cmd.Run()
	if !assert.NoError(t, err, "failed to list expected snapshots with exec in %s", repoDir) {
		return
	}
	var actualSnapshots []map[string]any
	err = json.Unmarshal(stdout.Bytes(), &actualSnapshots)
	if !assert.NoError(t, err) {
		return
	}
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
			d1:
				backend: restic
				options:
					repo: %s
					env:
						RESTIC_PASSWORD: '{{ .Secrets.pass }}'
			d2:
				backend: restic
				options:
					repo: %s
					env:
						RESTIC_PASSWORD: '{{ .Secrets.pass }}'
		jobs:
			my-job:
				recipe: bogus
				backup-to: [d1, d2]
	`, t.TempDir(), t.TempDir())))

	for range 2 {
		cmd := testutils.StandardBackups(t, "backup", "my-job")
		tc.Apply(cmd)
		err := cmd.Run()
		if !assert.NoError(t, err, "failed to backup") {
			return
		}
	}

	for _, dest := range []string{"d1", "d2"} {
		cmd := testutils.StandardBackups(t, "list-backups", dest, "--json")
		tc.Apply(cmd)
		stdout := bytes.NewBuffer(nil)
		cmd.Stdout = stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if !assert.NoError(t, err, "failed to list backups for %s", dest) {
			return
		}
		var output []proto.ListBackupsResponseItem
		err = json.Unmarshal(stdout.Bytes(), &output)
		if !assert.NoError(t, err, dest) {
			return
		}

		assert.Len(t, output, 2)
		for i := range 2 {
			backupTime, err := time.Parse(time.RFC3339, output[i].Time)
			assert.NoError(t, err, "failed to parse output[%d].Time in %s", i, dest)
			assert.WithinRange(t, backupTime,
				backupTime.Add(time.Minute*-2),
				backupTime.Add(time.Minute*2))

			assert.NotEmpty(t, output[i].Id, i)
			assert.NotEmpty(t, output[i].Extra, i)
			assert.Equal(t, "my-job", output[i].Job, i)
			assert.Equal(t, dest, output[i].Destination, i)
			assert.Greater(t, output[i].Size, 0, i)
		}

		assert.NotEqual(t, output[0].Id, output[1].Id, dest)
		assert.Less(t, output[0].Time, output[1].Time, dest)
	}
}
