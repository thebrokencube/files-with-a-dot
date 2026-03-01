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

**Two layers**: The CLI (`folio` binary) handles deterministic operations (validate, status, init, home). Claude workflows handle creative operations (plan, compose, review, add-pending). Each workflow's full instructions live in a reference file — read only what you need.

## Quick Orientation

Before handling any folio request, check for a folio.yml in the current directory (or use `--folio PATH`).

| What you find | What's available |
|---|---|
| No folio.yml | No folio infrastructure needed. |
| folio.yml with local outputs only | Local composition targets. |
| folio.yml with `external:` outputs | External system integration via co-located `tooling.yml`. |

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

### /folio add-pending

Add item to the `pending` list in folio.yml.

1. If text provided with command, use it. Otherwise ask.
2. Read folio.yml, locate or create `pending:` list
3. Append new item string
4. Write with targeted editing (don't reformat the whole file)

### CLI Pass-Throughs

These slash commands run the corresponding CLI command and report results:

| Command | Runs |
|---|---|
| `/folio setup` | `folio setup` |
| `/folio status` | `folio status` (mention `/folio compose` if stale targets exist) |
| `/folio validate` | `folio validate` |
| `/folio init` | `folio init --name "Name"` (ask for name if not provided) |
| `/folio gather <url>` | `folio gather <url>` (add `--materialize` or `--name` as needed) |
| `/folio home <cmd>` | `folio home <subcommand>` — run `folio home --help` for available commands |

If any CLI command fails, run `folio setup --check` first.

### Git Operations for ~/.folio

All git operations on `~/.folio` MUST use `folio home` subcommands (`push`, `pull`, etc.) — never raw `git add`, `git commit`, or `git push`. The CLI enforces conventional commit validation and handles remote sync.

## Tooling Resolution

External outputs resolve their push/pull method from `tooling.yml` (co-located with this skill file). Read `external:` from the target output, look up that system in tooling.yml, get the `pull`/`push` methods.

**Method types**: `cli:<tool>` = shell command, `mcp:<server>` = MCP tool call, `manual` = present to user, `manual:<hint>` = manual with guidance. Unlisted systems: pull=skip, push=manual.

> Jira push pipeline and other publish methods: see references/publish.md.

## Transform Types

The `transform` field on targets and tree nodes is a semantic hint for Claude, not a code branch. The CLI validates that the value is one of the allowed types but does not alter behavior based on which type is used.

| Type | Intent | Example |
|------|--------|---------|
| `distill` | Condense sources into a shorter, focused output | Thread analysis -> workflow guide |
| `extract` | Pull specific information out of broader sources | Spec -> API reference table |
| `adapt` | Reshape content for a different audience or format | Internal plan -> Jira epic description |
| `compose` | Combine multiple sources into a unified whole | Multiple plans -> initiative overview |

## Reference Files

- **references/gather.md** — Gather workflow: URL scaffold, materialize, deep research mode
- **references/compose.md** — Compose workflow: steps, tree targets, batch targets, iteration loop
- **references/publish.md** — Publish workflow: tooling resolution, Jira push pipeline, other targets
- **references/review.md** — Review workflow: steps, output format, cross-reference checks
- **references/plan.md** — Plan workflow: 7 phases, design gate, pre-commit review gate, lens system, re-run rules, agent prompt templates
- **references/stack.md** — Stack workflow: check/propagate/push actions, stacked-pr integration
- **references/schema.md** — folio.yml schema: YAML structure reference (shared across workflows)
