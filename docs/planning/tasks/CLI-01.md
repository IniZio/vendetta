# Task: CLI-01 CLI Scaffolding & Agent Config Generation

**Priority**: ⚡ Med
**Status**: [Completed]

## 🎯 Objective
Implement comprehensive agent configuration generation system using templates, supporting multiple AI agents (Cursor, OpenCode, Claude) with shared standard templates.

## 🛠 Implementation Details

### **Template System Architecture**
1. **Shared Templates** (`.vendatta/templates/`):
   - `skills/` - agentskills.io compliant YAML skills
   - `commands/` - Standardized command definitions
   - `rules/` - agents.md compliant markdown rules

2. **Agent-Specific Templates** (`.vendatta/agents/{agent}/`):
   - Template files with `.tpl` extension
   - Variable substitution for dynamic config generation

3. **Config-Driven Generation**:
   - Read `.vendatta/config.yaml` for enabled agents
   - Generate configs using Go templates with variable substitution
   - Copy referenced shared templates to agent directories

### **Supported Agents**
- **Cursor**: Generates `.cursor/mcp.json` with HTTP MCP connection
- **OpenCode**: Generates `opencode.json` + `.opencode/` directory with rules, skills, commands
- **Claude Desktop/Code**: Generates `claude_*_config.json` with `mcp-remote` connections

### **Variable Substitution**
- `{{.Host}}`, `{{.Port}}`, `{{.AuthToken}}` - MCP server settings
- `{{.ProjectName}}`, `{{.DatabaseURL}}` - Project-specific values
- `{{.RulesConfig}}`, `{{.SkillsConfig}}`, `{{.CommandsConfig}}` - JSON objects for enabled capabilities

### **Git Safety**
- Generated files added to `.gitignore`
- Only templates committed to version control
- Clean separation of source vs generated content

## 🧪 Proof of Work
- ✅ Template-based generation with variable substitution
- ✅ Shared templates following open standards (agentskills.io, agents.md)
- ✅ Multi-agent support (Cursor, OpenCode, Claude variants)
- ✅ MCP integration with proper authentication
- ✅ Git-ignored generated configs with committed templates
