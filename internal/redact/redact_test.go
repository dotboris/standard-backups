package redact

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/transform"
)

func TestEmptySecretError(t *testing.T) {
	r, err := New("")
	assert.Nil(t, r)
	assert.ErrorIs(t, err, ErrEmptySecret)
}

func TestNoopWithNoSecrets(t *testing.T) {
	r, err := New()
	if !assert.NoError(t, err) {
		return
	}

	str := "Hello world!"

	res, n, err := transform.String(r, str)
	assert.Equal(t, str, res)
	assert.Equal(t, len(str), n)
	assert.NoError(t, err)
}

func TestNoopWithNoMatch(t *testing.T) {
	r, err := New("bogus")
	if !assert.NoError(t, err) {
		return
	}

	str := "Hello world!"

	res, n, err := transform.String(r, str)
	assert.Equal(t, str, res)
	assert.Equal(t, len(str), n)
	assert.NoError(t, err)
}

func TestRedact(t *testing.T) {
	cases := []struct {
		name     string
		secrets  []string
		input    string
		expected string
	}{
		{
			name:     "one secret",
			secrets:  []string{"secret"},
			input:    "Hello secret world!",
			expected: "Hello *** world!",
		},
		{
			name:     "many secrets",
			secrets:  []string{"one", "two", "three"},
			input:    "Hello one - two - three world!",
			expected: "Hello *** - *** - *** world!",
		},
		{
			name:     "repeated secrets",
			secrets:  []string{"one", "two", "three"},
			input:    "Hello one two one three two three world!",
			expected: "Hello *** *** *** *** *** *** world!",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, err := New(c.secrets...)
			if !assert.NoError(t, err) {
				return
			}

			res, n, err := transform.String(r, c.input)
			assert.Equal(t, c.expected, res)
			assert.Equal(t, len(c.input), n)
			assert.NoError(t, err)
		})
	}
}

func TestOverlapSecretAcrossWrites(t *testing.T) {
	cases := []struct {
		name     string
		secrets  []string
		writes   []string
		expected string
	}{
		{
			name:     "simple",
			secrets:  []string{"beepboop"},
			writes:   []string{"hello beep", "boop world"},
			expected: "hello *** world",
		},
		{
			name:     "repeated chars",
			secrets:  []string{"AAAAA"},
			writes:   []string{"hello AAA", "AA world"},
			expected: "hello *** world",
		},
		{
			name: "multiple matching secrets",
			secrets: []string{
				"boop",
				"bboop",
				"beboop",
				"beeboop",
				"beepboop",
			},
			writes:   []string{"hello beep", "boop world"},
			expected: "hello *** world",
		},
		{
			name: "multiple matching secrets reverse",
			secrets: []string{
				"beepboop",
				"beeboop",
				"beboop",
				"bboop",
				"boop",
			},
			writes:   []string{"hello beep", "boop world"},
			expected: "hello *** world",
		},
		{
			name:     "simple no match",
			secrets:  []string{"beepboop"},
			writes:   []string{"hello beep", "beep world"},
			expected: "hello beepbeep world",
		},
		{
			name:     "repeated chars no match",
			secrets:  []string{"AAAAA"},
			writes:   []string{"hello AAA", "AB world"},
			expected: "hello AAAAB world",
		},
		{
			name: "multiple matching secrets no match",
			secrets: []string{
				"boop",
				"bboop",
				"beboop",
				"beeboop",
				"beepboop",
			},
			writes:   []string{"hello beep", "beep world"},
			expected: "hello beepbeep world",
		},
		{
			name:     "no match up to the end",
			secrets:  []string{"beepboop"},
			writes:   []string{"hello beep"},
			expected: "hello beep",
		},
		{
			name:     "overlap with previous match",
			secrets:  []string{"AAA"},
			writes:   []string{"BAAAA", "AAB"},
			expected: "B******B",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, err := New(c.secrets...)
			if !assert.NoError(t, err) {
				return
			}
			res := bytes.NewBuffer(nil)
			w := transform.NewWriter(res, r)

			for _, write := range c.writes {

				n, err := w.Write([]byte(write))
				if !assert.NoError(t, err) {
					return
				}
				assert.Equal(t, len(write), n)
			}

			err = w.Close()
			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, c.expected, res.String())
		})
	}
}
