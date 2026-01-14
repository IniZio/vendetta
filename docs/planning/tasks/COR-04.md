# COR-04: Convention-Based Hook System

**Priority**: ⚡ Med
**Status**: [Completed]

## 🎯 Objective
Implement convention-based hooks as the main operations in `.vendetta/hooks/` directory, allowing users to customize workspace behavior with full environment variable access.

## 🛠 Implementation Details

### **Hook Directory Structure**
```
.vendetta/hooks/
├── create.sh     # Executed during workspace create
├── up.sh         # Executed during workspace up
├── stop.sh       # Executed during workspace stop
└── down.sh       # Executed during workspace down
```

### **Execution Logic**
- **Main Operations**: Hook scripts replace default behavior (up.sh runs instead of default service startup)
- **Optional**: Missing hooks fall back to default behavior
- **Environment**: Full access to service discovery variables and workspace context
- **Permissions**: Scripts made executable with user warning about git commits
- **Error Handling**: Hook failures logged with recovery suggestions (Heroku/Fly.io style)

### **Hook Execution Points**
```go
// During workspace create - runs instead of default create logic
if createHook := filepath.Join(".vendetta", "hooks", "create.sh"); fileExists(createHook) {
    executeHook(createHook, envVars)
} else {
    // Default create behavior
}

// During workspace up - runs instead of default service startup
if upHook := filepath.Join(".vendetta", "hooks", "up.sh"); fileExists(upHook) {
    executeHook(upHook, envVars)
} else {
    // Default up behavior (docker-compose, etc.)
}
```

### **Hook Environment**
Hooks receive all `vendetta_SERVICE_*_URL` environment variables plus:
- `WORKSPACE_NAME`: Name of the current workspace
- `WORKTREE_PATH`: Absolute path to the worktree
- `CONTAINER_ID`: Docker container ID (when available)
- `HOST_USER`: Host system username
- `HOST_CWD`: Host current working directory

## 🧪 Testing Requirements

### **Unit Tests**
- ✅ Hook discovery finds scripts in correct locations
- ✅ Missing hooks are handled gracefully
- ✅ Scripts are made executable before execution
- ✅ Environment variables are passed correctly
- ✅ Hook failures don't crash workspace operations

### **Integration Tests**
- ✅ Hook execution order and timing
- ✅ Environment variable availability in hooks
- ✅ Hook output captured in logs
- ✅ Multiple hooks execute in sequence

### **E2E Scenarios**
```bash
# Test hook execution
vendetta workspace create hooks-demo

# Create hook as main operation
cat > .vendetta/hooks/up.sh << 'EOF'
#!/bin/bash
echo "Custom startup for $WORKSPACE_NAME"
echo "Web URL: $vendetta_SERVICE_WEB_URL"
docker-compose up -d  # Custom logic
npm run dev &         # Custom dev server
EOF

chmod +x .vendetta/hooks/up.sh

# Execute workspace up - hook replaces default behavior
vendetta workspace up hooks-demo

# Verify hook executed as main operation
# Check docker-compose started via hook
# Check npm dev server started via hook
```

## 📋 Implementation Steps

1. **Hook Discovery**: Add functions to find and validate hook scripts
2. **Hook Execution**: Implement safe hook execution with proper error handling
3. **Environment Setup**: Pass service discovery and workspace variables
4. **Integration**: Wire hooks into workspace lifecycle (create/up/stop/down)
5. **Testing**: Add comprehensive tests for hook functionality

## 🎯 Success Criteria
- ✅ Hooks execute at correct lifecycle points
- ✅ Missing hooks don't cause errors
- ✅ Environment variables available in hooks
- ✅ Hook failures are logged but don't prevent operations
- ✅ Scripts are made executable automatically

## 📚 Dependencies
- CLI-03: Workspace Command Group (sequential - hooks need workspace lifecycle)
- COR-03: Service Discovery (parallel - hooks use service variables)</content>
<parameter name="filePath">docs/planning/tasks/COR-04.md
