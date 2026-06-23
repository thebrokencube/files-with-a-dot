# Build & Release Convention

Source of truth for how dendrik tools are versioned, built, tagged, and released.
Owned by dendrik so every tool (and any tool later extracted to its own repo) releases the
same way. The build logic is provided by `dendrik build`; the GitHub orchestration is a thin
workflow shim.

## Two versions: binary and plugin

A CLI-backed plugin tool has **two independent version surfaces**. Conflating them is the
classic failure (a binary that ships but never reaches users, or a stale catalog), so they are
kept distinct, each canonical for its own surface:

| File | Is the… | Drives | Bump when |
|---|---|---|---|
| `cmd/<tool>/VERSION` | **binary version** | `dendrik build` (`-ldflags -X main.version`), the `release.yml` tag, and `bin/setup`'s download | the binary changes |
| `<plugin>/.claude-plugin/plugin.json` `.version` | **plugin version** | the Claude plugin auto-update, and the generated `marketplace.json` catalog versions | *any* bundle content changes (skill, setup, manifests, **or** a binary bump) |

`main.go` declares `var version = "dev"`, overridden at build time from `VERSION`. `plugin.json`
is hand-authored (incl. `version`); catalogs are generated from it — never carry an *independent*
catalog version (a stale one blocks auto-updates).

**The coupling rule:** a **binary** bump (`VERSION`) MUST be accompanied by a **plugin** bump
(`plugin.json.version`). Plugin auto-update fires on a `plugin.json.version` *change*, and that
update is what re-runs `bin/setup` to fetch the new binary — so a binary bump without a plugin
bump never reaches users. The reverse is free: a skill-only change bumps `plugin.json.version`
alone (no binary release). Enforced by `scripts/check-version-coupling` in CI; catalog↔plugin
consistency is enforced by the `version-consistency` CI job (regenerate + `git diff`). (A common
marketplace pattern auto-bumps changed plugins' versions at merge — a future automation option.)
How the plugin is packaged/published across harnesses (the `plugins.json` registry, generated
manifests, the CLI-backed binary-install path): see `distribution.md`.

semver is a compatibility contract, not just an identifier: MAJOR = breaking, MINOR = additive,
PATCH = fix. Honor it once anything depends on a tool.

## Build: `dendrik build`

`dendrik build [dir]` (default `.`) reads `<dir>/VERSION` and produces reproducible,
version-stamped binaries. It is the one place build flags live:

```
go build -trimpath -buildvcs=false -ldflags "-buildid= -X main.version=<VERSION>" -o <out>/<tool>-<os>-<arch>
```

- `--matrix` builds the release matrix (`darwin/arm64`, `linux/amd64`); default is the host platform.
- `--out DIR` (default `dist/`), `--version V` (override the VERSION file), `--json`.
- Artifact name: `<tool>-<os>-<arch>` (tool = directory basename).
- Reproducible flags (`-trimpath -buildvcs=false -buildid=`) make a rebuild byte-identical, so
  CI can reproduce any artifact.

`dendrik` builds *itself* with a plain `go build` (it can't run a not-yet-built copy of itself);
every other tool is built via `dendrik build`.

## Tags & releases

- **Per-tool tags**, `tool/vX.Y.Z` (e.g. `folio/v0.6.0`) — tool-scoped, so a tool keeps its
  release identity if it moves to its own repo. Tools version independently.
- **Never push tags manually.** Releases run via the `release` GitHub Action
  (`workflow_dispatch`); GitHub creates the tag server-side at the release commit. (Manual
  tag-and-push has been unreliable; this removes that step.)
- **Published releases are immutable.** The workflow fails if `tool/vX.Y.Z` already exists —
  bump VERSION rather than overwriting, so dependents can rely on a release never changing.

## Flow

**Skill / plugin-content change only (no binary change):**
1. Bump `<plugin>/.claude-plugin/plugin.json` `.version`; run `scripts/marketplace-generate` to
   refresh catalogs; commit + push. No binary release — the marketplace serves the new content
   from the repo and auto-update reaches users; `bin/setup` finds the binary `VERSION` unchanged
   and no-ops.

**Binary change (new binary to ship):**
1. Bump `cmd/<tool>/VERSION` **and** `plugin.json.version` (the coupling rule), run
   `scripts/marketplace-generate`, commit + push.
2. `gh workflow run release.yml -f tool=<tool>` (or the Actions UI button).
3. The workflow: bootstrap-builds dendrik → `dendrik build cmd/<tool> --matrix` → guards
   immutability → `gh release create tool/vX.Y.Z dist/* --generate-notes`.
4. Consumers converge: the bumped plugin version triggers auto-update → `bin/setup` downloads the
   new binary for `VERSION`; `dot sync` likewise pulls the release asset for the host platform.

## Related contract checks

`dendrik lint` enforces the build-adjacent conventions: `makefile-targets`, `makefile-gofiles`,
`version-flag`, `symlink-entries`, `go-work-sync` (see `contract.go`). The release convention
above is the human source of truth those checks point at.
