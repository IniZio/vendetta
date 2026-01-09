# Technical Specification: Lifecycle & Config

## 1. Project Configuration (`.vendatta/config.yaml`)

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
  role: full-stack # Base role for rule selection

hooks:
  setup: .vendatta/hooks/setup.sh
  dev: .vendatta/hooks/dev.sh
```

## 2. Lifecycle States

### **`init`**
Scaffolds the `.vendatta` directory. Creates the base rules and a default provider configuration.

### **`provision` (`oursky dev`)**
1.  **Worktree**: A new git worktree is created for the target branch.
2.  **Container**: The Provider creates a container, bind-mounting the worktree and the docker socket.
3.  **Ports**: Host ports are allocated and URLs calculated.
4.  **Setup**: The `setup.sh` hook is executed with `OURSKY_SERVICE_*` variables injected.

### **`run`**
The `dev.sh` hook is executed. This is intended for long-running processes like `npm run dev`.

### **`teardown`**
Triggered by `oursky kill`. Cleans up the container and optionally the worktree.
