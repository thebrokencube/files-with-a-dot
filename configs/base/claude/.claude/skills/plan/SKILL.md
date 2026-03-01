---
name: plan
description: Diverge-converge planner for non-trivial tasks. Two tiers — standard (single agent) and diverge (2 agents + review). Extends EnterPlanMode with upfront research and multi-agent planning.
user_invocable: true
---

# Plan

Planning skill for non-trivial implementation tasks. Two tiers:

- **Standard**: Single Plan agent proposes an approach. Good default.
- **Diverge**: Two agents propose independently through different lenses, then merge + review. Use when the design space is large or stakes are high.

Both tiers follow the same phases: understand → propose → (converge) → review → present → implement. Phases 1-4 run in normal mode (full tool access). Phase 5 enters built-in plan mode for structured approval.

## When to Use

Use `/plan` instead of `EnterPlanMode` for any non-trivial task — multi-file changes, architectural decisions, unclear requirements, or multiple valid approaches.

**Skip planning entirely** for trivial changes: single-file fixes, obvious bugs, one-line tweaks. Just do them.

## Commands

- `/plan` or `/plan <topic>` — Standard plan (single agent)
- `/plan diverge` or `/plan diverge <topic>` — Diverge plan (2 agents + review)

If a topic is provided, use it as the planning focus. Otherwise, infer from recent conversation context.

## Workflow

### Phase 1: Understand

Gather context before spawning any agents. This happens in the main conversation.

1. Identify the task scope from the user's request (or `<topic>` argument)
2. Read relevant files — entry points, existing implementations, tests, CLAUDE.md
3. Search for related patterns in the codebase (Grep/Glob)
4. Compile a **context summary** (max 30 lines): what exists, what needs to change, key constraints

This summary is passed to all downstream agents. Diversity comes from lenses, not information asymmetry.

### Phase 2: Propose

**Standard**: Launch 1 Plan agent with the context summary. It returns a proposal (max 80 lines).

**Diverge**: Launch 2 Plan agents in parallel, each with the same context summary but a different lens. Each returns a proposal (max 80 lines).

Default lenses (override by specifying custom lenses after `diverge`):
- **Pragmatic**: Minimize changes, reuse existing code, prefer the simplest approach that works
- **Thorough**: Consider edge cases, maintainability, architectural fit, future extensibility

### Phase 3: Converge (diverge only)

Merge the two proposals into a single plan. In the main conversation:

1. Read both proposals side by side
2. For each aspect of the plan, pick the stronger approach — or synthesize
3. Produce a **merged plan** (max 100 lines)

Convergence criteria:
- Every file to be changed is listed with what changes and why
- Implementation order is specified
- Trade-offs between the two proposals are noted where they diverged meaningfully
- If both proposals agreed on an approach, that's a strong signal — keep it

### Phase 4: Review

Launch 1 agent (subagent_type=general-purpose) to review the plan (standard: the single proposal; diverge: the merged plan). The review covers three checks:

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
   - If diverge: a brief note on where the two proposals differed and which was chosen
3. Call `ExitPlanMode` with `allowedPrompts` populated from the plan — e.g., if the plan includes running tests, include `{"tool": "Bash", "prompt": "run tests"}`. This presents the plan to the user for approval.

The user sees the plan file cleanly. They may request changes, ask questions, or reject. Iterate until approved or abandoned.

**Note**: Full conversation context (Phases 1-4) is retained in plan mode — only tool access is restricted (read-only + no Agent). This is fine since the heavy lifting is done.

### Phase 6: Implement

Only after user approval (ExitPlanMode accepted). Execute the plan step by step, following the specified order. If you discover something unexpected during implementation that contradicts the plan, stop and consult the user rather than improvising.

## Agent Prompts

### Standard Agent (Phase 2 — standard mode)

Use with `subagent_type: "Plan"`. Single instance.

```
You are planning an implementation for: {task_description}

## Context
{context_summary}

## Instructions
Propose an implementation plan. Your plan should:
- List every file to create or modify, with a description of the changes
- Specify implementation order and dependencies between steps
- Note any risks, assumptions, or open questions

Keep your proposal under 80 lines. Be concrete — file paths, function names, specific changes. No hand-waving.
```

### Diverge Agent (Phase 2 — diverge mode)

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

## Notes

- **Interaction with folio**: Folio handles project lifecycle (research, compile, audit). `/plan` handles implementation planning for code changes. They're complementary — use folio for knowledge work, `/plan` for "how should I build this."
- **When NOT to diverge**: If the task has an obvious single approach (e.g., "add a field to this form"), standard mode is faster and equally good. Diverge when you genuinely don't know the best approach, or when the wrong approach would be expensive to redo.
- **Custom lenses**: Users can specify lenses after `diverge`, e.g., `/plan diverge performance vs readability`. Parse the two sides of "vs" as lens names and craft descriptions that match.
