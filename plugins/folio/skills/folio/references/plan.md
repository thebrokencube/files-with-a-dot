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
  │design doc │────────→│work plan  │────────→  validate → review →
  │ COMMITTED │         │ COMMITTED │             commit → retro
  └───────────┘         └───────────┘
```

The forward path: gather sources (Phase 1), sketch the idea/architecture as an HTML page you react to and freeze it (Phase 1.5), freeze architecture in a design doc (Phase 4a-4b), decompose into tracks (Phase 5), write execution-level briefs (Phase 6), execute per track (Phase 7). Each agent commits its output before the next begins.

Phase 1.5 (idea/arch sketch) fronts the pipeline, runs the sketch past a fresh reviewer subagent then a birds-eye sign-off gate, and decides lightweight-vs-full — see `references/plan-idea.md`. Phase 7 is a loop: implement → validate → review (mandatory gate) → commit.

## Pipeline

The plan workflow runs as a 3-agent pipeline. Each agent operates in a separate session with bounded context — it reads only its input artifact, not prior conversation history.

| Agent | Phases | Input | Output (committed) |
|-------|--------|-------|-------------------|
| Design | 1-4 | User request + codebase | Design doc |
| Brief | 5-6 | Design doc | Work plan with tracks |
| Execute | 7-8 | Work plan | Code + retro |

Phase 1.5 (idea/arch sketch) runs inside the Design agent — lead-driven, before Phase 2, committed before the design doc. See `references/plan-idea.md`.

**Why three agents, not one:**

1. **Checkpoint principle**: Commit each artifact before starting the next — a late failure can't lose prior work.
2. **Bounded context**: Each phase needs different context; mixing wastes tokens on irrelevant information.
3. **Artifact-mediated handoff**: If an execution agent can't proceed from the brief alone, the brief is underspecified — fix the brief, don't compensate with conversation context.

**Invocation**: Agent 1 is invoked by `/folio plan`. Agents 2 and 3 are separate sessions — the user starts them after reviewing the prior agent's committed output.

**Hard rule: execution REQUIRES a committed plan artifact.** No mode skips this. In full mode it's a committed `README.md` work plan (Design → Brief → Execute). In lightweight (N==1) mode it's the frozen idea/arch sketch + a committed `track-1.md`. If you'd start execution with neither committed, STOP.

## Iteration Across Sessions

Planning is often iterative — a single session rarely produces a fully hardened design. The Design phase (Agent 1) can loop across multiple sessions:

```
Session 1: Research + first diverge-converge -> design doc v1
  User reviews, identifies weak areas
Session 2: Deep-dive weak areas -> design doc v2
  User reviews, still not confident on X
Session 3: Focused round on X -> design doc v3
  User: "ship it" -> proceed to Brief phase
```

Each session ends by **updating the design doc in place** — pinned constraints,
convergence status, open questions — and committing via `folio home push`. The next
session resumes by running `/folio <project>`, which surfaces the design doc's current
state inline. There is no separate handoff file. See plan-design.md Session Exit for
the update procedure and SKILL.md "Bare Invocation" for the resume flow.

**Key principle**: Each round should deepen specific areas, not repeat broad exploration. If round 1 produced a strong architecture but weak migration plan, round 2 focuses agent teams on migration feasibility — don't re-derive the architecture.

**Convergence tracking**: The design doc's `## Convergence Status` section MUST track
per-layer status: Direction (problem/approach/scope), Interfaces (contracts/tracks/
cross-cutting), and Implementation (technique choices). Use the status vocabulary:
EXPLORING, PROPOSED, SETTLED (Round N), AMENDED (Round N), NEEDS REVIEW, DEFERRED,
IN PROGRESS. Per-layer round budgets apply:

| Layer | Budget | Circuit breaker |
|-------|--------|----------------|
| Direction | 1-3 rounds | After 3 direction rounds without convergence, escalate |
| Interface | 1-2 rounds | After 2 interface rounds post-direction-settled, escalate |
| Implementation | 0-1 rounds | After 1 round, escalate (shouldn't normally run) |
| **Global maximum** | **5 rounds** | Hard escalate regardless of per-layer state |

When escalating, present explicit options: (a) accept current state, (b) narrow scope,
(c) spike on blocking question.

The design doc is the connective tissue across sessions. It MUST be complete enough that
a fresh session running `/folio <project>` can pick up without reading conversation
history or any other file beyond the design doc + observations.

## Invocation

- `/folio plan` — Infer topic from context, default lenses (pragmatic vs thorough)
- `/folio plan <topic>` — Explicit topic, default lenses

Custom lenses are specified naturally in the topic text (e.g., `/folio plan redesign auth, considering performance and readability`). The agent parses the user's intent and crafts lens descriptions accordingly. Default lenses (pragmatic vs thorough) apply when no lens guidance is given.

## Three Modes (decided at Phase 1.5 sketch-freeze)

The idea/arch sketch's track-title fan-out (`references/plan-idea.md`) picks the mode:

1. **Skip planning** — trivial single-file fix; never reaches `/folio plan` (do it directly).
2. **Lightweight (N == 1 track)** — the frozen sketch decomposes to a single `type(scope):` line
   (single clear change, single repo). Skip the design doc AND the brief; create one `track-1.md`
   and go straight to execution + commits. The committed plan artifact is the **sketch page + that
   single track file** — this satisfies the "execution requires a committed plan" rule.
3. **Full (N ≥ 2 tracks)** — run Phases 2-8: design doc → brief → (burndown?) → execute.

Co-signal + track-count details: `references/plan-idea.md`. If ambiguous, use the full pipeline.

Commit checkpoints are mandatory in every mode. Lightweight reduces artifacts, not checkpoints:
the sketch and the single track file must both be committed before execution begins. The Phase 1.5
sketch gates (fresh-subagent review + birds-eye sign-off) run in both modes; beyond them, in
lightweight mode the Phase 7 per-commit review + adversarial verify are the only code-level quality
gates (design/brief gates are skipped), so they are mandatory — see `references/plan-execute.md`.

## Re-run Rule

If design doc feedback requires rethinking (not just minor edits), re-run Phases 2-4 within
Agent 1. If work plan feedback requires restructuring, re-run Phases 5-6 in a new Agent 2
session. If execution reveals a flaw, classify by layer: direction-level flaw → re-run from
Phase 2 (new design round); interface-level flaw → re-run from Phase 5 (new brief) or patch
the interface spec; implementation question → agent resolves or asks user (no re-run). See
plan-execute.md Layered Escalation for the full classification table. Do not patch committed
artifacts inline — re-run the producing agent.

**Amend-design path**: For additive scope changes where core architecture is settled: (1) describe
the amendment, (2) get explicit approval, (3) edit the design doc, (4) re-run Phase 4b review
only, (5) commit. Use only when additive — if it contradicts existing decisions, re-run from
Phase 2.

## Agent Routing

Use the strongest available agent mode per phase, with graceful degradation:

| Phase | Preferred mode | Fallback mode | Rationale |
|-------|---------------|---------------|-----------|
| Phase 1 (Research) | Agent team (parallel explorers) | Sequential subagents | Breadth matters; parallel exploration finds more |
| Phase 1.5 (Sketch review) | Single fresh subagent (general-purpose) | Same | Single-file defect/render QA — lead-driven, no team |
| Phase 2 (Propose) | Agent team (parallel lenses) | 2 sequential subagents | Divergence benefits from independence |
| Phase 3 (Converge) | Single agent | Single agent | Convergence is inherently serial |
| Phase 4a (Fill) | Single agent | Single agent | Mechanical mapping from converge output |
| Phase 4b (Review) | 2 parallel opus subagents (devil's advocate + blast radius) | 1 opus subagent with adversarial prompt (single-agent fallback) | Adversarial perspectives catch what accuracy checks miss |
| Phase 5-6 (Brief) | Single agent | Single agent | Serial work, no parallelism benefit |
| Phase 7 (Execute) | Single agent per track | Single agent per track | Execution is focused, bounded |
| Phase 8 (Retro) | Single agent | Single agent | Reflection is serial |

**Detection**: Check if TeamCreate is available (tool inventory). If yes, use team mode for
Phases 1, 2, and 4b. If not, fall back to subagents. Both paths produce the same committed
artifacts — the handoff contract is identical regardless of mode.

**Materialization in team mode**: Team agents MUST write findings to files (not just
conversation memory). The convergence agent reads files, not conversation. This protects
against context compaction in long sessions.

## Phase References

-> Read references/plan-idea.md for Phase 1.5 (idea/arch sketch — fronts the pipeline)
-> Read references/plan-design.md for Phases 1-4 (Design agent)
-> Read references/plan-brief.md for Phases 5-6 (Brief agent)
-> Read references/plan-execute.md for Phases 7-8 (Execute agent)
-> Read references/burndown.md when plan brief declares shape: burndown (replaces Phase 7 step loop)
