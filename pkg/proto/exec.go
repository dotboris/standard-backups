package proto

import (
	"errors"
	"os"
)

type (
	ExecFunc    func(req *ExecRequest) error
	ExecRequest struct {
		Args            []string
		DestinationName string
		VariantName     string
		RawOptions      map[string]any
	}
)

func NewExecRequestFromEnv() (*ExecRequest, error) {
	destinationName, _ := os.LookupEnv(DESTINATION_NAME_ENV)
	variantName, _ := os.LookupEnv(VARIANT_NAME_ENV)
	options, err := getEnvJson[map[string]any](OPTIONS_ENV)
	if err != nil {
		return nil, err
	}
	args, err := getEnvJson[[]string](ARGS_ENV)
	if err != nil {
		return nil, err
	}
	return &ExecRequest{
		DestinationName: destinationName,
		VariantName:     variantName,
		RawOptions:      options,
		Args:            args,
	}, nil
}

func (r *ExecRequest) ToEnv() ([]string, error) {
	optionsEnv, err := toEnvJson(OPTIONS_ENV, r.RawOptions)
	if err != nil {
		return nil, err
	}
	argsEnv, err := toEnvJson(ARGS_ENV, r.Args)
	if err != nil {
		return nil, err
	}
	return []string{
		argsEnv,
		toEnvStr(DESTINATION_NAME_ENV, r.DestinationName),
		toEnvStr(VARIANT_NAME_ENV, r.VariantName),
		optionsEnv,
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
