# Vendatta

Vendatta eliminates the "it works on my machine" problem by providing isolated, reproducible development environments that work seamlessly with Coding Agents e.g. Cursor, OpenCode, Claude, etc.

## Key Features

- **Single Binary**: Zero-setup installation with no host dependencies
- **Branch Isolation**: Git worktrees provide unique filesystems for every branch
- **AI Agent Integration**: Automatic configuration for Cursor, OpenCode, Claude, and more via Model Context Protocol (MCP)
- **Service Discovery**: Automatic port mapping and environment variables for multi-service apps
- **Docker-in-Docker**: Run docker-compose projects inside isolated environments

## Quick Start

### Try It Now

Get started in 2 simple steps:

```bash
# 1. Install Vendatta
curl -fsSL https://raw.githubusercontent.com/IniZio/vendatta/main/install.sh | bash

# Add ~/.local/bin to your PATH if not already:
# export PATH="$HOME/.local/bin:$PATH"

# 2. Initialize in your project
vendatta init

# 3. Start an isolated development session
vendatta dev my-feature
```

That's it! Vendatta creates an isolated environment for your `my-feature` branch with automatic AI agent configuration.

#### Alternative: Build from Source

If you prefer to build from source:

```bash
# Requires Go 1.24+
go build -o vendatta cmd/vendatta/main.go
```

#### Updates

To update to the latest version:

```bash
vendatta update
```

### Understanding What Happened

- **Step 1**: Built a single Go binary that manages everything
- **Step 2**: Created a `.vendatta/` directory with basic configuration templates
- **Step 3**: Generated a Git worktree at `.vendatta/worktrees/my-feature/` and started any configured services

Your AI agents (Cursor, OpenCode, etc.) are now automatically configured to work with this isolated environment.

### Configure for Your Project

Vendatta works with your existing development setup. Edit `.vendatta/config.yaml` to define your services:

```yaml
# Example: Full-stack web app
services:
  db:
    command: "docker-compose up -d postgres"
  api:
    command: "cd server && npm run dev"
    depends_on: ["db"]
  web:
    command: "cd client && npm run dev"
    depends_on: ["api"]

# Enable AI agents
agents:
  - name: "cursor"
    enabled: true
  - name: "opencode"
    enabled: true
```

Run `./vendatta dev my-feature` again to apply your configuration.

## Configuration Reference

### Project Structure
```
.vendatta/
├── config.yaml          # Main project configuration
├── templates/           # Shared AI capabilities
│   ├── skills/          # Reusable AI skills
│   ├── commands/        # Development commands
│   └── rules/           # Coding guidelines
├── agents/              # Agent-specific overrides
└── worktrees/           # Auto-generated environments
```

### Main Configuration

The `.vendatta/config.yaml` file defines your development environment:

```yaml
# Project settings
name: "my-web-app"

# Services to run
services:
  db:
    command: "docker-compose up -d postgres"
    healthcheck:
      url: "http://localhost:5432/health"
  api:
    command: "cd server && npm run dev"
    depends_on: ["db"]
  web:
    command: "cd client && npm run dev"
    depends_on: ["api"]

# AI agents to configure
agents:
  - name: "cursor"
    enabled: true
  - name: "opencode"
    enabled: true

# MCP server settings
mcp:
  enabled: true
  port: 3001
```

### Customizing Templates

#### Adding AI Skills
Create `.vendatta/templates/skills/my-skill.yaml`:
```yaml
name: "my-custom-skill"
description: "Does something useful"
parameters:
  type: object
  properties:
    input: { type: "string" }
execute:
  command: "node"
  args: ["scripts/my-skill.js"]
```

#### Defining Commands
Create `.vendatta/templates/commands/my-command.yaml`:
```yaml
name: "deploy"
description: "Deploy to staging"
steps:
  - name: "Build"
    command: "npm run build"
  - name: "Deploy"
    command: "kubectl apply -f k8s/"
```

#### Setting Coding Rules
Create `.vendatta/templates/rules/team-standards.md`:
```markdown
---
title: "Team Standards"
applies_to: ["**/*.ts", "**/*.js"]
---

# Code Quality Standards
- Use TypeScript for new code
- Functions should be < 30 lines
- Always add return types
```

### Environment Variables

Use variables for dynamic configuration:

```yaml
# In config.yaml
mcp:
  port: "{{.Env.MCP_PORT}}"
```

```bash
export MCP_PORT=3001
./vendatta dev my-branch
```

### Service Discovery & Port Access

Vendatta automatically discovers running services and provides environment variables for easy access:

**Available in worktrees**: When you run `./vendatta dev branch-name`, your worktree environment gets these variables:

- `OURSKY_SERVICE_DB_URL` - Database connection URL
- `OURSKY_SERVICE_API_URL` - API service URL
- `OURSKY_SERVICE_WEB_URL` - Web frontend URL
- And more for each service you define

**Example usage in your code**:

```javascript
// In your frontend config
const apiUrl = process.env.OURSKY_SERVICE_API_URL || 'http://localhost:3001';

// In your API config
const dbUrl = process.env.OURSKY_SERVICE_DB_URL;
```

**Check available services**:

```bash
# In your worktree directory
env | grep OURSKY_SERVICE
```

This eliminates manual port management and ensures your services can communicate seamlessly across the isolated environment.

## Example: Full-Stack Development

1. **Set up your project**:
   ```bash
   ./vendatta init
   ```

2. **Configure services** (edit `.vendatta/config.yaml`):
   ```yaml
   services:
     db:
       command: "docker-compose up -d postgres"
     api:
       command: "cd server && npm run dev"
       depends_on: ["db"]
     web:
       command: "cd client && npm run dev"
       depends_on: ["api"]

   agents:
     - name: "cursor"
       enabled: true
   ```

3. **Start development**:
   ```bash
   ./vendatta dev new-feature
   ```

4. **Code with AI assistance**:
   - Open `.vendatta/worktrees/new-feature/` in Cursor
   - AI agent connects automatically with full environment access

## Complete Feature Walkthrough

This example demonstrates all Vendatta features in a real development workflow.

### 1. Initialize with Remote Templates

For existing projects, pull shared configurations and templates:

```bash
# Initialize the project
vendatta init

# Pull agent templates from a remote repository
vendatta templates pull https://github.com/IniZio/dotvendatta

# List pulled template repositories
vendatta templates list

# Merge templates into your configuration
vendatta templates merge
```

### 2. Configure Your Development Environment

Edit `.vendatta/config.yaml` to define your stack:

```yaml
name: "my-fullstack-app"

services:
  db:
    command: "docker-compose up -d postgres"
    healthcheck:
      url: "http://localhost:5432/health"
  api:
    command: "cd server && npm run dev"
    depends_on: ["db"]
  web:
    command: "cd client && npm run dev"
    depends_on: ["api"]

agents:
  - name: "cursor"
    enabled: true
  - name: "opencode"
    enabled: true

sync_targets:
  - name: "backup"
    url: "https://github.com/your-org/configs.git"
```

### 3. Start Development Session

```bash
# Start isolated development environment
vendatta dev feature-branch
```

The command starts the session in the background and exits once setup is complete. Vendatta will show progress as it:
- Initializes template remotes
- Merges AI agent templates
- Sets up Git worktree (at `.vendatta/worktrees/<branch>/`)
- Generates agent configurations in the worktree
- Creates and starts the container session
- Maps service ports (services start automatically in the container)
- Runs setup hooks (if configured)

Example output:
```
🚀 Starting dev session for branch 'feature-branch'...
📦 Initializing template remotes...
🔧 Merging AI agent templates...
🌳 Setting up Git worktree...
🤖 Generating AI agent configurations...
🐳 Creating docker session...
▶️  Starting session...
🌐 Service port mappings:
  📍 DB → http://localhost:5432
  📍 API → http://localhost:5000
  📍 WEB → http://localhost:3000
🔧 Running setup hook: .vendatta/hooks/setup.sh
✅ Setup hook completed successfully
🚀 Services starting in background...

🎉 Session my-project-feature-branch is ready!
📂 Worktree: /path/to/project/.vendatta/worktrees/feature-branch
💡 Open this directory in your AI agent (Cursor, OpenCode, etc.)
🔍 Use 'vendatta list' to see running sessions
🛑 Use 'vendatta kill my-project-feature-branch' to stop the session
⏳ Services may take a moment to fully start - check URLs when ready
```

### 4. Check Mapped Ports and Services

Once running, Vendatta automatically maps service ports. Check available services:

```bash
# See all running sessions
vendatta list

# Check environment variables for service URLs
env | grep OURSKY_SERVICE
# Output:
# OURSKY_SERVICE_DB_URL=postgresql://localhost:5432
# OURSKY_SERVICE_API_URL=http://localhost:5000
# OURSKY_SERVICE_WEB_URL=http://localhost:3000
```

### 5. Confirm Everything Works

- **Database**: Connect to `OURSKY_SERVICE_DB_URL`
- **API**: Visit `OURSKY_SERVICE_API_URL` or curl it
- **Web App**: Open `OURSKY_SERVICE_WEB_URL` in browser
- **AI Agents**: Open worktree in Cursor/OpenCode, agents connect automatically

### 6. Use MCP Agent Gateway

For direct AI agent integration:

```bash
# Start MCP server for a specific session
vendatta agent <session-id>
```

### 7. Sync Configurations

Push your `.vendatta` configs to remote targets:

```bash
# Sync to a specific target
vendatta remote sync backup

# Sync to all configured targets
vendatta remote sync-all
```

### 8. Clean Up

```bash
# Stop a specific session
vendatta kill <session-id>

# List all sessions before cleanup
vendatta list
```

### Checking Service Status

Services run inside the container. To check if they're healthy:

```bash
# Check container logs
docker logs <container-name>

# Access the running container
docker exec -it <container-name> /bin/bash

# Check service URLs from the port mappings
curl http://localhost:<port>/health
```

### Troubleshooting

- **Services not starting**: Check `.vendatta/config.yaml` syntax and that commands are correct
- **Ports not accessible**: Services may still be starting up - wait a moment
- **Container issues**: Check `docker ps` for running containers
- **Agents not connecting**: Verify MCP port (default 3001) is available and configs are generated
- **Git conflicts**: Pull latest changes before `vendatta dev`
- **Permission issues**: Ensure Docker is accessible and user has proper permissions

---
*Powered by OhMyOpenCode.*
