# Distribution Conventions

Canonical source: `pkg/dendrik/conventions/distribution.md`

Covers cross-harness distribution: packaging a tool's skill (and, for CLI-backed tools, its binary) into a plugin published to marketplaces that more than one AI coding agent can install.

Key points for quick reference:

- **Support tiers**: Claude Code is the golden path (guaranteed); Cursor, Codex, and other compatible tools are best-effort. Write for the agnostic baseline by default.
- **One registry**: `plugins.json` is canonical; a generate step builds the per-harness catalogs (`.claude-plugin/`, `.cursor-plugin/`, `.agents/plugins/`). Never hand-edit a catalog — they carry `_generated: DO NOT EDIT` and a validate step rejects drift.
- **Registry fields**: `name`, `description` (required); `tools` (optional, defaults to all supported); `path` (optional, overrides the default `plugins/<name>` location).
- **Per-plugin manifests**: the Claude `plugin.json` is canonical/hand-authored; Cursor is hand-authored only when behavior differs; Codex is generated (opt-in via `"codex"` in `tools`). Hooks/commands/`CLAUDE.md` don't transfer to other harnesses.
- **AGENTS.md is the baseline**; `CLAUDE.md`/`.cursorrules` are thin overlays that point at it.
- **CLI-backed plugins**: a plugin holds no binary. Golden path installs/updates it automatically; other harnesses assume it's on `PATH`. The install mechanism is the consuming tool's implementation.

See the canonical source for the full convention and related-convention links.
