# Staging Setup

Fresh staging deployment.

## Prerequisites

### macOS
```bash
brew install lxc
```

### Ubuntu
```bash
apt-get update
apt-get install lxc lxc-templates
```

### Generate SSH Key (if needed)
```bash
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519 -N ""
cat ~/.ssh/id_ed25519.pub
```

## Build Binary

From project root:
```bash
go build -o bin/nexus ./cmd/nexus
```

Check: `./bin/nexus version`

## Start Server

```bash
cd deploy/envs/staging
./ops/start.sh
```

Server: http://localhost:3001

Ctrl+C to stop.

## Register First User

In another terminal:
```bash
cd deploy/envs/staging
./ops/users.sh register alice 123456 "ssh-ed25519 AAAA..."
```

Get SSH pubkey: `cat ~/.ssh/id_ed25519.pub`

Outputs: `{user_id, github_username}`

## Create Workspace

```bash
./ops/workspaces.sh create alice my-dev-env
```

Outputs workspace JSON with `workspace_id`.

Wait for status to be "ready" (check with `./ops/workspaces.sh status {id}`).

## Connect to Workspace

From workspace status output, get `ssh_port` (usually 2236).

```bash
ssh -p 2236 dev@localhost
```

Inside:
```bash
cd /workspace        # Project code
ls -la               # Verify clone
bundle install       # Install dependencies
bundle exec puma     # Start server
```

## Access Services

From **outside** workspace:
```bash
curl http://localhost:5000   # Web service (port from service config)
```

From **inside** workspace:
```bash
curl http://localhost:5000   # Services on localhost
```

## Clean Up

Stop workspace:
```bash
./ops/workspaces.sh stop {id}
```

Delete workspace:
```bash
./ops/workspaces.sh delete {id}
```

## Customize Config

Edit `config/coordination.yaml`:

**Change port:**
```yaml
server:
  port: 3002  # or any available port
```

**Change provider:**
```yaml
provider:
  type: docker  # or qemu
```

**Change image:**
```yaml
workspace:
  image: ubuntu:20.04  # or any available image
```

Then restart: `./ops/start.sh`

## Troubleshoot

```bash
./ops/troubleshoot.sh
```

Checks: server, ports, LXC, SSH keys, workspaces.

## Common Issues

| Problem | Solution |
|---------|----------|
| Port 3001 in use | `lsof -ti:3001 \| xargs kill -9` |
| SSH connection refused | Check workspace status: `./ops/workspaces.sh status {id}` |
| LXC not found | Install LXC (see Prerequisites) |
| Workspace won't start | Check logs: `./ops/start.sh` (server logs) |
| Cannot clone repo | Check git SSH key: `ssh -T git@github.com` |
