# Technical Specification: Agent Gateway & Config Generation

## 1. Overview
The Agent Gateway is a Model Context Protocol (MCP) server built into the `oursky` binary, providing a bridge between AI agents (Cursor, OpenCode, Claude) and isolated development environments. The system includes comprehensive agent configuration generation from templates.

## 2. MCP Protocol Details
- **Transport**: JSON-RPC over Standard Input/Output (STDIO)
- **Session Context**: Started per session: `./oursky agent [session-id]`
- **Server Location**: Configurable via `.vendatta/config.yaml` (default: localhost:3001)

## 3. Capabilities

### **Core Tool: `exec`**
Executes commands in the session's isolated environment.
- **Input**: `cmd` (string) - Command to execute
- **Execution**: Routed via `Provider.Exec` with proper isolation
- **Output**: Combined stdout/stderr with exit codes

### **Dynamic Tool Loading**
Automatically registers skills from shared templates as MCP tools based on agent configuration.

## 4. Agent Configuration Generation System

### **Template Architecture**
```
.vendatta/
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
1. **Base Templates**: Start with built-in defaults from `.vendatta/templates/`
2. **Remote Templates**: Merge with templates from `vendatta config pull` sources
3. **Project Overrides**: Apply file-level overrides from `.vendatta/agents/{agent}/`
4. **Suppression Check**: Skip generation for rules/skills with empty override files
5. **Generate Configs**: Create final agent configurations in worktree directories

### **Override Mechanism**
- **Override**: Place a file in `.vendatta/agents/{agent}/rules/` or `skills/` to replace the base template
- **Suppression**: Create an empty file with the same name to prevent that rule/skill from being generated
- **Example**: Empty `.vendatta/agents/cursor/rules/legacy-code.md` prevents legacy-code rule generation

### **File Resolution Priority**
1. **Project Override**: `.vendatta/agents/cursor/rules/custom.md` (highest priority)
2. **Remote Template**: From `vendatta config pull` sources
3. **Base Template**: Built-in defaults (lowest priority)

### **Supported Agents**

#### **Cursor**
- **Template**: `.vendatta/agents/cursor/mcp.json.tpl`
- **Output**: `.cursor/mcp.json`
- **Format**: JSON with mcpServers object
- **Transport**: HTTP to MCP gateway

#### **OpenCode**
- **Templates**:
  - `opencode.json.tpl` → `opencode.json`
  - Shared rules → `.opencode/rules/`
  - Shared skills → `.opencode/skills/`
  - Shared commands → `.opencode/commands/`
- **Features**: MCP integration, custom rules, skills, commands

#### **Claude Desktop**
- **Template**: `claude_desktop_config.json.tpl`
- **Output**: `claude_desktop_config.json` (user's config directory)
- **Format**: JSON with mcpServers using `mcp-remote`

#### **Claude Code**
- **Template**: `claude_code_config.json.tpl`
- **Output**: `claude_code_config.json` (project or global)
- **Format**: Similar to Desktop but for CLI usage

## 5. Shared Templates (Open Standards)

### **Skills** (agentskills.io)
Standardized YAML format with metadata, parameters, execution, permissions.

### **Rules** (agents.md)
Markdown with frontmatter for applicability, priority, and content.

### **Commands**
YAML with steps, environment variables, and metadata.

## 6. Configuration Example

```yaml
# .vendatta/config.yaml
name: my-project
agents:
  - name: opencode
    rules: base
    enabled: true
  - name: cursor
    rules: base
    enabled: true

mcp:
  enabled: true
  port: 3001
  host: localhost
```

Generated configs connect all agents to the MCP gateway with appropriate authentication and capabilities.

## 8. Implementation Status
- ✅ Template merging from multiple sources (base, remotes, agents)
- ✅ Agent config file generation during `dev` command
- ✅ Support for Cursor, OpenCode, Claude Desktop/Code
- ✅ Populate RulesConfig, SkillsConfig, CommandsConfig as JSON
- 🚧 TODO: Implement authentication token generation
