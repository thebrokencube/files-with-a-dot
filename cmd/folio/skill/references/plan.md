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

The forward path: gather sources (Phase 1), freeze architecture in a design doc (Phase 4), decompose into tracks (Phase 5), write execution-level briefs (Phase 6), execute per track (Phase 7). Each agent commits its output before the next begins.

Phase 7 is a loop: implement → validate → review (mandatory gate) → commit.

## Pipeline

The plan workflow runs as a 3-agent pipeline. Each agent operates in a separate session with bounded context — it reads only its input artifact, not prior conversation history.

| Agent | Phases | Input | Output (committed) |
|-------|--------|-------|-------------------|
| Design | 1-4 | User request + codebase | Design doc |
| Brief | 5-6 | Design doc | Work brief with tracks |
| Execute | 7-8 | Work brief | Code + retro |

**Why three agents, not one:**

1. **Checkpoint principle**: Commit each artifact before starting the next — a late failure can't lose prior work.
2. **Bounded context**: Each phase needs different context; mixing wastes tokens on irrelevant information.
3. **Artifact-mediated handoff**: If an execution agent can't proceed from the brief alone, the brief is underspecified — fix the brief, don't compensate with conversation context.

**Invocation**: Agent 1 is invoked by `/folio plan`. Agents 2 and 3 are separate sessions — the user starts them after reviewing the prior agent's committed output.

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

Each session produces a handoff document (see plan-design.md Session Exit) that bridges to the next. The handoff identifies what's strong, what's weak, and what the next round should focus on. Exit criteria should be stated explicitly so the loop has a defined endpoint.

**Key principle**: Each round should deepen specific areas, not repeat broad exploration. If round 1 produced a strong architecture but weak migration plan, round 2 focuses agent teams on migration feasibility — don't re-derive the architecture.

Handoff docs are the connective tissue. They MUST be complete enough that a fresh session can pick up without reading conversation history. See plan-design.md Session Exit for the mandatory handoff template.

## Invocation

- `/folio plan` — Infer topic from context, default lenses (pragmatic vs thorough)
- `/folio plan <topic>` — Explicit topic, default lenses

Custom lenses are specified naturally in the topic text (e.g., `/folio plan redesign auth, considering performance and readability`). The agent parses the user's intent and crafts lens descriptions accordingly. Default lenses (pragmatic vs thorough) apply when no lens guidance is given.

## Lightweight Mode

When design doc scope is 5 or fewer files with clear implementation, collapse Design + Brief
into a single document. The design agent writes execution-level detail directly — skip Agent 2.
Criteria: file count, scope clarity, single-repo. If ambiguous, use the full pipeline.

The combined doc must include the brief's required sections: Context, Agent Setup, Tracks,
Build & Deploy, Execution Conventions (with commit sequence), and Handoff Prompts. These
sections appear in the design doc's Execution Brief — see `references/plan-brief.md` for
section specs. The handoff prompt is the acid test: if it needs context beyond "read the doc
and execute," the doc is underspecified.

Commit checkpoints are still mandatory: the combined design+brief doc must be committed before
execution begins. Lightweight mode reduces agents, not checkpoints.

## Re-run Rule

If design doc feedback requires rethinking (not just minor edits), re-run Phases 2-4 within
Agent 1. If work brief feedback requires restructuring, re-run Phases 5-6 in a new Agent 2
session. If execution reveals a design-level flaw, escalate to the user — re-run from Phase 2
(new design) or Phase 5 (new brief) depending on severity. Do not patch committed artifacts
inline — re-run the producing agent.

**Amend-design path**: For additive scope changes where core architecture is settled: (1) describe
the amendment, (2) get explicit approval, (3) edit the design doc, (4) re-run Phase 4 review
only, (5) commit. Use only when additive — if it contradicts existing decisions, re-run from
Phase 2.

## Phase References

-> Read references/plan-design.md for Phases 1-4 (Design agent)
-> Read references/plan-brief.md for Phases 5-6 (Brief agent)
-> Read references/plan-execute.md for Phases 7-8 (Execute agent)
