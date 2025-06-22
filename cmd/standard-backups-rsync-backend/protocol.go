package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Options struct {
	DestinationDir string `json:"destination-dir"`
}

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

func readCommand() (string, error) {
	command, err := requireEnv("STANDARD_BACKUPS_COMMAND")
	if err != nil {
		return "", err
	}
	return command, err
}

func readPaths() ([]string, error) {
	rawPaths, err := requireEnv("STANDARD_BACKUPS_PATHS")
	if err != nil {
		return nil, err
	}
	paths := strings.Split(rawPaths, ":")

	return paths, nil
}

func readOptions() (*Options, error) {
	rawOptions, err := requireEnv("STANDARD_BACKUPS_OPTIONS")
	if err != nil {
		return nil, err
	}
	var options Options
	err = json.Unmarshal([]byte(rawOptions), &options)
	if err != nil {
		return nil, err
	}

	return &options, nil
}
