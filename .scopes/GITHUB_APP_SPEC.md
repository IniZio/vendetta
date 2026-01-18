# GitHub App Integration Specification

## Overview
Replace GitHub CLI with GitHub App authentication to enable git push/pull in remote workspaces, supporting Gitpod-like collaborative development experience.

## Architecture

### Authentication Flow (USER-BASED - ACTUAL)
```
User creates workspace
  → Coordination Server checks GitHub authorization
  → If not authorized: Return OAuth URL
  → User authorizes app at: https://github.com/login/oauth/authorize?client_id=...&scope=repo
  → GitHub redirects to: https://linuxbox.tail31e11.ts.net/auth/github/callback?code=...
  → Server exchanges code for USER ACCESS TOKEN (tied to user, not app)
  → Server stores token with user_id
  → User auto-registered during authorization
  → Workspace receives token, uses for git operations
  → Commits appear as the USER (not as bot)
```

### GitHub App Configuration
- **Name**: nexus-workspace-automation
- **Homepage URL**: https://linuxbox.tail31e11.ts.net/
- **Authorization callback URL**: https://linuxbox.tail31e11.ts.net/auth/github/callback
- **Permissions** (minimum required):
  - `contents:read,write` — Clone, read, commit, push
  - `pull_requests:read,write` — Optional, for PR operations
  - `metadata:read` — Repository metadata
- **Events**: None required initially (server-initiated, not webhook-based)
- **Installation Type**: Per-user (each user authorizes individually)

### OAuth Callback Handler
```
POST /auth/github/callback
  ├─ Receive: code, state
  ├─ Validate: state token (CSRF protection)
  ├─ Exchange: code → installation_id, user_id, access_token
  ├─ Store: DB record linking user_id → installation_id
  └─ Return: Redirect to workspace setup success page
```

### Git Operations in Workspace
```
// Token usage in workspace:
git clone https://x-access-token:[token]@github.com/owner/repo.git
// OR via go-git:
auth := &http.BasicAuth{
  Username: "x-access-token",
  Password: token,
}
```

### Data Model
```go
// Table: github_installations (currently in-memory, needs SQLite persistence)
type GitHubInstallation struct {
  InstallationID  int64     // GitHub App installation ID (0 for user-based auth)
  UserID          string    // nexus user_id (GitHub username for MVP)
  GitHubUserID    int64     // GitHub user ID (numeric)
  GitHubUsername  string    // GitHub username (for lookup key)
  RepoFullName    string    // owner/repo (stored for audit/logging)
  Token           string    // User access token (NOT installation token)
  TokenExpiresAt  time.Time // Token expiry timestamp
  CreatedAt       time.Time
  UpdatedAt       time.Time
}

// Table: github_forks (tracks forked repos per user)
type GitHubFork struct {
  UserID          string    // nexus user_id
  OriginalOwner   string    // Original repo owner (e.g., "oursky")
  OriginalRepo    string    // Original repo name (e.g., "epson-eshop")
  ForkOwner       string    // Fork owner (user's account)
  ForkRepo        string    // Fork name (same as original)
  ForkURL         string    // Full fork HTTPS URL
  CreatedAt       time.Time
}
```

### Fork Management (NEW - REQUIRED FOR GIT PUSH)
For private repos owned by other users/orgs, workspace must NOT push directly. Instead:
1. **Auto-fork**: When workspace is created for private repo, auto-fork to user's account
2. **Track fork**: Record fork in `github_forks` table
3. **Use fork URL**: Workspace clones from fork (user has write access)
4. **Allow direct**: For repos user owns, push directly (no fork needed)

**Fork Detection Logic:**
```
If repo is private AND user doesn't own it AND repo not in whitelist:
  → Auto-fork to user's account
  → Store fork mapping
  → Use fork URL for workspace
Else:
  → Use repo URL as-is
```

### Security Considerations
- Store GitHub App private key in secure config (environment variable)
- User access tokens used for git operations (not app tokens)
- CSRF protection: `state` parameter in OAuth flow
- Scoped permissions: `repo` scope (read/write all repos user can access)
- No user credentials stored on server (tokens only)
- Never push to external orgs without explicit permission

## Removed Functionality
- `gh` CLI entirely removed from coordination server
- SSH key uploads to GitHub removed (use Deploy Keys for read-only access if needed)
- Manual authentication flow replaced with OAuth

## Affected Components
1. **Coordination Server**
   - `pkg/coordination/handlers.go` — Add OAuth callback handler
   - `pkg/coordination/auth.go` (new) — GitHub App token management
   - `pkg/coordination/models.go` — Add GitHubInstallation model
   - `pkg/coordination/server.go` — Initialize OAuth config

2. **Workspace Creation Flow**
   - `pkg/coordination/handlers_m4.go` — Redirect to GitHub App authorization URL
   - Workspace stores GitHub installation ID (not user credentials)

3. **Removed/Deprecated**
   - `pkg/github/auth.go` — Remove all `gh` CLI calls
   - `cmd/nexus/auth.go` — Remove `gh auth github` command
   - `pkg/ssh/upload.go` — Remove SSH key upload to GitHub

## Hosting Requirement
- **Public Address**: https://linuxbox.tail31e11.ts.net/ (provided, Tailscale subnet)
- **Callback endpoint**: Must be accessible from GitHub's servers
- **TLS/HTTPS**: Required by GitHub (already satisfied by Tailscale HTTPS)

## Success Criteria

### Phase 1: GitHub App OAuth (✅ COMPLETED)
- [x] GitHub App created and registered
- [x] OAuth callback handler implemented and tested
- [x] User-based authentication (user access tokens)
- [x] Workspace creation requires GitHub auth
- [x] Auto-user registration during authorization
- [x] All `gh` CLI dependencies removed from server code
- [x] CSRF protection with state tokens
- [x] In-memory GitHub installation storage

### Phase 2: Git Operations & Fork Management (🚧 IN PROGRESS)
- [ ] Auto-fork private repos to user's account
- [ ] Track fork mappings in database
- [ ] Workspace clones from fork (not original)
- [ ] Test: git clone with user token
- [ ] Test: git commit and push to fork
- [ ] Test: Push permissions for owned repos
- [ ] Pass token to workspace environment (`GITHUB_TOKEN`)
- [ ] Git credentials configuration in workspace

### Phase 3: Data Persistence (⏳ TODO - NEXT SESSION)
- [ ] Migrate from in-memory to SQLite
- [ ] Persist GitHubInstallation records
- [ ] Persist GitHubFork records
- [ ] Implement user registry in SQLite
- [ ] Database migrations/initialization
- [ ] Connection pooling
