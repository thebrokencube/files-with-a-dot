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
`release.md`'s), `description`. Optional: `author`, `pages`. The other harnesses derive from it:

- **Cursor** (`.cursor-plugin/plugin.json`) — hand-authored only when Cursor behavior must differ.
  Otherwise the Claude manifest serves Cursor too.
- **Codex** — opt-in by adding `"codex"` to `tools`; its manifest is generated (don't hand-edit).

Keep Claude-only surfaces out of Cursor- and Codex-compatible plugins. Hooks, commands, and
`CLAUDE.md` do not transfer; skills and agents are the portable content.

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
  the plugin's released version. Version contract: `release.md`.
- **One idempotent `bin/setup`.** It is self-locating (finds its own plugin dir, reads the bundled
  `VERSION`), self-contained (the plugin cache sandbox forbids sibling access), and safe to re-run —
  it installs the pinned binary only when it's missing or the version mismatches. This is the single
  "get the tool ready" entrypoint, not one of several install paths.
- **Run it the same way everywhere.** The skill instructs running `bin/setup` harness-neutrally — no
  per-harness env vars, no harness-specific preflight. New versions arrive through the harness's
  plugin update (a new `VERSION`), which the next `bin/setup` reconciles.

## Enforcement

This convention is guidance, not a `dendrik lint` contract yet. No check validates the
`plugins.json`-to-manifest relationship or the `_generated` header today; that validate step lives
in the consuming repo's own tooling. A dendrik-side manifest-sync check is possible later but isn't
built. The AGENTS.md baseline is handled by the `/dendrik` review framework, not the lint contract.

## Related conventions

- `release.md` — version source of truth (`VERSION`), build, tags; the version contract this builds on.
- `skill.md` — the `cmd/<tool>/skills/<tool>/` plugin-compatible skill layout.
- `cmd/dendrik/skill/references/review-type-agents-md.md` — AGENTS.md review dimensions.
