# Quick Push (Level 0)

Single-file operations without a forest. For the full walkthrough including prerequisites
and troubleshooting, see [docs/USAGE.md](../docs/USAGE.md#quick-start).

## Commands

```bash
jf push PROJ-123 description.md    # compile + push to Jira
jf pull PROJ-123 output.md         # pull Jira description to local file
```

## Agent Decision Points

- **When to use Level 0 vs forest**: If the user has one ticket, use Level 0. If they mention
  multiple related tickets, suggest `jf init` to create a forest.
- **`--plain-text` fallback**: If marklassian fails (Node.js issue), retry with `--plain-text` for
  plain text. Warn the user that rich formatting will be lost.
- **Pull preserves frontmatter**: If the target file has YAML frontmatter, `jf pull` keeps
  it and only replaces content below the closing `---` fence.

## Batch Operations

For non-interactive (agent) use, add `--yes` to skip the TTY confirmation prompt:

```bash
jf push --yes                  # push all stale nodes without prompting
jf sync --yes --dry-run --json # preview sync plan as structured JSON
```

Use `--dry-run --json` to inspect the plan before committing to execution.
