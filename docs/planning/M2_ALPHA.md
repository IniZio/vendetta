# Milestone: M2_Alpha (Namespaced Plugins & Determinism)

**Objective**: Transition vendetta into a modular, plugin-based platform with hierarchical discovery and `uv`-style version locking. This phase focuses on developer experience, parallel performance, and environment reproducibility.

## 🎯 Success Criteria
- [x] `pkg/plugins` registry implements hierarchical discovery and DAG resolution.
- [x] `vendetta.lock` ensures deterministic environment recreation across machines.
- [x] Remote plugins are fetched in parallel with 0 resolution overhead during workspace creation.
- [x] Agents receive namespaced capabilities (Rules, Skills, Commands).
- [ ] E2E test suite with 90%+ coverage and performance benchmarks (<30s startup, <500MB memory).
- [x] JSON schema validation for config.yaml with IDE support.
- [ ] Standardized Makefile with CI integration.
- [ ] Config extraction to plugins command for simplified sharing.
- [ ] LXC provider implementation as lightweight alternative to Docker.
- [ ] Standard config moved to `git@github.com:IniZio/vendetta-config.git`.

## 🛠 Implementation Tasks

| ID | Title | Priority | Status | Test Plan |
| :--- | :--- | :--- | :--- | :--- |
| **PLG-01** | Plugin Registry & DAG Resolver | 🔥 High | [✅ Completed] | [TP-PLG-01](#test-plan-plg-01) |
| **LCK-01** | Lockfile Manager (uv-style) | 🔥 High | [✅ Completed] | [TP-LCK-01](#test-plan-lck-01) |
| **CLI-04** | `vendetta plugin` Command Group | ⚡ Med | [✅ Completed] | [TP-CLI-04](#test-plan-cli-04) |
| **AGT-03** | Namespaced Agent Generation | ⚡ Med | [✅ Completed] | [TP-AGT-03](#test-plan-agt-03) |
| **VFY-02** | E2E Test Suite & Coverage | 🔥 High | [🚧 Pending] | [TP-VFY-02](#test-plan-vfy-02) |
| **SCH-01** | JSON Schema Validation | ⚡ Med | [✅ Completed] | [TP-SCH-01](#test-plan-sch-01) |
| **BLD-01** | Makefile & CI Infrastructure | ⚡ Med | [🚧 Pending] | [TP-BLD-01](#test-plan-bld-01) |
| **CFG-02** | Config Extraction to Plugins | ⚡ Med | [🚧 Pending] | [TP-CFG-02](#test-plan-cfg-02) |
| **LXC-01** | LXC Provider Implementation | ⚡ Med | [🚧 Pending] | [TP-LXC-01](#test-plan-lxc-01) |
| **CFG-01** | Remote Config Pull/Sync (Deprecated) | ❄️ Low | [❌ Deprecated] | N/A |

## 🔗 Task Dependencies
- **CFG-02** depends on PLG-01 (Plugin Registry) ✅
- **VFY-02** depends on all implementation tasks
- **SCH-01** depends on config parsing structs being stable
- **BLD-01** depends on test infrastructure from VFY-02

## 📋 Detailed Test Plans

### **TP-PLG-01: Plugin Registry & DAG Resolver**
**Objective**: Ensure modular plugins are correctly discovered, namespaced, and resolved without cycles.

**Unit Tests:**
- ✅ **Recursive Discovery**: Verify plugins are found at various depths in `.vendetta/plugins/`.
- ✅ **Namespace Isolation**: Ensure rules from `plugin-a` don't leak into `plugin-b`.
- ✅ **DAG Resolution**: Verify correct loading order for dependent plugins.
- ✅ **Cycle Detection**: Fail explicitly when A -> B -> A dependency exists.

**Integration Tests:**
- ✅ **Merge Logic**: Verify that namespaced rules are correctly merged into final agent configs.
- ✅ **Local Overrides**: Project-specific plugins at `.vendetta/plugins/` take precedence over remote ones.

---

### **TP-LCK-01: Lockfile Manager**
**Objective**: Validate deterministic behavior and parallel performance.

**Unit Tests:**
- ✅ **Lockfile Generation**: Ensure `vendetta.lock` captures exact SHAs for all remotes.
- ✅ **Integrity Checks**: Verify that tampering with a plugin's manifest triggers a lock mismatch.
- ✅ **Idempotency**: `workspace create` produces identical results when run twice with the same lockfile.

**Integration Tests:**
- ✅ **Parallel Fetch**: Verify that multiple remote plugins are cloned concurrently.
- ✅ **Offline Mode**: Workspace creation succeeds if all plugins are present in cache, even without internet.

---

### **TP-AGT-03: Namespaced Agent Generation**
**Objective**: Verify that agents (Cursor, OpenCode) receive isolated, namespaced instructions.

**E2E Scenarios:**
```bash
# Test 1: Namespaced Rules in Cursor
# 1. Add plugin 'vibegear/git' and 'local/verification'
# 2. Run 'vendetta workspace create test-workspace'
# 3. Verify: .cursor/rules/git_commit.mdc exists
# 4. Verify: .cursor/rules/verification_strict.mdc exists
# 5. Verify: NO collisions between similarly named rules in different plugins
```

---

### **TP-VFY-02: E2E Test Suite & Coverage**
**Objective**: Achieve 90%+ code coverage with comprehensive E2E testing, CI integration, and performance benchmarks.

**Unit Tests:**
- ✅ **TestEnvironment Framework**: Docker-based testing environment setup
- ✅ **Coverage Reports**: Use `go test -coverprofile=coverage.out ./...` and `go tool cover -html=coverage.out`
- ✅ **Coverage Threshold**: Enforce 90%+ coverage with CI checks
- ✅ **Benchmark Tests**: Startup time <30s, memory usage <500MB using `go test -bench=.`

**Integration Tests:**
- ✅ **Full Workspace Lifecycle**: End-to-end test of create/up/down/rm with all providers
- ✅ **Multi-Service Discovery**: Verify port mapping and environment variables
- ✅ **Agent Config Generation**: Test agent-specific configuration generation

**E2E Scenarios:**
```bash
# Test 1: Full Development Workflow
# 1. Initialize project with config.yaml
# 2. Create workspace with services
# 3. Start workspace and verify services running
# 4. Test agent configs generated correctly
# 5. Stop and remove workspace
# Expected: All steps pass, coverage >90%, benchmarks met
```

---

### **TP-SCH-01: JSON Schema Validation**
**Objective**: Enable IDE autocomplete and validation for config.yaml.

**Requirements:**
- Reference implementation from https://github.com/authgear/authgear-server for automatic schema generation
- Export final schema file from Go structs without hand-crafting
- Populate https://github.com/IniZio/vendetta-config with generic sharable plugins
- Note: Unlike eslint, plugins are OFF by default (opt-in), not ON by default

**Unit Tests:**
- ✅ **Schema Generation**: Create schema/config.schema.json from Go structs (auto-generated)
- ✅ **Validation Logic**: Validate config.yaml against auto-generated schema
- ✅ **Error Reporting**: Clear error messages for invalid configs

**Integration Tests:**
- ✅ **IDE Support**: VSCode/Cursor intellisense works with auto-generated schema
- ✅ **CLI Validation**: `vendetta config validate` command works
- ✅ **Plugin Registry**: Generic plugins available from vendetta-config repo

---

### **TP-BLD-01: Makefile & CI Infrastructure**
**Objective**: Standardize development and CI workflows.

**Requirements:**
- ✅ **test-unit**: Run unit tests with coverage reporting
- ✅ **test-integration**: Run integration tests with TestEnvironment
- ✅ **test-e2e**: Run full E2E test suite with Docker cleanup
- ✅ **build**: Cross-platform binary build (Linux/Darwin amd64, arm64)
- ✅ **install**: Install binary to $GOPATH/bin or ~/.local/bin
- ✅ **lint**: Run golangci-lint with standard rules
- ✅ **fmt**: Run gofmt and ensure no formatting issues
- ✅ **CI Integration**: GitHub Actions runs all targets on PR/merge
- ✅ **Platform Agnostic**: Use Makefile variables for OS/arch detection

**Test Plan:**
```bash
# Test 1: Makefile Targets
make test-unit      # Passes with >80% coverage
make test-integration # Passes with all services
make test-e2e       # Passes with full workflows
make build          # Produces working binary
make lint           # No linting errors
make fmt            # Code properly formatted
```

---

### **TP-CFG-02: Config Extraction to Plugins**
**Objective**: Simplify sharing by extracting local .vendetta/ configurations (rules, skills, commands) into dedicated plugins for easy distribution and reuse.

**Requirements:**
- Extract custom rules, skills, and commands from .vendetta/templates/ and .vendetta/agents/
- Generate proper plugin manifest with metadata (name, version, description, author)
- Create namespace to avoid conflicts (e.g., 'team-standards' plugin)
- Allow selective extraction (rules only, skills only, or all)
- Preserve local overrides after extraction

**Unit Tests:**
- ✅ **Extraction Logic**: Convert .vendetta/ directory structure to plugin format
- ✅ **Plugin Generation**: Create plugin.yaml manifest and organized file structure
- ✅ **Namespace Handling**: Ensure extracted plugins have unique namespaces
- ✅ **Selective Extraction**: Extract specific config types (rules/skills/commands)

**Integration Tests:**
- ✅ **CLI Command**: `vendetta config extract --plugin-name my-plugin [--rules|--skills|--commands]`
- ✅ **Plugin Installation**: Extracted plugin can be installed via plugin registry
- ✅ **Override Preservation**: Local customizations remain after extraction
- ✅ **Sharing Workflow**: Extracted plugin can be committed and shared via git

**E2E Scenarios:**
```bash
# Test 1: Extract Team Coding Standards
# 1. Add custom rules to .vendetta/templates/rules/ and .vendetta/agents/cursor/rules/
# 2. Run 'vendetta config extract --plugin-name team-standards --rules'
# 3. Verify plugin created in .vendetta/plugins/team-standards/ with manifest
# 4. Commit plugin to team repo and share with colleagues
# 5. Colleagues can install via 'vendetta plugin install https://github.com/team/configs'
# Expected: Team standards propagate without manual config syncing
```

---

### **TP-LXC-01: LXC Provider Implementation**
**Objective**: Complete lightweight container provider as alternative to Docker for faster development cycles.

**Unit Tests:**
- ✅ **LXC Launch**: Successfully create and start LXC containers
- ✅ **Template Management**: Use LXC templates for quick container setup
- ✅ **Network Configuration**: Bridge networking for service communication
- ✅ **Resource Limits**: CPU/memory allocation and monitoring

**Integration Tests:**
- ✅ **Container Lifecycle**: Create/start/stop/destroy LXC containers
- ✅ **Service Integration**: Run development services inside LXC
- ✅ **Performance**: Compare startup time vs Docker (<15s for LXC)

**E2E Scenarios:**
```bash
# Test 1: LXC Workspace Creation
# 1. Run 'vendetta workspace create test --provider lxc'
# 2. Verify LXC container starts with proper networking
# 3. Install dependencies and run services
# 4. Test faster startup compared to Docker
# Expected: Container running, services accessible, startup <15s
```

---

## 🏗 Infrastructure Requirements (for handover)

### **CI Integration**
- **Artifact Locking**: CI must run `vendetta plugin check` to ensure the lockfile is up-to-date with `config.yaml`.
- **Benchmarking**: Monitor workspace creation time to ensure parallel fetching remains < 10s for 5+ remote plugins.

### **Handover Guidelines**
- **Follow TDD**: All logic in `pkg/plugins` and `pkg/lock` must have 90%+ unit test coverage.
- **Mock Git**: Use `testify/mock` to simulate git remote operations in integration tests.
- **Standard Repo**: Populate `git@github.com:IniZio/vendetta-config.git` with generic sharable plugins (coding standards, development tools, framework-specific rules).
- **Plugin Philosophy**: Plugins are OFF by default (opt-in) - unlike eslint, users must explicitly enable plugins to avoid unexpected behavior.
- **Config Extraction**: Use CFG-02 for sharing team standards instead of complex remote syncing.
