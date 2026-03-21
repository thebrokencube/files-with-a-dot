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
- **Build artifacts are checked into git** — this repo is deployment-via-clone, so binaries and bundles must be committed. `dot sync` does not run build steps.
  - `cmd/folio/folio` — Go binary, rebuild with `make build` after code changes
  - `cmd/folio/internal/jira/md2adf.bundle.mjs` — embedded marklassian bundle, rebuild with `make bundle`
  - `cmd/folio/scripts/md2adf.bundle.mjs` — same bundle (source copy for rebuilds)
- Do NOT gitignore build outputs in this repo. Only gitignore transient dev state (`node_modules/`).
