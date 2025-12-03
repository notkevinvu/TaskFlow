---
description: Kickstart informal development tasks with automatic branch setup, planning, PR creation, and review cycles
trigger: kickstart
---

## User Input

```text
$ARGUMENTS
```

You **MUST** have a task description in $ARGUMENTS to proceed. If empty, ask the user: "What task would you like to kickstart? Provide a brief description:"

---

## Kickstart Workflow

This command automates informal development tasks (outside the formal speckit flow) with built-in quality gates.

### Phase Quick Reference

| Phase | Name | Key Actions | Gate |
|-------|------|-------------|------|
| 1 | Branch Setup | Sync main, create branch | Branch created |
| 2 | Plan Discovery | Find/create plan | Plan confirmed |
| 3 | Implementation | Execute plan, commit code | Code pushed |
| 4 | PR Creation | Create PR, initial review | PR created + reviewed |
| 5 | Review Fixes | Fix CRITICAL/MAJOR issues | All major issues fixed |
| 6 | CI/CD | Wait for checks, fix failures | CI passing |
| 7 | Final Review | Re-run review, verify clean | No new major issues |
| 8 | Summary | Document completion | Summary provided |

**INSTRUCTION:** Complete each phase in order. Do NOT skip phases. Mark each gate as complete before proceeding.

### Phase 1: Branch Setup

**INSTRUCTION:** Execute these steps in order, then display the status box.

1. **Sync main branch**:
   ```bash
   git checkout main && git pull origin main
   ```

2. **Create feature branch**:
   - Derive a branch name from the task description
   - Format: `feature/<short-kebab-case-name>` or `fix/<short-kebab-case-name>`
   - Example: "Add date range picker" → `feature/date-range-picker`
   ```bash
   git checkout -b feature/<derived-name>
   ```

3. **Display branch status** (GATE CHECK):
   ```
   ┌─────────────────────────────────────────────────────┐
   │              KICKSTART: BRANCH READY                │
   ├─────────────────────────────────────────────────────┤
   │  Task:   [Description from $ARGUMENTS]              │
   │  Branch: [Created branch name]                      │
   │  Base:   main (synced)                              │
   │  Gate:   ✅ PHASE 1 COMPLETE                        │
   └─────────────────────────────────────────────────────┘
   ```

---

### Phase 2: Plan Discovery

**INSTRUCTION:** Search for existing plans first. If none found, create one. Get user confirmation before proceeding.

4. **Search for existing plan artifacts**:

   First, try demongrep (preferred for token efficiency):
   ```bash
   demongrep search "[task keywords]"
   ```

   Look for:
   - `docs/plans/*.md` or `docs/plan-*.md`
   - `.specify/specs/**/plan.md`
   - Any markdown file with "plan" in the name related to the task
   - PRD or design docs that outline the task

   Fallback to native tools if demongrep returns no results:
   ```bash
   # Use Glob/Grep tools
   ```

5. **Plan status decision** (GATE CHECK):

   #### If plan/doc EXISTS:
   - Read and analyze the plan document
   - Display: "Found existing plan: `[path]`"
   - Summarize key points from the plan
   - Display: "✅ PHASE 2 COMPLETE - Plan confirmed"
   - Proceed to Phase 3

   #### If NO plan exists:
   - **Enable extended thinking** for thorough planning
   - Analyze the task requirements deeply
   - Consider:
     - What files need to be created/modified?
     - What dependencies exist?
     - What's the logical order of operations?
     - What edge cases should be handled?

   - **Ask clarifying questions** if needed (max 3-4 questions)
   - Create an inline implementation plan (not saved to file unless requested)
   - Display the plan for user confirmation before proceeding
   - Display: "✅ PHASE 2 COMPLETE - Plan created and confirmed"

---

### Phase 3: Implementation

**INSTRUCTION:** Use TodoWrite to track tasks. Complete implementation before pushing. Verify build/tests pass.

6. **Execute the plan**:
   - Use TodoWrite tool to track implementation tasks
   - Follow the plan step-by-step
   - Make atomic, focused changes
   - Run tests if applicable after each significant change

7. **Push initial changes** (GATE CHECK):
   ```bash
   git add -A && git commit -m "[descriptive message]" && git push -u origin [branch]
   ```
   - Display: "✅ PHASE 3 COMPLETE - Code pushed to origin/[branch]"

---

### Phase 4: PR Creation & Initial Review

**INSTRUCTION:** Create PR, then run review. Post review as PR comment. Parse and categorize all findings.

8. **Create Pull Request**:
   ```bash
   gh pr create --title "[Type] [Short description]" --body "$(cat <<'EOF'
   ## Summary
   [2-3 bullet points describing the changes]

   ## Test Plan
   [How to verify the changes work]

   ---
   🤖 Generated with [Claude Code](https://claude.ai/claude-code) via /kickstart
   EOF
   )"
   ```

9. **Execute PR review** (GATE CHECK):
   - Run: `/pr-review-toolkit:review-pr`
   - Wait for review to complete
   - **Post the review as a PR comment** for tracking purposes
   - Display: "✅ PHASE 4 COMPLETE - PR #[number] created and reviewed"

---

### Phase 5: Review Implementation (CRITICAL GATE)

**INSTRUCTION:** ALL Critical/Major issues MUST be fixed. Use TodoWrite to track each fix. Do NOT proceed until all major issues resolved.

10. **Analyze PR review results**:

    Parse the review and categorize findings:
    | Severity | Action Required |
    |----------|-----------------|
    | **CRITICAL** | MUST fix - blocks merge |
    | **MAJOR** | MUST fix - blocks merge |
    | **MINOR** | SHOULD fix - recommended |
    | **SUGGESTION** | MAY fix - optional |

11. **Implement required fixes**:

    **CRITICAL REQUIREMENT**:
    - ALL Critical and Major issues MUST be addressed
    - If any Critical/Major issues remain unaddressed, the PR is **NON-MERGEABLE**

    For Minor issues:
    - Lean towards implementing these as well
    - Skip only if time-constrained or if the fix introduces unnecessary complexity

    Track fixes using TodoWrite tool.

12. **Push review fixes** (GATE CHECK):
    ```bash
    git add -A && git commit -m "fix: Address PR review feedback" && git push
    ```
    - Display: "✅ PHASE 5 COMPLETE - All CRITICAL/MAJOR issues resolved"

---

### Phase 6: CI/CD Verification

**INSTRUCTION:** Wait for CI to complete. If failures occur, fix them immediately. Do NOT proceed until CI passes.

13. **Wait for CI/CD checks**:
    ```bash
    gh pr checks [PR_NUMBER] --watch
    ```

    Or poll status:
    ```bash
    gh pr checks [PR_NUMBER]
    ```

14. **Handle CI/CD failures** (GATE CHECK):

    If checks fail:
    - Retrieve failure logs: `gh run view [RUN_ID] --log-failed`
    - Analyze and fix the issues
    - Push fixes and repeat until CI passes

    If checks pass:
    - Display: "✅ PHASE 6 COMPLETE - CI/CD passing"
    - Proceed to Phase 7

---

### Phase 7: Final Verification

**INSTRUCTION:** Run a fresh PR review. Verify no new issues. Post final review as PR comment. Display merge checklist.

15. **Execute final PR review**:
    - Run: `/pr-review-toolkit:review-pr` again
    - Verify no new Critical/Major issues introduced
    - **Post the final review as a PR comment**

16. **Confirm merge readiness** (GATE CHECK):

    ```
    ┌─────────────────────────────────────────────────────┐
    │              KICKSTART: MERGE CHECKLIST             │
    ├─────────────────────────────────────────────────────┤
    │  [✓] All Critical issues resolved                   │
    │  [✓] All Major issues resolved                      │
    │  [✓] CI/CD checks passing                           │
    │  [✓] Final review completed                         │
    │  [✓] Summary provided below                         │
    │  Gate: ✅ PHASE 7 COMPLETE - Ready for merge        │
    └─────────────────────────────────────────────────────┘
    ```

---

### Phase 8: Session Summary

**INSTRUCTION:** Provide a comprehensive summary. Include all sections below. This marks the end of the kickstart workflow.

17. **Provide completion summary** (GATE CHECK):

    ```markdown
    ## Kickstart Summary

    ### Problem Statement
    [1-2 lines outlining what problem or task was targeted/solved]

    ### Context
    - [Relevant detail 1]
    - [Relevant detail 2]
    - [Relevant detail 3]
    - [Relevant detail 4 - if applicable]

    ### Artifacts/Docs Used
    - [Path to plan doc or "Inline plan created"]
    - [Any other referenced documents]

    ### Solution
    - [Change 1]
    - [Change 2]
    - [Change 3]
    - [Change 4]
    - [Change 5 - if applicable]

    ### PR Status
    - **PR:** #[NUMBER] - [URL]
    - **Branch:** [branch-name]
    - **Status:** Ready for human review and merge
    ```

    Display: "✅ PHASE 8 COMPLETE - Kickstart workflow finished"

---

## Error Handling

- **Git conflicts**: Stop and notify user, suggest resolution steps
- **demongrep unavailable**: Fallback to Glob/Grep tools
- **PR creation fails**: Show error, suggest manual creation via GitHub UI
- **CI/CD timeout**: Report status, suggest checking GitHub Actions directly
- **Review agent unavailable**: Proceed with manual review checklist

---

## Notes

- This command is for **informal development** - for formal features, use the speckit workflow (`/workflow`, `/specify`, `/plan`, etc.)
- The PR review gate is **non-negotiable** for Critical/Major issues
- All PR reviews should be posted as comments for audit trail
- Extended thinking is used for planning when no existing plan doc exists
