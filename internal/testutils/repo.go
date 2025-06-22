package testutils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func GetRepoRoot(t *testing.T) string {
	t.Helper()
	res, err := getRepoRoot()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return res
}

func getRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	origCwd := cwd
	if err != nil {
		return "", err
	}
	for {
		if cwd == "" {
			return "", fmt.Errorf("could not find repository root in %s", origCwd)
		}
		gitDir, err := os.Stat(path.Join(cwd, ".git"))
		if err == nil && gitDir.IsDir() {
			return cwd, nil
		}
		cwd = filepath.Dir(cwd)
	}
}
