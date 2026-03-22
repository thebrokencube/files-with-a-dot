# jf — Jira Forest CLI

jf maps a local filesystem of markdown files to Jira ticket descriptions.
A **forest** is a directory of `.md` files (with YAML frontmatter) that maps to
a Jira hierarchy. Directories become parent–child relationships, and `sync`
keeps everything in sync. Works at two levels: ad-hoc single-file push/pull
(Level 0) or full forest management with bidirectional sync (Level 1+).

## Prerequisites

- **Node.js** — markdown/ADF conversion
- **acli** — Atlassian CLI (`brew install acli`)
- **Jira auth** — `acli auth login`

Verify: `jf setup`

## Install

Symlinked via dotfiles: `cmd/jf/jf` → `~/.local/bin/jf` (see `symlink_map.txt`).

## Quick Start

**Clone an existing hierarchy:**

```bash
jf clone ACME-100           # scaffold local forest from Jira
cd api-redesign            # enter the scaffolded directory
jf tree                    # see the hierarchy
jf sync                    # push local edits, pull remote changes
```

**Start from scratch:**

```bash
mkdir my-effort && cd my-effort
jf init --project ACME      # create forest.yml
# create .md files with jira: KEY frontmatter
jf push                    # push all descriptions to Jira
```

**Ad-hoc (no forest needed):**

```bash
jf push ACME-123 notes.md   # push one file to one ticket
jf pull ACME-456 output.md  # pull one ticket to one file
```

## Documentation

- [docs/USAGE.md](docs/USAGE.md) — archetype workflows, command reference, frontmatter, troubleshooting
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — module structure, data models, pipeline, consolidation findings
- [docs/ROADMAP.md](docs/ROADMAP.md) — feature maturity (stable, experimental, planned, considering)

## Claude Code Integration

jf includes a skill file for Claude Code agent use:

- [skill/SKILL.md](skill/SKILL.md) — agent-facing reference with decision trees, conventions, and JQL patterns
