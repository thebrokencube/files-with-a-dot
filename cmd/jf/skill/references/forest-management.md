# Forest Management (Level 1)

Operations that require a `forest.yml` in the directory tree. For the full command
reference table, archetype workflows, and flag details, see
[docs/USAGE.md](../docs/USAGE.md#command-quick-reference).

## Agent Command Selection

Use these decision points when the user asks about their forest:

| User intent | Command | Key flags |
|-------------|---------|-----------|
| "What's in my forest?" | `jf tree --json` | Structured node list |
| "What needs syncing?" | `jf status --json` | Effective direction + staleness |
| "Is my forest valid?" | `jf validate --json` | Issues with levels |
| "Show me one ticket" | `jf show --json KEY` | Detail + stale/clean |
| "Sync everything" | `jf sync` | `--resolve local\|remote` for conflicts |
| "Create the TBD tickets" | `jf create-missing --dry-run` first | Preview, then run without `--dry-run` |
| "Push just this branch" | `jf push --subtree KEY` | Scoped push |

## Key Behaviors

- **Push order**: post-order DFS (children before parents)
- **Create order**: pre-order DFS (parents before children — Jira needs parent key first)
- **Sync conflict pre-scan**: checks bidirectional nodes (default behavior) before push/pull. Skips conflicts unless `--resolve` is set.
- **`--plain-text` on create-missing**: pushes plain text if marklassian fails, does NOT skip dedup check
- **Node resolution**: key (case-insensitive) → filename stem → file path

## Batch Operations

For non-interactive (agent) use:

| User intent | Command | Key flags |
|-------------|---------|-----------|
| "Preview sync plan" | `jf sync --dry-run --json` | Machine-parseable plan |
| "Execute without prompting" | `jf sync --yes` | Skip TTY confirmation |
| "Push all stale" | `jf push --yes` | Batch push |

Use `--dry-run --json` to inspect the plan before execution. Already-synced nodes appear as "skip" (state.json updated per-node).
