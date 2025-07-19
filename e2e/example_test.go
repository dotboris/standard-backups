package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

func TestExamplePrintConfig(t *testing.T) {
	cmd := testutils.StandardBackups(t, "print-config", "--no-color")
	cmd.Args = append(cmd.Args, testutils.ExampleConfigArgs...)
	stdout := bytes.Buffer{}
	cmd.Stdout = &stdout
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	err := cmd.Run()
	if !assert.NoError(t, err,
		fmt.Sprintf("stdout:\n%s\nstderr:\n%s",
			stdout.String(), stderr.String())) {
		return
	}
	snaps.MatchSnapshot(t, stdout.String())
}

func assertTreesMatch(t *testing.T, expectedPath string, actualPath string) bool {
	err := filepath.WalkDir(actualPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		sourcePath := filepath.Join(expectedPath, strings.TrimPrefix(p, actualPath))
		source, err := os.ReadFile(sourcePath)
		if !assert.NoError(t, err) {
			return nil
		}
		dest, err := os.ReadFile(p)
		if !assert.NoError(t, err) {
			return nil
		}

		assert.Equal(
			t,
			source,
			dest,
			fmt.Sprintf("expected content of %s to match %s", sourcePath, dest),
		)

		return nil
	})
	return assert.NoError(t, err)
}

func TestBackupRsyncLocal(t *testing.T) {
	root := testutils.GetRepoRoot(t)
	destDir := filepath.Join(root, "dist/backups/local/")
	err := os.MkdirAll(destDir, 0o755)
	if !assert.NoError(t, err) {
		return
	}

	listBackups := func() mapset.Set[string] {
		t.Helper()
		res := mapset.NewSet[string]()
		entries, err := os.ReadDir(destDir)
		if !assert.NoError(t, err) {
			t.FailNow()
			return nil
		}
		for _, entry := range entries {
			if entry.IsDir() {
				res.Add(path.Join(destDir, entry.Name()))
			}
		}
		return res
	}

	backupsBefore := listBackups()

	cmd := testutils.StandardBackups(t,
		"backup", "test",
		"--lockfile", path.Join(t.TempDir(), "standard-backups.pid"),
	)
	cmd.Args = append(cmd.Args, testutils.ExampleConfigArgs...)
	err = cmd.Run()
	if !assert.NoError(t, err) {
		return
	}

	backupsAfter := listBackups()
	diff := backupsAfter.Difference(backupsBefore)
	assert.Equal(t, 1, diff.Cardinality())
	newBackup, ok := diff.Pop()
	assert.Equal(t, true, ok)

	assertTreesMatch(t, root, newBackup)
}

func TestExampleResticLocal(t *testing.T) {
	root := testutils.GetRepoRoot(t)
	destDir := path.Join(root, "dist/backups/restic-local")
	_ = os.RemoveAll(path.Join(destDir))

	cmd := testutils.StandardBackups(t,
		"backup", "test-restic",
		"--lockfile", path.Join(t.TempDir(), "standard-backups.pid"),
	)
	cmd.Args = append(cmd.Args, testutils.ExampleConfigArgs...)
	err := cmd.Run()
	if !assert.NoError(t, err) {
		return
	}

	cmd = exec.Command("restic", "-v", "-r", destDir, "snapshots", "--json")
	cmd.Env = append(
		os.Environ(),
		"RESTIC_PASSWORD=supersecret",
	)
	stderr := bytes.NewBufferString("")
	cmd.Stderr = stderr
	resticOutput, err := cmd.Output()
	if !assert.NoError(t, err, stderr.String()) {
		return
	}
	var parsedOutput []map[string]any
	err = json.Unmarshal(resticOutput, &parsedOutput)
	if !assert.NoError(t, err) {
		return
	}
	assert.Len(t, parsedOutput, 1)
	assert.Equal(t, []any{
		"sb:dest:local-restic",
		"sb:job:test-restic",
	}, parsedOutput[0]["tags"])

	d := t.TempDir()
	cmd = exec.Command("restic", "-v", "-r", destDir, "restore", "latest", "--target", d)
	cmd.Env = append(
		os.Environ(),
		"RESTIC_PASSWORD=supersecret",
	)
	resticOutput, err = cmd.CombinedOutput()
	if !assert.NoError(t, err, string(resticOutput)) {
		return
	}

	assertTreesMatch(t, root, d)
}
