# files-with-a-dot

This repo is a dotfiles manager. Portable Agent Skills and this `AGENTS.md` form the
harness-agnostic kernel; each native adapter is retained only after its actual discovery and
behavioral contract is observed.

## Native delivery and admitted roots

- **Claude Code marketplace:** `plugins.json` is the canonical hand-edited inventory. Each `path`
  names a closed `plugins/<tool>` publishable bundle, never the `cmd/<tool>` implementation tree.
  `.claude-plugin/marketplace.json` and bundle `VERSION` mirrors are generated with
  `scripts/marketplace-generate`; Claude's installer, status line, Concise output style, and
  PostCompact terminal notice are Claude-native.
- **Direct user roots:** the map projects canonical policy and shared skills once to Claude
  (`~/.claude`), default OMP (`~/.omp/agent`), and Codex (`~/.codex`). Claude main and subagent
  skill discovery, thin tool-local `CLAUDE.md` pointers, Concise, and the repaired PostCompact
  hook are admitted. OMP native skills, `RULES.md`, terminal Ask, compaction policy retention, and
  an isolated marketplace fallback are admitted. Codex admits its user root beside `.system`,
  managed `OnRequest` approval behavior, and restart preservation.
- **Configuration boundaries:** OMP MCP definitions are a private-overlay migration at
  `~/.omp/agent/mcp.json`; public sources never carry their names or values. Codex configuration
  remains app-owned and unmanaged. Named OMP profiles and unproved adapters are not supported by
  these docs.
- **Claude bundles:** each bundle contains `.claude-plugin/plugin.json`, `skills/<tool>`,
  `bin/setup`, and generated `VERSION` only. Go source, tests, build files, runtime state, and
  private material stay out.
- **Versions are independent:** `cmd/<tool>/VERSION` owns the binary; the bundle mirrors it;
  `plugins/<tool>/.claude-plugin/plugin.json.version` owns Claude plugin updates. A binary bump
  requires a plugin bump, while a skill-only plugin bump is valid.

### Install a tool

In this repo, run that plugin's `bin/setup` once:

```sh
plugins/<tool>/bin/setup
```

The script is idempotent, self-locating, self-contained, and installs the release version mirrored
in the bundle into `~/.local/bin/`. `dot sync` installs the same binary version directly from the
implementation authority at `cmd/<tool>/VERSION`.

## Build system

- Each Go CLI is its own module under `cmd/<tool>/` — `cd cmd/<tool> && make build/test/check` (`make check` = fmt + vet + test, the full pre-commit gate).
- **Binaries are fetched, not committed.** This repo delivers its tools via GitHub Releases: the `cmd/<tool>/<tool>` binaries are gitignored and installed by each plugin's `bin/setup` (or built locally with `make build`) — see `pkg/dendrik/conventions/release.md`. The lone committed build artifact is the `go:embed`-ed `cmd/jf/internal/pipeline/md2adf.bundle.mjs`, which must be in-tree to compile (`make bundle`).

## Contributing

The contributor flow — setup, per-layer verify gates (`make check` / `dendrik lint` /
`dot validate`), and the PR process — lives in [CONTRIBUTING.md](CONTRIBUTING.md).
