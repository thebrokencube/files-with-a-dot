# Quick Push (Level 0)

Single-file operations without a forest. Use when you have one markdown file to push to a specific Jira ticket.

## Push

```bash
jf push PROJ-123 description.md
```

Pipeline: strip frontmatter -> compile markdown to ADF (via marklassian/Node) -> push via acli edit -> record in `.jf/state.json`.

**Supported markdown**: tables, fenced code blocks, blockquotes, nested lists, task lists, all heading levels, bold, italic, code, links. Full CommonMark spec via marklassian.

## Pull

```bash
jf pull PROJ-123 output.md
```

Pulls the Jira description as plain text to the local file. Preserves existing YAML frontmatter if the file already exists.

## Dry Run

Use `--dry-run` (when available) to preview compilation without pushing. Alternatively, inspect `jf schema` output for the ADF structure.

## Prerequisites

Run `jf setup --check` to verify:
- `node` is installed (required for marklassian ADF compilation)
- `acli` is installed (Jira CLI for push/pull)
- `JIRA_API_TOKEN` is set in environment (typically via `~/.env.local`)
