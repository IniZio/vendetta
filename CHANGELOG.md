# Changelog

## [Unreleased] - 2026-01-19

### Added

#### Dynamic Port Exposure
- **Service Port Auto-Assignment**: All services defined in `.nexus/config.yaml` are now automatically exposed to the host machine with dynamically allocated ports (typically 32768-65535)
- **Port Mapping Query**: New `GetPortMappings()` method in Docker provider to retrieve actual assigned host ports after container starts
- **Environment Variable Injection**: Host port mappings are injected into containers as environment variables (`NEXUS_SERVICE_<NAME>_PORT=<host_port>`)
- **Registry Storage**: Mapped ports are stored in `DBService.LocalPort` and returned in workspace status API
- **SSH Port Update**: SSH port is now correctly updated to the actual Docker-assigned port after container starts

**Benefits**:
- No port conflicts between workspaces
- Direct external access to services without SSH tunneling
- Service discovery via environment variables
- Unlimited concurrent workspaces

**Example**:
```yaml
# .nexus/config.yaml
services:
  postgres:
    port: 5432  # → Host port 32789 (auto-assigned)
  app:
    port: 5000  # → Host port 32788 (auto-assigned)
```

Inside container: `NEXUS_SERVICE_POSTGRES_PORT=32789`

**Files Modified**:
- `pkg/provider/docker/docker.go`: Port exposure logic and GetPortMappings method
- `pkg/coordination/handlers_m4.go`: Port mapping retrieval and environment injection
- `pkg/coordination/workspace_registry.go`: Service setup with port mappings
- `pkg/provider/provider.go`: Added Services map to Session struct

#### CLI Session Management
- **Login Command**: New `nexus login` command for local authentication
  - Creates session file at `~/.nexus/session.json`
  - Session expires after 30 days
  - Supports both interactive and non-interactive modes
  - Secure file permissions (0600)

- **Logout Command**: New `nexus logout` command to clear local session

- **Session Storage**: New `pkg/auth/session.go` module with:
  - `SaveSession()` - Store session locally
  - `LoadSession()` - Load and validate session
  - `ClearSession()` - Remove session file
  - `IsLoggedIn()` - Check session status

- **Helper Functions**: New `cmd/nexus/session.go` with:
  - `requireSession()` - Ensure user is logged in
  - `getUserID()` - Extract user ID from session

**Usage**:
```bash
$ nexus login
Username (GitHub username): myuser
Password (or press Enter to skip):

✓ Login successful!
  User ID: user_1768793567_myuser
  Session expires: 2026-02-18 03:32:47
```

**Files Added**:
- `pkg/auth/session.go`: Session management core
- `cmd/nexus/login.go`: Login command implementation
- `cmd/nexus/logout.go`: Logout command implementation
- `cmd/nexus/session.go`: Session helper functions

#### User-Scoped Workspace Filtering
- **API Filtering**: Workspace listing API now supports `?user=<user_id>` query parameter
- **User Isolation**: Workspaces can be filtered by owner for multi-tenant scenarios
- **Backward Compatible**: Without `user` parameter, returns all workspaces (existing behavior)

**Example**:
```bash
# List all workspaces
GET /api/v1/workspaces

# List workspaces for specific user
GET /api/v1/workspaces?user=user_1768793567_testuser
```

**Files Modified**:
- `pkg/coordination/handlers_m4.go`: Added user filtering logic to `handleM4ListWorkspacesRouter`

### Fixed
- **Mock Interface**: Updated `MockDockerClient` in tests to implement `ContainerInspect` method
- **Test Environment Variable**: Fixed test assertion to use `NEXUS_SERVICE_*` instead of `loom_SERVICE_*`

**Files Modified**:
- `pkg/provider/docker/docker_test.go`: Added ContainerInspect to mock, fixed environment variable check

### Documentation
- **Port Management Guide**: Comprehensive guide on dynamic port exposure, service access, and troubleshooting (`docs/PORT_MANAGEMENT.md`)
- **CLI Authentication Guide**: Complete reference for session management, login/logout, and API integration (`docs/CLI_AUTHENTICATION.md`)
- **README Updates**: Added features section highlighting dynamic port exposure and session management

**Files Added**:
- `docs/PORT_MANAGEMENT.md`: Port management documentation
- `docs/CLI_AUTHENTICATION.md`: CLI authentication documentation

**Files Modified**:
- `README.md`: Updated with new features and documentation links

### Technical Details

#### Dynamic Port Allocation Flow
1. Container created with all service ports exposed (`HostPort: "0"` for auto-assign)
2. Container started
3. `ContainerInspect` queries Docker for actual assigned ports
4. Port mappings injected as environment variables via `Exec`
5. SSH port updated in workspace registry
6. Service registry updated with mapped host ports
7. Status API returns complete port information

#### Session Architecture
- Sessions stored locally in `~/.nexus/session.json`
- Format: JSON with `user_id`, `access_token`, `expires_at`, `created_at`
- File permissions: 0600 (owner read/write only)
- Validation on load: checks expiration timestamp
- No network transmission required (local-only)

#### API Changes
- **Breaking**: None (all changes backward compatible)
- **New Query Parameters**:
  - `GET /api/v1/workspaces?user=<user_id>` - Filter workspaces by owner

#### Environment Variables Added
- `NEXUS_SERVICE_<NAME>_PORT` - Host port for each service (e.g., `NEXUS_SERVICE_POSTGRES_PORT=32789`)
- Injected into `/etc/environment` inside containers
- Available to all processes after sourcing environment

### Testing
- All unit tests passing
- Dynamic port exposure verified end-to-end
- User filtering tested with multiple users
- Session management commands tested (login/logout)
- Mock interfaces updated for new methods

### Dependencies
- Added: `golang.org/x/term v0.39.0` (for password input in login command)
- Updated: `golang.org/x/sys v0.39.0 => v0.40.0` (indirect dependency)

---

## Version Information
- Branch: (development)
- Commit: (unreleased)
- Date: 2026-01-19
