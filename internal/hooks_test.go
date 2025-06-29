package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/dotboris/standard-backups/internal/config"
	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRunHookSh(t *testing.T) {
	d := t.TempDir()
	outFile := path.Join(d, "out.txt")
	err := runHook(config.HookV1{
		Shell: "sh",
		Command: testutils.Dedent(fmt.Sprintf(`
			echo hello from $0 > %s
		`, outFile)),
	})
	if assert.NoError(t, err) {
		content, err := os.ReadFile(outFile)
		if assert.NoError(t, err) {
			assert.Equal(t, string(content), "hello from sh\n")
		}
	}
}

func TestRunHookBash(t *testing.T) {
	d := t.TempDir()
	outFile := path.Join(d, "out.txt")
	err := runHook(config.HookV1{
		Shell: "bash",
		Command: testutils.Dedent(fmt.Sprintf(`
			echo hello from $0 > %s
		`, outFile)),
	})
	if assert.NoError(t, err) {
		content, err := os.ReadFile(outFile)
		if assert.NoError(t, err) {
			assert.Equal(t, string(content), "hello from bash\n")
		}
	}
}

func TestRunHookUnsupportedShell(t *testing.T) {
	err := runHook(config.HookV1{
		Shell:   "bogus",
		Command: "bogus",
	})
	assert.ErrorIs(t, err, errUnsupportedShell)
}

func TestRunHookShError(t *testing.T) {
	err := runHook(config.HookV1{
		Shell: "sh",
		Command: testutils.Dedent(`
			exit 42
		`),
	})
	var exitError *exec.ExitError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &exitError) {
		assert.Equal(t, exitError.ExitCode(), 42)
	}
}

func TestRunHookBashError(t *testing.T) {
	err := runHook(config.HookV1{
		Shell: "sh",
		Command: testutils.Dedent(`
			exit 42
		`),
	})
	var exitError *exec.ExitError
	if assert.Error(t, err) && assert.ErrorAs(t, err, &exitError) {
		assert.Equal(t, exitError.ExitCode(), 42)
	}
}
