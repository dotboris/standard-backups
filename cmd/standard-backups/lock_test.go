package main

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setLockfilePath(t *testing.T, p string) {
	t.Helper()
	oldPath := lockfilePath
	lockfilePath = p
	t.Cleanup(func() {
		lockfilePath = oldPath
	})
}

func TestAcquireLock(t *testing.T) {
	d := t.TempDir()
	setLockfilePath(t, path.Join(d, "test.pid"))

	unlock, err := acquireLock()
	if assert.NoError(t, err) {
		unlock()
	}
}
