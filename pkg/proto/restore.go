package proto

import (
	"encoding/json"
	"errors"
	"fmt"
)

type (
	RestoreRequest struct {
		RawOptions      map[string]any
		DestinationName string
		BackupId        string
		OutputDir       string
	}
	RestoreFunc func(*RestoreRequest) error
)

func NewRestoreRequestFromEnv() (*RestoreRequest, error) {
	destinationName, err := requireEnv("STANDARD_BACKUPS_DESTINATION_NAME")
	if err != nil {
		return nil, err
	}
	backupId, err := requireEnv("STANDARD_BACKUPS_BACKUP_ID")
	if err != nil {
		return nil, err
	}
	outputDir, err := requireEnv("STANDARD_BACKUPS_OUTPUT_DIR")
	if err != nil {
		return nil, err
	}

	rawOptions, err := requireEnv("STANDARD_BACKUPS_OPTIONS")
	if err != nil {
		return nil, err
	}
	var options map[string]any
	err = json.Unmarshal([]byte(rawOptions), &options)
	if err != nil {
		return nil, err
	}

	return &RestoreRequest{
		RawOptions:      options,
		DestinationName: destinationName,
		BackupId:        backupId,
		OutputDir:       outputDir,
	}, err
}

func (r *RestoreRequest) ToEnv() ([]string, error) {
	jsonOptions, err := json.Marshal(r.RawOptions)
	if err != nil {
		return nil, err
	}

	return []string{
		fmt.Sprintf("STANDARD_BACKUPS_BACKUP_ID=%s", r.BackupId),
		fmt.Sprintf("STANDARD_BACKUPS_DESTINATION_NAME=%s",
			r.DestinationName),
		fmt.Sprintf("STANDARD_BACKUPS_OUTPUT_DIR=%s", r.OutputDir),
		fmt.Sprintf("STANDARD_BACKUPS_OPTIONS=%s",
			jsonOptions),
	}, nil
}

func (bc *BackendClient) Restore(req *RestoreRequest) error {
	env, err := req.ToEnv()
	if err != nil {
		return err
	}
	cmd := bc.cmd("restore", env)
	err = cmd.Run()
	return err
}

func (bi *BackendImpl) restore() error {
	if bi.Restore == nil {
		return errors.New("unhandled command backup")
	}
	req, err := NewRestoreRequestFromEnv()
	if err != nil {
		return err
	}
	return bi.Restore(req)
}
