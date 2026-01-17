# mochi Test Plan Specification

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Test Scope](#2-test-scope)
3. [Test Environment Requirements](#3-test-environment-requirements)
4. [Test Case Specifications](#4-test-case-specifications)
   - [CLI Commands](#41-cli-commands)
   - [Workspace Management](#42-workspace-management)
   - [Service Discovery](#43-service-discovery)
   - [Agent Configuration](#44-agent-configuration)
   - [Plugin System](#45-plugin-system)
   - [Lockfile Management](#46-lockfile-management)
   - [Hook System](#47-hook-system)
   - [Docker Provider](#48-docker-provider)
   - [Error Handling](#49-error-handling)
   - [Performance Tests](#410-performance-tests)
   - [Security Tests](#411-security-tests)
5. [Test Data Management](#5-test-data-management)
6. [Test Automation Framework](#6-test-automation-framework)
7. [CI/CD Integration](#7-cicd-integration)
8. [Risk Analysis](#8-risk-analysis)
9. [Appendix](#9-appendix)

---

## 1. Executive Summary

This document provides a comprehensive test plan specification for the mochi project. It defines detailed test cases for all functional areas, organized by module and priority. The test plan aims to achieve:

- **Unit Test Coverage**: 85%+ for all logic packages
- **Integration Test Coverage**: 100% for component interactions
- **E2E Test Coverage**: 100% for critical user workflows
- **Performance Targets**: Workspace creation <30s, startup <60s
- **Security Validation**: No critical vulnerabilities

---

## 2. Test Scope

### 2.1 In Scope

| Category | Modules | Priority |
|----------|---------|----------|
| CLI Commands | All command groups (init, workspace, config, plugin, remote) | High |
| Core Logic | Configuration parsing, lifecycle orchestration | High |
| Providers | Docker provider implementation | High |
| Integration | Git worktree, container lifecycle, port mapping | High |
| Agent Config | Template generation, overrides, suppression | High |
| Plugin System | Discovery, DAG resolution, cycle detection | High |
| Lockfile | Generation, verification, offline mode | High |
| Hook System | Discovery, execution, environment injection | Medium |

### 2.2 Out of Scope

| Category | Reason |
|----------|--------|
| LXC Provider | M2 Alpha feature (future) |
| QEMU Provider | M3 Beta feature (future) |
| Multi-Machine | M3 Beta feature (future) |
| Third-party agent internals | Outside mochi scope |

---

## 3. Test Environment Requirements

### 3.1 Hardware Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 2 cores | 4+ cores |
| Memory | 4 GB | 8+ GB |
| Disk | 10 GB free | 50+ GB free |
| Docker | Docker 20.10+ | Docker 24+ |

### 3.2 Software Requirements

| Software | Version | Required |
|----------|---------|----------|
| Go | 1.24+ | Yes |
| Git | 2.0+ | Yes |
| Docker | 20.10+ | Yes |
| Docker Compose | 2.0+ | Yes |

### 3.3 Test Fixtures Directory Structure

```
internal/testfixtures/
├── e2e/
│   ├── repos/
│   │   ├── empty-repo/
│   │   ├── multi-branch-repo/
│   │   └── with-dockerfile/
│   ├── configs/
│   │   ├── minimal.yaml
│   │   ├── full-stack.yaml
│   │   └── with-plugins.yaml
│   ├── plugins/
│   │   ├── valid-plugin/
│   │   │   ├── plugin.yaml
│   │   │   ├── rules/
│   │   │   └── skills/
│   │   └── invalid-plugin/
│   └── templates/
│       ├── skills/
│       ├── commands/
│       └── rules/
```

---

## 4. Test Case Specifications

### 4.1 CLI Commands

#### TC-CLI-001: Initialize Empty Project

| Attribute | Value |
|-----------|-------|
| Test ID | TC-CLI-001 |
| Module | CLI / Init |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- Git repository initialized
- No `.mochi` directory exists
- Docker daemon running

**Test Steps:**
1. Run `mochi init`
2. Verify `.mochi/` directory created
3. Verify `config.yaml` created with valid structure
4. Verify default templates directory created

**Expected Results:**
- `.mochi/` directory exists
- `config.yaml` contains required fields (name, services)
- `templates/` directory contains base templates
- Command exits with code 0

**Test Data:**
```yaml
# config.yaml expected structure
name: "test-project"
services: {}
plugins: []
```

**Automation:**
```go
func TestmochiInit(t *testing.T) {
    te := NewTestEnvironment(t)
    defer te.Cleanup()

    output := te.Runmochi("init")
    assert.Contains(t, output, "Initialized")

    assert.True(t, te.DirExists(".mochi"))
    assert.True(t, te.FileExists(".mochi/config.yaml"))

    config := te.ReadConfig(".mochi/config.yaml")
    assert.NotEmpty(t, config["name"])
}
```

---

#### TC-CLI-002: Invalid Workspace Name Rejection

| Attribute | Value |
|-----------|-------|
| Test ID | TC-CLI-002 |
| Module | CLI / Validation |
| Priority | P1 - High |
| Type | Unit |

**Preconditions:**
- None

**Test Steps:**
1. Attempt `mochi workspace create invalid@name`
2. Attempt `mochi workspace create "name with spaces"`
3. Attempt `mochi workspace create .`
4. Attempt `mochi workspace create ..`

**Expected Results:**
- All commands fail with validation error
- Clear error message indicating valid name format
- Exit code non-zero

**Test Data:**
| Input | Expected Behavior |
|-------|-------------------|
| `invalid@name` | Error: Invalid characters |
| `name with spaces` | Error: Spaces not allowed |
| `.` | Error: Reserved name |
| `..` | Error: Reserved name |

---

#### TC-CLI-003: Workspace List Command

| Attribute | Value |
|-----------|-------|
| Test ID | TC-CLI-003 |
| Module | CLI / Workspace |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multiple workspaces created (ws-1, ws-2, ws-3)
- Some workspaces running, some stopped

**Test Steps:**
1. Run `mochi workspace list`
2. Verify output shows all workspaces
3. Verify status indicators (running/stopped)
4. Verify port mappings displayed

**Expected Results:**
- All 3 workspaces listed
- Running workspaces marked as "Up"
- Stopped workspaces marked as "Down"
- Port mappings visible for running workspaces

**Automation:**
```go
func TestWorkspaceList(t *testing.T) {
    te := NewTestEnvironment(t)
    defer te.Cleanup()

    te.CreateWorkspace("ws-1")
    te.CreateWorkspace("ws-2")
    te.StartWorkspace("ws-1")

    output := te.Runmochi("workspace", "list")

    assert.Contains(t, output, "ws-1")
    assert.Contains(t, output, "ws-2")
    assert.Contains(t, output, "Up")
    assert.Contains(t, output, "Down")
}
```

---

### 4.2 Workspace Management

#### TC-WS-001: Complete Workspace Lifecycle

| Attribute | Value |
|-----------|-------|
| Test ID | TC-WS-001 |
| Module | Workspace |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- Git repository with main branch
- Docker daemon running
- No existing workspaces

**Test Steps:**
1. `mochi workspace create feature-test`
2. Verify worktree created at `.mochi/worktrees/feature-test/`
3. Verify agent configs generated
4. `mochi workspace up feature-test`
5. Verify container running
6. `mochi workspace down feature-test`
7. Verify container stopped
8. `mochi workspace rm feature-test`
9. Verify worktree removed

**Expected Results:**
- All commands succeed
- Worktree created on new branch
- Agent configs in correct locations
- Container running after up
- Container stopped after down
- All resources cleaned after rm

**Test Data:**
```bash
# Expected directory structure after create
.mochi/
└── worktrees/
    └── feature-test/
        ├── .cursor/
        │   └── mcp.json
        ├── opencode.json
        └── [project files from git worktree]
```

---

#### TC-WS-002: Workspace Context Awareness

| Attribute | Value |
|-----------|-------|
| Test ID | TC-WS-002 |
| Module | Workspace |
| Priority | P1 - High |
| Type | E2E |

**Preconditions:**
- Workspace `context-test` created and started

**Test Steps:**
1. `cd .mochi/worktrees/context-test`
2. Run `mochi workspace up` (no name specified)
3. Verify workspace starts without error
4. Run `mochi workspace down` (no name specified)
5. Verify workspace stops without error

**Expected Results:**
- Commands work without explicit workspace name
- Workspace identified correctly from current directory

---

#### TC-WS-003: Concurrent Workspace Isolation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-WS-003 |
| Module | Workspace |
| Priority | P0 - Critical |
| Type | Integration |

**Preconditions:**
- Docker daemon running
- No existing workspaces

**Test Steps:**
1. Create workspace `ws-a`
2. Create workspace `ws-b`
3. Start both workspaces in parallel
4. Verify both containers running
5. Verify different port mappings
6. Verify different container names
7. Verify worktrees are isolated

**Expected Results:**
- Both workspaces start successfully
- No port conflicts
- Container names unique (`mochi-workspace-ws-a`, `mochi-workspace-ws-b`)
- Worktrees contain different branches

**Test Data:**
```go
func TestConcurrentWorkspaceIsolation(t *testing.T) {
    te := NewTestEnvironment(t)
    defer te.Cleanup()

    var wg sync.WaitGroup

    // Create both workspaces
    wg.Add(2)
    go func() { defer wg.Done(); te.CreateWorkspace("ws-a") }()
    go func() { defer wg.Done(); te.CreateWorkspace("ws-b") }()
    wg.Wait()

    // Start both workspaces
    wg.Add(2)
    go func() { defer wg.Done(); te.StartWorkspace("ws-a") }()
    go func() { defer wg.Done(); te.StartWorkspace("ws-b") }()
    wg.Wait()

    // Verify both running with different ports
    containers := te.ListContainers()
    assert.Len(t, containers, 2)

    wsA := te.GetWorkspace("ws-a")
    wsB := te.GetWorkspace("ws-b")
    assert.NotEqual(t, wsA.Ports, wsB.Ports)
}
```

---

### 4.3 Service Discovery

#### TC-SD-001: Environment Variable Injection

| Attribute | Value |
|-----------|-------|
| Test ID | TC-SD-001 |
| Module | Service Discovery |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- Docker daemon running
- Config with multiple services defined

**Test Steps:**
1. Create config with services (web:3000, api:8080, db:5432)
2. `mochi workspace create service-test`
3. `mochi workspace up service-test`
4. `mochi workspace shell service-test`
5. Execute `env | grep mochi_SERVICE`

**Expected Results:**
- Environment variables present:
  - `mochi_SERVICE_WEB_URL=http://localhost:3000`
  - `mochi_SERVICE_API_URL=http://localhost:8080`
  - `mochi_SERVICE_DB_URL=postgresql://localhost:5432`

**Test Data:**
```yaml
# .mochi/config.yaml
services:
  web:
    command: "cd client && npm run dev"
  api:
    command: "cd server && npm run dev"
  db:
    command: "docker-compose up -d postgres"
```

---

#### TC-SD-002: Port Auto-Detection

| Attribute | Value |
|-----------|-------|
| Test ID | TC-SD-002 |
| Module | Service Discovery |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Docker daemon running
- Services with various command formats

**Test Steps:**
1. Create config with different service command formats
2. Start workspace
3. Verify port detection works for all formats

**Expected Results:**
- Port 3000 detected from `npm run dev -- --port 3000`
- Port 8080 detected from `uvicorn app:app --host 0.0.0.0 --port 8080`
- Port 5432 detected from `docker-compose up -d postgres`

**Test Data:**
```yaml
services:
  web:
    command: "npm run dev"
    port: 3000  # Explicit or auto-detected
  api:
    command: "uvicorn app:app --port 8080"
    port: 8080
  db:
    command: "docker-compose up -d postgres"
    port: 5432
```

---

#### TC-SD-003: Dynamic Port Assignment

| Attribute | Value |
|-----------|-------|
| Test ID | TC-SD-003 |
| Module | Service Discovery |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Docker daemon running
- Services with conflicting default ports

**Test Steps:**
1. Create two workspaces with same service configuration
2. Start both workspaces
3. Verify different host ports assigned
4. Verify environment variables reflect actual ports

**Expected Results:**
- First workspace: `mochi_SERVICE_WEB_URL=http://localhost:3000`
- Second workspace: `mochi_SERVICE_WEB_URL=http://localhost:3001` (or similar)

---

### 4.4 Agent Configuration

#### TC-AGT-001: Multi-Agent Config Generation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-AGT-001 |
| Module | Agent Configuration |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- Multiple agents enabled in config
- Templates defined

**Test Steps:**
1. Configure agents (cursor, opencode, claude-desktop)
2. `mochi workspace create agent-test`
3. Verify `.cursor/mcp.json` generated
4. Verify `opencode.json` generated
5. Verify `.opencode/` directory created
6. Verify `claude_desktop_config.json` generated

**Expected Results:**
- All agent configs generated
- MCP connection details correct
- No conflicts between agent configs

**Test Data:**
```yaml
# .mochi/config.yaml
agents:
  cursor:
    enabled: true
  opencode:
    enabled: true
  claude-desktop:
    enabled: true
```

---

#### TC-AGT-002: Template Override Mechanism

| Attribute | Value |
|-----------|-------|
| Test ID | TC-AGT-002 |
| Module | Agent Configuration |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Base templates exist
- Override files defined

**Test Steps:**
1. Create override file `.mochi/agents/cursor/rules/custom.md`
2. Create suppression file `.mochi/agents/opencode/rules/legacy.md` (empty)
3. Create workspace
4. Verify override applied
5. Verify suppressed rule not generated

**Expected Results:**
- Custom rule present in generated configs
- Suppressed rule absent from generated configs
- All other base rules present

---

#### TC-AGT-003: MCP Config Generation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-AGT-003 |
| Module | Agent Configuration |
| Priority | P0 - Critical |
| Type | Integration |

**Preconditions:**
- MCP config templates defined
- Agents enabled

**Test Steps:**
1. Create workspace
2. Verify `.cursor/mcp.json` generated with server connection details
3. Verify `opencode.json` has `mcp` section with connection info
4. Verify `claude_desktop_config.json` generated with MCP remote config

**Expected Results:**
- MCP connection configs generated for all enabled agents
- Configs contain correct server URL and connection settings
- No conflicts between agent configs

---

### 4.5 Plugin System

#### TC-PLG-001: Plugin Discovery

| Attribute | Value |
|-----------|-------|
| Test ID | TC-PLG-001 |
| Module | Plugin System |
| Priority | P1 - High |
| Type | Unit |

**Preconditions:**
- Multiple plugins in various locations

**Test Steps:**
1. Place plugins in `.mochi/plugins/`
2. Place plugins in `.mochi/plugins/subdir/`
3. Run plugin discovery
4. Verify all plugins found

**Expected Results:**
- All 3 plugins discovered
- Plugin names correctly extracted from `plugin.yaml`

**Test Data:**
```
.mochi/plugins/
├── plugin-a/
│   └── plugin.yaml
└── subdir/
    └── plugin-b/
        └── plugin.yaml
```

---

#### TC-PLG-002: DAG Resolution and Loading Order

| Attribute | Value |
|-----------|-------|
| Test ID | TC-PLG-002 |
| Module | Plugin System |
| Priority | P0 - Critical |
| Type | Unit |

**Preconditions:**
- Plugins with dependencies defined

**Test Steps:**
1. Configure plugins with dependencies:
   - Plugin A depends on B
   - Plugin B depends on C
2. Run plugin resolution
3. Verify loading order (C, B, A)
4. Verify no cycles detected

**Expected Results:**
- Plugins loaded in correct dependency order
- No cycles detected
- All plugins loaded successfully

**Test Data:**
```yaml
# plugin-a/plugin.yaml
name: plugin-a
depends_on:
  - plugin-b

# plugin-b/plugin.yaml
name: plugin-b
depends_on:
  - plugin-c

# plugin-c/plugin.yaml
name: plugin-c
depends_on: []
```

---

#### TC-PLG-003: Cycle Detection

| Attribute | Value |
|-----------|-------|
| Test ID | TC-PLG-003 |
| Module | Plugin System |
| Priority | P0 - Critical |
| Type | Unit |

**Preconditions:**
- Plugins with circular dependency

**Test Steps:**
1. Configure plugins with cycle (A -> B -> C -> A)
2. Run plugin resolution
3. Verify error reported with cycle details

**Expected Results:**
- Error with clear cycle path (A -> B -> C -> A)
- No plugins loaded
- Exit code non-zero

---

#### TC-PLG-004: Parallel Plugin Fetching

| Attribute | Value |
|-----------|-------|
| Test ID | TC-PLG-004 |
| Module | Plugin System |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multiple remote plugins configured
- Network available

**Test Steps:**
1. Configure 5 remote plugins
2. Measure fetch time
3. Verify parallel execution (time < 5 * sequential time)

**Expected Results:**
- All plugins fetched
- Total time significantly less than sequential fetch
- No race conditions

---

### 4.6 Lockfile Management

#### TC-LCK-001: Lockfile Generation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-LCK-001 |
| Module | Lockfile |
| Priority | P0 - Critical |
| Type | Integration |

**Preconditions:**
- Plugins configured
- No lockfile exists

**Test Steps:**
1. `mochi plugin update`
2. Verify `mochi.lock` created
3. Verify lockfile contains all plugins with versions
4. Verify checksums present

**Expected Results:**
- Lockfile created
- All plugins listed with commit SHAs
- Content hash present
- Valid JSON structure

**Test Data:**
```json
{
  "version": 1,
  "metadata": {
    "content_hash": "sha256:..."
  },
  "plugins": [
    {
      "name": "mochi/standard",
      "url": "https://github.com/IniZio/laichi-config.git",
      "revision": "abc123..."
    }
  ]
}
```

---

#### TC-LCK-002: Deterministic Recreation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-LCK-002 |
| Module | Lockfile |
| Priority | P0 - Critical |
| Type | Integration |

**Preconditions:**
- Lockfile exists with plugins
- Plugin cache populated

**Test Steps:**
1. Delete all plugin caches
2. Run `mochi workspace create` with lockfile
3. Verify plugins fetched from lockfile
4. Repeat step 1-3
5. Verify identical results

**Expected Results:**
- Same plugins fetched
- Same versions used
- Results reproducible across runs

---

#### TC-LCK-003: Offline Mode

| Attribute | Value |
|-----------|-------|
| Test ID | TC-LCK-003 |
| Module | Lockfile |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Lockfile exists
- All plugins cached
- Network disconnected

**Test Steps:**
1. Disconnect network
2. Run `mochi workspace create`
3. Verify workspace creation succeeds

**Expected Results:**
- All plugins loaded from cache
- No network errors
- Workspace created successfully

---

#### TC-LCK-004: Lockfile Tampering Detection

| Attribute | Value |
|-----------|-------|
| Test ID | TC-LCK-004 |
| Module | Lockfile |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Valid lockfile exists

**Test Steps:**
1. Modify lockfile content (change SHA)
2. Run workspace creation
3. Verify error or warning about mismatch

**Expected Results:**
- Mismatch detected
- Clear error message
- Option to regenerate lockfile

---

### 4.7 Hook System

#### TC-HOOK-001: Hook Discovery

| Attribute | Value |
|-----------|-------|
| Test ID | TC-HOOK-001 |
| Module | Hook System |
| Priority | P1 - High |
| Type | Unit |

**Preconditions:**
- Hook scripts exist in `.mochi/hooks/`

**Test Steps:**
1. Create hooks (create.sh, up.sh, down.sh)
2. Run hook discovery
3. Verify all hooks found

**Expected Results:**
- All 3 hooks discovered
- Script paths correctly resolved

---

#### TC-HOOK-002: Hook Execution with Environment

| Attribute | Value |
|-----------|-------|
| Test ID | TC-HOOK-002 |
| Module | Hook System |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- Hook script with environment variable access
- Services configured

**Test Steps:**
1. Create `.mochi/hooks/up.sh`:
   ```bash
   #!/bin/bash
   echo "Web URL: $mochi_SERVICE_WEB_URL" >> /tmp/hook-test.log
   ```
2. Start workspace
3. Check `/tmp/hook-test.log`

**Expected Results:**
- Hook executed
- Environment variable accessible
- Log file contains expected URL

---

#### TC-HOOK-003: Hook Failure Handling

| Attribute | Value |
|-----------|-------|
| Test ID | TC-HOOK-003 |
| Module | Hook System |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Hook script that fails

**Test Steps:**
1. Create failing hook:
   ```bash
   #!/bin/bash
   exit 1
   ```
2. Run workspace up
3. Verify error message with recovery suggestions
4. Verify workspace state is consistent

**Expected Results:**
- Error clearly displayed
- Recovery suggestions provided
- No orphaned resources

---

#### TC-HOOK-004: Missing Hook Fallback

| Attribute | Value |
|-----------|-------|
| Test ID | TC-HOOK-004 |
| Module | Hook System |
| Priority | P2 - Medium |
| Type | Unit |

**Preconditions:**
- No hooks defined

**Test Steps:**
1. Run workspace up
2. Verify default behavior works

**Expected Results:**
- No errors about missing hooks
- Default service startup works

---

### 4.8 Docker Provider

#### TC-DCK-001: Container Creation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-DCK-001 |
| Module | Docker Provider |
| Priority | P0 - Critical |
| Type | Integration |

**Preconditions:**
- Docker daemon running
- Valid image available

**Test Steps:**
1. Run workspace up
2. Verify container created
3. Verify container labels
4. Verify container name format

**Expected Results:**
- Container exists
- Label `mochi.session.id` present
- Name format: `mochi-workspace-<name>`

---

#### TC-DCK-002: Docker-in-Docker Support

| Attribute | Value |
|-----------|-------|
| Test ID | TC-DCK-002 |
| Module | Docker Provider |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Dockerfile or docker-compose.yml exists
- DinD enabled in config

**Test Steps:**
1. Create workspace with DinD
2. Exec into container
3. Run `docker version`

**Expected Results:**
- Docker commands work inside container
- Docker version output present

---

#### TC-DCK-003: Port Mapping

| Attribute | Value |
|-----------|-------|
| Test ID | TC-DCK-003 |
| Module | Docker Provider |
| Priority | P0 - Critical |
| Type | Integration |

**Preconditions:**
- Services with ports configured

**Test Steps:**
1. Start workspace
2. Check `docker port <container>`
3. Verify ports accessible from host

**Expected Results:**
- Port bindings visible
- Services accessible on mapped ports
- No port conflicts

---

#### TC-DCK-004: Bind Mount for Worktree

| Attribute | Value |
|-----------|-------|
| Test ID | TC-DCK-004 |
| Module | Docker Provider |
| Priority | P0 - Critical |
| Type | Integration |

**Preconditions:**
- Workspace created
- Container running

**Test Steps:**
1. Exec into container
2. Check mounted directory
3. Verify worktree files accessible

**Expected Results:**
- Worktree mounted at expected path
- Files visible inside container
- Changes sync between host and container

---

### 4.9 Error Handling

#### TC-ERR-001: Invalid Configuration

| Attribute | Value |
|-----------|-------|
| Test ID | TC-ERR-001 |
| Module | Error Handling |
| Priority | P1 - High |
| Type | Unit |

**Preconditions:**
- None

**Test Steps:**
1. Create invalid config (missing required fields)
2. Run any mochi command
3. Verify clear error message

**Expected Results:**
- Error message identifies specific issue
- Suggestion for fix provided
- Exit code non-zero

---

#### TC-ERR-002: Docker Not Available

| Attribute | Value |
|-----------|-------|
| Test ID | TC-ERR-002 |
| Module | Error Handling |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Docker daemon not running

**Test Steps:**
1. Run workspace up
2. Verify error about Docker not available
3. Verify command exits cleanly

**Expected Results:**
- Clear error about Docker
- No panic
- Helpful message for user

---

#### TC-ERR-003: Git Operations Failure

| Attribute | Value |
|-----------|-------|
| Test ID | TC-ERR-003 |
| Module | Error Handling |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Not in git repository

**Test Steps:**
1. Run workspace create
2. Verify error about git repository required

**Expected Results:**
- Clear error message
- Suggestion to run `git init`

---

#### TC-ERR-004: Resource Cleanup on Error

| Attribute | Value |
|-----------|-------|
| Test ID | TC-ERR-004 |
| Module | Error Handling |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Partial workspace creation

**Test Steps:**
1. Start workspace create
2. Kill process mid-execution
3. Verify no orphaned resources
4. Verify retry works

**Expected Results:**
- No leaked containers
- No dangling worktrees
- Clean state for retry

---

### 4.10 Performance Tests

#### TC-PERF-001: Workspace Creation Time

| Attribute | Value |
|-----------|-------|
| Test ID | TC-PERF-001 |
| Module | Performance |
| Priority | P1 - High |
| Type | Benchmark |

**Preconditions:**
- Docker daemon running
- No existing resources

**Test Steps:**
1. Measure time for `workspace create` with no plugins
2. Measure time for `workspace create` with 5 plugins
3. Measure time for `workspace create` with 10 plugins

**Expected Results:**
- No plugins: < 10s
- 5 plugins: < 30s
- 10 plugins: < 60s

**Automation:**
```go
func BenchmarkWorkspaceCreate(b *testing.B) {
    for i := 0; i < b.N; i++ {
        te := NewTestEnvironment(nil)
        start := time.Now()
        te.CreateWorkspace(fmt.Sprintf("bench-%d", i))
        duration := time.Since(start)
        if duration > 30*time.Second {
            b.Errorf("Workspace creation took too long: %v", duration)
        }
        te.Cleanup()
    }
}
```

---

#### TC-PERF-002: Memory Usage

| Attribute | Value |
|-----------|-------|
| Test ID | TC-PERF-002 |
| Module | Performance |
| Priority | P1 - High |
| Type | Benchmark |

**Preconditions:**
- Docker daemon running
- Workspace created and running

**Test Steps:**
1. Monitor memory usage during workspace operations
2. Check for memory leaks over time

**Expected Results:**
- Base memory: < 100MB
- Peak memory: < 500MB
- No memory leaks

---

#### TC-PERF-003: Parallel Plugin Fetch

| Attribute | Value |
|-----------|-------|
| Test ID | TC-PERF-003 |
| Module | Performance |
| Priority | P2 - Medium |
| Type | Benchmark |

**Preconditions:**
- 5 remote plugins configured
- Network available

**Test Steps:**
1. Measure time for parallel fetch
2. Compare with sequential fetch estimate

**Expected Results:**
- Parallel fetch: < 10s
- Speedup: > 3x vs sequential

---

### 4.11 Security Tests

#### TC-SEC-001: Authentication Token Generation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-SEC-001 |
| Module | Security |
| Priority | P0 - Critical |
| Type | Unit |

**Preconditions:**
- None

**Test Steps:**
1. Generate auth token
2. Verify token length and complexity
3. Verify tokens are unique per session

**Expected Results:**
- Token meets complexity requirements
- No predictable patterns
- Tokens unique per session

---

#### TC-SEC-002: Container Isolation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-SEC-002 |
| Module | Security |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Workspace running

**Test Steps:**
1. Attempt to access host system from container
2. Verify container cannot break isolation
3. Verify resource limits enforced

**Expected Results:**
- No host access
- Resource limits applied
- Proper isolation maintained

---

#### TC-SEC-003: Template Injection Prevention

| Attribute | Value |
|-----------|-------|
| Test ID | TC-SEC-003 |
| Module | Security |
| Priority | P1 - High |
| Type | Unit |

**Preconditions:**
- Templates with user content

**Test Steps:**
1. Create template with malicious content
2. Process template
3. Verify no code execution

**Expected Results:**
- Content sanitized
- No code execution
- Safe output

## 5. Test Data Management

### 5.1 Test Fixtures

| Fixture | Type | Purpose |
|---------|------|---------|
| `empty-repo` | Git repository | Basic initialization tests |
| `multi-branch-repo` | Git repository | Branch isolation tests |
| `with-dockerfile` | Git repository | DinD tests |
| `minimal.yaml` | Config | Basic config tests |
| `full-stack.yaml` | Config | Service discovery tests |
| `with-plugins.yaml` | Config | Plugin system tests |

### 5.2 Test Data Patterns

```go
// Test data patterns for service configurations
var serviceConfigs = []struct {
    name    string
    config  string
}{
    {
        name: "single_service",
        config: `
services:
  web:
    command: "npm run dev"
`,
    },
    {
        name: "multi_service",
        config: `
services:
  web:
    command: "npm run dev"
  api:
    command: "cd server && npm run dev"
    depends_on: ["db"]
  db:
    command: "docker-compose up -d postgres"
`,
    },
}
```

---

## 6. Test Automation Framework

### 6.1 TestEnvironment Structure

```go
type TestEnvironment struct {
    t           *testing.T
    baseDir     string
    gitRepoDir  string
    mochiBin string
    containers  []string
    worktrees   []string
}

// Methods
func (te *TestEnvironment) InitGitRepo() error
func (te *TestEnvironment) Runmochi(args ...string) (string, error)
func (te *TestEnvironment) CreateWorkspace(name string) error
func (te *TestEnvironment) StartWorkspace(name string) error
func (te *TestEnvironment) StopWorkspace(name string) error
func (te *TestEnvironment) RemoveWorkspace(name string) error
func (te *TestEnvironment) Cleanup() error
func (te *TestEnvironment) WriteConfig(config string) error
func (te *TestEnvironment) WriteFile(path, content string) error
func (te *TestEnvironment) ReadFile(path string) string
func (te *TestEnvironment) DirExists(path string) bool
func (te *TestEnvironment) FileExists(path string) bool
```

### 6.2 Assertion Helpers

```go
func assertWorktreeExists(t *testing.T, name string)
func assertAgentConfigsGenerated(t *testing.T, name string)
func assertContainerRunning(t *testing.T, name string)
func assertServicesAccessible(t *testing.T, name string)
func assertEnvVarPresent(t *testing.T, container, pattern string)
```

---

## 7. CI/CD Integration

### 7.1 GitHub Actions Workflow

```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go test ./pkg/... -coverprofile=coverage.out
      - uses: codecov/codecov-action@v4
        with:
          file: coverage.out

  integration:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:dind
        options: --privileged
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go test -tags=integration ./... -v

  e2e:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:dind
        options: --privileged
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: make test-e2e
      - run: make benchmark
        continue-on-error: true

  perf:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:dind
        options: --privileged
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go test -bench=. -benchmem ./...
      - uses: benchmark-action/github-action-benchmark@v1
        with:
          name: Go Benchmark
          tool: 'go'
          output-file-path: benchmark.txt
```

### 7.2 Makefile Targets

```makefile
.PHONY: test-unit test-integration test-e2e test-all test-bench clean

test-unit:
	go test ./pkg/... -coverprofile=coverage.out

test-integration:
	go test -tags=integration ./...

test-e2e:
	go test -tags=e2e ./e2e/... -v

test-all: test-unit test-integration test-e2e

test-bench:
	go test -bench=. -benchmem ./...

clean:
	rm -rf coverage.out benchmark.txt bin/
```

---

## 8. Risk Analysis

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Docker daemon unavailable | Medium | High | Skip integration tests, mock provider |
| Network failures during plugin fetch | Medium | Medium | Implement retry with backoff, offline mode |
| Race conditions in parallel operations | Low | High | Use proper synchronization, test concurrency |
| Resource exhaustion in CI | Low | Medium | Limit parallel test execution, use CI caching |
| Flaky E2E tests | Medium | Medium | Improve test isolation, add timeouts |

---

## 9. Appendix

### 9.1 Test Case Index

| ID | Module | Priority | Type |
|----|--------|----------|------|
| TC-CLI-001 | CLI / Init | P0 | E2E |
| TC-CLI-002 | CLI / Validation | P1 | Unit |
| TC-CLI-003 | CLI / Workspace | P1 | Integration |
| TC-WS-001 | Workspace | P0 | E2E |
| TC-WS-002 | Workspace | P1 | E2E |
| TC-WS-003 | Workspace | P0 | Integration |
| TC-SD-001 | Service Discovery | P0 | E2E |
| TC-SD-002 | Service Discovery | P1 | Integration |
| TC-SD-003 | Service Discovery | P1 | Integration |
| TC-AGT-001 | Agent Configuration | P0 | E2E |
| TC-AGT-002 | Agent Configuration | P1 | Integration |
| TC-AGT-003 | Agent Configuration | P0 | Integration |
| TC-PLG-001 | Plugin System | P1 | Unit |
| TC-PLG-002 | Plugin System | P0 | Unit |
| TC-PLG-003 | Plugin System | P0 | Unit |
| TC-PLG-004 | Plugin System | P1 | Integration |
| TC-LCK-001 | Lockfile | P0 | Integration |
| TC-LCK-002 | Lockfile | P0 | Integration |
| TC-LCK-003 | Lockfile | P1 | Integration |
| TC-LCK-004 | Lockfile | P1 | Integration |
| TC-HOOK-001 | Hook System | P1 | Unit |
| TC-HOOK-002 | Hook System | P0 | E2E |
| TC-HOOK-003 | Hook System | P1 | Integration |
| TC-HOOK-004 | Hook System | P2 | Unit |
| TC-DCK-001 | Docker Provider | P0 | Integration |
| TC-DCK-002 | Docker Provider | P1 | Integration |
| TC-DCK-003 | Docker Provider | P0 | Integration |
| TC-DCK-004 | Docker Provider | P0 | Integration |
| TC-ERR-001 | Error Handling | P1 | Unit |
| TC-ERR-002 | Error Handling | P1 | Integration |
| TC-ERR-003 | Error Handling | P1 | Integration |
| TC-ERR-004 | Error Handling | P1 | Integration |
| TC-PERF-001 | Performance | P1 | Benchmark |
| TC-PERF-002 | Performance | P1 | Benchmark |
| TC-PERF-003 | Performance | P2 | Benchmark |
| TC-SEC-001 | Security | P0 | Unit |
| TC-SEC-002 | Security | P1 | Integration |
| TC-SEC-003 | Security | P1 | Unit |

### 9.2 Coverage Goals by Module

| Module | Unit Coverage | Integration Coverage | E2E Coverage |
|--------|---------------|---------------------|--------------|
| CLI Commands | 90% | 100% | 100% |
| Workspace | 85% | 100% | 100% |
| Service Discovery | 90% | 100% | 100% |
| Agent Configuration | 85% | 100% | 100% |
| Plugin System | 95% | 100% | 90% |
| Lockfile | 90% | 100% | 90% |
| Hook System | 85% | 100% | 90% |
| Docker Provider | 80% | 100% | 90% |
| Error Handling | 90% | 90% | 80% |
| Performance | N/A | N/A | 100% |
| Security | 90% | 90% | 80% |

### 9.3 Test Execution Schedule

| Phase | Tests | Duration | Frequency |
|-------|-------|----------|-----------|
| Unit | All TC-* unit tests | 2-5 min | Every commit |
| Integration | All integration tests | 10-15 min | Every PR |
| E2E | All E2E tests | 20-30 min | Every PR |
| Performance | Benchmark tests | 5-10 min | Daily |
| Security | Security tests | 10-15 min | Weekly |

---

## 10. M3 Beta Test Cases (Multi-Machine Orchestration)

### 10.1 QEMU Provider Tests

#### TC-QEM-001: QEMU VM Creation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-QEM-001 |
| Module | QEMU Provider |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- QEMU installed
- Valid OS image available

**Test Steps:**
1. `mochi workspace create-multi test --machines 1 --provider qemu`
2. Verify VM started
3. Verify SSH access available
4. Verify network configured

**Expected Results:**
- VM running within 60s
- SSH connection successful
- Unique IP assigned

---

#### TC-QEM-002: Multi-Machine QEMU Session

| Attribute | Value |
|-----------|-------|
| Test ID | TC-QEM-002 |
| Module | QEMU Provider |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- QEMU installed
- Sufficient system resources (8GB+ RAM)

**Test Steps:**
1. `mochi workspace create-multi multi-test --machines 3 --provider qemu`
2. Verify 3 VMs started
3. Verify unique IPs assigned
4. Verify SSH to each VM

**Expected Results:**
- All 3 VMs running
- No IP conflicts
- Each VM accessible via SSH

---

#### TC-QEM-003: OS Image Management

| Attribute | Value |
|-----------|-------|
| Test ID | TC-QEM-003 |
| Module | QEMU Provider |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- QEMU installed
- No cached images

**Test Steps:**
1. `mochi machine image list`
2. `mochi machine image pull ubuntu-22.04`
3. Verify download completes
4. Verify image cached

**Expected Results:**
- Image downloaded successfully
- Image cached for reuse
- Integrity verification passes

---

#### TC-QEM-004: QEMU Resource Limits

| Attribute | Value |
|-----------|-------|
| Test ID | TC-QEM-004 |
| Module | QEMU Provider |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- QEMU installed
- Resource limits configured

**Test Steps:**
1. Create VM with resource limits:
   ```bash
   mochi workspace create test --provider qemu --memory 2g --cpus 2
   ```
2. Verify VM respects limits
3. Verify host system stability

**Expected Results:**
- VM uses <= configured memory
- VM uses <= configured CPUs
- No resource exhaustion on host

---

### 10.2 Multi-Machine Session Management

#### TC-MUL-001: Session Creation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-MUL-001 |
| Module | Multi-Machine |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- Provider available (QEMU or Docker)

**Test Steps:**
1. Define multi-machine config:
   ```yaml
   machines:
     - name: web
       role: frontend
     - name: api
       role: backend
     - name: db
       role: database
   ```
2. `mochi workspace create-multi app`
3. Verify all machines created
4. Verify session state persisted

**Expected Results:**
- All 3 machines created
- Session state saved
- Machine roles assigned correctly

---

#### TC-MUL-002: Machine Roles and Dependencies

| Attribute | Value |
|-----------|-------|
| Test ID | TC-MUL-002 |
| Module | Multi-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multi-machine config with dependencies

**Test Steps:**
1. Create config with dependencies:
   ```yaml
   machines:
     - name: db
       role: database
     - name: api
       role: backend
       depends_on: [db]
     - name: web
       role: frontend
       depends_on: [api]
   ```
2. Start session
3. Verify startup order
4. Verify all machines reachable

**Expected Results:**
- Machines start in correct order (db -> api -> web)
- No race conditions
- All machines fully operational

---

#### TC-MUL-003: Session Scaling

| Attribute | Value |
|-----------|-------|
| Test ID | TC-MUL-003 |
| Module | Multi-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multi-machine session running

**Test Steps:**
1. `mochi workspace scale app --add 1 --role worker`
2. Verify new machine added
3. Verify session still functional
4. `mochi workspace scale app --remove worker-1`
5. Verify machine removed

**Expected Results:**
- Machine added successfully
- Service continues running
- Machine removed cleanly

---

#### TC-MUL-004: Health Monitoring

| Attribute | Value |
|-----------|-------|
| Test ID | TC-MUL-004 |
| Module | Multi-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multi-machine session running

**Test Steps:**
1. `mochi workspace status app`
2. Verify all machines showing healthy
3. Simulate machine failure
4. Verify auto-recovery triggered
5. Verify health status updated

**Expected Results:**
- All machines report healthy
- Failed machine detected
- Auto-recovery initiated
- Status updated correctly

---

### 10.3 Cross-Machine Coordination

#### TC-COR-001: Service Registry

| Attribute | Value |
|-----------|-------|
| Test ID | TC-COR-001 |
| Module | Cross-Machine |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- Multi-machine session with web and db machines

**Test Steps:**
1. Create multi-machine session
2. Deploy service to db machine
3. Verify service registered
4. Deploy service to web machine
5. Verify web can discover db service

**Expected Results:**
- Service registered automatically
- DNS resolution works across machines
- Service accessible from other machines

---

#### TC-COR-002: Cross-Machine Networking

| Attribute | Value |
|-----------|-------|
| Test ID | TC-COR-002 |
| Module | Cross-Machine |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- Multi-machine session with 2+ machines

**Test Steps:**
1. Create web and api machines
2. Configure network bridge
3. From web machine: curl http://api.internal:8080
4. Verify response received

**Expected Results:**
- Network bridge established
- Internal DNS resolution works
- Cross-machine traffic flows
- Latency < 100ms

---

#### TC-COR-003: Load Balancing

| Attribute | Value |
|-----------|-------|
| Test ID | TC-COR-003 |
| Module | Cross-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multi-machine session with multiple workers

**Test Steps:**
1. Create session with 3 workers
2. Configure load balancer
3. Send 100 requests to load balancer
4. Verify requests distributed across workers

**Expected Results:**
- Requests distributed evenly
- No request failures
- Load balancer configuration valid

---

#### TC-COR-004: Encrypted Communication

| Attribute | Value |
|-----------|-------|
| Test ID | TC-COR-004 |
| Module | Cross-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multi-machine session
- TLS certificates configured

**Test Steps:**
1. Configure TLS for cross-machine traffic
2. Send sensitive data between machines
3. Verify encryption in transit
4. Attempt man-in-the-middle attack
5. Verify attack blocked

**Expected Results:**
- Traffic encrypted
- Certificates validated
- Unauthorized access blocked

---

### 10.4 Multi-Machine CLI Commands

#### TC-CLI-MUL-001: create-multi Command

| Attribute | Value |
|-----------|-------|
| Test ID | TC-CLI-MUL-001 |
| Module | CLI / Multi-Machine |
| Priority | P0 - Critical |
| Type | E2E |

**Preconditions:**
- Provider available

**Test Steps:**
1. `mochi workspace create-multi --help`
2. Verify help output correct
3. `mochi workspace create-multi test --machines 2 --provider docker`
4. Verify session created

**Expected Results:**
- Help text complete and accurate
- Command creates session
- Machines properly configured

---

#### TC-CLI-MUL-002: connect Command

| Attribute | Value |
|-----------|-------|
| Test ID | TC-CLI-MUL-002 |
| Module | CLI / Multi-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multi-machine session running

**Test Steps:**
1. `mochi workspace connect test machine-1`
2. Execute command inside machine
3. Verify command runs correctly
4. Exit and return to host

**Expected Results:**
- SSH connection established
- Commands execute in remote machine
- Clean exit back to host

---

#### TC-CLI-MUL-003: scale Command

| Attribute | Value |
|-----------|-------|
| Test ID | TC-CLI-MUL-003 |
| Module | CLI / Multi-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multi-machine session running

**Test Steps:**
1. `mochi workspace scale --help`
2. Verify help output correct
3. `mochi workspace scale test --add 2`
4. Verify 2 new machines added
5. `mochi workspace scale test --remove machine-3`
6. Verify machine removed

**Expected Results:**
- Scale command works correctly
- New machines functional
- Removed machines cleaned up

---

#### TC-CLI-MUL-004: status Command

| Attribute | Value |
|-----------|-------|
| Test ID | TC-CLI-MUL-004 |
| Module | CLI / Multi-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multi-machine session running

**Test Steps:**
1. `mochi workspace status test`
2. Verify output shows all machines
3. Verify status indicators correct
4. Verify resource usage displayed

**Expected Results:**
- All machines listed
- Status indicators accurate
- Resource metrics correct

---

### 10.5 Enhanced Lockfile for Multi-Machine

#### TC-LCK-MUL-001: Multi-Machine Lockfile Generation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-LCK-MUL-001 |
| Module | Lockfile / Multi-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multi-machine config with 3 machines

**Test Steps:**
1. `mochi plugin update`
2. Verify `mochi.lock` created
3. Verify lockfile contains all machine configs
4. Verify OS images and versions pinned

**Expected Results:**
- Lockfile includes all machines
- Image versions locked
- Network config included

---

#### TC-LCK-MUL-002: Deterministic Multi-Machine Recreation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-LCK-MUL-002 |
| Module | Lockfile / Multi-Machine |
| Priority | P0 - Critical |
| Type | Integration |

**Preconditions:**
- Multi-machine lockfile exists

**Test Steps:**
1. Destroy all machines
2. Recreate from lockfile
3. Verify identical configuration
4. Repeat 3 times
5. Verify reproducibility

**Expected Results:**
- Identical machines created each time
- No configuration drift
- Reproducible across environments

---

#### TC-LCK-MUL-003: Offline Multi-Machine Mode

| Attribute | Value |
|-----------|-------|
| Test ID | TC-LCK-MUL-003 |
| Module | Lockfile / Multi-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- All images cached
- Lockfile exists

**Test Steps:**
1. Disconnect network
2. Run `mochi workspace create-multi test`
3. Verify machines created from cache

**Expected Results:**
- Machines created from cached images
- No network requests
- Successful creation

---

### 10.6 M3 Beta Performance Tests

#### TC-PERF-MUL-001: Multi-Machine Startup Time

| Attribute | Value |
|-----------|-------|
| Test ID | TC-PERF-MUL-001 |
| Module | Performance / Multi-Machine |
| Priority | P1 - High |
| Type | Benchmark |

**Preconditions:**
- Provider available
- Sufficient resources

**Test Steps:**
1. Measure time for 1 machine: should be < 60s
2. Measure time for 3 machines (parallel): should be < 90s
3. Measure time for 5 machines (parallel): should be < 120s

**Expected Results:**
- Single machine: < 60s
- 3 machines: < 90s
- 5 machines: < 120s

---

#### TC-PERF-MUL-002: Cross-Machine Latency

| Attribute | Value |
|-----------|-------|
| Test ID | TC-PERF-MUL-002 |
| Module | Performance / Multi-Machine |
| Priority | P1 - High |
| Type | Benchmark |

**Preconditions:**
- Multi-machine session running

**Test Steps:**
1. Send 1000 pings between machines
2. Measure latency percentiles (p50, p95, p99)

**Expected Results:**
- p50 latency: < 10ms
- p95 latency: < 50ms
- p99 latency: < 100ms

---

#### TC-PERF-MUL-003: QEMU Memory Overhead

| Attribute | Value |
|-----------|-------|
| Test ID | TC-PERF-MUL-003 |
| Module | Performance / Multi-Machine |
| Priority | P2 - Medium |
| Type | Benchmark |

**Preconditions:**
- QEMU installed
- Multiple VMs running

**Test Steps:**
1. Measure memory with 1 VM (2GB configured)
2. Measure memory with 3 VMs (2GB each)
3. Compare to Docker equivalent

**Expected Results:**
- 1 VM: ~2.5GB total
- 3 VMs: ~8GB total
- Acceptable overhead vs Docker

---

### 10.7 M3 Beta Security Tests

#### TC-SEC-MUL-001: VM Isolation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-SEC-MUL-001 |
| Module | Security / Multi-Machine |
| Priority | P0 - Critical |
| Type | Integration |

**Preconditions:**
- QEMU session running

**Test Steps:**
1. Attempt to escape VM to host
2. Attempt to access other VMs' resources
3. Verify isolation boundaries

**Expected Results:**
- No host escape possible
- No cross-VM access
- Proper network isolation

---

#### TC-SEC-MUL-002: Network Segmentation

| Attribute | Value |
|-----------|-------|
| Test ID | TC-SEC-MUL-002 |
| Module | Security / Multi-Machine |
| Priority | P1 - High |
| Type | Integration |

**Preconditions:**
- Multi-machine session with network isolation

**Test Steps:**
1. Configure network policies
2. Attempt unauthorized access between machines
3. Verify blocked traffic

**Expected Results:**
- Network policies enforced
- Unauthorized access blocked
- Audit logs capture attempts

---

### 10.8 M3 Beta Test Case Index

| ID | Module | Priority | Type |
|----|--------|----------|------|
| TC-QEM-001 | QEMU Provider | P0 | E2E |
| TC-QEM-002 | QEMU Provider | P0 | E2E |
| TC-QEM-003 | QEMU Provider | P1 | Integration |
| TC-QEM-004 | QEMU Provider | P1 | Integration |
| TC-MUL-001 | Multi-Machine | P0 | E2E |
| TC-MUL-002 | Multi-Machine | P1 | Integration |
| TC-MUL-003 | Multi-Machine | P1 | Integration |
| TC-MUL-004 | Multi-Machine | P1 | Integration |
| TC-COR-001 | Cross-Machine | P0 | E2E |
| TC-COR-002 | Cross-Machine | P0 | E2E |
| TC-COR-003 | Cross-Machine | P1 | Integration |
| TC-COR-004 | Cross-Machine | P1 | Integration |
| TC-CLI-MUL-001 | CLI / Multi-Machine | P0 | E2E |
| TC-CLI-MUL-002 | CLI / Multi-Machine | P1 | Integration |
| TC-CLI-MUL-003 | CLI / Multi-Machine | P1 | Integration |
| TC-CLI-MUL-004 | CLI / Multi-Machine | P1 | Integration |
| TC-LCK-MUL-001 | Lockfile / Multi-Machine | P1 | Integration |
| TC-LCK-MUL-002 | Lockfile / Multi-Machine | P0 | Integration |
| TC-LCK-MUL-003 | Lockfile / Multi-Machine | P1 | Integration |
| TC-PERF-MUL-001 | Performance / Multi-Machine | P1 | Benchmark |
| TC-PERF-MUL-002 | Performance / Multi-Machine | P1 | Benchmark |
| TC-PERF-MUL-003 | Performance / Multi-Machine | P2 | Benchmark |
| TC-SEC-MUL-001 | Security / Multi-Machine | P0 | Integration |
| TC-SEC-MUL-002 | Security / Multi-Machine | P1 | Integration |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-12 | QA Team | Initial version |
| 1.1.0 | 2026-01-12 | QA Team | Added M3 Beta test cases |

---

**End of Document**
