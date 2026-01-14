# Coordination Server API Documentation

## Overview

The vendetta Coordination Server provides a centralized API for managing remote nodes and dispatching commands across distributed development environments.

## Features

- **Node Registry**: Track and manage remote nodes
- **Command Dispatcher**: Send commands to specific nodes
- **Service Discovery**: Discover services running across nodes
- **Real-time Updates**: WebSocket/Server-Sent Events for live monitoring
- **Authentication**: JWT-based auth with configurable security
- **Health Monitoring**: Comprehensive health checks and metrics

## Configuration

The coordination server is configured via YAML files and environment variables.

### Default Configuration File
```yaml
server:
  host: "0.0.0.0"
  port: 3001
  auth_token: "vendetta-coordination-token"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"

registry:
  provider: "memory"
  sync_interval: "30s"
  health_check_interval: "10s"
  node_timeout: "60s"
  max_retries: 3

websocket:
  enabled: true
  path: "/ws"
  origins: ["*"]
  ping_period: "30s"

auth:
  enabled: false
  jwt_secret: "vendetta-jwt-secret-key-minimum-16-chars"
  token_expiry: "24h"

logging:
  level: "info"
  format: "json"
  output: "stdout"
  max_size: 100
  max_backups: 3
  max_age: 28
```

### Environment Variables

- `VENDETTA_COORD_HOST`: Override server host
- `VENDETTA_COORD_PORT`: Override server port  
- `VENDETTA_JWT_SECRET`: Override JWT secret

## API Endpoints

### Node Management

#### Register Node
```http
POST /api/v1/nodes
Content-Type: application/json

{
  "id": "node-1",
  "name": "Development Node 1",
  "provider": "docker",
  "status": "active",
  "address": "localhost",
  "port": 8080,
  "labels": {
    "env": "dev",
    "region": "us-west"
  },
  "capabilities": {
    "docker": true,
    "kubernetes": false
  }
}
```

#### List Nodes
```http
GET /api/v1/nodes
```

Response:
```json
{
  "nodes": [...],
  "count": 5
}
```

#### Get Node
```http
GET /api/v1/nodes/{id}
```

#### Get Node Status
```http
GET /api/v1/nodes/{id}/status
```

Response:
```json
{
  "node_id": "node-1",
  "status": "active",
  "last_seen": "2026-01-13T18:17:00Z",
  "uptime": "2h30m15s"
}
```

#### Update Node
```http
PUT /api/v1/nodes/{id}
Content-Type: application/json

{
  "status": "inactive",
  "port": 8081
}
```

#### Unregister Node
```http
DELETE /api/v1/nodes/{id}
```

### Command Dispatch

#### Send Command
```http
POST /api/v1/nodes/{id}/commands
Content-Type: application/json

{
  "type": "exec",
  "action": "echo 'hello world'",
  "params": {
    "timeout": "30s"
  },
  "user": "developer"
}
```

Response:
```json
{
  "id": "cmd_123456_node-1",
  "node_id": "node-1",
  "command": {...},
  "status": "success",
  "output": "hello world",
  "duration": "100ms",
  "finished": "2026-01-13T18:17:00Z"
}
```

#### Report Command Result
```http
POST /api/v1/commands/{id}/result
Content-Type: application/json

{
  "id": "cmd_123456_node-1",
  "node_id": "node-1",
  "status": "success",
  "output": "Command executed successfully",
  "duration": "150ms",
  "finished": "2026-01-13T18:17:00Z"
}
```

### Service Discovery

#### List Services
```http
GET /api/v1/services
```

Response:
```json
{
  "services": {
    "node-1": [
      {
        "id": "web-server",
        "name": "Web Server",
        "type": "http",
        "status": "running",
        "port": 8080,
        "endpoint": "http://localhost:8080"
      }
    ]
  },
  "nodes": 3
}
```

### Monitoring

#### Health Check
```http
GET /health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2026-01-13T18:17:00Z",
  "total_nodes": 5,
  "active_nodes": 4,
  "version": "1.0.0"
}
```

#### Metrics
```http
GET /metrics
```

Response:
```json
{
  "timestamp": "2026-01-13T18:17:00Z",
  "nodes": {
    "total": 5,
    "by_status": {
      "active": 4,
      "inactive": 1
    },
    "by_provider": {
      "docker": 3,
      "lxc": 2
    }
  },
  "services": {
    "total": 12
  },
  "websocket": {
    "connected_clients": 2
  }
}
```

## Real-time Updates

### WebSocket/SSE Connection

The server provides real-time updates via Server-Sent Events at `/ws`.

```javascript
const eventSource = new EventSource('http://localhost:3001/ws');

eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Event:', data.type, data.data);
};
```

#### Event Types

- `initial_state`: Current state when connecting
- `node_registered`: New node registered
- `node_updated`: Node information updated
- `node_unregistered`: Node removed
- `command_result`: Command execution result

## CLI Usage

### Generate Configuration
```bash
vendetta coordination config
```

### Start Server
```bash
vendetta coordination start
```

### Check Status
```bash
vendetta coordination status
```

## Integration Examples

### Registering a Docker Node
```bash
curl -X POST http://localhost:3001/api/v1/nodes \
  -H "Content-Type: application/json" \
  -d '{
    "id": "docker-node-1",
    "name": "Docker Development Node",
    "provider": "docker",
    "status": "active",
    "address": "localhost",
    "port": 2376,
    "labels": {"env": "development"},
    "capabilities": {"docker": true, "compose": true}
  }'
```

### Sending a Command
```bash
curl -X POST http://localhost:3001/api/v1/nodes/docker-node-1/commands \
  -H "Content-Type: application/json" \
  -d '{
    "type": "exec",
    "action": "docker ps",
    "params": {"timeout": "10s"}
  }'
```

### Monitoring with SSE
```bash
curl -N http://localhost:3001/ws
```

## Security Considerations

1. **Authentication**: Enable JWT authentication in production
2. **Network**: Limit origins in WebSocket configuration
3. **Firewall**: Restrict access to coordination server
4. **TLS**: Use reverse proxy for HTTPS termination

## Architecture

The coordination server follows a provider-agnostic design:

- **Registry Interface**: Abstract node storage
- **In-Memory Implementation**: Default for development
- **Extensible**: Can be extended for persistence backends
- **Transport Agnostic**: Works with SSH, HTTP, WebSocket

## Development

### Running Tests
```bash
go test ./pkg/coordination/...
```

### Building
```bash
go build -o bin/vendetta ./cmd/vendetta/
```

### Configuration Development
```go
cfg, err := coordination.LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}

server := coordination.NewServer(cfg)
if err := server.Start(); err != nil {
    log.Fatal(err)
}
```
