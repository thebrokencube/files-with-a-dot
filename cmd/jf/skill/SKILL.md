# Jira Forest (jf)

Standalone CLI for managing Jira ticket hierarchies as local markdown forests. Works at two levels: Level 0 (single-file push/pull) and Level 1 (forest-aware operations with `forest.yml`).

## Quick Reference

| Command | What it does |
|---------|-------------|
| `jf push <KEY> <FILE>` | Compile markdown to ADF and push to Jira description |
| `jf pull <KEY> <FILE>` | Pull Jira description to local markdown |
| `jf discover` | Detect forest tree from filesystem |
| `jf tree` | Show forest hierarchy |
| `jf list [--json]` | Flat list of all nodes |
| `jf validate` | Check forest integrity |
| `jf status [--json]` | Forest summary with staleness |
| `jf show <target>` | Single-node detail view |
| `jf sync` | Push all + pull all |
| `jf create-missing` | Create Jira tickets for TBD nodes |
| `jf rm <KEY>...` | Remove node files from forest by key |
| `jf setup` | Check and install prerequisites |
| `jf init` | Scaffold `forest.yml` in current directory |
| `jf schema` | Emit JSON Schema for forest.yml and frontmatter |

Most commands accept `--json` for structured output.

## Level Detection

Before any operation, detect the working level:

1. Run `jf setup --check --json` — verify environment (node, acli, JIRA_API_TOKEN)
2. Try `jf discover --json` from the working directory
   - If forest found: **Level 1** (forest-aware operations available)
   - If no forest: **Level 0** (single-file push/pull only)

## Level 0: Single-File Operations

For ad-hoc push/pull without a forest:

```bash
jf push BEN-123 description.md    # compile + push
jf pull BEN-123 output.md         # pull to local file
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

Frontmatter fields: `jira` (required), `label`, `type`, `sync` (push/pull), `order` (sibling sort).

## Lifecycle

Park (deactivate) and unpark (reactivate) tickets using a combination of `jf rm` and MCP-driven Jira operations. These are skill-level workflows — the jf binary only handles the local file removal.

-> See references/lifecycle.md for park and unpark workflows.

## folio Integration

When used within a folio project, `folio jira push` delegates to `jf push` and adds folio-specific autoTouch for staleness tracking. Use `jf` directly for forest-level operations; use `folio jira` when you need folio's target system integration.

## Gotchas

-> See references/gotchas.md for Jira-specific pitfalls.
