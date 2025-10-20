package testutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"
)

type DumpBackend struct {
	Path   string
	outDir string
	t      *testing.T
}

func NewDumpBackend(t *testing.T) *DumpBackend {
	d := t.TempDir()
	backendPath := path.Join(d, "backend.sh")
	err := os.WriteFile(
		backendPath,
		[]byte(Dedent(fmt.Sprintf(`
			#!/usr/bin/env bash
			set -euo pipefail
			dir='%s'
			while IFS= read -r line; do
				var="${line%%%%=*}"
				value="${line#*=}"
				echo -n "$value" > "$dir/$var.json"
			done <<< $(env | grep "^STANDARD_BACKUPS_")
		`, d))),
		0o755,
	)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return nil
	}

	return &DumpBackend{
		Path:   backendPath,
		outDir: d,
		t:      t,
	}
}

func (b *DumpBackend) ReadBytes(name string) []byte {
	bytes, err := os.ReadFile(path.Join(b.outDir, fmt.Sprintf("%s.json", name)))
	if err != nil {
		b.t.Error(err)
		b.t.FailNow()
		return nil
	}
	return bytes
}

func (b *DumpBackend) ReadString(name string) string {
	return string(b.ReadBytes(name))
}

func (b *DumpBackend) ReadJsonMap(name string) map[string]any {
	bytes := b.ReadBytes(name)

	var res map[string]any
	err := json.Unmarshal(bytes, &res)
	if err != nil {
		b.t.Error(err)
		b.t.FailNow()
		return nil
	}
	return res
}

func (b *DumpBackend) ReadJsonArray(name string) []any {
	bytes := b.ReadBytes(name)

	var res []any
	err := json.Unmarshal(bytes, &res)
	if err != nil {
		b.t.Error(err)
		b.t.FailNow()
		return nil
	}
	return res
}
