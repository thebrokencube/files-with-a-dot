# What Is dendrik

This is the source of truth for what dendrik is. Other dendrik docs describe their own slice and link here for identity rather than restating it.

## Identity

dendrik is the shared foundation the dotfiles CLI tools — folio, jf, dot — are built on. Concretely, it is two things:

- **A convention contract** (`dendrik lint`) — a validator that checks CLI tools against shared conventions across three layers: Go build infrastructure, Skill agent-discovery, and Bridge integration. It keeps independently-built tools consistent.
- **A type-dispatching review framework** (`/dendrik`) — quality review of agentic documents (SKILL.md, CLAUDE.md, AGENTS.md, README.md, reference files, slash-commands) that detects each document's type and applies type-appropriate criteria.
- **A build/release provider** (`dendrik build`) — produces reproducible, version-stamped release artifacts per the [build & release convention](../../../pkg/dendrik/conventions/release.md); the `release` workflow is a thin shim over it.

The contract is enforced deterministically; the review framework is judgmental quality guidance.

Longer term, the aim is for the tools to share more building blocks rather than each reinventing them — direction, not what's built today.

## Where conventions live

dendrik's conventions are not denormalized across prose; each lives in one source of truth:

| Source | Owns |
|---|---|
| `pkg/dendrik/conventions/cli.md` | CLI conventions — exit codes, flags, output modes, command structure |
| `pkg/dendrik/conventions/skill.md` | Skill conventions — SKILL.md structure, frontmatter, progressive disclosure |
| `pkg/dendrik/conventions/release.md` | Build & release — version source (VERSION), `dendrik build`, tag scheme, immutability |
| `pkg/dendrik/conventions/contract.go` | The enforced contract — the canonical `Contract` slice every `dendrik lint` check derives from |
| `cmd/dendrik/skill/references/review-*.md` | The review framework — shared dimensions plus the per-type leaves |
| `cmd/dendrik/skill/references/contract-checks.md` | The one enumeration of the contract's checks (derived from `contract.go`) |

High-level docs (this one, README, CLAUDE.md, SKILL.md) describe the *shape* of these and point here; they don't carry the volatile specifics (exact counts, enumerated check lists), which live in `contract.go` and `contract-checks.md`.

## Not yet / deferred

To keep this honest, a few things dendrik does *not* do today:

- **The Right-Layer stance — including keeping volatile specifics out of high-level docs — is review guidance, not a `dendrik lint` check.** It's a judgmental opinion the review framework applies; an automated form is possible future work, not built.
- **MCP transport is deferred.** The CLI (text plus `--json`) is the canonical agent interface; an MCP surface would be a later additive adapter.
