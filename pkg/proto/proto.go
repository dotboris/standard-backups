package proto

import (
	"fmt"
	"os"
)

func requireEnv(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("missing required environment variable %s", name)
	}
	if value == "" {
		return "", fmt.Errorf("required environment variable %s has an empty value", name)
	}
	return value, nil
}

func GetCommand() (string, error) {
	command, err := requireEnv("STANDARD_BACKUPS_COMMAND")
	if err != nil {
		return "", err
	}
	return command, err
}
