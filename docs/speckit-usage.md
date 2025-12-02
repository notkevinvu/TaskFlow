# Speckit Feature Development Workflow

Speckit is an AI-assisted feature development framework that provides structured commands for planning, specifying, and implementing features in TaskFlow.

## Quick Start

### Option 1: Guided Workflow (Recommended)

```bash
# Single command guides you through the entire flow with PR gates
/workflow "Add user notification preferences"
```

The `/workflow` command will:
1. Detect your current phase (or start fresh)
2. Guide you through spec → plan → tasks → implement
3. Offer PR creation at appropriate gates based on feature complexity
4. Support resuming if you take a break

### Option 2: Manual Commands

```bash
# 1. Start with a feature idea
/specify "Add user notification preferences"

# 2. Clarify any ambiguities (optional)
/clarify

# 3. Create implementation plan
/plan

# 4. Generate task breakdown
/tasks

# 5. Execute implementation
/implement
```

## Commands Reference

### `/workflow "feature description"` (Recommended)

Orchestrates the complete feature development lifecycle with intelligent phase detection and PR gates.

**What it does:**
- Detects current workflow phase based on existing artifacts
- Guides through spec → plan → tasks → implement with prompts
- Offers PR creation at appropriate gates based on feature tier
- Supports resuming work after breaks

**Feature Tiers:**
| Tier | Criteria | PR Strategy |
|------|----------|-------------|
| Tier 1 (Small) | < 5 tasks, single component | Single PR with all artifacts |
| Tier 2 (Medium) | 5-15 tasks, new APIs | Spec PR → Plan + Implementation PR |
| Tier 3 (Large) | 15+ tasks, architectural | Spec PR → Plan PR → Implementation PR(s) |

**Example:**
```bash
/workflow "Allow users to configure email notification frequency"

# Or override tier detection:
/workflow --tier 3 "Add multi-tenant workspace support"
```

**Resume capability:**
```bash
# If you're already on a feature branch, /workflow detects your progress:
/workflow
# Output: "Detected feature branch 'feature/notifications' at PLAN phase. Continue?"
```

---

### `/specify "feature description"`

Creates a feature specification from a natural language description.

**What it does:**
- Creates a new feature branch (e.g., `feature/001-notification-preferences`)
- Generates `spec.md` with user scenarios, requirements, and success criteria
- Creates a quality checklist for requirement validation

**Output location:** `.specify/specs/BRANCH_NAME/spec.md`

**Example:**
```
/specify "Allow users to configure email notification frequency for at-risk tasks"
```

---

### `/clarify`

Identifies underspecified areas in the current feature spec by asking targeted clarification questions.

**What it does:**
- Analyzes the spec for ambiguities
- Asks up to 5 clarification questions
- Encodes answers back into the spec

**When to use:** After `/specify` if you want to refine requirements before planning.

---

### `/plan`

Generates an implementation plan based on the feature specification.

**What it does:**
- **Phase 0:** Research existing codebase patterns
- **Phase 1:** Design data models and API contracts
- Creates `plan.md`, `research.md`, `data-model.md`
- Validates against the project constitution

**Output location:** `.specify/specs/BRANCH_NAME/plan.md`

**Constitution check:** The plan must comply with TaskFlow's constitution principles:
- Clean Architecture layers
- Type safety requirements
- Testing standards
- Design system consistency
- Simplicity (YAGNI)

---

### `/tasks`

Breaks down the implementation plan into actionable, prioritized tasks.

**What it does:**
- Creates `tasks.md` organized by user story
- Assigns task IDs (T001, T002, etc.)
- Marks parallelizable tasks with `[P]`
- Groups by phase (Setup → Foundational → User Stories → Polish)

**Output location:** `.specify/specs/BRANCH_NAME/tasks.md`

**Task format:**
```markdown
- [ ] [T001] [P] [US1] Create notification preferences table in database
- [ ] [T002] [US1] Add NotificationPreference domain model
- [ ] [T003] [P] [US1] Implement notification preference repository
```

---

### `/checklist`

Generates a custom checklist for validating feature requirements.

**What it does:**
- Creates checklists that act as "unit tests for requirements"
- Validates completeness, clarity, and consistency
- NOT for verification testing - for requirements quality

**Output location:** `.specify/specs/BRANCH_NAME/checklists/`

---

### `/analyze`

Performs read-only consistency analysis across all feature artifacts.

**What it does:**
- Validates consistency between `spec.md`, `plan.md`, and `tasks.md`
- Checks constitution compliance
- Reports gaps or conflicts
- Does NOT modify any files

**When to use:** Before starting implementation to catch issues early.

---

### `/implement`

Executes the implementation workflow based on the task breakdown.

**What it does:**
- Verifies prerequisites (spec, plan, tasks exist)
- Checks checklist status
- Guides incremental implementation
- Ensures each user story is independently deployable

---

### `/constitution`

Creates or updates the project constitution.

**What it does:**
- Interactive principle definition
- Syncs dependent templates
- Version tracking

**Location:** `.specify/memory/constitution.md`

---

### `/taskstoissues`

Converts tasks into GitHub issues for project tracking.

**What it does:**
- Creates GitHub issues from `tasks.md`
- Preserves task IDs and dependencies
- Links to feature specification

---

## Workflow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    FEATURE DEVELOPMENT                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. SPECIFY ──► 2. CLARIFY ──► 3. PLAN ──► 4. TASKS         │
│      │              │              │            │            │
│      ▼              ▼              ▼            ▼            │
│   spec.md      (refinements)   plan.md     tasks.md         │
│                                research.md                   │
│                                data-model.md                 │
│                                                              │
│  5. ANALYZE ──► 6. IMPLEMENT                                │
│      │              │                                        │
│      ▼              ▼                                        │
│   (validation)   Working Code                               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
.specify/
├── memory/
│   └── constitution.md      # Project governance principles
├── scripts/
│   ├── check-prerequisites.ps1
│   ├── common.ps1
│   ├── create-new-feature.ps1
│   ├── setup-plan.ps1
│   └── update-agent-context.ps1
├── specs/                   # Feature specifications (per branch)
│   └── feature-001-name/
│       ├── spec.md
│       ├── plan.md
│       ├── tasks.md
│       ├── research.md
│       ├── data-model.md
│       └── checklists/
└── templates/
    ├── agent-file-template.md
    ├── checklist-template.md
    ├── plan-template.md
    ├── spec-template.md
    └── tasks-template.md
```

## Best Practices

### 1. Start with Clear Requirements

The better your initial `/specify` description, the better the outputs:

**Good:**
```
/specify "Add email notification preferences allowing users to choose frequency
(immediate, daily digest, weekly) for at-risk task alerts and task completion
reminders"
```

**Less effective:**
```
/specify "Add notifications"
```

### 2. Use `/clarify` for Complex Features

For features touching multiple systems or with unclear scope, run `/clarify` before `/plan`.

### 3. Review Before Implementation

Always run `/analyze` before `/implement` to catch:
- Spec/plan mismatches
- Missing tasks for requirements
- Constitution violations

### 4. Incremental Delivery

Each user story in `/tasks` should be:
- Independently testable
- Independently deployable
- Independently demonstrable

### 5. Keep Constitution Updated

If project patterns evolve, update the constitution using `/constitution` to keep future features aligned.

## Integration with TaskFlow

The speckit workflow integrates with TaskFlow's existing conventions:

| Speckit Artifact | TaskFlow Convention |
|------------------|---------------------|
| `spec.md` requirements | PRD user stories format |
| `plan.md` architecture | Clean Architecture layers |
| `tasks.md` breakdown | Handler → Service → Repository pattern |
| Quality checklist | Testing requirements from constitution |

## Troubleshooting

### "Prerequisite check failed"

Run `/specify` first to create the feature branch and specification.

### "Constitution violation detected"

Review `.specify/memory/constitution.md` and ensure your plan complies with the five core principles.

### "No tasks generated"

Ensure `/plan` completed successfully before running `/tasks`.

---

**See also:**
- `.specify/memory/constitution.md` - Project governance principles
- `.claude/CLAUDE.md` - Development conventions
- `docs/product/PRD.md` - Product requirements
