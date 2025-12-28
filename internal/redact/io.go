package redact

import (
	"io"
	"os"

	"golang.org/x/text/transform"
)

var (
	redactTransformer *RedactTransformer
	Stdout            io.WriteCloser
	Stderr            io.WriteCloser
)

func init() {
	r, err := NewTransformer()
	if err != nil {
		panic(err)
	}
	redactTransformer = r
	Stdout = transform.NewWriter(os.Stdout, r)
	Stderr = transform.NewWriter(os.Stderr, r)
}

func AddSecrets(secrets ...string) error {
	return redactTransformer.AddSecrets(secrets...)
}
