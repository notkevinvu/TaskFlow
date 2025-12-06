---
description: Guided feature development with spec → plan → implement flow and automatic PR gates
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Workflow Orchestrator

This command guides you through the complete feature development lifecycle with appropriate review gates.

### Phase Detection

1. **Detect current state** by checking:

   a. **Current branch**:
      ```bash
      git branch --show-current
      ```
      - If on `main` or `master`: No feature in progress
      - If on `feature/*` or `[0-9]*-*`: Feature branch detected

   b. **Existing artifacts** (if on feature branch):
      - Check for `.specify/specs/BRANCH_NAME/spec.md`
      - Check for `.specify/specs/BRANCH_NAME/plan.md`
      - Check for `.specify/specs/BRANCH_NAME/tasks.md`
      - Check for `.specify/specs/BRANCH_NAME/checklists/`

   c. **Determine current phase**:
      | Condition | Current Phase | Next Action |
      |-----------|---------------|-------------|
      | On main, no feature description provided | START | Ask for feature description |
      | On main, feature description provided | SPECIFY | Run /specify |
      | spec.md exists, no plan.md | PLAN | Offer PR gate, then /plan |
      | plan.md exists, no tasks.md | TASKS | Run /tasks |
      | tasks.md exists, incomplete tasks | IMPLEMENT | Run /implement |
      | All tasks complete | COMPLETE | Offer final PR |

2. **Determine feature tier** (for PR gate decisions):

   Analyze the feature description and spec to determine complexity:

   | Tier | Criteria | PR Strategy |
   |------|----------|-------------|
   | **Tier 1** (Small) | Single component, < 5 tasks, no new APIs | Single PR with all artifacts |
   | **Tier 2** (Medium) | Multiple components, 5-15 tasks, new API endpoints | PR 1: Spec → PR 2: Plan + Implementation |
   | **Tier 3** (Large) | Cross-system, 15+ tasks, architectural changes | PR 1: Spec → PR 2: Plan → PR 3+: Implementation |

   Display the detected tier to the user:
   ```
   📊 Feature Tier: [TIER] ([Small/Medium/Large])
   └─ PR Strategy: [Description of PR approach]
   ```

### Workflow Execution

3. **Execute based on current phase**:

   #### Phase: START (No feature in progress)

   If user provided a feature description in $ARGUMENTS:
   - Proceed to run `/specify "$ARGUMENTS"`

   If no description provided:
   - Ask user: "What feature would you like to build? Provide a description:"
   - Wait for response, then run `/specify` with their description

   #### Phase: SPECIFY → PLAN Gate

   After spec.md is created (or if it already exists without plan.md):

   ```markdown
   ## ✅ Specification Complete

   **Feature:** [Feature name from spec]
   **Branch:** [Current branch name]
   **Spec:** `.specify/specs/[BRANCH]/spec.md`

   ### PR Gate Decision

   Based on **Tier [N]** complexity:

   | Option | Action |
   |--------|--------|
   | **A** | Open PR for spec review now, then continue to planning |
   | **B** | Skip spec PR, continue directly to planning |
   | **C** | Pause here - I'll review the spec manually first |

   **Recommended for Tier [N]:** [A or B based on tier]

   Your choice:
   ```

   - If user chooses A: Create PR with `gh pr create`, then run `/plan`
   - If user chooses B: Run `/plan` directly
   - If user chooses C: Stop and wait for user to return

   #### Phase: PLAN → TASKS

   After plan.md is created:

   For **Tier 3 only**, offer plan PR:
   ```markdown
   ## ✅ Technical Plan Complete

   **Plan:** `.specify/specs/[BRANCH]/plan.md`

   ### PR Gate (Tier 3 Feature)

   Large features benefit from architectural review before implementation.

   | Option | Action |
   |--------|--------|
   | **A** | Open PR for plan review, then generate tasks |
   | **B** | Skip plan PR, continue to task generation |

   Your choice:
   ```

   For **Tier 1-2**: Automatically proceed to `/tasks`

   #### Phase: TASKS → IMPLEMENT

   After tasks.md is created:

   ```markdown
   ## ✅ Task Breakdown Complete

   **Tasks:** `.specify/specs/[BRANCH]/tasks.md`
   **Total tasks:** [N]
   **Parallelizable:** [M] tasks marked [P]

   Ready to begin implementation?

   | Option | Action |
   |--------|--------|
   | **A** | Start implementation now |
   | **B** | Run /analyze first to validate consistency |
   | **C** | Pause - I want to review tasks first |

   Your choice:
   ```

   - If A: Run `/implement`
   - If B: Run `/analyze`, then `/implement`
   - If C: Stop and wait

   #### Phase: IMPLEMENT → COMPLETE

   After all tasks are marked complete in tasks.md:

   **First, check for database migrations:**
   ```bash
   ls backend/migrations/*.up.sql 2>/dev/null | wc -l
   ```

   If new migration files were created during this feature:
   ```markdown
   ## ⚠️ Database Migration Required

   This feature includes database schema changes:

   | Migration | File |
   |-----------|------|
   | 000005 | `backend/migrations/000005_subtasks_support.up.sql` |

   **You must apply these migrations before testing!**

   | Option | Action |
   |--------|--------|
   | **M** | Run `/migrate` to apply migrations now |
   | **S** | Skip - I'll apply migrations manually later |

   Your choice:
   ```

   - If M: Run the `/migrate` command
   - If S: Warn user that feature won't work until migrations are applied

   **Then show completion summary:**

   ```markdown
   ## 🎉 Implementation Complete!

   **Summary:**
   - [X] Specification defined
   - [X] Technical plan created
   - [X] [N] tasks completed
   - [X/⚠️] Database migrations [applied/pending]

   ### Final Steps

   | Option | Action |
   |--------|--------|
   | **A** | Run code review, then open PR |
   | **B** | Open PR directly (skip review) |
   | **C** | I'll handle the PR manually |

   Your choice:
   ```

   - If A: Use Task tool with code-reviewer agent, then create PR
   - If B: Create PR with `gh pr create`
   - If C: Stop

### PR Creation Format

When creating PRs, use this format:

```bash
gh pr create --title "[Feature] [Short description from spec]" --body "$(cat <<'EOF'
## Summary

[2-3 bullet points from spec.md success criteria]

## Changes

[List key files/components modified from tasks.md]

## Artifacts

- Spec: `.specify/specs/[BRANCH]/spec.md`
- Plan: `.specify/specs/[BRANCH]/plan.md`
- Tasks: `.specify/specs/[BRANCH]/tasks.md`

## Test Plan

[From spec.md acceptance criteria]

---
🤖 Generated with [Claude Code](https://claude.ai/claude-code) using speckit workflow
EOF
)"
```

### Status Display

At each phase, display current workflow status:

```
┌─────────────────────────────────────────────────────┐
│              SPECKIT WORKFLOW STATUS                │
├─────────────────────────────────────────────────────┤
│                                                     │
│  Feature: [Name]                                    │
│  Branch:  [feature/xxx]                             │
│  Tier:    [1/2/3] ([Small/Medium/Large])            │
│                                                     │
│  Progress:                                          │
│  [✓] Specify  →  [✓] Plan  →  [ ] Tasks  →  [ ] Implement  │
│       ↓              ↓                              │
│     PR #12        (skip)                            │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Resume Capability

If the user runs `/workflow` on an existing feature branch:
1. Detect current phase from artifacts
2. Display status showing completed phases
3. Offer to continue from current phase

This allows picking up work after breaks or session changes.

### Error Handling

- If `/specify` fails: Report error, suggest checking feature description
- If `/plan` fails: Report error, suggest running `/clarify` first
- If `/implement` fails mid-way: Report progress, offer to retry failed task
- If PR creation fails: Show gh error, suggest manual PR creation

### Notes

- This command orchestrates existing speckit commands - it doesn't replace them
- Users can still run individual commands (`/specify`, `/plan`, etc.) directly
- PR gates are recommendations - users can always skip them
- Tier detection is heuristic - users can override by specifying tier in arguments
  - Example: `/workflow --tier 3 "Add user authentication"`
