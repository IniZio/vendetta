# Staging API Reference

Base: `http://localhost:3001`

## Health

```
GET /health
```

Check server is up.

## Users

### Register User
```
POST /api/v1/users/register-github
Content-Type: application/json

{
  "github_username": "alice",
  "github_id": 123456,
  "ssh_pubkey": "ssh-ed25519 AAAA...",
  "ssh_pubkey_fingerprint": "SHA256:..."
}
```

Returns: `{user_id, github_username}`

Use script: `../ops/users.sh register alice 123456 "ssh-ed25519 AAAA..."`

## Workspaces

### Create
```
POST /api/v1/workspaces/create-from-repo
Content-Type: application/json

{
  "github_username": "alice",
  "workspace_name": "feature-x",
  "repo": {
    "owner": "oursky",
    "name": "epson-eshop",
    "url": "git@github.com:oursky/epson-eshop.git",
    "branch": "main",
    "is_fork": false
  },
  "provider": "lxc",
  "image": "ubuntu:22.04",
  "services": [
    {"name": "web", "command": "bundle exec puma -p 5000", "port": 5000}
  ]
}
```

Returns: `{workspace_id, name, status}`

Use script: `../ops/workspaces.sh create alice feature-x`

### List
```
GET /api/v1/workspaces
```

Returns: `{workspaces: [{id, name, status, provider, ...}]}`

Use script: `../ops/workspaces.sh list`

### Get Status
```
GET /api/v1/workspaces/{workspace-id}/status
```

Returns: `{id, name, status, ssh_port, services}`

Use script: `../ops/workspaces.sh status {id}`

### Stop
```
POST /api/v1/workspaces/{workspace-id}/stop
```

Use script: `../ops/workspaces.sh stop {id}`

### Delete
```
DELETE /api/v1/workspaces/{workspace-id}
```

Use script: `../ops/workspaces.sh delete {id}`

## Quick Examples

Register user:
```bash
curl -X POST http://localhost:3001/api/v1/users/register-github \
  -H "Content-Type: application/json" \
  -d '{"github_username":"alice","github_id":123456,"ssh_pubkey":"ssh-ed25519 AAAA...","ssh_pubkey_fingerprint":"SHA256:..."}'
```

Create workspace:
```bash
curl -X POST http://localhost:3001/api/v1/workspaces/create-from-repo \
  -H "Content-Type: application/json" \
  -d '{"github_username":"alice","workspace_name":"feature-x","repo":{"owner":"oursky","name":"epson-eshop","url":"git@github.com:oursky/epson-eshop.git","branch":"main","is_fork":false},"provider":"lxc","image":"ubuntu:22.04","services":[{"name":"web","command":"bundle exec puma -p 5000","port":5000}]}'
```

Get status:
```bash
curl http://localhost:3001/api/v1/workspaces/{id}/status | jq
```

List all:
```bash
curl http://localhost:3001/api/v1/workspaces | jq
```
