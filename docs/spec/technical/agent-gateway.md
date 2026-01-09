# Technical Specification: Agent Gateway & Config Generation

## 1. Overview
The Agent Gateway is a Model Context Protocol (MCP) server built into the `oursky` binary, providing a bridge between AI agents (Cursor, OpenCode, Claude) and isolated development environments. The system includes comprehensive agent configuration generation from templates.

## 2. MCP Protocol Details
- **Transport**: JSON-RPC over Standard Input/Output (STDIO)
- **Session Context**: Started per session: `./oursky agent [session-id]`
- **Server Location**: Configurable via `.oursky/config.yaml` (default: localhost:3001)

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
.oursky/
├── config.yaml                 # Main configuration with agent list
├── templates/                  # Shared templates (open standards)
│   ├── skills/                 # agentskills.io compliant
│   ├── commands/               # Standardized command definitions
│   └── rules/                  # agents.md compliant
├── agents/                     # Agent-specific config templates
│   ├── cursor/mcp.json.tpl
│   ├── opencode/opencode.json.tpl
│   ├── claude-desktop/claude_desktop_config.json.tpl
│   └── claude-code/claude_code_config.json.tpl
└── worktrees/                  # Generated worktrees (gitignored)
```

### **Generation Process**
1. **Read** `.oursky/config.yaml` for enabled agents
2. **Load** agent-specific templates from `.oursky/agents/{agent}/`
3. **Substitute** variables (MCP host/port, auth tokens, project metadata)
4. **Copy** referenced shared templates to agent directories
5. **Generate** final configs in correct locations

### **Template Variables**
- `{{.Host}}` - MCP server host (default: localhost)
- `{{.Port}}` - MCP server port (default: 3001)
- `{{.AuthToken}}` - Authentication token
- `{{.ProjectName}}` - Project name for rules
- `{{.DatabaseURL}}` - Database connection string
- `{{.RulesConfig}}` - JSON object of enabled rules
- `{{.SkillsConfig}}` - JSON object of enabled skills
- `{{.CommandsConfig}}` - JSON object of enabled commands

### **Supported Agents**

#### **Cursor**
- **Template**: `.oursky/agents/cursor/mcp.json.tpl`
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
# .oursky/config.yaml
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
