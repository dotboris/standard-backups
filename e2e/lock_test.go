package e2e

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestBackupLock(t *testing.T) {
	logFile := path.Join(t.TempDir(), "backends.log")
	tc := testutils.NewTestConfig(t)
	backend1 := testutils.NewBlockingBackend(t, "backend1", logFile)
	tc.AddBackend("backend1", backend1.Path)
	backend2 := testutils.NewBlockingBackend(t, "backend2", logFile)
	tc.AddBackend("backend2", backend2.Path)
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(`
		version: 1
		backends:
			backend1: {enable: true}
			backend2: {enable: true}
		destinations:
			dest1:
				backend: backend1
			dest2:
				backend: backend2
		jobs:
			test1:
				recipe: bogus
				backup-to: [dest1]
			test2:
				recipe: bogus
				backup-to: [dest2]
	`))

	lockFile := path.Join(t.TempDir(), "standard-backups.pid")
	root := testutils.GetRepoRoot(t)

	startCmd := func(jobName string) *exec.Cmd {
		t.Helper()
		cmd := exec.Command(
			"./dist/standard-backups",
			"backup", jobName,
			"--log-level", "debug",
			"--config-dir", tc.Dir,
			"--lockfile", lockFile,
		)
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			t.Error(err)
			t.FailNow()
			return nil
		}

		t.Cleanup(func() {
			_ = cmd.Process.Kill()
		})

		return cmd
	}

	cmd1 := startCmd("test1")
	time.Sleep(250 * time.Millisecond)
	cmd2 := startCmd("test2")

	backend1.Unblock()
	backend2.Unblock()

	err := cmd1.Wait()
	if !assert.NoError(t, err) {
		return
	}
	err = cmd2.Wait()
	if !assert.NoError(t, err) {
		return
	}

	log, err := os.ReadFile(logFile)
	if assert.NoError(t, err) {
		assert.Equal(t,
			testutils.Dedent(`
				backend1: started
				backend1: waiting
				backend1: unblocked
				backend1: done
				backend2: started
				backend2: waiting
				backend2: unblocked
				backend2: done
			`),
			strings.TrimRight(string(log), "\n"),
		)
	}
}

func TestBackupLockTimeout(t *testing.T) {
	logFile := path.Join(t.TempDir(), "backends.log")
	tc := testutils.NewTestConfig(t)
	backend1 := testutils.NewBlockingBackend(t, "backend1", logFile)
	tc.AddBackend("backend1", backend1.Path)
	backend2 := testutils.NewBlockingBackend(t, "backend2", logFile)
	tc.AddBackend("backend2", backend2.Path)
	tc.AddBogusRecipe(t, "bogus")
	tc.WriteConfig(testutils.DedentYaml(`
		version: 1
		backends:
			backend1: {enable: true}
			backend2: {enable: true}
		destinations:
			dest1:
				backend: backend1
			dest2:
				backend: backend2
		jobs:
			test1:
				recipe: bogus
				backup-to: [dest1]
			test2:
				recipe: bogus
				backup-to: [dest2]
	`))

	lockFile := path.Join(t.TempDir(), "standard-backups.pid")
	root := testutils.GetRepoRoot(t)

	startCmd := func(jobName string, stderr io.Writer) *exec.Cmd {
		t.Helper()
		cmd := exec.Command(
			"./dist/standard-backups",
			"backup", jobName,
			"--log-level", "debug",
			"--config-dir", tc.Dir,
			"--lockfile", lockFile,
			"--lock-timeout", "1s",
			"--no-color",
		)
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = stderr
		err := cmd.Start()
		if err != nil {
			t.Error(err)
			t.FailNow()
			return nil
		}

		t.Cleanup(func() {
			_ = cmd.Process.Kill()
		})

		return cmd
	}

	cmd1 := startCmd("test1", os.Stderr)
	time.Sleep(250 * time.Millisecond)
	cmd2Stderr := bytes.NewBuffer(nil)
	cmd2 := startCmd("test2", cmd2Stderr)

	err := cmd2.Wait()
	var exitErr *exec.ExitError
	if assert.ErrorAs(t, err, &exitErr) {
		assert.Equal(t, 1, exitErr.ExitCode())
		assert.Contains(t,
			cmd2Stderr.String(),
			fmt.Sprintf(
				"Error: failed to acquire lock %s after %d\n",
				lockFile, 1*time.Second),
		)
	}

	backend1.Unblock()
	backend2.Unblock()
	err = cmd1.Wait()
	if !assert.NoError(t, err) {
		return
	}
}
