# What Is dendrik

This is the source of truth for what dendrik is. Every other dendrik doc points here for identity; they describe their own slice and link back rather than restating this.

## Identity

dendrik is the **platform**: a library of well-scoped, composable primitives for agentic knowledge/CLI tooling. Consumer tools — folio, jf, dot — *compose* those primitives rather than reinventing them, so a good building block written once propagates to every consumer. The name leans into this: one root branching out into many tools.

Today that platform is expressed concretely as two surfaces:

- **A convention contract** (`dendrik lint`) — a validator that checks CLI tools against shared conventions across three layers (Go build infrastructure, Skill agent-discovery, Bridge integration), so independently-developed tools stay composable.
- **A type-dispatching review framework** (`/dendrik`) — quality review of agentic documents (SKILL.md, CLAUDE.md, AGENTS.md, README.md, reference files, slash-commands) that detects the document's *type* and applies type-appropriate criteria, because each type has a different job, audience, and dominant failure mode.

The contract is what's *enforced* deterministically; the review framework is *judgmental* quality guidance. Together they are the built, shipping expression of the platform — not its full extent.

## Primitive maturity

The platform is a catalog of recurring building blocks consumer tools keep re-implementing. Most are not yet built. The honest current state, qualitatively:

| Primitive | Maturity | What that means |
|---|---|---|
| Skill frameworks (authoring + per-type review lenses) | **Fully owned** | The review framework and authoring conventions live in dendrik and are used. |
| Convention vocabulary (the enforced contract) | **Mostly (~80%)** | The CLI/skill/bridge contract is codified and enforced; some conventions remain review-only opinions, not Go checks. |
| Template structure (typed-artifact skeletons + fill/validate) | **Partial** | Some structure is checked (doc naming, README sections), but there's no general template primitive yet. |
| Indices (generated, queryable rosters of typed artifacts) | **Absent** | Not built. |
| Lifecycle tooling (`active` / `archive` / `delete` over a folder) | **Absent** | Not built. |

This is a catalog, not a commitment — each primitive earns dendrik ownership only once it's proven well-scoped with real consumers. Do not read the absent rows as roadmap promises.

## Where conventions live

The platform's conventions are not denormalized across prose. They live in these sources of truth:

| Source | Owns |
|---|---|
| `pkg/dendrik/conventions/cli.md` | CLI conventions — exit codes, flags, output modes, command structure |
| `pkg/dendrik/conventions/skill.md` | Skill conventions — SKILL.md structure, frontmatter, progressive disclosure |
| `pkg/dendrik/conventions/contract.go` | The **enforced** contract — the canonical `Contract` slice every `dendrik lint` check derives from |
| `cmd/dendrik/skill/references/review-*.md` | The review framework — shared dimensions plus the per-type review leaves |

High-level docs (this one, README, CLAUDE.md, SKILL.md) describe the *shape* of these and point here; they do not carry the exact, volatile specifics (counts, enumerated check lists). Those live in `contract.go` and the one derived enumeration at `cmd/dendrik/skill/references/contract-checks.md`.

## Scoped out / deferred

To keep the identity honest, these are explicitly *not* part of the built platform today:

- **The Right-Layer stance is a review-local opinion, not a Go check.** dendrik's review framework recommends layering content by load model and audience (and flags volatile-specifics denormalization), but this is judgmental review guidance phrased as dendrik's opinion — not a deterministic `dendrik lint` check.
- **MCP transport is deferred and additive.** The CLI (text plus `--json`) is the canonical agent interface. An MCP surface would be a tool-specific adapter over the same command functions — a future option, not a current target.
- **An automated info-layering check is a future effort.** Codifying "high-level docs carry no volatile specifics" as an automated review/lint dimension (what counts as a volatile specific, per-type thresholds, warn-vs-fail) deserves its own scoped design and is not built.
