# GitHub App Integration Specification

## Overview
Replace GitHub CLI with GitHub App authentication to enable git push/pull in remote workspaces, supporting Gitpod-like collaborative development experience.

## Architecture

### Authentication Flow
```
User creates workspace
  → Coordination Server creates GitHub App installation
  → User authorizes app at: https://github.com/apps/[app-name]/installations/new?state=...
  → GitHub redirects to: https://linuxbox.tail31e11.ts.net/auth/github/callback
  → Server exchanges authorization code for installation_id + user_id
  → Server generates short-lived installation access token (1 hour expiry)
  → Workspace receives token, uses for git push/pull
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
// New DB table: github_installations
type GitHubInstallation struct {
  InstallationID  int64     // GitHub App installation ID
  UserID          string    // nexus user_id (from DBUser)
  GitHubUserID    int64     // GitHub user ID
  RepoFullName    string    // owner/repo (per-installation, can be multiple)
  Token           string    // Installation access token (1-hour expiry)
  TokenExpiresAt  time.Time
  CreatedAt       time.Time
  UpdatedAt       time.Time
}
```

### Security Considerations
- Store GitHub App private key in secure config (environment variable)
- Installation tokens expire after 1 hour (automatic rotation)
- CSRF protection: `state` parameter in OAuth flow
- Scoped permissions: minimal required per repo
- No user credentials stored on server (tokens only)

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
- [ ] GitHub App created and registered
- [ ] OAuth callback handler implemented and tested
- [ ] Workspace can clone private repos with installation token
- [ ] Workspace can push commits using installation token
- [ ] Installation token automatically refreshed before expiry
- [ ] Multi-user: Different users can authorize for different repos
- [ ] All `gh` CLI dependencies removed from server code
