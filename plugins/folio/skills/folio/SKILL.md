---
name: folio
description: "Use when planning non-trivial tasks, composing outputs, or managing
  knowledge work projects. Lifecycle toolkit with folio.yml-driven source-to-target
  composition and diverge-converge planning. Trigger on: design doc, spike, retro,
  wrap-up, observations, handoff, brief, gather, publish, burndown, work tracks."
---

# Folio

Lifecycle toolkit for knowledge work. Local source files compose into external targets (Jira descriptions, Google Docs, specs). `folio.yml` declares structure; freshness is derived from recorded input digests when available, with `unknown` reported honestly when no complete snapshot exists.

## Setup

`folio` needs its binary installed. Run this plugin's `bin/setup` once to install it (and again after a
plugin update). It is idempotent — safe to re-run, and a no-op when the pinned version is already installed.

**Two layers**: The CLI (`folio` binary) handles deterministic operations (validate, status, init, home). Agent workflows handle creative operations (plan, compose, observe). Each workflow's full instructions live in a reference file — read only what you need.
Freshness snapshots are recorded by `folio touch <target>` after a composition and review. The
command stores input digests without rewriting outputs. Add `--final` only when the target has a
direct external output and an existing local review copy; final targets are terminal for Folio
propagation, but their local stale or missing state remains visible.

**Phase markers**: Mark each phase transition with a **single line** — no rationale or restatement of the goal. Example: `Phase 2: propose (pragmatic + thorough)`.

## Quick Orientation

Before handling any folio request, check for a folio.yml in the current directory (or use `--folio PATH`).

| What you find | What's available |
|---|---|
| No folio.yml | No folio infrastructure needed. |
| folio.yml with local outputs only | Local composition targets. |
| folio.yml with `external:` outputs | External system integration via co-located `tooling.yml`. |

**Multi-store (container)**: `FOLIO_UMBRELLA` names the plain control-plane directory that owns `stores.yml`; `FOLIO_HOME` names the content work root for this invocation. The registry lists independent stores plus `default:`. A bare project lookup prefers an explicit Folio work root, then a cwd-owned Folio store, then the configured default; `folio stores list --json` fans out across registered stores, `--folio <store>:<project>` selects a Folio store, and `folio home push/pull [<store>]` syncs one store at a time. A compatibility alias accepts an umbrella-valued `FOLIO_HOME` with a warning. Absent `stores.yml`, the legacy isolated-home behavior remains.

## Lifecycle Model

Folio tracks a knowledge lifecycle:

```
observation -> spike -> sketch -> design -> plan[tracks] -> implementation -> retro
     ^                      |                                                 |
     |                      '- N==1 lightweight: sketch -> implementation      |
     '---------------------- findings feed back ------------------------------'
```

**Lifecycle types** progress through stages: observation, spike, sketch, design, plan, track, retro.
**References** (labels: research, insight, guide, domain, review) feed in at any stage.
**Outputs** are composed artifacts for external systems.

**Two-tier residency**: Lifecycle types always stay project-scoped. References that prove cross-cutting promote to the selected Folio store's `vault/<label>/` — a shared knowledge layer outside any project, folio-local to that store (never a global/registered store). Source paths use the `vault:` prefix (e.g., `vault:research/2026-03-01-comparable-dvc.md`), resolved from the owning project store or the selected content work root for store-only operations. The vault has no folio.yml — its directory structure is its index.

`folio status` shows a lifecycle summary header with counts per stage.

| From | To | Trigger |
|------|------|---------|
| observation | spike | "I should investigate this" |
| spike | sketch | Lay out the whole shape as a birds-eye to react to (or enter sketch directly) |
| sketch | design | Sketch frozen AND N≥2 tracks — harden into a design doc |
| sketch | implementation | N==1 (lightweight) — single track + commits, skip design + brief |
| design | plan | Design approved, ready to execute |
| plan | implementation | Tracks created from plan |
| implementation | retro | Work complete or paused |
| retro | observation | Findings feed back into new observations |

## Bare Invocation

When `/folio` is called with no subcommand (ARGUMENTS is empty, missing, or just freeform discussion):

1. Run `folio home list` to get the project dashboard
2. Read `references/lifecycle.md` and present active projects in deterministic manifest declaration order. Derive lifecycle stage from authored artifact presence; do not infer recency from directory metadata.
3. Ask: **"Which project? (number or name — or a command like `plan`, `compose`, `wrap-up`)"**
4. When the user picks a project, use the **Path** column from `folio home list` with the selected content root; do not construct paths from `~/.folio`:
   - Pass the returned path to `--folio` when it is already a file path.
   - For a session workspace, keep using its literal path via `FOLIO_HOME=<workspace>`.
   - Run `folio status --folio <selected-root>/active/<path>/folio.yml` only when the selected root is known.
   - **Resume summary (always print after status).** Folio state IS the handoff — no
     handoff files exist. The next-session pickup happens by surfacing the most current
     state inline. After `folio status`, do the following and display each as a labeled
     block:
     a. **Active work tracks**: Enumerate `work/active/*/` directories in the order
        established by the project's manifest source declarations; append unreferenced
        tracks in lexical path order. For each:
        - Derive the lifecycle stage from authored artifact presence.
        - Print the evidence (spike, sketch, design, plan, implementation, or retro)
          instead of a directory-metadata age label.
     b. **Design state** (if any track has one): Read the first design doc in that
        deterministic track order and print these sections verbatim under
        "Design state — `<relative-path>`":
        - `## Pinned Constraints`
        - `## Open Questions`
        - `## Convergence Status`
        - `### Non-Negotiable Constraints` (under `## Direction`)
        If the doc predates this template and lacks those exact section names, fall back
        to reading `## Direction → ### Scope Boundary` + `## Convergence Status` and
        flag the doc as "uses older template — consider migrating."
     c. **Latest round** (if any): Highest-numbered `<track>/agent-research/NNNN-round/`
        directory for the surfaced track. List filenames inside it (e.g.
        `converged.md`, `review-*.md`, lens files).
     d. **First spike** (if no design doc): If (b) yielded nothing, list the first
        active spike from the same deterministic track order.
     e. **Open observations**: Run `folio observe list` (or read folio.yml).
     f. **Multi-track note**: If more than one work track is active, explicitly remind
        the user: "Multiple work tracks active. Surfaced state follows manifest
        declaration order. Switch with: `/folio <project> <track-slug>` (not yet
        supported as a flag — for now ask the user which to focus on)."
     -> Read references/lifecycle.md for derivation rules, digest freshness, deterministic session entry display, artifact routing, schema migration hints, and fallback behavior.
     `Next: [lifecycle-derived action] — [brief rationale]`
   - If the user provides work items alongside the pick, route them through the artifact routing guidance in each derivation rule

     -> Read references/lifecycle.md for derivation rules, stale detection, artifact routing, session entry display, schema migration hints, and fallback behavior.

**Why the resume summary matters.** Sessions used to end with a `/tmp/handoff-*.md` file
that the next session had to be told to read. That's been removed — see plan-design.md
Session Exit. The price of removing handoff files is that bare invocation MUST surface
enough current state to pick up cold. Do not skip the resume summary.

If the user's ARGUMENTS text doesn't match any known subcommand but isn't empty, check in order: (1) if it matches "wrap-up", route to `/folio wrap-up`; (2) if it's a knowledge lookup (keywords like "find", "search", "look for", "stuff about", "anything on"), route to `/folio find` with the extracted query; (3) otherwise, treat it as freeform discussion about the folio system — answer the question directly.

## Workflows

### /folio find <query>

Search across folio knowledge for a topic. Vault-first, then current project, then all active projects.

-> Read references/find.md for full workflow (search order, output format).

### /folio gather [url|topic|path]

Bring sources into the folio. Two skill shapes: **snapshot** (`/folio gather <topic>`) captures new knowledge; **re-seed** (`/folio gather <existing-file-path>`) updates existing research. CLI scaffolds source entries from URLs.

-> Read references/gather.md for full workflow (includes CLI flags, snapshot, and re-seed).

### /folio plan [topic]

Multi-agent diverge-converge planning for non-trivial tasks. Use instead of EnterPlanMode for multi-file changes, architectural decisions, or unclear requirements. Skip for trivial single-file fixes.

Custom lenses can be specified naturally in the topic text; defaults to pragmatic vs thorough.

-> Read references/plan.md for full workflow (includes agent prompt templates).

### /folio dispatch

Dispatch a selected track to a named receiver by reading only its committed work-plan README and rendering the transient, provider-neutral dispatch projection. The committed work-plan README is the only semantic handoff input; enforce the pre-dispatch gate, show visible refusal when a required condition is unresolved, and do not read the design, sketch, spike, transcript, prompt file, or implicit parent context. Never replace bare `/folio` resume.

-> Read references/plan-brief.md for the full dispatch projection and pre-dispatch gate.

### /folio compose [target]

Compose sources into targets in DAG order. Composition is creative assembly — sources are working memory; targets are communication condensed for their audience.

-> Read references/compose.md for full workflow. See references/schema.md for folio.yml structure.

### /folio publish [target]

Send composed output to external systems (Jira, Google Docs, Slack). Resolves push method from tooling.yml.

-> Read references/publish.md for full workflow (includes Jira push pipeline).

### /folio lint [project]

Periodic knowledge integrity pass. Two-layer scan (CLI deterministic + LLM semantic) across active projects, producing a structured cleanup plan before acting. Findings tracked in a persistent `folio-hygiene` project. Scoped mode (`/folio lint <project>`) runs on a single project.

-> Read references/lint.md for full workflow (phases, scoped mode, guardrails).

### /folio observe

Manage the open-items queue in folio.yml — bugs, gaps, ideas, debt, tasks. All mutations go through CLI commands (never hand-edit). Includes type disambiguation via the alignment protocol.

-> Read references/observe.md for full workflow (CLI commands, type disambiguation, alignment routing).

### /folio wrap-up

End-of-session workflow: retro, archive completed tracks, create successor tracks, update folio state, workspace cleanup. Steps are safe to abandon mid-flow; the one always-captured output is the retro (see references/wrap-up.md).

-> Read references/wrap-up.md for full workflow.

### CLI Quick Reference

The `folio` binary handles all deterministic operations. Run `folio --help` for the full list. Key commands by category:

**Data** — query project state (read-only, safe to run anytime):

| Command | Purpose |
|---|---|
| `folio validate` | Check folio.yml structural integrity |
| `folio status` | Derive and display target state (mention `/folio compose` if stale) |
| `folio stale` | List stale/missing/unknown targets |
| `folio dag` | Show target dependency graph |
| `folio health` | Project health report (types, naming, observations) |
| `folio stores list [--json]` | List registered stores (multi-store registry); `find` fans out across these |

**Composition** — create and manage artifacts:

| Command | Purpose |
|---|---|
| `folio new <type> <topic>` | Scaffold typed artifact (`--dry-run` to preview). Vault types: `vault:research`, `vault:domain`, `vault:guide`, `vault:insight` |
| `folio gather <url>` | Add source entry from URL (`--materialize --type <type>` or `--name` as needed) |
| `folio touch <target>` | Record input-digest freshness without rewriting outputs; `--final` marks a reviewed terminal target |
| `folio observe 'type(scope): description'` | Add observation (auto-syncs: pull + push). Types: `idea`, `gap`, `bug`, `debt`, `task`. Use `--no-sync` to skip |
| `folio observe list` | List all observations (add `--json` for structured output) |
| `folio observe resolve "#N" "#N2" ...` | Resolve by index (auto-syncs). **Batch multiple in one call** to avoid index shift. Use `--no-sync` to skip |
| `folio observe types` | Show valid types and descriptions |
| `folio observe lint` | Check format and inline path refs |
| `folio archive` | Move work track from active to archive |

**Management** — setup and home operations:

| Command | Purpose |
|---|---|
| `folio init --name "Name"` | Bootstrap new folio.yml (`--path` overrides the auto-derived slug) |
| `folio setup` | Check folio dependencies (`--check` for non-interactive) |
| `folio home <cmd>` | Folio content-root operations (list, push, pull, archive, activate, health, workspace) |
| `folio home workspace list` | List jj workspaces (one per active Claude session) |
| `folio home workspace create` | Manually create a jj workspace |
| `folio home workspace cleanup [path]` | Remove a workspace — errors if unpushed changes exist |

**Fleet** — the code and dotfiles repos folio hovers over (read-only status; isolated checkouts):

| Command | Purpose |
|---|---|
| `folio fleet status [--dirty] [--json]` | Read-only branch + dirty state across every registered store, grouped by kind |
| `folio fleet workarea open <store> <branch>` | Create an isolated checkout at `<umbrella>/.worktrees/<store>/<slug>` (`-b` to fork off a base other than the store's default branch) |
| `folio fleet workarea list` | Every work area — folio-placed, plus hand-made ones read from `jj workspace list` / `git worktree list` |
| `folio fleet workarea reap [--force]` | Remove folio-placed areas (tier-correct; keeps dirty/unpushed) |

Folio never commits or opens a PR for a code store — `folio home push <code-store>` positions the branch and hands off to `/commit`.

Some commands have corresponding skill workflows that add creative/judgmental work on top: `gather` (snapshot/re-seed), `observe` (type disambiguation via alignment).

**Flag ordering**: Flags go **before** positional arguments. `folio new --folio my-project spike topic` works; `folio new spike topic --folio my-project` does not.

If any CLI command fails, run `folio setup --check` first.

## Isolation and Cleanup

Use a work area for code and a temporary workspace for a Folio store; never create either beside or inside its source checkout. Fetch before branching. Do not enter another workspace or rewrite its working-copy change. Remove a work-area or workspace registration through the owning CLI instead of deleting its directory.

The dotfiles repository is the exception: create its temporary jj workspace under `/tmp/fwad-<topic>`, work there, push its bookmark from that workspace, and create a PR from the colocated checkout.

For a repository or workspace you do not own, use `jj -R <repo-path>` with `--ignore-working-copy`; do not enter it. In your own work area, run commands from the target repository. Never use command substitution, backticks, `git -C`, `--git-dir`, or `--work-tree`.

## Terminology Note

**"workspace"** in folio CLI always means a **jj workspace** of a Folio store — a session-isolated checkout under `/tmp/folio-ws-<id>`. It is NOT a synonym for a folio project. To list folio projects, use `folio home list`. To list jj workspaces, use `folio home workspace list`.

**"work area"** is the code-side counterpart: an isolated checkout of a `code` store, at `<umbrella>/.worktrees/<store>/<slug>`. Created with `folio fleet workarea open`, not `folio home workspace create`.

## Session Lifecycle

Before any folio operation, establish the content root for this session:

1. In legacy isolated-home mode, the configured `FOLIO_HOME` is the content root. In container mode, `FOLIO_UMBRELLA` identifies the registry and `FOLIO_HOME` may identify an existing session workspace.
2. If you do not already have a workspace path, run `folio home workspace create` and capture the printed path. From an umbrella, this selects the configured default Folio store; a code or dotfiles cwd never becomes a Folio workspace.
3. **Store the path as a literal string** — do not rely on an exported variable across separate tool calls.
4. If create fails: surface the error, do not proceed.

After creation, use the captured workspace path in two ways:

- **Bash calls**: prefix each command with both roots when known, for example `FOLIO_UMBRELLA="$HOME/.folio" FOLIO_HOME=/tmp/folio-ws-123 folio status`.
- **Edit/Read/Write tools**: use the literal path, for example `/tmp/folio-ws-123/active/my-project/folio.yml`. Never use `$FOLIO_HOME` in tool file paths — those tools do not expand shell variables.

**Push and pull behavior:**
- `folio home push` rebases @ onto `main` before setting the bookmark. Concurrent sessions cannot cause bookmark divergence. On content conflict: errors with instructions to resolve.
- `folio home pull` fetches + rebases if a remote exists; rebases onto local `main` if not.

**Path convention:** Always use the captured workspace path or an explicit `--folio` path. Never hardcode `~/.folio/`; the umbrella is a control root, not a project root.

**Mandatory cleanup at session end:**
Before ending a session that used folio, run:
```
FOLIO_UMBRELLA="$HOME/.folio" FOLIO_HOME=<path> folio home workspace cleanup <path>
```
This errors if unpushed changes exist (run `folio home push` first), then removes the workspace. Do not skip this step — leaked workspaces are reaped after 2 days, but clean exit is preferred.
Sessions that never invoked `/folio` have nothing to clean up.

**Only clean up your own workspace.** `folio home workspace list` shows all workspaces but does not indicate which ones belong to active sessions. Cleaning up another session's workspace will break that session. Never run cleanup on a workspace you did not create in this session.

### Git Operations for a Folio work root

All git or jj operations on a Folio content work root MUST use `folio home` subcommands (`push`, `pull`, etc.) — never raw `git add`, `git commit`, or `git push`. The CLI enforces conventional commit validation and handles remote sync.

**Jira operations**: Use the `/jf` skill for all Jira work — push, sync, view, search, create.

## Tooling Resolution

External outputs resolve their push/pull method from `tooling.yml` (co-located with this skill file). Read `external:` from the target output, look up that system in tooling.yml, get the `pull`/`push` methods.

**Method types**: `cli:<tool>` = shell command, `mcp:<server>` = MCP tool call, `manual` = present to user, `manual:<hint>` = manual with guidance. Unlisted systems: pull=skip, push=manual.

**Manual methods are inviolable.** When a target's publish method resolves to `manual` or `manual:<hint>` (e.g., `manual:paste-from-markdown`), never substitute an MCP tool or API call to push content programmatically. Compose the output file, then present it to the user (or copy to clipboard). This applies even if an MCP tool exists that could technically write to the target system — the method type is an explicit choice about how content reaches that system, not a limitation to work around.

**Jira routing**: Use `jf` (Jira Forest CLI) for ALL Jira operations — push, sync, view, search, create. Never call MCP Jira tools directly for writes. If `jf` cannot accomplish an operation, hard-stop and ask the user before falling back to MCP. Jira push pipeline and other publish methods: see references/publish.md.

## Review Gates

Two gate types, proportional to risk:

| Type | Behavior | Used when |
|------|----------|-----------|
| **Hard** | Stop. Present summary. Require explicit "yes" to proceed. | Destructive/external-facing operations |
| **Soft** | Present summary. Proceed unless user objects. | Local/reversible operations |

Some Hard gates are fronted by an **automated fresh-subagent QA/review loop** that runs to completion
before the gate presents — the Phase 1.5 sketch reviewer and the Phase 4b adversarial design review.
That loop is an internal fix-loop, not a user gate; the Hard gate is the human sign-off that follows.

### Gate placement

| Workflow | Gate | Placement | What's shown |
|----------|------|-----------|-------------|
| publish | Hard | Before each push | Target, system, method, first 5 lines |
| compose | Soft | After composition loop, before final status | Targets composed, paths, sizes (cap 5) |
| gather (snapshot) | Soft | Before file write | Proposed filename, length, 3 key facts |
| gather (re-seed) | Soft | Before file update | Summary of changes to existing file |
| plan | Hard | Phase 1.5 idea/arch freeze | Rendered birds-eye sketch page (show it, don't summarize) |
| plan | Hard | Phase 4b pre-commit | Review design doc before commit |
| plan | Hard | Phase 6 Test Strategy | Test strategy section, user approval required |
| plan | Hard | Phase 6 pre-commit | Already defined in plan.md |
| dispatch | Hard | Before any provider invocation | Selected target, landed brief, required context, named receiver, or visible refusal |
| lint | Hard | Before Phase 3 execution | Cleanup plan summary, action count by effort level |
| lint | Hard | Before cross-project mutation | Proposed command, target project |
| lint | Soft | At 20 observations | Finding count, confirmation to continue |
| wrap-up | Hard | Before archiving (step 3) | List of tracks to archive, batch confirmation |

## Materialization Invariants

Every workflow phase that produces knowledge materializes it as a typed artifact
before the next phase begins. This is enforced by the skill, not optional.

| Workflow | Phase | Artifact | Command |
|----------|-------|----------|---------|
| gather | snapshot (Shape A) | reference file | `folio new <inferred-type> <topic>` |
| gather | re-seed (Shape C) | updated reference file | edit existing file |
| plan | Phase 1 research | spike(s) | `folio new spike <topic>` |
| plan | Phase 1.5 idea/arch | sketch HTML page | `folio new sketch <topic>` (creates work dir, colocates) |
| plan | Phase 4b design | design doc | `folio new design <topic>` (creates work dir if none exists, colocates inside it) |
| plan | Phase 7 retro | retro file | `folio new retro <topic>` (colocates with work dir if topic matches) |

"Materialized" means: file exists on disk, registered in folio.yml, committed
via `folio home push`. Agent memory and conversation context are ephemeral —
they do not count as materialization.

## Source Declaration Checklist

Before creating or updating any folio artifact (design doc, spike, retro, brief, idea/arch sketch), verify:

1. **List sources read.** Enumerate every file, vault entry, or external reference you
   consumed to produce this artifact. If you cannot list them, you haven't done the work.
   Route each source to its correct type: transient → observation; investigation → spike;
   external knowledge → vault research; plan-level detail → brief.
2. **Register in folio.yml.** Every source must appear in `sources:` with a `depends_on`
   pointing to the artifact being created. Use `folio home push` to commit.
3. **Verify sources exist.** Every path in `depends_on` must resolve to a real file.
   Run `folio validate` to catch broken references.
4. **No orphan synthesis.** If the artifact synthesizes from conversation context alone
   (no file sources), STOP — materialize the source knowledge first (spike, gather),
   then create the artifact from the materialized source.

This checklist applies at every materialization point in the table above. Skipping it
is how provenance chains break.

## Reference Files

- **references/blueprint.md** — folio's blueprint: which dendrik concepts folio composes (+ honest conformance); ids resolve in dendrik's building-blocks map
- **references/find.md** — Find workflow: vault-first search order, tiered scope expansion, output format
- **references/lint.md** — Lint workflow: periodic knowledge integrity pass, two-layer scan, cleanup plan, folio-hygiene project
- **references/observe.md** — Observe workflow: CLI commands, type disambiguation, alignment routing
- **references/alignment.md** — Alignment protocol: claim-first questioning, invocation contract, confidence-based exit
- **references/adversarial-review.md** — Cross-cutting principle: every subjective judgment needs pushback (3 tiers: self-challenge, adversarial prompt, parallel adversarial)
- **references/gather.md** — Gather workflow: URL scaffold, snapshot (Shape A), re-seed (Shape C), phase structure
- **references/compose.md** — Compose workflow: steps, forest targets, batch targets, iteration loop
- **references/publish.md** — Publish workflow: tooling resolution, Jira push pipeline, Notion templates, other targets
- **references/notion-proposal-template.md** — Default Notion template: feedback table with reviewer stance/comments
- **references/plan.md** — Plan workflow: pipeline overview, phase routing, lightweight mode, re-run rules
  - **references/plan-idea.md** — Plan Phase 1.5 (idea/arch sketch): HTML-first birds-eye page, visual vocabulary, fresh-subagent reviewer gate, sign-off gate, lightweight track-count decision
  - **references/plan-design.md** — Plan Phases 1-4 (Design agent): understand, propose, converge, fill/review design doc
- **references/lifecycle.md** — Lifecycle derivation: type/status-based suggestions, digest freshness, deterministic session entry display, artifact routing, schema migration hints
  - **references/plan-execute.md** — Plan Phases 7-8 (Execute agent): implement per track, retro
- **references/schema.md** — folio.yml schema: YAML structure reference (shared across workflows)
- **references/progressive-disclosure.md** — Cross-cutting principle: action first, context second, history last. Applied to briefs, handoffs, compose outputs
- **references/wrap-up.md** — Wrap-up workflow: session-end retro, archive, successor tracks, handoff doc
- **references/testing.md** — Integration testing: FOLIO_HOME-isolated test loops, setup/teardown patterns
- **references/burndown.md** — Burndown execution: wave-based batch work with ratchet, checkpoints, and flywheel learning
- **references/migrate.md** — Migration guide: moving lifecycle artifacts from reference/ to work/, classification framework, cluster moves, per-artifact checklist
- **references/container-migration.md** — Container migration runbook: demote single-home `~/.folio` into the multi-store umbrella (clone-beside, push gate, two-mv swap, dotfile-managed stores.yml, second-machine bootstrap, rollback). Paired with `cmd/folio/scripts/migrate-container.sh`
