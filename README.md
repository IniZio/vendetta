# nexus

Remote SSH access to isolated dev environments. Workspaces on Docker/LXC/QEMU.

---

## Start Staging

```bash
cd deploy/envs/staging
./ops/start.sh
```

Server: http://localhost:3001

Full guide: [deploy/envs/staging/README.md](deploy/envs/staging/README.md)

---

## Dev: Connect to Workspace

1. Get SSH key: `cat ~/.ssh/id_ed25519.pub`
2. Register with admin
3. Get workspace ID from admin
4. SSH in: `ssh -p 2236 dev@localhost`
5. Inside: `cd /workspace` → Your code

---

## Admin: Deploy & Manage

[Staging Deployment Guide](deploy/envs/staging/README.md)

- Start server: `deploy/envs/staging/ops/start.sh`
- Register users: `deploy/envs/staging/ops/users.sh`
- Create workspaces: `deploy/envs/staging/ops/workspaces.sh`
- Troubleshoot: `deploy/envs/staging/ops/troubleshoot.sh`

---

## API Reference

[deploy/envs/staging/docs/API.md](deploy/envs/staging/docs/API.md)

---

## Setup from Scratch

[deploy/envs/staging/docs/SETUP.md](deploy/envs/staging/docs/SETUP.md)

---

## Build

```bash
go build -o bin/nexus ./cmd/nexus
```

Test:
```bash
go test ./...
```
