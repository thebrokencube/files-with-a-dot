---
name: architect
description: Diverge-converge architect for non-trivial tasks. Two agents propose through different lenses, then merge + review. Extends EnterPlanMode with upfront research and multi-agent planning.
user_invocable: true
---

# Architect

Multi-agent planning skill for non-trivial implementation tasks. Two agents propose independently through different lenses, then their proposals are merged and reviewed before presenting to the user.

Phases: understand → propose → converge → review → present → implement → retrospective. Phases 1-4 run in normal mode (full tool access). Phase 5 enters built-in plan mode for structured approval. Phase 6 implements the approved plan (with review before each commit). Phase 7 runs post-implementation.

## When to Use

Use `/architect` instead of `EnterPlanMode` for any non-trivial task — multi-file changes, architectural decisions, unclear requirements, or multiple valid approaches.

**Skip planning entirely** for trivial changes: single-file fixes, obvious bugs, one-line tweaks. Just do them.

## Commands

- `/architect` or `/architect <topic>` — Plan with default lenses (pragmatic vs thorough)
- `/architect <lens1> vs <lens2>` — Plan with custom lenses

If a topic is provided, use it as the planning focus. Otherwise, infer from recent conversation context.

## Workflow

### Phase 1: Understand

Gather context before spawning any agents. This happens in the main conversation.

1. Identify the task scope from the user's request (or `<topic>` argument)
2. Read relevant files — entry points, existing implementations, tests, CLAUDE.md
3. Search for related patterns in the codebase (Grep/Glob)
4. **Check for folio context**: Run `folio home list` to find active projects. If any project is relevant to the task, read its `folio.yml` and pull in relevant sources, cross-references, and tasks/pending items.
5. Compile a **context summary** (max 30 lines): what exists, what needs to change, key constraints, and relevant folio context (if any)

This summary is passed to all downstream agents. Diversity comes from lenses, not information asymmetry.

### Phase 2: Propose

Launch 2 Plan agents in parallel, each with the same context summary but a different lens. Each returns a proposal (max 80 lines).

Default lenses (override with `/architect <lens1> vs <lens2>`):
- **Pragmatic**: Minimize changes, reuse existing code, prefer the simplest approach that works
- **Thorough**: Consider edge cases, maintainability, architectural fit, future extensibility

### Phase 3: Converge

Launch 1 agent (subagent_type: general-purpose) to merge the two proposals into a single plan (max 100 lines).

Convergence criteria:
- Every file to be changed is listed with what changes and why
- Implementation order is specified
- Trade-offs between the two proposals are noted where they diverged meaningfully
- If both proposals agreed on an approach, that's a strong signal — keep it

### Phase 4: Review

Launch 1 agent (subagent_type=general-purpose) to review the merged plan. The review covers three checks:

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
3. Call `ExitPlanMode` with `allowedPrompts` populated from the plan — e.g., if the plan includes running tests, include `{"tool": "Bash", "prompt": "run tests"}`. This presents the plan to the user for approval.

The user sees the plan file cleanly. They may request changes, ask questions, or reject. Iterate until approved or abandoned.

**Note**: Full conversation context (Phases 1-4) is retained in plan mode — only tool access is restricted (read-only + no Agent). This is fine since the heavy lifting is done.

### Phase 6: Implement

Only after user approval (ExitPlanMode accepted). Execute the plan step by step, following the specified order. If you discover something unexpected during implementation that contradicts the plan, stop and consult the user rather than improvising.

**Review before each commit**: Before every commit, launch 2 review agents (subagent_type: general-purpose) — one checking accuracy, one checking scope — then converge findings and fix issues before committing. This catches implementation bugs that planning can't — typos, stale references, broken cross-references.

### Phase 7: Retrospective

**Mandatory — no skip.** After all implementation commits are complete, launch 1 agent (subagent_type: general-purpose) to review the planning process itself — what worked, what added friction, what to change next time. Only capture actionable findings, not session notes. Present findings to the user and ask where to record them and if any warrant immediate changes.

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

### Retrospective Agent (Phase 7)

Use with `subagent_type: "general-purpose"`.

```
You are reviewing how an /architect planning process went.

## Original Task
{task_description}

## Process Summary
{process_summary}

## Instructions
Briefly review the planning process:
- What worked well? What added friction?
- Were the lenses useful? Did convergence surface good trade-offs?
- Was the folio context helpful (if used)?
- What should change next time?

Only capture actionable findings, not session notes. Keep under 20 lines.
```

## Notes

- **Interaction with folio**: Phase 1 checks for active folio projects. If one is relevant, its sources and cross-references inform the context summary. After implementation, folio targets that reference changed code may need recompilation — flag this in the retrospective.
- **Custom lenses**: Users can specify lenses with `/architect <lens1> vs <lens2>`, e.g., `/architect performance vs readability`. Parse the two sides of "vs" as lens names and craft descriptions that match.
