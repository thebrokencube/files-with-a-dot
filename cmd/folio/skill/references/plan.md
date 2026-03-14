# Plan Workflow

Read by `/folio plan [topic]`. Multi-agent diverge-converge planning for non-trivial tasks.

## When to Use

Use `/folio plan` instead of `EnterPlanMode` for any non-trivial task — multi-file changes, architectural decisions, unclear requirements, or multiple valid approaches.

**Skip planning entirely** for trivial changes: single-file fixes, obvious bugs, one-line tweaks. Just do them.

## Workflow

```
  Agent 1 (Design)       Agent 2 (Brief)        Agent 3 (Execute)
  ─────────────────      ────────────────        ─────────────────
  Phases 1-4             Phases 5-6              Phases 7-8
        │                      │                       │
        ▼                      ▼                       ▼
  ┌───────────┐         ┌───────────┐           implement per track
  │design doc │────────→│work brief │────────→  validate → review →
  │ COMMITTED │         │ COMMITTED │             commit → retro
  └───────────┘         └───────────┘
```

The forward path: gather sources (Phase 1), freeze architecture in a design doc (Phase 4), decompose into tracks by risk profile (Phase 5), write execution-level briefs per track (Phase 6), execute per track (Phase 7). Each agent commits its output before the next begins.

Phase 7 is a loop per track step: implement → validate → review (mandatory gate) → commit. No commit without review.

When the plan has external targets (Jira hierarchy, branch topology), execution feeds into compose/publish to sync outputs. This is optional — not every plan has external targets.

The feedback loop: implementation experience feeds the Phase 8 retrospective, whose actionable findings become observation items that inform future cycles. Design docs and work briefs persist as durable reference material, not disposable intermediates.

## Pipeline

The plan workflow runs as a 3-agent pipeline. Each agent operates in a separate session with bounded context — it reads only its input artifact, not prior conversation history.

| Agent | Phases | Input | Output (committed) |
|-------|--------|-------|-------------------|
| Design | 1-4 | User request + codebase | Design doc |
| Brief | 5-6 | Design doc | Work brief with tracks |
| Execute | 7-8 | Work brief | Code + retro |

**Why three agents, not one:**

1. **Checkpoint principle**: When producing N artifacts with soft dependencies, commit each before starting the next. A single session that designs, briefs, and implements accumulates risk — a late failure (context overflow, wrong assumption) can lose work from all prior phases. Committed artifacts are checkpoints.
2. **Bounded context depth**: Design decisions require different context than implementation details. The design agent reads surveys, spikes, and comparable systems. The execution agent reads function signatures and test fixtures. Mixing both in one session wastes context on information irrelevant to the current phase.
3. **Artifact-mediated handoff**: The design doc and work brief are the contracts between agents. If an execution agent can't proceed from the brief alone, the brief is underspecified — fix the brief, don't compensate with conversation context.

**Invocation**: Agent 1 is invoked by `/folio plan`. Agents 2 and 3 are separate sessions — the user starts them after reviewing the prior agent's committed output.

**Collapsing the pipeline**: For simple changes where the design doc and work brief would be trivially small, Agents 1 and 2 can run in the same session. The Pipeline section defines the default — collapse deliberately, not by default.

## Invocation

- `/folio plan` — Infer topic from context, default lenses (pragmatic vs thorough)
- `/folio plan <topic>` — Explicit topic, default lenses

Custom lenses are specified naturally in the topic text (e.g., `/folio plan redesign auth, considering performance and readability`). The agent parses the user's intent and crafts lens descriptions accordingly. Default lenses (pragmatic vs thorough) apply when no lens guidance is given.

## Phases

Phases 1-4 run in Agent 1 (Design). Phases 5-6 run in Agent 2 (Brief). Phases 7-8 run in Agent 3 (Execute). Each agent session has full tool access. See the Pipeline section for agent boundaries and handoff rules.

### Phase 1: Understand

Gather context before spawning any agents. This happens in the main conversation.

1. Identify the task scope from the user's request (or `<topic>` argument)
2. Read relevant files — entry points, existing implementations, tests, CLAUDE.md
3. Search for related patterns in the codebase (Grep/Glob)
4. **Check for folio context**: Run `folio home list` to find active projects. If any project is relevant to the task, read its `folio.yml` and pull in relevant sources, cross-references, and observations.
5. **Present folio findings to the user.** If a project matched: summarize project name, key
   sources pulled in, relevant observations. Wait for confirmation before continuing. If no
   project matched: list the active projects that were considered, ask if any of them are
   relevant (the user may see a match the search missed), and ask if a new folio project
   should be created to track this work. Only proceed without folio context after explicit
   user confirmation.
6. **Check source freshness — MUST run if folio project matched.** For every source with a
   `derived_from` entry, compare the `cached` date against today. If any source is >14 days
   stale, STOP and present the list to the user: "These sources may be outdated: [list with
   ages]. Refresh before planning?" Wait for the user's response. If yes, use the gather
   workflow (see `references/gather.md`) per stale source. If no, note staleness in the
   context summary so lenses can account for it. Do not auto-run gather. Do not proceed to
   step 7 until the user has acknowledged or dismissed the staleness report.
7. **Pin hard constraints**: Separate the user's stated decisions and explicit preferences
   (hard constraints) from open trade-offs. Hard constraints are non-negotiable — lenses
   must not re-evaluate them. Include pinned constraints as a distinct section at the top
   of the context summary.
8. Compile a **context summary** (max 30 lines): pinned hard constraints first, then what
   exists, what needs to change, key trade-offs, and relevant folio context (if any)

This summary is passed to all downstream agents. Diversity comes from lenses, not information asymmetry.

**Context checkpoint**: If Phase 1 involved substantial research (multiple file reads, web
fetches, folio source review), summarize and compact the findings into the context summary
before spawning Phase 2 agents. Propose agents operate within bounded context — they should
receive a distilled summary, not raw research transcripts. For research-heavy tasks, consider
splitting Phases 1-4 across sub-agents when Phase 1 alone consumes significant context.

**Materialization gate**: If Phase 1 research produces substantial findings (not just reading
existing files), materialize them as spike files via `folio new spike <topic>` before Phase 2
begins. Commit via `folio home push`. The spikes become sources that propose agents can reference.

**Live data**: When research requires live data (npm versions, API docs, external specs), use
WebFetch to verify — do not rely on training data for version numbers or API surfaces. Flag
any facts that couldn't be live-verified so the review phase can catch staleness.

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
- Architectural decisions, type definitions, and key function signatures are pre-decided —
  implementation-level detail (test names, commit structure, validation commands) is deferred
  to the Brief agent
- **Option-value interactions**: When rejecting an option from either proposal, note what
  conditions would reinstate it. This preserves the reasoning — if assumptions change later,
  the team knows which rejected options become viable again without re-running the full
  diverge-converge

After the converge agent returns, briefly summarize (3–5 lines) the key divergence decisions to the user — which proposal won on each point and why. Informational only, not blocking. Proceed to Phase 4 immediately after.

### Phase 4: Review

Launch 1 agent (subagent_type: general-purpose) to review the merged plan. The review covers:

1. **Accuracy**: Does the plan reference correct file paths, function names, and APIs? Are assumptions about existing code valid?
2. **Feasibility**: Can each step actually be implemented as described? Are there missing dependencies or ordering issues?
3. **Scope**: Is everything in the plan necessary for the task? **Meta-review: should any of this work not exist?** Flag anything that's over-engineered, gold-plated, or solving a problem the user didn't ask about.

Review output: max 40 lines. For each issue found, state: what's wrong, where, and a suggested fix.

**Design doc (mandatory):** After the review, commit the design doc. Every plan produces one — lightweight for simple changes, but it always exists.

1. Scaffold via `folio new design <topic>`. This creates
   `reference/design/YYYY-MM-DD-<topic>.md` with the design template and registers it in
   folio.yml. Fill in from converge output:
   - **Problem**: What and why
   - **Architecture**: Key decisions, type definitions, function signatures
   - **Divergence decisions**: Table from converge phase
   - **What's NOT included**: Explicit scope boundary
   - **Design Provenance**: Agent count, lens names, review findings
2. **Scope approval gate (hard):** Present the **What's NOT Included** section to the user
   for explicit sign-off before committing. This is scope negotiation, not just documentation —
   gaps here cause re-runs. Wait for "yes" before proceeding.
3. If a folio project exists: commit via `folio home push`
4. If no folio project: use `--no-register` and write to the plan file's directory instead
5. Present to user: "Design doc committed."

The committed design doc is the contract for Agent 2.

**Session exit sequence (mandatory at every agent boundary):**
1. **Retro prompt**: Ask "Anything worth retroing on before we move to the next phase?"
   Capture findings as observation items or a retro file (see Phase 8 for recording rules).
2. **Handoff gate**: Present the next phase and offer two options:
   - **Continue** (default): Spawn the next agent via the Agent tool with fresh context.
     The agent receives only the committed artifact path and the standard setup instructions
     — no conversation history leaks across the boundary. This consumes parent context budget
     but avoids manual session switching.
   - **New session**: Provide a paste-able prompt for the user to start a fresh session.
     Use this when the parent context is already heavy or the user prefers manual control.

   Format: "Design doc committed at [path]. **Continue to Brief phase, or hand off to a
   new session?**"

**Multi-perspective review variant** (`/folio plan --pe-review`): When specified, replace the
single Phase 4 review agent with 5 parallel agents, each with a distinct perspective: API
surface, blast radius, migration risk, test coverage, and UX. Converge their findings before
the design doc commit. This is heavier but catches issues the single-perspective review misses —
use for high-stakes or cross-cutting changes.

### Phase 5: Decompose (Agent 2)

**Separate session from Agent 1.** Reads the committed design doc — no prior conversation context.

Analyze the design doc and break it into implementation tracks:

1. **Read the design doc.** This is the only input — do not rely on conversation history.
2. **Identify tracks.** Each track is an independently executable stream of work. Tracks
   should be scoped so an execution agent can pick up any single track without needing
   context from other tracks.
3. **Sequence by risk.** Order tracks so the highest-risk work runs first — failures surface
   early rather than after low-risk work is committed. Risk factors: architectural novelty,
   cross-file dependencies, test coverage gaps, integration surface area.
4. **Determine track dependencies.** Mark which tracks are independent (parallelizable) vs.
   sequential (each depends on the prior track's output).
5. **Present the track structure to the user.** List tracks with: name, risk assessment,
   sequencing rationale, and estimated commit count. Wait for approval before proceeding
   to Phase 6.

### Phase 6: Brief (Agent 2)

Populate each track with execution-level detail. The brief must be self-contained — the
execution agent reads only the brief, not the design doc or conversation history.

Every brief has four required sections, in order:

#### Context section (required)

Distill the design doc into the minimum context an execution agent needs to make judgment
calls when implementation deviates from the plan. This replaces "go read the design doc."

Include:
- **What this work is** (2-3 sentences): the problem, the goal, what repo(s) are involved
- **Key design decisions** (bulleted, non-negotiable): architectural choices the execution
  agent must not revisit. Sourced from the design doc's Architecture and Divergence Decisions
  sections. Frame as constraints: "Do X" / "Do NOT do Y" / "X stays because Y."
- **Scope boundary** (bulleted): what is NOT included, framed as stop signals. Sourced from
  the design doc's What's NOT Included section. The execution agent should recognize when
  it's drifting out of scope.

Target: 10-15 lines. Dense, not conversational. Every line should change how the agent
behaves — if a line doesn't affect execution decisions, cut it.

#### Agent Setup section (required)

Tell the execution agent how to prepare before touching any code. Three parts:

1. **Skill loading**: Which skills to invoke and why. Standard set:
   - `/folio status` (or any `/folio` subcommand) — loads folio conventions. Call out the
     key rule: `~/.folio` commits use `folio home push`, never raw git.
   - `/commit` — loads commit format. Note repo-specific conventions (e.g., dotfiles use
     versioned tags, ~/.folio uses plain conventional commits).
   - Additional skills as needed (e.g., `/stacked-pr` if the work involves branch stacks).

2. **Repo mapping**: Which tracks operate in which repos. Execution agents work in one repo
   at a time — make the mapping unambiguous.

3. **Escalation triggers**: Specific to this work — when should the agent stop and ask the
   user instead of improvising? Common triggers:
   - Validation failures after migration steps
   - File paths or signatures that don't match the brief (implies codebase changed since
     briefing)
   - Temptation to add code not described in the track
   - Commits flagged as high-complexity in the track description

#### Tracks section (required)

For each track, specify:
- Exact file changes (create, modify, delete) with paths
- Function signatures, struct diffs, type definitions
- Test names and what they validate
- Commit message(s) and what each commit contains
- Validation commands (build, test, lint)

#### Execution Conventions section (required)

Commit format, scope target (max commits — typically ~5), validation commands,
module/package path, push workflow, and repo-specific patterns.

Include a **Folio integration** subsection: targets to add for branches, observation
items to resolve on completion, `folio home push` checkpoints at milestones. Execution agents
should maintain folio state as they go — not as a final cleanup step.

---

Scaffold the brief under `work/active/YYYY-MM-DD-<topic>/README.md`. If the brief needs
per-track detail files (large tracks), create `track-N.md` siblings.

**Brief verification gate (hard):** Before committing, launch verification agents to confirm
all referenced file paths, line numbers, and function signatures are accurate. Approximate
references cause mid-execution corrections — verify them upfront. Fix any inaccuracies, then
commit.

Commit via `folio home push`. The committed work brief is the contract for Agent 3.

**Session exit sequence** (same as Phase 4 — retro prompt, then handoff gate):
1. Ask "Anything worth retroing on before we move to execution?"
2. "Work brief committed at [path]. **Continue to Execute phase, or hand off to a new
   session?**"

### Phase 7: Execute (Agent 3)

**Separate session from Agent 2.** Reads the committed work brief — no prior conversation context. If the brief has multiple tracks, each track can be executed independently (same or separate sessions).

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

**Folio integration**: If a relevant folio project exists, record design decisions, progress, and rationale in the folio project as work progresses — not as a final cleanup step. This means updating folio.yml observations, adding reference files for significant decisions, and keeping cross-references current throughout implementation. All `~/.folio` commits must use `folio home push` (see SKILL.md § Git Operations).

### Phase 8: Retrospective (every agent)

**Mandatory — no skip.** Every agent session ends with a retro prompt (see session exit
sequence in Phases 4 and 6). The execution agent (Agent 3) does the most thorough retro
since it has implementation context. Design and Brief agents do lightweight retros focused
on process friction and artifact quality. Do NOT delegate retros to a subagent. Cover:

- What worked well? What added friction?
- Was the work brief sufficient? Could the execution agent proceed without reading the design doc?
- Were track boundaries well-chosen? Did tracks interfere or require unexpected coordination?
- Was the risk sequencing useful? Did high-risk-first ordering surface issues early?
- What should change next time?
- If the plan changed folio source files, flag whether targets need recompilation.

Only capture actionable findings, not session notes. Findings that aren't worth planning
aren't actionable — note them in the retro summary and move on.

**Recording findings**: Retro findings worth preserving MUST be materialized via
`folio new retro <topic>` — not just added as observation items. The retro file captures the full
context; observation items capture only the actionable follow-ups. Both are written: the retro file
is the durable artifact, observation items are the work triggers. Commit via `folio home push`.

For lightweight retros (few findings, simple context), observation items alone are sufficient — use
judgment on whether the full context warrants a retro file.

## Re-run Rule

If design doc feedback requires rethinking (not just minor edits), re-run Phases 2-4 within Agent 1. If work brief feedback requires restructuring tracks, re-run Phases 5-6 in a new Agent 2 session. If execution reveals a design-level flaw (not just a brief gap), escalate to the user — they may re-run from Phase 2 (new design) or Phase 5 (new brief) depending on severity. Do not patch committed artifacts inline — re-run the producing agent.

**Amend-design path**: For additive scope changes where the core architecture is settled (adding
a track, extending a type, covering an edge case), a full diverge-converge re-run is overkill.
Instead: (1) describe the amendment to the user, (2) get explicit approval, (3) edit the design
doc directly, (4) re-run Phase 4 review only against the amended section, (5) commit the update.
Use this path only when the change is additive — if it contradicts existing decisions, re-run
from Phase 2.

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

Focus on architecture, file-level changes, and key design trade-offs. Defer per-function implementation detail, test strategy, and commit structure to the execution brief.

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
- Pre-decide function signatures, type definitions, and edge-case handling where feasible

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

### Implementation Review Agent (Phase 7)

Use with `subagent_type: "general-purpose"`. Launch two instances — one with accuracy focus, one with scope focus. Needs file access.

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

## Notes

- **Interaction with folio**: Phase 1 checks for active folio projects. If one is relevant, its sources and cross-references inform the context summary.
- **Design doc**: Always produced in Phase 4. The doc is the authoritative input for Agent 2.
- **Custom lenses**: Users can specify lenses naturally in the topic text (e.g., `/folio plan redesign auth, considering performance and readability`). Parse the user's intent and craft lens descriptions accordingly.
- **Retrospective findings**: Phase 8 captures findings as observation items in the folio project, not as recursive `/folio plan` invocations.
- **Agent boundaries**: The Pipeline section defines default boundaries. When collapsing agents (running multiple phases in one session), preserve the commit checkpoints — the design doc and work brief must still be committed before execution begins.
