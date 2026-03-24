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

## Commands

| Command | What it does |
|---------|-------------|
| `jf clone <KEY>` | Scaffold local forest from Jira hierarchy |
| `jf init` | Create `forest.yml` in current directory |
| `jf setup` | Check prerequisites (node, acli, auth) |
| `jf push [KEY FILE]` | Push markdown to Jira (ad-hoc or forest-wide) |
| `jf pull [KEY FILE]` | Pull Jira description to local markdown |
| `jf sync` | Bidirectional sync (push stale + pull pull-eligible nodes) |
| `jf tree` | Show forest hierarchy |
| `jf list [--json]` | Flat list of all nodes |
| `jf show <target>` | Single-node detail view |
| `jf status [--json]` | Forest summary with staleness |
| `jf validate` | Check forest integrity |
| `jf create-missing` | Create Jira tickets for TBD nodes |
| `jf search <text>` | Find Jira tickets by text/project/type |
| `jf rm <KEY>...` | Remove node files from forest |
| `jf schema` | Emit JSON Schema for forest.yml and frontmatter |

## Code Structure

```
cmd/jf/
├── main.go            # Entry point, command dispatch
├── cmd_*.go           # One file per command (push, pull, sync, etc.)
├── cmd_*_test.go      # Tests for each command
├── helpers.go         # Shared utilities
├── internal/          # Internal packages (forest, jira, markdown)
├── scripts/           # Build/test scripts
├── docs/              # USAGE.md, ARCHITECTURE.md, ROADMAP.md
├── skill/             # Claude Code skill (SKILL.md + references)
├── Makefile           # build, test, check targets
└── testdata/          # Test fixtures
```

## Documentation

- [docs/USAGE.md](docs/USAGE.md) — archetype workflows, command reference, frontmatter, troubleshooting
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — module structure, data models, pipeline, consolidation findings
- [docs/ROADMAP.md](docs/ROADMAP.md) — feature maturity (stable, experimental, planned, considering)

## Claude Code Integration

jf includes a skill file for Claude Code agent use:

- [skill/SKILL.md](skill/SKILL.md) — agent-facing reference with decision trees, conventions, and JQL patterns
