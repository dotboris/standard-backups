package e2e

import (
	"os/exec"
	"testing"

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
