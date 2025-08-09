package config

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSecretLiteral(t *testing.T) {
	res, err := loadSecret(SecretConfigV1{Literal: "supersecret"})
	if assert.NoError(t, err) {
		assert.Equal(t, "supersecret", res)
	}
}

func TestLoadSecretFromFile(t *testing.T) {
	file := path.Join(t.TempDir(), "my-secret.txt")
	err := os.WriteFile(file, []byte("supersecret from file"), 0o600)
	if !assert.NoError(t, err) {
		return
	}
	res, err := loadSecret(SecretConfigV1{FromFile: file})
	if assert.NoError(t, err) {
		assert.Equal(t, "supersecret from file", res)
	}
}

func TestLoadSecretFromFileNotFound(t *testing.T) {
	_, err := loadSecret(SecretConfigV1{FromFile: "does-not-exist.txt"})
	assert.ErrorIs(t, err, os.ErrNotExist)
}
