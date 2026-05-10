package proto

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	ARGS_ENV             = "STANDARD_BACKUPS_ARGS"
	BACKUP_ID_ENV        = "STANDARD_BACKUPS_BACKUP_ID"
	COMMAND_ENV          = "STANDARD_BACKUPS_COMMAND"
	DESTINATION_NAME_ENV = "STANDARD_BACKUPS_DESTINATION_NAME"
	EXCLUDE_ENV          = "STANDARD_BACKUPS_EXCLUDE"
	JOB_NAME_ENV         = "STANDARD_BACKUPS_JOB_NAME"
	OPTIONS_ENV          = "STANDARD_BACKUPS_OPTIONS"
	OUTPUT_DIR_ENV       = "STANDARD_BACKUPS_OUTPUT_DIR"
	PATHS_ENV            = "STANDARD_BACKUPS_PATHS"
	VARIANT_NAME_ENV     = "STANDARD_BACKUPS_VARIANT_NAME"
)

func getEnvStr(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("missing required environment variable %s", name)
	}
	if value == "" {
		return "", fmt.Errorf("required environment variable %s has an empty value", name)
	}
	return value, nil
}

func getEnvJson[T any](name string) (T, error) {
	var res T
	raw, err := getEnvStr(name)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal([]byte(raw), &res)
	if err != nil {
		return res, err
	}
	return res, nil
}

func toEnvStr(name, value string) string {
	return fmt.Sprintf("%s=%s", name, value)
}

func toEnvJson(name string, value any) (string, error) {
	jsonVal, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal %s to json: %w", name, err)
	}
	return fmt.Sprintf("%s=%s", name, jsonVal), nil
}
