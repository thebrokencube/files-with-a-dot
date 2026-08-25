# dendrik -- Dotfiles tool contract linter

dendrik is the shared foundation the dotfiles CLI tools (folio, jf, dot) are built on. See [docs/00-what-is-dendrik.md](docs/00-what-is-dendrik.md) for what that means.

Validates CLI tools against a convention contract across Go, Skill, and Bridge layers.

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

Release binaries install to `~/.local/bin/dendrik` through `dot sync` or `plugins/dendrik/bin/setup`.
Skill at `plugins/dendrik/skills/dendrik/`, symlinked to `~/.claude/skills/dendrik`.
After code changes: run `make check && make build`; the in-tree binary is transient and not committed.

## Code Conventions

### Linter Architecture

`lint.GatherToolData` owns filesystem reads and builds `ToolData`; `lint.Run` dispatches the pure
Go, Skill, and Bridge layer checks.

### Adding a Check

1. Add the pure check function under `pkg/dendrik/lint/`
2. Register it in the layer dispatch and `pkg/dendrik/conventions/contract.go`
3. Gather any new filesystem data only in `lint.GatherToolData`
4. Return `[]lint.Result` with a contract severity and remediation

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
