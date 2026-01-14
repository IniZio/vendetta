# vendetta

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
curl -fsSL https://raw.githubusercontent.com/IniZio/vendetta/main/install.sh | bash

# Initialize in your project
vendetta init

# Create isolated workspace for your branch
vendetta workspace create my-feature && vendetta workspace up my-feature
```

**That's it!** Your AI agents (Cursor, OpenCode, etc.) are now automatically configured to work in the isolated environment.

### 📺 See It In Action

[![asciicast](https://asciinema.org/a/ziY3DyDyNjSFf9VP.svg)](https://asciinema.org/a/ziY3DyDyNjSFf9VP)

**Complete workflow demo** showing installation, multiple parallel workspaces, service isolation, Git integration, and cleanup.

### Demo Steps

**Installation & Setup:**
```bash
# Install vendetta
curl -fsSL https://raw.githubusercontent.com/IniZio/vendetta/main/install.sh | bash

# Initialize project
vendetta init

# Configure services
cat .vendetta/config.yaml
```

**Create Multiple Workspaces:**
```bash
# Create first workspace
vendetta workspace create feature/auth
vendetta workspace up feature/auth

# Create second workspace (runs in parallel)
vendetta workspace create feature/payments
vendetta workspace up feature/payments
```

**Check Parallel Execution:**
```bash
# Verify both workspaces running with isolated ports
docker ps

# Output shows:
# vendetta-demo-project-feature-auth (port 8080)
# vendetta-demo-project-feature-payments (port 8081)
```

**Git Workflow:**
```bash
# Work in each workspace
cd .vendetta/worktrees/feature/auth
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
vendetta workspace down feature/auth
vendetta workspace rm feature/auth
vendetta workspace down feature/payments
vendetta workspace rm feature/payments
```

### Understanding What Happened

- **Step 1**: Built a single Go binary that manages everything
- **Step 2**: Created a `.vendetta/` directory with basic configuration and hook templates
- **Step 3**: Created a workspace with branch, worktree, agent configs, and started the isolated environment

Your AI agents (Cursor, OpenCode, etc.) are now automatically configured to work with this isolated workspace.

### Configure for Your Project

vendetta works with your existing development setup. Edit `.vendetta/config.yaml` to define your services:

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

Run `vendetta workspace create my-feature && vendetta workspace up my-feature` to create and start your workspace.

## ⚙️ Configuration

Configure your development environment in `.vendetta/config.yaml`:

```yaml
name: "my-fullstack-app"

extends:
  - inizio/vendetta-config-inizio

plugins:
  - golang
  - node

services:
  db: {command: 'docker-compose up -d postgres'}
  api: {command: 'cd server && npm run dev', depends_on: [db]}
  web: {command: 'cd client && npm run dev', depends_on: [api]}
```

All capabilities from loaded plugins are automatically enabled!

User-specific settings are auto-detected and stored in `~/.config/vendetta/config.yaml`:

```yaml
provider: "docker"  # Preferred container provider
# All settings are auto-detected (no manual config needed)
```

## 🛠️ Advanced Commands

**Configuration Management:**
```bash
vendetta apply          # Apply latest config to agent configs
vendetta update         # Update all extends to latest versions
vendetta plugin update  # Update plugins to latest versions
vendetta plugin list    # List loaded plugins
```

**Workspace Management:**
```bash
vendetta workspace create <name>  # Create new workspace
vendetta workspace up <name>      # Start workspace
vendetta workspace down <name>    # Stop workspace
vendetta workspace list           # List all workspaces
vendetta workspace rm <name>      # Remove workspace
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

**Distributed Development**: Remote execution across multiple nodes with provider-agnostic support
```bash
# Local development
vendetta workspace create my-feature && vendetta workspace up my-feature

# Remote development with coordination server
vendetta node add my-node remote.example.com
vendetta workspace create my-feature --node my-node
```

**Service Discovery**: Automatic environment variables for service URLs
```bash
# In your worktree
env | grep vendetta_SERVICE
# vendetta_SERVICE_DB_URL=postgresql://localhost:5432
# vendetta_SERVICE_API_URL=http://localhost:5000
# vendetta_SERVICE_WEB_URL=http://localhost:3000
```

**AI Agent Integration**: Automatically configures Cursor, OpenCode, and Claude agents for isolated development

**Multi-Provider Support**: Docker, LXC, and QEMU with unified interface
```yaml
# Choose your provider
provider: docker  # or lxc, qemu

# Remote configuration
remote:
  node: remote-server.example.com
  user: developer
  port: 22
```

**Advanced Service Orchestration**: Health monitoring and auto-restart capabilities
- Dependency resolution and startup ordering
- Health checks with configurable intervals
- Automatic restart on service failure
- Real-time service status tracking

## 📚 Documentation

- [Configuration Guide](docs/spec/product/configuration.md)
- [Plugin System](docs/spec/technical/plugins.md)
- [API Reference](docs/spec/technical/architecture.md)
