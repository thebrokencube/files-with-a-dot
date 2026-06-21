# Dotfiles

See **[AGENTS.md](AGENTS.md)** for the harness-agnostic baseline — the plugin marketplace
(`plugins.json`, generated catalogs, per-plugin `bin/setup`) and the build system.

## Claude specifics

- Add this marketplace: `/plugin marketplace add thebrokencube/files-with-a-dot`
- Install a tool: `/plugin install <tool>@files-with-a-dot` (folio, jf, or dendrik), then run the
  plugin's `bin/setup` once to install its binary.
- Refresh the catalog after the repo updates: `/plugin marketplace update files-with-a-dot`.
