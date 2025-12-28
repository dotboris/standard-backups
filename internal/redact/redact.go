package redact

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sync"

	"golang.org/x/text/transform"
)

type RedactTransformer struct {
	transform.NopResetter
	secrets     [][]byte
	secretsLock sync.RWMutex
}

var ErrEmptySecret = errors.New("secret is empty")

const REPLACE = "***"

func NewTransformer(secrets ...string) (*RedactTransformer, error) {
	bSecrets := make([][]byte, len(secrets))
	for i, s := range secrets {
		b, err := convertSecret(s)
		if err != nil {
			return nil, fmt.Errorf("bad secret at index %d: %w", i, err)
		}
		bSecrets[i] = b
	}

	return &RedactTransformer{
		secrets: bSecrets,
	}, nil
}

func convertSecret(s string) ([]byte, error) {
	if s == "" {
		return nil, ErrEmptySecret
	}
	return []byte(s), nil
}

func (r *RedactTransformer) AddSecrets(secrets ...string) error {
	r.secretsLock.Lock()
	defer r.secretsLock.Unlock()

	for i, s := range secrets {
		b, err := convertSecret(s)
		if err != nil {
			return fmt.Errorf("bad secret at index %d: %w", i, err)
		}
		r.secrets = append(r.secrets, b)
	}
	return nil
}

// Transform implements [transform.Transformer].
func (r *RedactTransformer) Transform(dst []byte, src []byte, atEOF bool) (int, int, error) {
	r.secretsLock.RLock()
	defer r.secretsLock.RUnlock()

	if len(r.secrets) == 0 {
		n, err := checkCopy(dst, src)
		return n, n, err
	}

	var nDst, nSrc int
	for {
		iSecret := math.MaxInt
		nSecret := 0

		// Find the closest secret
		for _, secret := range r.secrets {
			currIndex := bytes.Index(src[nSrc:], secret)
			if currIndex != -1 && currIndex < iSecret {
				iSecret = currIndex
				nSecret = len(secret)
			}
		}

		// No more secrets in buffer, tap out
		if iSecret == math.MaxInt {
			break
		}

		// Copy up to secret
		n, err := checkCopy(dst[nDst:], src[nSrc:nSrc+iSecret])
		nSrc += n
		nDst += n
		if err != nil {
			return nDst, nSrc, err
		}

		// Copy secret replacement
		n, err = checkCopy(dst[nDst:], []byte(REPLACE))
		nSrc += nSecret
		nDst += n
		if err != nil {
			return nDst, nSrc, err
		}
	}

	if atEOF {
		// Copy whatever's left
		n, err := checkCopy(dst[nDst:], src[nSrc:])
		nSrc += n
		nDst += n
		return nDst, nSrc, err
	}

	// Find the biggest overlap between our secrets and the tail of src
	overlap := 0
	for _, secret := range r.secrets {
		start := min(
			len(secret)-1, // Don't check full secret, should have already matched
			len(src)-nSrc, // Only consider what's left of src
		)
		// Go backwards to land on the biggest first
		for n := start; n > overlap; n-- {
			if bytes.Equal(src[len(src)-n:], secret[:n]) {
				overlap = max(overlap, n)
				break
			}
		}
	}

	// Copy leftover accounting for overlap
	n, err := checkCopy(dst[nDst:], src[nSrc:len(src)-overlap])
	nSrc += n
	nDst += n
	if err == nil && overlap != 0 {
		// if there's overlap, we're missing data from src
		err = transform.ErrShortSrc
	}
	return nDst, nSrc, err
}

func checkCopy(dst, src []byte) (int, error) {
	n := copy(dst, src)
	if n < len(src) {
		return n, transform.ErrShortDst
	}
	return n, nil
}
