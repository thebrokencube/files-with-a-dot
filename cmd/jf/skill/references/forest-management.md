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

## Snapshot Workflow

For programmatic forest sync, use the snapshot-first approach:

| User intent | Command | Key flags |
|-------------|---------|-----------|
| "Snapshot the forest" | `jf snapshot --json` | Full plan + tokens |
| "Snapshot one ticket" | `jf snapshot KEY --json` | Single-node plan + token |
| "Preview sync plan" | `jf snapshot` (no --json) | Human-readable plan display |

### Token Lifecycle

1. `jf snapshot --json` writes `.jf/snapshots/latest.json` with per-node tokens
2. Tokens are snapshot-scoped: new snapshot invalidates all old tokens
3. TTL: 5 minutes (file-scoped). Process-scoped snapshots (auto-snapshot in sync/push) have no TTL.
4. Token errors: `TOKEN_INVALID` (content changed) or `SNAPSHOT_EXPIRED` (TTL exceeded)
5. Recovery: re-run `jf snapshot --json`

### Error Recovery

On any token error during push/pull:
1. Re-run `jf snapshot --json`
2. Already-executed nodes appear as "skip" (state.json updated per-node)
3. Re-evaluate plan: some Tier 2 blocks may have resolved
4. If user previously approved a Tier 2 block and diff is unchanged, re-execute without re-asking
