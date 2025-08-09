package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
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

func TestResticBackup(t *testing.T) {
	repoDir := t.TempDir()

	tc := testutils.NewTestConfig(t)
	tc.AddBackend("restic", "dist/standard-backups-restic-backend")
	sourceDir := tc.AddBogusRecipe(t, "bogus")
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
	cmd.Args = append(cmd.Args, tc.Args()...)
	err := cmd.Run()
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

	// Check that we have the tags
	cmd = exec.Command("restic", "-r", repoDir, "snapshots", "--json")
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD=supersecret")
	cmd.Dir = testutils.GetRepoRoot(t)
	stdout := bytes.NewBufferString("")
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if !assert.NoError(t, err, "restic repo %s has not been initialized", repoDir) {
		return
	}
	var output []map[string]any
	err = json.Unmarshal(stdout.Bytes(), &output)
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, output, 1)
	assert.Equal(t, []any{
		"sb:dest:my-dest",
		"sb:job:my-job",
	}, output[0]["tags"])

	// Test restore
	restoreDir := t.TempDir()
	cmd = exec.Command("restic",
		"-v",
		"-r", repoDir,
		"restore", fmt.Sprintf("latest:%s", sourceDir),
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
	cmd.Args = append(cmd.Args, tc.Args()...)
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
		cmd.Args = append(cmd.Args, tc.Args()...)
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
