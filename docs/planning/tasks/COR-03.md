# COR-03: Service Discovery Environment Variables Fix

**Priority**: 🔥 High
**Status**: [Completed]

## 🎯 Objective
Fix the critical bug where service discovery environment variables are not injected into running containers, breaking the core service discovery feature.

## 🚨 Root Cause
In `pkg/provider/docker/docker.go`, the Docker provider collects service URL environment variables but fails to pass them to the container configuration. The `env` slice is populated correctly but never used in `ContainerCreate()`.

## 🛠 Implementation Details

### **Current Broken Code** (lines ~104-113 in docker.go)
```go
env := []string{}
for name, svc := range cfg.Services {
    if svc.Port > 0 {
        pStr := fmt.Sprintf("%d/tcp", svc.Port)
        if bindings, ok := json.NetworkSettings.Ports[nat.Port(pStr)]; ok && len(bindings) > 0 {
            url := fmt.Sprintf("http://localhost:%s", bindings[0].HostPort)
            env = append(env, fmt.Sprintf("vendetta_SERVICE_%s_URL=%s", name, url))
        }
    }
}
// ❌ MISSING: env is never used in ContainerCreate
```

### **Fix Implementation**
```go
resp, err := p.cli.ContainerCreate(ctx, &container.Config{
    Image: imgName,
    Tty:   true,
    Labels: map[string]string{
        "vendetta.session.id": sessionID,
    },
    Cmd:          []string{"/bin/bash"},
    Env:          env,  // ✅ ADD THIS: Pass environment variables
    ExposedPorts: exposedPorts,
}, &container.HostConfig{
    Mounts:       mounts,
    PortBindings: portBindings,
    Privileged:   cfg.Docker.DinD,
}, nil, nil, sessionID)
```

### **Port Auto-Detection**
- **Command Analysis**: Parse service commands to detect exposed ports
- **Docker Compose**: Extract ports from `docker-compose.yml` or command args
- **Dev Servers**: Detect common patterns (`npm run dev`, `yarn start`, etc.)
- **Protocol Guessing**: postgres → `postgresql://`, web services → `http://`

### **Environment Variable Format**
- **Pattern**: `vendetta_SERVICE_{SERVICE_NAME}_URL={PROTOCOL}://localhost:{PORT}`
- **Examples**:
  - `vendetta_SERVICE_WEB_URL=http://localhost:3000`
  - `vendetta_SERVICE_API_URL=http://localhost:8080`
  - `vendetta_SERVICE_DB_URL=postgresql://localhost:5432`
- **Service Name**: Uppercased service key from config.yaml

## 🧪 Testing Requirements

### **Unit Tests**
- ✅ Environment variable generation follows correct pattern
- ✅ Service names are uppercased properly
- ✅ Port mappings are resolved to localhost URLs
- ✅ Multiple services generate multiple variables

### **Integration Tests**
- ✅ Container receives environment variables on creation
- ✅ Variables accessible in container shell: `env | grep vendetta_SERVICE`
- ✅ Variables available in hook scripts during execution

### **E2E Scenarios**
```bash
# Test service discovery
vendetta workspace create discovery-test

# Configure services with commands (ports auto-detected)
cat > .vendetta/config.yaml << EOF
services:
  web:
    command: "cd client && npm run dev"
  api:
    command: "cd server && npm run dev"
  db:
    command: "docker-compose up -d postgres"
EOF

# Start workspace - environment variables injected before services start
vendetta workspace up discovery-test

# Verify environment variables available in container
vendetta workspace shell discovery-test
env | grep vendetta_SERVICE
# Expected output:
# vendetta_SERVICE_WEB_URL=http://localhost:3000
# vendetta_SERVICE_API_URL=http://localhost:8080
# vendetta_SERVICE_DB_URL=postgresql://localhost:5432
```

## 📋 Implementation Steps

1. **Locate Bug**: Find the environment variable collection code in docker.go
2. **Apply Fix**: Add `Env: env` to the ContainerCreate call
3. **Test Fix**: Run existing E2E tests to verify functionality
4. **Add Tests**: Create comprehensive tests for service discovery

## 🎯 Success Criteria
- ✅ `vendetta_SERVICE_*_URL` variables available in running containers
- ✅ Variables accessible in hook scripts
- ✅ Multiple services work correctly
- ✅ Existing E2E tests pass
- ✅ New integration tests added

## 📚 Dependencies
- None - This is an isolated fix to the Docker provider</content>
<parameter name="filePath">docs/planning/tasks/COR-03.md
