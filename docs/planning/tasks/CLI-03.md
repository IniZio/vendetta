# CLI-03: Workspace Command Group Implementation

**Priority**: 🔥 High
**Status**: [Completed]

## 🎯 Objective
Implement the complete workspace command group with context awareness, proper isolation, and intuitive user experience.

## 🛠 Implementation Details

### **Command Structure**
```bash
vendetta workspace create <name>     # Create workspace + configs
vendetta workspace up [name] [-d]    # Start services (detached optional)
vendetta workspace shell [name]      # Interactive shell
vendetta workspace stop [name]       # Stop services
vendetta workspace down [name]       # Stop + cleanup
vendetta workspace list              # List all workspaces
vendetta workspace rm <name>         # Remove workspace
# Note: 'agent' command removed per latest spec
```

### **Context Awareness**
- If inside `.vendetta/worktrees/<name>/`, auto-detect workspace name
- Commands accept optional `[name]` parameter for explicit specification
- Validate workspace existence before operations

### **Workspace Isolation**
- **Worktree Creation**: `git worktree add .vendetta/worktrees/<name>`
- **Branch Handling**: Auto-create branch if doesn't exist; for conflicts: stash changes, create/checkout branch, stash pop
- **Container Naming**: `vendetta-workspace-<name>` for easy identification

### **Agent Config Generation**
- Generate configs during `workspace create`
- Place configs in worktree root (`.cursor/mcp.json`, `opencode.json`, etc.)
- Support file-based overrides from `.vendetta/agents/`

### **Service Discovery Integration**
- Collect port mappings during container startup
- Inject `vendetta_SERVICE_*_URL` environment variables
- Make variables available in hooks and container shell

## 🧪 Testing Requirements

### **Unit Tests**
- ✅ Command parsing and validation
- ✅ Context detection logic
- ✅ Error handling for invalid workspaces
- ✅ Git worktree operations
- ✅ Docker container naming

### **Integration Tests**
- ✅ Full workspace lifecycle
- ✅ Agent config generation
- ✅ Hook execution with environment variables
- ✅ Port forwarding and service discovery

### **E2E Scenarios**
```bash
# Happy path
vendetta workspace create feature-x
vendetta workspace up
# Verify: Worktree created, container running, configs generated

# Context awareness
cd .vendetta/worktrees/feature-x
vendetta workspace up  # Should work without specifying name

# Error handling
vendetta workspace up nonexistent  # Should fail gracefully
```

## 📋 Implementation Steps

1. **CLI Structure**: Add workspace command group with subcommands
2. **Context Detection**: Implement workspace auto-detection logic
3. **Worktree Management**: Integrate with existing worktree manager
4. **Container Operations**: Extend Docker provider for workspace lifecycle
5. **Agent Integration**: Generate configs during workspace creation
6. **Service Discovery**: Fix and integrate environment variable injection

## 🎯 Success Criteria
- ✅ All workspace commands work correctly
- ✅ Context awareness functions properly
- ✅ Proper isolation between workspaces
- ✅ Agent configs generated and functional
- ✅ Service discovery variables injected correctly

## 📚 Dependencies
- COR-03: Service Discovery Environment Variables (parallel)
- AGT-02: File-Based Agent Config Overrides (sequential)</content>
<parameter name="filePath">docs/planning/tasks/CLI-03.md
