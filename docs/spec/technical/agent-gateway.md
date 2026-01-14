# Technical Specification: Agent Configuration Generation

## 1. Overview
The Agent Configuration system generates appropriate configuration files for AI agents (Cursor, OpenCode, Claude) to work seamlessly with isolated development environments. Configurations are generated from templates with support for project-specific overrides and plugin-based extensions.

## 2. Configuration Generation Process

## 4. Agent Configuration Generation System

### **Template Architecture**
```
.vendetta/
├── config.yaml                 # Main configuration
├── templates/                  # Shared templates (open standards)
│   ├── skills/                 # agentskills.io compliant
│   ├── commands/               # Standardized command definitions
│   └── rules/                  # agents.md compliant
├── agents/                     # Agent-specific file overrides
│   ├── cursor/
│   │   ├── rules/
│   │   │   ├── typescript.md    # Override specific rule
│   │   │   └── legacy-code.md   # Empty file = suppress this rule
│   │   └── skills/
│   │       └── debug.yaml       # Override specific skill
│   ├── opencode/
│   └── claude-desktop/
└── worktrees/                  # Generated worktrees (gitignored)
    └── <branch>/
        ├── .cursor/mcp.json     # Generated final configs
        ├── opencode.json
        └── ...
```

### **Generation Process**
1. **Base Templates**: Start with built-in defaults from `.vendetta/templates/`
2. **Remote Templates**: Merge with templates from `vendetta config pull` sources
3. **Project Overrides**: Apply file-level overrides from `.vendetta/agents/{agent}/`
4. **Suppression Check**: Skip generation for rules/skills with empty override files
5. **Generate Configs**: Create final agent configurations in worktree directories

### **Override Mechanism**
- **Override**: Place a file in `.vendetta/agents/{agent}/rules/` or `skills/` to replace the base template
- **Suppression**: Create an empty file with the same name to prevent that rule/skill from being generated
- **Example**: Empty `.vendetta/agents/cursor/rules/legacy-code.md` prevents legacy-code rule generation

### **File Resolution Priority**
1. **Project Override**: `.vendetta/agents/cursor/rules/custom.md` (highest priority)
2. **Remote Template**: From `vendetta config pull` sources
3. **Base Template**: Built-in defaults (lowest priority)

### **Supported Agents**

#### **Cursor**
- **Configuration**: Agent-specific rules and settings
- **Output**: `.cursor/` directory with rules and configurations
- **Format**: Markdown rules files (.mdc) and configuration files

#### **OpenCode**
- **Configuration**: Project-specific settings and capabilities
- **Output**: `opencode.json` and `.opencode/` directory
- **Features**: Custom rules, skills, and command definitions

#### **Claude Desktop/Code**
- **Configuration**: MCP-compatible configuration files
- **Output**: `claude_desktop_config.json` or `claude_code_config.json`
- **Format**: JSON configuration for Claude agent integration

## 5. Shared Templates (Open Standards)

### **Skills** (agentskills.io)
Standardized YAML format with metadata, parameters, execution, permissions.

### **Rules** (agents.md)
Markdown with frontmatter for applicability, priority, and content.

### **Commands**
YAML with steps, environment variables, and metadata.

## 6. Configuration Example

```yaml
# .vendetta/config.yaml
name: my-project
agents:
  - name: opencode
    enabled: true
  - name: cursor
    enabled: true
```

Generated configurations provide each agent with appropriate settings and capabilities for the development environment.

## 7. Implementation Status
- ✅ Template merging from multiple sources (base, remotes, agents)
- ✅ Agent config file generation during workspace creation
- ✅ Support for Cursor, OpenCode, Claude Desktop/Code
- ✅ Plugin-based rule and skill extensions
- ✅ Override and suppression mechanisms
