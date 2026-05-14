package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/dotboris/standard-backups/internal/testbackend"
	"github.com/dotboris/standard-backups/pkg/proto"
)

func trace(traceDir string, command string, req any) error {
	if traceDir == "" {
		return nil
	}

	err := os.MkdirAll(traceDir, 0o755)
	if err != nil {
		return err
	}

	p := path.Join(traceDir, fmt.Sprintf("%s.json", command))
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to close %s: %s", p, err)
		}
	}()

	enc := json.NewEncoder(file)
	err = enc.Encode(req)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	traceDir := os.Getenv(testbackend.TRACE_DIR_ENV)
	var impl testbackend.Impl
	rawImpl, ok := os.LookupEnv(testbackend.IMPL_ENV)
	if ok {
		err := json.Unmarshal([]byte(rawImpl), &impl)
		if err != nil {
			log.Fatalf("failed to parse %s=%s: %s", testbackend.IMPL_ENV, rawImpl, err)
		}
	}

	b := proto.BackendImpl{}
	if impl.Backup.Enable {
		b.Backup = func(req *proto.BackupRequest) error {
			err := trace(traceDir, "backup", req)
			if err != nil {
				return err
			}
			if impl.Backup.Error != "" {
				return errors.New(impl.Backup.Error)
			}
			return nil
		}
	}
	if impl.Exec.Enable {
		b.Exec = func(req *proto.ExecRequest) error {
			err := trace(traceDir, "exec", req)
			if err != nil {
				return err
			}
			if impl.Exec.Error != "" {
				return errors.New(impl.Exec.Error)
			}
			return nil
		}
	}
	if impl.ListBackups.Enable {
		b.ListBackups = func(req *proto.ListBackupsRequest) (*proto.ListBackupsResponse, error) {
			err := trace(traceDir, "list-backups", req)
			if err != nil {
				return nil, err
			}
			if impl.ListBackups.Error != "" {
				return nil, errors.New(impl.ListBackups.Error)
			}
			return impl.ListBackups.Res, nil
		}
	}
	if impl.Restore.Enable {
		b.Restore = func(req *proto.RestoreRequest) error {
			err := trace(traceDir, "restore", req)
			if err != nil {
				return err
			}
			if impl.Restore.Error != "" {
				return errors.New(impl.Restore.Error)
			}
			return nil
		}
	}

	fmt.Fprintln(os.Stderr, "starting test backend")
	fmt.Fprintf(os.Stderr, "traceDir=%s\n", traceDir)
	fmt.Fprintf(os.Stderr, "impl=%#+v\n", impl)
	fmt.Fprintf(os.Stderr, "backend=%#+v\n", b)

	b.Execute()
}
