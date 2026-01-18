# Phase 4: Error Handling & Polish Guide

**Purpose**: Production-ready error responses and user experience  
**Priority**: High (affects user perception)  
**Effort**: 40 hours

---

## Error Response Standards

All API errors must follow this structure:

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "status": 400,
  "details": {
    "field": "optional_additional_context"
  },
  "request_id": "req-abc123"
}
```

---

## HTTP Status Codes

| Code | Use Case | Example |
|------|----------|---------|
| 200 | Success | Workspace status retrieved |
| 201 | Created | Workspace created |
| 202 | Accepted | Workspace creation started (async) |
| 400 | Bad Request | Invalid JSON, missing field |
| 401 | Unauthorized | GitHub auth failed |
| 409 | Conflict | Workspace already exists |
| 500 | Server Error | Unexpected error |

---

## API Error Codes

### User Errors (4xx)

**auth_required** (401)
- User not authenticated
- Solution: Run `nexus auth github`

**invalid_request** (400)
- Malformed JSON or missing fields
- Include `details.missing_fields` array

**repo_not_found** (404)
- Repository doesn't exist or not accessible
- Include `details.owner` and `details.repo`

**workspace_exists** (409)
- Workspace with name already exists
- Include `details.workspace_name`

**invalid_provider** (400)
- Provider (docker/lxc/qemu) not available
- Include `details.provider` and `details.available_providers`

**insufficient_resources** (503)
- Not enough disk/memory to create workspace
- Include `details.available_disk` and `details.required_disk`

### Server Errors (5xx)

**internal_error** (500)
- Unexpected error (log full stack trace)
- Include `details.trace_id` for support

**provider_error** (500)
- LXC/Docker/QEMU operation failed
- Include `details.provider` and `details.operation`

**database_error** (500)
- Database operation failed
- Include `details.operation`

---

## Error Handling by Component

### GitHub CLI Integration

**Scenario**: User not authenticated with GitHub

```go
err := github.AuthenticateWithGH()
if errors.Is(err, github.ErrNotAuthenticated) {
    return &ErrorResponse{
        Code:    "auth_required",
        Message: "Please authenticate with GitHub: nexus auth github",
        Status:  http.StatusUnauthorized,
        Details: map[string]interface{}{
            "command": "nexus auth github",
        },
    }
}
```

**Scenario**: SSH key upload fails (already exists)

```go
err := ssh.UploadPublicKeyToGitHub(pubKey)
if err != nil {
    if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "Key already exists") {
        return &ErrorResponse{
            Code:    "ssh_key_exists",
            Message: "SSH key already registered with GitHub. Remove old key or use existing.",
            Status:  http.StatusConflict,
        }
    }
}
```

### Workspace Operations

**Scenario**: Provider not available

```go
if provider == "lxc" && !isLXCAvailable() {
    return &ErrorResponse{
        Code:    "provider_not_available",
        Message: "LXC provider not available. Install: apt-get install lxc",
        Status:  http.StatusServiceUnavailable,
        Details: map[string]interface{}{
            "provider":           "lxc",
            "install_command":    "apt-get install lxc",
            "supported_providers": []string{"docker", "qemu"},
        },
    }
}
```

**Scenario**: Insufficient resources

```go
if availableDisk < requiredDisk {
    return &ErrorResponse{
        Code:    "insufficient_resources",
        Message: fmt.Sprintf("Not enough disk space. Need %dGB, have %dGB",
            requiredDisk/1e9, availableDisk/1e9),
        Status:  http.StatusServiceUnavailable,
        Details: map[string]interface{}{
            "available_disk": availableDisk,
            "required_disk":  requiredDisk,
            "suggestion":     "Delete unused workspaces or increase disk size",
        },
    }
}
```

---

## CLI Error Messages

### User-Friendly Messages

Instead of:
```
error: unable to create workspace
```

Use:
```
Error: Failed to create workspace.

Reason: SSH key not found

Solution:
  1. Generate SSH key: ssh-keygen -t ed25519
  2. Upload to GitHub: nexus ssh setup
  3. Try again: nexus workspace create owner/repo

Need help? Check: docs/troubleshooting.md
```

### Implementation Pattern

```go
func handleError(err error) {
    switch err.(type) {
    case *RepoNotFoundError:
        fmt.Fprintf(os.Stderr, `Error: Repository not found

Repo: %s
URL:  %s

Check:
  1. Repository exists and is public/accessible
  2. You have permission to access it
  3. Repository URL is correct

Help: nexus help workspace create
`, err.Repo, err.URL)

    case *InsufficientResourcesError:
        fmt.Fprintf(os.Stderr, `Error: Insufficient resources

Required: %s
Available: %s

Solution: Delete unused workspaces or increase server capacity
`, err.Required, err.Available)

    default:
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    }
}
```

---

## Input Validation

### Repository URLs

```go
func ValidateRepoURL(url string) error {
    if !strings.HasPrefix(url, "https://github.com/") &&
       !strings.HasPrefix(url, "git@github.com:") {
        return &ValidationError{
            Field:   "repo",
            Message: "Only GitHub repositories supported",
            Example: "https://github.com/owner/repo.git",
        }
    }
    return nil
}
```

### Workspace Names

```go
func ValidateWorkspaceName(name string) error {
    if len(name) > 64 {
        return &ValidationError{
            Field:   "workspace_name",
            Message: "Name too long (max 64 characters)",
        }
    }
    if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(name) {
        return &ValidationError{
            Field:   "workspace_name",
            Message: "Name must contain only lowercase letters, numbers, and hyphens",
            Example: "my-feature-branch",
        }
    }
    return nil
}
```

### SSH Keys

```go
func ValidateSSHKey(pubKey string) error {
    if !strings.HasPrefix(pubKey, "ssh-ed25519 ") &&
       !strings.HasPrefix(pubKey, "ssh-rsa ") {
        return &ValidationError{
            Field:   "ssh_pubkey",
            Message: "Invalid SSH key format",
            Example: "ssh-ed25519 AAAA... user@host",
        }
    }
    return nil
}
```

---

## Logging & Observability

### Request Logging

```go
type RequestLog struct {
    RequestID  string        `json:"request_id"`
    Method     string        `json:"method"`
    Path       string        `json:"path"`
    Status     int           `json:"status"`
    Duration   time.Duration `json:"duration_ms"`
    Error      string        `json:"error,omitempty"`
    UserID     string        `json:"user_id,omitempty"`
    Timestamp  time.Time     `json:"timestamp"`
}
```

All requests logged with:
- Unique request ID (for support)
- Duration (performance monitoring)
- Error details (debugging)
- User ID (auditing)

### Error Logging

```go
func logError(ctx context.Context, err error) {
    log.WithError(err).
        WithField("request_id", ctx.Value("request_id")).
        WithField("user_id", ctx.Value("user_id")).
        WithField("component", "workspace_creation").
        WithField("operation", "provision_container").
        Error("Workspace creation failed")
}
```

---

## Timeout Handling

### Default Timeouts

| Operation | Timeout | Rationale |
|-----------|---------|-----------|
| API request | 30s | HTTP standard |
| Workspace creation | 5min | Container startup slow |
| Workspace deletion | 60s | Container cleanup fast |
| Git clone | 5min | Network dependent |
| SSH connection | 10s | Fast network operation |
| Service health check | 30s | Service startup time |

### Timeout Messages

```
Error: Workspace creation timed out (5 minutes exceeded)

This usually means:
  1. Server is overloaded (try again later)
  2. Container startup is slow (check server logs)
  3. Network is slow (check connection)

Status: Check workspace: nexus workspace status WORKSPACE_ID
Logs:   Check server logs for details
Help:   Contact: support@nexus.dev
```

---

## Graceful Degradation

### Feature Fallback

If GitHub CLI not installed:
```
⚠️  GitHub CLI not found. Installing...

If this hangs, install manually:
  macOS: brew install gh
  Ubuntu: apt-get install gh

Or use manual setup:
  nexus auth github --manual
```

### Provider Fallback

If LXC not available:
```
⚠️  LXC not available. Trying Docker...

✅ Docker found. Using Docker provider.

To use LXC instead:
  Install: apt-get install lxc
  Reconfigure: nexus config set provider lxc
```

---

## Testing Error Scenarios

All error conditions must be tested:

### Unit Tests

```go
func TestErrorResponse_BadRequest(t *testing.T) {
    req := &CreateWorkspaceRequest{WorkspaceName: ""}
    err := validateWorkspaceRequest(req)
    
    assert.NotNil(t, err)
    assert.Equal(t, "invalid_request", err.Code)
    assert.Equal(t, http.StatusBadRequest, err.Status)
}

func TestErrorResponse_Conflict(t *testing.T) {
    createWorkspace("my-ws")
    req := &CreateWorkspaceRequest{WorkspaceName: "my-ws"}
    
    err := createWorkspace("my-ws")
    
    assert.NotNil(t, err)
    assert.Equal(t, "workspace_exists", err.Code)
    assert.Equal(t, http.StatusConflict, err.Status)
}
```

### Integration Tests

```go
func TestErrorHandling_GHAuthFailed(t *testing.T) {
    // Mock gh CLI to fail
    os.Setenv("PATH", "/nonexistent:"+os.Getenv("PATH"))
    
    err := nexus.AuthGitHub()
    
    assert.Equal(t, "auth_required", err.Code)
    assert.Contains(t, err.Message, "nexus auth github")
}
```

---

## Success Criteria (Phase 4)

- ✅ All API errors follow standard format
- ✅ All error codes documented
- ✅ CLI errors include actionable solutions
- ✅ Input validation on all endpoints
- ✅ Timeout handling with clear messages
- ✅ Graceful degradation implemented
- ✅ Request logging enabled
- ✅ Error test coverage >95%

---

## Polish Checklist

- [ ] All error messages reviewed for clarity
- [ ] CLI help messages complete
- [ ] Logging configured for production
- [ ] Timeout values tested and tuned
- [ ] Fallback providers configured
- [ ] Error documentation complete
- [ ] Support contact info displayed in errors
- [ ] Error tests pass 100%

---

**Next**: CI/CD setup in PHASE_4_CI_CD.md
