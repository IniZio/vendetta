# CLI Authentication

## Overview

Nexus CLI supports local session management to authenticate users and scope workspace operations. Sessions are stored locally in `~/.nexus/session.json` and expire after 30 days.

## Getting Started

### Login

Authenticate with Nexus to create a local session:

```bash
nexus login
```

**Interactive prompt**:
```
Nexus Login
===========

Username (GitHub username): your-username
Password (or press Enter to skip): 

✓ Login successful!
  User ID: user_1768793567_your-username
  Session expires: 2026-02-18 03:32:47

Your session has been saved to ~/.nexus/session.json
You can now use 'nexus workspace' commands.
```

### Check Login Status

```bash
# Verify session is active
test -f ~/.nexus/session.json && echo "Logged in" || echo "Not logged in"

# View session details
cat ~/.nexus/session.json | jq
```

**Example output**:
```json
{
  "user_id": "user_1768793567_testuser",
  "access_token": "local-session-token",
  "expires_at": "2026-02-18T03:32:47Z",
  "created_at": "2026-01-19T03:32:47Z"
}
```

### Logout

Clear your local session:

```bash
nexus logout
```

Output:
```
✓ Successfully logged out
Your local session has been cleared.
```

## Session Management

### Session File Location

Sessions are stored at: `~/.nexus/session.json`

### Session Expiration

- **Default duration**: 30 days
- Sessions automatically expire after the expiration time
- Expired sessions require re-login

### Session Security

- Session file has permissions `0600` (owner read/write only)
- Stored locally on your machine only
- Not transmitted over network except for API authentication

## Using Sessions with API

### Filtering Workspaces by User

When logged in, CLI commands can filter workspaces by your user ID:

```bash
# Get your user ID from session
USER_ID=$(cat ~/.nexus/session.json | jq -r '.user_id')

# List only your workspaces
curl "http://localhost:3001/api/v1/workspaces?user=$USER_ID"
```

### Example: List My Workspaces

```bash
#!/bin/bash
SESSION_FILE=~/.nexus/session.json

if [ ! -f "$SESSION_FILE" ]; then
  echo "Not logged in. Run 'nexus login' first."
  exit 1
fi

USER_ID=$(jq -r '.user_id' "$SESSION_FILE")
EXPIRES_AT=$(jq -r '.expires_at' "$SESSION_FILE")

# Check if expired
if [[ $(date -d "$EXPIRES_AT" +%s) -lt $(date +%s) ]]; then
  echo "Session expired. Please run 'nexus login' again."
  exit 1
fi

# List workspaces for this user
curl -s "http://localhost:3001/api/v1/workspaces?user=$USER_ID" | jq
```

## Non-Interactive Login

For automation and scripts, you can provide credentials non-interactively:

```bash
# Using echo
echo "username" | nexus login

# Using here-doc
nexus login <<EOF
username

EOF
```

**Note**: Password prompt is skipped in non-interactive mode for security.

## API Integration

### User-Scoped Endpoints

The workspace listing API supports user filtering:

```bash
# List all workspaces (no filter)
GET /api/v1/workspaces

# List workspaces for specific user
GET /api/v1/workspaces?user=user_1768793567_testuser
```

**Response format**:
```json
{
  "workspaces": [
    {
      "id": null,
      "name": "my-workspace",
      "owner": "user_1768793567_testuser",
      "status": "running",
      "provider": "docker",
      "ssh_port": 32787,
      "created_at": "2026-01-19T03:34:24Z",
      "services_count": 3
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

### Authentication Header (Future)

Currently, the API uses query parameters for user filtering. In future versions, authentication may use headers:

```bash
# Proposed future format
curl http://localhost:3001/api/v1/workspaces \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Troubleshooting

### Session Not Found

**Error**: `no active session found. Please run 'nexus login' first`

**Solution**: Run `nexus login` to create a new session

### Session Expired

**Error**: `session expired. Please run 'nexus login' again`

**Solution**: Re-authenticate with `nexus login`

### Permission Denied

**Error**: `failed to write session file: permission denied`

**Solution**: Ensure `~/.nexus` directory is writable:
```bash
mkdir -p ~/.nexus
chmod 700 ~/.nexus
```

### Session File Corrupted

**Error**: `failed to unmarshal session: ...`

**Solution**: Remove corrupted file and re-login:
```bash
rm ~/.nexus/session.json
nexus login
```

## Security Best Practices

1. **Never share your session file**
   - Treat `~/.nexus/session.json` like a password
   - Don't commit it to version control
   - Don't copy it to shared locations

2. **Use logout when done**
   - Run `nexus logout` on shared machines
   - Clear sessions before giving away devices

3. **Monitor expiration**
   - Sessions expire after 30 days
   - Re-authenticate periodically for security

4. **File permissions**
   - Session files are created with `0600` permissions
   - Only your user account can read the file
   - Don't change these permissions

## Integration with CI/CD

For CI/CD pipelines, consider these approaches:

### Option 1: Store User ID as Secret

```yaml
# .github/workflows/test.yml
env:
  NEXUS_USER_ID: ${{ secrets.NEXUS_USER_ID }}

steps:
  - name: List workspaces
    run: |
      curl "http://nexus-server:3001/api/v1/workspaces?user=$NEXUS_USER_ID"
```

### Option 2: Create Session in Pipeline

```bash
#!/bin/bash
# ci-login.sh

# Create session programmatically
cat > ~/.nexus/session.json <<EOF
{
  "user_id": "${CI_USER_ID}",
  "access_token": "${CI_ACCESS_TOKEN}",
  "expires_at": "$(date -u -d '+30 days' --rfc-3339=seconds | sed 's/ /T/' | sed 's/+00:00/Z/')",
  "created_at": "$(date -u --rfc-3339=seconds | sed 's/ /T/' | sed 's/+00:00/Z/')"
}
EOF

chmod 600 ~/.nexus/session.json
```

## Future Enhancements

Planned authentication improvements:

1. **OAuth Integration**: GitHub OAuth for seamless authentication
2. **SSO Support**: Enterprise SSO providers
3. **API Keys**: Long-lived tokens for automation
4. **Role-Based Access**: Team workspaces and permissions
5. **Multi-Factor Auth**: Enhanced security for production use

## See Also

- [Port Management](./PORT_MANAGEMENT.md) - Accessing workspace services
- [Workspace Commands](./WORKSPACE_COMMANDS.md) - CLI reference
- [API Reference](./API_REFERENCE.md) - Complete API documentation
