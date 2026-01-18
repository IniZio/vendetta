package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectGHCLI(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "gh CLI installed",
			setupMock: func() {},
			wantErr:   false,
		},
		{
			name: "gh CLI found in PATH",
			setupMock: func() {
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			path, err := DetectGHCLI()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, path)
			}
		})
	}
}

func TestCheckAuthStatus(t *testing.T) {
	tests := []struct {
		name        string
		ghCLIPath   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid gh CLI path",
			ghCLIPath: "gh",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticated, err := CheckAuthStatus(tt.ghCLIPath)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.IsType(t, true, authenticated)
			}
		})
	}
}

func TestExecuteGHCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		checkFunc func(string)
	}{
		{
			name:    "simple gh api call",
			args:    []string{"api", "user", "--jq", ".login"},
			wantErr: false,
			checkFunc: func(output string) {
				// Should return something (either username or error from gh)
				assert.NotEmpty(t, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ExecuteGHCommand("gh", tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				if err == nil && tt.checkFunc != nil {
					tt.checkFunc(output)
				}
			}
		})
	}
}

func TestAuthenticateWithGH(t *testing.T) {
	t.Skip("Skipping interactive auth test")
}

func TestExecuteGHCommandFailure(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "invalid command",
			args:    []string{"invalid-command-xyz"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ExecuteGHCommand("gh", tt.args...)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}
