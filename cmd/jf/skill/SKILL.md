---
name: jf
description: "Use when managing Jira tickets, creating/editing issues, pushing descriptions, searching with JQL, managing ticket lifecycle (parking, repurposing), or when needing Jira conventions, project field defaults, and content gotchas."
---

# Jira Forest (jf)

Standalone CLI and canonical Jira reference. Manages ticket hierarchies as local markdown forests, provides conventions for ticket structure, and holds configuration for project-specific field defaults via `~/.jf.yml`. Works at two levels: Level 0 (single-file push/pull) and Level 1 (forest-aware operations with `forest.yml`).

## Quick Reference

| Command | What it does |
|---------|-------------|
| `jf clone <KEY>` | Scaffold local forest from Jira hierarchy |
| `jf init` | Create `forest.yml` in current directory |
| `jf setup` | Check prerequisites (node, acli, auth) |
| `jf push <KEY> <FILE>` | Compile markdown to ADF and push to Jira description |
| `jf pull <KEY> <FILE>` | Pull Jira description to local markdown |
| `jf sync` | Push all stale + pull all pull-mode nodes |
| `jf tree` | Show forest hierarchy |
| `jf list [--json]` | Flat list of all nodes |
| `jf show <target>` | Single-node detail view |
| `jf status [--json]` | Forest summary with staleness |
| `jf validate` | Check forest integrity |
| `jf create-missing` | Create Jira tickets for TBD nodes |
| `jf search <text> [--json]` | Find Jira tickets by text/project/type |
| `jf rm <KEY>...` | Remove node files from forest |
| `jf schema` | Emit JSON Schema for forest.yml and frontmatter |

Most commands accept `--json` for structured output.

**Flag ordering**: All `jf` commands require flags before positional arguments. `jf push --dir /tmp KEY FILE` works; `jf push KEY FILE --dir /tmp` errors.

## Level Detection

Before any operation, detect the working level:

1. Run `jf setup --check --json` — verify environment (node, acli, JIRA_API_TOKEN)
2. Try `jf tree --json` from the working directory
   - If forest found: **Level 1** (forest-aware operations available)
   - If no forest: **Level 0** (single-file push/pull only)

## Level 0: Single-File Operations

For ad-hoc push/pull without a forest:

```bash
jf push PROJ-123 description.md    # compile + push
jf pull PROJ-123 output.md         # pull to local file
```

-> See references/quick-push.md for details.

## Level 1: Forest Operations

With a `forest.yml` in the directory tree:

```bash
jf status --json     # what's stale?
jf sync              # push all stale + pull all pull-mode nodes
jf create-missing    # create tickets for TBD nodes
```

-> See references/forest-management.md for discovery, validation, sync workflows.

## Forest Structure

A forest is a directory tree with:
- `forest.yml` at the root (schema version, defaults for sync/type/project)
- `.md` files with YAML frontmatter containing `jira: KEY` (or `jira: TBD`)
- Directory `README.md` files become parent nodes; files in directories become children

Frontmatter fields: `jira` (required), `label`, `type`, `sync` (push/pull/both), `order` (sibling sort).
See [docs/USAGE.md](../docs/USAGE.md#frontmatter-reference) for field details, inheritance, and label derivation.

`jf clone` scaffolds a forest with `sync: both` by default (override with `--sync push|pull|both`). Content is pulled from Jira initially; a state baseline is recorded so the first `jf sync` has real content hashes for conflict detection.

`jf tree --json` outputs `[]NodeInfo` (same structure as `jf list --json`). `jf tree --verbose` shows sync direction icons and file paths.

For architecture details (data models, pipeline internals, module structure), see [docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md).

## Lifecycle

Park tickets to permanently deactivate noise — over-decomposed, superseded, or unneeded tickets become blank placeholders in a parking lot epic. Parked tickets can be repurposed later instead of creating new ones.

-> See references/lifecycle.md for the park workflow and repurposing guidance.

## Conventions

Ticket naming, description structure, content rules, and project field defaults.

-> See references/conventions.md for all conventions and project-specific creation templates.

## Configuration

`~/.jf.yml` holds per-environment config: cloud ID, project field defaults, parking lot settings.

-> See references/configuration.md for the full schema.

## JQL & Bulk Operations

-> See references/jql-patterns.md for NL-to-JQL translation and bulk acli operations.

## Gotchas

-> See references/gotchas.md for Jira-specific pitfalls (content rendering, MCP optimization, acli quirks).
