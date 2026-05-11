# Dotfiles

## Build System
- Each Go CLI has its own module root under `cmd/<tool>/` — use `cd cmd/<tool> && make build/test/check`
- `make check` runs fmt + vet + test (full pre-commit validation)
- **Build artifacts are checked into git** — this repo is deployment-via-clone, so binaries and bundles must be committed. `dot sync` does not run build steps.
  - `cmd/folio/folio` — Go binary
  - `cmd/jf/jf` — Go binary
  - `cmd/jf/internal/pipeline/md2adf.bundle.mjs` — embedded marklassian bundle, rebuild with `make bundle`
  - `cmd/dendrik/dendrik` — Go binary
- Do NOT gitignore build outputs in this repo. Only gitignore transient dev state (`node_modules/`).
