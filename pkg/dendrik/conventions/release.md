# Build & Release Convention

Source of truth for how dendrik tools are versioned, built, tagged, and released.
Owned by dendrik so every tool (and any tool later extracted to its own repo) releases the
same way. The build logic is provided by `dendrik build`; the GitHub orchestration is a thin
workflow shim.

## Version: single source of truth

Each tool's version is the trimmed contents of `cmd/<tool>/VERSION` (semver, e.g. `0.6.0`).
Nothing else stores the version: `main.go` declares `var version = "dev"`, overridden at build
time via `-ldflags -X main.version=<VERSION>`. To release a new version, **bump the VERSION
file first** (a normal commit). When a tool gains a `plugin.json` (marketplace distribution),
`plugin.json.version` becomes the source and `VERSION` mirrors it; `marketplace.json` mirrors
`plugin.json` (name/description/version/keywords synced from it) and must never carry an
*independent* version — a stale, unsynced value blocks plugin auto-updates. (Reference sync:
guideline-plugin-marketplace's `scripts/bump-changed-plugin-versions.sh`.)

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

1. Bump `cmd/<tool>/VERSION`; commit; push.
2. `gh workflow run release.yml -f tool=<tool>` (or the Actions UI button).
3. The workflow: bootstrap-builds dendrik → `dendrik build cmd/<tool> --matrix` → guards
   immutability → `gh release create tool/vX.Y.Z dist/* --generate-notes`.
4. Consumers read the same version: `dot sync` downloads the release asset for the host
   platform; the marketplace references the version via `plugin.json`.

## Related contract checks

`dendrik lint` enforces the build-adjacent conventions: `makefile-targets`, `makefile-gofiles`,
`version-flag`, `symlink-entries`, `go-work-sync` (see `contract.go`). The release convention
above is the human source of truth those checks point at.
