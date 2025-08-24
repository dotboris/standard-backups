package proto

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type (
	ExecFunc    func(req *ExecRequest) error
	ExecRequest struct {
		Args            []string
		DestinationName string
		RawOptions      map[string]any
	}
)

func NewExecRequestFromEnv() (*ExecRequest, error) {
	destinationName, _ := os.LookupEnv("STANDARD_BACKUPS_DESTINATION_NAME")

	rawOptions, err := requireEnv("STANDARD_BACKUPS_OPTIONS")
	if err != nil {
		return nil, err
	}
	var options map[string]any
	err = json.Unmarshal([]byte(rawOptions), &options)
	if err != nil {
		return nil, err
	}

	rawArgs, err := requireEnv("STANDARD_BACKUPS_ARGS")
	if err != nil {
		return nil, err
	}
	var args []string
	err = json.Unmarshal([]byte(rawArgs), &args)
	if err != nil {
		return nil, err
	}

	return &ExecRequest{
		DestinationName: destinationName,
		RawOptions:      options,
		Args:            args,
	}, nil
}

func (r *ExecRequest) ToEnv() ([]string, error) {
	jsonOptions, err := json.Marshal(r.RawOptions)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal destination options %v as JSON: %w",
			r.RawOptions,
			err,
		)
	}

	jsonArgs, err := json.Marshal(r.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal args %v as JSON: %w", r.RawOptions, err)
	}

	return []string{
		fmt.Sprintf("STANDARD_BACKUPS_ARGS=%s", jsonArgs),
		fmt.Sprintf("STANDARD_BACKUPS_DESTINATION_NAME=%s", r.DestinationName),
		fmt.Sprintf("STANDARD_BACKUPS_OPTIONS=%s", jsonOptions),
	}, nil
}

func (bc *BackendClient) Exec(req *ExecRequest) error {
	env, err := req.ToEnv()
	if err != nil {
		return err
	}
	cmd := bc.cmd("exec", env)
	err = cmd.Run()
	return err
}

func (bi *BackendImpl) exec() error {
	if bi.Exec == nil {
		return errors.New("unhandled command exec")
	}
	req, err := NewExecRequestFromEnv()
	if err != nil {
		return err
	}
	return bi.Exec(req)
}
