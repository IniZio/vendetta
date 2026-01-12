package lxc

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibegear/vendatta/pkg/provider"
)

func TestNewLXCProvider(t *testing.T) {
	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	assert.NotNil(t, lxcProvider)
	assert.Equal(t, "lxc", lxcProvider.Name())
}

func TestLXCProvider_Name(t *testing.T) {
	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	assert.Equal(t, "lxc", lxcProvider.Name())
}

func TestLXCProvider_Lifecycle(t *testing.T) {
	lxcProvider, err := NewLXCProvider()
	if err != nil {
		t.Skip("LXC not available:", err)
	}

	ctx := context.Background()
	sessionID := "test-lxc-lifecycle"

	// 1. Create temporary workspace directory
	tmpDir, err := os.MkdirTemp("", "vendatta-lxc-test-workspace")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a canary file in the workspace
	canaryPath := filepath.Join(tmpDir, "canary.txt")
	err = os.WriteFile(canaryPath, []byte("ok"), 0644)
	require.NoError(t, err)

	// 2. Create the container
	session, err := lxcProvider.Create(ctx, sessionID, tmpDir, nil)
	require.NoError(t, err)
	require.Equal(t, "running", session.Status)

	// 3. Verify mount inside the container
	var stdout bytes.Buffer
	execOpts := provider.ExecOptions{
		Cmd:          []string{"ls", "-al", "/workspace/canary.txt"},
		Stdout:       true,
		StdoutWriter: &stdout,
		Stderr:       true,
		StderrWriter: &stdout, // Capture stderr to stdout for easier debugging
	}
	err = lxcProvider.Exec(ctx, sessionID, execOpts)
	// We use strings.Contains instead of require.NoError because the error output of 'lxc exec' can be complex.
	// We check the output for 'canary.txt' and assert it's a success.
	require.NoError(t, err, "Failed to exec inside container to verify mount. Output: %s", stdout.String())
	assert.Contains(t, stdout.String(), "canary.txt", "Canary file not found in /workspace mount")

	// 4. Destroy the container
	err = lxcProvider.Destroy(ctx, sessionID)
	require.NoError(t, err)
}
