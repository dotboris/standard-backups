package proto

import (
	"errors"
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
	destinationName, err := getEnvStr(DESTINATION_NAME_ENV)
	if err != nil {
		return nil, err
	}
	backupId, err := getEnvStr(BACKUP_ID_ENV)
	if err != nil {
		return nil, err
	}
	outputDir, err := getEnvStr(OUTPUT_DIR_ENV)
	if err != nil {
		return nil, err
	}
	options, err := getEnvJson[map[string]any](OPTIONS_ENV)
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
	optionsEnv, err := toEnvJson(OPTIONS_ENV, r.RawOptions)
	if err != nil {
		return nil, err
	}
	return []string{
		toEnvStr(BACKUP_ID_ENV, r.BackupId),
		toEnvStr(DESTINATION_NAME_ENV, r.DestinationName),
		toEnvStr(OUTPUT_DIR_ENV, r.OutputDir),
		optionsEnv,
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
