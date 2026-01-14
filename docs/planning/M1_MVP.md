# Milestone: M1_Workspace (Core CLI & Isolation)

**Objective**: Complete workspace-centric CLI with isolation, lifecycle hooks, and simplified agent configuration. Deliver a clean, intuitive developer experience with comprehensive test coverage.

## 🎯 Success Criteria
- [x] `vendetta workspace create <name>` creates isolated workspace with agent configs
- [x] `vendetta workspace up <name>` starts services with port forwarding and optional hooks
- [x] Service discovery injects `vendetta_SERVICE_*_URL` environment variables correctly
- [x] Agent configurations generated for supported AI agents
- [x] Optional lifecycle hooks execute from `.vendetta/hooks/` directory
- [ ] Full E2E test suite with 90%+ coverage and CI integration
- [ ] JSON schema validation for config.yaml
- [ ] Makefile for development and CI workflows

## 🛠 Implementation Tasks

| ID | Title | Priority | Status | Test Plan |
| :--- | :--- | :--- | :--- | :--- |
| **[CLI-03](./tasks/CLI-03.md)** | Workspace Command Group Implementation | 🔥 High | [In Progress] | [TP-CLI-03](#test-plan-cli-03) |
| **[COR-03](./tasks/COR-03.md)** | Service Discovery Environment Variables Fix | 🔥 High | [Completed] | [TP-COR-03](#test-plan-cor-03) |
| **[COR-04](./tasks/COR-04.md)** | Convention-Based Hook System | ⚡ Med | [In Progress] | [TP-COR-04](#test-plan-cor-04) |
| **[AGT-02](./tasks/AGT-02.md)** | File-Based Agent Config Overrides | ⚡ Med | [In Progress] | [TP-AGT-02](#test-plan-agt-02) |
| **[CFG-01](./tasks/CFG-01.md)** | Config Pull/Sync Commands | 🟢 Low | [Pending] | [TP-CFG-01](#test-plan-cfg-01) |
| **[VFY-02](./tasks/VFY-02.md)** | Comprehensive E2E Test Suite | 🔥 High | [Pending] | [TP-VFY-02](#test-plan-vfy-02) |

## 📋 Detailed Test Plans

### **TP-CLI-03: Workspace Command Group Testing**
**Objective**: Validate workspace lifecycle commands work correctly with proper isolation

**Unit Tests:**
- ✅ `workspace create` creates Git worktree and agent configs
- ✅ `workspace up` starts container and executes hooks
- ✅ `workspace stop/down` properly stops container and services
- ✅ Context awareness detects workspace from current directory
- ✅ Error handling for non-existent workspaces

**Integration Tests:**
- ✅ Full workspace lifecycle: create → up → stop → down → rm
- ✅ Port forwarding and service discovery
- ✅ Agent config generation in worktree
- ✅ Hook execution order and error handling

**E2E Scenarios:**
```bash
# Test 1: Basic workspace lifecycle
vendetta workspace create feature-x
cd .vendetta/worktrees/feature-x
vendetta workspace up    # Should auto-detect workspace
# Verify: Container running, services accessible, agent configs present

# Test 2: Hook execution
echo "#!/bin/bash\necho 'Hook executed'" > .vendetta/hooks/up.sh
chmod +x .vendetta/hooks/up.sh
vendetta workspace up feature-x
# Verify: Hook output in logs

# Test 3: Service discovery
vendetta workspace up feature-x
# Verify: vendetta_SERVICE_WEB_URL=localhost:3000 in container env
```

---

### **TP-COR-03: Service Discovery Testing**
**Objective**: Ensure environment variables are properly injected into running containers before services start

**Unit Tests:**
- ✅ Port auto-detection from service commands (docker-compose, npm, etc.)
- ✅ Environment variable generation follows `vendetta_SERVICE_{NAME}_URL` pattern
- ✅ Protocol guessing from service nature (postgres → postgresql://, web → http://)
- ✅ Multiple services generate multiple environment variables

**Integration Tests:**
- ✅ Container receives environment variables during creation (before services start)
- ✅ Variables available in hook scripts during execution
- ✅ Variables accessible in running container shell

**E2E Scenarios:**
```bash
# Test service discovery with command-based services
vendetta workspace create discovery-test
cat > .vendetta/config.yaml << EOF
services:
  web:
    command: "cd client && npm run dev"
  api:
    command: "cd server && npm run dev"
  db:
    command: "docker-compose up -d postgres"
EOF
vendetta workspace up discovery-test

# Verify environment variables available before services start
vendetta workspace shell discovery-test
env | grep vendetta_SERVICE
# Expected: vendetta_SERVICE_WEB_URL=http://localhost:3000
#          vendetta_SERVICE_API_URL=http://localhost:8080
#          vendetta_SERVICE_DB_URL=postgresql://localhost:5432
```

---

### **TP-COR-04: Hook System Testing**
**Objective**: Validate convention-based hooks as main operations with environment variable access

**Unit Tests:**
- ✅ Hook discovery finds scripts in `.vendetta/hooks/`
- ✅ Missing hooks use default behavior (no errors)
- ✅ Scripts are made executable before execution (with user prompt/warning)
- ✅ Hook execution receives service discovery variables and workspace context

**Integration Tests:**
- ✅ Hooks execute as main operations (up.sh replaces docker-compose up)
- ✅ Environment variables available during hook execution
- ✅ Hook failures logged with recovery suggestions
- ✅ Hook modifications committed to git (user warned)

**E2E Scenarios:**
```bash
# Test hooks as main operations
vendetta workspace create hooks-demo
mkdir -p .vendetta/hooks

# Create up.sh to replace default behavior
cat > .vendetta/hooks/up.sh << 'EOF'
#!/bin/bash
echo "Starting custom services..."
docker-compose up -d
npm run dev &
EOF
chmod +x .vendetta/hooks/up.sh

# Execute workspace up - should run up.sh instead of default
vendetta workspace up hooks-demo

# Verify services started via hook
# Check docker-compose containers running
# Check npm dev server on expected port
```

---

### **TP-AGT-02: Agent Config Generation Testing**
**Objective**: Verify AI agent configuration generation for supported agents

**Unit Tests:**
- ✅ Base templates loaded from `.vendetta/templates/`
- ✅ Agent-specific configurations generated correctly
- ✅ Override files replace base templates per agent specs
- ✅ Empty files suppress template generation

**Integration Tests:**
- ✅ Config generation creates correct files per agent formats
- ✅ Template merging works with project overrides
- ✅ Agent directories processed correctly

**E2E Scenarios:**
```bash
# Test config generation for agents
vendetta workspace create config-test

# Create agent override
mkdir -p .vendetta/agents/cursor
cat > .vendetta/agents/cursor/.cursorrules << EOF
# Custom Cursor Rules
- Use TypeScript for all new files
- Prefer functional components over class components
EOF

# Create suppression for specific rules
touch .vendetta/agents/opencode/rules/legacy.md  # Empty file suppresses

vendetta workspace create config-test
# Verify: Agent configs generated in worktree
# Verify: Overrides applied correctly
# Verify: Suppressions work
```

---

### **TP-CFG-01: Config Management Testing**
**Objective**: Validate remote config pulling and syncing with proper merge precedence

**Unit Tests:**
- ✅ `config pull` clones remote repositories with branch specification
- ✅ Remote refs stored in state for merge tracking
- ✅ Fast-forward merges prefer remote when possible
- ✅ Conflict resolution uses chezmoi-style reconciliation

**Integration Tests:**
- ✅ Template precedence: remote (fast-forward) > local > base
- ✅ Sync creates filtered branch with only `.vendetta`
- ✅ Merge conflicts trigger interactive reconciliation

**E2E Scenarios:**
```bash
# Test config pulling with branch
vendetta config pull https://github.com/example/templates.git --branch develop

# Test template merging with stored refs
# Remote templates override local when fast-forward possible
# Conflicts show diff and allow selection

# Test config syncing
cat > .vendetta/config.yaml << EOF
sync_targets:
  - name: team-configs
    url: https://github.com/company/dev-templates.git
EOF
vendetta config sync team-configs
# Verify: .vendetta directory merged and pushed
```

---

### **TP-VFY-02: Comprehensive E2E Test Suite**
**Objective**: 90%+ test coverage with CI integration

**Test Categories:**
- ✅ **Unit Tests**: All public functions, error paths, edge cases
- ✅ **Integration Tests**: Component interactions, Docker operations
- ✅ **E2E Tests**: Full user workflows, real containers
- ✅ **Performance Tests**: Startup time < 30s, memory usage < 500MB
- ✅ **Compatibility Tests**: Multiple OS, Docker versions

**CI Pipeline:**
```yaml
# .github/workflows/test.yml
name: Test Suite
on: [push, pull_request]
jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: make test-unit

  integration:
    runs-on: ubuntu-latest
    services:
      docker: docker:dind
    steps:
      - uses: actions/checkout@v3
      - run: make test-integration

  e2e:
    runs-on: ubuntu-latest
    services:
      docker: docker:dind
    steps:
      - uses: actions/checkout@v3
      - run: make test-e2e
```

**Makefile Requirements:**
```makefile
# Makefile - Development and CI entrypoint
.PHONY: test-unit test-integration test-e2e test-all
.PHONY: build install clean lint fmt

test-unit:
	go test ./... -coverprofile=coverage.out

test-integration:
	go test -tags=integration ./...

test-e2e:
	go test -tags=e2e ./...

test-all: test-unit test-integration test-e2e

build:
	go build -o bin/vendetta ./cmd/vendetta

install: build
	cp bin/vendetta ~/.local/bin/

lint:
	golangci-lint run

fmt:
	go fmt ./...
```

**Coverage Goals:**
- **Unit Tests**: 85%+ coverage
- **Integration Tests**: All major workflows
- **E2E Tests**: Critical user journeys
- **Performance**: No obvious bottlenecks (memory leaks, N+1 queries)
- **Configuration**: JSON schema validation for IDE support

## 📝 Implementation Notes

**Development Workflow:**
1. Implement workspace command group (create/up/down/rm/shell)
2. Add JSON schema validation for config.yaml
3. Create Makefile for development and CI workflows
4. Implement port auto-detection from service commands
5. Add comprehensive unit tests for each component
6. Build integration tests for component interactions
7. Create E2E test scenarios for user workflows
8. Implement branch conflict handling with stash/pop logic
9. Add error recovery with debug command suggestions

**Branch Conflict Handling:**
- **Detection**: Check if branch exists and has uncommitted changes
- **Resolution**: `git stash`, checkout/create branch, `git stash pop`
- **User Confirmation**: Prompt before stashing uncommitted work
- **Validation**: Like git branch naming (alphanumeric, hyphens, underscores)

**Error Recovery & Debugging:**
- **Hook Failures**: Log errors with suggested fixes (like Heroku/Fly.io)
- **Container Issues**: Provide docker inspect/exec commands for debugging
- **Service Discovery**: Show port mapping and environment variable details
- **Git Conflicts**: Clear instructions for manual resolution

**Hook Environment Variables:**
- **Service Discovery**: `vendetta_SERVICE_{NAME}_URL` (protocol guessed from service type)
- **Workspace Context**: `WORKSPACE_NAME`, `WORKTREE_PATH`
- **Container Info**: `CONTAINER_ID` (when available)
- **Host Info**: `HOST_USER`, `HOST_CWD`

**Service Discovery URL Format:**
- **Protocol Guessing**: postgres images → `postgresql://`, web services → `http://`
- **Display Only**: Protocol hints shown in logs but URLs provide hostname:port
- **Future Remote**: Support for remote workspaces (k8s, LXC) with dynamic host detection

**Workspace Context Detection:**
- **Scope**: Only within `.vendetta/worktrees/<name>/` directory
- **Auto-detection**: Parse directory path to extract workspace name
- **Validation**: Verify worktree exists and is valid

**Configuration Validation:**
- **JSON Schema**: Create `schema/config.schema.json` for config.yaml validation
- **IDE Support**: Schema enables autocomplete and validation in VSCode, Cursor, etc.
- **Validation**: Runtime validation with helpful error messages
- **Versioning**: Schema versioning for backward compatibility

**Quality Gates:**
- ✅ All unit tests pass with 85%+ coverage
- ✅ Integration tests validate component interactions
- ✅ E2E tests cover critical user journeys
- ✅ JSON schema validation enables IDE intellisense
- ✅ Makefile serves as platform-agnostic CI entrypoint

**Risk Mitigation:**
- **Docker Compatibility**: Test with multiple Docker versions
- **Git Operations**: Handle various repository states gracefully
- **Network Issues**: Robust error handling for remote operations
- **Resource Cleanup**: Ensure workspaces are properly cleaned up
