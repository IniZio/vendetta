# CLI UX Specification: Project vendetta

## 1. Design Philosophy
- **Speed**: Commands should provide immediate feedback (sub-second for information, clear progress for IO).
- **Clarity**: Status updates should use colors and clear labels.
- **Actionability**: Error messages must suggest a solution (e.g., "Docker not running. Start Docker and try again.").
- **Scriptability**: Global `--json` flag for all commands to allow integration into other tools.

## 2. Command Structure

| Command | Usage | Feedback Pattern |
| :--- | :--- | :--- |
| `init` | `vendetta init` | Interactive prompts + Success checklist. |
| `workspace create <name>` | `vendetta workspace create <branch>` | Progress bars for worktree creation & config generation. |
| `workspace up [name]` | `vendetta workspace up [branch]` | Starts container + runs hooks + port forwarding. Blocks for logs unless `-d`. |
| `workspace shell [name]` | `vendetta workspace shell [branch]` | Opens interactive shell in workspace container. |
| `workspace stop [name]` | `vendetta workspace stop [branch]` | Stops container but preserves state. |
| `workspace down [name]` | `vendetta workspace down [branch]` | Stops and removes container/network. |
| `workspace list` | `vendetta workspace list` | Tabular data with color-coded status (Active=Green, Stopped=Yellow). |
| `workspace rm <name>` | `vendetta workspace rm <branch>` | Deletes worktree and associated resources. |
| `config pull <url>` | `vendetta config pull <url>` [--branch=branch] | Pulls shared templates/capabilities from Git repo. |
| `config sync <target>` | `vendetta config sync <target>` | Syncs .vendetta directory to configured remote target. |
| `config sync-all` | `vendetta config sync-all` | Syncs .vendetta to all configured remote targets. |
| `version update` | `vendetta version update` | Updates CLI binary to latest version. |
| `internal mcp <id>` | `vendetta internal mcp <session-id>` | Hidden: MCP server for AI agent tool execution. |

## 3. Command Groups

### **Workspace Commands**
Primary commands for managing isolated development environments:
- `workspace create <name>` - Creates and configures a new workspace
- `workspace up [name]` - Starts the workspace and runs services
- `workspace shell [name]` - Opens interactive shell in workspace
- `workspace stop [name]` - Stops workspace services
- `workspace down [name]` - Stops and cleans up workspace
- `workspace list` - Shows all active workspaces
- `workspace rm <name>` - Permanently removes workspace

### **Config Commands**
Commands for managing shared configurations and templates:
- `config pull <url>` - Pulls templates from remote Git repositories
- `config sync <target>` - Syncs local config to remote targets
- `config sync-all` - Syncs to all configured remote targets

### **Utility Commands**
- `version update` - Updates the CLI to latest version
- `init` - Initializes project with basic configuration

## 4. Feedback Elements

### **Progress Indicators**
For long-running tasks like `Image Pull` or `Worktree Setup`:
```text
[1/3] Creating worktree 'feature-x'... OK
[2/3] Pulling docker image 'node:20'... [=====>    ] 60%
[3/3] Running setup hook... 
```

### **Error Handling**
Errors should follow the **Context-Problem-Solution** pattern:
```text
Error: Failed to bind-mount worktree.
Problem: The directory '/home/user/repo' is not shared with Docker.
Solution: Add this directory to Docker Desktop > Settings > Resources > File Sharing.
```

## 4. Visual Language
- **Accent Color**: Sky Blue (`#00BFFF`).
- **Success**: Green Checkmark (`✔`).
- **Warning**: Yellow Triangle (`⚠`).
- **Error**: Red Cross (`✖`).

## 5. Agent Interoperability UX

### **Automatic Config Generation**
- Agent configs are generated automatically during `init` and `dev` commands
- Templates use variable substitution for project-specific settings
- Generated files are gitignored to prevent version control pollution

### **Multi-Agent Support**
- Simultaneous configuration for Cursor, OpenCode, Claude Desktop/Code
- Each agent gets appropriate config format and connection settings
- Shared templates ensure consistency across agents

### **MCP Gateway**
- `vendetta agent <session-id>` starts the MCP server for the session
- Robust connection handling with automatic recovery
- Secure tool execution within isolated environments

### **Template System**
- Shared templates follow open standards (agentskills.io, agents.md)
- Agent-specific templates with `.tpl` extension
- Easy customization and extension of capabilities
