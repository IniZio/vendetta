# Documentation Structure

This directory contains the complete documentation for the vendetta project.

## Overview

```
docs/
├── spec/                    # Technical specifications (final documents)
│   ├── m3.md             # M3: Simplified QEMU Provider (CURRENT)
│   ├── product/           # Product specifications and user stories
│   │   ├── configuration.md
│   │   ├── overview.md
│   │   └── user_stories.md
│   ├── technical/         # Technical architecture and implementation specs
│   │   ├── agent-gateway.md
│   │   ├── architecture.md
│   │   ├── lifecycle.md
│   │   └── plugins.md
│   ├── testing/           # Testing strategies and test cases
│   │   ├── strategy.md
│   │   └── cases.md
│   └── ux/               # User experience specifications
│       └── cli-ux.md
└── planning/                # Planning and milestone documents
    ├── README.md           # Planning overview
    ├── M1_MVP.md         # M1 MVP specification (archived)
    ├── M2_ALPHA.md        # M2 Alpha specification (archived)
    └── tasks/              # Individual task specifications
        ├── CLI-01.md
        ├── CLI-02.md
        ├── CLI-03.md
        ├── CFG-01.md
        ├── COR-01.md
        ├── COR-02.md
        ├── COR-03.md
        ├── COR-04.md
        ├── INF-01.md
        ├── AGT-02.md
        ├── VFY-01.md
        └── VFY-02.md
```

## Key Documents

### Current Specifications
- **`docs/spec/m3.md`**: M3 Simplified QEMU Provider (master specification)
- **`docs/planning/README.md`**: Planning process and milestone overview

### Product Specifications
- Configuration management, user stories, product overview
- CLI/UX specifications for user experience

### Technical Specifications  
- System architecture, agent gateway, plugin system
- Lifecycle management, testing strategies

### Planning Documents
- Historical milestone specifications (M1 MVP, M2 Alpha)
- Detailed task breakdowns and implementation plans

## Document Status

### Active (Current)
- ✅ **M3 Specification**: `docs/spec/m3.md` - Simplified QEMU provider
- ✅ **Product Specs**: All `docs/spec/product/*.md` files
- ✅ **Technical Specs**: All `docs/spec/technical/*.md` files
- ✅ **Testing Specs**: All `docs/spec/testing/*.md` files

### Archived (Superseded)
- 📦 **M1 MVP**: `docs/planning/M1_MVP.md` - Completed milestone
- 📦 **M2 Alpha**: `docs/planning/M2_ALPHA.md` - Completed milestone
- 📦 **Planning Tasks**: `docs/planning/tasks/*.md` - Historical task tracking

## Usage

- **For current implementation**: Refer to `docs/spec/m3.md`
- **For product context**: See `docs/spec/product/` directory
- **For technical details**: Consult `docs/spec/technical/` directory  
- **For testing guidance**: Review `docs/spec/testing/` directory
- **For historical context**: Check `docs/planning/` directory

## Maintenance

- Keep `docs/spec/m3.md` as the single source of truth for M3
- Archive completed milestones to `docs/planning/` directory
- Update task documents in `docs/planning/tasks/` during development
- Ensure all cross-references are updated when specifications change
