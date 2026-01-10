---
title: "project"
description: "Vendatta Project Knowledge Base - Architecture, Conventions, and Anti-Patterns"
globs: ["**/*"]
alwaysApply: true
---

## OVERVIEW
Vendatta eliminates the "it works on my machine" problem by providing isolated, reproducible development environments. Built in Go 1.24, it orchestrates Git worktrees, Docker/LXC containers, and AI Agent configurations (Cursor, OpenCode, Claude) via Model Context Protocol (MCP).

## STRUCTURE
```
vibegear/
├── cmd/
│   └── oursky/        # CLI entry point (main.go)
├── pkg/
│   ├── config/        # YAML/JSON config parsing & Agent rule generation
│   ├── ctrl/          # Core orchestration logic (Controller)
│   ├── templates/     # Rule/Skill/Command merging & rendering
│   ├── provider/      # Session providers (Docker, LXC)
│   └── worktree/      # Git worktree management
├── internal/          # Shared internal utilities
├── docs/              # Specifications and planning tasks
├── example/           # Full-stack example project
└── .vendatta/         # Core configuration templates & rules
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Add Agent | `pkg/config/config.go` | Update `agentConfigs` map and generation logic |
| Modify Lifecycle | `pkg/ctrl/ctrl.go` | `WorkspaceCreate`, `WorkspaceUp`, `WorkspaceDown` |
| New Provider | `pkg/provider/` | Implement `Provider` interface |
| Rule Merging | `pkg/templates/` | `merge.go` recursive merging logic |
| CLI Commands | `cmd/oursky/main.go` | Root command and subcommands |

## TDD (Test-Driven Development)
**MANDATORY for all logic changes.** Follow RED-GREEN-REFACTOR:
1. **RED**: Write failing test in `*_test.go`
2. **GREEN**: Implement minimal code to pass
3. **REFACTOR**: Clean up while keeping tests green

**Rules:**
- Never write implementation before test
- Use `testify/assert` and `testify/require`
- Test file naming: `*.test.go` alongside source

## CONVENTIONS
- **Language**: Go 1.24
- **Error Handling**: Always wrap errors: `fmt.Errorf("failed to...: %w", err)`
- **Configuration**: Declarative YAML in `.vendatta/config.yaml`
- **Agent Rules**: Markdown with frontmatter, managed in `.vendatta/`
- **Naming**: `pkg/` for exported modules, `internal/` for private implementation

## ANTI-PATTERNS (THIS PROJECT)
- **Manual Ports**: Never hardcode ports in code; use `Service` discovery
- **Absolute Paths**: Never use absolute paths in templates (use `{{.ProjectName}}`)
- **interface{}**: Avoid empty interfaces unless truly dynamic (prefer Generics or Interfaces)
- **Large Commits**: 3+ files changed = split into multiple atomic commits
- **Missing Tests**: No logic PR should be merged without 80%+ coverage on new code

## UNIQUE STYLES
- **Factory Pattern**: Controllers and Providers created via `New...()` functions
- **Template First**: Agent settings should be generated from templates, not hardcoded
- **Isolation**: Every branch must be able to run in a dedicated worktree without interference

## COMMANDS
```bash
go run cmd/oursky/main.go init              # Initialize project
go run cmd/oursky/main.go workspace create  # Create worktree & agent rules
go run cmd/oursky/main.go workspace up      # Start session & services
go test ./...                               # Run all tests
```

## NOTES
- **LXC Support**: Under development (see M2 milestone)
- **MCP Gateway**: Built-in server on port 3001 by default
- **Security**: Worktree directories are gitignored via `.vendatta/worktrees/`
