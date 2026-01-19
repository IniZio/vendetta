# Port Management in Nexus

## Overview

Nexus automatically exposes workspace service ports to the host machine with dynamic port allocation. This eliminates port conflicts and enables external access to services running in workspaces.

## How It Works

When a workspace is created:

1. **Service Discovery**: Nexus reads `.nexus/config.yaml` from your repository
2. **Port Exposure**: All services with defined ports are exposed to the host
3. **Dynamic Allocation**: Docker auto-assigns available host ports (typically 32768-65535)
4. **Environment Injection**: Port mappings are injected as environment variables
5. **Registry Storage**: Mapped ports are stored and returned in API responses

## Port Mapping Example

```yaml
# .nexus/config.yaml
services:
  postgres:
    command: "docker-compose up postgres"
    port: 5432
  redis:
    command: "docker-compose up redis"
    port: 6379
  app:
    command: "npm start"
    port: 5000
```

**Result after workspace creation**:

| Service | Container Port | Host Port | Environment Variable |
|---------|---------------|-----------|----------------------|
| postgres | 5432 | 32789 | `NEXUS_SERVICE_POSTGRES_PORT=32789` |
| redis | 6379 | 32790 | `NEXUS_SERVICE_REDIS_PORT=32790` |
| app | 5000 | 32788 | `NEXUS_SERVICE_APP_PORT=32788` |

## Accessing Services

### 1. Direct Host Access

Services are accessible from your host machine using the mapped ports:

```bash
# Get workspace status with port mappings
curl http://localhost:3001/api/v1/workspaces/ws-123/status | jq '.services'

# Access postgres directly
psql -h localhost -p 32789 -U postgres

# Access redis
redis-cli -h localhost -p 32790

# Access web app
curl http://localhost:32788
```

### 2. Within the Container

Inside the workspace container, services use their internal ports:

```bash
# SSH into workspace
ssh -p 32787 dev@localhost

# Access postgres internally
psql -h localhost -p 5432 -U postgres

# Access redis internally
redis-cli -h localhost -p 6379
```

### 3. Using Environment Variables

Applications can read the exposed host ports programmatically:

```bash
# Inside container - read environment variables
source /etc/environment
echo $NEXUS_SERVICE_APP_PORT  # 32788
echo $NEXUS_SERVICE_POSTGRES_PORT  # 32789
```

## API Response Format

The workspace status endpoint returns complete port information:

```json
{
  "workspace_id": "ws-1768793388045888119",
  "status": "running",
  "ssh": {
    "host": "localhost",
    "port": 32787
  },
  "services": {
    "app": {
      "name": "app",
      "status": "running",
      "port": 5000,           // Container port
      "mapped_port": 32788,   // Host port (use this from outside)
      "url": "http://localhost:32788",
      "health": "healthy"
    },
    "postgres": {
      "name": "postgres",
      "port": 5432,
      "mapped_port": 32789,
      "url": "http://localhost:32789"
    }
  }
}
```

## Benefits

### 1. No Port Conflicts
Each workspace gets unique ports automatically assigned by Docker. Run unlimited workspaces concurrently without manual port management.

### 2. External Access
Services are accessible from:
- Your host machine
- Other machines on the network (if host firewall allows)
- CI/CD pipelines
- Testing tools

### 3. Service Discovery
Applications can discover their exposed ports via environment variables, making configuration dynamic and portable.

### 4. Simple Integration
No need for SSH tunneling or port forwarding. Services are directly accessible using standard tools.

## Configuration

Service ports are defined in `.nexus/config.yaml`:

```yaml
version: "1.0"
provider:
  type: docker
  docker:
    image: ubuntu:22.04

services:
  # Basic service with port
  api:
    command: "./start-api.sh"
    port: 8080
  
  # Service without port (worker processes)
  worker:
    command: "./start-worker.sh"
    # No port defined - won't be exposed
  
  # Database service
  database:
    command: "docker-compose up db"
    port: 5432
```

**Rules**:
- Services with `port: 0` or no port defined are NOT exposed to the host
- Only TCP ports are supported
- Ports must be > 0 and < 65536
- Container services must actually listen on the specified ports

## Troubleshooting

### Service Not Accessible

**Problem**: Cannot connect to service on mapped port

**Solutions**:
1. Check service is actually running inside container:
   ```bash
   ssh -p <ssh_port> dev@localhost
   docker ps  # if using docker-compose
   netstat -tlnp | grep <internal_port>
   ```

2. Verify port mapping is correct:
   ```bash
   curl http://localhost:3001/api/v1/workspaces/<workspace_id>/status | jq '.services'
   ```

3. Check Docker port bindings:
   ```bash
   docker port <workspace_id>
   ```

4. Test from inside container first:
   ```bash
   ssh -p <ssh_port> dev@localhost
   curl http://localhost:<internal_port>
   ```

### Port Already in Use

This shouldn't happen with dynamic allocation, but if it does:

```bash
# Find process using the port
lsof -i :<port>

# Or
netstat -tlnp | grep <port>
```

### Environment Variables Not Set

**Problem**: `$NEXUS_SERVICE_*_PORT` variables are empty

**Solutions**:
1. Source the environment file:
   ```bash
   source /etc/environment
   ```

2. Check if variables were injected:
   ```bash
   cat /etc/environment | grep NEXUS_SERVICE
   ```

3. Restart your shell or application to pick up new environment

## Advanced Usage

### Multiple Workspaces

Each workspace gets independent port mappings:

```bash
# Workspace 1
ws-123: app -> 32788, postgres -> 32789

# Workspace 2
ws-456: app -> 32791, postgres -> 32792
```

No conflicts, fully isolated.

### Network Access from External Machine

If you want to access services from another machine on your network:

1. Find your host machine's IP:
   ```bash
   ip addr show | grep "inet "
   ```

2. Ensure firewall allows the port:
   ```bash
   sudo ufw allow 32789/tcp  # Example for postgres
   ```

3. Connect from remote machine:
   ```bash
   psql -h 192.168.1.100 -p 32789 -U postgres
   ```

### Using with CI/CD

In CI/CD pipelines, query the workspace status API to get port mappings:

```bash
#!/bin/bash
WORKSPACE_ID="ws-123"
API_URL="http://nexus-server:3001"

# Get postgres port
POSTGRES_PORT=$(curl -s $API_URL/api/v1/workspaces/$WORKSPACE_ID/status | \
  jq -r '.services.postgres.mapped_port')

# Run tests
DATABASE_URL="postgresql://localhost:$POSTGRES_PORT/testdb" npm test
```

## See Also

- [SSH Tunneling Guide](./SSH_TUNNELING.md) - Alternative access method
- [Workspace Configuration](./CONFIGURATION.md) - Complete config reference
- [Service Health Checks](./HEALTH_CHECKS.md) - Monitoring service status
