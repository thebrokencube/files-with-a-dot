# dot -- Dotfiles manager

Shell-based CLI for installing, syncing, and validating the dotfiles environment.

## Build

No build step. `dot` is a shell script, not a compiled binary.

## Test

```bash
dot validate
```

Runs shellcheck, bash syntax checks, symlink_map.txt verification, and skill frontmatter validation.

## Binary Distribution

Shell script at `cmd/dot/dot`, symlinked to `~/.local/bin/dot` via `symlink_map.txt`.

## Code Conventions

### Architecture

Functional Core, Imperative Shell (FCIS):
- `cmd/dot/lib/` -- 11 reusable shell function libraries
- `cmd/dot/dot` -- entry point dispatcher
- Core scripts: `sync.sh`, `cleanup.sh`, `health.sh`
- Utilities: `cmd/dot/scripts/` (bootstrap, uninstall, validate)

### Shell Standards

- Bash 3.2 compatible (macOS system bash)
- `set -euo pipefail` in all scripts
- `trap ERR` for detailed error info

## Safe Operation Conventions

- `DOTFILES_NONINTERACTIVE=1` auto-set when stdin isn't a TTY
- `confirm()` function with `--force` / `-f` support
- Automatic backups via `lib/backup.sh` (disable with `NO_BACKUP=true`)
- `--dry-run` for preview operations

## Deep Context

@ARCHITECTURE.md
