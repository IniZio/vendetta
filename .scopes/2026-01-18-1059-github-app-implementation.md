# Scope: GitHub App Integration for Remote Workspaces

**Date**: 2026-01-18  
**Status**: Planning  
**Spec**: `.scopes/GITHUB_APP_SPEC.md`  
**Hosting**: https://linuxbox.tail31e11.ts.net/

---

## Task Breakdown

### Phase 1: Setup & Configuration

#### 1.1 Create GitHub App Registration
- **Owner**: Manual (one-time setup)
- **Actions**:
  1. Go to https://github.com/settings/apps/new
  2. Fill form with spec details:
     - App name: `nexus-workspace-automation`
     - Homepage URL: `https://linuxbox.tail31e11.ts.net/`
     - Callback URL: `https://linuxbox.tail31e11.ts.net/auth/github/callback`
     - Permissions: contents (read,write), metadata (read)
  3. Create app, save:
     - App ID
     - Client ID  
     - Client Secret (store in env: `GITHUB_APP_CLIENT_SECRET`)
     - Private Key (PEM format, store as `GITHUB_APP_PRIVATE_KEY`)
- **Deliverable**: App credentials in `.env.local` (gitignored)

#### 1.2 Configuration Infrastructure
- **Files to create**:
  - `pkg/github/app.go` — GitHub App config struct + initialization
  - `pkg/github/oauth.go` — OAuth flow handlers
- **Implementation**:
  ```go
  type AppConfig struct {
    AppID          int64
    ClientID       string
    ClientSecret   string
    PrivateKey     []byte
    RedirectURL    string  // https://linuxbox.tail31e11.ts.net/auth/github/callback
  }
  ```
- **Env vars**:
  - `GITHUB_APP_ID`
  - `GITHUB_APP_CLIENT_ID`
  - `GITHUB_APP_CLIENT_SECRET`
  - `GITHUB_APP_PRIVATE_KEY` (base64-encoded PEM)

---

### Phase 2: Core Implementation

#### 2.1 Database Model
- **File**: `pkg/coordination/models.go`
- **New struct**: `GitHubInstallation`
  ```go
  type GitHubInstallation struct {
    InstallationID  int64
    UserID          string        // From DBUser
    GitHubUserID    int64         // GitHub user ID
    RepoFullName    string        // owner/repo
    Token           string        // Access token (encrypted in DB)
    TokenExpiresAt  time.Time
    CreatedAt       time.Time
    UpdatedAt       time.Time
  }
  ```
- **DB Migration**: Add table `github_installations` (if using SQL)

#### 2.2 GitHub App Token Manager
- **File**: `pkg/github/app.go` (new)
- **Functions**:
  - `GenerateJWT(appID, privateKey)` → JWT for server auth
  - `ExchangeCodeForToken(code, state)` → Exchange OAuth code for installation ID + user info
  - `GenerateInstallationAccessToken(appID, privateKey, installationID)` → Get short-lived token
  - `RefreshInstallationToken(db, userID)` → Rotate token before expiry
- **Implementation notes**:
  - Use `golang-jwt/jwt` for JWT generation
  - Use `net/http` for GitHub REST API calls
  - Implement CSRF protection with `state` token

#### 2.3 OAuth Callback Handler
- **File**: `pkg/coordination/handlers.go`
- **New handler**: `HandleGitHubOAuthCallback(w, r)`
  ```go
  func HandleGitHubOAuthCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")
    
    // Validate state (CSRF protection)
    if !validateState(state) {
      http.Error(w, "Invalid state", http.StatusBadRequest)
      return
    }
    
    // Exchange code for installation ID + user info
    installation, err := app.ExchangeCodeForToken(code)
    
    // Store in DB
    db.CreateGitHubInstallation(installation)
    
    // Redirect to workspace setup success
    http.Redirect(w, r, "/workspace/auth-success", http.StatusSeeOther)
  }
  ```
- **Route**: `POST /auth/github/callback`

#### 2.4 Workspace Authorization Flow
- **File**: `pkg/coordination/handlers_m4.go`
- **Modification**: `CreateWorkspace` handler
  - Check if user has GitHub installation
  - If not: Return `auth_required` status with authorization URL
  - If yes: Use stored `installation_id` to get access token
- **New response field**:
  ```go
  type M4CreateWorkspaceResponse struct {
    // ... existing fields ...
    GitHubAuthURL  string // URL to authorize GitHub App if needed
    AuthRequired   bool   // true if user hasn't authorized app yet
  }
  ```

---

### Phase 3: Integration with Workspace Operations

#### 3.1 Git Credentials in Workspace
- **File**: `pkg/coordination/handlers.go`
- **New function**: `GetGitCredentials(userID, workspaceID)`
  - Return installation access token
  - Validate token hasn't expired (< 5 min remaining → refresh)
- **Workspace receives credentials**:
  - Installation token via secure API endpoint
  - Used in `~/.netrc` or `~/.git-credentials` for HTTPS operations
  - OR passed to workspace as env var: `GITHUB_TOKEN`

#### 3.2 Remove GitHub CLI
- **Files to delete**:
  - `pkg/github/auth.go` (all `gh` CLI functions)
  - `pkg/ssh/upload.go` (SSH key upload)
- **Files to modify**:
  - `cmd/nexus/auth.go` — Remove `authGitHubCmd` (keep `authStatusCmd`)
  - `cmd/nexus/workspace.go` — Use workspace server auth instead of local `gh`
  - `pkg/github/workspace.go` — Replace `gh` calls with API or direct SSH
- **Remove dependency**: `exec.Command("gh", ...)`

#### 3.3 Go-Git Integration
- **File**: `pkg/templates/manager.go`
- **Update**: Add support for authenticated git operations
  ```go
  type TemplateRepo struct {
    URL    string
    Branch string
    Auth   transport.AuthMethod  // Add this
  }
  
  func (m *Manager) cloneRepo(repo TemplateRepo, repoDir string) error {
    options := &git.CloneOptions{
      URL:  repo.URL,
      Auth: repo.Auth,  // Use installation token auth
    }
    // ...
  }
  ```

---

### Phase 4: Testing & Verification

#### 4.1 Unit Tests
- **File**: `pkg/github/app_test.go` (new)
- Tests for:
  - JWT generation
  - OAuth code exchange (mock GitHub)
  - Token refresh logic
- **Coverage**: 80%+

#### 4.2 Integration Tests
- **File**: `pkg/coordination/handlers_test.go`
- Tests for:
  - OAuth callback flow (mock GitHub responses)
  - Workspace creation with auth flow
  - Token retrieval in workspace
- **Setup**: Mock GitHub API server

#### 4.3 E2E Test
- **Manual**:
  1. Register real GitHub App
  2. Visit `https://linuxbox.tail31e11.ts.net/workspace/auth/github` (generates authorization URL)
  3. Authorize app on GitHub.com (redirects to callback)
  4. Verify installation stored in DB
  5. Create workspace with repo
  6. Inside workspace: `git clone` with token, verify it works
  7. Inside workspace: `git push` verify write access

---

## Implementation Order (Dependencies)

```
1. Create GitHub App (manual setup)
   ↓
2. Add configuration infrastructure (2.1)
   ↓
3. Implement token manager (2.2)
   ↓
4. Implement OAuth callback (2.3)
   ↓
5. Integrate with workspace creation (2.4 + 3.1)
   ↓
6. Remove old GitHub CLI code (3.2)
   ↓
7. Update git operations (3.3)
   ↓
8. Tests & verification (4.1 + 4.2 + 4.3)
```

## Code Changes Summary

| File | Change | Type |
|------|--------|------|
| `pkg/github/app.go` | New | Create |
| `pkg/github/oauth.go` | New | Create |
| `pkg/coordination/models.go` | GitHubInstallation | Add |
| `pkg/coordination/handlers.go` | OAuth callback + GetGitCredentials | Add |
| `pkg/coordination/handlers_m4.go` | AuthRequired + GitHubAuthURL | Modify |
| `pkg/github/auth.go` | All functions | Delete |
| `pkg/github/workspace.go` | Remove `gh` calls | Modify |
| `pkg/ssh/upload.go` | SSH key upload | Delete |
| `cmd/nexus/auth.go` | Remove authGitHubCmd | Modify |
| `cmd/nexus/workspace.go` | Remove local `gh` usage | Modify |
| `go.mod` | Remove `gh` dependency | Modify |
| `pkg/github/app_test.go` | JWT, exchange, refresh tests | Create |

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Token expiry race condition | Workspace loses access mid-operation | Pre-refresh tokens at 5min mark, implement retry logic |
| GitHub App not authorized | Workspace creation fails | User-friendly error messages + auth URL |
| Installation token generation fails | Workspace stuck | Fallback: manual token endpoint for testing |
| Tailscale DNS changes | OAuth callback unreachable | Monitor DNS, set up alerting |
| Multiple workspaces same user | Token reuse conflicts | One token per workspace, implement rotation |

## Success Acceptance Criteria

- [ ] GitHub App created and configured
- [ ] OAuth callback receives authorization code
- [ ] Installation ID stored in database
- [ ] Workspace can retrieve installation token
- [ ] Workspace clones private repo with token
- [ ] Workspace pushes commits with token
- [ ] Token auto-refreshes before expiry
- [ ] All `gh` CLI code removed
- [ ] Tests passing (>80% coverage)
- [ ] E2E test successful with real GitHub
- [ ] Multi-user scenarios tested

## Timeline Estimate

- Phase 1 (Setup): 15 min
- Phase 2 (Core): 4-6 hours
- Phase 3 (Integration): 2-3 hours
- Phase 4 (Testing): 2-3 hours
- **Total**: ~1-2 sprint days

## Delegation Readiness

✅ All requirements clear
✅ Specification written
✅ Code locations identified  
✅ Test strategy defined
✅ Hosting confirmed  
✅ GitHub App credentials needed (manual setup)

**Next**: Delegate to backend-dev for implementation starting with Phase 1.
