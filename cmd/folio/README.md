# folio
Folio manages knowledge work projects — plans, references, and compiled outputs. The CLI handles
deterministic project management; the portable Agent Skill handles creative work where the host
exposes it.

## Install

Symlinked via dotfiles: `cmd/folio/folio` → `~/.local/bin/folio` (see `symlink_map.txt`).

## Quick Start

```bash
folio init --name "My Project"   # create folio.yml in current directory
folio status                     # show project status
folio observe 'idea(scope): description'  # add an observation
folio home push                  # push to ~/.folio home
```
Use the `folio` Agent Skill for non-trivial planning and composition in an admitted host. The
portable workflow is the same; invocation syntax is host-native.

## Agent Skill workflows

| Skill | What it does |
|-------|-------------|
| `folio plan` | Plan non-trivial tasks by exploring options, then converging on an approach. |
| `folio compose` | Turn project sources into audience-appropriate outputs. |
| `folio gather` | Add sources to a project or research a topic. |
| `folio publish` | Push compiled outputs to external targets. |

## How It Works

```
folio gather → design doc (lock) → implementation plan → execute
          per step: implement → review (gate) → commit
          if targets: → folio compose → folio publish
          retro findings feed back into future work
```

The `folio plan` Agent Skill runs the full pipeline: gather context, freeze architecture in a
design doc (mandatory lock gate), derive the implementation plan, then execute step by step with
mandatory review before each commit. When the plan has external targets (Jira, branch topology),
execution feeds into compose/publish. Retrospective findings loop back as observations for future
cycles.

## Commands

| Command | What it does |
|---------|-------------|
| `folio status` | Show project status (sources, targets, staleness) |
| `folio stale` | Find stale projects needing attention |
| `folio validate` | Validate folio.yml structure |
| `folio init` | Initialize a new folio project |
| `folio new <type> <topic>` | Scaffold a typed artifact |
| `folio observe <text>` | Add/resolve/list observations |
| `folio health` | Project health report |
| `folio home list` | List home-synced projects |
| `folio home push` | Push project to home |
| `folio gather <url>` | Scaffold source entry from URL |
| `folio dag` | Show project DAG |

## Code Structure

```
cmd/folio/
├── main.go            # Entry point, command dispatch
├── cmd_*.go           # One file per command
├── cmd_*_test.go      # Tests for each command
├── helpers.go         # Shared utilities
├── internal/          # Internal packages (folio, home, jira)
├── scripts/           # Build/test scripts
├── Makefile           # build, test, check targets
└── testdata/          # Test fixtures

plugins/folio/skills/folio/  # Canonical Agent Skill and references
```

Project data lives in `~/.folio/`:

```
~/.folio/
├── active/<project>/
│   ├── folio.yml           # Project manifest (sources, targets, DAG)
│   ├── reference/          # Research, analysis, retrospectives
│   ├── work/               # Implementation plans and tracks
│   └── output/             # Composed outputs ready to share
├── archive/                # Completed/shelved projects
└── vault/                  # Cross-cutting knowledge (no folio.yml)
    ├── research/           # Tool surveys, ecosystem landscapes
    ├── domain/             # Business/technical domain knowledge
    ├── guide/              # Reusable procedures
    └── insight/            # Patterns extracted from experience
```

Projects source from the vault via `vault:` prefix paths (e.g., `vault:research/comparable-dvc.md`). Lifecycle types (spike, design, plan, retro) stay project-scoped; references promote to vault when proven cross-cutting.

Status is derived from file modification times — no separate tracking needed.

## Releasing

Bump `cmd/folio/VERSION`, then dispatch the release workflow — GitHub creates the `folio/vX.Y.Z` tag and uploads binaries. Never push tags by hand; published releases are immutable (bump VERSION to re-release).

```bash
gh workflow run release.yml -f tool=folio
```

See the [build & release convention](../../pkg/dendrik/conventions/release.md).

## Documentation

- [Getting Started](docs/01-getting-started.md) — mental model, project anatomy, quick start
- [Workflows](docs/02-workflows.md) — research, planning, composition, publishing patterns
- [Reference](docs/03-reference.md) — folio.yml schema, CLI commands, artifact types
