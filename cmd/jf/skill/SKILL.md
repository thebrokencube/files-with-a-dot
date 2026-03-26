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
| `jf init` | Create `.jf/forest.yml` in current directory |
| `jf setup` | Check prerequisites (node, acli, auth) |
| `jf push <KEY> <FILE>` | Compile markdown to ADF and push to Jira description (`--dry-run`, `--plain-text`) |
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

jf uses a 3-tier safety system. Every sync operation runs Read-Plan-Execute:
1. **Read**: Snapshot both local and remote state for all nodes
2. **Plan**: Pure function evaluates 8 rules per node, produces actions
3. **Execute**: Only processes Plan output — no independent decisions

### Tiers

| Tier | Gate | Examples | Override |
|------|------|---------|---------|
| 1: Always safe | None | Push with baseline + local-only change + substantive content | N/A |
| 2: Interactive | TTY prompt | First push/pull with content on other side | Human-only (not agents) |
| 2b: Conflict | --resolve flag | Both sides changed since baseline | Re-run with --resolve local\|remote |
| 3: Impossible | No mechanism | Push empty content; push when remote unreachable | None |

### Blocked Operations

When an operation is blocked, jf prints the reason AND an action hint:
```
BLOCKED PROJ-789: empty content — no override
BLOCKED PROJ-456: first sync, remote has content — resolve in terminal
BLOCKED PROJ-123: conflict — use --resolve local|remote
```

The hint tells you exactly what to do:
- **"no override"** (Tier 3): Fix the underlying issue — add substantive content, or check network connectivity.
- **"resolve in terminal"** (Tier 2): Requires interactive TTY confirmation from a human.
- **"use --resolve local|remote"** (Tier 2b): Re-run with `--resolve local` (keep local) or `--resolve remote` (keep remote).

### Plan Display

`--dry-run` shows the plan without executing. BLOCKED items sort first; summary line at bottom:
```
── Plan ──────────────────────────────────────────
  BLOCKED PROJ-789  stub.md (empty content — no override)
  PUSH    PROJ-123  README.md (local changed)
  SKIP    PROJ-456  sub/README.md (unchanged)
── 1 push, 1 blocked, 1 skip ──
```

For machine-parseable output, use `--json`:
```bash
jf sync --dry-run --json
```
Returns structured JSON with action, key, file, reason, tier, and hint fields. Agents should prefer `--json` for programmatic plan inspection.

### Batch Safety

Multi-node operations (sync, push, pull) have a batch safety gate:
- **TTY mode**: Plan displays, execution proceeds after `--yes` confirmation
- **Non-TTY mode (agents)**: Use `--yes` flag to confirm batch execution, or use `--dry-run --json` to inspect the plan first

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
See [docs/03-reference.md](../docs/03-reference.md) for field details, inheritance, and label derivation.

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

For architecture details (data models, pipeline internals, module structure), see [docs/04-architecture.md](../docs/04-architecture.md).

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

## When to Use MCP Directly

**HARD RULE**: Always use jf commands for Jira operations. If you must use MCP directly (e.g., Jira MCP tools like `editJiraIssue`, `createJiraIssue`), you MUST explain why jf cannot handle the operation BEFORE making the MCP call. Never skip this explanation.

Valid reasons to use MCP directly:
- **Parking lot repurposing**: jf has no `jf unpark` command yet, and acli doesn't support updating the parent field — must use MCP to reparent a parked ticket
- **Field updates not supported by jf**: Changing status, assignee, or custom fields that jf doesn't manage
- **Queries outside jf scope**: Complex JQL searches, field metadata discovery

Invalid reasons (use jf instead):
- Creating new tickets → `jf create-missing`
- Pushing descriptions → `jf push` or `jf sync`
- Pulling descriptions → `jf pull` or `jf sync`
- Searching tickets → `jf search`
