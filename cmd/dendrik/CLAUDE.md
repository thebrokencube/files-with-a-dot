# dendrik -- Dotfiles tool contract linter

Validates CLI tools against a 29-check contract across Go, Skill, and Bridge layers.

## Build

```bash
cd cmd/dendrik
make build
```

## Test

```bash
cd cmd/dendrik
make test
```

Self-lint (validates its own conventions):

```bash
dendrik lint cmd/dendrik
```

## Binary Distribution

Built binary at `cmd/dendrik/dendrik`, symlinked to `~/.local/bin/dendrik` via `symlink_map.txt`.
Skill at `cmd/dendrik/skill/`, symlinked to `~/.claude/skills/dendrik`.
After code changes: rebuild the binary and commit it.

## Code Conventions

### Linter Architecture

All linters are pure functions. The orchestrator (`cmd_lint.go`) handles I/O via `gatherToolData()` which builds a `ToolData` struct, then passes it to each layer:

- `lint_go.go` -- Go layer checks (10 checks)
- `lint_skill.go` -- Skill layer checks (9 checks)
- `lint_bridge.go` -- Bridge layer checks (10 checks)

### Adding a Check

1. Add the check function to the appropriate `lint_*.go` file
2. Register it in the layer's check list
3. Use the `ToolData` struct for all file access (no direct I/O in check functions)
4. Return `[]conventions.Issue` with severity `Error` or `Warning`

### Shared Package

`pkg/dendrik/` provides shared types used by other tools:
- `conventions/` -- contract definitions
- `agentskills/` -- SKILL.md validation
- `output_format.go` -- JSON formatting, `WriteResult()`
- Flag parsing: `NewFlagSet()`, `Parse()`
- Exit codes: `ExitOK`, `ExitUserError`, `ExitExternalErr`, `ExitConflict`

## Deep Context

@docs/01-getting-started.md
@docs/03-reference.md
