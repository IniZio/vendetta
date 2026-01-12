# Vendatta

**Isolated development environments that work with AI agents**

Eliminate "it works on my machine" by providing reproducible dev environments that integrate seamlessly with Cursor, OpenCode, Claude, and other AI coding assistants.

## ✨ Features

- **🔒 Branch Isolation**: Git worktrees + containers for true environment separation
- **📦 Single Binary**: Zero dependencies, works anywhere
- **🔌 Plugin System**: Extensible rules, skills, and commands
- **🚀 Service Discovery**: Automatic port mapping and environment variables
- **🐳 Multi-Provider**: Docker, LXC, and more container backends

## 🚀 Quick Start

```bash
# Install (single binary, no dependencies)
curl -fsSL https://raw.githubusercontent.com/IniZio/vendatta/main/install.sh | bash

# Initialize in your project
vendatta init

# Create isolated workspace for your branch
vendatta workspace create my-feature && vendatta workspace up my-feature
```

**That's it!** Your AI agents (Cursor, OpenCode, etc.) are now automatically configured to work in the isolated environment.

### 📺 See It In Action

[![asciicast](https://asciinema.org/a/ziY3DyDyNjSFf9VP.svg)](https://asciinema.org/a/ziY3DyDyNjSFf9VP)

**Complete workflow demo** showing installation, multiple parallel workspaces, service isolation, Git integration, and cleanup.

### Demo Steps

**Installation & Setup:**
```bash
# Install Vendatta
curl -fsSL https://raw.githubusercontent.com/IniZio/vendatta/main/install.sh | bash

# Initialize project
vendatta init

# Configure services
cat .vendatta/config.yaml
```

**Create Multiple Workspaces:**
```bash
# Create first workspace
vendatta workspace create feature/auth
vendatta workspace up feature/auth

# Create second workspace (runs in parallel)
vendatta workspace create feature/payments
vendatta workspace up feature/payments
```

**Check Parallel Execution:**
```bash
# Verify both workspaces running with isolated ports
docker ps

# Output shows:
# vendatta-demo-project-feature-auth (port 8080)
# vendatta-demo-project-feature-payments (port 8081)
```

**Git Workflow:**
```bash
# Work in each workspace
cd .vendatta/worktrees/feature/auth
git add .
git commit -m "Add auth feature"

cd ../feature/payments
git add .
git commit -m "Add payments feature"

# Merge to main
git checkout main
git merge feature/auth
git merge feature/payments
```

**Cleanup:**
```bash
# Stop and remove workspaces
vendatta workspace down feature/auth
vendatta workspace rm feature/auth
vendatta workspace down feature/payments
vendatta workspace rm feature/payments
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


```

Run `vendatta workspace create my-feature && vendatta workspace up my-feature` to create and start your workspace.

## ⚙️ Configuration

Configure your development environment in `.vendatta/config.yaml`:

```yaml
name: "my-fullstack-app"

extends:
  - inizio/vendatta-config-inizio

plugins:
  - golang
  - node

services:
  db: {command: 'docker-compose up -d postgres'}
  api: {command: 'cd server && npm run dev', depends_on: [db]}
  web: {command: 'cd client && npm run dev', depends_on: [api]}
```

All capabilities from loaded plugins are automatically enabled!

User-specific settings are auto-detected and stored in `~/.config/vendatta/config.yaml`:

```yaml
provider: "docker"  # Preferred container provider
# All settings are auto-detected (no manual config needed)
```

## 🛠️ Advanced Commands

**Configuration Management:**
```bash
vendatta apply          # Apply latest config to agent configs
vendatta plugin update  # Update plugins to latest versions
vendatta plugin list    # List loaded plugins
```

**Workspace Management:**
```bash
vendatta workspace create <name>  # Create new workspace
vendatta workspace up <name>      # Start workspace
vendatta workspace down <name>    # Stop workspace
vendatta workspace list           # List all workspaces
vendatta workspace rm <name>      # Remove workspace
```

**Generated Structure:**
```
.cursor/
└── rules/
    └── [plugin-name]/
        ├── rule1.md
        └── rule2.md

.opencode/
├── rules/[plugin-name]/
├── skills/[plugin-name]/
└── commands/[plugin-name]/
```

📖 **Full Documentation**: [Configuration Guide](docs/spec/product/configuration.md)

## 🌟 What It Does

**Service Discovery**: Automatic environment variables for service URLs
```bash
# In your worktree
env | grep VENDATTA_SERVICE
# VENDATTA_SERVICE_DB_URL=postgresql://localhost:5432
# VENDATTA_SERVICE_API_URL=http://localhost:5000
# VENDATTA_SERVICE_WEB_URL=http://localhost:3000
```

**AI Agent Integration**: Automatically configures Cursor, OpenCode, and Claude agents for isolated development

**Plugin System**: Load capabilities from remote repos, enable what you need:
- **Rules**: Coding standards and linting
- **Skills**: AI capabilities (web search, file ops, etc.)
- **Commands**: Development workflows

## 📚 Documentation

- [Configuration Guide](docs/spec/product/configuration.md)
- [Plugin System](docs/spec/technical/plugins.md)
- [API Reference](docs/spec/technical/architecture.md)
