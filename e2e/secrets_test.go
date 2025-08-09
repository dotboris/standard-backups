package e2e

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestSecretsPassedToBackend(t *testing.T) {
	secretFile1 := path.Join(t.TempDir(), "secret1.txt")
	err := os.WriteFile(secretFile1, []byte("file secret 1"), 0o644)
	if !assert.NoError(t, err) {
		return
	}
	secretFile2 := path.Join(t.TempDir(), "secret2.txt")
	err = os.WriteFile(secretFile2, []byte("file secret 2\n"), 0o644)
	if !assert.NoError(t, err) {
		return
	}
	b := testutils.NewDumpBackend(t)
	tc := testutils.NewTestConfig(t)
	tc.AddBogusRecipe(t, "bogus")
	tc.AddBackend("test-backend", b.Path)
	tc.WriteConfig(fmt.Sprintf(testutils.DedentYaml(`
		version: 1
		secrets:
			literal:
				literal: supersecret
			file1:
				from-file: %s
			file2:
				from-file: %s
		destinations:
			my-dest:
				backend: test-backend
				options:
					literal: '{{ .Secrets.literal }}'
					file1: '{{ .Secrets.file1 }}'
					file2: '{{ .Secrets.file2 }}'
		jobs:
			my-job:
				recipe: bogus
				backup-to: [my-dest]
	`), secretFile1, secretFile2))

	lockFile := path.Join(t.TempDir(), "standard-backups.pid")

	cmd := testutils.StandardBackups(t,
		"validate-config",
		"--lockfile", lockFile,
	)
	cmd.Args = append(cmd.Args, tc.Args()...)
	err = cmd.Run()
	assert.NoError(t, err)

	cmd = testutils.StandardBackups(t,
		"backup", "my-job",
		"--lockfile", lockFile,
	)
	cmd.Args = append(cmd.Args, tc.Args()...)
	err = cmd.Run()
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]any{
			"literal": "supersecret",
			"file1":   "file secret 1",
			"file2":   "file secret 2\n",
		}, b.ReadOptions())
	}
}
