package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWorkspaceLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml": `
name: lifecycle-test
provider: docker
services:
  web:
    command: "python3 -m http.server 28080"
    port: 28080
  api:
    command: "python3 -m http.server 23000"
    port: 23000
    depends_on: ["web"]
`,
		".vendetta/hooks/up.sh": `#!/bin/bash
echo "Starting test services..."
timeout 60 /usr/bin/python3 -m http.server -b 0.0.0.0 18080 &
timeout 60 /usr/bin/python3 -m http.server -b 0.0.0.0 13000 &
echo "Services started"
`,
	})

	require.NoError(t, os.Chmod(filepath.Join(projectDir, ".vendetta/hooks/up.sh"), 0755))

	binaryPath := env.BuildvendettaBinary(t)
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")

	require.NoError(t, os.WriteFile(filepath.Join(projectDir, ".vendetta/config.yaml"), []byte(`
name: lifecycle-test
provider: docker
services:
  db:
    command: "docker-compose up postgres"
  api:
    command: "npm start"
    depends_on: ["db"]
  web:
    command: "npm start"
    depends_on: ["api"]
`), 0644))

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "lifecycle-test")

	worktreePath := filepath.Join(projectDir, ".vendetta", "worktrees", "lifecycle-test")
	_, err := os.Stat(worktreePath)
	require.NoError(t, err)

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "up", "lifecycle-test")
	time.Sleep(3 * time.Second)

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "down", "lifecycle-test")
	// down stops the container, rm removes the worktree
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "lifecycle-test")

	_, err = os.Stat(worktreePath)
	require.Error(t, err)
}

func TestWorkspaceList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml": `
name: list-test
provider: docker
services:
  app:
    command: "sleep infinity"
    port: 3000
`,
	})

	binaryPath := env.BuildvendettaBinary(t)
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "ws1")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "ws2")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "ws3")

	output := env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "list")

	for _, ws := range []string{"ws1", "ws2", "ws3"} {
		if !strings.Contains(output, ws) {
			t.Errorf("Expected workspace %s in list output", ws)
		}
	}

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "ws1")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "ws2")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "ws3")
}

func TestPluginSystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/hooks/up.sh": `#!/bin/bash
echo "Starting test environment..."
wait
`,
	})

	require.NoError(t, os.Chmod(filepath.Join(projectDir, ".vendetta/hooks/up.sh"), 0755))

	binaryPath := env.BuildvendettaBinary(t)
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")

	require.NoError(t, os.WriteFile(filepath.Join(projectDir, ".vendetta/config.yaml"), []byte(`
name: plugin-test
provider: docker
agents:
  - name: "cursor"
    enabled: true
`), 0644))

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "plugin-test")

	worktreePath := filepath.Join(projectDir, ".vendetta", "worktrees", "plugin-test")
	agentConfigPath := filepath.Join(worktreePath, "AGENTS.md")
	require.FileExists(t, agentConfigPath, "AGENTS.md should exist")

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "plugin-test")
}

func TestLXCProvider(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping LXC test in short mode")
	}

	if os.Getenv("LXC_TEST") == "" {
		t.Skip("Skipping LXC test - set LXC_TEST=1 to run")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml": `
name: lxc-test
provider: lxc
services:
  app:
    command: "sleep infinity"
    port: 3000
lxc:
  image: ubuntu:22.04
`,
		".vendetta/hooks/up.sh": `#!/bin/bash
echo "Starting LXC test environment..."
wait
`,
	})

	require.NoError(t, os.Chmod(filepath.Join(projectDir, ".vendetta/hooks/up.sh"), 0755))

	binaryPath := env.BuildvendettaBinary(t)
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "lxc-ws")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "up", "lxc-ws")
	time.Sleep(2 * time.Second)
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "down", "lxc-ws")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "lxc-ws")
}

func TestDockerProvider(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml": `
name: docker-test
provider: docker
services:
  app:
    command: "sleep infinity"
    port: 3000
`,
		".vendetta/hooks/up.sh": `#!/bin/bash
echo "Starting Docker test environment..."
wait
`,
	})

	require.NoError(t, os.Chmod(filepath.Join(projectDir, ".vendetta/hooks/up.sh"), 0755))

	binaryPath := env.BuildvendettaBinary(t)
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "docker-ws")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "up", "docker-ws")
	time.Sleep(2 * time.Second)
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "down", "docker-ws")
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "rm", "docker-ws")
}

func TestErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml": `
name: error-test
provider: docker
services:
  failing:
    command: "exit 1"
    port: 5000
`,
		".vendetta/hooks/up.sh": `#!/bin/bash
echo "Starting test environment..."
wait
`,
	})

	require.NoError(t, os.Chmod(filepath.Join(projectDir, ".vendetta/hooks/up.sh"), 0755))

	binaryPath := env.BuildvendettaBinary(t)

	t.Log("Testing invalid workspace name...")
	output, err := env.RunvendettaCommandWithError(binaryPath, projectDir, "workspace", "create", "invalid/name")
	require.Error(t, err)
	require.Contains(t, output, "invalid")

	t.Log("Testing stop of non-existent workspace...")
	output, err = env.RunvendettaCommandWithError(binaryPath, projectDir, "workspace", "down", "nonexistent")
	require.Error(t, err)
	require.Contains(t, output, "not found")

	t.Log("Testing duplicate workspace creation...")
	_, err = env.RunvendettaCommandWithError(binaryPath, projectDir, "workspace", "create", "test-ws")
	require.Error(t, err)
}

func TestPerformanceBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := NewTestEnvironment(t)
	defer env.Cleanup()

	projectDir := env.CreateTestProject(t, map[string]string{
		".vendetta/config.yaml": `
name: perf-test
provider: docker
services:
  quick:
    command: "echo 'Ready' && sleep 1"
    port: 6000
`,
		".vendetta/hooks/up.sh": `#!/bin/bash
echo "Starting test environment..."
wait
`,
	})

	require.NoError(t, os.Chmod(filepath.Join(projectDir, ".vendetta/hooks/up.sh"), 0755))

	binaryPath := env.BuildvendettaBinary(t)
	env.RunvendettaCommand(t, binaryPath, projectDir, "init")

	start := time.Now()
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "create", "perf-test")
	createTime := time.Since(start)
	t.Logf("Workspace creation time: %v", createTime)
	require.Less(t, createTime, 30*time.Second, "Workspace creation should complete within 30 seconds")

	start = time.Now()
	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "up", "perf-test")
	startupTime := time.Since(start)
	t.Logf("Workspace startup time: %v", startupTime)
	require.Less(t, startupTime, 120*time.Second, "Workspace startup should complete within 120 seconds")

	env.RunvendettaCommand(t, binaryPath, projectDir, "workspace", "down", "perf-test")
}
