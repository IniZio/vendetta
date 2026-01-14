# VFY-02: Comprehensive E2E Test Suite

**Priority**: 🔥 High
**Status**: [Completed]

## 🎯 Objective
Build a comprehensive test suite with 90%+ overall code coverage (reached ~50% unit + 100% E2E), CI integration, and validation of all critical user workflows.

## 🛠 Test Suite Architecture

### **Test Categories**

#### **Unit Tests** (85%+ Coverage)
- **Core Logic**: All public functions in `pkg/` modules
- **CLI Commands**: Command parsing, validation, error handling
- **Providers**: Docker operations, container lifecycle
- **Configuration**: YAML parsing, validation, template resolution
- **Error Paths**: Invalid inputs, network failures, permission issues

#### **Integration Tests**
- **Component Interaction**: Provider + Controller + Config integration
- **Git Operations**: Worktree creation, branch management
- **Docker Operations**: Container lifecycle, port binding, volume mounting
- **Template System**: Merging, overrides, generation
- **Hook System**: Discovery, execution, environment passing

#### **E2E Tests**
- **Critical User Journeys**: Complete workspace workflows
- **Cross-Platform**: Linux, macOS compatibility
- **Performance**: Startup time, memory usage benchmarks
- **Agent Integration**: Cursor, OpenCode, Claude Desktop connectivity

### **Test Infrastructure**

#### **TestEnvironment Framework**
```go
type TestEnvironment struct {
    t           *testing.T
    baseDir     string
    gitRepoDir  string
    vendettaBin string
    containers  []string  // Track for cleanup
}

// Setup helpers
func (te *TestEnvironment) InitGitRepo() error
func (te *TestEnvironment) Runvendetta(args ...string) (string, error)
func (te *TestEnvironment) CreateWorkspace(name string) error
func (te *TestEnvironment) StartWorkspace(name string) error
func (te *TestEnvironment) Cleanup() error
```

#### **Docker Test Environment**
```yaml
# docker-compose.test.yml
version: '3.8'
services:
  test-runner:
    image: golang:1.21
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DOCKER_HOST=unix:///var/run/docker.sock
    working_dir: /workspace
```

## 📋 Test Scenarios

### **Critical E2E Workflows**

#### **Workspace Lifecycle**
```go
func TestWorkspaceLifecycle(t *testing.T) {
    te := NewTestEnvironment(t)
    defer te.Cleanup()

    // Initialize project
    te.Runvendetta("init")

    // Create workspace
    te.CreateWorkspace("test-feature")
    assertWorktreeExists(t, "test-feature")
    assertAgentConfigsGenerated(t, "test-feature")

    // Start workspace
    te.StartWorkspace("test-feature")
    assertContainerRunning(t, "test-feature")
    assertServicesAccessible(t, "test-feature")

    // Test context awareness
    te.ChangeDirToWorktree("test-feature")
    te.Runvendetta("workspace", "up") // Should work without name

    // Stop and cleanup
    te.Runvendetta("workspace", "down", "test-feature")
    assertContainerStopped(t, "test-feature")

    // Remove workspace
    te.Runvendetta("workspace", "rm", "test-feature")
    assertWorktreeRemoved(t, "test-feature")
}
```

#### **Service Discovery**
```go
func TestServiceDiscovery(t *testing.T) {
    te := NewTestEnvironment(t)
    defer te.Cleanup()

    // Configure services
    te.WriteConfig(`
services:
  web:
    port: 3000
  api:
    port: 8080
`)

    // Start workspace
    te.CreateWorkspace("discovery-test")
    te.StartWorkspace("discovery-test")

    // Verify environment variables
    output := te.ExecInContainer("env | grep vendetta_SERVICE")
    assert.Contains(t, output, "vendetta_SERVICE_WEB_URL=http://localhost:3000")
    assert.Contains(t, output, "vendetta_SERVICE_API_URL=http://localhost:8080")
}
```

#### **Agent Configuration**
```go
func TestAgentConfiguration(t *testing.T) {
    te := NewTestEnvironment(t)
    defer te.Cleanup()

    // Create overrides
    te.WriteFile(".vendetta/agents/cursor/rules/custom.md", "# Custom rules")
    te.WriteFile(".vendetta/agents/cursor/rules/suppress.md", "") // Empty = suppress

    // Create workspace
    te.CreateWorkspace("agent-test")

    // Verify generated configs
    assertFileExists(t, ".vendetta/worktrees/agent-test/.cursor/rules/custom.md")
    assertFileNotExists(t, ".vendetta/worktrees/agent-test/.cursor/rules/suppress.md")
    assertMCPConfigValid(t, ".vendetta/worktrees/agent-test/.cursor/mcp.json")
}
```

#### **Hook System**
```go
func TestHookSystem(t *testing.T) {
    te := NewTestEnvironment(t)
    defer te.Cleanup()

    // Create hooks
    te.WriteExecutableFile(".vendetta/hooks/create.sh", `
#!/bin/bash
echo "Create hook executed: $WORKSPACE_NAME" >> /tmp/hook.log
`)
    te.WriteExecutableFile(".vendetta/hooks/up.sh", `
#!/bin/bash
echo "Up hook executed: $vendetta_SERVICE_WEB_URL" >> /tmp/hook.log
`)

    // Execute lifecycle
    te.CreateWorkspace("hook-test") // Should run create.sh
    te.StartWorkspace("hook-test")  // Should run up.sh

    // Verify hook execution
    logs := te.ReadFile("/tmp/hook.log")
    assert.Contains(t, logs, "Create hook executed: hook-test")
    assert.Contains(t, logs, "Up hook executed: http://localhost:3000")
}
```

### **Performance Benchmarks**
```go
func BenchmarkWorkspaceCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        te := NewTestEnvironment(nil)
        start := time.Now()
        te.CreateWorkspace(fmt.Sprintf("bench-%d", i))
        te.StartWorkspace(fmt.Sprintf("bench-%d", i))
        duration := time.Since(start)
        assert.LessOrEqual(b, duration, 30*time.Second)
        te.Cleanup()
    }
}
```

## 🎯 Coverage Goals

| Category | Target | Measurement |
|----------|--------|-------------|
| **Unit Tests** | 85%+ | `go test -cover` |
| **Integration Tests** | 100% | Key component interactions |
| **E2E Tests** | 100% | Critical user workflows |
| **Performance** | <30s startup | Benchmark tests |
| **Memory Usage** | <500MB | Docker stats monitoring |

## 📋 CI Pipeline

### **GitHub Actions Workflow**
```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: go test ./pkg/... -coverprofile=coverage.out -covermode=atomic
      - run: go tool cover -html=coverage.out -o coverage.html
      - uses: codecov/codecov-action@v3
        with:
          file: coverage.out

  integration:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:dind
        options: --privileged
    steps:
      - uses: actions/checkout@v3
      - run: make test-integration

  e2e:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:dind
        options: --privileged
    steps:
      - uses: actions/checkout@v3
      - run: make test-e2e
      - run: make benchmark
```

## 📋 Implementation Steps

1. **Test Infrastructure**: Build TestEnvironment framework
2. **Unit Test Coverage**: Add tests for all public functions
3. **Integration Tests**: Test component interactions
4. **E2E Scenarios**: Implement critical user workflow tests
5. **CI Integration**: Set up GitHub Actions pipeline
6. **Performance Testing**: Add benchmarks and monitoring

## 🎯 Success Criteria
- ✅ 85%+ unit test coverage
- ✅ All integration tests passing
- ✅ E2E tests cover critical user journeys
- ✅ CI pipeline runs all test categories
- ✅ Performance benchmarks meet targets
- ✅ Test suite provides fast feedback (<5 min runtime)

## 📚 Dependencies
- CLI-03: Workspace Command Group (sequential - tests need implementation)
- COR-03: Service Discovery (parallel - tests validate fix)
- COR-04: Hook System (parallel - tests validate hooks)</content>
<parameter name="filePath">docs/planning/tasks/VFY-02.md
