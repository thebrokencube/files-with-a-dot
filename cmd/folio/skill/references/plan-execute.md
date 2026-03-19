# Plan Workflow — Execute Phase (Phases 7-8)

Read by Agent 3 (Execute). Self-contained for execution sessions.

## Phase 7: Execute

Reads the committed work brief — no prior conversation context. If the brief has multiple
tracks, each track can be executed independently (same or separate sessions).

If you discover something unexpected during implementation that contradicts the brief, stop and consult the user rather than improvising.

For each track step, execute this sequence in order. Do NOT skip or reorder steps.

1. **Implement** one logical unit (one feature, fix, or refactor). If a step spans
   multiple concerns, split it at commit time.
2. **Content extraction check** (if this step moved or extracted content across files): Diff
   old content against new locations. Verify nothing was dropped, duplicated, or silently
   truncated. Do not rely solely on review agents — run an explicit before/after comparison.
3. **Validate — hard gate.** Run the validation commands from the brief's execution conventions
   in order: build, then test, then lint. If ANY command fails, STOP and fix. Do NOT proceed
   to step 4 until all validation commands pass.
   **Skill file drift check**: If this step changed CLI behavior, command names, schema fields,
   or flag semantics, grep the skill files (`~/.claude/skills/`) for stale references before
   proceeding. Skill docs that reference old field names or removed commands cause downstream
   confusion — fix them in the same commit or as a follow-up commit in the same track.
4. **Review — hard gate.** Launch 2 review agents (subagent_type: general-purpose) — one
   checking accuracy, one checking scope. Converge findings and fix issues. Do NOT run
   `git commit` until both reviews complete and all findings are resolved. If fixes are
   mechanical (typos, imports, paths), proceed to step 5. If fixes change logic or add code
   paths, return to step 3.
5. **Commit.** One logical unit per commit.
6. **Repeat** from step 1 for the next step.

### Implementation Review Prompt

Use with `subagent_type: "general-purpose"` and `model: "sonnet"`. Launch two instances — one accuracy, one scope. Needs file access.

```
You are reviewing code changes before a commit. Your job is to catch implementation bugs.

## Original Task
{task_description}

## Work Brief
{brief_summary}

## Changes to Review
{change_description}

## Your Focus: {focus}
{focus_description}

For each issue found, state: what's wrong, where, and a concrete fix.
Keep your review under 40 lines. Only flag real issues.
```

**Focus descriptions:**

- **Accuracy**: "Verify the changes match the work brief. Check for typos, stale references, incorrect file paths, broken imports, and wrong function signatures. Read the actual changed files."
- **Scope**: "Check that changes are necessary and sufficient. Flag anything that wasn't in the brief, unnecessary abstractions, or missing pieces. Meta-review: should any of these changes NOT exist? Verify the commit bundles one logical unit — flag if unrelated concerns are mixed."

**Folio integration**: If a relevant folio project exists, record design decisions, progress, and rationale in the folio project as work progresses — not as a final cleanup step. This means updating folio.yml observations, adding reference files for significant decisions, and keeping cross-references current throughout implementation. All `~/.folio` commits must use `folio home push` (see SKILL.md § Git Operations).

If execution reveals a design-level flaw, escalate to the user — see plan.md for re-run rules.

## Phase 8: Retrospective

**Mandatory — no skip.** Every agent session ends with a retro. The execution agent does the
most thorough retro since it has implementation context. Design and Brief agents do lightweight
retros focused on process friction and artifact quality. Do NOT delegate retros to a subagent.

Cover:
- Was the work brief sufficient? Could the execution agent proceed without reading the design doc?
- Were track boundaries well-chosen? Did tracks interfere or require unexpected coordination?
- Was the risk sequencing useful? Did high-risk-first ordering surface issues early?
- If the plan changed folio source files, flag whether targets need recompilation.

Only capture actionable findings, not session notes. Findings that aren't worth planning
aren't actionable — note them and move on.

**Recording findings**: Retro findings worth preserving MUST be materialized via
`folio new retro <topic>` — not just added as observation items. The retro file captures the
full context; observation items capture the actionable follow-ups. Both are written: the retro
file is the durable artifact, observation items are the work triggers. Commit via
`folio home push`. For lightweight retros (few findings), observation items alone suffice.

## Session Exit

Phase 8 retro is the exit sequence for the final agent. No forward handoff.
Record actionable findings via `folio new retro <topic>` and observation items.
Commit via `folio home push`.
