# vendetta

**Isolated development environments that work with AI agents**

Eliminate "it works on my machine" by providing reproducible dev environments that integrate seamlessly with Cursor, OpenCode, Claude, and other AI coding assistants.

## 🚀 5-Minute Onboarding Guide

Get up and running with vendetta in 5 minutes.

### Step 1: Install Vendetta

```bash
curl -fsSL https://raw.githubusercontent.com/IniZio/vendetta/main/install.sh | bash
```

Verify the installation:

```bash
vendetta version
# Output: Vendetta Version: 1.0.0, Go Version: go1.24.x, Build Date: ...
```

### Step 2: Create Your First Project

```bash
# Create a new project directory
mkdir my-app && cd my-app

# Initialize vendetta (creates .vendetta/ directory)
vendetta init

# Initialize git (required for workspace creation)
git init && git add . && git commit -m "Initial commit"
```

### Step 3: Add a Remote Node

Workspaces run on remote nodes (your own machines or servers).

```bash
# Add your current machine as a remote node
vendetta node add my-machine user@localhost

# Follow the instructions to copy your SSH key to enable passwordless access
ssh-copy-id -i ~/.ssh/id_ed25519_vendetta user@localhost
```

### Step 4: Create Your First Workspace

```bash
# Create a workspace on your remote node
vendetta workspace create my-feature --node my-machine

# Start the workspace
vendetta workspace up my-feature
```

Your workspace is now running in an isolated Docker container on your remote node!

### Step 5: Access Your Workspace

#### Option A: VSCode Remote

```bash
# Connect with VSCode to the worktree
code .vendetta/worktrees/my-feature/
```

#### Option B: Vim/Neovim

```bash
# Edit files directly in the worktree
vim .vendetta/worktrees/my-feature/
```

#### Option C: Shell Access

```bash
# Open a shell in your workspace
vendetta workspace shell my-feature

# Inside the workspace, start a test webserver
python3 -m http.server 8080
```

### Step 6: Verify It Works

From your host machine or in the workspace shell:

```bash
# Test the webserver (if you started one)
curl http://localhost:8080

# List your workspaces
vendetta workspace list

# Check running services on the remote node
ssh user@localhost "docker ps | grep vendetta"
```

### Step 7: Clean Up

```bash
# Stop the workspace when done
vendetta workspace down my-feature

# Remove the workspace completely
vendetta workspace rm my-feature
```

---

## 📋 Command Reference

### Installation & Version

| Command | Description |
|---------|-------------|
| `curl -fsSL ... install.sh \| bash` | Install vendetta |
| `vendetta version` | Show version info |
| `vendetta --help` | Show help |

### Remote Node Management

| Command | Description |
|---------|-------------|
| `vendetta node add <name> <host>` | Add a remote node (auto-generates SSH key) |
| `vendetta node list` | List configured nodes |
| `vendetta coordination start` | Start coordination server (for multi-node setup) |

### Workspace Management

| Command | Description |
|---------|-------------|
| `vendetta workspace create <name> --node <node>` | Create workspace on remote node |
| `vendetta workspace up <name>` | Start a workspace |
| `vendetta workspace down <name>` | Stop a workspace |
| `vendetta workspace shell <name>` | Open shell in workspace |
| `vendetta workspace list` | List all workspaces |
| `vendetta workspace rm <name>` | Remove a workspace |

---

## 🔧 Configuration

Edit `.vendetta/config.yaml` to customize your environment:

```yaml
name: my-app
provider: docker

services:
  web:
    command: "npm run dev"
    port: 3000
  api:
    command: "npm run server"
    port: 4000

docker:
  image: node:20
```

### Node-Specific Configuration

Add to `.vendetta/config.yaml` for remote workspace creation:

```yaml
remote:
  node: my-machine  # Name of the node from vendetta node list
```

---

## 🌐 Multi-Machine Setup (Optional)

For team collaboration or using multiple remote machines:

### On Each Remote Machine

```bash
# Install vendetta
curl -fsSL https://raw.githubusercontent.com/IniZio/vendetta/main/install.sh | bash

# Start the coordination server
vendetta coordination start &
# Server runs on http://localhost:3001

# Start the node agent
VENDETTA_COORD_URL="http://localhost:3001" vendetta agent start &
```

### On Your Local Machine

```bash
# Add each remote node
vendetta node add server-1 user@192.168.1.100
vendetta node add server-2 user@192.168.1.101

# Create workspace on specific node
vendetta workspace create feature --node server-1
```

---

## ✨ Features

- **🔒 Branch Isolation**: Git worktrees + containers for true environment separation
- **📦 Single Binary**: Zero dependencies, works anywhere
- **🔌 Plugin System**: Extensible rules, skills, and commands
- **🚀 Service Discovery**: Automatic port mapping and environment variables
- **🐳 Multi-Provider**: Docker, LXC, and more container backends
- **🤖 AI Agent Ready**: Auto-configures Cursor, OpenCode, Claude, and more
- **🌍 Remote Development**: Works seamlessly across multiple machines

---

## 📚 Documentation

- [Configuration Guide](docs/spec/product/configuration.md)
- [Plugin System](docs/spec/technical/plugins.md)
- [Architecture](docs/spec/technical/architecture.md)
- [Remote Development Guide](docs/REMOTE_CONNECTION_GUIDE.md)

---

## License

MIT
