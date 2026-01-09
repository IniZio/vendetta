# User Stories: Project Oursky

## 1. Feature Development (The Main Path)
> *As a full-stack developer, I want to start working on a new feature branch in an isolated environment so that my local machine stays clean and I can switch between branches instantly.*

- **Acceptance Criteria**:
    - Running `oursky dev branch-name` creates a dedicated worktree.
    - Dependencies are installed inside a container, not on the host.
    - My existing `docker-compose` setup for the database works inside the Oursky environment (DinD).

## 2. Agent Collaboration (BYOA)
> *As a developer using any AI agent (Cursor, OpenCode, Claude), I want seamless integration with my isolated development environment so that agents can execute commands, follow project standards, and access shared capabilities.*

- **Acceptance Criteria**:
    - Automatic generation of agent configs during `init`/`dev` commands
    - Support for Cursor, OpenCode, Claude Desktop, and Claude Code
    - MCP gateway provides secure tool execution (`exec`, dynamic skills)
    - Agents inherit shared rules, skills, and commands from standard templates
    - Template system allows customization while following open standards

## 3. Microservices & API Discovery
> *As a frontend developer, I want to automatically know the URL of my backend API running in the Oursky environment so that I don't have to manually update my `.env.local` every time.*

- **Acceptance Criteria**:
    - Oursky discovers the host-mapped port for the `api` service.
    - The environment variable `OURSKY_SERVICE_API_URL` is automatically injected into my frontend container.

## 4. Multi-Agent Standardization
> *As a team lead, I want standardized agent configurations across my team so that all developers get consistent AI assistance regardless of their preferred tools.*

- **Acceptance Criteria**:
    - Shared templates in `.vendatta/templates/` follow open standards
    - Agent-specific configs generated from templates with team standards
    - Easy customization of capabilities per project or team needs
    - Version-controlled templates with gitignored generated configs

## 5. Teardown & Cleanup
> *As a developer, I want to completely remove all artifacts (containers, worktrees) related to a finished feature branch so that I don't waste disk space.*

- **Acceptance Criteria**:
    - `oursky kill session-id` destroys the container and the worktree.
    - No dangling Docker volumes or git worktrees remain.
