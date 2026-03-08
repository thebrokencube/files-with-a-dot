# Dotfiles

## Versioning Convention
This repo uses a **custom semver scheme** that differs from the commit skill defaults:
- **PATCH** for all normal changes (feat, fix, refactor, docs, etc.)
- **MINOR** for breaking changes only
- **MAJOR** reserved until interfaces stabilize (pre-1.0 mindset)

This overrides the commit skill's default rule where `feat` bumps MINOR.

## Build System
- Module root is `cmd/folio/` (not repo root) — use `cd cmd/folio && make build/test/check`
- `make check` runs fmt + vet + test (full pre-commit validation)
- The `folio` binary at `cmd/folio/folio` is checked into git — rebuild and commit it after code changes
