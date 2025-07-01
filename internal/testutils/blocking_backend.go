package testutils

import (
	"fmt"
	"os"
	"path"
	"testing"
)

// BlockingBackend is standard-backups backend meant for testing. When invoked,
// it blocks waiting for a signal from the testing process. All its actions are
// logged to a file so that they can be examined later on.
type BlockingBackend struct {
	Path     string
	waitFile string
	t        *testing.T
}

func NewBlockingBackend(t *testing.T, name string, logFile string) *BlockingBackend {
	t.Helper()
	d := t.TempDir()
	waitFile := path.Join(d, "wait.txt")
	backendPath := path.Join(d, fmt.Sprintf("%s.sh", name))
	err := os.WriteFile(
		backendPath,
		[]byte(Dedent(fmt.Sprintf(`
			#!/bin/bash
			name=%s
			log_file=%s
			wait_file=%s
			log () {
				echo "$name: $*"
				echo "$name: $*" >> $log_file
			}
			log started
			log waiting
			while true; do
				if [[ -e "$wait_file" ]]; then
					break
				fi
				sleep 0.25
			done
			log unblocked
			log done
		`, name, logFile, waitFile))),
		0o755,
	)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return nil
	}
	return &BlockingBackend{
		Path:     backendPath,
		waitFile: waitFile,
		t:        t,
	}
}

func (b *BlockingBackend) Unblock() {
	b.t.Helper()
	err := os.WriteFile(b.waitFile, []byte{}, 0o644)
	if err != nil {
		b.t.Error(err)
		b.t.FailNow()
		return
	}
}
