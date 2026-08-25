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
| `jf view <KEY>` | Fetch remote issue details from Jira |
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
├── docs/              # 01-getting-started, 02-workflows, 03-reference, 04-architecture
├── Makefile           # build, test, check targets
└── testdata/          # Test fixtures

plugins/jf/skills/jf/  # Canonical Agent Skill and references
```

## Releasing

Bump `cmd/jf/VERSION`, then dispatch the release workflow — GitHub creates the `jf/vX.Y.Z` tag and uploads binaries. Never push tags by hand; published releases are immutable (bump VERSION to re-release).

```bash
gh workflow run release.yml -f tool=jf
```

See the [build & release convention](../../pkg/dendrik/conventions/release.md).

## Documentation

- [Getting Started](docs/01-getting-started.md) -- prerequisites, .jf/ directory model, quick start, levels
- [Workflows](docs/02-workflows.md) -- archetype workflows, multi-forest management, recovery
- [Reference](docs/03-reference.md) -- command reference, frontmatter, lint rules, safety, troubleshooting
- [Architecture](docs/04-architecture.md) -- module structure, data models, pipeline, command routing

## Claude Code Integration

jf includes a bundled Agent Skill:

- [`plugins/jf/skills/jf/SKILL.md`](../../plugins/jf/skills/jf/SKILL.md) — agent-facing workflows and references
