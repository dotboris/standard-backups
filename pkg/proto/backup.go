package proto

import (
	"errors"
	"os"
)

type (
	BackupFunc    func(req *BackupRequest) error
	BackupRequest struct {
		Paths           []string
		Exclude         []string
		DestinationName string
		VariantName     string
		JobName         string
		RawOptions      map[string]any
	}
)

func NewBackupRequestFromEnv() (*BackupRequest, error) {
	paths, err := getEnvJson[[]string](PATHS_ENV)
	if err != nil {
		return nil, err
	}
	exclude, err := getEnvJson[[]string](EXCLUDE_ENV)
	destinationName, err := getEnvStr(DESTINATION_NAME_ENV)
	if err != nil {
		return nil, err
	}
	variantName := os.Getenv(VARIANT_NAME_ENV)
	jobName, err := getEnvStr(JOB_NAME_ENV)
	if err != nil {
		return nil, err
	}
	options, err := getEnvJson[map[string]any](OPTIONS_ENV)
	if err != nil {
		return nil, err
	}
	return &BackupRequest{
		Paths:           paths,
		Exclude:         exclude,
		DestinationName: destinationName,
		VariantName:     variantName,
		JobName:         jobName,
		RawOptions:      options,
	}, nil
}

func (br *BackupRequest) ToEnv() ([]string, error) {
	pathsEnv, err := toEnvJson(PATHS_ENV, br.Paths)
	if err != nil {
		return nil, err
	}
	excludeEnv, err := toEnvJson(EXCLUDE_ENV, br.Exclude)
	if err != nil {
		return nil, err
	}
	optionsEnv, err := toEnvJson(OPTIONS_ENV, br.RawOptions)
	if err != nil {
		return nil, err
	}
	return []string{
		pathsEnv,
		excludeEnv,
		toEnvStr(DESTINATION_NAME_ENV, br.DestinationName),
		toEnvStr(VARIANT_NAME_ENV, br.VariantName),
		toEnvStr(JOB_NAME_ENV, br.JobName),
		optionsEnv,
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
