# Distribution Conventions

Canonical source: `pkg/dendrik/conventions/distribution.md`.

- **Portable kernel:** Agent Skills + repository `AGENTS.md` + deterministic CLIs/scripts.
- **Admitted direct roots:** canonical skills and policy project once to Claude, default OMP, and
  Codex user roots after native discovery and invocation proof. Codex configuration is app-owned;
  named OMP profiles and generic roots are absent.
- **Claude marketplace:** Claude Code alone has the generated catalog, closed bundles, and
  self-locating installer. OMP's isolated marketplace fallback is proven without a catalog here.
- **Registry:** root `plugins.json` points at closed `plugins/<tool>` bundles, never `cmd/<tool>`.
- **Bundle:** native Claude manifest, `skills/<tool>`, self-locating `bin/setup`, generated `VERSION`.
  Source, tests, build files, runtime state, and private material are forbidden.
- **Versions:** `cmd/<tool>/VERSION` owns the binary; bundle `VERSION` mirrors it; native Claude
  `plugin.json.version` owns plugin updates. A binary bump requires a plugin bump, while a
  skill-only plugin bump is valid.
- **Generation:** only `.claude-plugin/marketplace.json` is generated. Native validation and isolated
  behavioral proof—not file presence—establish support.
- **Private boundary:** private overlays remain local sync inputs and never enter public distribution.
  Native OMP MCP definitions remain private and are never documented by name or value.
- **Admission:** add another native adapter only after primary-source research and behavioral proof;
  extract generic machinery only after a second proven adapter demonstrates duplication.
