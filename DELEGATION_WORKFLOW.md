# Sisyphus Delegation Workflow

**Purpose**: Enable parallel execution with minimal direct involvement from Sisyphus. Sisyphus orchestrates, teammates execute.

---

## Core Principle

```
┌─────────────────────────────────────────┐
│         SISYPHUS (Orchestrator)         │
│  • Understands requirements             │
│  • Plans & delegates                    │
│  • Synthesizes results                  │
│  • Makes decisions                      │
└──────────────────┬──────────────────────┘
                   │
        ┌──────────┼──────────┬──────────────┐
        │          │          │              │
        ▼          ▼          ▼              ▼
   ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────────┐
   │ Docs   │ │Backend │ │Explore │ │   Oracle     │
   │Writer  │ │Dev     │ │/Search │ │   (Complex)  │
   │ 40%    │ │ 40%    │ │ 10%    │ │     5%       │
   └────────┘ └────────┘ └────────┘ └──────────────┘
```

**Time Split**: 
- 40% Documentation (create, update, maintain)
- 40% Backend/Implementation (code, files, tests)
- 10% Research (discovery, analysis)
- 5% Strategic decisions (architecture, trade-offs)
- **5% Sisyphus Coordination** (tiny!)

---

## Task Delegation Framework

### Template for Every Task

```yaml
TASK: [Brief description]
TYPE: [documentation|implementation|research|decision]

WHAT: [Specific deliverable]
WHO: [Agent responsible]
HOW: [Agent's specific instructions]
VERIFY: [How to confirm success]
DEPEND: [Dependencies/blockers]
```

---

## Agent Capabilities & Tasks

### 📝 Document-Writer Agent (40% of work)

**Capabilities**: Create, update, reorganize documentation. Own all README files, API docs, guides.

**Tasks**:
- [ ] Create new specs/documentation
- [ ] Update cross-references
- [ ] Maintain README/navigation
- [ ] Create migration guides
- [ ] Document architecture decisions
- [ ] Maintain changelog

**Delegation Pattern**:
```
TASK: Create Sprint 1 documentation
WHO: document-writer
HOW:
  MUST DO:
    - Use sprint-template.md as base
    - Add M3.1 tasks: COR-05, COR-06, AGT-03, INF-02, VFY-03
    - Set success criteria from SPRINT_FRAMEWORK.md
    - Include daily standup placeholders
  MUST NOT DO:
    - Create task details (use backlog.md)
    - Estimate beyond team capacity
    - Make task assignments without input
VERIFY:
  - Sprint document complete and readable
  - All success criteria clear
  - No broken internal links
  - Team can start sprint immediately
```

---

### 💻 Backend Developer Agent (40% of work)

**Capabilities**: Code changes, file operations, build/test execution, refactoring.

**Tasks**:
- [ ] Implement features
- [ ] Fix bugs
- [ ] Run tests/builds
- [ ] Refactor code
- [ ] Execute file operations
- [ ] Create commits (with my approval)

**Delegation Pattern**:
```
TASK: Implement coordination server foundation
WHO: backend-dev (or sisyphus:sisyphus-junior-high for complex work)
HOW:
  MUST DO:
    - Follow existing code patterns in pkg/
    - Use testify for assertions
    - Write tests BEFORE implementation (TDD)
    - Run lsp_diagnostics on all changed files
    - Keep commits focused & atomic
    - Update CHANGELOG.md
  MUST NOT DO:
    - Use interface{} without justification
    - Suppress type errors (as any, @ts-ignore)
    - Delete failing tests
    - Create giant PRs (3+ files = split commits)
VERIFY:
  - Tests pass
  - 80%+ coverage on new code
  - lsp_diagnostics clean
  - Code review checklist complete
```

---

### 🔍 Explore Agent (10% of work)

**Capabilities**: Fast codebase search, pattern discovery, file finding.

**Tasks**:
- [ ] Find code patterns
- [ ] Locate implementations
- [ ] Analyze dependencies
- [ ] Map file structures
- [ ] Identify naming conventions

**Delegation Pattern**:
```
TASK: Find all SSH key handling code
WHO: explore
HOW:
  MUST DO:
    - Search for SSH-related patterns
    - Find all key generation code
    - Identify key storage locations
    - Map dependencies between modules
  RETURN:
    - File list with locations
    - Pattern summary
    - Dependency map
VERIFY:
  - Results match manual spot-checks
  - No critical files missed
```

---

### 🧠 Oracle Agent (5% of work)

**Capabilities**: Architecture review, complex design decisions, trade-off analysis.

**Tasks**:
- [ ] Review major architectural changes
- [ ] Solve complex debugging issues
- [ ] Evaluate design trade-offs
- [ ] Provide strategic guidance

**When to Use**: After 2+ failed attempts, or before major decisions.

**Delegation Pattern**:
```
TASK: Decide on coordination server architecture
WHO: oracle
HOW:
  - Review: M3 spec + current coordination-api.md
  - Analyze: Node agent pattern vs direct execution
  - Consider: Performance, maintainability, team capacity
  - Recommend: Specific approach with reasoning
VERIFY:
  - Recommendation matches project constraints
  - Trade-offs clearly explained
  - Implementation path clear
```

---

### 🎨 Frontend-Engineer Agent (special case)

**When to Use**: ANY visual/styling/UI changes

**Pattern**:
```
TASK: Design CLI output for workspace status
WHO: frontend-ui-ux-engineer
HOW:
  MUST DO:
    - Create clear, scannable output
    - Use consistent visual language
    - Test readability at terminal width
    - Provide color/emoji guidance
  RETURN:
    - Mock output examples
    - CSS/styling code
    - ASCII art if needed
VERIFY:
  - Output is readable
  - Consistent with existing CLI style
```

---

### ✅ QA-Tester Agent (special case)

**When to Use**: After implementation for E2E verification

**Pattern**:
```
TASK: Test remote Docker workspace creation
WHO: qa-tester
HOW:
  - Use interactive CLI via tmux
  - Create workspace
  - Start services
  - Verify remote access
  - Check port mapping
VERIFY:
  - Workflow succeeds end-to-end
  - Error messages are clear
  - Performance acceptable
```

---

## Work Patterns

### Pattern 1: Parallel Execution (No Dependencies)

```
User Request: "Add user management (USR-01 through USR-05)"

PARALLEL:
  └─ document-writer: Create USR-01 spec doc
  └─ document-writer: Create USR-02 API spec
  └─ backend-dev: Implement user registry (USR-01)
  └─ backend-dev: Implement registration API (USR-02)
  └─ explore: Find existing user management patterns
  
THEN (after results):
  └─ Sisyphus: Review specs + code
  └─ backend-dev: Integrate & test
  └─ document-writer: Update docs/README.md
```

### Pattern 2: Sequential with Dependencies

```
User Request: "Implement coordination server"

STEP 1:
  └─ oracle: Architecture review & recommendation
  └─ document-writer: Create architecture doc
  └─ backend-dev: Create project structure

STEP 2 (depends on Step 1):
  └─ backend-dev: Implement core components
  └─ document-writer: Update README

STEP 3 (depends on Step 2):
  └─ backend-dev: Integration testing
  └─ qa-tester: E2E validation
  └─ Sisyphus: Final review & decision
```

### Pattern 3: Rapid Iteration

```
User Request: "Fix coordination server crashes"

STEP 1:
  └─ explore: Find crash locations
  └─ oracle: Analyze root causes (if unclear)

STEP 2:
  └─ backend-dev: Implement fixes (TDD)
  └─ backend-dev: Verify with tests

STEP 3:
  └─ backend-dev: Push to branch
  └─ Sisyphus: Code review & merge decision
```

---

## Sisyphus Responsibilities (Minimal)

### What Sisyphus DOES
1. **Understand** the requirement/request
2. **Plan** which agents to delegate to
3. **Create** delegation prompts (MUST DO/MUST NOT DO)
4. **Synthesize** results from multiple agents
5. **Make** final decisions (merge? proceed? pivot?)
6. **Track** progress via todos

### What Sisyphus DOES NOT DO
- ❌ Create documentation (delegate to document-writer)
- ❌ Write code (delegate to backend-dev)
- ❌ Search codebase (delegate to explore)
- ❌ Manually move files (delegate to backend-dev)
- ❌ Review code line-by-line (delegate to oracle if complex)

---

## Workflow Template for M3 Sprint Work

### Every Sprint

```
Monday (Planning):
  PARALLEL:
    └─ document-writer: Create sprint doc from template
    └─ backend-dev: Set up branch/environment
    └─ explore: Identify relevant code areas
  
  THEN:
    └─ Sisyphus: Review, adjust if needed, approve sprint start

Days 2-9 (Execution - MINIMAL Sisyphus):
  ASYNC:
    └─ backend-dev: Implement tasks (with TDD)
    └─ document-writer: Update docs as code changes
    └─ Sisyphus: Monitor todos, unblock if needed
  
  SYNC (optional):
    └─ Daily standups (team async updates)
    └─ Blocker resolution (Sisyphus mediation if needed)

Day 10 (Review & Retro):
  PARALLEL:
    └─ backend-dev: Demo work, run final tests
    └─ document-writer: Prepare sprint summary
    └─ qa-tester: E2E validation (if critical)
  
  THEN:
    └─ Sisyphus: Final review, merge decisions, lessons learned
    └─ Sisyphus: Plan Sprint N+1

```

---

## Example: Sprint 1 Coordination (Jan 20-Feb 2)

**Goal**: Coordination server foundation (33% → 60%)

### Monday Jan 20 (2 hours Sisyphus time)

```
PARALLEL:
  task(
    subagent_type: "document-writer",
    description: "Create Sprint 1 doc",
    prompt: """
    Create docs/sprints/active/sprint-01.md:
    - Dates: Jan 20 - Feb 2
    - Goal: Coordination server foundation
    - Tasks: COR-05, COR-06, AGT-03, INF-02, VFY-03
    - Success criteria from SPRINT_FRAMEWORK.md
    - Daily standup placeholders
    - Risk log section
    RETURN: Complete sprint document ready for team
    """
  )
  
  task(
    subagent_type: "sisyphus-junior-high",
    description: "Set up sprint infrastructure",
    prompt: """
    1. Create feature branch: git checkout -b sprint-01/coordination-server
    2. Create pkg/coordination/ directory structure
    3. Add TODO comments for COR-05, COR-06, etc
    4. Push branch
    RETURN: Branch ready for implementation
    """
  )
  
  task(
    subagent_type: "explore",
    description: "Map existing SSH/transport code",
    prompt: """
    Find all SSH-related code:
    - Current SSH implementation in QEMU provider
    - Connection pooling patterns
    - Key generation code
    - Transport layer candidates
    RETURN: File locations + pattern summary
    """
  )
```

**Sisyphus Review (30 min)**: 
- Verify sprint doc is complete
- Verify branch is ready
- Verify explore results make sense
- **Decision**: "Sprint 1 approved. Kickoff with team."

### Days 2-9 (MINIMAL Sisyphus)

```
backend-dev:
  ✓ Implementing COR-05, COR-06, AGT-03, etc
  ✓ TDD approach (tests first)
  ✓ Pushing to sprint-01 branch
  ✓ Daily standup async in sprint doc

document-writer:
  ✓ Updating docs as code emerges
  ✓ Keeping sprint doc fresh
  ✓ Creating implementation notes

Sisyphus:
  ✓ Check todos daily (5 min)
  ✓ Unblock if needed (emerge only)
  ✓ Monitor lsp_diagnostics results
  ✓ That's it!
```

### Day 10 (Final Review - 1 hour)

```
PARALLEL:
  backend-dev:
    ✓ Run full test suite
    ✓ Verify lsp_diagnostics clean
    ✓ Prepare code for review
  
  document-writer:
    ✓ Complete sprint retrospective doc
    ✓ Update SPRINT_FRAMEWORK.md with learnings
    ✓ Create Sprint 2 skeleton
  
  qa-tester:
    ✓ E2E test: Start coordination server
    ✓ Verify node registration works
    ✓ Check SSH pooling stability

Sisyphus:
  ✓ Review PRs (Oracle helps if complex)
  ✓ Merge approved PRs
  ✓ Review retro learnings
  ✓ Approve Sprint 2 plan
  ✓ Document decisions
```

---

## Success Metrics for This Workflow

| Metric | Target | How to Track |
|--------|--------|--------------|
| **Sisyphus Time/Sprint** | <5 hours | todos + timer |
| **Parallel Execution** | 70%+ | dependency graph |
| **Task Cycle Time** | <1 day | PR timestamps |
| **Agent Utilization** | 80%+ | background task logs |
| **Rework Rate** | <10% | revision requests |
| **Team Velocity** | 80%+ completion | sprint results |

---

## When to Escalate to Sisyphus

❌ **Don't escalate for**:
- Minor documentation updates
- Standard code changes
- Finding code patterns
- Creating new files

✅ **DO escalate for**:
- Architecture decisions (→ oracle)
- Conflicting requirements
- Unblock persistent blockers
- Major scope changes
- Sprint planning/retrospectives

---

## Sisyphus as Orchestrator, Not Executor

This workflow is **optimized for efficiency**:
- Parallelize everything possible
- Minimize Sisyphus context switches
- Maximize specialist autonomy
- Rapid feedback loops
- Minimal ceremony

**Result**: Sisyphus focuses on **direction**, not **execution**. 

---

**Workflow Established**: January 17, 2026
