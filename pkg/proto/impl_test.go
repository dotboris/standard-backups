package proto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackendImplUnknownCommand(t *testing.T) {
	t.Setenv(COMMAND_ENV, "bogus")
	b := &BackendImpl{}
	err := b.execute()
	assert.EqualError(t, err, "unknown command bogus")
}
