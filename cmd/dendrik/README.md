# dendrik — Tool Contract Linter

dendrik validates that CLI tools in the dotfiles repo follow the dendrik platform conventions. It runs a contract across three layers (Go, Skill, Bridge) and reports violations with actionable remediation.

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
├── cmd_lint.go       # Orchestrator: flags, gatherToolData(), output formatting
├── lint_go.go        # GoLint: 10 checks (go-mod-linked through readme-doc-links)
├── lint_skill.go     # SkillLint: Layer 1 delegation + 4 Layer 2 checks
├── lint_bridge.go    # BridgeLint: 9 checks (dendrik-import through run-has-json)
├── *_test.go         # Tests for each linter
├── Makefile          # build, test, check targets
└── skill/            # Claude Code skill (SKILL.md + references)

pkg/dendrik/
├── conventions/
│   └── contract.go   # 29 ContractEntry structs with ID, rationale, remediation
├── agentskills/
│   └── validate.go   # Layer 1 SKILL.md validator (standalone, does its own I/O)
└── output_format.go  # Output type (inert formatter, parallel to Palette)
```

## Contract Layers

- **Go layer** — Build infrastructure: go.mod, go.work, main dispatch, Makefile, README
- **Skill layer** — Agent discovery: SKILL.md frontmatter, links, arrow refs, activation guidance
- **Bridge layer** — Integration: dendrik imports, exit constants, JSON output, symlink entries

All linters are pure functions. The orchestrator (`cmd_lint.go`) handles all filesystem reads via `gatherToolData()` → `ToolData` struct → linters.

## Documentation

- [Getting Started](docs/01-getting-started.md) — first lint run, reading results, fixing violations
- [Reference](docs/03-reference.md) — full check catalog, severity model, output formats, flags

## Claude Code Integration

dendrik includes a skill file for Claude Code agent use:

- [skill/SKILL.md](skill/SKILL.md) — agent-facing reference with check IDs and remediation guidance
