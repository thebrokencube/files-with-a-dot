# Plan Workflow — Design Phase (Phases 1-4)

Read by Agent 1 (Design). Self-contained for design sessions.

## Phase 1: Understand

Gather context before spawning any agents. This happens in the main conversation.

1. Identify the task scope from the user's request (or `<topic>` argument)
2. **Scope interview** (fires when task replaces or removes functionality): Ask the user
   "What should be preserved? What can be dropped? What needs to change?" Capture answers
   as hard constraints for the context summary.
3. Read relevant files — entry points, existing implementations, tests, CLAUDE.md
4. Search for related patterns in the codebase (Grep/Glob)
5. **Check for folio context**: Run `folio home list` to find active projects. If any project is relevant to the task, read its `folio.yml` and pull in relevant sources, cross-references, and observations.
6. **Present folio findings to the user.** If a project matched: summarize project name, key
   sources pulled in, relevant observations. Wait for confirmation before continuing. If no
   project matched: list the active projects that were considered, ask if any of them are
   relevant (the user may see a match the search missed), and ask if a new folio project
   should be created to track this work. Only proceed without folio context after explicit
   user confirmation.
7. **Check source freshness — MUST run if folio project matched.** For every source with a
   `derived_from` entry, compare the `cached` date against today. If any source is >14 days
   stale, STOP and present the list to the user: "These sources may be outdated: [list with
   ages]. Refresh before planning?" Wait for the user's response. If yes, use the gather
   workflow (see `references/gather.md`) per stale source. If no, note staleness in the
   context summary so lenses can account for it. Do not auto-run gather. Do not proceed to
   step 8 until the user has acknowledged or dismissed the staleness report.
8. **Pin hard constraints**: Separate the user's stated decisions and explicit preferences
   (hard constraints) from open trade-offs. Hard constraints are non-negotiable — lenses
   must not re-evaluate them. Include pinned constraints as a distinct section at the top
   of the context summary.
9. **Extensibility interview** (fires for tooling/infra tasks): Ask the user "What's coming
   next for this area?" Capture answers as soft context — they inform lens trade-offs but
   are not hard constraints.
10. Compile a **context summary** (max 30 lines): pinned hard constraints first, then what
    exists, what needs to change, key trade-offs, and relevant folio context (if any)
11. **Framing gate (hard):** Present the conceptual model — how the pieces fit together,
    what changes, what stays — to the user. Wait for confirmation before spawning propose
    agents. This catches misalignment before diverge-converge, not after.

**Context checkpoint**: If Phase 1 involved substantial research, compact findings into the
context summary before spawning Phase 2 agents — they receive a distilled summary, not raw
transcripts. For research-heavy tasks, consider splitting Phases 1-4 across sub-agents.

**Materialization gate**: If research produces substantial findings, materialize as spike files
via `folio new spike <topic>` before Phase 2. Commit via `folio home push`. Link each spike
conclusion to its source artifact: `[conclusion] <- [source path]`.

**Live data**: Use WebFetch to verify live data (npm versions, API docs, external specs) — do
not rely on training data. Flag facts that couldn't be live-verified.

## Phase 2: Propose

Launch 2 Plan agents in parallel, each with the same context summary but a different lens. Each returns a proposal (max 80 lines).

### Model Routing

Subagents use explicit model selection to balance cost and capability:

| Role | model | Rationale |
|------|-------|-----------|
| Propose | sonnet | Breadth exploration, constrained output |
| Converge | session default | Synthesis needs depth |
| Review | opus | Complex architectural reviews need depth (obs #42) |

When `model` is omitted, the agent inherits the session default.

Default lenses:
- **Pragmatic**: Minimize changes, reuse existing code, prefer the simplest approach that works
- **Thorough**: Consider edge cases, maintainability, architectural fit, future extensibility

### Propose Agent Prompt

Use with `subagent_type: "Plan"` and `model: "sonnet"`. Launch two instances in parallel with different lens values.

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
- Verify design doc claims against actual source code. Flag aspirational statements
  reported as facts. MUST read files, not assume prompt context is accurate.

Focus on architecture, file-level changes, and key design trade-offs. Defer per-function implementation detail, test strategy, and commit structure to the execution brief.

Keep your proposal under 80 lines. Be concrete — file paths, function names, specific changes. No hand-waving.
```

**Default lens descriptions:**

- **Pragmatic**: "Minimize the number of files changed and lines of code written. Reuse existing patterns and utilities. Prefer the simplest correct solution. Avoid new abstractions unless they pay for themselves immediately."
- **Thorough**: "Consider edge cases, error handling, and how this change interacts with the rest of the codebase. Prioritize maintainability and architectural consistency. Flag potential issues even if fixing them adds scope."

## Phase 3: Converge

Launch 1 agent (subagent_type: general-purpose, model: session default) to merge the two proposals into a single plan (max 100 lines).

Convergence criteria:
- Every file to be changed is listed with what changes and why
- Implementation order is specified
- Trade-offs noted where proposals diverged; agreements are strong signals — keep them
- Architectural decisions, type definitions, and key function signatures are pre-decided —
  implementation-level detail deferred to the Brief agent
- **Option-value interactions**: When rejecting an option, note what conditions would
  reinstate it — preserves reasoning without re-running diverge-converge

After the converge agent returns, briefly summarize (3–5 lines) the key divergence decisions to the user — which proposal won on each point and why. Informational only, not blocking. Proceed to Phase 4 immediately after.

### Converge Agent Prompt

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
- Pre-decide function signatures, type definitions, and edge-case handling where feasible

Keep the merged plan under 100 lines. Be concrete.
```

## Phase 4: Review

Launch 1 agent (subagent_type: general-purpose) to review the merged plan. The review covers:

1. **Accuracy**: Does the plan reference correct file paths, function names, and APIs? Are assumptions about existing code valid?
2. **Feasibility**: Can each step actually be implemented as described? Are there missing dependencies or ordering issues?
3. **Scope**: Is everything in the plan necessary for the task? **Meta-review: should any of this work not exist?** Flag anything that's over-engineered, gold-plated, or solving a problem the user didn't ask about.

Review output: max 40 lines. For each issue found, state: what's wrong, where, and a suggested fix.

Loop until the review returns zero issues. Cap at 3 iterations — if issues persist after 3
rounds, present remaining issues to the user for judgment.

**Design doc (mandatory):** After the review, commit the design doc. Every plan produces one — lightweight for simple changes, but it always exists.

1. Scaffold via `folio new design <topic>`. This creates
   `reference/design/YYYY-MM-DD-<topic>.md` with the design template and registers it in
   folio.yml.
2. Fill in from converge output: Problem, Architecture, Divergence decisions, What's NOT
   included, Design Provenance (agent count, lens names, review findings).
3. **Scope approval gate (hard):** Present the **What's NOT Included** section to the user
   for explicit sign-off before committing. This is scope negotiation, not just documentation —
   gaps here cause re-runs. Wait for "yes" before proceeding.
4. If a folio project exists: commit via `folio home push`
5. If no folio project: use `--no-register` and write to the plan file's directory instead
6. Present to user: "Design doc committed."

The committed design doc is the contract for Agent 2.

### Review Agent Prompt

Use with `subagent_type: "general-purpose"` and `model: "opus"`. Needs file access to verify claims.

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

### Multi-Perspective Review (`--pe-review`)

When `/folio plan --pe-review` is specified, replace the single Phase 4 review agent with 5
parallel agents (API surface, blast radius, migration risk, test coverage, UX). Converge their
findings before the design doc commit. Use for high-stakes or cross-cutting changes.

For re-run and amend-design rules, see plan.md.

## Session Exit (mandatory)

1. **Retro prompt**: "Anything worth retroing on before we move to the next phase?"
   Materialize findings via `folio new retro <topic>` and observation items. Commit via
   `folio home push`. For lightweight retros, observation items alone suffice.
2. **Handoff gate** — two options:
   - **Continue** (default): Spawn next agent via Agent tool with fresh context. Pass only
     the committed artifact path and setup instructions — no conversation history.
   - **New session**: Provide a paste-able prompt for the user to start fresh.
   Format: "[Artifact] committed at [path]. **Continue to [next phase], or hand off to a
   new session?**"
3. **Clipboard delivery** (mandatory for new-session handoff): Write the handoff prompt to a
   temp file and `pbcopy < /tmp/handoff-prompt.txt`. The prompt exists in the doc for
   durability, but clipboard is how the user actually starts the next session.
