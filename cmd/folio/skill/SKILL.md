---
name: folio
description: "Use when planning non-trivial tasks, composing outputs, or managing
  knowledge work projects. Lifecycle toolkit with folio.yml-driven source-to-target
  composition and diverge-converge planning. Trigger on: design doc, spike, retro,
  wrap-up, observations, handoff, brief, gather, publish, burndown, work tracks."
user_invocable: true
argument-hint: "[gather|plan|compose|publish|lint|find|observe|wrap-up|status|...] [args]"
---

# Folio

Lifecycle toolkit for knowledge work. Local source files compose into external targets (Jira descriptions, Google Docs, specs). `folio.yml` declares structure; status is derived from file mtimes.

**Two layers**: The CLI (`folio` binary) handles deterministic operations (validate, status, init, home). Claude workflows handle creative operations (plan, compose, observe). Each workflow's full instructions live in a reference file — read only what you need.

**Process narration**: Before starting any multi-step workflow or phase transition, state what you're about to do and why. Example: "Starting Phase 2 — spawning two propose agents with pragmatic and thorough lenses." This prevents ambiguity about which phase you're in and lets the user course-correct before work begins, not after.

## Quick Orientation

Before handling any folio request, check for a folio.yml in the current directory (or use `--folio PATH`).

| What you find | What's available |
|---|---|
| No folio.yml | No folio infrastructure needed. |
| folio.yml with local outputs only | Local composition targets. |
| folio.yml with `external:` outputs | External system integration via co-located `tooling.yml`. |

## Lifecycle Model

Folio tracks a knowledge lifecycle:

```
observation -> spike -> design -> plan[tracks] -> implementation -> retro
     ^                                                                |
     '---------------------- findings feed back ----------------------'
```

**Lifecycle types** progress through stages: observation, spike, design, plan, track, retro.
**References** (labels: research, insight, guide, domain, review) feed in at any stage.
**Outputs** are composed artifacts for external systems.

**Two-tier residency**: Lifecycle types always stay project-scoped. References that prove cross-cutting promote to `~/.folio/vault/<label>/` — a shared knowledge layer outside any project. Source paths use the `vault:` prefix (e.g., `vault:research/2026-03-01-comparable-dvc.md`) which resolves to `~/.folio/vault/`. The vault has no folio.yml — its directory structure is its index.

`folio status` shows a lifecycle summary header with counts per stage.

| From | To | Trigger |
|------|------|---------|
| observation | spike | "I should investigate this" |
| spike | design | Findings warrant a solution |
| design | plan | Design approved, ready to execute |
| plan | implementation | Tracks created from plan |
| implementation | retro | Work complete or paused |
| retro | observation | Findings feed back into new observations |

## Bare Invocation

When `/folio` is called with no subcommand (ARGUMENTS is empty, missing, or just freeform discussion):

1. Run `folio home list` to get the project dashboard
2. Derive recency and lifecycle stage per project using the Session Entry Display format in `references/lifecycle.md`. Present a recency-ranked list (most recently touched work track first), grouped into "Recently active" and "Also active but stale."
3. Ask: **"Which project? (number or name — or a command like `plan`, `compose`, `wrap-up`)"**
4. When the user picks a project, use the **Path** column from `folio home list` output to resolve the folio.yml location:
   - Active projects live at `~/.folio/active/<path>/folio.yml`
   - Archived projects live at `~/.folio/archive/<path>/folio.yml`
   - Run `folio status --folio ~/.folio/active/<path>/folio.yml`
   - Prepend a lifecycle suggestion line before the raw status output:
     `Next: [lifecycle-derived action] — [brief rationale]`
   - If the user provides work items alongside the pick, route them through the artifact routing guidance in each derivation rule

     -> Read references/lifecycle.md for derivation rules, stale detection, artifact routing, session entry display, schema migration hints, and fallback behavior.

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

End-of-session workflow: retro, archive completed tracks, create successor tracks, write handoff doc, workspace cleanup. Each step is independently safe — skip or abandon at any point.

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

**Composition** — create and manage artifacts:

| Command | Purpose |
|---|---|
| `folio new <type> <topic>` | Scaffold typed artifact (`--dry-run` to preview). Vault types: `vault:research`, `vault:domain`, `vault:guide`, `vault:insight` |
| `folio gather <url>` | Add source entry from URL (`--materialize --type <type>` or `--name` as needed) |
| `folio touch <target>` | Mark a target as current |
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
| `folio home <cmd>` | FOLIO_HOME operations (list, push, pull, archive, activate, health, workspace) |
| `folio home workspace list` | List jj workspaces (one per active Claude session) |
| `folio home workspace create` | Manually create a jj workspace |
| `folio home workspace cleanup [path]` | Remove a workspace — errors if unpushed changes exist |

Some commands have corresponding skill workflows that add creative/judgmental work on top: `gather` (snapshot/re-seed), `observe` (type disambiguation via alignment).

**Flag ordering**: Flags go **before** positional arguments. `folio new --folio my-project spike topic` works; `folio new spike topic --folio my-project` does not.

If any CLI command fails, run `folio setup --check` first.

## Terminology Note

**"workspace"** in folio CLI always means a **jj workspace** (a session-isolated checkout of `~/.folio`). It is NOT a synonym for a folio project. To list folio projects, use `folio home list`. To list jj workspaces, use `folio home workspace list`.

## Session Lifecycle

Before any folio operation, ensure a jj workspace exists for this session:

1. If `~/.folio/.jj` does not exist: skip (git-based home, no workspace needed)
2. If you already have a workspace path from this session: skip (already initialized)
3. Run `folio home workspace create` and capture the printed path (last line of output)
4. **Store the path as a literal string** — do NOT rely on env vars. Each Bash call starts
   a fresh shell, so `export FOLIO_HOME=...` does not persist between calls.
5. If create fails: surface the error, do not proceed

After init, use the captured workspace path in two ways:

- **Bash calls**: prefix every `folio` command with `FOLIO_HOME=<path>`, e.g.
  `FOLIO_HOME=/tmp/folio-ws-123 folio status`. This sets the var for that single command.
- **Edit/Read/Write tools**: use the literal path, e.g.
  `/tmp/folio-ws-123/active/my-project/folio.yml`. Never use `$FOLIO_HOME` in tool
  file_path parameters — those tools do not expand shell variables.

**Push and pull behavior:**
- `folio home push` rebases @ onto `main` before setting the bookmark. Concurrent sessions
  cannot cause bookmark divergence. On content conflict: errors with instructions to resolve.
- `folio home pull` fetches + rebases if a remote exists; rebases onto local `main` if not.

**Path convention:** Always use the captured workspace path (or the `--folio` flag) to
reference folio files — never hardcode `~/.folio/`. With jj workspace isolation, `~/.folio`
is the default workspace and may be at a different jj change than the session workspace.

**Mandatory cleanup at session end:**
Before ending a session that used folio, run:
```
FOLIO_HOME=<path> folio home workspace cleanup <path>
```
This errors if unpushed changes exist (run `folio home push` first), then removes the workspace.
Do not skip this step — leaked workspaces are reaped after 2 days, but clean exit is preferred.
Sessions that never invoked `/folio` have nothing to clean up.

**Only clean up your own workspace.** `folio home workspace list` shows all workspaces but
does not indicate which ones belong to active sessions. Cleaning up another session's workspace
will break that session. Never run cleanup on a workspace you didn't create in this session.
If the user asks about stale workspaces, list them but let the user decide which to remove.

**Workspace commands:**
- `folio home workspace create` — create a workspace (called by skill on first `/folio` invocation)
- `folio home workspace list` — list all workspaces (does NOT indicate ownership — treat unknown workspaces as potentially active)
- `folio home workspace cleanup [path]` — remove a workspace (requires empty @, only use on YOUR workspace)

### Git Operations for ~/.folio

All git operations on `~/.folio` MUST use `folio home` subcommands (`push`, `pull`, etc.) — never raw `git add`, `git commit`, or `git push`. The CLI enforces conventional commit validation and handles remote sync.

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

### Gate placement

| Workflow | Gate | Placement | What's shown |
|----------|------|-----------|-------------|
| publish | Hard | Before each push | Target, system, method, first 5 lines |
| compose | Soft | After composition loop, before final status | Targets composed, paths, sizes (cap 5) |
| gather (snapshot) | Soft | Before file write | Proposed filename, length, 3 key facts |
| gather (re-seed) | Soft | Before file update | Summary of changes to existing file |
| plan | Hard | Phase 4b pre-commit | Review design doc before commit |
| plan | Hard | Phase 6 Test Strategy | Test strategy section, user approval required |
| plan | Hard | Phase 6 pre-commit | Already defined in plan.md |
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
| plan | Phase 4b design | design doc | `folio new design <topic>` (creates work dir if none exists, colocates inside it) |
| plan | Phase 7 retro | retro file | `folio new retro <topic>` (colocates with work dir if topic matches) |

"Materialized" means: file exists on disk, registered in folio.yml, committed
via `folio home push`. Agent memory and conversation context are ephemeral —
they do not count as materialization.

## Source Declaration Checklist

Before creating or updating any folio artifact (design doc, spike, retro, brief), verify:

1. **List sources read.** Enumerate every file, vault entry, or external reference you
   consumed to produce this artifact. If you cannot list them, you haven't done the work.
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
  - **references/plan-design.md** — Plan Phases 1-4 (Design agent): understand, propose, converge, fill/review design doc
  - **references/plan-brief.md** — Plan Phases 5-6 (Brief agent): decompose tracks, write execution brief
  - **references/plan-execute.md** — Plan Phases 7-8 (Execute agent): implement per track, retro
- **references/schema.md** — folio.yml schema: YAML structure reference (shared across workflows)
- **references/progressive-disclosure.md** — Cross-cutting principle: action first, context second, history last. Applied to briefs, handoffs, compose outputs
- **references/lifecycle.md** — Lifecycle derivation: type/status-based suggestions, stale detection, session entry display, artifact routing, schema migration hints
- **references/wrap-up.md** — Wrap-up workflow: session-end retro, archive, successor tracks, handoff doc
- **references/testing.md** — Integration testing: FOLIO_HOME-isolated test loops, setup/teardown patterns
- **references/burndown.md** — Burndown execution: wave-based batch work with ratchet, checkpoints, and flywheel learning
- **references/migrate.md** — Migration guide: moving lifecycle artifacts from reference/ to work/, classification framework, cluster moves, per-artifact checklist
