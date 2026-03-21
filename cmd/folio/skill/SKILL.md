---
name: folio
description: Knowledge work lifecycle — plan, compose, review. Manages project
  structure via folio.yml with source-to-target composition and diverge-converge
  planning for non-trivial tasks.
user_invocable: true
argument-hint: "[gather|plan|compose|publish|review|status|...] [args]"
---

# Folio

Lifecycle toolkit for knowledge work. Local source files compose into external targets (Jira descriptions, Google Docs, specs). `folio.yml` declares structure; status is derived from file mtimes.

**Two layers**: The CLI (`folio` binary) handles deterministic operations (validate, status, init, home). Claude workflows handle creative operations (plan, compose, review, observe). Each workflow's full instructions live in a reference file — read only what you need.

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
2. Present active projects as a compact numbered list:
   ```
   Active projects:
     1. files-with-a-dot          (1 target, 29 observations)
     2. app-benefits Structure    (7 targets, 1 observation)
     ...
   ```
   Highlight projects with high observation counts or many targets.
3. Ask: **"Which project? (number or name — or a command like `plan`, `compose`)"**
4. When the user picks a project, use the **Path** column from `folio home list` output to resolve the folio.yml location:
   - Active projects live at `~/.folio/active/<path>/folio.yml`
   - Archived projects live at `~/.folio/archive/<path>/folio.yml`
   - Run `folio status --folio ~/.folio/active/<path>/folio.yml`
   - Suggest next actions based on what's stale, has observations, or is ready to compose/publish

If the user's ARGUMENTS text doesn't match any known subcommand but isn't empty, treat it as freeform discussion about the folio system — answer the question directly.

## Workflows

### /folio gather [url|topic]

Bring sources into the folio. CLI scaffolds source entries from URLs; skill mode does deep research on a topic.

-> Read references/gather.md for full workflow (includes CLI flags and skill research mode).

### /folio plan [topic]

Multi-agent diverge-converge planning for non-trivial tasks. Use instead of EnterPlanMode for multi-file changes, architectural decisions, or unclear requirements. Skip for trivial single-file fixes.

Custom lenses can be specified naturally in the topic text; defaults to pragmatic vs thorough.

-> Read references/plan.md for full workflow (includes agent prompt templates).

### /folio compose [target]

Compose sources into targets in DAG order. Composition is creative assembly — sources are working memory; targets are communication condensed for their audience.

Previously: `/folio compile`

-> Read references/compose.md for full workflow. See references/schema.md for folio.yml structure.

### /folio publish [target]

Send composed output to external systems (Jira, Google Docs, Slack). Resolves push method from tooling.yml.

-> Read references/publish.md for full workflow (includes Jira push pipeline).

### /folio stack [check|propagate|push]

Unified stack management — morning standup view, propagation, and push in one workflow. Bridges folio topology (`dag --branches`) with stacked-pr skill mechanics.

Default action is `check`. Requires targets with `branch` fields in folio.yml.

-> Read references/stack.md for full workflow (includes check/propagate/push actions).

### /folio review [scope]

Project health check — like `git status` for the compilation system. Reports status without fixing anything.

Scope: no arg or `local` = local checks only. `external` = also fetch and compare. Specific target ID = just that target.

Previously: `/folio audit`

-> Read references/review.md for full workflow.

### /folio observe

Manage observations in folio.yml. **Do NOT edit the `observations:` list in folio.yml by hand** — not to add, remove, reorder, or reformat entries. All mutations go through CLI commands, which enforce format validation. `folio home push` runs lint as a gate and will reject malformed observations.

Workflow:
1. `folio observe types` — get valid types and descriptions
2. `folio observe list --json` — check existing scopes (avoid near-duplicates)
3. `folio observe 'type(scope): description'` — append (validates format)
4. `folio observe resolve <#N|substring>` — delete resolved items
5. `folio observe lint` — check format + inline path refs

If no text provided with the command, ask the user for the observation.

### CLI Pass-Throughs

These slash commands run the corresponding CLI command and report results:

| Command | Runs |
|---|---|
| `/folio setup` | `folio setup` |
| `/folio status` | `folio status` (mention `/folio compose` if stale targets exist) |
| `/folio validate` | `folio validate` |
| `/folio init` | `folio init --name "Name"` (ask for name if not provided) |
| `/folio gather <url>` | `folio gather <url>` (add `--materialize --type <type>` or `--name` as needed) |
| `/folio new <type> <topic>` | `folio new <type> <topic>` — scaffold typed artifact at correct path |
| `/folio health` | `folio health` — project health report (types, naming, observations) |
| `/folio home <cmd>` | `folio home <subcommand>` — run `folio home --help` for available commands |
| `/folio jira <cmd>` | `folio jira <subcommand>` — Jira pipeline: compile, push, create, view, search |

**Flag ordering**: The folio CLI uses Go's `flag` package, which requires flags **before** positional arguments. `folio new --folio my-project spike topic` works; `folio new spike topic --folio my-project` silently ignores `--folio`. This applies to all commands.

If any CLI command fails, run `folio setup --check` first.

### Git Operations for ~/.folio

All git operations on `~/.folio` MUST use `folio home` subcommands (`push`, `pull`, etc.) — never raw `git add`, `git commit`, or `git push`. The CLI enforces conventional commit validation and handles remote sync.

## Tooling Resolution

External outputs resolve their push/pull method from `tooling.yml` (co-located with this skill file). Read `external:` from the target output, look up that system in tooling.yml, get the `pull`/`push` methods.

**Method types**: `cli:<tool>` = shell command, `mcp:<server>` = MCP tool call, `manual` = present to user, `manual:<hint>` = manual with guidance. Unlisted systems: pull=skip, push=manual.

**Jira routing**: Use `folio jira` for ALL Jira operations — reads, writes, searches, everything. Never call MCP Jira tools directly. If `folio jira` cannot accomplish an operation, hard-stop and ask the user before falling back to MCP. Jira push pipeline and other publish methods: see references/publish.md.

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
| gather (skill) | Soft | Before file write | Proposed filename, length, 3 key facts |
| stack push | Hard | Before push | Branches, local vs remote tips, force-with-lease |
| stack propagate | Soft | After propagation | Rebased branches, conflicts resolved, stale remainder |
| stack check | None | — | Read-only |
| review | None | — | Read-only |
| plan | Hard | Phase 6 pre-commit | Already defined in plan.md |

## Materialization Invariants

Every workflow phase that produces knowledge materializes it as a typed artifact
before the next phase begins. This is enforced by the skill, not optional.

| Workflow | Phase | Artifact | Command |
|----------|-------|----------|---------|
| gather | deep research | reference file | `folio new <inferred-type> <topic>` |
| plan | Phase 1 research | spike(s) | `folio new spike <topic>` |
| plan | Phase 4b design | design doc | `folio new design <topic>` (colocates with work dir if topic matches) |
| plan | Phase 7 retro | retro file | `folio new retro <topic>` (colocates with work dir if topic matches) |

"Materialized" means: file exists on disk, registered in folio.yml, committed
via `folio home push`. Agent memory and conversation context are ephemeral —
they do not count as materialization.

## Reference Files

- **references/gather.md** — Gather workflow: URL scaffold, materialize, deep research mode
- **references/compose.md** — Compose workflow: steps, tree targets, batch targets, iteration loop
- **references/publish.md** — Publish workflow: tooling resolution, Jira push pipeline, other targets
- **references/review.md** — Review workflow: steps, output format, cross-reference checks
- **references/plan.md** — Plan workflow: pipeline overview, phase routing, lightweight mode, re-run rules
- **references/stack.md** — Stack workflow: check/propagate/push actions, stacked-pr integration
- **references/schema.md** — folio.yml schema: YAML structure reference (shared across workflows)
