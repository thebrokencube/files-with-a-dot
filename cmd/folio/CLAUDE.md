# folio -- Knowledge work lifecycle CLI

Manages folio projects: observations, spikes, designs, plans, work tracks, retros, and composition targets.

## Build

```bash
cd cmd/folio
make build
```

## Test

```bash
cd cmd/folio
make test
```

Full pre-commit check (format, vet, test):

```bash
cd cmd/folio
make check
```

## Binary Distribution

Release binaries install to `~/.local/bin/folio` through `dot sync` or `plugins/folio/bin/setup`.
Skill at `plugins/folio/skills/folio/`, symlinked to `~/.claude/skills/folio`.
After code changes: run `make check && make build`; the in-tree binary is transient and not committed.

## Code Conventions

### Adding a Command

1. Create `cmd_<name>.go` with `runXxx(args []string) int`
2. Register in `main.go` switch dispatch
3. Add `--dry-run` / `-n` flag if the command mutates files or external state
4. Use `dendrik.NewFlagSet()` and `dendrik.Parse()` for flag parsing
5. Use `dendrik.WriteResult()` for `--json` output

### Internal Packages

- `internal/config/` -- folio.yml parsing and validation
- `internal/graph/` -- DAG resolution and dependency tracking
- `internal/health/` -- project health checks
- `internal/home/` -- folio home git operations
- `internal/status/` -- status derivation
- `internal/validate/` -- folio.yml validation rules

### Exit Codes

- `0` -- success
- `1` -- user/validation error
- `2` -- external tool error

## Safe Operation Conventions

- `--dry-run` / `-n` on destructive commands (`new`, `archive`) -- always preview first
- `folio home` subcommands for all git operations -- never raw git in `~/.folio`
- folio.yml mutations only through CLI commands, not hand-editing
- `--force` required for recomposing with same DAG

## Deep Context

@docs/01-getting-started.md
@docs/02-workflows.md
@docs/03-reference.md
