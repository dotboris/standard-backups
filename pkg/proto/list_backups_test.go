package proto

import (
	"errors"
	"testing"

	"github.com/dotboris/standard-backups/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var testListBackupsReqFull = &ListBackupsRequest{
	RawOptions: map[string]any{
		"key":    "value",
		"bool":   true,
		"number": 42.0,
		"array":  []any{1.0, 2.0, 3.0},
		"object": map[string]any{"foo": "bar"},
	},
	DestinationName: "my-dest",
	VariantName:     "my-variant",
}

var testListBackupsReqMinimal = &ListBackupsRequest{
	DestinationName: "my-dest",
}

func TestListBackupsE2E(t *testing.T) {
	testCases := map[string]*ListBackupsRequest{
		"full":    testListBackupsReqFull,
		"minimal": testListBackupsReqMinimal,
	}
	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			testutils.SetEnvFromToEnv(t, req)
			var (
				gotReq *ListBackupsRequest
				called bool
			)
			b := &BackendImpl{
				ListBackups: func(r *ListBackupsRequest) (*ListBackupsResponse, error) {
					called = true
					gotReq = r
					return nil, nil
				},
			}
			t.Setenv(COMMAND_ENV, "list-backups")
			err := b.execute()
			if assert.NoError(t, err) {
				assert.True(t, called, "ListBackups func was not called")
				assert.Equal(t, req, gotReq)
			}
		})
	}
}

func TestListBackupsError(t *testing.T) {
	t.Setenv(COMMAND_ENV, "list-backups")
	testutils.SetEnvFromToEnv(t, testListBackupsReqFull)
	expectedErr := errors.New("oops")
	b := &BackendImpl{
		ListBackups: func(r *ListBackupsRequest) (*ListBackupsResponse, error) {
			return nil, expectedErr
		},
	}
	err := b.execute()
	assert.ErrorIs(t, err, expectedErr)
}

func TestListBackupsNotImplemented(t *testing.T) {
	t.Setenv(COMMAND_ENV, "list-backups")
	testutils.SetEnvFromToEnv(t, testListBackupsReqFull)
	b := &BackendImpl{}
	err := b.execute()
	assert.EqualError(t, err, "unhandled command list-backups")
}
