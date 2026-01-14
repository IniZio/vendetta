# Task: COR-01 Orchestration Engine

**Priority**: 🔥 High
**Status**: [Completed]

## 🎯 Objective
Coordinate the lifecycle of a dev environment: Worktree creation, Container provisioning, and Hook execution.

## 🛠 Implementation Details
1.  **Worktree Manager**:
    - Add `git worktree` at `.vendetta/worktrees/[branch]`.
    - Handle absolute path resolution for Docker.
2.  **Lifecycle Orchestration**:
    - Sequence: `Load Config` -> `Add Worktree` -> `Provider Create` -> `Provider Start` -> `Exec Setup Hook`.
3.  **Env Injection**:
    - Discover ports from active provider session.
    - Inject `vendetta_SERVICE_[NAME]_URL` into the `Exec` context for hooks.

## 🧪 Proof of Work
- Running `vendetta dev branch-name` creates the worktree directory.
- `setup.sh` hook runs successfully inside the container.
- Environment variables are present in the hook's execution context.
