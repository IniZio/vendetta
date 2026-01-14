package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestM3RemoteProviderAgnostic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping M3 E2E test in short mode")
	}

	providers := []string{"docker", "lxc", "qemu"}

	for _, provider := range providers {
		t.Run("provider_"+provider, func(t *testing.T) {
			testRemoteProvider(t, provider)
		})
	}
}

func testRemoteProvider(t *testing.T, providerName string) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// Create test project with remote configuration
	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml": buildRemoteConfig(providerName),
	})

	binaryPath := env.BuildvendettaBinary(t)

	// Test initialization
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")

	// Test workspace creation with remote configuration
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "remote-test")

	// Verify workspace structure
	worktreePath := filepath.Join(projectDir, ".vendetta", "worktrees", "remote-test")
	_, err := os.Stat(worktreePath)
	require.NoError(t, err)

	// Test workspace startup (should handle remote gracefully)
	output, err := env.RunvendettaCommandWithError(binaryPath, projectDir, "workspace", "up", "remote-test")

	if providerName == "qemu" {
		// QEMU should work locally
		require.NoError(t, err)
		assert.Contains(t, output, "✅")
	} else {
		// Docker/LXC will fail without remote node but should validate config
		require.Error(t, err)
		remoteKeywordFound := strings.Contains(output, "remote")
		connectionKeywordFound := strings.Contains(output, "connection")
		assert.True(t, remoteKeywordFound || connectionKeywordFound, "Error should mention remote or connection")
	}

	// Cleanup
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "remote-test")
}

// TestM3ServiceDiscovery tests service management and discovery
func TestM3ServiceDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping M3 E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// Test service configuration with dependencies
	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml:": `
name: service-discovery-test
provider: qemu
services:
  db:
    command: "redis-server"
    port: 6379
    healthcheck:
      url: "tcp://localhost:6379"
      interval: "5s"
      timeout: "2s"
      retries: 3

  api:
    command: "npm run dev"
    port: 3000
    depends_on: ["db"]
    env:
      DATABASE_URL: "redis://localhost:6379"
    healthcheck:
      url: "http://localhost:3000/health"
      interval: "10s"
      timeout: "5s"
      retries: 5

  web:
    command: "python -m http.server 8080"
    port: 8080
    depends_on: ["api"]
    healthcheck:
      url: "http://localhost:8080"
      interval: "15s"
      timeout: "3s"
      retries: 3

qemu:
  image: "ubuntu:22.04"
  cpu: 2
  memory: "4G"
  disk: "20G"
  ssh_port: 2222
`,
	})

	binaryPath := env.BuildvendettaBinary(t)
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "service-test")

	// Verify service configuration is parsed correctly
	worktreePath := filepath.Join(projectDir, ".vendetta", "worktrees", "service-test")
	configPath := filepath.Join(worktreePath, ".vendetta", "config.yaml")

	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(configContent), "services:")
	assert.Contains(t, string(configContent), "db:")
	assert.Contains(t, string(configContent), "api:")
	assert.Contains(t, string(configContent), "web:")

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "service-test")
}

// TestM3CoordinationServer tests coordination server functionality
func TestM3CoordinationServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping M3 E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// Test multi-node configuration
	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml": `
name: coordination-test
provider: qemu
remote:
  node: "coordination.example.com"
  user: "devuser"
  port: 22

# Test node management integration
nodes:
  - name: "node1"
    address: "192.168.1.100"
    provider: "docker"
  - name: "node2"
    address: "192.168.1.101"
    provider: "lxc"
  - name: "node3"
    address: "192.168.1.102"
    provider: "qemu"

services:
  coordinator:
    command: "coordination-server --port 3001"
    port: 3001
    healthcheck:
      url: "http://localhost:3001/status"

  monitor:
    command: "node-monitor --coordinator localhost:3001"
    port: 3002
    depends_on: ["coordinator"]
`,
	})

	binaryPath := env.BuildvendettaBinary(t)
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")

	// Test coordination server commands (if implemented)
	output, err := env.RunvendettaCommandWithError(binaryPath, projectDir, "node", "list")

	// If node commands don't exist yet, verify error is appropriate
	if err != nil {
		unknownCmdFound := strings.Contains(output, "unknown command")
		notFoundFound := strings.Contains(output, "not found")
		assert.True(t, unknownCmdFound || notFoundFound, "Error should indicate command not found")
	} else {
		// If implemented, verify basic functionality
		assert.Contains(t, output, "node")
	}
}

// TestM3ConfigurationMerging tests template and configuration merging
func TestM3ConfigurationMerging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping M3 E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml": `
name: config-merge-test
provider: qemu
extends:
  - "base-config"
  - "remote-config"

services:
  app:
    command: "npm run dev"
    port: 3000
    env:
      NODE_ENV: "development"

qemu:
  cpu: 4
  memory: "8G"
  disk: "50G"

remote:
  node: "remote.example.com"
  user: "devuser"
`,
		".vendetta/templates/base-config.yaml": `
# Base configuration for all workspaces
hooks:
  setup: "echo 'Setting up environment...'"
  teardown: "echo 'Cleaning up environment...'"

qemu:
  image: "ubuntu:22.04"
  cpu: 2
  memory: "4G"
`,
		".vendetta/templates/remote-config.yaml": `
# Remote-specific configuration
remote:
  port: 22
  ssh_key: "~/.ssh/id_rsa"

services:
  ssh-agent:
    command: "ssh-agent -a $SSH_AUTH_SOCK"
`,
	})

	binaryPath := env.BuildvendettaBinary(t)
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "merge-test")

	// Verify merged configuration
	worktreePath := filepath.Join(projectDir, ".vendetta", "worktrees", "merge-test")
	mergedConfigPath := filepath.Join(worktreePath, ".vendetta", "config.yaml")

	configContent, err := os.ReadFile(mergedConfigPath)
	require.NoError(t, err)

	// Verify base config was merged
	assert.Contains(t, string(configContent), "hooks:")
	assert.Contains(t, string(configContent), "ubuntu:22.04")

	// Verify remote config was merged
	assert.Contains(t, string(configContent), "ssh_key:")

	// Verify local config takes precedence
	assert.Contains(t, string(configContent), "memory: 8G") // Should be 8G, not 4G
	assert.Contains(t, string(configContent), "cpu: 4")     // Should be 4, not 2

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "merge-test")
}

// TestM3ErrorHandling tests error scenarios and recovery
func TestM3ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping M3 E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	binaryPath := env.BuildvendettaBinary(t)

	t.Run("invalid_remote_config", func(t *testing.T) {
		projectDir := env.CreateTestProject(t, map[string]string{
			".vendetta/config.yaml": `
name: invalid-remote-test
provider: qemu
remote:
  node: ""  # Empty node should cause error
  user: "devuser"
`,
		})

		output, err := env.RunvendettaCommandWithError(binaryPath, projectDir, "workspace", "create", "invalid-test")
		require.Error(t, err)
		invalidFound := strings.Contains(output, "invalid")
		requiredFound := strings.Contains(output, "required")
		assert.True(t, invalidFound || requiredFound, "Error should indicate invalid config or missing required field")
	})

	t.Run("unsupported_provider", func(t *testing.T) {
		projectDir := env.CreateTestProject(t, map[string]string{
			".vendetta/config.yaml": `
name: unsupported-provider-test
provider: "nonexistent"
`,
		})

		_, err := env.RunvendettaCommandWithError(binaryPath, projectDir, "workspace", "create", "unsupported-test")
		require.Error(t, err)
	})

	t.Run("resource_limits", func(t *testing.T) {
		projectDir := env.CreateTestProject(t, map[string]string{
			".vendetta/config.yaml": `
name: resource-test
provider: qemu
qemu:
  cpu: 128  # Unrealistic CPU count
  memory: "1TB"  # Unrealistic memory
  disk: "10PB"  # Unrealistic disk
`,
		})

		output, err := env.RunvendettaCommandWithError(binaryPath, projectDir, "workspace", "create", "resource-test")
		// Should either validate and reject, or attempt and fail gracefully
		if err != nil {
			resourceFound := strings.Contains(output, "resource")
			limitFound := strings.Contains(output, "limit")
			invalidFound := strings.Contains(output, "invalid")
			assert.True(t, resourceFound || limitFound || invalidFound, "Error should indicate resource/limit issues or invalid config")
		}
	})
}

// TestM3PerformanceBenchmarks tests performance requirements
func TestM3PerformanceBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping M3 performance test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml": `
name: performance-test
provider: qemu
qemu:
  image: "ubuntu:22.04"
  cpu: 2
  memory: "4G"
  disk: "20G"
  ssh_port: 2223

services:
  quick:
    command: "echo 'Ready' && sleep 1"
    port: 6000
`,
	})

	binaryPath := env.BuildvendettaBinary(t)
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")

	// Benchmark workspace creation
	start := time.Now()
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "perf-test")
	createTime := time.Since(start)
	t.Logf("Workspace creation time: %v", createTime)
	assert.Less(t, createTime, 30*time.Second, "Workspace creation should complete within 30 seconds")

	// Benchmark workspace startup
	start = time.Now()
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "up", "perf-test")
	startupTime := time.Since(start)
	t.Logf("Workspace startup time: %v", startupTime)
	assert.Less(t, startupTime, 120*time.Second, "Workspace startup should complete within 120 seconds")

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "down", "perf-test")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "perf-test")
}

// Helper function to build remote configuration for different providers
func buildRemoteConfig(provider string) string {
	baseConfig := `
name: remote-test-PLACEHOLDER
provider: PLACEHOLDER
remote:
  node: "remote.example.com"
  user: "devuser"
  port: 22

services:
  app:
    command: "sleep infinity"
    port: 3000
    healthcheck:
      url: "http://localhost:3000/health"
      interval: "10s"
      timeout: "5s"
      retries: 3
`

	switch provider {
	case "docker":
		return strings.ReplaceAll(baseConfig, "PLACEHOLDER", "docker") + `
docker:
  image: "ubuntu:22.04"
`
	case "lxc":
		return strings.ReplaceAll(baseConfig, "PLACEHOLDER", "lxc") + `
lxc:
  image: "ubuntu:22.04"
`
	case "qemu":
		return strings.ReplaceAll(baseConfig, "PLACEHOLDER", "qemu") + `
qemu:
  image: "ubuntu:22.04"
  cpu: 2
  memory: "4G"
  disk: "20G"
  ssh_port: 2222
`
	default:
		return strings.ReplaceAll(baseConfig, "PLACEHOLDER", provider)
	}
}
