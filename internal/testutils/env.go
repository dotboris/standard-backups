package testutils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func SetEnv(t *testing.T, env []string) {
	t.Helper()
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		t.Setenv(key, value)
	}
}

type ToEnver interface {
	ToEnv() ([]string, error)
}

func SetEnvFromToEnv(t *testing.T, toEnver ToEnver) {
	t.Helper()
	env, err := toEnver.ToEnv()
	if !assert.NoError(t, err) {
		t.Error(err)
		t.FailNow()
		return
	}
	SetEnv(t, env)
}
