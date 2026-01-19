# Nexus Testing Guide: epson-eshop Example

This guide walks you through testing Nexus with the IniZio/epson-eshop repository after pushing changes and creating a new release.

## Prerequisites

1. **Push Changes**:
   ```bash
   git push origin main
   ```

2. **Create GitHub Release**:
   ```bash
   # Create and push a tag
   git tag -a v0.8.0 -m "Release v0.8.0: Dynamic port exposure and CLI auth"
   git push origin v0.8.0
   
   # Or create release via GitHub UI:
   # https://github.com/IniZio/nexus/releases/new
   ```

3. **Install Latest Nexus** (on testing machine):
   ```bash
   # Download from release
   wget https://github.com/IniZio/nexus/releases/download/v0.8.0/nexus-linux-amd64
   chmod +x nexus-linux-amd64
   sudo mv nexus-linux-amd64 /usr/local/bin/nexus
   
   # Or build from source
   git clone https://github.com/IniZio/nexus
   cd nexus
   make build
   sudo cp bin/nexus /usr/local/bin/
   ```

## Part 1: Start Coordination Server

### 1.1 Setup GitHub App Authentication

First, ensure you have a GitHub App configured (or use the existing one):

```bash
# Set GitHub App credentials
export GITHUB_APP_ID="your-app-id"
export GITHUB_APP_PRIVATE_KEY_PATH="/path/to/private-key.pem"
export GITHUB_WEBHOOK_SECRET="your-webhook-secret"
```

### 1.2 Start the Coordination Server

```bash
# Start coordination server
nexus coordination start

# Or run in background
nohup nexus coordination start > /tmp/nexus-coordination.log 2>&1 &

# Verify it's running
curl http://localhost:3001/health
```

**Expected Output**:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "total_nodes": 0,
  "active_nodes": 0,
  "timestamp": "2026-01-19T..."
}
```

## Part 2: Setup GitHub Installation

### 2.1 Authenticate with GitHub

Option A: **Use OAuth Flow** (Recommended):
```bash
# Visit the GitHub OAuth URL
# The server will log the URL when it starts
# Example: https://github.com/login/oauth/authorize?client_id=...

# Complete the OAuth flow in your browser
# You'll be redirected back with installation completed
```

Option B: **Manual Database Entry** (For Testing):
```bash
# Create test installation
cd /path/to/nexus
cat > /tmp/add_github_install.go << 'EOF'
package main

import (
	"database/sql"
	"log"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", ".nexus-runtime/data/nexus.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userID := "user_test_IniZio"
	token := "YOUR_GITHUB_TOKEN"  // Replace with actual token
	expiresAt := time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339)
	now := time.Now().Format(time.RFC3339)

	_, err = db.Exec(`
		INSERT OR REPLACE INTO github_installations 
		(installation_id, user_id, github_user_id, github_username, repo_full_name, token, token_expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, 12345, userID, 1234, "IniZio", "IniZio/epson-eshop", token, expiresAt, now, now)
	
	if err != nil {
		log.Fatal(err)
	}
	
	log.Println("GitHub installation added")
}
EOF

go run /tmp/add_github_install.go
```

## Part 3: Create Workspace from epson-eshop

### 3.1 Create Workspace via API

```bash
# Create workspace
RESPONSE=$(curl -X POST http://localhost:3001/api/v1/workspaces/create-from-repo \
  -H "Content-Type: application/json" \
  -d '{
    "github_username": "IniZio",
    "workspace_name": "epson-eshop-test",
    "repo": {
      "owner": "IniZio",
      "name": "epson-eshop",
      "url": "https://github.com/IniZio/epson-eshop",
      "branch": "master"
    },
    "provider": "docker"
  }')

echo $RESPONSE | jq

# Extract workspace ID
WORKSPACE_ID=$(echo $RESPONSE | jq -r '.workspace_id')
echo "Workspace ID: $WORKSPACE_ID"
```

**Expected Output**:
```json
{
  "workspace_id": "ws-1768793663913416282",
  "status": "creating",
  "ssh_port": 2290,
  "polling_url": "/api/v1/workspaces/ws-1768793663913416282/status",
  "estimated_time_seconds": 60,
  "fork_created": false,
  "created_at": "2026-01-19T03:34:24.538Z"
}
```

### 3.2 Monitor Provisioning Progress

```bash
# Wait for provisioning (20-30 seconds)
sleep 25

# Check status
curl http://localhost:3001/api/v1/workspaces/$WORKSPACE_ID/status | jq
```

**Expected Output**:
```json
{
  "workspace_id": "ws-1768793663913416282",
  "owner": "user_test_IniZio",
  "name": "epson-eshop-test",
  "status": "running",
  "provider": "docker",
  "ssh": {
    "host": "localhost",
    "port": 32787,
    "user": "dev",
    "key_required": "~/.ssh/id_ed25519"
  },
  "services": {
    "postgres": {
      "name": "postgres",
      "status": "running",
      "port": 5432,
      "mapped_port": 32789,
      "health": "healthy",
      "url": "http://localhost:32789",
      "last_check": "2026-01-19T03:28:31Z"
    },
    "redis": {
      "name": "redis",
      "status": "running",
      "port": 6379,
      "mapped_port": 32790,
      "health": "healthy",
      "url": "http://localhost:32790"
    },
    "app": {
      "name": "app",
      "status": "running",
      "port": 5000,
      "mapped_port": 32788,
      "health": "healthy",
      "url": "http://localhost:32788"
    }
  },
  "repository": {
    "owner": "IniZio",
    "name": "epson-eshop",
    "url": "https://github.com/IniZio/epson-eshop",
    "branch": "master"
  },
  "created_at": "2026-01-19T03:34:24Z",
  "updated_at": "2026-01-19T03:34:50Z"
}
```

### 3.3 Save Port Information

```bash
# Extract ports for testing
SSH_PORT=$(curl -s http://localhost:3001/api/v1/workspaces/$WORKSPACE_ID/status | jq -r '.ssh.port')
POSTGRES_PORT=$(curl -s http://localhost:3001/api/v1/workspaces/$WORKSPACE_ID/status | jq -r '.services.postgres.mapped_port')
REDIS_PORT=$(curl -s http://localhost:3001/api/v1/workspaces/$WORKSPACE_ID/status | jq -r '.services.redis.mapped_port')
APP_PORT=$(curl -s http://localhost:3001/api/v1/workspaces/$WORKSPACE_ID/status | jq -r '.services.app.mapped_port')

echo "SSH Port: $SSH_PORT"
echo "Postgres Port: $POSTGRES_PORT"
echo "Redis Port: $REDIS_PORT"
echo "App Port: $APP_PORT"
```

## Part 4: Test Workspace Access

### 4.1 SSH into Workspace

```bash
# SSH into the workspace
ssh -p $SSH_PORT dev@localhost

# Once inside:
# Check repository was cloned
ls -la /workspace
cd /workspace
ls -la

# Check environment variables
cat /etc/environment | grep NEXUS_SERVICE

# Expected output:
# export NEXUS_SERVICE_POSTGRES_PORT=32789
# export NEXUS_SERVICE_REDIS_PORT=32790
# export NEXUS_SERVICE_APP_PORT=32788
```

### 4.2 Verify Repository Structure

```bash
# Inside the workspace container
cd /workspace

# Check epson-eshop structure
ls -la
# Expected: .nexus/, src/, package.json, docker-compose.yml, etc.

# Check .nexus/config.yaml
cat .nexus/config.yaml
```

**Expected config.yaml**:
```yaml
version: "1.0"
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

## Part 5: Test Services

### 5.1 Test PostgreSQL Access

**From Host Machine**:
```bash
# Test PostgreSQL connection from host
psql -h localhost -p $POSTGRES_PORT -U postgres -d postgres -c '\l'

# Or use Docker
docker exec $WORKSPACE_ID psql -U postgres -d postgres -c '\l'
```

**From Inside Container**:
```bash
# SSH into workspace
ssh -p $SSH_PORT dev@localhost

# Access postgres internally
psql -h localhost -p 5432 -U postgres -d postgres -c '\l'
```

### 5.2 Test Redis Access

**From Host Machine**:
```bash
# Test Redis connection
redis-cli -h localhost -p $REDIS_PORT ping
# Expected: PONG

# Set and get a test key
redis-cli -h localhost -p $REDIS_PORT SET test "Hello from host"
redis-cli -h localhost -p $REDIS_PORT GET test
```

**From Inside Container**:
```bash
# Inside workspace
redis-cli -h localhost -p 6379 ping
redis-cli -h localhost -p 6379 SET test "Hello from container"
redis-cli -h localhost -p 6379 GET test
```

### 5.3 Start and Test Application

**Inside Container**:
```bash
# SSH into workspace
ssh -p $SSH_PORT dev@localhost

cd /workspace

# Install dependencies (if needed)
npm install

# Start services with docker-compose
docker-compose up -d postgres redis

# Wait for services to start
sleep 5

# Start the application
npm start &

# Wait for app to start
sleep 10

# Test from inside container
curl http://localhost:5000
```

**From Host Machine**:
```bash
# Test app endpoint from host
curl http://localhost:$APP_PORT

# Test specific routes (based on epson-eshop API)
curl http://localhost:$APP_PORT/api/health
curl http://localhost:$APP_PORT/api/products
```

## Part 6: Test Port Mapping Environment Variables

### 6.1 Verify Environment Variables

```bash
# SSH into workspace
ssh -p $SSH_PORT dev@localhost

# Source environment
source /etc/environment

# Check all NEXUS variables
env | grep NEXUS_SERVICE

# Test using environment variables in app
cat > /tmp/test_env.js << 'EOF'
const postgresPort = process.env.NEXUS_SERVICE_POSTGRES_PORT;
const redisPort = process.env.NEXUS_SERVICE_REDIS_PORT;
const appPort = process.env.NEXUS_SERVICE_APP_PORT;

console.log(`Postgres: localhost:${postgresPort}`);
console.log(`Redis: localhost:${redisPort}`);
console.log(`App exposed on: localhost:${appPort}`);
EOF

node /tmp/test_env.js
```

### 6.2 Test Connection from Application

```bash
# Inside container, create test script
cat > /tmp/test_connections.js << 'EOF'
const { Client } = require('pg');
const redis = require('redis');

async function testConnections() {
  // Test Postgres
  const pgClient = new Client({
    host: 'localhost',
    port: 5432,
    user: 'postgres',
    database: 'postgres'
  });
  
  try {
    await pgClient.connect();
    const res = await pgClient.query('SELECT NOW()');
    console.log('✓ Postgres connected:', res.rows[0].now);
    await pgClient.end();
  } catch (err) {
    console.error('✗ Postgres error:', err.message);
  }
  
  // Test Redis
  const redisClient = redis.createClient({
    host: 'localhost',
    port: 6379
  });
  
  redisClient.on('error', (err) => console.error('✗ Redis error:', err));
  redisClient.on('ready', () => {
    console.log('✓ Redis connected');
    redisClient.quit();
  });
}

testConnections();
EOF

node /tmp/test_connections.js
```

## Part 7: Test Complete Workflow

### 7.1 Full Application Test

```bash
# SSH into workspace
ssh -p $SSH_PORT dev@localhost
cd /workspace

# 1. Start services
docker-compose up -d postgres redis

# 2. Run database migrations (if applicable)
npm run migrate

# 3. Start application
npm start &

# Wait for app to fully start
sleep 15

# 4. Test application endpoints
curl http://localhost:5000/api/health
curl http://localhost:5000/api/products
curl -X POST http://localhost:5000/api/products \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Product", "price": 99.99}'

# 5. Verify data persisted in Postgres
psql -h localhost -p 5432 -U postgres -d epson_eshop -c 'SELECT * FROM products;'

# 6. Verify cache in Redis
redis-cli -h localhost -p 6379 KEYS "*"
```

### 7.2 Test from External Machine

If you want to test from another machine on the network:

```bash
# On another machine on the same network
# Find the host machine's IP
HOST_IP="192.168.1.100"  # Replace with actual IP

# Test Postgres
psql -h $HOST_IP -p 32789 -U postgres -d postgres -c '\l'

# Test Redis
redis-cli -h $HOST_IP -p 32790 ping

# Test App
curl http://$HOST_IP:32788/api/health
```

## Part 8: Test User-Scoped Filtering

### 8.1 Create Multiple Users

```bash
# Create workspaces for different users
curl -X POST http://localhost:3001/api/v1/workspaces/create-from-repo \
  -H "Content-Type: application/json" \
  -d '{
    "github_username": "user1",
    "workspace_name": "user1-workspace",
    "repo": {...},
    "provider": "docker"
  }'

curl -X POST http://localhost:3001/api/v1/workspaces/create-from-repo \
  -H "Content-Type: application/json" \
  -d '{
    "github_username": "user2",
    "workspace_name": "user2-workspace",
    "repo": {...},
    "provider": "docker"
  }'
```

### 8.2 Test Filtering

```bash
# List all workspaces
curl http://localhost:3001/api/v1/workspaces | jq '.total'

# List only user1's workspaces
curl "http://localhost:3001/api/v1/workspaces?user=user_test_user1" | jq

# List only user2's workspaces
curl "http://localhost:3001/api/v1/workspaces?user=user_test_user2" | jq
```

## Part 9: Cleanup

### 9.1 Stop and Remove Workspace

```bash
# Stop workspace (via API)
curl -X POST http://localhost:3001/api/v1/workspaces/$WORKSPACE_ID/stop

# Or directly with Docker
docker stop $WORKSPACE_ID
docker rm $WORKSPACE_ID

# Clean up workspace directory
rm -rf /tmp/nexus-workspaces/$WORKSPACE_ID
```

### 9.2 Stop Coordination Server

```bash
# If running in foreground, Ctrl+C

# If running in background
pkill -f "nexus coordination start"
```

## Troubleshooting

### Issue: Workspace Creation Fails

**Check**:
```bash
# View coordination server logs
tail -f /tmp/nexus-coordination.log

# Check Docker status
docker ps -a | grep ws-

# Check workspace registry
curl http://localhost:3001/api/v1/workspaces | jq
```

### Issue: Cannot Connect to Services

**Check**:
```bash
# Verify container is running
docker ps | grep $WORKSPACE_ID

# Check port mappings
docker port $WORKSPACE_ID

# Test port from host
nc -zv localhost $POSTGRES_PORT
nc -zv localhost $REDIS_PORT
nc -zv localhost $APP_PORT
```

### Issue: Environment Variables Not Set

**Check**:
```bash
# SSH into container
ssh -p $SSH_PORT dev@localhost

# Check /etc/environment
cat /etc/environment | grep NEXUS

# Source environment manually
source /etc/environment
env | grep NEXUS_SERVICE
```

### Issue: GitHub Authentication Fails

**Check**:
```bash
# Verify GitHub installation in database
sqlite3 .nexus-runtime/data/nexus.db "SELECT * FROM github_installations;"

# Check token validity
curl -H "Authorization: token YOUR_GITHUB_TOKEN" https://api.github.com/user

# Verify coordination server has GitHub App config
curl http://localhost:3001/health | jq
```

## Expected Results Summary

✅ **Workspace Created**: Status shows "running"
✅ **Repository Cloned**: /workspace contains epson-eshop code
✅ **Ports Exposed**: All services have unique host ports
✅ **Environment Variables Set**: NEXUS_SERVICE_*_PORT variables present
✅ **SSH Access**: Can connect via ssh -p $SSH_PORT dev@localhost
✅ **Service Access**: Can connect to postgres, redis from host machine
✅ **Application Running**: App accessible on mapped port
✅ **User Filtering**: Can filter workspaces by user_id

## Next Steps

1. **Test with Your Own Repository**: Replace epson-eshop with your repo
2. **Production Deployment**: Follow deployment guide in `deploy/`
3. **CI/CD Integration**: Use documented API for automation
4. **Custom Configurations**: Modify `.nexus/config.yaml` for your services

## Support

- Documentation: `docs/PORT_MANAGEMENT.md`, `docs/CLI_AUTHENTICATION.md`
- Issues: https://github.com/IniZio/nexus/issues
- API Reference: `docs/API_REFERENCE.md`
