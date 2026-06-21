---
name: jf
description: "Use when managing Jira tickets, creating/editing issues, pushing descriptions and titles, searching with JQL, managing ticket lifecycle (parking, repurposing), or when needing Jira conventions, project field defaults, and content gotchas."
user_invocable: true
argument-hint: "[push|pull|sync|search|tree|status|clone|create-missing|...] [args]"
---

# Jira Forest (jf)

Standalone CLI and canonical Jira reference. Manages ticket hierarchies as local markdown forests, provides conventions for ticket structure, and holds configuration for project-specific field defaults via `~/.jf.yml`. Works at two levels: Level 0 (single-file push/pull) and Level 1 (forest-aware operations with `forest.yml`).

## Setup

`jf` needs its binary installed. Run this plugin's `bin/setup` once to install it (and again after a
plugin update). It is idempotent — safe to re-run, and a no-op when the pinned version is already installed.

## Quick Reference

| Command | What it does |
|---------|-------------|
| `jf clone <KEY>` | Scaffold local forest from Jira hierarchy |
| `jf init` | Create `.jf/forest.yml` in current directory |
| `jf setup` | Check prerequisites (node, acli, auth) |
| `jf push <KEY> <FILE>` | Push description + title to Jira (`--dry-run`, `--plain-text`) |
| `jf pull <KEY> <FILE>` | Pull Jira description to local markdown (`--dry-run`) |
| `jf sync` | Push all stale + pull all pull-eligible nodes (`--dry-run`, `--resolve`, `--json`) |
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

**Flag ordering**: Flags can appear before or after positional arguments. Both `jf push --dir /tmp KEY FILE` and `jf push KEY FILE --dir /tmp` work.

## Safety Model

3-tier system: Tier 1 (always safe, no gate), Tier 2 (interactive TTY prompt or `--resolve` flag), Tier 3 (impossible, no override). Every sync runs Read-Plan-Execute. When blocked, jf prints the reason and an action hint.

-> Read references/safety-model.md for tier details, blocked operation examples, plan display format, and the failure decision tree.

## Preflight Checklist (MANDATORY)

Run this sequence before ANY jf operation. Do not skip steps.

1. **Config**: Read `~/.jf.yml` — get cloud_id, project defaults, parking lot epic keys
2. **Environment**: `jf setup --check --json` — verify node, acli, JIRA_API_TOKEN
3. **Forest**: `jf tree --json` from the working directory
   - Forest found -> **Level 1** (forest-aware operations available)
   - No forest -> **Level 0** (single-file push/pull only)
4. **Status** (Level 1 only): `jf status --json` — check staleness before sync

If you skip this checklist and a jf operation fails, come back here first. For the failure decision tree, see references/safety-model.md.

## Level 0: Single-File Operations

For ad-hoc push/pull without a forest:

```bash
jf push PROJ-123 description.md    # compile + push
jf pull PROJ-123 output.md         # pull to local file
```

-> See references/quick-push.md for details.

## Level 1: Forest Operations

With a `.jf/forest.yml` in the directory tree:

```bash
jf status --json     # what's stale?
jf sync              # push all stale + pull all pull-eligible nodes
jf create-missing    # create tickets for TBD nodes
```

-> See references/forest-management.md for discovery, validation, sync workflows.

## Forest Structure

A forest is a directory with a `.jf/` subdirectory containing:
- `.jf/forest.yml` — schema version, defaults for sync/type/project
- `.jf/*.md` — node files with YAML frontmatter containing `jira: KEY` (or `jira: TBD`)
- `.jf/subdir/README.md` — directory README.md files become parent nodes; sibling files become children

```
my-project/
├── .jf/
│   ├── forest.yml        # config
│   ├── README.md         # root node (epic)
│   ├── feature-a.md      # child of root
│   └── epics/
│       ├── README.md     # nested parent
│       └── story.md      # child of epics/
```

The `.jf/` directory is the forest root. All node discovery happens inside it.

Frontmatter fields: `jira` (required), `label`, `type`, `sync` (override-only: push/pull), `order` (sibling sort).

**`label` maps to the Jira summary field (the ticket title).** `jf push` and `jf sync` update both the description (from the markdown body) and the title (from `label`). Do not use MCP to update the summary separately — changing `label` and pushing is the correct workflow. (The name `label` is a known source of confusion — humans think "title", Jira's API calls it "summary", jf calls it "label". Consider renaming to `summary` or `title` in a future version.)

For field details, inheritance, and label derivation, see `docs/03-reference.md` in the jf source tree (`cmd/jf/docs/`).

### Sync Direction

`sync:` is **override-only** — omit it for the common case. When absent, the engine derives effective direction from content mutability:
- **Mutable content** (lint + roundtrip pass) → push+pull
- **Read-only content** (lint or roundtrip fail) → demoted to pull-only
- **Empty content** → blocked (Tier 3)

Use `sync: pull` to force pull-only (never push even if mutable). Use `sync: push` to force push-only (never pull even if remote changes). Explicit `sync: both` is valid but equivalent to omitting — it's the default derived behavior.

**Override vs mutability**: When `sync: push` is set but content is read-only (lint/roundtrip fail), the engine **skips** the node (can't push read-only content, can't pull because sync says push-only). When `sync: pull` is set, mutability is irrelevant — the node always pulls.

These overrides are for specific use cases, not the default.

`jf clone` scaffolds a forest without `sync:` fields (override with `--sync push|pull|both`). Content is pulled from Jira initially; a state baseline is recorded so the first `jf sync` has real content hashes for conflict detection.

`jf status` shows the **effective derived direction** per node — not the raw sync field.

`jf tree --json` outputs `[]NodeInfo` (same structure as `jf list --json`). `jf tree --verbose` shows sync direction icons and file paths.

For architecture details (data models, pipeline internals, module structure), see `docs/04-architecture.md` in the jf source tree (`cmd/jf/docs/`).

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

## Restructure

Workflow for creating new epics/projects, reparenting tickets, and reorganizing Jira hierarchy.

-> See references/restructure.md for the full phased playbook (prepare → create → reparent → sync → verify → close → update folio).

## When to Use MCP Directly

**HARD RULE**: Always use jf commands for Jira operations. If you must use MCP directly (e.g., Jira MCP tools like `editJiraIssue`, `createJiraIssue`), you MUST explain why jf cannot handle the operation BEFORE making the MCP call. Never skip this explanation.

Valid reasons to use MCP directly:
- **Reparenting tickets**: jf has no reparent command — use MCP `editJiraIssue` with `parent: { key: "..." }`
- **Transitioning status**: jf has no transition command — use MCP `transitionJiraIssue`
- **Creating issue links**: jf has no link command — use MCP `createIssueLink`
- **Parking lot repurposing**: jf has no `jf unpark` command yet, and acli doesn't support updating the parent field — must use MCP to reparent a parked ticket
- **Field updates not supported by jf**: Changing assignee or custom fields that jf doesn't manage
- **Queries outside jf scope**: Complex JQL searches, field metadata discovery

Invalid reasons (use jf instead):
- Creating new tickets → `jf create-missing`
- Pushing descriptions → `jf push` or `jf sync`
- Updating ticket title/summary → change `label` in frontmatter, then `jf push` (push updates both description and title)
- Pulling descriptions → `jf pull` or `jf sync`
- Searching tickets → `jf search`

## Testing

-> See references/testing.md for clone-based integration testing (setup, sync validation, teardown).

## Reference Files

- **references/safety-model.md** — Safety tiers, blocked operations, plan display, failure decision tree
- **references/quick-push.md** — Level 0 single-file push/pull details
- **references/forest-management.md** — Forest discovery, validation, sync workflows
- **references/lifecycle.md** — Park workflow and ticket repurposing
- **references/conventions.md** — Ticket naming, description structure, content rules, project templates
- **references/configuration.md** — `~/.jf.yml` schema
- **references/jql-patterns.md** — NL-to-JQL translation and bulk acli operations
- **references/gotchas.md** — Jira-specific pitfalls (content rendering, MCP, acli quirks)
- **references/restructure.md** — Epic/project creation, reparenting, hierarchy reorganization
- **references/testing.md** — Clone-based integration testing
