package e2e

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvironment provides a testing harness for E2E tests
type TestEnvironment struct {
	t          *testing.T
	tempDir    string
	binaryPath string
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	tempDir, err := os.MkdirTemp("", "vendatta-e2e-*")
	require.NoError(t, err)

	return &TestEnvironment{
		t:       t,
		tempDir: tempDir,
	}
}

// Cleanup removes the test environment
func (env *TestEnvironment) Cleanup() {
	// Clean up docker containers
	cmd := exec.Command("docker", "ps", "-q", "--filter", "label=oursky.session.id")
	if output, err := cmd.Output(); err == nil {
		containerIDs := strings.Fields(string(output))
		for _, id := range containerIDs {
			exec.Command("docker", "rm", "-f", id).Run()
		}
	}

	if env.tempDir != "" {
		os.RemoveAll(env.tempDir)
	}
}

// CreateTestProject creates a test project with the given files
func (env *TestEnvironment) CreateTestProject(t *testing.T, files map[string]string) string {
	projectDir := filepath.Join(env.tempDir, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	for path, content := range files {
		fullPath := filepath.Join(projectDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = projectDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = projectDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = projectDir
	require.NoError(t, cmd.Run())

	// Create initial commit
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = projectDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = projectDir
	require.NoError(t, cmd.Run())

	return projectDir
}

// BuildVendattaBinary builds the vendatta binary for testing
func (env *TestEnvironment) BuildVendattaBinary(t *testing.T) string {
	if env.binaryPath != "" {
		return env.binaryPath
	}

	binaryPath := filepath.Join(env.tempDir, "vendatta")
	cmd := exec.Command("go", "build", "-o", binaryPath, "cmd/oursky/main.go")
	cmd.Dir = ".."
	require.NoError(t, cmd.Run())

	env.binaryPath = binaryPath
	return binaryPath
}

// RunVendattaCommand runs a vendatta command and returns the output
func (env *TestEnvironment) RunVendattaCommand(t *testing.T, binaryPath, projectDir string, args ...string) string {
	output, err := env.RunVendattaCommandWithError(binaryPath, projectDir, args...)
	if err != nil {
		t.Logf("Command failed: %s %v in %s", binaryPath, args, projectDir)
		t.Logf("Output: %s", output)
	}
	require.NoError(t, err)
	return output
}

// RunVendattaCommandWithError runs a vendatta command and returns output and error
func (env *TestEnvironment) RunVendattaCommandWithError(binaryPath, projectDir string, args ...string) (string, error) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = projectDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String() + stderr.String()
	return output, err
}

// VerifyServiceHealth checks if a service is healthy
func (env *TestEnvironment) VerifyServiceHealth(t *testing.T, url string, timeout time.Duration) {
	client := &http.Client{Timeout: 5 * time.Second}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}

	t.Fatalf("Service at %s did not become healthy within %v", url, timeout)
}

// VerifyServiceDown checks if a service is down
func (env *TestEnvironment) VerifyServiceDown(t *testing.T, url string, timeout time.Duration) {
	client := &http.Client{Timeout: 5 * time.Second}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err != nil {
			return // Service is down
		}
		resp.Body.Close()
		if resp.StatusCode >= 500 {
			return // Service error means it's down
		}
		time.Sleep(500 * time.Millisecond)
	}

	t.Fatalf("Service at %s did not go down within %v", url, timeout)
}

// VerifyFileExists checks if a file exists
func (env *TestEnvironment) VerifyFileExists(t *testing.T, path string) {
	_, err := os.Stat(path)
	assert.NoError(t, err, "File should exist: %s", path)
}
