# AGT-02: File-Based Agent Config Overrides

**Priority**: ⚡ Med
**Status**: [Completed]

## 🎯 Objective
Replace complex YAML variable selection with simple file-based agent configuration overrides and suppression mechanism.

## 🛠 Implementation Details

### **Override System Architecture**
```
.vendetta/
├── templates/           # Base templates (built-in + remote)
│   ├── rules/
│   │   └── typescript.md
│   ├── skills/
│   │   └── debug.yaml
│   └── commands/
│       └── commit.md
├── agents/              # Project-level overrides
│   └── cursor/
│       ├── rules/
│       │   ├── typescript.md     # Override base template
│       │   └── legacy.md         # Empty file = suppress
│       ├── skills/
│       │   └── custom.yaml       # Override base skill
│       └── commands/
│           └── pr-create.md      # Override base command
└── worktrees/<name>/    # Generated configs
    └── .cursor/
        ├── mcp.json
        ├── rules/
        │   ├── typescript.md     # From override
        │   └── custom.md         # From base (legacy suppressed)
        └── commands/
            └── pr-create.md      # From override
```

### **Merging Logic**
1. **Base Layer**: Load built-in templates + remote templates
2. **Override Layer**: Apply project-specific files from `.vendetta/agents/{agent}/`
3. **Suppression**: Empty files in override layer prevent generation
4. **Generation**: Create final configs in worktree during `workspace create`

### **File Resolution Priority**
```go
// Pseudocode for template resolution
func resolveTemplate(agent, templatePath string) (content string, shouldGenerate bool) {
// Check for override (highest priority)
overridePath := filepath.Join(".vendetta", "agents", agent, templatePath)
if fileExists(overridePath) {
    if isEmptyFile(overridePath) {
        return "", false  // Suppress generation
    }
    return readFile(overridePath), true  // Use override
}
        return readFile(overridePath), true  // Use override
    }

    // Check remote templates
    remotePath := filepath.Join(".vendetta", "remotes", remoteName, templatePath)
    if fileExists(remotePath) {
        return readFile(remotePath), true
    }

    // Fall back to base template
    basePath := filepath.Join(".vendetta", "templates", templatePath)
    if fileExists(basePath) {
        return readFile(basePath), true
    }

    return "", false  // No template found
}
```

### **Agent-Specific Generation & Formats**

**Cursor** (Sources: https://mcpconfig.com/cursor-mcp/, https://docs.cursor.com/context/rules, https://cursor.com/docs/agent/chat/commands):
- **MCP Config**: `.cursor/mcp.json` with `mcpServers` object
- **Rules**: `.cursor/rules/*.mdc` files (YAML frontmatter + Markdown)
- **Commands**: `.cursor/commands/*.md` files (plain Markdown workflows)
- **Legacy Support**: `.cursorrules` (deprecated but functional)

**OpenCode** (Sources: https://opencode.ai/docs/config/, https://opencode.ai/docs/mcp-servers/, https://opencode.ai/docs/skills/):
- **MCP Config**: `mcp` section in `opencode.json` with server definitions
- **Rules**: `AGENTS.md` in project root
- **Skills**: `.opencode/skill/*/SKILL.md` with YAML frontmatter (agentskills.io format)

**Claude**: Generates respective config files in project root (format TBD - likely agentskills.io compatible)

## 🧪 Testing Requirements

### **Unit Tests**
- ✅ Template resolution priority (override > remote > base) for rules, skills, commands
- ✅ Suppression via empty files works correctly for all template types
- ✅ File copying and content replacement with format preservation
- ✅ Agent-specific output locations and formats (Cursor: .mdc/.md, OpenCode: SKILL.md)
- ✅ Error handling for missing templates and invalid formats

### **Integration Tests**
- ✅ Full config generation during workspace creation (MCP + rules + skills + commands)
- ✅ Override files take precedence over base templates for all types
- ✅ Suppressed templates are not generated across all categories
- ✅ Generated configs are functional (Cursor MCP connection, command workflows, etc.)

### **E2E Scenarios**
```bash
# Test override mechanism
vendetta workspace create override-test

# Create custom rule override using modern .mdc format
mkdir -p .vendetta/agents/cursor/rules
cat > .vendetta/agents/cursor/rules/typescript.mdc << EOF
---
description: Custom TypeScript coding standards
globs:
  - "src/**/*.ts"
  - "src/**/*.tsx"
alwaysApply: false
---
# Custom TypeScript Rules
- Use 4 spaces instead of 2
- Prefer const over let
- Use interfaces over types for object shapes
EOF

# Create custom command override
mkdir -p .vendetta/agents/cursor/commands
cat > .vendetta/agents/cursor/commands/pr-create.md << EOF
# Create Pull Request

## Overview
Standardized PR creation workflow for our team.

## Steps
1. Ensure tests pass
2. Update CHANGELOG.md
3. Create PR with proper template
4. Request reviews from appropriate team members
EOF

# Create suppression
touch .vendetta/agents/cursor/rules/legacy.md  # Empty = suppress

# Generate configs
vendetta workspace create override-test

# Verify results
ls .vendetta/worktrees/override-test/.cursor/rules/
# Should contain: typescript.mdc (custom) but NOT legacy.md

ls .vendetta/worktrees/override-test/.cursor/commands/
# Should contain: pr-create.md (custom override)

cat .vendetta/worktrees/override-test/.cursor/rules/typescript.mdc
# Should contain YAML frontmatter + custom content
```

## 📋 Implementation Steps

1. **Template Resolution**: Implement priority-based template resolution
2. **Suppression Logic**: Add empty file detection and suppression
3. **File Operations**: Implement safe file copying and content handling
4. **Agent Integration**: Wire into workspace creation process
5. **Config Generation**: Update GenerateAgentConfigs() to use new system

## 🎯 Success Criteria
- ✅ Override files replace base templates correctly
- ✅ Empty files suppress template generation
- ✅ Agent configs generate in correct locations
- ✅ Generated configs are functional for respective agents
- ✅ Backward compatibility with existing configs

## 📚 Dependencies
- CLI-03: Workspace Command Group (sequential - needs workspace creation)</content>
<parameter name="filePath">docs/planning/tasks/AGT-02.md
