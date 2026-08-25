# dendrik — Tool Contract Linter

dendrik is the shared foundation the dotfiles CLI tools (folio, jf, dot) are built on. See [docs/00-what-is-dendrik.md](docs/00-what-is-dendrik.md) for what that means.

This CLI is the contract surface: it validates that CLI tools in the dotfiles repo follow the dendrik conventions, running a contract across three layers (Go, Skill, Bridge) and reporting violations with actionable remediation.

## Install

Symlinked via dotfiles: `cmd/dendrik/dendrik` → `~/.local/bin/dendrik` (see `symlink_map.txt`).

## Quick Start

```bash
dendrik lint cmd/jf          # lint a single tool
dendrik lint cmd/folio        # lint another tool
dendrik lint cmd/dendrik      # lint itself

dendrik lint cmd/jf --json    # structured output for agents
dendrik lint cmd/jf --strict  # promote warnings to errors
dendrik lint --explain go-mod-linked  # show rationale for a check
```

## Commands

| Command | What it does |
|---------|-------------|
| `dendrik lint <path>` | Run contract validation against a tool directory |
| `dendrik lint --json` | JSON output (ResultEnvelope format) |
| `dendrik lint --strict` | Promote warnings to errors |
| `dendrik lint --fix` | Apply mechanical fixes (go.work / symlink_map wiring), then re-lint |
| `dendrik lint --explain <id>` | Show rationale and remediation for a check ID |
| `dendrik version` | Show version |

## Examples

Validate a tool before release (warnings block too):

```bash
dendrik lint cmd/jf --strict
# jf: 0 error(s), 0 warning(s)   → ready to tag
```

Understand and fix a specific failure:

```bash
dendrik lint cmd/jf
#   W [version-flag] main.go does not handle a --version flag (main.go)
dendrik lint --explain version-flag   # rationale + remediation for that check
```

Drive dendrik from an agent or script (structured output):

```bash
dendrik lint cmd/folio --json | jq '.data.errors'
```

## Code Structure

```
cmd/dendrik/
├── main.go           # Entry point, command dispatch
├── cmd_lint.go       # Thin CLI shell over pkg/dendrik/lint
├── *_test.go         # CLI tests
└── Makefile          # build, test, check targets

pkg/dendrik/
├── conventions/
│   └── contract.go   # canonical Contract slice
├── agentskills/
│   └── validate.go   # Agent Skills validator
├── lint/             # importable gather + pure lint core
└── output_format.go  # inert Output formatter

plugins/dendrik/skills/dendrik/ # Canonical Agent Skill and references
```

## Contract Layers

- **Go layer** — Build infrastructure: go.mod, go.work, main dispatch, Makefile, README
- **Skill layer** — Agent discovery: SKILL.md frontmatter, links, arrow refs, activation guidance
- **Bridge layer** — Integration: dendrik imports, exit constants, JSON output, symlink entries

`lint.GatherToolData` owns filesystem reads; `lint.Run` and the per-layer checks remain pure over `ToolData`.

## Releasing

Bump `cmd/dendrik/VERSION`, then dispatch the release workflow — GitHub creates the `dendrik/vX.Y.Z` tag and uploads binaries. Never push tags by hand; published releases are immutable (bump VERSION to re-release).

```bash
gh workflow run release.yml -f tool=dendrik
```

See the [build & release convention](../../pkg/dendrik/conventions/release.md).

## Documentation

- [Getting Started](docs/01-getting-started.md) — first lint run, reading results, fixing violations
- [Starting a new tool](docs/02-new-tool.md) — the copy-an-exemplar + `lint --fix` recipe
- [Reference](docs/03-reference.md) — full check catalog, severity model, output formats, flags

## Claude Code Integration

dendrik includes a bundled Agent Skill:

- [`plugins/dendrik/skills/dendrik/SKILL.md`](../../plugins/dendrik/skills/dendrik/SKILL.md) — agent-facing workflow and references
