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

## Programmatic Push/Pull

For agent-driven operations using the snapshot-first workflow:

```bash
jf push KEY FILE --token TOKEN              # execute with snapshot token
jf push KEY FILE --token TOKEN --plain-text # with plain-text fallback
jf pull KEY FILE --token TOKEN              # pull with snapshot token
```

The `--token` flag:
- Validates against `.jf/snapshots/latest.json` (written by `jf snapshot`)
- Skips the internal Read phase (snapshot already captured state)
- Records state after success (next snapshot sees the node as synced)
- Returns structured JSON on stdout: `{"data":{"status":"ok",...}}` or `{"data":{"status":"error","code":"TOKEN_INVALID",...}}`
- Exit code 1 on any validation failure
