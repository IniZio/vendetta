# Staging Environment

Self-contained deploy guide. All ops here.

## Start Staging

```bash
cd deploy/envs/staging
./ops/start.sh
```

Server: `http://localhost:3001`  
Logs: stdout (stop with Ctrl+C)

## First Time

### 1. Register User
```bash
./ops/users.sh register alice 123456 "ssh-ed25519 AAAA..."
```

- `alice` = GitHub username
- `123456` = GitHub user ID
- SSH pubkey from `cat ~/.ssh/id_ed25519.pub`

### 2. Create Workspace
```bash
./ops/workspaces.sh create alice feature-x
```

Creates workspace from `epson-eshop` repo. Outputs workspace ID.

### 3. Connect
```bash
ssh -p 2236 dev@localhost
```

Inside:
```bash
cd /workspace  # Your code
bundle install
bundle exec puma
```

## Daily Ops

| Task | Command |
|------|---------|
| Start server | `./ops/start.sh` |
| List workspaces | `./ops/workspaces.sh list` |
| Get workspace status | `./ops/workspaces.sh status <id>` |
| Stop workspace | `./ops/workspaces.sh stop <id>` |
| Delete workspace | `./ops/workspaces.sh delete <id>` |
| SSH into workspace | `ssh -p 2236 dev@localhost` |
| Health check | `curl http://localhost:3001/health` |

## Config

Edit `config/coordination.yaml` before starting:
- Port: Change `server.port`
- Provider: Change `provider.type` (lxc/docker/qemu)
- Image: Change `image` (ubuntu:22.04, etc)

Then start: `./ops/start.sh`

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Port 3001 in use | `lsof -ti:3001 \| xargs kill -9` |
| SSH port conflict | Check `config/coordination.yaml` → `services.ssh.port` |
| LXC not available | `apt-get install lxc` (Ubuntu) or `brew install lxc` (Mac) |
| Workspace won't start | `./ops/troubleshoot.sh` |

## Learn More

- [API Reference](docs/API.md)
- [Setup Details](docs/SETUP.md)
- [Workspace Creation](docs/WORKSPACES.md)
