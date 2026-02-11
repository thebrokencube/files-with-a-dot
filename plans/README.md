# Project Plans

Tracking files for architectural work on the dotfiles repo.

## Projects

| Project | Status | Depends on | Description |
|---------|--------|------------|-------------|
| [fcis-refactor](fcis-refactor.yml) | **Complete** | — | v1: Extract libs, non-interactive protocol |
| [migration-system](migration-system.yml) | Pending | — | Consolidated version-gated migrations |
| [fcis-v2-standalone-scripts](fcis-v2-standalone-scripts.yml) | Pending | migration-system | v2: Sourced libs → standalone scripts |
| [stow-migration](stow-migration.yml) | Pending | v2, migration-system | Replace custom symlinks with GNU Stow |
| [ci-lint](ci-lint.yml) | Pending | — | GitHub Actions lint workflow |

## Dependency graph

```
fcis-refactor (COMPLETE)
    │
    ├── ci-lint (independent)
    │
    ├── migration-system (independent)
    │       │
    │       ├── fcis-v2 (steps 1-6 can start before migration-system,
    │       │            but step 7 requires it complete)
    │       │       │
    │       │       └── stow-migration
    │       │
    │       └── stow-migration (adds stow-readiness migration)
```

## Conventions for agents

When picking up work from these plans:

### Finding work
1. Read this README for the project map
2. Pick a project whose dependencies are met
3. Read the project's .yml file for pending steps
4. Steps are ordered — work top to bottom

### Verification (every commit)
1. `./scripts/validate.sh` — shellcheck, syntax, symlink sources, skills
2. `./sync.sh --dry-run --skip-pull` — confirm output matches expectations
3. If the step modifies cleanup.sh: `./cleanup.sh --dry-run`
4. If the step modifies health.sh: `./health.sh`

### Commit conventions
- Message format: `v<version>: type(scope): description`
- Tag every commit: `git tag v<version>`
- Bump types: patch (non-breaking), minor (breaking change)
- Version is determined at commit time: check `git describe --tags --abbrev=0`
  for current version, then bump accordingly
- Do NOT add Co-Authored-By or other trailers

### Key files to understand
- `ARCHITECTURE.md` — current design, decisions, and conventions
- `dot` — CLI entry point, orchestrates everything
- `sync.sh` — main sync orchestrator (biggest script)
- `lib/` — sourced helper libraries
- `scripts/` — standalone scripts (as they get created)
- `shared/` — config files organized by tool (stow packages)
