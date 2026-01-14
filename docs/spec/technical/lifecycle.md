# Technical Specification: Lifecycle & Config

## 1. Project Configuration (`.vendetta/config.yaml`)

vendetta uses a declarative YAML configuration system with full JSON schema validation and IDE support.

### Schema Validation
- **Auto-generated schema**: JSON schema is automatically generated from Go structs
- **IDE integration**: VSCode, Cursor, and other editors provide autocomplete and validation
- **Validation commands**:
  ```bash
  vendetta config generate-schema  # Generate .vendetta/schema/config.schema.json
  vendetta config validate         # Validate current config.yaml
  ```

### Schema Location
- **Schema file**: `.vendetta/schema/config.schema.json`
- **Auto-generated**: Updated automatically when config structs change

```yaml
name: project-name
provider: docker

# Service port definitions for discovery
services:
  web: 3000
  api: 8080

docker:
  image: node:20-alpine
  dind: true  # Enables Docker-in-Docker

agent:
  enabled: true

hooks:
  setup: .vendetta/hooks/setup.sh
  dev: .vendetta/hooks/dev.sh
```

## 2. Lifecycle States

### **`init`**
Scaffolds the `.vendetta` directory. Creates the base configuration and templates.

### **`workspace create <name>`**
1.  **Branch**: Creates or switches to the specified git branch.
2.  **Worktree**: Creates a git worktree in `.vendetta/worktrees/<name>/` (if `-w` flag used).
3.  **Agent Configs**: Generates AI agent configurations (Cursor, OpenCode, etc.) from templates.
4.  **Hooks**: Runs `.vendetta/hooks/create.sh` if it exists.

### **`workspace up [name]`**
1.  **Container**: Starts the Docker container with worktree bind-mounted.
2.  **Port Forwarding**: Maps service ports and injects `vendetta_SERVICE_*` environment variables.
3.  **Hooks**: Executes `.vendetta/hooks/up.sh` if it exists.
4.  **Blocking**: Streams logs and maintains session until Ctrl+C (or detached with `-d`).

### **`workspace stop [name]`**
Stops the container but preserves state and resources.

### **`workspace down [name]`**
Stops and removes the container, networks, and temporary resources.

### **`workspace rm <name>`**
Deletes the worktree directory and all associated workspace resources.
