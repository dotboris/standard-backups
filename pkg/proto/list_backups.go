package proto

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
)

type (
	ListBackupsRequest struct {
		RawOptions      map[string]any
		DestinationName string
		VariantName     string
	}
	ListBackupsResponseItem struct {
		Id          string         `json:"id"`
		Time        string         `json:"time"`
		Job         string         `json:"job"`
		Destination string         `json:"destination"`
		Variant     string         `json:"variant"`
		Size        int            `json:"size"` // Size of the backup in bytes
		Extra       map[string]any `json:"extra"`
	}
	ListBackupsResponse struct {
		Backups []ListBackupsResponseItem `json:"backups"`
	}
	ListBackupsFunc func(req *ListBackupsRequest) (*ListBackupsResponse, error)
)

func NewListBackupsRequestsFromEnv() (*ListBackupsRequest, error) {
	options, err := getEnvJson[map[string]any](OPTIONS_ENV)
	if err != nil {
		return nil, err
	}
	destinationName, err := getEnvStr(DESTINATION_NAME_ENV)
	if err != nil {
		return nil, err
	}
	variantName := os.Getenv(VARIANT_NAME_ENV)
	return &ListBackupsRequest{
		RawOptions:      options,
		DestinationName: destinationName,
		VariantName:     variantName,
	}, nil
}

func (lbr *ListBackupsRequest) ToEnv() ([]string, error) {
	optionsEnv, err := toEnvJson(OPTIONS_ENV, lbr.RawOptions)
	if err != nil {
		return nil, err
	}
	return []string{
		toEnvStr(DESTINATION_NAME_ENV, lbr.DestinationName),
		toEnvStr(VARIANT_NAME_ENV, lbr.VariantName),
		optionsEnv,
	}, nil
}

func (bc *BackendClient) ListBackups(req *ListBackupsRequest) (*ListBackupsResponse, error) {
	env, err := req.ToEnv()
	if err != nil {
		return nil, err
	}
	cmd := bc.cmd("list-backups", env)
	stdout := bytes.NewBuffer(nil)
	cmd.Stdout = stdout
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	var res ListBackupsResponse
	err = json.Unmarshal(stdout.Bytes(), &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (bi *BackendImpl) listBackups() error {
	if bi.ListBackups == nil {
		return errors.New("unhandled command list-backups")
	}
	req, err := NewListBackupsRequestsFromEnv()
	if err != nil {
		return err
	}
	res, err := bi.ListBackups(req)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(res)
	if err != nil {
		return err
	}
	return nil
}
