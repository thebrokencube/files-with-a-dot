---
name: folio
description: Knowledge work lifecycle — plan, compile, audit. Manages project
  structure via folio.yml with source-to-target compilation and diverge-converge
  planning for non-trivial tasks.
user_invocable: true
argument-hint: "[plan|compile|review|status|...] [args]"
---

# Folio

Lifecycle toolkit for knowledge work. Local source files compile into external targets (Jira descriptions, Google Docs, specs). `folio.yml` declares structure; status is derived from file mtimes.

**Two layers**: The CLI (`folio` binary) handles deterministic operations (validate, status, init, home). Claude workflows handle creative operations (plan, compile, audit, add-pending). Each workflow's full instructions live in a reference file — read only what you need.

## Quick Orientation

Before handling any folio request, check for a folio.yml in the current directory (or use `--folio PATH`).

| What you find | What's available |
|---|---|
| No folio.yml | No folio infrastructure needed. |
| folio.yml with local outputs only | Local compilation targets. |
| folio.yml with `external:` outputs | External system integration via co-located `tooling.yml`. |

## Workflows

### /folio plan [topic]

Multi-agent diverge-converge planning for non-trivial tasks. Use instead of EnterPlanMode for multi-file changes, architectural decisions, or unclear requirements. Skip for trivial single-file fixes.

Custom lenses can be specified naturally in the topic text; defaults to pragmatic vs thorough.

-> Read references/plan.md for full workflow (includes agent prompt templates).

### /folio compile [target]

Compile sources into targets in DAG order. Compilation is distillation — sources are working memory; targets are communication condensed for their audience.

-> Read references/compile.md for full workflow (includes folio.yml schema reference).

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
| `/folio status` | `folio project status` (mention `/folio compile` if stale targets exist) |
| `/folio validate` | `folio project validate` |
| `/folio init` | `folio project init --name "Name"` (ask for name if not provided) |
| `/folio home <cmd>` | `folio home <subcommand>` — run `folio home --help` for available commands |

If any CLI command fails, run `folio setup --check` first.

### Git Operations for ~/.folio

All git operations on `~/.folio` MUST use `folio home` subcommands (`push`, `pull`, etc.) — never raw `git add`, `git commit`, or `git push`. The CLI enforces conventional commit validation and handles remote sync.

## Tooling Resolution

External outputs resolve their push/pull method from `tooling.yml` (co-located with this skill file). Read `external:` from the target output, look up that system in tooling.yml, get the `pull`/`push` methods.

**Method types**: `cli:<tool>` = shell command, `mcp:<server>` = MCP tool call, `manual` = present to user, `manual:<hint>` = manual with guidance. Unlisted systems: pull=skip, push=manual.

### Jira Push Pipeline

Tree targets with `system: jira` and `compiled_ext: .json` use a three-phase pipeline:

```
source .md -> lint (md-to-adf --lint) -> precompile (md-to-adf --acli) -> compiled .json -> push (acli)
```

| Placeholder | Resolves from |
|---|---|
| `{id}` | Tree node `id` (Jira key) |
| `{source}` | Tree node `file` |
| `{compiled}` | `{compiled_dir}/{id}{compiled_ext}` |

Example:
```bash
md-to-adf --lint epic.md                                               # 1. Lint
md-to-adf --acli BEN-48284 < epic.md > compiled/jira/BEN-48284.json   # 2. Precompile
acli jira workitem edit --from-json compiled/jira/BEN-48284.json --yes # 3. Push
```

**md-to-adf limitations** (caught by `--lint`): no tables, no fenced code blocks, no blockquotes, no nested lists, no h3+. Flatten source files before compilation.

## Transform Types

The `transform` field on targets and tree nodes is a semantic hint for Claude, not a code branch. The CLI validates that the value is one of the allowed types but does not alter behavior based on which type is used.

| Type | Intent | Example |
|------|--------|---------|
| `distill` | Condense sources into a shorter, focused output | Thread analysis -> workflow guide |
| `extract` | Pull specific information out of broader sources | Spec -> API reference table |
| `adapt` | Reshape content for a different audience or format | Internal plan -> Jira epic description |
| `compose` | Combine multiple sources into a unified whole | Multiple plans -> initiative overview |

## Reference Files

- **references/compile.md** — Compile workflow: folio.yml schema, steps, tree targets, batch targets
- **references/review.md** — Review workflow: steps, output format, cross-reference checks
- **references/plan.md** — Plan workflow: 7 phases, lens system, re-run rules, agent prompt templates
