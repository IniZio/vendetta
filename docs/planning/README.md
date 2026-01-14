# Project Planning

vendetta uses a **Milestone > Project > Task** hierarchy.

## 🏷 Projects Legend

| Code | Name | Description |
| :--- | :--- | :--- |
| **INF** | Infrastructure | Docker, LXC, Worktrees, Networking. |
| **COR** | Core / Control | Orchestration logic, config parsing, lifecycle. |
| **AGT** | Agent Integration | SSH, Agent Scaffold sync, AI agent configs. |
| **CLI** | CLI / UX | Command structure, output formatting, scaffolding. |

## 📅 Milestones

- [x] **[M1: CLI MVP](./M1_MVP.md)** - ✅ COMPLETED (Working Docker+Worktree + Agent Integration)
- [x] **[M2: Alpha](./M2_ALPHA.md)** - ✅ COMPLETED (Namespaced Plugins, UV-style Locking, Remote Configs)
- [ ] **[M3: Provider-Agnostic Remote Nodes](../spec/m3.md)** - 🚧 ACTIVE (33% Complete - Coordination Server Critical Path)

## 📋 Current Status Documents

### M3 Implementation Tracking
- **[M3_IMPLEMENTATION_STATUS.md](./M3_IMPLEMENTATION_STATUS.md)** - **UPDATED**: Corrected 20% completion analysis  
- **[M3_ARCHITECTURAL_CORRECTION.md](./M3_ARCHITECTURAL_CORRECTION.md)** - **NEW**: Critical architectural understanding corrections
- **[M3_NEXT_STEPS.md](./M3_NEXT_STEPS.md)** - Updated prioritized implementation plan (8-10 weeks)
- **[M3_ROADMAP.md](./M3_ROADMAP.md)** - Updated roadmap with critical path focus
- **[M3_COORDINATION_SERVER_PLAN.md](./M3_COORDINATION_SERVER_PLAN.md)** - Critical component implementation plan
- **[M3_PROVIDER_REMOTE_SUPPORT.md](./M3_PROVIDER_REMOTE_SUPPORT.md)** - Docker/LXC remote support plan

### Progress Summary - CORRECTED
**M3 Current Status**: 20% Complete 🔴 (Corrected from 33%)  
**Critical Path**: Transport Layer + Coordination Server + Node Agents  
**Timeline**: 8-10 weeks to completion  
**Next Milestone**: M3.1 (50% complete) - February 14, 2026  
**Critical Update**: Architecture corrections revealed 80% of remote infrastructure missing
