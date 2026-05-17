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
	"github.com/stretchr/testify/require"
)

func TestExamplePrintConfig(t *testing.T) {
	cmd := testutils.StandardBackups(t, "print-config", "--no-color")
	testutils.ApplyExampleConfig(t, cmd)
	stdout := bytes.Buffer{}
	cmd.Stdout = &stdout
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	err := cmd.Run()
	require.NoError(t, err,
		fmt.Sprintf("stdout:\n%s\nstderr:\n%s",
			stdout.String(), stderr.String()))
	clean := strings.ReplaceAll(stdout.String(), testutils.GetRepoRoot(t), "[root]")
	snaps.MatchSnapshot(t, clean)
}

func TestExampleListBackends(t *testing.T) {
	cmd := testutils.StandardBackups(t, "list-backends", "--no-color")
	testutils.ApplyExampleConfig(t, cmd)
	stdout := bytes.Buffer{}
	cmd.Stdout = &stdout
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	err := cmd.Run()
	require.NoError(t, err,
		fmt.Sprintf("stdout:\n%s\nstderr:\n%s",
			stdout.String(), stderr.String()))
	clean := strings.ReplaceAll(stdout.String(), testutils.GetRepoRoot(t), "[root]")
	snaps.MatchSnapshot(t, clean)
}

func TestExampleListRecipes(t *testing.T) {
	cmd := testutils.StandardBackups(t, "list-recipes", "--no-color")
	testutils.ApplyExampleConfig(t, cmd)
	stdout := bytes.Buffer{}
	cmd.Stdout = &stdout
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	err := cmd.Run()
	require.NoError(t, err,
		fmt.Sprintf("stdout:\n%s\nstderr:\n%s",
			stdout.String(), stderr.String()))
	clean := strings.ReplaceAll(stdout.String(), testutils.GetRepoRoot(t), "[root]")
	snaps.MatchSnapshot(t, clean)
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
		require.NoError(t, err)
		dest, err := os.ReadFile(p)
		require.NoError(t, err)

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
	require.NoError(t, err)

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

	cmd := testutils.StandardBackups(t, "backup", "test")
	testutils.ApplyExampleConfig(t, cmd)
	err = cmd.Run()
	require.NoError(t, err)

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

	cmd := testutils.StandardBackups(t, "backup", "test-restic")
	stderr := bytes.NewBufferString("")
	cmd.Stderr = stderr
	testutils.ApplyExampleConfig(t, cmd)
	err := cmd.Run()
	require.NoError(t, err, stderr)

	cmd = exec.Command("restic", "-v", "-r", destDir, "snapshots", "--json")
	cmd.Env = append(
		os.Environ(),
		"RESTIC_PASSWORD=supersecret",
	)
	stderr = bytes.NewBufferString("")
	cmd.Stderr = stderr
	resticOutput, err := cmd.Output()
	require.NoError(t, err, stderr.String())
	var parsedOutput []map[string]any
	err = json.Unmarshal(resticOutput, &parsedOutput)
	require.NoError(t, err)
	assert.Len(t, parsedOutput, 1)
	assert.Equal(t, []any{
		"sb:dest:local-restic",
		"sb:job:test-restic",
		"sb:variant:medium",
	}, parsedOutput[0]["tags"])

	d := t.TempDir()
	cmd = exec.Command("restic", "-v", "-r", destDir, "restore", "latest", "--target", d)
	cmd.Env = append(
		os.Environ(),
		"RESTIC_PASSWORD=supersecret",
	)
	resticOutput, err = cmd.CombinedOutput()
	require.NoError(t, err, string(resticOutput))

	assertTreesMatch(t, root, d)
}
