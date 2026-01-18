# Scope: Nexus Runtime Folder Structure Refactor

**Date**: 2026-01-18 13:30 UTC  
**Status**: ✅ COMPLETE  
**Effort**: ~5-6 hours  
**Complexity**: Medium (affects multiple subsystems)  

---

## Why This Matters

Currently, `.nexus/` mixes committed configuration with runtime artifacts (nexus.db, server.pid, server.log). This creates:
- ❌ Gitignore confusion (too many patterns)
- ❌ Accidental commits of database files
- ❌ Unclear which files are versioned
- ❌ Poor production deployment structure

**Solution**: Separate into `.nexus/` (config) and `.nexus-runtime/` (data), with centralized path resolution supporting environment variables for production (`/var/lib/nexus/`).

---

## Design Overview

### Committed Files (`.nexus/`)
```
.nexus/
├── config.yaml          # Project configuration
├── agents/              # Agent templates (Cursor, Claude, etc)
├── templates/           # Reusable command/skill templates
├── hooks/               # Lifecycle hooks (up.sh, dev.sh, teardown.sh)
├── remotes/             # Remote plugin repos (git submodules)
└── plugins/             # Local plugins
```

### Runtime Files (`.nexus-runtime/` - GITIGNORED)
```
.nexus-runtime/
├── data/
│   ├── nexus.db         # SQLite database
│   ├── nexus.db-shm     # SQLite shared memory
│   └── nexus.db-wal     # SQLite write-ahead log
├── state/
│   ├── server.pid       # Server process ID
│   ├── server.lock      # Server lock file
│   └── worktrees/       # Workspace state (moved from .nexus/worktrees/)
├── logs/
│   ├── server.log       # Coordination server logs
│   ├── usage.json       # Metrics/usage logs
│   └── archive/         # Rotated logs
└── cache/
    ├── ssh/             # SSH key cache
    ├── templates/       # Compiled templates
    └── plugins/         # Plugin cache
```

### User-Local (per-machine, NOT versioned)
```
~/.config/nexus/         # User config
~/.cache/nexus/          # User cache
```

### Production (system-wide)
```
/var/lib/nexus/data/     # System data
/var/lib/nexus/state/    # System state
/var/log/nexus/          # System logs
/var/cache/nexus/        # System cache
```

---

## Implementation Phases

### Phase 1: Path Resolution Layer (1-2 hours) ⏳ START HERE
**File**: `pkg/paths/paths.go` (NEW)

Functions to implement:
- `GetDataDir(projectRoot string) string` → `.nexus-runtime/data/` or `$NEXUS_DATA_DIR`
- `GetStateDir(projectRoot string) string` → `.nexus-runtime/state/` or `$NEXUS_STATE_DIR`
- `GetLogsDir(projectRoot string) string` → `.nexus-runtime/logs/` or `$NEXUS_LOGS_DIR`
- `GetCacheDir(projectRoot string) string` → `~/.cache/nexus/` or `$NEXUS_CACHE_DIR`
- `GetConfigDir(projectRoot string) string` → `.nexus/`
- `GetPIDFile(projectRoot string) string` → `GetStateDir() + "/server.pid"`
- `GetDatabasePath(projectRoot string) string` → `GetDataDir() + "/nexus.db"`
- `EnsureDir(path string) error` → Create directory with proper permissions

Tests: `pkg/paths/paths_test.go` - test all env var overrides

### Phase 2: Update Path Usage (1-2 hours) ⏳ NEXT
Update these files to use `pkg/paths`:

1. **`pkg/coordination/server.go`**
   - Line ~29: `StartServer()` - use `paths.GetDatabasePath()` for DB_PATH default
   - Line ~40: Set PID file location

2. **`pkg/coordination/registry_sqlite.go`**
   - Update `NewSQLiteRegistry()` to use `paths.GetDatabasePath()`
   - Ensure directory created with `paths.EnsureDir()`

3. **`pkg/coordination/config.go`**
   - Update `LoadConfig()` to respect config path
   - Generate default config to `.nexus-runtime/coordination.yaml`

4. **`pkg/metrics/logger.go`** (if exists)
   - Log files to `paths.GetLogsDir()`
   - Create archive directory

5. **`pkg/ctrl/ctrl.go`**
   - `Init()` function: Create both `.nexus/` and `.nexus-runtime/` directories
   - Worktrees path: `.nexus-runtime/state/worktrees/`

6. **`cmd/nexus/main.go`**
   - `coordinationRestartCmd`: Use `paths.GetPIDFile()` for PID file
   - Any other paths

### Phase 3: Update Initialization (30 minutes) ⏳ THEN
**File**: `pkg/ctrl/ctrl.go` - Update `Init()` function

Should create:
```
.nexus/
├── config.yaml
├── agents/
├── templates/
├── hooks/
├── plugins/
└── remotes/

.nexus-runtime/
├── data/
├── state/worktrees/
├── logs/archive/
└── cache/
```

### Phase 4: Update Gitignore (5 minutes) ⏳ THEN
**File**: `.gitignore`

Replace:
```gitignore
.nexus/worktrees/
.nexus/plugins/
.nexus/*.lock
.nexus/*.db
.nexus/server.pid
.nexus/server.log
```

With:
```gitignore
# Runtime artifacts - NEVER commit
.nexus-runtime/
```

Also remove `.nexus/plugins/` from gitignore since plugins ARE committed now.

### Phase 5: Migration Script (30 minutes) ⏳ THEN
**File**: `scripts/migrate-runtime-dirs.sh` (NEW)

Script to migrate existing projects:
```bash
#!/bin/bash
# Migrate .nexus/ runtime files to .nexus-runtime/

mkdir -p .nexus-runtime/{data,state/worktrees,logs/archive,cache/{ssh,templates,plugins}}

# Move data
[ -f .nexus/nexus.db ] && mv .nexus/nexus.db .nexus-runtime/data/
[ -f .nexus/nexus.db-shm ] && mv .nexus/nexus.db-shm .nexus-runtime/data/
[ -f .nexus/nexus.db-wal ] && mv .nexus/nexus.db-wal .nexus-runtime/data/

# Move state
[ -d .nexus/worktrees ] && mv .nexus/worktrees/* .nexus-runtime/state/worktrees/ 2>/dev/null
[ -f .nexus/server.pid ] && mv .nexus/server.pid .nexus-runtime/state/

# Move logs
[ -f .nexus/server.log ] && mv .nexus/server.log .nexus-runtime/logs/
[ -f .nexus/usage.json ] && mv .nexus/usage.json .nexus-runtime/logs/

echo "✅ Migration complete"
```

### Phase 6: Testing (1 hour) ⏳ THEN
- `pkg/paths/paths_test.go` - Unit tests for all path functions
- Test environment variable overrides
- Test directory creation
- Manual test: Run server, verify files appear in `.nexus-runtime/`

### Phase 7: Documentation & Deployment Examples (30 min) ⏳ THEN
Update:
- `DEPLOYMENT_GUIDE.md` - Update path references
- Add Docker example with `/var/lib/nexus/` volumes
- Add Kubernetes example with environment variables
- Add migration guide for existing projects

### Phase 8: Commit & Verify (30 min) ⏳ LAST
- Commit with comprehensive message
- Rebuild and test
- Verify staging deployment works with new structure
- Tag release if ready

---

## File Change Summary

| Phase | Files | Action |
|-------|-------|--------|
| 1 | `pkg/paths/paths.go` (NEW) | Create path resolution layer |
| 1 | `pkg/paths/paths_test.go` (NEW) | Unit tests |
| 2 | `pkg/coordination/server.go` | Update StartServer() |
| 2 | `pkg/coordination/registry_sqlite.go` | Use paths.GetDatabasePath() |
| 2 | `pkg/coordination/config.go` | Update config loading |
| 2 | `pkg/ctrl/ctrl.go` | Update Init(), worktrees path |
| 2 | `cmd/nexus/main.go` | Update PID file handling |
| 3 | `pkg/ctrl/ctrl.go` | Create full directory structure |
| 4 | `.gitignore` | Simplify to `.nexus-runtime/` |
| 5 | `scripts/migrate-runtime-dirs.sh` (NEW) | Migration script |
| 6 | Multiple test files | Add comprehensive tests |
| 7 | `DEPLOYMENT_GUIDE.md` | Update examples |
| 7 | `README.md` | Document new structure |
| 8 | All above | Final commit |

---

## Resumption Points

**If stopping mid-implementation, note where to resume:**

- **After Phase 1**: Path layer created and tested. Ready for file updates.
- **After Phase 2**: All files updated to use paths. Ready for init changes.
- **After Phase 3**: Full directory structure created. Ready for gitignore update.
- **After Phase 4**: Gitignore clean. Ready for migration script.
- **After Phase 5**: Migration script ready. Ready for testing.
- **After Phase 6**: All tests passing. Ready for docs.
- **After Phase 7**: Docs updated. Ready for commit.

---

## Success Criteria

- ✅ All path functions tested and working
- ✅ Server starts and creates files in `.nexus-runtime/`
- ✅ Database persists correctly
- ✅ `.gitignore` has single `.nexus-runtime/` entry
- ✅ Migration script works for existing projects
- ✅ Docker deployment works with `/var/lib/nexus/`
- ✅ Staging server deploys successfully
- ✅ All tests pass

---

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Breaking existing projects | Migration script + backward compatibility period |
| Files in wrong places | Comprehensive testing before release |
| Path confusion | Centralized `pkg/paths/` - single source of truth |
| Production deployment issues | Test with Docker/K8s before release |

---

## Timeline

```
Phase 1: 1-2 hours
Phase 2: 1-2 hours
Phase 3: 30 minutes
Phase 4: 5 minutes
Phase 5: 30 minutes
Phase 6: 1 hour
Phase 7: 30 minutes
Phase 8: 30 minutes
─────────────────
Total: ~5-6 hours
```

**Estimated completion**: Within 1-1.5 hours of focused work

---

## Next Action

**START WITH PHASE 1**: Create `pkg/paths/paths.go` and implement path resolution layer.

When resuming after break: Return to this file, note which phase was last completed, and continue from next phase.
