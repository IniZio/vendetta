# CFG-01: Config Pull/Sync Commands

**Priority**: 🟢 Low
**Status**: [In Progress]

## 🎯 Objective
Implement config pull and sync commands for sharing templates and configurations across teams and projects.

## 🛠 Implementation Details

### **Config Pull Command**
```bash
vendetta config pull <url> [--branch=branch]
```

**Functionality:**
- Clone remote Git repository into `.vendetta/remotes/`
- Extract templates from repository
- Merge with existing templates (remote overrides base)
- Support branch specification for template versioning

### **Config Sync Commands**
```bash
vendetta config sync <target>      # Sync to specific target
vendetta config sync-all          # Sync to all configured targets
```

**Functionality:**
- Read sync targets from `.vendetta/config.yaml`
- Create filtered Git branch with only `.vendetta` directory
- Push to remote repository
- Clean up temporary branches

### **Configuration Schema**
```yaml
# .vendetta/config.yaml
sync_targets:
  - name: "team-templates"
    url: "https://github.com/company/vendetta-templates.git"
  - name: "project-configs"
    url: "https://github.com/project/vendetta-configs.git"
```

### **Template Merging**
- **Remote Ref Tracking**: Store latest commit SHA for each remote in `.vendetta/state.json`
- **Fast-Forward Merging**: When remote has linear history from stored ref, prefer remote templates
- **Conflict Resolution**: Use chezmoi-style interactive reconciliation for merge conflicts
- **Precedence Order**: remote (fast-forward) > local modifications > base templates

## 🧪 Testing Requirements

### **Unit Tests**
- ✅ Git clone operations with branch specification
- ✅ Template extraction and merging logic
- ✅ Sync target configuration parsing
- ✅ Filtered branch creation and cleanup

### **Integration Tests**
- ✅ Full config pull workflow
- ✅ Template merging from multiple remotes
- ✅ Sync operations to configured targets
- ✅ Git authentication and error handling

### **E2E Scenarios**
```bash
# Test config pulling
vendetta config pull https://github.com/company/templates.git --branch=main

# Verify templates merged
ls .vendetta/templates/
# Should contain both base and remote templates

# Test config syncing
cat > .vendetta/config.yaml << EOF
sync_targets:
  - name: test-sync
    url: https://github.com/test/repo.git
EOF

# Create some config
echo "test config" > .vendetta/test.txt

vendetta config sync test-sync
# Verify: .vendetta directory pushed to remote repo
```

## 📋 Implementation Steps

1. **Config Pull**: Implement Git cloning and template extraction
2. **Template Merging**: Add logic to merge remote templates with base
3. **Config Sync**: Implement filtered Git operations for sync targets
4. **Configuration**: Update config schema to support sync targets
5. **Error Handling**: Add robust Git operation error handling

## 🎯 Success Criteria
- ✅ Remote repositories can be pulled successfully
- ✅ Templates merge correctly with proper priority
- ✅ Config sync works with filtered Git operations
- ✅ Multiple sync targets supported
- ✅ Git authentication and errors handled gracefully

## 📚 Dependencies
- None - This is independent functionality for config management</content>
<parameter name="filePath">docs/planning/tasks/CFG-01.md
