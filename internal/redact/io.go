package redact

import (
	"io"
	"os"

	"golang.org/x/text/transform"
)

type WriteCloser = io.WriteCloser

var (
	redactTransformer *RedactTransformer
	Stdout            WriteCloser
	Stderr            WriteCloser
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
