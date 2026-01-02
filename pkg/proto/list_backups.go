package proto

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type (
	ListBackupsRequest struct {
		RawOptions      map[string]any
		DestinationName string
	}
	ListBackupsResponseItem struct {
		Id          string         `json:"id"`
		Time        string         `json:"time"`
		Job         string         `json:"job"`
		Destination string         `json:"destination"`
		Bytes       int            `json:"bytes"`
		Extra       map[string]any `json:"extra"`
	}
	ListBackupsResponse struct {
		Backups []ListBackupsResponseItem `json:"backups"`
	}
	ListBackupsFunc func(req *ListBackupsRequest) (*ListBackupsResponse, error)
)

func NewListBackupsRequestsFromEnv() (*ListBackupsRequest, error) {
	rawOptions, err := requireEnv("STANDARD_BACKUPS_OPTIONS")
	if err != nil {
		return nil, err
	}
	var options map[string]any
	err = json.Unmarshal([]byte(rawOptions), &options)
	if err != nil {
		return nil, err
	}

	destinationName, err := requireEnv("STANDARD_BACKUPS_DESTINATION_NAME")
	if err != nil {
		return nil, err
	}

	return &ListBackupsRequest{
		RawOptions:      options,
		DestinationName: destinationName,
	}, nil
}

func (lbr *ListBackupsRequest) ToEnv() ([]string, error) {
	jsonOptions, err := json.Marshal(lbr.RawOptions)
	if err != nil {
		return nil, err
	}

	return []string{
		fmt.Sprintf("STANDARD_BACKUPS_DESTINATION_NAME=%s",
			lbr.DestinationName),
		fmt.Sprintf("STANDARD_BACKUPS_OPTIONS=%s",
			jsonOptions),
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
	bs, err := json.Marshal(res)
	if err != nil {
		return err
	}
	fmt.Printf("%s", bs)
	return nil
}
