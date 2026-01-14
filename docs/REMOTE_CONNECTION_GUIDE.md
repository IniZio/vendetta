# Remote Development with vendetta M3

## OIDC Authentication with Authgear

For production use, vendetta supports OIDC authentication via Authgear:

```yaml
# .vendetta/config.yaml
auth:
  enabled: true
  issuer: "https://certain-thing-311.authgear.cloud"
  client_id: "9a0da7557c863ff9"
  # client_secret: ""  # NOT REQUIRED for public clients
```

**Client Secret**: The client secret is **NOT required** for this Authgear configuration. It uses public client flow which doesn't require client authentication.

---

## Step-by-Step Guide: Connect from Another Device on Same Network

This guide covers setting up both **local** and **remote** transport testing to prove the M3 implementation works correctly.

---

## Prerequisites

### On Both Machines (Host & Remote)

```bash
# 1. Install vendetta
curl -fsSL https://raw.githubusercontent.com/IniZio/vendetta/main/install.sh | bash

# 2. Verify installation
vendetta --version

# 3. Check available providers
# On Linux:
which docker    # for Docker provider
which lxc       # for LXC provider
which qemu-system-x86_64  # for QEMU provider

# 4. Ensure SSH server is running (for remote connections)
sudo systemctl status sshd
# If not running:
sudo systemctl start sshd
```

### Network Requirements

- Both machines must be on the **same LAN**
- Firewall must allow SSH (port 22) and MCP gateway (default: 3001)
- Note down IP addresses:
  ```bash
  # On remote machine:
  hostname -I | awk '{print $1}'
  # Example: 192.168.1.100
  ```

---

## Part 1: Local Transport Testing (Proves Local Stack Works)

### Step 1.1: Initialize a Test Project

```bash
# Create a test project directory
mkdir -p ~/vendetta-test
cd ~/vendetta-test

# Initialize vendetta
vendetta init

# Verify structure
ls -la
# Should show: .vendetta/ directory
```

### Step 1.2: Create a Local Workspace (Docker Provider)

```bash
# Create config.yaml for local Docker workspace
cat > .vendetta/config.yaml << 'EOF'
name: local-test-project
provider: docker

services:
  api:
    command: "python3 -m http.server 8080"
    port: 8080
  db:
    command: "redis-server --port 6379"
    port: 6379

docker:
  image: "python:3.11-slim"
EOF

# Create workspace
vendetta workspace create local-demo

# Verify workspace created
vendetta workspace list
# Should show: local-demo (status: created)

# Start workspace
vendetta workspace up local-demo
# Wait for Docker container to start...
```

### Step 1.3: Test Local MCP Gateway

```bash
# The MCP gateway should be running on port 3001
curl http://localhost:3001/health
# Expected response: {"status":"healthy","total_nodes":0,...}

# Check available tools via MCP
# Connect using MCP client (e.g., via Claude Desktop or Cursor)

# Verify agent configs were generated
cat .cursor/mcp.json 2>/dev/null || cat .opencode/mcp.json 2>/dev/null
```

### Step 1.4: Test Local Transport Layer Directly

```bash
# Test SSH transport (localhost simulation)
go test ./pkg/transport/... -run TestSSHTransport -v

# Test HTTP transport
go test ./pkg/transport/... -run TestHTTPTransport -v

# Test connection pooling
go test ./pkg/transport/... -run TestPool -v
```

---

## Part 2: Remote Transport Testing (Proves Remote Stack Works)

### Step 2.1: Configure SSH Access (Passwordless)

**On LOCAL machine:**
```bash
# Generate SSH key if not exists
ls ~/.ssh/id_rsa 2>/dev/null || ssh-keygen -t ed25519 -C "vendetta@local"

# Copy public key to remote machine
ssh-copy-id user@192.168.1.100
# OR manually:
cat ~/.ssh/id_rsa.pub | ssh user@192.168.1.100 'mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys'

# Test connection
ssh user@192.168.1.100 "echo 'SSH connection successful'"
```

### Step 2.2: Set Up Remote Machine

**On REMOTE machine (192.168.1.100):**
```bash
# Install vendetta
curl -fsSL https://raw.githubusercontent.com/IniZio/vendetta/main/install.sh | bash

# Create a shared directory for workspaces
mkdir -p ~/vendetta-shared
cd ~/vendetta-shared

# Initialize vendetta
vendetta init

# Create minimal config
cat > .vendetta/config.yaml << 'EOF'
name: remote-test
provider: docker
docker:
  image: "python:3.11-slim"
EOF
```

### Step 2.3: Configure Remote Access in Local Project

**On LOCAL machine, edit `~/vendetta-test/.vendetta/config.yaml`:**

```yaml
name: remote-test-project
provider: docker

# Remote node configuration
remote:
  node: "192.168.1.100"
  user: "user"
  port: 22
  # Optional: specify SSH key
  # ssh_key: "~/.ssh/id_ed25519"

services:
  api:
    command: "python3 -m http.server 8080"
    port: 8080

docker:
  image: "python:3.11-slim"
```

### Step 2.4: Create Remote Workspace

```bash
cd ~/vendetta-test

# Create workspace (will connect to remote via SSH)
vendetta workspace create remote-demo

# Verify it connected
vendetta workspace list
# Should show: remote-demo (status: created or running on remote)
```

### Step 2.5: Test Remote Execution

```bash
# Execute command on remote via transport layer
vendetta workspace shell remote-demo
# This should SSH into the remote container/VM

# Inside remote shell:
pwd  # Should show remote workspace path
ls -la
exit
```

---

## Part 3: Start Coordination Server (M3 Core Feature)

The coordination server enables provider-agnostic remote node management.

### Step 3.1: Generate Coordination Server Config

**On LOCAL machine:**
```bash
# Create coordination config
mkdir -p ~/.config/vendetta
cat > ~/.config/vendetta/coordination.yaml << 'EOF'
server:
  host: "0.0.0.0"
  port: 3001
  auth_token: "your-secure-token-change-this"

registry:
  provider: "memory"
  health_check_interval: "10s"
  node_timeout: "60s"

websocket:
  enabled: true
  path: "/ws"

auth:
  enabled: false  # Set to true for OIDC authentication
  jwt_secret: "your-jwt-secret-minimum-16-chars"
  oidc:
    issuer: "https://certain-thing-311.authgear.cloud"
    client_id: "9a0da7557c863ff9"
    # client_secret: ""  # Optional for confidential clients
    scopes: "openid profile email"

logging:
  level: "info"
  format: "json"
EOF
```

### Step 3.2: Start Coordination Server

```bash
# Start in background
vendetta coordination start --config ~/.config/vendetta/coordination.yaml

# Or run with specific port
vendetta coordination start --host 0.0.0.0 --port 3001

# Verify server is running
curl http://localhost:3001/health
# Expected: {"status":"healthy",...}

# Check metrics
curl http://localhost:3001/metrics
```

### Step 3.3: Register Remote Node

**On REMOTE machine:**
```bash
# Start node agent (connects to coordination server)
export VENDETTA_COORD_HOST="192.168.1.100"  # or localhost if same machine
export VENDETTA_COORD_PORT="3001"

vendetta agent start --coordination-url http://192.168.1.100:3001
```

**On LOCAL machine:**
```bash
# List registered nodes
curl -H "Authorization: Bearer your-secure-token" http://localhost:3001/api/v1/nodes

# Expected response:
# {"nodes":[...],"count":1}
```

---

## Part 4: Full Remote Workspace Workflow

### Step 4.1: Create Complete Remote Configuration

**On LOCAL machine, `~/vendetta-test/.vendetta/config.yaml`:**

```yaml
name: full-remote-demo
provider: docker

# Remote node specification
remote:
  node: "192.168.1.100"
  user: "user"
  port: 22
  ssh_key: "~/.ssh/id_ed25519"

# Service definitions
services:
  app:
    command: "python3 -m http.server 8080"
    port: 8080
    healthcheck:
      url: "http://localhost:8080"
      interval: "10s"
      timeout: "5s"
      retries: 3
  db:
    command: "redis-server --port 6379"
    port: 6379

# Provider-specific config
docker:
  image: "python:3.11-slim"
  network: "host"
```

### Step 4.2: Create and Start Remote Workspace

```bash
cd ~/vendetta-test

# Create workspace (prepares remote environment)
vendetta workspace create full-remote

# Start workspace (bootstraps provider on remote)
vendetta workspace up full-remote

# Monitor startup
vendetta workspace status full-remote

# Connect to remote workspace shell
vendetta workspace shell full-remote
```

### Step 4.3: Verify Remote Services

```bash
# Check service status via coordination server
curl http://localhost:3001/api/v1/services

# Expected output shows services running on remote node
```

---

## Part 5: Troubleshooting Remote Connections

### Issue: SSH Connection Failed

```bash
# 1. Verify SSH is running on remote
ssh user@192.168.1.100 "systemctl status sshd"

# 2. Check firewall
sudo ufw status
# Allow SSH if needed:
sudo ufw allow 22/tcp

# 3. Verify key-based auth
ssh -v user@192.168.1.100
# Look for "Authentications that can continue: publickey"

# 4. Test SSH command execution
ssh user@192.168.1.100 "echo 'SSH works'"
```

### Issue: Coordination Server Unreachable

```bash
# 1. Check if server is running
ps aux | grep vendetta
pgrep -a vendetta

# 2. Check port is listening
netstat -tlnp | grep 3001
# OR
ss -tlnp | grep 3001

# 3. Check firewall
sudo ufw status
sudo ufw allow 3001/tcp

# 4. Test from remote machine
curl http://192.168.1.100:3001/health
```

### Issue: Node Not Registering

```bash
# 1. Check node agent logs
# On remote machine:
journalctl -u vendetta-agent -n 100

# 2. Verify auth token matches
# coordination.yaml: auth_token: "your-secure-token"
# Node agent should use same token

# 3. Check network connectivity
nc -zv 192.168.1.100 3001
```

### Issue: Docker Provider Not Working on Remote

```bash
# On remote machine:
# 1. Verify Docker is installed and running
docker ps
sudo systemctl status docker

# 2. Add user to docker group
sudo usermod -aG docker $USER
# Then logout/login

# 3. Test Docker access
docker run --rm hello-world
```

---

## Part 6: Security for Local Network Testing

### Enable Authentication

```yaml
# ~/.config/vendetta/coordination.yaml
auth:
  enabled: true
  jwt_secret: "your-secure-minimum-16-char-secret"
  token_expiry: "24h"
```

### Use TLS for Production

```yaml
# For production, use reverse proxy with TLS (nginx, traefik, etc.)
server:
  host: "0.0.0.0"
  port: 443  # HTTPS port
  
# Configure nginx as reverse proxy:
# server {
#     listen 443 ssl;
#     server_name vendetta.example.com;
#     
#     ssl_certificate /etc/letsencrypt/live/vendetta.example.com/fullchain.pem;
#     ssl_certificate_key /etc/letsencrypt/live/vendetta.example.com/privkey.pem;
#     
#     location / {
#         proxy_pass http://localhost:3001;
#         proxy_http_version 1.1;
#         proxy_set_header Upgrade $http_upgrade;
#         proxy_set_header Connection "upgrade";
#     }
# }
```

---

## Quick Reference: Command Cheat Sheet

| Action | Command |
|--------|---------|
| Initialize project | `vendetta init` |
| Create workspace | `vendetta workspace create <name>` |
| Start workspace | `vendetta workspace up <name>` |
| Stop workspace | `vendetta workspace down <name>` |
| List workspaces | `vendetta workspace list` |
| Connect to workspace | `vendetta workspace shell <name>` |
| Remove workspace | `vendetta workspace rm <name>` |
| Start coordination server | `vendetta coordination start` |
| Start node agent | `vendetta agent start` |
| Check health | `curl http://localhost:3001/health` |
| List nodes | `curl http://localhost:3001/api/v1/nodes` |
| List services | `curl http://localhost:3001/api/v1/services` |

---

## Expected Results

After completing all steps, you should have:

1. ✅ **Local transport working**: SSH/HTTP transports tested locally
2. ✅ **Remote transport working**: Can connect to remote node via SSH
3. ✅ **Coordination server running**: HTTP API responds on port 3001
4. ✅ **Remote workspace created**: Workspace exists on remote machine
5. ✅ **Remote execution working**: Commands execute on remote via transport
6. ✅ **Node registered**: Remote node appears in coordination server
7. ✅ **Services running**: Services accessible on remote machine

---

## Architecture Reference

```
┌─────────────────────────────────────────────────────────────────┐
│                        LOCAL MACHINE                             │
│  ┌─────────────────┐    ┌──────────────────┐                    │
│  │  vendetta CLI   │───>│ Coordination     │                    │
│  │                 │    │ Server (3001)    │                    │
│  └─────────────────┘    └────────┬─────────┘                    │
│                                  │                               │
│                    SSH (port 22) │                               │
│                                  ▼                               │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │              FIREWALL / ROUTER (same LAN)                   │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                  │
                                  │ Network (192.168.1.100)
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                       REMOTE MACHINE                             │
│  ┌─────────────────┐    ┌──────────────────┐                    │
│  │ SSH Server      │<───│ vendetta Node    │                    │
│  │ (port 22)       │    │ Agent            │                    │
│  └─────────────────┘    └────────┬─────────┘                    │
│                                  │                               │
│                    Provider API │                               │
│                                  ▼                               │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Docker/LXC/QEMU Provider                                   │ │
│  │  └── Running containers/VMs for workspaces                  │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```
