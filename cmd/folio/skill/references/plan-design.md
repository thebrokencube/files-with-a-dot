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
11. **Source coverage check**: Before spawning propose agents, verify the problem space is
    understood. If the topic involves a domain the agent has not spiked on, STOP and suggest
    `/folio gather` first. Evidence of sufficient gathering: at least one spike or research
    reference in folio.yml sources for each key domain the task touches.
12. **Framing gate (hard):** Present the conceptual model — how the pieces fit together,
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

**Scaffold work directory (mandatory):** Before spawning Phase 2 agents, scaffold the design
doc early via `folio new design <topic>`. This creates the work directory at
`work/active/YYYY-MM-DD-<topic>/` and the design doc at
`reference/design/YYYY-MM-DD-<topic>.md` inside it. The design doc starts as a template —
it gets filled in during Phase 4. Early scaffolding ensures:
- `folio new round` has a work dir to create rounds in
- Vault research produced during Phase 2 can be registered with `depends_on` pointing to the
  design doc path
- The provenance chain is established incrementally, not retrofitted at the end

Commit the scaffold: `folio home push -m 'docs(<topic>): scaffold design doc and work dir'`

## Phase 2: Propose

Launch propose agents in parallel, each with the same context summary but a different lens.

### Standard vs Team Mode

| Mode | When | Agents | Output |
|------|------|--------|--------|
| **Standard** (default) | Single-concern, bounded scope | 2 subagents | Each returns proposal text (max 80 lines) |
| **Team** | Multi-concern, deep design, platform-level | 3-5 teammates | Each materializes findings to a file |

**Use team mode when**: The task spans multiple independent concerns that benefit from dedicated deep-dive focus — architecture + migration feasibility + evolvability, or performance + correctness + UX. Each lens should represent a genuinely different evaluation axis, not just "more thorough." The user can request team mode or custom lenses naturally in the topic text (e.g., "go nuts with agent teams").

**Recommended lens: devil's advocate.** When running 3+ agents in team mode, include a
devil's advocate agent that argues against the proposed approach. Its job is to find the
strongest counterarguments, identify simpler alternatives, and challenge whether the full
scope is warranted. This lens consistently prevents over-engineering — it forces the
convergence agent to address objections rather than rubber-stamp consensus. The devil's
advocate prompt should explicitly instruct the agent to be genuinely adversarial, not to
strawman and concede.

### Team Mode Protocol

When using team mode, each agent follows a standard protocol:

1. **Read shared context** — the context summary plus any files listed as required reading
2. **Explore** — read actual source code, fetch external docs, examine real files. Produce concrete findings, not summaries of summaries.
3. **Materialize to file** — write findings to `agent-research/{NNNN}-round/{lens-slug}.md`
   inside the work directory. The round directory is created when the team spins up (see
   Round Directory below). This is mandatory — agent memory is ephemeral, files survive
   context compaction and session boundaries.
4. **Signal done** — report completion with file path and 3-5 line summary of key findings

**Provenance registration (team mode):** After all propose agents complete, the lead agent
registers any vault research files they produced as sources in folio.yml. Each vault research
source gets a `depends_on` entry pointing to the design doc (scaffolded in Phase 1). This
builds the provenance chain incrementally — research → design doc. Use `folio home push` to
commit the updated folio.yml before proceeding to convergence.

The convergence agent (Phase 3) reads all materialized files, not conversation summaries.
It writes its output to `converged.md` in the same round directory.

### Round Directory

Each diverge-converge cycle gets a numbered round directory inside `agent-research/`:

```
agent-research/
  0001-round/
    pragmatic.md
    thorough.md
    converged.md
  0002-round/
    migration-feasibility.md
    performance.md
    converged.md
```

**Create the round directory when spinning up the agent team** — before dispatching propose
agents. Use the CLI to scaffold the next round:

```bash
folio new round <topic> --folio <path>
```

This auto-increments from existing `agent-research/????-round/` directories. First round is
`0001-round`. The work dir is created when the design doc is scaffolded after Phase 1.

Previous rounds stay intact for reference. The design doc's provenance section should note
which round produced the final convergence.

### Model Routing

Subagents use explicit model selection to balance cost and capability:

| Role | model | Rationale |
|------|-------|-----------|
| Propose (standard) | sonnet | Breadth exploration, constrained output |
| Propose (team) | session default | Deep exploration needs full capability |
| Converge | session default | Synthesis needs depth |
| Review | opus | Complex architectural reviews need depth (obs #42) |

When `model` is omitted, the agent inherits the session default.

Default lenses (standard 2-agent mode):
- **Pragmatic**: Minimize changes, reuse existing code, prefer the simplest approach that works
- **Thorough**: Consider edge cases, maintainability, architectural fit, future extensibility

### Propose Agent Prompt (Standard Mode)

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

### Propose Agent Prompt (Team Mode)

Use with `subagent_type: "general-purpose"` and session default model. Launch N instances in parallel.

```
You are a research/design agent on a team exploring: {task_description}

## Context
{context_summary}

## Your Lens: {lens_name}
{lens_description}

## Required Reading
{file_list}

## Protocol
1. Read all required files first for shared context
2. Explore the codebase through your lens — read actual source code, not just descriptions
3. Write your findings to: {round_dir}/{lens_slug}.md
   **Write ONLY to this path. Do not create, modify, or write to any other file.**
4. Your file MUST include:
   - Concrete findings with file paths and line references
   - Specific recommendations (not "consider X" — say "do X because Y")
   - Risks and mitigations for your area of focus
   - Open questions that need resolution
5. Report back with: file path, 3-5 line summary of key findings

Be concrete. Read real code. No hand-waving.
```

**Default lens descriptions:**

- **Pragmatic**: "Minimize the number of files changed and lines of code written. Reuse existing patterns and utilities. Prefer the simplest correct solution. Avoid new abstractions unless they pay for themselves immediately."
- **Thorough**: "Consider edge cases, error handling, and how this change interacts with the rest of the codebase. Prioritize maintainability and architectural consistency. Flag potential issues even if fixing them adds scope."

## Phase 3: Converge

Launch 1 agent (subagent_type: general-purpose, model: session default) to merge proposals into a single plan.

**Standard mode**: Converge agent receives proposal text directly (max 100 lines output).
**Team mode**: Converge agent receives file paths to all materialized proposals. It reads each file, then synthesizes. Output may be longer for complex multi-lens convergence — cap at 200 lines for team mode.

Convergence criteria:
- Every file to be changed is listed with what changes and why
- Implementation order is specified
- Trade-offs noted where proposals diverged; agreements are strong signals — keep them
- Architectural decisions, type definitions, and key function signatures are pre-decided —
  implementation-level detail deferred to the Brief agent
- **Per-layer convergence status**: The converge agent MUST report status for each planning
  layer — Direction (problem/approach/scope), Interfaces (contracts/tracks/cross-cutting),
  and Implementation (technique choices). Use the status vocabulary: EXPLORING, PROPOSED,
  SETTLED (Round N), AMENDED (Round N), NEEDS REVIEW, DEFERRED, IN PROGRESS. A unified "converged"
  assessment is insufficient — layer-level status enables the Session Exit gate to offer
  early exit when direction and interfaces are settled.
- **Option-value interactions**: When rejecting an option, note what conditions would
  reinstate it — preserves reasoning without re-running diverge-converge
- **Conflict resolution priority** (team mode): When proposals conflict, the converge agent
  must state which proposal wins and why. If a prior design doc or user constraint exists,
  it is authoritative over agent proposals.

**Knowledge gap feedback loop**: If propose agents identify knowledge gaps (unknown
constraints, unverified assumptions, missing domain context), STOP the converge phase.
Record the gaps as observations (`folio observe`). Suggest `/folio gather` for each gap.
Resume planning in a new session after gathering. Do not paper over gaps with assumptions.

After the converge agent returns, briefly summarize (3–5 lines) the key divergence decisions to the user — which proposal won on each point and why. Informational only, not blocking. Proceed to Phase 4 immediately after.

**Team mode materialization**: The converge agent MUST write its output to
`{round_dir}/converged.md` in addition to returning it. This protects against context
compaction in long sessions and preserves the convergence artifact for future reference.

### Converge Agent Prompt (Standard Mode)

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

### Converge Agent Prompt (Team Mode)

Use with `subagent_type: "general-purpose"`.

```
You are synthesizing findings from {N} research/design agents into a unified plan.

## Original Task
{task_description}

## Proposal Files
Read each of these files fully before synthesizing:
{file_path_list}

## Conflict Resolution
- If proposals agree, that's strong signal — keep it
- If proposals conflict, pick the stronger approach with explicit rationale
- Prior design docs and user-stated constraints are authoritative over proposals
- When rejecting an approach, note what conditions would reinstate it

## Instructions
Synthesize into a single unified plan:
- Resolve all conflicts — no "on one hand / on the other hand"
- Every file to change listed with what changes and why
- Implementation order with dependencies
- Pre-decide key signatures, types, and architectural decisions

Write your output to: {round_dir}/converged.md
Also return a 5-10 line summary of key decisions made.
```

## Phase 4a: Fill Design Doc

The design doc was scaffolded after Phase 1 — it already exists at
`reference/design/YYYY-MM-DD-<topic>.md` inside the work directory. Every plan produces
one — lightweight for simple changes, but it always exists.

1. Fill in the design doc using the layer-tagged template below. Source content from the
   converge output, mapping it to the appropriate layer sections.

   ```markdown
   # [Project Name] — Design Document

   ## Direction
   ### Problem Statement
   ### Alternatives Considered
   ### Chosen Approach
   ### Scope Boundary
   ### Non-Negotiable Constraints

   ## Interfaces
   ### Component Contracts
   [Type signatures, API shapes, data models — cross-boundary only.]
   ### Track Decomposition
   [Track list with file manifests, dependencies, sequencing.]
   ### Cross-Cutting Decisions
   [Shared types, system-wide patterns, technology choices affecting multiple tracks.]

   ## Implementation Notes (optional, often empty)
   [Per-component technique choices, only when JIT resolution is risky.]
   [Marked as advisory, not prescriptive.]

   ## Convergence Status
   - Direction: [EXPLORING | PROPOSED | SETTLED (Round N) | AMENDED (Round N)]
   - Interfaces: [EXPLORING | PROPOSED | SETTLED (Round N) | NEEDS REVIEW]
   - Implementation: [DEFERRED | IN PROGRESS | SETTLED (Round N)]

   ## Design Provenance
   [Agent count, lens names, review findings, round-by-round record.]
   ```

   The Direction section captures *what* and *why*. The Interfaces section captures
   cross-boundary contracts and track structure. Implementation Notes are optional and
   advisory — most projects leave this empty because execution agents resolve technique
   choices JIT via spikes. Convergence Status tracks per-layer progress across rounds.
   Design Provenance records how the design was produced (agent count, lenses, review fixes).

**Do NOT commit or proceed to review until the design doc is fully filled.** The review
agent reviews the design doc as written — not the converge output.

## Phase 4b: Review Design Doc (hard gate — blocks commit)

**Prerequisite: Phase 4a complete.** The design doc must be filled before review begins.
The review agent reviews the actual design doc artifact, not the converge output — this
ensures the committed artifact is the one that was validated.

Launch 1 agent (subagent_type: general-purpose, model: opus) to review the design doc. The review covers:

1. **Accuracy**: Does the design doc reference correct file paths, function names, and APIs? Are assumptions about existing code valid? Read the actual source files to verify.
2. **Feasibility**: Can each step actually be implemented as described? Are there missing dependencies or ordering issues?
3. **Scope**: Is everything in the design doc necessary for the task? **Meta-review: should any of this work not exist?** Flag anything that's over-engineered, gold-plated, or solving a problem the user didn't ask about.
4. **Completeness**: Are there gaps between the converge output and the design doc — decisions that were made during convergence but lost during fill?

Review output: max 40 lines. For each issue found, state: what's wrong, where, and a suggested fix.

Loop: fix issues in the design doc, re-run review. Cap at 5 iterations — if issues persist
after 5 rounds, present remaining issues to the user for judgment.

### Review Gate Checklist (must pass before commit)

- [ ] Review agent returned zero issues, OR 5 iterations completed AND user judged remaining issues
- [ ] All review fixes applied to the design doc (not just noted)

**After the review gate passes:**

1. **Scope approval gate (hard):** Present the **Scope Boundary** section to the user
   for explicit sign-off before committing. This is scope negotiation, not just documentation —
   gaps here cause re-runs. Wait for "yes" before proceeding.
2. **Provenance validation:** Before committing, verify all vault research produced during
   the round is registered in folio.yml with `depends_on` pointing to the design doc. Check
   that the design doc's `depends_on` list in folio.yml includes every vault research source
   that informed it. Missing links mean broken provenance — fix before committing.
3. If a folio project exists: commit via `folio home push`
4. If no folio project: use `--no-register` and write to the plan file's directory instead
5. Present to user: "Design doc committed."

The committed design doc is the contract for Agent 2.

### Review Agent Prompt

Use with `subagent_type: "general-purpose"` and `model: "opus"`. Needs file access to verify claims.

```
You are reviewing a design document. Your job is to find problems before the design is committed and handed off to implementation.

## Original Task
{task_description}

## Design Doc Path
{design_doc_path}

## Prior Issues (verify these are resolved)
{If round > 1: list of issues from previous review round}
{If round 1: "This is the first review round. No prior issues."}

## Instructions
Read the design doc at the path above. Review it against the actual codebase:

1. **Accuracy**: Verify file paths, function names, and API references exist and are correct. Read the actual files.
2. **Feasibility**: Can each step be implemented as described? Are there missing imports, wrong method signatures, or ordering issues?
3. **Scope**: Is everything necessary? Meta-review: should any part of this design NOT exist? Flag over-engineering, unnecessary abstractions, or work the user didn't ask for.
4. **Completeness**: Are any architectural decisions or interface contracts missing? Could an execution agent implement from this doc alone?
Do not flag issues that linters, formatters, or test suites would catch — those are handled by deterministic tools.

Report in two sections:
1. **Prior fix verification**: For each prior issue, state RESOLVED or STILL PRESENT with evidence.
2. **New findings**: Issues not in the prior list. For each: what's wrong, where, and a concrete fix.
Keep your review under 40 lines. Only flag real issues — don't nitpick style or add suggestions beyond the task scope.
```

### Multi-Perspective Review (`--pe-review`)

When `/folio plan --pe-review` is specified, replace the single Phase 4b review agent with 5
parallel agents (API surface, blast radius, migration risk, test coverage, UX). Converge their
findings before the design doc commit. Use for high-stakes or cross-cutting changes.

For re-run and amend-design rules, see plan.md.

## Session Exit (mandatory)

Every session that produces design work MUST complete all four steps before ending.

### Step 1: Confidence Check

Check the design doc's Convergence Status. The gate behavior depends on layer state:

**When Direction and Interfaces are both SETTLED:**

Present the stop-at-interfaces gate:

> Interfaces settled. Options:
> 1. **Stop here** (default) — execution agents resolve implementation JIT via spikes.
> 2. **One focused round** — pre-specify cross-cutting implementation choices only.
> 3. **Full implementation architecture** — pre-specify all component techniques.
>
> Recommendation: Stop here unless the project involves novel patterns with no codebase precedent.

Option 1 proceeds to Brief phase. Option 2 runs one more design round scoped to
Implementation layer only. Option 3 runs a full round (rare — most projects don't need it).

**When Direction or Interfaces are NOT yet SETTLED:**

Ask the user: **"Is this design hardened enough to move forward, or should we iterate more?"**

Three possible outcomes:

| Response | Action |
|----------|--------|
| **Ship it** | Proceed to Brief phase (Agent 2) or implementation |
| **Iterate** | Produce iteration handoff (Step 3b) identifying weak areas for next round |
| **Not sure** | Run `/folio review` on the design doc to surface gaps, then re-ask |

Planning is iterative. A single round of diverge-converge rarely produces a hardened plan.
It's perfectly fine — expected, even — to loop through multiple sessions before moving to
implementation. Each round should deepen specific areas, not repeat broad exploration.

### Step 2: Retro

"Anything worth retroing on before we wrap?"

Materialize findings via `folio new retro <topic>` and observation items. Commit via
`folio home push`. For lightweight retros, observation items alone suffice.

Retro scope includes the planning process itself — what worked about the agent team
composition, lens selection, convergence quality, context management.

### Step 3: Handoff Document

Write a handoff document to `/tmp/handoff-{topic}.md`. This is the single source of truth
for the next session — whether that's the next phase or another planning round.

**Progressive disclosure** (see `references/progressive-disclosure.md`): The handoff is
structured in layers. A new session that knows what to do reads 10 lines. A session that
needs to understand *why* reads deeper. Action first, context second, history last.

**Layer 1 — TL;DR + Start Here (always read)**

```markdown
# Handoff: {Topic} — {Phase/State}

## TL;DR
{One sentence: what happened.} {One sentence: what's open.} {One sentence: what to do.}

## Start Here

{Copy-paste prompt block for the next session. Self-contained — includes topic, key
constraints, design doc path, vault sources, and mode (e.g., "full agent teams powered").}

Skills to invoke: `/folio status --folio <path>`, `/commit`, {others as needed}
```

The TL;DR is max 3 sentences. The Start Here block is the prompt the user pastes — it must
work without reading anything else in the handoff. If the prompt needs context that isn't
in the design doc, the handoff is underspecified.

**Layer 2 — Context (read if the session needs to understand why)**

```markdown
## Open Questions
{Bulleted list of unresolved decisions. Only questions the next session must address.}

## Key Decisions (this session only)
{Numbered list of decisions made THIS session — prior sessions' decisions live in the
design doc. Don't repeat what's already committed.}

## Exit Criteria
{Concrete checklist for when the next round/phase is done.}
```

Keep Layer 2 under 30 lines. If a decision needs rationale, one sentence — not a paragraph.
Link to the design doc for full context rather than restating it.

**Layer 3 — Reference (skim or skip)**

```markdown
## Artifacts
{File tree of what was created/modified, with paths. Compact — no descriptions unless
the path isn't self-explanatory.}

## Folio
- Project: `{path}`
- Commit: `folio home push -m "type(scope): description"`

## Temp Files
{List /tmp files. One line each.}
```

**What NOT to include in the handoff:**
- Full decision history from prior rounds (that's the design doc's job)
- Restated design doc sections (link, don't copy)
- Convergence status tables that repeat settled items (only list what's open)
- Agent research summaries (those are in the round directories)

**Checklist before writing the handoff** (prevents forgetting things):
- [ ] All artifacts committed to folio? (`folio home push`)
- [ ] TL;DR is 3 sentences or fewer?
- [ ] Start Here prompt works without reading the rest of the handoff?
- [ ] Key Decisions only lists THIS session's decisions?
- [ ] Open questions are questions, not restated context?
- [ ] Exit criteria stated (if iterating)?

### Step 4: Handoff Validation (mandatory)

Spawn a fresh subagent with ONLY the handoff doc + design doc (no conversation context).
The subagent reports ambiguities — anything that requires context the handoff doesn't
provide. Fix ambiguities before session ends. Commit updated handoff via `folio home push`.

### Step 5: Clipboard Delivery

```bash
cat /tmp/handoff-{topic}.md | pbcopy
```

Always copy the full handoff to clipboard. The user starts the next session by pasting it.
Tell the user: "Handoff copied to clipboard. Paste it to start the next session."

## Exit Criteria Templates

Define these before starting each phase transition. Concrete checklists, not prose.

### Gather → Design
- [ ] All knowledge domains relevant to the task have sources in folio.yml
- [ ] No source >14 days stale for active domains
- [ ] Key trade-offs identified (at least 2 competing approaches)

### Design Round N → Round N+1
- [ ] Specific weak areas listed **with layer tag** (direction/interface/implementation)
- [ ] Prior round's weak areas resolved or explicitly accepted
- [ ] Per-layer round budgets respected (direction: 1-3, interface: 1-2, implementation: 0-1)
- [ ] No more than 5 total rounds without user checkpoint

### Design → Brief
- [ ] Direction and Interfaces SETTLED in Convergence Status
- [ ] High-risk areas validated (spike or code exploration)
- [ ] User signed off on scope boundary (Scope Boundary section)
- [ ] Design doc committed via `folio home push`
