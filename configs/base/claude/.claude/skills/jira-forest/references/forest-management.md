# Forest Management (Level 1)

Operations that require a `forest.yml` in the directory tree.

## Discovery

```bash
jf discover              # human-readable tree
jf discover --json       # structured output with all node metadata
```

Walks the filesystem from `forest.yml` downward. Files with `jira:` in frontmatter become nodes. Directory `README.md` files become parent nodes for sibling files.

## Validation

```bash
jf validate              # human-readable issues
jf validate --json       # structured: { valid, nodes, issues[] }
```

Checks: duplicate keys, TBD nodes missing type/label, invalid sync values.

## Status

```bash
jf status                # summary: total nodes, TBD count, stale counts
jf status --json         # structured StatusResult
```

Staleness is tracked in `.jf/state.json` (file mtime vs last push timestamp).

## Show

```bash
jf show TEST-123         # detail view for one node
jf show --json TEST-123  # structured NodeInfo
```

Resolves by Jira key (case-insensitive), file stem, or file path.

## Sync

```bash
jf sync                  # push all push-mode nodes, then pull all pull-mode nodes
```

Push uses post-order DFS (children before parents). Skips TBD nodes.

## Create Missing

```bash
jf create-missing --dry-run    # preview what would be created
jf create-missing              # create tickets for all TBD nodes
jf create-missing --force      # skip dedup check
```

Pre-order DFS (parents before children) so parent keys are available for child creation. Each TBD node gets: acli create -> capture key -> rewrite frontmatter -> push description.

Dedup check: searches Jira for existing tickets with matching summary before creating.

## Init

```bash
jf init --project BEN          # scaffold forest.yml in current directory
```

## Tree View

```bash
jf tree                        # indented hierarchy with sync direction icons
jf list --json                 # flat list of all nodes as JSON
```
