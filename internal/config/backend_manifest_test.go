package config

import (
	"os"
	"path"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestLoadBackendManifestsSingleFile(t *testing.T) {
	d := t.TempDir()
	p := path.Join(d, "example.yaml")
	err := os.WriteFile(p,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: example 1
			description: the first example
			bin: /path/to/backend
			protocol-version: 1
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	backendManifests, err := LoadBackendManifests([]string{d})
	if assert.NoError(t, err) {
		assert.Equal(t, []BackendManifestV1{
			{
				path:            p,
				Version:         1,
				Name:            "example 1",
				Description:     "the first example",
				Bin:             "/path/to/backend",
				ProtocolVersion: 1,
			},
		}, backendManifests)
	}
}

func TestLoadBackendManifestsMultipleFiles(t *testing.T) {
	d1 := t.TempDir()
	p1 := path.Join(d1, "backend1.yaml")
	err := os.WriteFile(p1,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: backend1
			description: the backend1
			bin: /path/to/backend1
			protocol-version: 1
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	d2 := t.TempDir()
	p2 := path.Join(d2, "backend2.yaml")
	err = os.WriteFile(p2,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: backend2
			description: the backend2
			bin: /path/to/backend2
			protocol-version: 1
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	backendManifests, err := LoadBackendManifests([]string{d1, d2})
	if assert.NoError(t, err) {
		assert.Equal(t, []BackendManifestV1{
			{
				path:            p1,
				Version:         1,
				Name:            "backend1",
				Description:     "the backend1",
				Bin:             "/path/to/backend1",
				ProtocolVersion: 1,
			},
			{
				path:            p2,
				Version:         1,
				Name:            "backend2",
				Description:     "the backend2",
				Bin:             "/path/to/backend2",
				ProtocolVersion: 1,
			},
		}, backendManifests)
	}
}

func TestLoadBackendManifestsIgnoreNonYaml(t *testing.T) {
	d := t.TempDir()
	err := os.WriteFile(path.Join(d, "bogus.txt"), []byte("bogus"), 0o644)
	if !assert.NoError(t, err) {
		return
	}
	backendManifests, err := LoadBackendManifests([]string{d})
	if assert.NoError(t, err) {
		assert.Equal(t, []BackendManifestV1{}, backendManifests)
	}
}

func TestLoadBackendManifestsEmptyDir(t *testing.T) {
	d := t.TempDir()
	backendManifests, err := LoadBackendManifests([]string{d})
	if assert.NoError(t, err) {
		assert.Equal(t, []BackendManifestV1{}, backendManifests)
	}
}

func TestLoadBackendManifestsMissingDir(t *testing.T) {
	d := t.TempDir()
	backendManifests, err := LoadBackendManifests([]string{path.Join(d, "does-not-exist")})
	if assert.NoError(t, err) {
		assert.Equal(t, []BackendManifestV1{}, backendManifests)
	}
}

func TestLoadBackendManifestsMinimalFields(t *testing.T) {
	d := t.TempDir()
	p := path.Join(d, "backend.yaml")
	err := os.WriteFile(p,
		[]byte(testutils.DedentYaml(`
			version: 1
			name: backend
			bin: /usr/bin/my-backend
			protocol-version: 1
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	backendManifests, err := LoadBackendManifests([]string{d})
	if assert.NoError(t, err) {
		assert.Equal(t, []BackendManifestV1{
			{
				path:            p,
				Version:         1,
				Name:            "backend",
				Bin:             "/usr/bin/my-backend",
				ProtocolVersion: 1,
			},
		}, backendManifests)
	}
}

func TestLoadBackendManifestsInvalidEmptyFile(t *testing.T) {
	d := t.TempDir()
	err := os.WriteFile(path.Join(d, "backend.yaml"), []byte(""), 0o644)
	if !assert.NoError(t, err) {
		return
	}
	_, err = LoadBackendManifests([]string{d})
	assert.Error(t, err)
}

func TestLoadBackendManifestsInvalidBadVersion(t *testing.T) {
	d := t.TempDir()
	err := os.WriteFile(
		path.Join(d, "backend.yaml"),
		[]byte(testutils.DedentYaml(`
			version: -1
			name: backend
			bin: /usr/bin/backend
			protocol-version: 1
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	_, err = LoadBackendManifests([]string{d})
	assert.Error(t, err)
}

func TestLoadBackendManifestsInvalidBadProtocolVersion(t *testing.T) {
	d := t.TempDir()
	err := os.WriteFile(
		path.Join(d, "backend.yaml"),
		[]byte(testutils.DedentYaml(`
			version: 1
			name: backend
			bin: /usr/bin/backend
			protocol-version: -1
		`)),
		0o644,
	)
	if !assert.NoError(t, err) {
		return
	}
	_, err = LoadBackendManifests([]string{d})
	assert.Error(t, err)
}
