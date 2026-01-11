# Vendatta E2E Test Plan

## Overview
This document outlines the comprehensive test plan for Vendatta's end-to-end (e2e) functionality, covering all possible flows and features.

## Current Test Status

### Implemented Flows ✅
- **Project Initialization**: `vendatta init` creates .vendatta directory and default config
- **Workspace Management**:
  - `vendatta workspace create <name>`: Creates git worktree and generates configs
  - `vendatta workspace up [name]`: Starts isolated environment with Docker/LXC
  - `vendatta workspace down [name]`: Stops running workspace
  - `vendatta workspace list`: Lists all workspaces
  - `vendatta workspace rm <name>`: Removes workspace completely
  - `vendatta workspace shell [name]`: Opens shell in workspace

### Partially Implemented Flows ⚠️
- **Service Management**: Services defined in config but startup automation incomplete
- **Agent Configuration**: Templates generated but MCP server not started

### Missing Flows ❌
- **Config Management**:
  - `vendatta config pull <url>`: Pull remote templates
  - `vendatta config list`: List pulled remotes
  - `vendatta config sync [target]`: Sync configs to remote
  - `vendatta config sync-all`: Sync to all configured targets
  - `vendatta config generate-schema`: Generate JSON schema
  - `vendatta config validate`: Validate config against schema
- **Plugin System**:
  - `vendatta plugin list`: List available plugins
  - `vendatta plugin check`: Validate plugin dependencies
- **Agent Management**:
  - `vendatta agent <session-id>`: Start MCP server for session

## Test Scenarios

### 1. Basic Workspace Lifecycle
**Objective**: Verify complete workspace creation, startup, and teardown
**Steps**:
1. Initialize project: `init`
2. Create workspace: `workspace create test-ws`
3. Start workspace: `workspace up test-ws`
4. Verify worktree exists and configs generated
5. Stop workspace: `workspace down test-ws`
6. Remove workspace: `workspace rm test-ws`
**Expected**: All commands succeed, resources cleaned up

### 2. Multi-Workspace Management
**Objective**: Test multiple concurrent workspaces
**Steps**:
1. Create multiple workspaces: `workspace create ws1`, `workspace create ws2`
2. Start both: `workspace up ws1`, `workspace up ws2`
3. List workspaces: `workspace list`
4. Verify isolation (different worktrees, ports)
5. Stop and remove all
**Expected**: No conflicts, proper isolation

### 3. Provider Testing
**Objective**: Test different execution providers
**Configurations**:
- Docker provider (default)
- LXC provider (when implemented)
**Steps**: Same as basic lifecycle with different provider configs
**Expected**: Environment starts with correct provider

### 4. Service Integration
**Objective**: Verify automatic service startup and discovery
**Configuration**: Define services in config.yaml
**Steps**:
1. Create workspace with services
2. Start workspace
3. Verify services running on mapped ports
4. Check environment variables (OURSKY_SERVICE_*)
**Expected**: Services accessible, ports mapped correctly

### 5. Agent Configuration
**Objective**: Test AI agent integration setup
**Configurations**: Different agents (Cursor, OpenCode, Claude, etc.)
**Steps**:
1. Configure agents in config.yaml
2. Create workspace
3. Verify agent-specific files generated (.cursor/, .opencode/, etc.)
4. Check MCP configurations if applicable
**Expected**: Agent configs match templates

### 6. Plugin System
**Objective**: Test plugin discovery and integration
**Steps**:
1. Create plugin manifests in .vendatta/plugins/
2. Run `plugin list`
3. Run `plugin check`
4. Create workspace with plugins
5. Verify plugin rules/commands merged
**Expected**: Plugins discovered, rules applied

### 7. Config Remote Management
**Objective**: Test remote template pulling and syncing
**Steps**:
1. Pull remote config: `config pull <git-url>`
2. List remotes: `config list`
3. Create workspace (should use remote templates)
4. Sync changes: `config sync <target>`
**Expected**: Remote templates applied, sync successful

### 8. Error Handling
**Objective**: Test graceful failure scenarios
**Cases**:
- Invalid workspace names
- Missing dependencies (git, docker)
- Port conflicts
- Corrupted configs
- Non-existent workspaces
**Expected**: Clear error messages, no crashes

### 9. Performance Benchmarks
**Objective**: Ensure acceptable startup times
**Metrics**:
- Workspace creation: < 30s
- Workspace startup: < 60s
- Resource usage within limits
**Steps**: Time operations, monitor resources

### 10. Git Integration
**Objective**: Test git worktree management
**Scenarios**:
- Branch creation/deletion
- Worktree isolation
- Conflict resolution
- Detached HEAD handling
**Expected**: Git operations work seamlessly

## Test Environment Requirements

### System Dependencies
- Go 1.24+
- Git
- Docker
- LXC (optional, for LXC provider)

### Test Data
- Sample project with various configurations
- Remote git repositories for config testing
- Plugin manifests
- Different agent configurations

## Implementation Notes

### Current Issues
1. **MCP Server**: Implemented but not started in workspace lifecycle
2. **Service Startup**: Services defined but startup hook incomplete
3. **Plugin Commands**: CLI commands missing despite test expectations
4. **Config Remotes**: Git-based config resolution not implemented
5. **LXC Provider**: Marked as under development

### Recommendations
1. **Remove MCP**: If not actively used, remove to reduce complexity
2. **Complete Service Integration**: Fix up.sh generation and service startup
3. **Implement Missing Commands**: Add config and plugin CLI commands
4. **Add Config Remote Support**: Implement git-based template pulling
5. **Improve Error Handling**: Better validation and user feedback

### Test Automation
- All tests should be runnable with `go test ./e2e`
- Tests should clean up resources (containers, worktrees)
- Parallel execution where possible
- Clear failure diagnostics