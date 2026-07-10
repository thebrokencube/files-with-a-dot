# Plan Workflow — Execute Phase (Phases 7-8)

Read by Agent 3 (Execute). Self-contained for execution sessions.

## Phase 7: Execute

Reads the committed work plan — no prior conversation context. If the plan has multiple
tracks, each track can be executed independently (same or separate sessions).

If you discover something unexpected during implementation that contradicts the plan, stop and consult the user rather than improvising.

For each track, execute this sequence in order. Do NOT skip or reorder steps.

### Shape Dispatch

Before starting Step 0, check the work plan's `shape:` field:
- `shape: burndown` → Read references/burndown.md. It replaces the per-track
  step sequence (Steps 0-7 within Phase 7) with a wave-based batch loop.
  Validation and review gates still apply per-group. Phase 8 (Retro) still applies.
- No shape field → Continue with Steps 0-7 below.

**Lightweight (N==1) plans** have no design doc or brief — the committed plan is the frozen
idea/arch sketch + a single `track-1.md` at `work/active/<date>-<topic>/track-1.md`, authored at
sketch-freeze with execution-level detail for the one track (the author who froze the sketch writes
it — no Brief agent runs). Execute it as a single track. With design/brief gates absent, the
per-commit review + adversarial verify (Step 5 gate below) are the ONLY quality gates and are
mandatory. If the work fans out mid-execution, escalate per the layered-escalation table.

0. **Spike** — before touching any code for a track:
   1. Read the track spec (file paths, interfaces, constraints, deferral markers)
   2. Read every file the track will modify — full contents
   3. For each `[RESOLVE IN SPIKE]` marker: make the implementation choice, grounded in
      actual code
   4. Draft implementation sequence: what changes in what order
   5. **Conflict check**: For each file path in spec — does it exist? For each type
      signature — does it match? For each CLI op referenced — does the command/subcommand/flag
      exist (read source or `--help`)? For each scope assumption — still valid?
   6. If conflicts found: classify per the layered escalation table below and escalate.
      If clean: proceed to Step 1.

   Spike output is internal to the execution agent (not a separate committed artifact).
   If the track has no `[RESOLVE IN SPIKE]` markers and all paths check out, the spike
   is fast — just confirm and proceed.

For each track step after the spike, execute:

1. **Implement** one logical unit (one feature, fix, or refactor). If a step spans
   multiple concerns, split it at commit time.
2. **Test gate (Go repos only).** If the repo uses Go (`go.mod` exists or Makefile contains
   `go test`), write or update tests before production code. RED-GREEN-REFACTOR:
   - Write a failing test for the expected behavior
   - Run `make test` — verify it fails with the expected error
   - Write minimal code to pass
   - Run `make test` — verify it passes
   - Refactor if needed, re-run tests
   For skill-file-only changes (no Go code), note "no automated tests for skill files" and
   proceed to the next step.
3. **Content extraction check** (if this step moved or extracted content across files): Diff
   old content against new locations. Verify nothing was dropped, duplicated, or silently
   truncated. Do not rely solely on review agents — run an explicit before/after comparison.
4. **Validate — hard gate.** Run the validation commands from the plan's execution conventions
   in order: build, then test, then lint. If ANY command fails, STOP and fix. Do NOT proceed
   to step 5 until all validation commands pass.
   **Skill file drift check**: If this step changed CLI behavior, command names, schema fields,
   or flag semantics, grep the skill files (`~/.claude/skills/`) for stale references before
   proceeding. Skill docs that reference old field names or removed commands cause downstream
   confusion — fix them in the same commit or as a follow-up commit in the same track.
5. **Review — hard gate.** Launch 1 review agent (subagent_type: general-purpose,
   model: "opus") covering accuracy, scope, and code quality. If the review returns issues,
   fix them and re-dispatch the review agent. Loop until zero issues. Cap at 5 iterations —
   if still failing, escalate to the user. Do NOT run `git commit` until the review passes
   clean.
6. **Commit checklist (ALL must pass):**
   - [ ] Tests pass (step 2)
   - [ ] Validation commands pass (step 4)
   - [ ] Review returned zero issues (step 5, cap 5 iterations)
   - [ ] Content extraction checked (step 3, if applicable)
   One logical unit per commit.
7. **Repeat** from step 1 for the next step.

### Implementation Review Prompt

Use with `subagent_type: "general-purpose"` and `model: "opus"`. Launch one instance with comprehensive checklist. Needs file access.

```
You are reviewing code changes before a commit. Your job is to catch implementation bugs.

## Original Task
{task_description}

## Work Brief
{brief_summary}

## Changes to Review
{change_description}

## Prior Issues (verify these are resolved)
{If round > 1: list of issues from previous review round}
{If round 1: "This is the first review round. No prior issues."}

## Review Checklist
1. **Accuracy**: Verify changes match the work brief. Check for typos, stale references,
   incorrect file paths, broken imports, wrong function signatures. Read the actual files.
2. **Scope**: Check changes are necessary and sufficient. Flag anything not in the brief,
   unnecessary abstractions, or missing pieces. Verify the commit bundles one logical unit.
3. **Code quality**: Check for bugs, edge cases, error handling gaps, and style issues.
   Flag anything that would fail review on a real PR.
4. **Adversarial check**: Challenge the implementation — is there a simpler way to achieve
   the same result? Are any abstractions unnecessary? Would a future reader understand why
   this approach was chosen over alternatives? (See `references/adversarial-review.md`.)
Do not flag issues that linters, formatters, or test suites would catch — those are handled by deterministic tools.

Report in two sections:
1. **Prior fix verification**: For each prior issue, state RESOLVED or STILL PRESENT with evidence.
2. **New findings**: Issues not in the prior list. For each: what's wrong, where, and a concrete fix.
Keep your review under 40 lines. Only flag real issues.
```

**Folio integration**: If a relevant folio project exists, record design decisions, progress, and rationale in the folio project as work progresses — not as a final cleanup step. This means updating folio.yml observations, adding reference files for significant decisions, and keeping cross-references current throughout implementation. All `~/.folio` commits must use `folio home push` (see SKILL.md § Git Operations).

### Layered Escalation

When the spike or implementation discovers a flaw, classify it by layer and act accordingly:

| Type | Trigger | Blast radius | Action |
|------|---------|-------------|--------|
| Direction flaw | Goal wrong, scope boundary misdrawn, constraint infeasible | Stop ALL tracks | User re-evaluates direction. Amendment via plan.md's amend-design path. |
| Interface flaw | File path doesn't exist, signature mismatch, cross-track dependency broken | Stop AFFECTED track | User patches interface spec. Other tracks continue. |
| Implementation question | Technique choice, code structure within single file | No stop | Agent decides (single-file consequence) or asks user (cross-file/user-visible). |

**Backward traversal**: When a flaw is discovered, record it as an observation. Classify as
**additive** (doesn't invalidate existing decisions) or **contradictory** (invalidates).
Additive: edit the design doc at the appropriate layer → re-run Phase 4 review only → commit.
Contradictory: re-run from Phase 2 (direction) or Phase 5 (brief) per plan.md re-run rules.

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

## Exit Criteria Templates

### Brief → Execute
- [ ] All commit checklist items passed for every commit
- [ ] Work plan's validation commands pass end-to-end
- [ ] No unresolved escalations to user

### Execute → Done
- [ ] All tracks completed with passing validation
- [ ] Retro materialized (`folio new retro` or observation items)
- [ ] Relevant observations resolved (`folio observe resolve`)
- [ ] Open work captured as observations or design-doc Open Questions if work continues

## Session Exit

Phase 8 retro is the exit sequence for the final agent. No separate handoff file —
folio state IS the handoff. Record actionable findings via `folio new retro <topic>` and
observation items. If work continues in another session, update the design doc's
Open Questions + Convergence Status and log follow-on items via `folio observe`. Commit
via `folio home push`. The next session resumes by running `/folio <project>` — see
SKILL.md "Bare Invocation" for the resume flow.

**Validation (mandatory if work continues)**: Spawn a fresh subagent with ONLY the design
doc path + work plan path + observations list (no conversation context). Ask whether the
state is sufficient to resume. Fix anything missing in the design doc or observations
before session ends.
