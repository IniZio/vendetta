# Testing eDealer on Fresh Remote Machine - Coordination Server Architecture

This guide walks you through testing the eDealer workspace on a fresh remote machine that connects to the coordination server running on the current host (linuxbox).

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│ CURRENT HOST (linuxbox)                                         │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Coordination Server (runs on 0.0.0.0:3001)              │   │
│  │ - Manages workspaces across all nodes                   │   │
│  │ - Provides registry of available providers              │   │
│  │ - Handles workspace lifecycle                           │   │
│  └─────────────────────────────────────────────────────────┘   │
│                          ▲                                       │
│                          │ HTTP/WebSocket                        │
│                          │ (port 3001)                           │
└──────────────────────────┼───────────────────────────────────────┘
                           │
                    Network Connection
                           │
┌──────────────────────────┼───────────────────────────────────────┐
│ REMOTE MACHINE (fresh)   │                                       │
│                          ▼                                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Nexus CLI                                               │   │
│  │ - Connects to coordination server on linuxbox:3001      │   │
│  │ - Sends workspace creation requests                     │   │
│  │ - Manages local workspace operations                    │   │
│  └─────────────────────────────────────────────────────────┘   │
│                          │                                       │
│                          ▼                                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ LXC Containers (created on linuxbox or remote)         │   │
│  │ - edealer workspace with all services                  │   │
│  │ - Accessible via SSH from remote machine               │   │
│  │ - Connected via coordination server registry           │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

### On Current Host (linuxbox)

**Already running:**
- Coordination server on `0.0.0.0:3001`
- LXC provider configured
- Network accessible to remote machine

**Verify it's running:**
```bash
curl http://localhost:3001/health
# Should respond with: {"status":"ok"}

netstat -tuln | grep 3001
# Should show: 0.0.0.0:3001
```

### On Fresh Remote Machine

Install only:
- Go 1.24+ (for nexus CLI)
- Git
- SSH client (usually pre-installed)
- curl/jq (for testing)

**No LXC needed on remote machine!** The remote connects to the coordination server which manages workspaces.

## Step-by-Step Testing

### **STEP 1: Prepare Current Host (linuxbox)**

Ensure coordination server is running:
```bash
# On linuxbox, in nexus repo directory
./bin/nexus coordination start &

# Verify it's listening
sleep 2
curl -s http://0.0.0.0:3001/health | jq .
```

**Note the IP address or hostname of linuxbox for the next steps.**

---

### **STEP 2: Install Prerequisites on Fresh Remote Machine**

```bash
# Install Go 1.24
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Git
sudo apt-get update
sudo apt-get install -y git curl jq

# Install SSH keys (if needed for workspace access)
# Copy your SSH public key to ~/.ssh/authorized_keys if not already there
```

---

### **STEP 3: Clone Nexus Repository on Remote Machine**

```bash
cd ~/projects  # or your preferred directory
git clone https://github.com/IniZio/nexus.git
cd nexus

# Verify you have the edealer configuration
ls -la examples/epson-eshop/edealer/.nexus/
ls -la examples/epson-eshop/edealer/tests/
```

---

### **STEP 4: Build Nexus CLI on Remote Machine**

```bash
make build

# Verify build succeeded
./bin/nexus --version
```

---

### **STEP 5: Configure Connection to Coordination Server**

The nexus CLI needs to know where the coordination server is running.

**Option A: Set environment variable (temporary)**
```bash
export NEXUS_COORDINATION_URL=http://<linuxbox-ip>:3001

# Test connection
curl -s $NEXUS_COORDINATION_URL/health | jq .
```

**Option B: Configure in nexus config (persistent)**
```bash
# Edit or create ~/.nexus/config.yaml
cat > ~/.nexus/config.yaml << 'EOF'
name: nexus-remote-client
coordination:
  url: http://<linuxbox-ip>:3001
  timeout: 30s
EOF

# Test connection
./bin/nexus workspace list
```

Replace `<linuxbox-ip>` with the actual IP or hostname of your current host running the coordination server.

---

### **STEP 6: Verify Connection to Coordination Server**

```bash
# From remote machine, test coordination server connection
curl -s http://<linuxbox-ip>:3001/health | jq .

# Should respond:
# {
#   "status": "ok"
# }

# If you've configured nexus, also test:
./bin/nexus workspace list
# Should show existing workspaces (if any)
```

---

### **STEP 7: Create eDealer Workspace via Coordination Server**

```bash
# From remote machine
./bin/nexus workspace create \
  --config examples/epson-eshop/edealer/.nexus/config.yaml \
  --coordination-url http://<linuxbox-ip>:3001 \
  edealer

# OR if you set environment variable:
./bin/nexus workspace create \
  --config examples/epson-eshop/edealer/.nexus/config.yaml \
  edealer

# The coordination server on linuxbox will:
# 1. Receive the workspace creation request
# 2. Create LXC container with edealer services
# 3. Initialize database with seed data
# 4. Return workspace details to remote CLI

# Wait 2-3 minutes for workspace initialization
./bin/nexus workspace show edealer
```

---

### **STEP 8: Access Workspace from Remote Machine**

```bash
# Get workspace details
./bin/nexus workspace show edealer

# SSH into workspace
# Get SSH key path from workspace info
./bin/nexus workspace show edealer --ssh

# SSH connection (replace <workspace-host> with actual host)
ssh -i ~/.ssh/nexus_edealer_key dev@<workspace-host>

# Inside workspace, verify services
psql -h db -U postgres -d edealer_development -c "SELECT version();"
redis-cli -h redis-db ping

exit
```

---

### **STEP 9: Run Playwright Tests from Remote Machine**

```bash
# Access workspace via nexus CLI
./bin/nexus workspace shell edealer

# Inside the workspace shell:
cd /workspace

# Install Playwright
npm install

# Run tests
npm run playwright:test

# Expected: 7/7 tests PASS
```

---

### **STEP 10: Access Web Application**

The web application runs in the workspace (created on linuxbox or remote LXC):

```bash
# Get workspace network info
./bin/nexus workspace show edealer

# Access from remote machine:
# - User Portal: http://<workspace-ip>:23100
# - Admin Portal: http://<workspace-ip>:23100/admins

# OR if you have network access:
curl http://localhost:23100/up | jq .
```

---

### **STEP 11: Generate Test Report**

```bash
# From workspace shell
npm run playwright:report

exit

# Report is accessible at:
# examples/epson-eshop/edealer/playwright-report/index.html
```

---

### **STEP 12: Verify Database and Services**

```bash
# Check all services are healthy
./bin/nexus workspace show edealer --services

# Verify seed data
./bin/nexus workspace shell edealer -c \
  "cd /workspace && bundle exec rails runner 'puts Admin.pluck(:email).inspect'"

# Expected: ["admin@example.com"]
```

---

### **STEP 13: Cleanup**

```bash
# Delete workspace (from remote machine)
./bin/nexus workspace delete edealer

# Coordination server on linuxbox continues running for other workspaces
```

---

## Troubleshooting

### Issue: Cannot Connect to Coordination Server

```bash
# Check coordination server is running (on linuxbox)
curl http://localhost:3001/health

# Check network connectivity from remote
ping <linuxbox-ip>
telnet <linuxbox-ip> 3001

# Check firewall
sudo ufw allow 3001/tcp  # if using UFW

# Verify URL is correct
echo $NEXUS_COORDINATION_URL
# Should be: http://<linuxbox-ip>:3001
```

### Issue: Workspace Creation Fails

```bash
# Check coordination server logs (on linuxbox)
journalctl -u nexus -f

# Check workspace status
./bin/nexus workspace show edealer --logs

# Verify LXC provider is available
lxc list  # on linuxbox
```

### Issue: Cannot SSH into Workspace

```bash
# Verify SSH key was generated
ls -la ~/.ssh/nexus_edealer_key

# Check SSH access
ssh-keyscan -p 22 <workspace-ip>

# Try with verbose output
ssh -vvv -i ~/.ssh/nexus_edealer_key dev@<workspace-ip>
```

### Issue: Playwright Tests Fail

```bash
# Install Chromium
./bin/nexus workspace shell edealer -c "npx playwright install"

# Run with verbose output
./bin/nexus workspace shell edealer -c \
  "cd /workspace && npm run playwright:test -- --verbose"
```

---

## Performance Expectations

| Operation | Time | Notes |
|-----------|------|-------|
| Install prerequisites | 5-10 min | One-time on remote |
| Clone repository | 2-5 min | One-time |
| Build nexus CLI | 2-5 min | One-time |
| Connect to coordination server | < 1 sec | Network latency |
| Create workspace | 2-3 min | Via coordination server |
| Initialize database | ~30 sec | Seed data loading |
| Start services | ~30 sec | Health checks |
| Run Playwright tests | 2-3 min | Browser automation |
| **Total (first run)** | **~25 min** | Mostly setup |
| **Subsequent runs** | **~5 min** | Recreate + test |

---

## Success Checklist

- [ ] Prerequisites installed on remote machine
- [ ] Nexus repository cloned
- [ ] Nexus CLI built successfully
- [ ] Connected to coordination server on linuxbox
- [ ] Workspace created via coordination server
- [ ] Can SSH into workspace
- [ ] Playwright tests: 7/7 PASS
- [ ] Test report generated
- [ ] Seed data verified (admin account exists)
- [ ] All services healthy
- [ ] Web application accessible
- [ ] Workspace deleted successfully

---

## Key Differences from Local Testing

| Aspect | Local | Remote Client |
|--------|-------|---------------|
| **Coordination Server** | Runs on same machine | Runs on linuxbox (network call) |
| **Workspace Creation** | `./bin/nexus workspace create` | API call to coordination server |
| **Service Port Access** | `localhost:23100` | Network address of workspace |
| **SSH Access** | SSH to localhost | SSH to workspace IP/host |
| **CLI Installation** | Full build with dependencies | Only CLI needed |
| **Setup Time** | ~15-20 min | ~25 min (includes network setup) |

---

## Advanced Usage

### Create Multiple Workspaces

```bash
# Create multiple workspaces for parallel testing
./bin/nexus workspace create \
  --config examples/epson-eshop/edealer/.nexus/config.yaml \
  edealer-1

./bin/nexus workspace create \
  --config examples/epson-eshop/edealer/.nexus/config.yaml \
  edealer-2

# Run tests in parallel
./bin/nexus workspace shell edealer-1 -c "cd /workspace && npm run playwright:test" &
./bin/nexus workspace shell edealer-2 -c "cd /workspace && npm run playwright:test" &

wait
```

### Monitor Coordination Server

```bash
# On linuxbox
curl http://localhost:3001/workspaces | jq .

# Shows all workspaces managed by coordination server
```

### Share Workspace URL

```bash
# Get workspace details
./bin/nexus workspace show edealer

# Share with team:
# - User Portal: http://<workspace-ip>:23100
# - Admin Portal: http://<workspace-ip>:23100/admins
# - Admin credentials: admin@example.com / 1234Qwer!
```

---

## TL;DR - Quick Commands

```bash
# On CURRENT HOST (linuxbox) - already done
./bin/nexus coordination start &

# On REMOTE MACHINE - fresh setup
export PATH=$PATH:/usr/local/go/bin
export NEXUS_COORDINATION_URL=http://<linuxbox-ip>:3001

git clone https://github.com/IniZio/nexus.git && cd nexus
make build

# Create and test
./bin/nexus workspace create \
  --config examples/epson-eshop/edealer/.nexus/config.yaml \
  edealer

./bin/nexus workspace shell edealer -c \
  "cd /workspace && npm install && npm run playwright:test"

# Cleanup
./bin/nexus workspace delete edealer
```

---

## Documentation References

- **Testing Guide**: `edealer-testing-guide.md`
- **Workspace Setup**: `examples/epson-eshop/edealer/.nexus/SETUP_GUIDE.md`
- **Main README**: `README.md`
- **Coordination Server Docs**: `docs/COORDINATION.md` (if available)
