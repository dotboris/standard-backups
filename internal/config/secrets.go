package config

import (
	"errors"
	"fmt"
	"os"
)

var errLoadSecretUnimplemented = errors.New("config.loadSecret: unimplemented")

func loadSecrets(secretsConfig map[string]SecretConfigV1) (map[string]string, error) {
	res := map[string]string{}
	for key, c := range secretsConfig {
		value, err := loadSecret(c)
		if err != nil {
			return nil, fmt.Errorf("failed to load secret %s: %w", key, err)
		}
		res[key] = value
	}
	return res, nil
}

func loadSecret(secret SecretConfigV1) (string, error) {
	if secret.FromFile != "" {
		res, err := os.ReadFile(secret.FromFile)
		if err != nil {
			return "", err
		}
		return string(res), nil
	} else if secret.Literal != "" {
		return secret.Literal, nil
	} else {
		return "", errLoadSecretUnimplemented
	}
}
