# Milestone: M2_Alpha (Namespaced Plugins & Determinism)

**Objective**: Transition Vendatta into a modular, plugin-based platform with hierarchical discovery and `uv`-style version locking. This phase focuses on developer experience, parallel performance, and environment reproducibility.

## 🎯 Success Criteria
- [ ] `pkg/plugins` registry implements hierarchical discovery and DAG resolution.
- [ ] `vendatta.lock` ensures deterministic environment recreation across machines.
- [ ] Remote plugins are fetched in parallel with 0 resolution overhead during workspace creation.
- [ ] Agents receive namespaced capabilities (Rules, Skills, Commands).
- [ ] Standard config moved to `git@github.com:IniZio/vendatta-config.git`.

## 🛠 Implementation Tasks

| ID | Title | Priority | Status | Test Plan |
| :--- | :--- | :--- | :--- | :--- |
| **PLG-01** | Plugin Registry & DAG Resolver | 🔥 High | [✅ Completed] | [TP-PLG-01](#test-plan-plg-01) |
| **LCK-01** | Lockfile Manager (uv-style) | 🔥 High | [✅ Completed] | [TP-LCK-01](#test-plan-lck-01) |
| **CLI-04** | `vendatta plugin` Command Group | ⚡ Med | [✅ Completed] | [TP-CLI-04](#test-plan-cli-04) |
| **AGT-03** | Namespaced Agent Generation | ⚡ Med | [✅ Completed] | [TP-AGT-03](#test-plan-agt-03) |

## 📋 Detailed Test Plans

### **TP-PLG-01: Plugin Registry & DAG Resolver**
**Objective**: Ensure modular plugins are correctly discovered, namespaced, and resolved without cycles.

**Unit Tests:**
- ✅ **Recursive Discovery**: Verify plugins are found at various depths in `.vendatta/plugins/`.
- ✅ **Namespace Isolation**: Ensure rules from `plugin-a` don't leak into `plugin-b`.
- ✅ **DAG Resolution**: Verify correct loading order for dependent plugins.
- ✅ **Cycle Detection**: Fail explicitly when A -> B -> A dependency exists.

**Integration Tests:**
- ✅ **Merge Logic**: Verify that namespaced rules are correctly merged into final agent configs.
- ✅ **Local Overrides**: Project-specific plugins at `.vendatta/plugins/` take precedence over remote ones.

---

### **TP-LCK-01: Lockfile Manager**
**Objective**: Validate deterministic behavior and parallel performance.

**Unit Tests:**
- ✅ **Lockfile Generation**: Ensure `vendatta.lock` captures exact SHAs for all remotes.
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
# 2. Run 'vendatta workspace create test-workspace'
# 3. Verify: .cursor/rules/git_commit.mdc exists
# 4. Verify: .cursor/rules/verification_strict.mdc exists
# 5. Verify: NO collisions between similarly named rules in different plugins
```

---

## 🏗 Infrastructure Requirements (for handover)

### **CI Integration**
- **Artifact Locking**: CI must run `vendatta plugin check` to ensure the lockfile is up-to-date with `config.yaml`.
- **Benchmarking**: Monitor workspace creation time to ensure parallel fetching remains < 10s for 5+ remote plugins.

### **Handover Guidelines**
- **Follow TDD**: All logic in `pkg/plugins` and `pkg/lock` must have 90%+ unit test coverage.
- **Mock Git**: Use `testify/mock` to simulate git remote operations in integration tests.
- **Standard Repo**: Use `git@github.com:IniZio/vendatta-config.git` as the reference for all "Standard" capability tests.
