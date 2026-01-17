# Configuration Reference: Project mochi

## 1. Overview

mochi uses a declarative configuration system based on YAML and JSON templates. This reference covers all available configuration options and how to use them effectively.

## 2. Main Configuration (`config.yaml`)

### **Root Structure**
```yaml
name: "project-name"           # Required: Project identifier
description: "Optional description"

extends:                       # Base configurations to extend
  - inizio/nexus-config-inizio

plugins:                       # Conditional plugins to load
  - golang                     # Only if go.mod exists
  - node                       # Only if package.json exists

services: {}                   # Container services definition
```

### **Extends Configuration**

Extends allow you to inherit base configurations from remote repositories, similar to ESLint's extends.

#### **Extend Sources**
```yaml
extends:
  - inizio/nexus-config-inizio     # Base config from remote repo
  - company/base-config               # Company-wide base config
  - owner/repo@branch                 # With optional branch (default: main)
```

Extends are loaded first and provide the foundation rules, skills, and commands.

#### **Updating Extends**
```bash
mochi update         # Fetch latest versions of all extends
```

This command:
- Fetches latest content from remote repositories
- Updates the cached copies in `.mochi/remotes/`
- Displays updated repos with their commit SHAs

**Note:** Normal operations (workspace create, apply) use cached extends with no network calls. Run `mochi update` periodically to get the latest templates.

### **Plugins Configuration**

Plugins add conditional capabilities based on project structure. Plugins are only loaded if their conditions are met (e.g., presence of specific files).

#### **Plugin Examples**
```yaml
plugins:
  - golang          # Loads if go.mod or go.sum exists
  - node            # Loads if package.json exists
  - python          # Loads if requirements.txt or pyproject.toml exists
```

Plugins are loaded after extends and add project-specific capabilities.

#### **Structured Capability Organization**
Plugin capabilities are organized in structured directories:

```
.cursor/rules/[plugin-name]/
.opencode/rules/[plugin-name]/
.opencode/skills/[plugin-name]/
.opencode/commands/[plugin-name]/
```

When you load plugins, all their capabilities are automatically enabled and organized by plugin. This provides a "batteries included" experience where adding a plugin gives you a complete set of capabilities.

For customization, use local overrides in `.mochi/templates/` to modify or remove specific capabilities.

---

### **Reproducible Locking (`mochi.lock`)**

mochi automatically generates a `mochi.lock` file to ensure that your team is always using the exact same versions of all extends.

**Command Workflow:**
1. `mochi init`: Initializes project with optional extends.
2. `mochi update`: Updates all extends to their latest versions and refreshes the cache.
3. `mochi workspace create`: Uses cached extends for offline-safe, deterministic workspace creation.

### **Services Configuration**

#### **Basic Service Definition**
```yaml
services:
  web:
    command: "cd client && npm run dev"
    healthcheck:
      url: "http://localhost:3000"
      interval: 5s
      timeout: 3s
      retries: 5
```

#### **Service Options**
| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `command` | string | Yes | Shell command to start the service |
| `healthcheck.url` | string | No | Health check endpoint URL |
| `healthcheck.interval` | duration | No | Check frequency (default: 5s) |
| `healthcheck.timeout` | duration | No | Check timeout (default: 3s) |
| `healthcheck.retries` | int | No | Max retry attempts (default: 5) |
| `depends_on` | array | No | Services that must start first |

#### **Example: Full-Stack Setup**
```yaml
services:
  db:
    command: "docker-compose up -d postgres"
    healthcheck:
      url: "http://localhost:5432/health"
      interval: 10s
      retries: 10

  api:
    command: "cd server && npm run dev"
    healthcheck:
      url: "http://localhost:5000/api/health"
    depends_on: ["db"]

  web:
    command: "cd client && npm run dev"
    healthcheck:
      url: "http://localhost:3000"
    depends_on: ["api"]
```



### **MCP Configuration**

MCP (Model Context Protocol) servers are automatically configured and started when AI agents are enabled. The MCP server provides secure tool execution and environment access for AI agents.

**Default Configuration:**
- **Port**: 3001
- **Host**: localhost
- **Auto-enabled**: When any agents are detected/enabled

No manual MCP configuration is required - mochi handles this automatically.

### **User-Specific Configuration (`$XDG_CONFIG_HOME/mochi/config.yaml`)**

mochi auto-generates a default user configuration at `$XDG_CONFIG_HOME/mochi/config.yaml` (typically `~/.config/mochi/config.yaml`). This file contains your personal preferences and is never committed to version control.

AI agents are automatically detected from installed CLIs - no manual configuration required.

```yaml
# Preferred container provider
provider: "docker"
```

#### **Auto-Detection**
mochi scans your system for installed AI agents:
- **Cursor**: Detects VS Code with Cursor extension
- **OpenCode**: Detects OpenCode installation
- **Claude Desktop/Code**: Detects Anthropic CLI tools

#### **Supported Agents**
| Agent | Detection Method | Generated Config |
|-------|------------------|------------------|
| `cursor` | VS Code + Cursor extension | `.cursor/mcp.json` |
| `opencode` | `opencode` CLI available | `opencode.json` + `.opencode/` |
| `claude-desktop` | `claude` desktop app | `claude_desktop_config.json` |
| `claude-code` | `claude` CLI tool | `claude_code_config.json` |

### **Docker Configuration**

#### **Container Runtime**
```yaml
docker:
  image: "ubuntu:22.04"         # Base container image
  privileged: false             # Run in privileged mode
  memory: "2g"                  # Memory limit
  cpu: "1.0"                    # CPU limit (cores)
```

**Docker-in-Docker**: Automatically enabled when `Dockerfile` or `docker-compose.yml` files are detected in the project. No manual configuration required.

### **Hooks Configuration**

#### **Convention-Based Lifecycle Scripts**
Hooks are now convention-based and located in `.mochi/hooks/`:

**Base Project Hooks:**
- `.mochi/hooks/create.sh` - Executed during `workspace create` (optional)
- `.mochi/hooks/up.sh` - Executed during `workspace up` (optional)
- `.mochi/hooks/stop.sh` - Executed during `workspace stop` (optional)
- `.mochi/hooks/down.sh` - Executed during `workspace down` (optional)

**Execution:** Scripts must be executable (chmod +x) if present.



## 3. Template System

### **Template Variables**
All `.tpl` files support variable substitution using Go template syntax:

```yaml
# In any .tpl file
mcp:
  server: "http://{{.Host}}:{{.Port}}"
  token: "{{.AuthToken}}"
```

#### **Available Variables**
| Variable | Description | Example |
|----------|-------------|---------|
| `{{.Host}}` | MCP server host | `localhost` |
| `{{.Port}}` | MCP server port | `3001` |
| `{{.AuthToken}}` | Authentication token | `abc123...` |
| `{{.ProjectName}}` | Project name | `my-project` |
| `{{.DatabaseURL}}` | Database connection | `postgresql://...` |

### **Skills Templates** (`templates/skills/`)

Following [agentskills.io](https://agentskills.io) specification:

```yaml
name: "web-search"
description: "Search the web for information"
version: "1.0.0"
author: "Your Team"

parameters:
  type: object
  properties:
    query:
      type: string
      description: "The search query"
    limit:
      type: integer
      default: 10
  required: ["query"]

execute:
  type: "http"
  url: "https://api.searchengine.com/search"
  method: "GET"

permissions:
  - "web:read"
```

### **Commands Templates** (`templates/commands/`)

Standardized command definitions:

```yaml
name: "build"
description: "Build the project"
aliases: ["compile", "make"]

steps:
  - name: "Install dependencies"
    command: "npm install"
  - name: "Lint code"
    command: "npm run lint"
  - name: "Run tests"
    command: "npm test"
  - name: "Build artifacts"
    command: "npm run build"

env:
  NODE_ENV: "production"
```

### **Rules Templates** (`templates/rules/`)

Following [agents.md](https://github.com/agentsmd/agents.md) format:

```markdown
---
title: "Code Quality Standards"
version: "1.0.0"
applies_to: ["**/*.js", "**/*.ts", "**/*.py"]
priority: "high"
---

# Code Quality Standards

## Naming Conventions
- Use camelCase for variables and functions
- Use PascalCase for classes and components
- Use UPPER_CASE for constants

## Code Structure
- Keep functions under 50 lines
- Use early returns to reduce nesting
- Group related functionality together

## Documentation
- Add JSDoc comments for public APIs
- Document complex business logic
- Keep README files current
```

## 4. Agent-Specific Configuration

Agent configurations are generated based on your `$XDG_CONFIG_HOME/mochi/config.yaml` settings and the enabled capabilities from `config.yaml`. Each supported agent gets customized configuration files.

### **Cursor Configuration**
Generated: `.cursor/mcp.json` (when cursor is enabled in `config.local.yaml`)

```json
{
  "mcpServers": {
    "project-name": {
      "type": "http",
      "url": "http://localhost:3001",
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  }
}
```

### **OpenCode Configuration**
Generated: `opencode.json` + `.opencode/` directory (when opencode is enabled)

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4-5",
  "mcp": {
    "project-name": {
      "type": "remote",
      "url": "http://localhost:3001",
      "enabled": true
    }
  },
  "rules": ["code-quality", "collaboration"],
  "skills": ["web-search", "file-ops"],
  "commands": ["build", "deploy"]
}
```

### **Claude Desktop/Code Configuration**
Generated: `claude_desktop_config.json` / `claude_code_config.json` (when enabled)

```json
{
  "mcpServers": {
    "project-name": {
      "command": "npx",
      "args": ["-y", "mcp-remote", "http://localhost:3001"],
      "env": {
        "MCP_AUTH_TOKEN": "YOUR_TOKEN"
      }
    }
  }
}
```

### **OpenCode Configuration**
Generated: `opencode.json` + `.opencode/` directory

```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4-5",
  "mcp": {
    "project-name": {
      "type": "remote",
      "url": "http://localhost:3001",
      "enabled": true
    }
  },
  "rules": ["code-quality", "collaboration"],
  "skills": ["web-search", "file-operations"],
  "commands": ["build", "deploy"]
}
```

### **Claude Desktop/Code Configuration**
Generated: `claude_desktop_config.json` / `claude_code_config.json`

```json
{
  "mcpServers": {
    "project-name": {
      "command": "npx",
      "args": ["-y", "mcp-remote", "http://localhost:3001"],
      "env": {
        "MCP_AUTH_TOKEN": "YOUR_TOKEN"
      }
    }
  }
}
```

## 5. Environment Variables

mochi supports environment variable substitution in configuration:

### **In config.yaml**
```yaml
mcp:
  port: "{{.Env.MCP_PORT}}"
  host: "{{.Env.MCP_HOST}}"
```

### **In Templates**
```yaml
# In any .tpl file
database_url: "{{.Env.DATABASE_URL}}"
api_key: "{{.Env.OPENAI_API_KEY}}"
```

### **Common Environment Variables**
```bash
# MCP Configuration
MCP_PORT=3001
MCP_HOST=localhost
MCP_AUTH_TOKEN=your-secret-token

# Project Configuration
PROJECT_NAME=my-awesome-project
DATABASE_URL=postgresql://user:pass@localhost:5432/db

# API Keys
OPENAI_API_KEY=sk-...
GITHUB_TOKEN=ghp_...
```

## 6. Best Practices

### **Configuration Organization**
- Keep `config.yaml` focused on services and agents
- Use templates for reusable capabilities
- Document custom configurations with comments

### **Security**
- Never commit sensitive data to version control
- Use environment variables for secrets
- Regularly rotate authentication tokens

### **Performance**
- Use appropriate health check intervals
- Configure resource limits for containers
- Enable D in D only when needed

### **Maintenance**
- Version control your `.mochi/` directory
- Test configuration changes in isolated branches
- Document custom templates and their purpose

## 7. Troubleshooting

### **Common Issues**

#### **MCP Connection Failed**
```yaml
# Check mcp configuration
mcp:
  enabled: true
  port: 3001
  host: "localhost"
```

#### **Agent Config Not Generated**
Ensure the AI agent CLI is installed and available in PATH. mochi auto-detects supported agents.

#### **Container Won't Start**
```yaml
# Check service dependencies
services:
  api:
    depends_on: ["db"]  # Wait for database first
```

#### **Template Variables Not Substituted**
- Ensure variables are defined in the correct context
- Check for typos in variable names
- Verify environment variables are set

### **Debugging Commands**
```bash
# Check MCP server status
curl http://localhost:3001/health

# View generated configurations
cat .cursor/mcp.json
cat opencode.json

# Check container logs
docker logs mochi-session-123
```

## 8. Migration Guide

### **Upgrading from Manual Config**
1. Run `mochi init` to generate new structure
2. Move existing configs to appropriate template directories
3. Update `config.yaml` with your settings
4. Test with `mochi dev test-branch`

### **From Other Tools**
- **docker-compose**: Move service definitions to `services:` section
- **Manual scripts**: Convert to templates with `{{.Variable}}` syntax
- **Environment files**: Use `{{.Env.VAR_NAME}}` substitution
