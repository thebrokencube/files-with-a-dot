# Distribution Conventions

Canonical source: `pkg/dendrik/conventions/distribution.md`.

- **Portable kernel:** Agent Skills + repository `AGENTS.md` + deterministic CLIs/scripts.
- **Proven adapter:** Claude Code only. Cursor, Codex, and other targets emit nothing until native
  contracts and isolated discovery/invocation proof exist.
- **Registry:** root `plugins.json` points at closed `plugins/<tool>` bundles, never `cmd/<tool>`.
- **Bundle:** native Claude manifest, `skills/<tool>`, self-locating `bin/setup`, generated `VERSION`.
  Source, tests, build files, runtime state, and private material are forbidden.
- **Versions:** `cmd/<tool>/VERSION` owns the binary; bundle `VERSION` mirrors it; native
  `plugin.json.version` independently owns plugin updates. Binary bump requires plugin bump.
- **Generation:** only `.claude-plugin/marketplace.json` is generated. Native validation and isolated
  behavioral proof—not file presence—establish support.
- **Private boundary:** private overlays remain local sync inputs and never enter public distribution.
- **Admission:** add another native adapter only after primary-source research and behavioral proof;
  extract generic machinery only after a second proven adapter demonstrates duplication.
