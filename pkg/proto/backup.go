package proto

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type BackupFunc func(req *BackupRequest) error

type BackupRequest struct {
	Paths      []string
	RawOptions map[string]any
}

func NewBackupRequestFromEnv() (*BackupRequest, error) {
	rawPaths, err := requireEnv("STANDARD_BACKUPS_PATHS")
	if err != nil {
		return nil, err
	}
	paths := strings.Split(rawPaths, ":")

	rawOptions, err := requireEnv("STANDARD_BACKUPS_OPTIONS")
	if err != nil {
		return nil, err
	}
	var options map[string]any
	err = json.Unmarshal([]byte(rawOptions), &options)
	if err != nil {
		return nil, err
	}

	return &BackupRequest{
		Paths:      paths,
		RawOptions: options,
	}, nil
}

func (br *BackupRequest) ToEnv() ([]string, error) {
	jsonOptions, err := json.Marshal(br.RawOptions)
	if err != nil {
		return nil, err
	}

	return []string{
		fmt.Sprintf("STANDARD_BACKUPS_PATHS=%s",
			strings.Join(br.Paths, ":")),
		fmt.Sprintf("STANDARD_BACKUPS_OPTIONS=%s",
			jsonOptions),
	}, nil
}

func (bc *BackendClient) Backup(req *BackupRequest) error {
	env, err := req.ToEnv()
	if err != nil {
		return err
	}
	cmd := bc.cmd("backup", env)
	err = cmd.Run()
	return err
}

func (bi *BackendImpl) backup() error {
	if bi.Backup == nil {
		return errors.New("unhandled command backup")
	}
	req, err := NewBackupRequestFromEnv()
	if err != nil {
		return err
	}
	return bi.Backup(req)
}
