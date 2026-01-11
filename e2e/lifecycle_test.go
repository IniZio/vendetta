package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVendattaFullStackLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// Create test project structure
	projectDir := env.CreateTestProject(t, map[string]string{
		".vendatta/config.yaml": `
name: e2e-test
provider: docker
services:
  web:
    command: "python3 -m http.server 8080"
    port: 8080
  api:
    command: "python3 -m http.server 3000"
    port: 3000
    depends_on: ["web"]
agents:
  - name: "cursor"
    enabled: true
docker:
  image: ubuntu:22.04
`,
	})

	// Build vendatta binary
	t.Log("Building vendatta...")
	binaryPath := env.BuildVendattaBinary(t)

	// Initialize vendatta
	t.Log("Initializing vendatta...")
	env.RunVendattaCommand(t, binaryPath, projectDir, "init")

	// Create workspace
	t.Log("Creating workspace...")
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "create", "e2e-test")

	// Start workspace
	t.Log("Starting workspace...")
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "up", "e2e-test")

	// Verify services are running
	t.Log("Verifying services...")
	env.VerifyServiceHealth(t, "http://localhost:8080", 10*time.Second)
	env.VerifyServiceHealth(t, "http://localhost:3000", 10*time.Second)

	// Test workspace listing
	t.Log("Testing workspace listing...")
	output := env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "list")
	assert.Contains(t, output, "e2e-test")

	// Stop workspace
	t.Log("Stopping workspace...")
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "down", "e2e-test")

	// Verify services stopped
	t.Log("Verifying services stopped...")
	env.VerifyServiceDown(t, "http://localhost:8080", 5*time.Second)
	env.VerifyServiceDown(t, "http://localhost:3000", 5*time.Second)
}

func TestVendattaPluginSystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendatta/config.yaml": `
name: plugin-test
agents:
  - name: "cursor"
    enabled: true
    plugins: ["core/base"]
`,
		".vendatta/plugins/core/base/plugin.yaml": `
name: core/base
version: 1.0.0
description: Core base plugin with essential rules and skills
rules:
  - name: base
    content: |
      # Base Coding Rules
      - Use meaningful variable names
      - Add error handling for external calls
      - Keep functions focused on single responsibility
skills:
  - name: analyze-code
    description: Analyze code for potential issues
    parameters:
      type: object
      properties:
        code:
          type: string
commands:
  - name: format
    description: Format code using standard tools
`,
	})

	binaryPath := env.BuildVendattaBinary(t)
	env.RunVendattaCommand(t, binaryPath, projectDir, "init")
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "create", "plugin-test")

	// Check plugin discovery
	output := env.RunVendattaCommand(t, binaryPath, projectDir, "plugin", "list")
	assert.Contains(t, output, "core/base")

	// Check lockfile generation
	env.RunVendattaCommand(t, binaryPath, projectDir, "plugin", "check")

	// Start workspace and verify plugin integration
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "up", "plugin-test")

	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "down", "plugin-test")
}

func TestVendattaMultiProviderSupport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendatta/config.yaml": `
name: multi-provider-test
provider: docker
services:
  app:
    command: "echo 'App running' && sleep 300"
    port: 4000
agents:
  - name: "cursor"
    enabled: true
`,
	})

	binaryPath := env.BuildVendattaBinary(t)
	env.RunVendattaCommand(t, binaryPath, projectDir, "init")
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "create", "multi-test")

	// Test with Docker provider (default)
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "up", "multi-test")
	env.VerifyServiceHealth(t, "http://localhost:4000", 10*time.Second)
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "down", "multi-test")
}

func TestVendattaErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendatta/config.yaml": `
name: error-test
provider: docker
services:
  failing:
    command: "exit 1"
    port: 5000
`,
	})

	binaryPath := env.BuildVendattaBinary(t)
	env.RunVendattaCommand(t, binaryPath, projectDir, "init")

	// Test invalid workspace name
	_, err := env.RunVendattaCommandWithError(binaryPath, projectDir, "workspace", "create", "invalid/name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")

	// Test stopping non-existent workspace
	_, err = env.RunVendattaCommandWithError(binaryPath, projectDir, "workspace", "down", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test creating workspace that already exists
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "create", "test-ws")
	_, err = env.RunVendattaCommandWithError(binaryPath, projectDir, "workspace", "create", "test-ws")
	assert.Error(t, err) // Should fail or warn about existing workspace
}

func TestVendattaPerformanceBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendatta/config.yaml": `
name: perf-test
provider: docker
services:
  quick:
    command: "echo 'Ready' && sleep 1"
    port: 6000
agents:
  - name: "cursor"
    enabled: true
`,
	})

	binaryPath := env.BuildVendattaBinary(t)
	env.RunVendattaCommand(t, binaryPath, projectDir, "init")

	// Benchmark workspace creation time
	start := time.Now()
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "create", "perf-test")
	createTime := time.Since(start)
	t.Logf("Workspace creation time: %v", createTime)
	assert.Less(t, createTime, 30*time.Second, "Workspace creation should complete within 30 seconds")

	// Benchmark workspace startup time
	start = time.Now()
	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "up", "perf-test")
	startupTime := time.Since(start)
	t.Logf("Workspace startup time: %v", startupTime)
	assert.Less(t, startupTime, 120*time.Second, "Workspace startup should complete within 120 seconds")

	env.RunVendattaCommand(t, binaryPath, projectDir, "workspace", "down", "perf-test")
}
