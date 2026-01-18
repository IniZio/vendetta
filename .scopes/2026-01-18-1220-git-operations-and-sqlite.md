# Scope: Git Operations & SQLite Persistence for GitHub-Managed Workspaces

**Date**: 2026-01-18  
**Status**: ✅ COMPLETE  
**Spec**: `.scopes/GITHUB_APP_SPEC.md`  
**Previous Phase**: Phase 1 OAuth (Completed)  
**Current Status**: Phase 2 & 3 Combined

---

## Overview

Phase 1 (GitHub App OAuth) is **✅ COMPLETED**. Now implementing:
- **Phase 2**: Git operations and fork management (automatic forking for private repos)
- **Phase 3**: SQLite persistence (replace in-memory storage)

---

## Phase 2A: Fork Management

### 2A.1 Fork API Integration
- **File**: `pkg/github/fork.go` (new)
- **Functions**:
  - `GetUserRepos(ctx, token)` — List repos user owns
  - `ForkRepository(ctx, token, owner, repo)` — Fork repo to user account
  - `IsForkOf(token, userRepo, origOwner, origRepo)` — Verify fork relationship
  - `GetForkURL(token, owner, repo)` — Get fork URL for user
- **Implementation**:
  - Use GitHub REST API: `POST /user/repos/{template_owner}/{template_repo}/forks`
  - Handle already-forked case (idempotent)
  - Extract fork URL from response

### 2A.2 Fork Detection in Workspace Creation
- **File**: `pkg/coordination/handlers_m4.go`
- **Update `handleM4CreateWorkspace`**:
  ```
  Before provisioning:
    1. Get repo metadata (public/private, owner)
    2. If repo is private AND user doesn't own it:
       → Call ForkRepository(token, origOwner, origRepo)
       → Store fork mapping in database
       → Use fork URL for workspace.RepoURL
    3. If repo is owned by user OR public:
       → Use original URL as-is
  ```
- **New response fields** (to `M4CreateWorkspaceResponse`):
  - `fork_created: bool` — Whether fork was created
  - `fork_url: string` — Fork HTTPS URL if created

### 2A.3 Fork Tracking Database
- **Table**: `github_forks`
- **Fields**:
  ```go
  UserID string
  OriginalOwner string
  OriginalRepo string
  ForkOwner string
  ForkURL string
  CreatedAt time.Time
  ```
- **Duplicate prevention**: `UNIQUE(UserID, OriginalOwner, OriginalRepo)`

---

## Phase 2B: Git Operations in Workspace

### 2B.1 Token Injection into Workspace
- **File**: `pkg/coordination/handlers_m4.go` (modify `provisionWorkspace`)
- **Pass token to workspace via**:
  - Option A (simpler): Environment variable `GITHUB_TOKEN`
  - Option B (secure): `.netrc` file in workspace home
  - **Use**: Option A for MVP (set in workspace startup script)

### 2B.2 Test Git Operations
- **Manual test** (via workspace SSH):
  ```bash
  # Inside workspace (user=dev, at ~/)
  echo "export GITHUB_TOKEN=$GITHUB_TOKEN" >> ~/.bashrc
  source ~/.bashrc
  
  # Test clone (fork was auto-created)
  git clone https://${GITHUB_TOKEN}@github.com/$USER/epson-eshop.git
  cd epson-eshop
  
  # Test commit
  git config user.name "Test User"
  git config user.email "test@example.com"
  echo "test" >> test.txt
  git add test.txt
  git commit -m "test: workspace integration"
  
  # Test push
  git push origin main
  ```
- **Automated test** (optional):
  - Create integration test that SSHes into workspace
  - Runs git operations
  - Verifies commits appear on GitHub fork

---

## Phase 3: SQLite Persistence

### 3.1 Database Schema
- **File**: `pkg/coordination/db.go` (new)
- **Tables**:
  ```sql
  CREATE TABLE github_installations (
    id INTEGER PRIMARY KEY,
    installation_id INTEGER,
    user_id TEXT NOT NULL UNIQUE,
    github_user_id INTEGER NOT NULL,
    github_username TEXT NOT NULL,
    repo_full_name TEXT,
    token TEXT NOT NULL,
    token_expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  
  CREATE TABLE github_forks (
    id INTEGER PRIMARY KEY,
    user_id TEXT NOT NULL,
    original_owner TEXT NOT NULL,
    original_repo TEXT NOT NULL,
    fork_owner TEXT NOT NULL,
    fork_url TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, original_owner, original_repo)
  );
  
  CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    public_key TEXT,
    workspace_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  
  CREATE TABLE workspaces (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    workspace_name TEXT NOT NULL,
    status TEXT,
    provider TEXT,
    image TEXT,
    repo_owner TEXT,
    repo_name TEXT,
    repo_url TEXT,
    repo_branch TEXT,
    repo_commit TEXT,
    fork_created BOOLEAN DEFAULT 0,
    fork_url TEXT,
    ssh_port INTEGER,
    ssh_host TEXT,
    node_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id)
  );
  ```

### 3.2 Registry Implementation with SQLite
- **File**: `pkg/coordination/registry_sqlite.go` (new)
- **Implement**:
  - `SQLiteRegistry` — implements `Registry` interface
  - `SQLiteUserRegistry` — implements `UserRegistry` interface
  - Connection pooling (sync.Pool for sqlite3 connections)
  - Migration runner (auto-create tables on startup)
- **Replace in-memory**:
  - `NewInMemoryRegistry()` → `NewSQLiteRegistry(dbPath)`
  - `NewInMemoryUserRegistry()` → Embedded in SQLiteRegistry

### 3.3 GitHub Installation Persistence
- **File**: `pkg/coordination/registry_sqlite.go`
- **Methods**:
  - `StoreGitHubInstallation(installation *GitHubInstallation)` → INSERT/UPDATE
  - `GetGitHubInstallation(userID string)` → SELECT
  - `DeleteGitHubInstallation(userID string)` → DELETE

### 3.4 Migration Strategy
- **Startup**:
  1. Check if DB file exists
  2. If not: Create file, run all migrations
  3. If yes: Check schema version, run pending migrations
- **Migrations**:
  - Version 1: Initial schema (tables above)
  - Version 2+: Future schema changes
- **Storage**: Migration version in `_schema_version` table

### 3.5 Database Configuration
- **File**: `pkg/coordination/server.go`
- **Environment variable**: `DB_PATH` (default: `.nexus/nexus.db`)
- **Initialize DB** in `NewServer()`:
  ```go
  db, err := NewSQLiteRegistry(os.Getenv("DB_PATH"))
  if err != nil {
    panic(fmt.Sprintf("Failed to initialize database: %v", err))
  }
  srv.registry = db
  ```

---

## Implementation Order (Dependencies)

```
Phase 2A (Fork Management)
├─ 2A.1: Fork API integration (pkg/github/fork.go)
├─ 2A.2: Fork detection in workspace creation
└─ 2A.3: Fork tracking DB table (schema-only, in-memory for now)
   ↓
Phase 2B (Git Operations)
├─ 2B.1: Token injection into workspace
└─ 2B.2: Test git operations (manual E2E)
   ↓
Phase 3 (SQLite Persistence) - Can start in parallel
├─ 3.1: Design schema
├─ 3.2: Registry implementation
├─ 3.3: GitHub installation persistence
├─ 3.4: Migration runner
└─ 3.5: Database configuration
   ↓
Integration & Testing
├─ Migrate existing in-memory data
├─ Verify persistence across restarts
└─ Full E2E test: Auth → Fork → Workspace → Git operations
```

---

## Code Changes Summary

| File | Change | Phase |
|------|--------|-------|
| `pkg/github/fork.go` | New fork API integration | 2A |
| `pkg/coordination/handlers_m4.go` | Add fork detection + token injection | 2A, 2B |
| `pkg/coordination/models.go` | Add GitHubFork model | 2A |
| `pkg/coordination/db.go` | Database schema + migrations | 3 |
| `pkg/coordination/registry_sqlite.go` | SQLite registry implementation | 3 |
| `pkg/coordination/server.go` | Initialize SQLite DB | 3 |
| `go.mod` | Add sqlite3 driver | 3 |

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Fork API rate limits | Workspace creation fails | Cache fork status, batch requests |
| Fork already exists | Idempotent operation expected | Check if fork exists before creating |
| Token not available in workspace | Git clone/push fails | Test token passing early |
| Database corruption | Data loss, service unavailable | Backup before migrations, use transactions |
| SQLite concurrent writes | Race conditions | Use WAL mode, connection pooling |

---

## Testing Strategy

### Unit Tests
- `pkg/github/fork_test.go`: Fork API mocking
- `pkg/coordination/registry_sqlite_test.go`: CRUD operations
- Mock GitHub API responses

### Integration Tests
- Fork creation + workspace provisioning
- Token passing to workspace
- Database persistence across restarts

### E2E Test Flow
1. ✅ User authorizes via GitHub (already working)
2. ⏳ Workspace created for private repo
3. ⏳ Auto-fork to user account
4. ⏳ Token available in workspace
5. ⏳ git clone works
6. ⏳ git push works
7. ⏳ Commits visible on GitHub fork
8. ⏳ Restart coordination server
9. ⏳ Data still persists (SQLite)

---

## Acceptance Criteria

- [ ] Fork detection correctly identifies private/owned repos
- [ ] Auto-fork works for private repos not owned by user
- [ ] Fork tracking accurate in database
- [ ] GitHub token injected into workspace environment
- [ ] Git clone works with token (authenticated access)
- [ ] Git commit works (user identity preserved)
- [ ] Git push works to fork (no direct push to external orgs)
- [ ] SQLite schema created on startup
- [ ] Schema migrations run automatically
- [ ] All in-memory data persists in SQLite
- [ ] Service restart preserves GitHub installations + forks
- [ ] No direct pushes to "oursky" org repos (only to forks)

---

## Timeline Estimate

- **Phase 2A (Fork Management)**: 2-3 hours
  - GitHub API integration
  - Fork detection logic
  - Database schema design

- **Phase 2B (Git Operations)**: 1-2 hours
  - Token injection
  - Manual E2E testing

- **Phase 3 (SQLite Persistence)**: 3-4 hours
  - Schema + migrations
  - Registry implementation
  - Data migration + testing

- **Total**: ~8-10 hours (1-1.5 sprint days)

---

## Delegation Readiness

✅ All requirements clear
✅ Specification written
✅ Code locations identified  
✅ Test strategy defined
✅ Architecture decisions made
✅ Risk analysis completed

**Next**: Delegate to backend-dev for Phase 2A (Fork Management) starting with GitHub API integration.
