# Vendatta

Vendatta eliminates the "it works on my machine" problem by providing isolated, reproducible development environments that work seamlessly with Coding Agents e.g. Cursor, OpenCode, Claude, etc.

## Key Features

- **Single Binary**: Zero-setup installation with no host dependencies
- **Branch Isolation**: Git worktrees provide unique filesystems for every branch
- **AI Agent Integration**: Automatic configuration for Cursor, OpenCode, Claude, and more
- **Service Discovery**: Automatic port mapping and environment variables for multi-service apps
- **Docker-in-Docker**: Run docker-compose projects inside isolated environments

## Quick Start

### Try It Now

Get started in 3 simple steps:

```bash
# 1. Install Vendatta
curl -fsSL https://raw.githubusercontent.com/IniZio/vendatta/main/install.sh | bash

# Add ~/.local/bin to your PATH if not already:
# export PATH="$HOME/.local/bin:$PATH"

# 2. Initialize in your project
vendatta init

# 3. Create and start your workspace
vendatta workspace create my-feature
vendatta workspace up my-feature
```

That's it! Vendatta creates an isolated workspace for your `my-feature` branch with automatic AI agent configuration.

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
- **Step 2**: Created a `.vendatta/` directory with basic configuration and hook templates
- **Step 3**: Created a workspace with branch, worktree, agent configs, and started the isolated environment

Your AI agents (Cursor, OpenCode, etc.) are now automatically configured to work with this isolated workspace.

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

Run `vendatta workspace create my-feature && vendatta workspace up my-feature` to create and start your workspace.

## Configuration Reference

### Project Structure
```
.vendatta/
├── config.yaml          # Main project configuration
├── schema/              # Auto-generated JSON schemas
│   └── config.schema.json # Schema for config.yaml validation
├── hooks/               # Lifecycle scripts (convention-based)
│   ├── create.sh        # Runs during workspace creation
│   ├── up.sh            # Runs during workspace startup
│   ├── stop.sh          # Runs during workspace stop
│   └── down.sh          # Runs during workspace teardown
├── templates/           # Shared AI capabilities
│   ├── skills/          # Reusable AI skills
│   ├── commands/        # Development commands
│   └── rules/           # Coding guidelines
├── agents/              # Agent-specific file overrides
│   └── cursor/
│       └── rules/       # Override/suppress specific rules
└── worktrees/           # Auto-generated environments (gitignored)
```

### Schema Validation & IDE Support

Vendatta provides full JSON schema validation for `config.yaml` with IDE autocompletion:

```bash
# Generate or update the JSON schema
vendatta config generate-schema

# Validate your current config.yaml
vendatta config validate
```

The schema enables:
- **Autocomplete** in VSCode, Cursor, and other editors
- **Validation** with helpful error messages
- **Documentation** tooltips for all configuration options
- **Type safety** for complex configurations

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


### Service Discovery & Port Access

Vendatta automatically discovers running services and provides environment variables for easy access:

**Available in worktrees**: When you run `vendatta workspace up branch-name`, your workspace environment gets these variables:

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
   vendatta init
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

3. **Create and start your workspace**:
   ```bash
   vendatta workspace create new-feature
   vendatta workspace up new-feature
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
vendatta config pull https://github.com/IniZio/dotvendatta

# List pulled template repositories
vendatta config list

# Merge templates into your configuration
# (automatic during workspace creation)
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
# Create and start isolated development workspace
vendatta workspace create feature-branch
vendatta workspace up feature-branch
```

The `up` command starts the session and blocks for logs. Vendatta will show progress as it:
- Merges AI agent templates
- Sets up Git worktree (at `.vendatta/worktrees/<branch>/`)
- Generates agent configurations in the worktree
- Creates and starts the container session
- Maps service ports (services start automatically in the container)
- Runs lifecycle hooks from `.vendatta/hooks/`

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
🔍 Use 'vendatta workspace list' to see running workspaces
🛑 Use 'vendatta workspace down my-project-feature-branch' to stop the workspace
⏳ Services may take a moment to fully start - check URLs when ready
```

### 4. Check Mapped Ports and Services

Once running, Vendatta automatically maps service ports. Check available services:

```bash
# See all running workspaces
vendatta workspace list

# Check environment variables for service URLs (inside workspace)
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

### 6. Sync Configurations

Push your `.vendatta` configs to remote targets:

```bash
# Sync to a specific target
vendatta config sync backup

# Sync to all configured targets
vendatta config sync-all
```

### 7. Clean Up

```bash
# Stop a specific workspace
vendatta workspace stop <name>

# Remove workspace entirely
vendatta workspace rm <name>

# List all workspaces before cleanup
vendatta workspace list
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
- **Agents not connecting**: Verify agent configs are generated in worktree
- **Git conflicts**: Pull latest changes before `vendatta workspace create`
- **Permission issues**: Ensure Docker is accessible and user has proper permissions

---
*Powered by OhMyOpenCode.*
