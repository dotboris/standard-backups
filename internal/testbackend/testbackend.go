package testbackend

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/stretchr/testify/require"
)

const (
	IMPL_ENV      = "TEST_BACKEND_IMPL"
	TRACE_DIR_ENV = "TEST_BACKEND_TRACE_DIR"
	NAME          = "test"
	BIN           = "dist/standard-backups-test-backend"
)

type BaseImpl struct {
	Enable bool
	Error  string
}
type ListBackupsImpl struct {
	BaseImpl
	Res *proto.ListBackupsResponse
}
type Impl struct {
	Backup      BaseImpl
	Exec        BaseImpl
	ListBackups ListBackupsImpl
	Restore     BaseImpl
}

type TestBackend struct {
	t        *testing.T
	traceDir string
	impl     Impl
}

// New creates a new interface for controlling the test backend
func New(t *testing.T, impl Impl) *TestBackend {
	t.Helper()
	return &TestBackend{
		t:        t,
		traceDir: t.TempDir(),
		impl:     impl,
	}
}

func (b *TestBackend) AddSelf(tc *testutils.TestConfig) {
	b.t.Helper()
	tc.AddBackend(NAME, BIN)
}

func (b *TestBackend) Apply(cmd *exec.Cmd) {
	b.t.Helper()
	implJson, err := json.Marshal(b.impl)
	require.NoError(b.t, err)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("%s=%s", TRACE_DIR_ENV, b.traceDir),
		fmt.Sprintf("%s=%s", IMPL_ENV, implJson),
	)
}

func (b *TestBackend) RequireTrace(command string) []byte {
	b.t.Helper()
	p := path.Join(b.traceDir, fmt.Sprintf("%s.json", command))
	res, err := os.ReadFile(p)
	require.NoError(b.t, err)
	return res
}
