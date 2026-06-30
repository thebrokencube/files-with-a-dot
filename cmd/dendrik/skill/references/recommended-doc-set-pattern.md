# Recommended Doc-Set Pattern

The shape dendrik **recommends** for a repo's documentation set. This is advice, not a check —
the cohesion checks live in `review-shared.md` (Right Layer?) and the composition pass in
`review-orchestration.md`. Reach for this when a review wants to point at a target shape rather
than just flag a single miss. Patterns attributed to the public AGENTS.md convention and
Diátaxis; adapt to repo shape (default below is **single-purpose**).

## The shape (single-purpose repo)

A cohesive set routes a reader through one arc, each doc owning one job:

| Stage | Doc | Job |
|---|---|---|
| Front door | `README.md` (humans) / `AGENTS.md` (agents) | orient in a few lines, then route to detail — never dead-end |
| Understand | `ARCHITECTURE.md`, deep guides | the model behind the code, referenced not inlined |
| Contribute | `CONTRIBUTING.md` or a guide | setup → understand → change → PR, end to end |
| Verify | a read-only `doctor` | prove the docs still match the code (see below) |

## Contributor journey

State the arc explicitly — **setup → understand → change → PR** — and make each step reachable
by link from the front door. A reader (or agent) should never have to guess the next step. Link
foundational concepts to their definition at first mention; keep a glossary reachable from the
front door, not orphaned.

## Routing-skill kernel

A thin skill (or slash-command) that fronts a guide must **route, not restate**: point at the
harness-agnostic guide and stop. The transferable kernel is *single-source + intent-routing* — a
skill earns its keep only by adding discovery or behavioral value, never by paraphrasing a doc
that already exists. (Public AGENTS.md convention: behavior and conventions live in the agnostic
layer; tool-specific files are thin adapters.)

## Single-owner edit contract

When a set is large enough that "who owns this fact?" is non-obvious, add a **Maintaining these
docs** table to the front-door/agent doc:

| Fact type | Owner | Form | Rule |
|---|---|---|---|
| (e.g. phase contract) | the one authoritative doc | prose + code anchor | others link, never restate |

Skip it for a two-file set — it is overhead there. Recommend it only once single-source findings
are already firing.

## Read-only verifier (the durable end state)

The strongest defense against doc rot is a repo-owned, **read-only** verifier (a `doctor`-style
command) that:

- reads *live* repo files (dispatch tables, lockfiles, manifests, the docs themselves),
- reports evidence + status, and **never mutates** (no installs, no writes),
- runs on every commit.

dendrik's sampled doc-claims grounding catches the cheap cases in review; a repo's own verifier
is where deeper, repo-specific grounding belongs. Recommend adopting one rather than asking
dendrik to read the whole tree.

## Form consistency

Each doc should know its Diátaxis type (tutorial / how-to / reference / explanation) and not blur
them — a how-to that drifts into explanation loses both jobs. For agentic docs the type is
usually the file (AGENTS.md, ARCHITECTURE.md, a skill); keep each one's form consistent across
the set.

## Repo-shape weighting

| Shape | Weight differently |
|---|---|
| **Single-purpose** (default) | the arc above as written |
| **Marketplace** | add ownership/placement routing (which package owns what); heavier front-door signposting |
| **Monorepo** | per-package docs scoped to the nearest `AGENTS.md`; the root routes, packages own detail |

## How this maps to dendrik's review

| Pattern element | Review home |
|---|---|
| Front door routes, doesn't dead-end | README ↔ AGENTS.md split row (`review-orchestration.md`) |
| Skill routes, doesn't restate | Skill ↔ references row |
| One owner per fact / edit contract | Right Layer? (`review-shared.md`) + Adapter-drift / CLAUDE-hoarding rows |
| Claims match code | Doc-claims grounding row; deeper → repo's own verifier |
| No conflicting facts across docs | Cross-doc contradiction row |
