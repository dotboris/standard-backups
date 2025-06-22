package e2e

import (
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

func TestExampleConfigDump(t *testing.T) {
	cmd := exec.Command(
		"./dist/standard-backups",
		"config", "dump",
		"--config-dir", "examples/config/etc/standard-backups",
		"--no-color",
	)
	cmd.Dir = testutils.GetRepoRoot(t)
	output, err := cmd.CombinedOutput()
	if !assert.NoError(t, err, string(output)) {
		return
	}
	snaps.MatchSnapshot(t, string(output))
}

func TestBackupRsyncLocal(t *testing.T) {
	root := testutils.GetRepoRoot(t)
	destDir := filepath.Join(root, "dist/backups/local/")
	err := os.MkdirAll(destDir, 0755)
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

	cmd := exec.Command(
		"./dist/standard-backups",
		"backup", "test",
		"--config-dir", "examples/config/etc/standard-backups",
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if !assert.NoError(t, err, string(output)) {
		return
	}

	backupsAfter := listBackups()
	diff := backupsAfter.Difference(backupsBefore)
	assert.Equal(t, 1, diff.Cardinality())
	newBackup, ok := diff.Pop()
	assert.Equal(t, true, ok)

	err = filepath.WalkDir(newBackup, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		sourcePath := filepath.Join(root, "internal", strings.TrimPrefix(p, newBackup))
		source, err := os.ReadFile(sourcePath)
		if !assert.NoError(t, err) {
			return nil
		}
		dest, err := os.ReadFile(p)
		if !assert.NoError(t, err) {
			return nil
		}

		assert.Equal(t, source, dest, fmt.Sprintf("expected content of %s to match %s", sourcePath, dest))

		return nil
	})
	assert.NoError(t, err)
}
