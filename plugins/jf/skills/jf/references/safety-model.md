# Safety Model

jf uses a 3-tier safety system. Every sync operation runs Read-Plan-Execute:
1. **Read**: Snapshot both local and remote state for all nodes
2. **Plan**: Pure function evaluates 8 rules per node, produces actions
3. **Execute**: Only processes Plan output — no independent decisions

## Tiers

| Tier | Gate | Examples | Override |
|------|------|---------|---------|
| 1: Always safe | None | Push with baseline + local-only change + substantive content | N/A |
| 2: Interactive | TTY prompt | First push/pull with content on other side | Human-only (not agents) |
| 2b: Conflict | --resolve flag | Both sides changed since baseline | Re-run with --resolve local\|remote |
| 3: Impossible | No mechanism | Push empty content; push when remote unreachable | None |

## Blocked Operations

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

## Plan Display

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

## Batch Safety

Multi-node operations (sync, push, pull) have a batch safety gate:
- **TTY mode**: Plan displays, execution proceeds after `--yes` confirmation
- **Non-TTY mode (agents)**: Use `--yes` flag to confirm batch execution, or use `--dry-run --json` to inspect the plan first

## When jf Fails: Decision Tree

- **"roundtrip diverges at line N"** -> The content has characters (smart quotes, special unicode) or structures (tables, code blocks) that don't survive the md-to-ADF-to-md roundtrip. Options: fix the source content, or use `--plain-text` as a fallback.
- **"read-only: lint issue"** -> The content uses markdown features not in the supported subset (h1, h3+, tables, code blocks, blockquotes, nested lists, images). Simplify the content or set `sync: pull`.
- **"empty content"** -> The file has no substantive content beyond frontmatter. Add content before pushing.
- **"conflict" / "both sides changed"** -> Re-run with `--resolve local` or `--resolve remote`.
- **"first sync, remote has content"** -> Requires interactive TTY. A human must confirm in terminal.
- **"remote-unknown" / "cannot reach Jira"** -> Check network, JIRA_API_TOKEN, and `jf setup --check`.
- **Path errors / file not found** -> Use absolute paths. If inside a forest, ensure you're in the right directory.
- **Any other error** -> Run with `--dry-run --json` for structured output. Report to user with the full error.
