# Plan Workflow

Read by `/folio plan [topic]`. Multi-agent diverge-converge planning for non-trivial tasks.

## When to Use

Use `/folio plan` instead of `EnterPlanMode` for any non-trivial task — multi-file changes, architectural decisions, unclear requirements, or multiple valid approaches.

**Skip planning entirely** for trivial changes: single-file fixes, obvious bugs, one-line tweaks. Just do them.

## Invocation

- `/folio plan` — Infer topic from context, default lenses (pragmatic vs thorough)
- `/folio plan <topic>` — Explicit topic, default lenses

Custom lenses are specified naturally in the topic text (e.g., `/folio plan redesign auth, considering performance and readability`). The agent parses the user's intent and crafts lens descriptions accordingly. Default lenses (pragmatic vs thorough) apply when no lens guidance is given.

## Phases

Phases 1-4 run in normal mode (full tool access). Phase 5 enters built-in plan mode for structured approval. Phase 6 implements the approved plan. Phase 7 runs post-implementation.

### Phase 1: Understand

Gather context before spawning any agents. This happens in the main conversation.

1. Identify the task scope from the user's request (or `<topic>` argument)
2. Read relevant files — entry points, existing implementations, tests, CLAUDE.md
3. Search for related patterns in the codebase (Grep/Glob)
4. **Check for folio context**: Run `folio home list` to find active projects. If any project is relevant to the task, read its `folio.yml` and pull in relevant sources, cross-references, and tasks/pending items.
5. **Present folio findings to the user.** If a project matched: summarize project name, key
   sources pulled in, relevant pending tasks. Wait for confirmation before continuing. If no
   project matched: list the active projects that were considered, ask if any of them are
   relevant (the user may see a match the search missed), and ask if a new folio project
   should be created to track this work. Only proceed without folio context after explicit
   user confirmation.
6. **Pin hard constraints**: Separate the user's stated decisions and explicit preferences
   (hard constraints) from open trade-offs. Hard constraints are non-negotiable — lenses
   must not re-evaluate them. Include pinned constraints as a distinct section at the top
   of the context summary.
7. Compile a **context summary** (max 30 lines): pinned hard constraints first, then what
   exists, what needs to change, key trade-offs, and relevant folio context (if any)

This summary is passed to all downstream agents. Diversity comes from lenses, not information asymmetry.

### Phase 2: Propose

Launch 2 Plan agents in parallel, each with the same context summary but a different lens. Each returns a proposal (max 80 lines).

Default lenses:
- **Pragmatic**: Minimize changes, reuse existing code, prefer the simplest approach that works
- **Thorough**: Consider edge cases, maintainability, architectural fit, future extensibility

### Phase 3: Converge

Launch 1 agent (subagent_type: general-purpose) to merge the two proposals into a single plan (max 100 lines).

Convergence criteria:
- Every file to be changed is listed with what changes and why
- Implementation order is specified
- Trade-offs between the two proposals are noted where they diverged meaningfully
- If both proposals agreed on an approach, that's a strong signal — keep it

After the converge agent returns, briefly summarize (3–5 lines) the key divergence decisions to the user — which proposal won on each point and why. Informational only, not blocking. Proceed to Phase 4 immediately after.

### Phase 4: Review

Launch 1 agent (subagent_type: general-purpose) to review the merged plan. The review covers three checks:

1. **Accuracy**: Does the plan reference correct file paths, function names, and APIs? Are assumptions about existing code valid?
2. **Feasibility**: Can each step actually be implemented as described? Are there missing dependencies or ordering issues?
3. **Scope**: Is everything in the plan necessary for the task? **Meta-review: should any of this work not exist?** Flag anything that's over-engineered, gold-plated, or solving a problem the user didn't ask about.

Review output: max 40 lines. For each issue found, state: what's wrong, where, and a suggested fix.

### Phase 5: Present (enter plan mode)

After Phases 1-4 complete, hand off to built-in plan mode for structured approval:

1. Call `EnterPlanMode`. The user will be prompted to consent to entering plan mode.
2. Once in plan mode, write the final plan to the **plan file path provided in the plan mode system message**. Include:
   - The plan (with any review fixes applied)
   - A summary of what the review flagged and how it was addressed
   - A brief note on where the two proposals differed and which was chosen
   - **The implementation order must end with an explicit step for Phase 7 (retrospective).** This is not implicit — it must appear as a numbered step so it cannot be skipped during execution.
   - **Execution conventions**: Project-specific implementation idioms — commit workflow
     (e.g., conventional commits, `folio home push`), tool flags (e.g., `rm -f` not `rm`),
     and repo-specific patterns discovered in Phase 1.
3. Call `ExitPlanMode` with `allowedPrompts` populated from the plan — e.g., if the plan includes running tests, include `{"tool": "Bash", "prompt": "run tests"}`. This presents the plan to the user for approval.

The user sees the plan file cleanly. They may request changes, ask questions, or reject. Iterate until approved or abandoned.

**Note**: Full conversation context (Phases 1-4) is retained in plan mode — only tool access is restricted (read-only + no Agent). This is fine since the heavy lifting is done.

### Phase 6: Implement

Only after user approval (ExitPlanMode accepted). Execute the plan step by step, following the specified order. If you discover something unexpected during implementation that contradicts the plan, stop and consult the user rather than improvising.

**Review before each commit**: Before every commit, launch 2 review agents (subagent_type: general-purpose) — one checking accuracy, one checking scope — then converge findings and fix issues before committing. This catches implementation bugs that planning can't — typos, stale references, broken cross-references.

**Content extraction check**: When a step moves or extracts content across files, diff the
old content against the new locations before committing. Verify nothing was dropped,
duplicated, or silently truncated. Do not rely solely on review agents — run an explicit
before/after comparison.

**Folio integration**: If a relevant folio project exists, record design decisions, progress, and rationale in the folio project as work progresses — not as a final cleanup step. This means updating folio.yml tasks/pending, adding reference files for significant decisions, and keeping cross-references current throughout implementation. All `~/.folio` commits must use `folio home push` (see SKILL.md § Git Operations).

### Phase 7: Retrospective

**Mandatory — no skip.** After all implementation commits are complete, review the planning
process in the main conversation (do NOT delegate to a subagent — the retrospective needs
full session context to be useful). Cover:

- What worked well? What added friction?
- Were the lenses useful? Did convergence surface good trade-offs?
- Was the folio context helpful (if used)?
- What should change next time?
- If the plan changed folio source files, flag whether targets need recompilation.

Only capture actionable findings, not session notes. Findings that aren't worth planning
aren't actionable — note them in the retro summary and move on.

**Recording findings**: For each actionable finding, invoke `/folio plan <finding>` to vet
and implement it as a durable change. This is not optional — verbal agreements evaporate.

**Recursion guard**: A `/folio plan` invoked from Phase 7 runs Phases 1-6 normally but
replaces its own Phase 7 with a single-sentence summary ("Retrospective finding recorded,
no further action needed."). Only one level of nesting is permitted.

## Re-run Rule

If plan feedback from the user requires rethinking (not just minor edits), re-run phases 2-4 (full diverge-converge). Do not patch the existing plan.

## Agent Prompts

### Propose Agent (Phase 2)

Use with `subagent_type: "Plan"`. Launch two instances in parallel with different lens values.

```
You are planning an implementation for: {task_description}

## Context
{context_summary}

## Your Lens: {lens_name}
{lens_description}

## Instructions
Propose an implementation plan. Your plan should:
- List every file to create or modify, with a description of the changes
- Specify implementation order and dependencies between steps
- Note any risks, assumptions, or open questions
- Stay within your lens — let it guide your trade-off decisions

Keep your proposal under 80 lines. Be concrete — file paths, function names, specific changes. No hand-waving.
```

**Default lens descriptions:**

- **Pragmatic**: "Minimize the number of files changed and lines of code written. Reuse existing patterns and utilities. Prefer the simplest correct solution. Avoid new abstractions unless they pay for themselves immediately."
- **Thorough**: "Consider edge cases, error handling, and how this change interacts with the rest of the codebase. Prioritize maintainability and architectural consistency. Flag potential issues even if fixing them adds scope."

### Converge Agent (Phase 3)

Use with `subagent_type: "general-purpose"`.

```
You are merging two implementation proposals into a single plan.

## Original Task
{task_description}

## Proposal A (Pragmatic)
{proposal_a}

## Proposal B (Thorough)
{proposal_b}

## Instructions
Merge the two proposals into a single implementation plan:
- For each aspect, pick the stronger approach — or synthesize from both
- Note where the proposals diverged and which was chosen, with rationale
- If both agreed on an approach, that's a strong signal — keep it
- Every file to change must be listed with what changes and why
- Specify implementation order

Keep the merged plan under 100 lines. Be concrete.
```

### Review Agent (Phase 4)

Use with `subagent_type: "general-purpose"`. Needs file access to verify claims.

```
You are reviewing an implementation plan. Your job is to find problems before code is written.

## Original Task
{task_description}

## Plan to Review
{plan_content}

## Review Checklist
1. **Accuracy**: Verify file paths, function names, and API references exist and are correct. Read the actual files.
2. **Feasibility**: Can each step be implemented as described? Are there missing imports, wrong method signatures, or ordering issues?
3. **Scope**: Is everything necessary? Meta-review: should any part of this plan NOT exist? Flag over-engineering, unnecessary abstractions, or work the user didn't ask for.

For each issue found, state: what's wrong, where in the plan, and a concrete fix.
Keep your review under 40 lines. Only flag real issues — don't nitpick style or add suggestions beyond the task scope.
```

### Implementation Review Agent (Phase 6)

Use with `subagent_type: "general-purpose"`. Launch two instances — one with accuracy focus, one with scope focus. Needs file access.

```
You are reviewing code changes before a commit. Your job is to catch implementation bugs.

## Original Task
{task_description}

## Approved Plan
{plan_summary}

## Changes to Review
{change_description}

## Your Focus: {focus}
{focus_description}

For each issue found, state: what's wrong, where, and a concrete fix.
Keep your review under 40 lines. Only flag real issues.
```

**Focus descriptions:**

- **Accuracy**: "Verify the changes match the approved plan. Check for typos, stale references, incorrect file paths, broken imports, and wrong function signatures. Read the actual changed files."
- **Scope**: "Check that changes are necessary and sufficient. Flag anything that wasn't in the plan, unnecessary abstractions, or missing pieces. Meta-review: should any of these changes NOT exist?"

## Notes

- **Interaction with folio**: Phase 1 checks for active folio projects. If one is relevant, its sources and cross-references inform the context summary.
- **Custom lenses**: Users can specify lenses naturally in the topic text (e.g., `/folio plan redesign auth, considering performance and readability`). Parse the user's intent and craft lens descriptions accordingly.
- **Retrospective nesting**: When `/folio plan` is invoked from a Phase 7 retrospective, its own Phase 7 is replaced by a one-line summary. Only one level of nesting is permitted.
