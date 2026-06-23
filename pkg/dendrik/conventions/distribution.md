# Distribution Convention

dendrik is the shared foundation the dotfiles CLI tools (folio, jf, dot) are built on. This file
holds its conventions for **cross-harness distribution**: packaging a tool's skill (and, for
CLI-backed tools, its binary) into a plugin that more than one AI coding agent can install. See
`cmd/dendrik/docs/00-what-is-dendrik.md`.

This is the source of truth for the convention. It builds on emerging cross-tool patterns — `AGENTS.md`
as a portable, agent-agnostic baseline and the `.agents/plugins/` directory convention multiple agents
recognize — and states dendrik's stance for any dendrik-built tool.

## Support tiers

Claude Code is the primary harness: the golden path, fully supported. Cursor, Codex, and other
compatible tools are best-effort — we don't preclude them and we take contributions, but only Claude
Code is guaranteed. Write for the agnostic baseline by default, and keep harness-specific features
on the golden path.

## One registry, generated manifests

A repo-root `plugins.json` registry is canonical. A generate step builds the per-harness marketplace
catalogs from it. Never hand-edit a catalog:

| Harness | Generated catalog |
|---|---|
| Claude Code (golden path) | `.claude-plugin/marketplace.json` |
| Cursor | `.cursor-plugin/marketplace.json` |
| Codex + `.agents/plugins/`-convention tools | `.agents/plugins/marketplace.json` |

Each generated catalog carries a `_generated: DO NOT EDIT` header. A validate step (CI and
pre-commit) rejects manual edits and stale output. The point is one place to manage plugins, no
matter how many tools read them.

### Registry entry fields

- `name`, `description` — required.
- `tools` — optional. Which harnesses the plugin targets; defaults to all supported. Restrict with
  `["claude"]`, `["cursor"]`, `["codex"]`, or a combination. Most plugins need no `tools` field;
  tool-specific ones are the exception.
- `path` — optional. Overrides the default `plugins/<name>` location, for plugins that live
  elsewhere in the repo (for example, co-located with a tool's source).

### Per-plugin manifests

The Claude manifest, `<plugin>/.claude-plugin/plugin.json`, is canonical and hand-authored. Required
fields: `name` (kebab-case), `version` (semver; the bump rule and source-of-truth chain are
`release.md`'s), `description`. Optional: `author`, `pages`. The per-harness catalogs (and any
per-harness manifest) are generated from it plus `plugins.json` — **uniformly, for every harness the
plugin targets** (all supported by default; opt a plugin out of one with `tools`). Hand-author a
harness-specific manifest only when that harness genuinely needs behavior the Claude manifest can't
express — that's a per-plugin exception, not a per-harness rule.

Keep Claude-only surfaces out of the other harnesses' plugins. Hooks, commands, and `CLAUDE.md` do not
transfer; skills and agents are the portable content.

## AGENTS.md is the baseline

Behavior-changing content — commands, conventions, constraints — lives in `AGENTS.md`, the shared
baseline every harness reads. Per-harness files (`CLAUDE.md`, `.cursorrules`) are thin overlays that
point at it, not copies of it. The review dimensions for AGENTS.md live in
`cmd/dendrik/skill/references/review-type-agents-md.md`; this convention just names it as the
distribution baseline.

## CLI-backed plugins: the binary

A plugin holds skills, commands, hooks, agents, and MCP servers — not a binary. So a CLI-backed
dendrik tool ships its skill in the plugin and installs its binary through a setup step. Keep that
step **uniform across harnesses** — don't special-case one.

- **Plugins do not carry binaries.** The binary is fetched from the tool's GitHub release, pinned to
  the tool's **binary version** (`cmd/<tool>/VERSION`) — distinct from the **plugin version**
  (`plugin.json.version`). The two are independent surfaces; see `release.md` for the contract and
  the coupling rule between them.
- **One idempotent `bin/setup`.** It is self-locating (finds its own plugin dir, reads the bundled
  `VERSION` — the binary version), self-contained (the plugin cache sandbox forbids sibling access),
  and safe to re-run — it installs the binary only when it's missing or the version mismatches. This
  is the single "get the tool ready" entrypoint, not one of several install paths.
- **User-facing docs reference the plugin's bundled `bin/setup`, never a repo path.** An installed
  plugin lives in the harness's plugin cache with `bin/setup` at the plugin root — there is no
  `cmd/<tool>/` prefix (that's `plugins.json`'s repo `path`, stripped on publish). So install docs
  say "the plugin's `bin/setup`" (self-locating; the skill runs it on first use); `cmd/<tool>/bin/setup`
  is the in-repo developer path only. Conflating them sends consumers to a path that doesn't exist.
- **Run it the same way everywhere.** The skill instructs running `bin/setup` harness-neutrally — no
  per-harness env vars, no harness-specific preflight. A new binary reaches users when a plugin
  update (a bumped `plugin.json.version`) re-runs `bin/setup`, which reconciles to the binary
  `VERSION`. That is why a binary bump must carry a plugin bump (the coupling rule in `release.md`).

## Enforcement

Two version invariants are enforced in CI (see `release.md`):
- **Catalog ↔ plugin**: the `version-consistency` job re-runs `scripts/marketplace-generate` and
  asserts `git diff --exit-code`, so every catalog version matches its `plugin.json.version`.
- **Binary ↔ plugin coupling**: `scripts/check-version-coupling` fails a change that bumps a
  tool's binary `VERSION` without bumping its `plugin.json.version` (else the new binary never
  reaches users).

The broader `plugins.json`-to-manifest relationship and the `_generated` header are otherwise
guidance enforced by the consuming repo's tooling; a dendrik-side per-tool contract check is
possible later (the platform-extraction case) but isn't built. The AGENTS.md baseline is handled
by the `/dendrik` review framework, not the lint contract.

## Related conventions

- `release.md` — version source of truth (`VERSION`), build, tags; the version contract this builds on.
- `skill.md` — the `cmd/<tool>/skills/<tool>/` plugin-compatible skill layout.
- `cmd/dendrik/skill/references/review-type-agents-md.md` — AGENTS.md review dimensions.
