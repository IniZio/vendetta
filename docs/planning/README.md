# Project Planning

Oursky uses a **Milestone > Project > Task** hierarchy.

## 🏷 Projects Legend

| Code | Name | Description |
| :--- | :--- | :--- |
| **INF** | Infrastructure | Docker, LXC, Worktrees, Networking. |
| **COR** | Core / Control | Orchestration logic, config parsing, lifecycle. |
| **AGT** | Agent Gateway | MCP server, SSH, Agent Scaffold sync. |
| **CLI** | CLI / UX | Command structure, output formatting, scaffolding. |

## 📅 Milestones

- [ ] **[M1: CLI MVP](./M1_MVP.md)** - ✅ COMPLETED (Working Docker+Worktree + MCP)
- [ ] **[M2: Alpha](./M2_ALPHA.md)** - 🚧 ACTIVE (Namespaced Plugins, UV-style Locking, Remote Configs)
- [ ] **[M3: Beta](./M3_BETA.md)** - 📝 SPECCED (QEMU, multi-machine coordination)
