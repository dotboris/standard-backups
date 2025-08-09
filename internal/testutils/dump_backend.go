package testutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"
)

type DumpBackend struct {
	Path        string
	optionsPath string
	t           *testing.T
}

func NewDumpBackend(t *testing.T) *DumpBackend {
	d := t.TempDir()
	optionsPath := path.Join(d, "options.json")
	backendPath := path.Join(d, "backend.sh")
	err := os.WriteFile(
		backendPath,
		[]byte(Dedent(fmt.Sprintf(`
			#!/bin/bash
			set -euo pipefail
			echo "$STANDARD_BACKUPS_OPTIONS" > %s
		`, optionsPath))),
		0o755,
	)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return nil
	}

	return &DumpBackend{
		Path:        backendPath,
		optionsPath: optionsPath,
		t:           t,
	}
}

func (b *DumpBackend) ReadOptions() map[string]any {
	bytes, err := os.ReadFile(b.optionsPath)
	if err != nil {
		b.t.Error(err)
		b.t.FailNow()
		return nil
	}
	var res map[string]any
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		b.t.Error(err)
		b.t.FailNow()
		return nil
	}
	return res
}
