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

- Each Go CLI is its own module under `cmd/<tool>/` — `cd cmd/<tool> && make build/test/check` (`make check` = fmt + vet + test, the full pre-commit gate).
- **Binaries are fetched, not committed.** This repo delivers its tools via GitHub Releases: the `cmd/<tool>/<tool>` binaries are gitignored and installed by each plugin's `bin/setup` (or built locally with `make build`) — see `pkg/dendrik/conventions/release.md`. The lone committed build artifact is the `go:embed`-ed `cmd/jf/internal/pipeline/md2adf.bundle.mjs`, which must be in-tree to compile (`make bundle`).

## Contributing

The contributor flow — setup, per-layer verify gates (`make check` / `dendrik lint` /
`dot validate`), and the PR process — lives in [CONTRIBUTING.md](CONTRIBUTING.md).
