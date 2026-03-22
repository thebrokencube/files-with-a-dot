# Roadmap

Feature maturity classification for jf. Each feature is categorized by readiness:

- **stable** — production-ready, breaking changes follow versioning convention
- **experimental** — works but interface may change without major version bump
- **planned** — committed to building, no timeline
- **considering** — gathering feedback, may not happen

## Stable

### Core Forest Model

- `forest.yml` schema version 1
- YAML frontmatter discovery (`jira:`, `label:`, `type:`, `sync:`, `order:`)
- Filesystem-as-tree: directories become hierarchy, README.md becomes parent
- Label derivation chain: frontmatter → first `#` heading → filename stem (README.md → parent dirname)
- Default inheritance: `forest.yml defaults` → per-node frontmatter override

### Push Pipeline

- Markdown → ADF conversion via marklassian (`md2adf.bundle.mjs`)
- Frontmatter stripping before compilation (`StripFrontmatter`)
- `--force` fallback to plain text ADF when conversion fails
- Summary field update from node label during forest push
- acli transport: `acli jira workitem edit --from-json`

### Pull Pipeline

- ADF → markdown conversion via extended-markdown-adf-parser (`adf2md.bundle.mjs`)
- `ExtractDescriptionADF` for JSON field extraction
- Frontmatter preservation on pull (`mergeWithFrontmatter`)
- Plain text fallback when ADF conversion fails

### State Tracking

- `.jf/state.json` for per-node push/pull timestamps and content hashes
- Push staleness detection via file mtime comparison (`IsStale`)
- Pull staleness detection (`IsPullStale`)
- Atomic state writes via tempfile + rename (`SaveState`)

### Conflict Detection and Resolution

- Content hash comparison for local and remote changes (`DetectConflict`)
- `ConflictStatus` enum: `ConflictNone`, `ConflictLocalOnly`, `ConflictRemoteOnly`, `ConflictBoth`
- `--resolve local|remote` on sync to force winner

### Forest Inspection Commands

- `discover` — filesystem walk with JSON output
- `tree` — indented hierarchy view
- `list` — flat node list with JSON output
- `show` — single-node detail with state
- `status` — forest summary with staleness counts
- `validate` — integrity checks (key uniqueness, TBD completeness, field values, stem uniqueness)

### TBD Lifecycle

- `create-missing` with `--dry-run` preview
- Pre-order creation (parents before children)
- Dedup check before creation (JQL search for matching summary)
- Frontmatter rewrite: `jira: TBD` → `jira: PROJ-123`
- Description push after creation

### Forest Scaffolding

- `init` — create `forest.yml` with project defaults
- `clone` — scaffold from Jira hierarchy with recursive child fetching

### Prerequisites Checking

- `setup` — verify Node.js, acli, Jira auth
- `QuickCheck` guard on Jira-touching commands
- JSON output via `--check --json`

### Flag Ordering Enforcement

- `parseFlags` detects trailing flags after positional arguments
- Consistent error message across all commands

## Experimental

### Search Command

Thin JQL wrapper via `Pipeline.Search`. Supports `--project`, `--type`, `--limit` filters.
`--json` passes through to acli for structured output.

### Clone Hierarchy Depth Limiting

`--depth` flag on clone to limit recursive child fetching. Default 0 (unlimited).

### Park Workflow

Skill-level operation (not a CLI command). Combines `jf rm` with MCP-driven Jira
transitions, reparenting, and content clearing. Defined in `skill/references/lifecycle.md`.

### `rm` Command

Basic node file removal. Refuses if the node has children (no cascade). No confirmation
prompt, no `--force` override, no Jira-side cleanup.

## Planned

### `--json` for Tree

`tree` is the only inspection command without structured output.
It would emit `[]NodeInfo` (matching `discover --json`).

See: [ARCHITECTURE.md — Finding (a)](ARCHITECTURE.md#a-tree-and-search-lack---json)

### `--dry-run` for Push, Pull, and Sync

Preview what would be pushed/pulled without side effects. Only `create-missing`
currently has `--dry-run`.

See: [ARCHITECTURE.md — Finding (d)](ARCHITECTURE.md#d-no---dry-run-on-push-pull-or-sync)

### Tree-Drawing Helper Extraction

`printDiscoverTree` and `printTree` share the same `├─`/`└─`/`│ ` connector logic.
Extract a shared tree-walker that takes a per-node format function.

See: [ARCHITECTURE.md — Finding (c)](ARCHITECTURE.md#c-tree-drawing-connector-logic-duplicated)

### Sync Forest-Load Deduplication

`runSync` loads the forest once for conflict pre-scan, then `pushForest` and `pullForest`
each load it again. Refactor to load once and pass through.

See: [ARCHITECTURE.md — Finding (b)](ARCHITECTURE.md#b-sync-triple-loads-the-forest)

## Considering

### MCP Transport

Alternative to acli for Jira API calls. Would use MCP tools directly instead of
shelling out to acli. Removes the acli dependency but adds MCP protocol complexity.

### Multi-Cloud-ID Support

Multiple Atlassian sites in a single session. Currently `~/.jf.yml` holds one
`cloud_id`. Would need per-project or per-forest cloud ID configuration.

### Plugin Marketplace Packaging

Package the `/jf` skill as a standalone Claude Code plugin, installable without the
full dotfiles repo.

### Custom Field Registry

Shared mapping of custom field IDs to human-readable names (e.g.,
`customfield_12345` → `T-Shirt Size`). Currently each project in `~/.jf.yml` uses
raw field names.

### Subtree-Scoped Pull and Sync

`push` supports `--subtree` but `pull` and `sync` do not. Would allow pulling or
syncing only a branch of the forest.

### Forest-Level Hooks

Pre-push validation (e.g., lint markdown, check required sections) and post-pull
transforms (e.g., normalize formatting). Would use a `hooks:` section in `forest.yml`.

### Clone `--sync` Flag

`clone` hardcodes `sync: both` for all scaffolded nodes. A `--sync` flag would let
users choose the default sync direction at clone time.

See: [ARCHITECTURE.md — Finding (e)](ARCHITECTURE.md#e-clone-hardcodes-sync-both-for-all-scaffolded-nodes)

