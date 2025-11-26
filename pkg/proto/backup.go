package proto

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type (
	BackupFunc    func(req *BackupRequest) error
	BackupRequest struct {
		Paths           []string
		Exclude         []string
		DestinationName string
		JobName         string
		RawOptions      map[string]any
	}
)

func NewBackupRequestFromEnv() (*BackupRequest, error) {
	rawPaths, err := requireEnv("STANDARD_BACKUPS_PATHS")
	if err != nil {
		return nil, err
	}
	paths := strings.Split(rawPaths, ":")

	exclude := []string{}
	rawExclude, _ := os.LookupEnv("STANDARD_BACKUPS_EXCLUDE")
	if rawExclude != "" {
		exclude = strings.Split(rawExclude, ":")
	}

	destinationName, err := requireEnv("STANDARD_BACKUPS_DESTINATION_NAME")
	if err != nil {
		return nil, err
	}

	jobName, err := requireEnv("STANDARD_BACKUPS_JOB_NAME")
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

	return &BackupRequest{
		Paths:           paths,
		Exclude:         exclude,
		DestinationName: destinationName,
		JobName:         jobName,
		RawOptions:      options,
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
		fmt.Sprintf("STANDARD_BACKUPS_EXCLUDE=%s",
			strings.Join(br.Exclude, ":")),
		fmt.Sprintf("STANDARD_BACKUPS_DESTINATION_NAME=%s", br.DestinationName),
		fmt.Sprintf("STANDARD_BACKUPS_JOB_NAME=%s", br.JobName),
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
