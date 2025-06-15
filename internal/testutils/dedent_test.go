package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDedent(t *testing.T) {
	assert.Equal(
		t,
		"cool text\nmore text\n\teven indented\n\t\tvery indented",
		Dedent(`
			cool text
			more text
				even indented
					very indented
		`),
	)
}

func TestDedentYaml(t *testing.T) {
	assert.Equal(
		t,
		"foo: 1\nbar:\n  nested: yay",
		DedentYaml(`
			foo: 1
			bar:
				nested: yay
		`),
	)
}
