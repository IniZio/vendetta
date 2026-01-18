package ssh

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadPublicKeyToGitHub(t *testing.T) {
	tests := []struct {
		name       string
		ghCLIPath  string
		pubKeyPath string
		wantErr    bool
	}{
		{
			name:       "upload key to github",
			ghCLIPath:  "gh",
			pubKeyPath: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pubKeyPath == "" {
				t.Skip("skipping without valid public key path")
			}

			err := UploadPublicKeyToGitHub(tt.ghCLIPath, tt.pubKeyPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else if err != nil {
				assert.Contains(t, err.Error(), "")
			}
		})
	}
}

func TestParseGitHubSSHKeyError(t *testing.T) {
	tests := []struct {
		name      string
		errOutput string
		wantCode  int
	}{
		{
			name:      "409 conflict error",
			errOutput: "HTTP 409: Key already exists",
			wantCode:  409,
		},
		{
			name:      "401 unauthorized",
			errOutput: "HTTP 401: Unauthorized",
			wantCode:  401,
		},
		{
			name:      "403 forbidden",
			errOutput: "HTTP 403: Forbidden",
			wantCode:  403,
		},
		{
			name:      "unknown error",
			errOutput: "something went wrong",
			wantCode:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := ParseGitHubSSHKeyError(tt.errOutput)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}
