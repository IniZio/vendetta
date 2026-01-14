# vendetta Example Project

This directory contains a complete working example of an vendetta-configured project. It demonstrates a full-stack web application with database, API, and frontend services, plus AI agent integration.

## What's Included

### Services
- **Database**: PostgreSQL with health checks
- **API**: Node.js/Express server with REST endpoints
- **Frontend**: Vite + React application

### AI Agent Support
- **Cursor**: Agent configuration for VS Code integration
- **OpenCode**: `opencode.json` + `.opencode/` directory
- **Claude**: Desktop and Code configurations

### Capabilities
- **Skills**: Web search, file operations, data analysis
- **Commands**: Build, deploy, git operations
- **Rules**: Code quality standards, collaboration guidelines

## Quick Start

1. **Initialize vendetta**:
   ```bash
   cd example
   ../vendetta init
   ```

2. **Start development environment**:
   ```bash
   ../vendetta dev example-branch
   ```

3. **Open in your AI agent**:
   - **Cursor**: Open `.vendetta/worktrees/example-branch/`
   - **OpenCode**: Uses the generated `opencode.json`
   - **Claude**: Uses the generated config files

## Configuration Structure

```
.vendetta/
├── config.yaml          # Project configuration
├── templates/           # Shared AI capabilities
│   ├── skills/          # Reusable skills
│   ├── commands/        # Development commands
│   └── rules/           # Coding guidelines
├── agents/              # Agent-specific templates
└── worktrees/           # Generated environments
```

## Testing the Setup

### API Endpoints
```bash
# Check health
curl http://localhost:5000/api/health

# Get items
curl http://localhost:5000/api/items

# Add an item
curl -X POST http://localhost:5000/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Item","description":"Example item"}'
```

### Frontend
Visit `http://localhost:3000` to see the React application.

### Database
PostgreSQL runs on `localhost:5432` with credentials from `.env`.

## Customizing

### Add Your Own Skills
Create `.vendetta/templates/skills/my-skill.yaml`:
```yaml
name: "my-skill"
description: "Does something useful"
parameters:
  type: object
  properties:
    input: { type: "string" }
execute:
  command: "node"
  args: ["scripts/my-skill.js"]
```

### Modify Rules
Edit `.vendetta/templates/rules/code-quality.md` to match your team's standards.

### Configure Services
Update `.vendetta/config.yaml` to change ports, add services, or enable different agents.

## Troubleshooting

### Services Won't Start
Check Docker is running and ports are available:
```bash
docker ps
netstat -tlnp | grep :3000
```

### Agent Not Connecting
Check generated agent configs in the worktree.

### Configuration Issues
Check generated configs in `.vendetta/worktrees/<branch>/`

## Learn More

- [Main README](../../README.md) - General vendetta documentation
- [Configuration Reference](../../docs/spec/product/configuration.md) - Detailed config options
- [Technical Specs](../../docs/spec/technical/) - Architecture details
