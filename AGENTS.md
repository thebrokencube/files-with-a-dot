# files-with-a-dot

This repo is both a dotfiles manager and a **cross-harness plugin marketplace** for the Go CLI
tools it ships: `folio`, `jf`, and `dendrik`. This file is the harness-agnostic baseline; a thin
`CLAUDE.md` overlays Claude-specific notes on top.

## Plugin marketplace

The repo distributes its CLI tools as plugins that work on any agent harness.

- **`plugins.json` (repo root) is the one canonical, hand-edited manifest.** It lists each plugin
  (`name`, `path`, `description`). Edit this file and nothing else.
- **Per-harness catalogs are generated, never hand-edited.** `scripts/marketplace-generate` reads
  `plugins.json` and writes:
  - `.claude-plugin/marketplace.json` — Claude (guaranteed)
  - `.cursor-plugin/marketplace.json` — Cursor (best-effort; format pending confirmation)
  - `.agents/plugins/marketplace.json` — Codex / agents (best-effort; same caveat)

  Each carries a `_generated: DO NOT EDIT` header. Regenerate after editing `plugins.json`; the
  smoke test fails on drift.
- **Each plugin lives at `cmd/<tool>/`** with a `.claude-plugin/plugin.json` (name, version,
  description), its `skills/<tool>` skill, and a uniform `bin/setup`.

### Install a tool

In this repo, run that plugin's `bin/setup` once:

```sh
cmd/<tool>/bin/setup
```

Installed as a plugin, the skill runs the bundled `bin/setup` from the plugin root on first use
(no `cmd/<tool>/` prefix — that's the repo path only). It is the same script either way:
idempotent — self-locating, self-contained, safe to re-run, and a no-op when the pinned
version (from the plugin's `VERSION`) is already installed. It downloads the matching release
binary into `~/.local/bin/`.

## Build system

- Each Go CLI has its own module root under `cmd/<tool>/` — use `cd cmd/<tool> && make build/test/check`
- `make check` runs fmt + vet + test (full pre-commit validation)
- **Build artifacts are checked into git** — this repo is deployment-via-clone, so binaries and bundles must be committed. `dot sync` does not run build steps.
  - `cmd/folio/folio` — Go binary
  - `cmd/jf/jf` — Go binary
  - `cmd/jf/internal/pipeline/md2adf.bundle.mjs` — embedded marklassian bundle, rebuild with `make bundle`
  - `cmd/dendrik/dendrik` — Go binary
- Do NOT gitignore build outputs in this repo. Only gitignore transient dev state (`node_modules/`).

## Contributing

The contributor flow — setup, per-layer verify gates (`make check` / `dendrik lint` /
`dot validate`), and the PR process — lives in [CONTRIBUTING.md](CONTRIBUTING.md).
