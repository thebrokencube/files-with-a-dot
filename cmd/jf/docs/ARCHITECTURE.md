# Architecture

## Overview

jf is a CLI that maps a local filesystem of markdown files to Jira ticket descriptions,
using a tree-shaped "forest" model to represent hierarchy. The data flows through a pipeline:

```
filesystem (.md files with YAML frontmatter)
     │
     ▼
forest model (discover → tree of Nodes)
     │
     ▼
pipeline (compile markdown → ADF JSON, or pull ADF → markdown)
     │
     ▼
acli (Atlassian CLI — transport layer to Jira REST API)
     │
     ▼
Jira (ticket descriptions, summaries, hierarchy)
```

## Module Structure

### `cmd/jf/` (root package)

| File | Responsibility |
|------|---------------|
| `main.go` | Command router with three-tier dispatch (no-prereq, local-only, Jira-touching). Defines `version` constant and `printUsage()`. |
| `helpers.go` | Shared utilities: `parseFlags()` for trailing-flag detection, `loadForest()` and `loadForestOrFail()` for forest discovery with error handling. |
| `helpers_test.go` | Tests for `parseFlags()` — trailing flag detection, normal ordering, positional-only. |
| `cmd_push.go` | `runPush()` — Level 0 single-file push (`pushSingle`) and forest-mode push (`pushForest`). Routes through engine Read-Plan-Execute pipeline. Supports `--dry-run`, `--yes`, `--json`, `--plain-text`. |
| `cmd_pull.go` | `runPull()` — Level 0 single-file pull (`pullSingle`) and forest-mode pull (`pullForest`). Routes through engine Read-Plan-Execute pipeline. Supports `--dry-run`, `--yes`, `--json`. |
| `cmd_sync.go` | `runSync()` — Loads forest, routes through engine Read-Plan-Execute pipeline. Supports `--dry-run`, `--yes`, `--json`, `--resolve local\|remote`. |
| `cmd_tree.go` | `runTree()` — Forest hierarchy view with `--json` and `--verbose` flags. Includes `printTree()` for ASCII tree output with optional sync-direction icons and file paths. |
| `cmd_list.go` | `runList()` — Flat list of all nodes with optional JSON. Includes `nodeToInfo()` to convert `Node` to `output.NodeInfo`. |
| `cmd_show.go` | `runShow()` — Single-node detail view with state. Includes `nodeStatus()` for stale/clean/unknown and `syncDisplay()`. |
| `cmd_status.go` | `runStatus()` — Forest-wide summary with effective derived direction (push+pull, pull-only, push-only, empty) and staleness counts. |
| `cmd_validate.go` | `runValidate()` — Runs `forest.Validate()` and reports issues as text or JSON. |
| `cmd_clone.go` | `runClone()` — Scaffolds a local forest from a Jira hierarchy. `--sync` omitted by default (derives from mutability); explicit `--sync push\|pull\|both` overrides. Records state baseline after pulling. Includes `fetchIssue()`, `fetchTree()`, `parseSearchResults()`, `cloneFrontmatter()`, `scaffoldTree()`, `generateForestYAML()`, `slugify()`, `countNodes()`. |
| `cmd_snapshot.go` | `runSnapshot()` — Snapshot local+remote state with tokens. Routes through engine Read-Plan, writes plan-level JSON to `.jf/snapshots/latest.json`. Supports `--json`, `--dir`, optional KEY/KEY+FILE modes. Includes `enrichBlockedEntry()`, `truncateContent()`. |
| `cmd_search.go` | `runSearch()` — Thin JQL wrapper. Includes `buildSearchJQL()` to construct JQL from text/project/type filters. |
| `cmd_create.go` | `runCreateMissing()` — Creates Jira tickets for TBD nodes. Includes `dryRunCreate()`, `executeCreate()`, `dedupCheck()`, `buildCreatePayload()`, `rewriteFrontmatterKey()`, `rewriteTBDLine()`, `isTBDLine()`. |
| `cmd_init.go` | `runInit()` — Creates `forest.yml` with defaults. Uses `defaultForestYml` template. |
| `cmd_setup.go` | `runSetup()` — Prerequisite checker. Includes `setupCheck()` (non-interactive) and `setupInteractive()`. |
| `cmd_schema.go` | `runSchema()` — Emits JSON Schema for forest.yml and frontmatter (input) or command output types (output). |
| `cmd_rm.go` | `runRm()` — Removes node files from the forest. Refuses if the node has children. |

### `internal/forest/`

| File | Responsibility |
|------|---------------|
| `schema.go` | Core types (`Forest`, `ForestDefaults`, `Node`, `Frontmatter`). Parsing: `ParseForestFile()`, `ParseFrontmatter()`, `ParseFrontmatterFile()`. Label derivation: `DeriveLabel()`, `firstHeading()`. Frontmatter extraction: `extractFrontmatterBytes()`. Helper: `IsTBD()`. |
| `discover.go` | Filesystem walk and tree building. `FindForest()` walks up directories for `forest.yml`. `Discover()` scans `.md` files, builds parent-child hierarchy using README.md-as-parent convention. Defaults `node.Sync` to forest default, then to `"both"` if still empty (derived sync). Includes `findAncestorParent()`, `sortChildren()`. |
| `traverse.go` | DFS traversal utilities. `PostOrder()` (children before parents — used for push). `PreOrder()` (parents before children — used for create-missing). `Subtree()` and `Resolve()` for target lookup by key, filename stem, or file path. `Flatten()` for pre-order flat list. |
| `validate.go` | `Validate()` orchestrates four checks: `checkKeyUniqueness()`, `checkTBDNodes()`, `checkFieldValues()`, `checkStemUniqueness()`. Defines `ValidationIssue` struct. |
| `state.go` | State tracking in `.jf/state.json`. Types: `State`, `NodeState`. Methods: `LoadState()`, `SaveState()` (atomic via tempfile+rename), `IsStale()`, `RecordPush()`, `RecordPull()`, `RecordSync()`, `ComputeHash()`, `IsPullStale()`, `MutabilityCache()`, `SetMutability()`. |

### `internal/pipeline/`

| File | Responsibility |
|------|---------------|
| `acli.go` | Transport layer. `Pipeline` struct with injectable `Runner`. Methods: `Compile()`, `Push()`, `View()`, `Search()`, `Create()`. Helper: `ExtractJiraKey()`. Defines `Runner` type and `DefaultRunner`. |
| `lint.go` | Validate restricted markdown subset. `Lint()` returns `[]LintIssue` for unsupported constructs (tables, code blocks, h1/h3+, etc.). `FormatLintIssues()` for display. |
| `normalize.go` | Canonicalize markdown for comparison. `NormalizeMarkdown()` collapses blank lines, trims trailing whitespace. `ComputeLocalHash()` for content-addressed caching. |
| `roundtrip.go` | Verify md→ADF→md fidelity. `CheckRoundtrip()` compiles then converts back, comparing normalized output. |
| `adf.go` | Markdown→ADF conversion. `CompileMarkdown()` strips frontmatter and shells out to embedded `md2adf.bundle.mjs` via Node.js. `StripFrontmatter()` removes YAML fences. Script management: `ensureScript()` (write-once temp file). |
| `adf2md.go` | ADF→markdown conversion. `ConvertADF()` shells out to embedded `adf2md.bundle.mjs` via Node.js. `ExtractDescriptionADF()` extracts `.fields.description` from acli JSON. Script management: `ensureADF2MDScript()`. |

### `internal/setup/`

| File | Responsibility |
|------|---------------|
| `check.go` | Prerequisite checking. `CheckAll()` runs `checkNode()`, `checkAcli()`, `checkJiraAuth()`. `QuickCheck()` returns one-line error for the first failure. Defines `CheckResult` struct and `Checker` type (injectable for testing). |

### `internal/output/`

| File | Responsibility |
|------|---------------|
| `json.go` | JSON output envelope. `Result()` marshals to stdout. `Error()` writes structured error to stderr. Types: `ErrorResult`, `NodeResult`, `NodeInfo`, `StatusResult`, `ValidateResult`, `ValidateIssue`. |

## Command Routing

Commands are dispatched in `main.go` through three switch blocks, forming prerequisite tiers:

### Tier 1: No Prerequisites

No external tools needed. Safe to run immediately.

```
setup    → runSetup()
init     → runInit()
schema   → runSchema()
version  → prints version constant
```

### Tier 2: Local-Only

Requires a `forest.yml` (found via `FindForest()`), but no Jira connectivity.

```
tree     → runTree()
list     → runList()
validate → runValidate()
status   → runStatus()
show     → runShow()
rm       → runRm()
```

### Tier 3: Jira-Touching

Guarded by `setup.QuickCheck(setup.DefaultChecker)` — checks Node.js, acli, and Jira auth
before dispatching.

```
snapshot       → runSnapshot()
push           → runPush()
pull           → runPull()
sync           → runSync()
create-missing → runCreateMissing()
search         → runSearch()
clone          → runClone()
```

The `QuickCheck` guard pattern: `QuickCheck()` calls `CheckAll()` which runs `checkNode()`,
`checkAcli()`, and `checkJiraAuth()`. If any check fails, it returns a one-line error message
(e.g., `"✗ acli not found. Run: jf setup"`). If all pass, it returns an empty string.

## Data Models

### Forest Configuration

From `internal/forest/schema.go`:

```go
type Forest struct {
    Schema   int            `yaml:"schema"`
    Defaults ForestDefaults `yaml:"defaults"`
    Acli     string         `yaml:"acli"` // version constraint, optional
    Dir      string         `yaml:"-"`    // runtime: directory containing forest.yml
}

type ForestDefaults struct {
    Sync    string `yaml:"sync"`    // push | pull | both (optional, override-only; defaults to "both")
    Type    string `yaml:"type"`    // Jira issue type (for creation)
    Field   string `yaml:"field"`   // description | comment
    Project string `yaml:"project"` // Jira project key
}
```

### Node and Frontmatter

From `internal/forest/schema.go`:

```go
type Node struct {
    Key      string  // from jira: field ("ACME-123" or "TBD")
    Label    string  // from label: field, first # heading, or filename stem
    Type     string  // from type: field or defaults
    Sync     string  // from sync: field or defaults (push | pull | both; defaults to "both")
    Order    int     // from order: field (0 = unset, alphabetical)
    File     string  // relative path from forest root
    Parent   *Node   `yaml:"-" json:"-"` // nil for root nodes
    Children []*Node // child nodes
}

type Frontmatter struct {
    Jira  string `yaml:"jira"`
    Label string `yaml:"label"`
    Type  string `yaml:"type"`
    Sync  string `yaml:"sync"`
    Order int    `yaml:"order"`
}
```

### State Tracking

From `internal/forest/state.go`:

```go
type State struct {
    Nodes map[string]NodeState `json:"nodes"`
}

type NodeState struct {
    LastPush   time.Time `json:"last_push,omitempty"`
    LastPull   time.Time `json:"last_pull,omitempty"`
    LocalHash  string    `json:"local_hash,omitempty"`  // sha256 of content below frontmatter
    RemoteHash string    `json:"remote_hash,omitempty"` // sha256 of ADF JSON from Jira
}
```

Conflict detection uses a `ConflictStatus` enum:

```go
type ConflictStatus int

const (
    ConflictNone       ConflictStatus = iota
    ConflictLocalOnly                 // local changed, remote unchanged
    ConflictRemoteOnly                // remote changed, local unchanged
    ConflictBoth                      // both sides changed
)
```

### Output Types

From `internal/output/json.go`:

```go
type ErrorResult struct {
    Error  string `json:"error"`
    Detail string `json:"detail,omitempty"`
}

type NodeResult struct {
    Node   string `json:"node"`
    File   string `json:"file,omitempty"`
    Status string `json:"status"` // "ok" | "error" | "skipped"
    Error  string `json:"error,omitempty"`
    Detail string `json:"detail,omitempty"`
    Size   int    `json:"size,omitempty"`
}

type NodeInfo struct {
    Key      string `json:"key"`
    Label    string `json:"label"`
    Type     string `json:"type"`
    Sync     string `json:"sync"`
    File     string `json:"file"`
    Parent   string `json:"parent,omitempty"`
    Children int    `json:"children"`
    Status   string `json:"status,omitempty"` // "stale" | "clean" | "unknown"
}

type StatusResult struct {
    Forest    string `json:"forest"`
    Total     int    `json:"total"`
    TBD       int    `json:"tbd"`
    PushTotal int    `json:"push_total"`  // push-eligible (mutable + push-only + empty)
    PushStale int    `json:"push_stale"`
    PullTotal int    `json:"pull_total"`  // pull-only (explicit + read-only demoted)
    PullStale int    `json:"pull_stale"`
    Mutable   int    `json:"mutable"`
    ReadOnly  int    `json:"read_only"`
    Empty     int    `json:"empty"`
}

type ValidateResult struct {
    Valid  bool            `json:"valid"`
    Nodes  int             `json:"nodes"`
    Issues []ValidateIssue `json:"issues,omitempty"`
}

type ValidateIssue struct {
    Level   string `json:"level"` // "error" | "warning"
    Message string `json:"message"`
}
```

### Validation Types

From `internal/forest/validate.go`:

```go
type ValidationIssue struct {
    Level   string // "error" | "warning"
    File    string
    Message string
}
```

### Setup Types

From `internal/setup/check.go`:

```go
type CheckResult struct {
    Name   string `json:"name"`
    Status string `json:"status"` // "ok" | "missing" | "outdated"
    Detail string `json:"detail"`
    Fix    string `json:"fix"`
}

type Checker func(name string, args ...string) (string, error)
```

## Pipeline Detail

### Push Pipeline

The push pipeline converts local markdown to a Jira description. Content must pass
lint and roundtrip checks before pushing (see Lint and Mutability in USAGE.md):

1. **Read source** — `os.ReadFile()` reads the `.md` file
2. **Lint + Roundtrip** — `Lint()` validates the restricted subset, `CheckRoundtrip()` verifies md→ADF→md fidelity. Nodes that fail are read-only (skip push, demote to pull-only in both mode).
3. **`Pipeline.Compile(id, source, summary)`** (`acli.go:30`)
   - Calls `CompileMarkdown(source)` which:
     - Strips frontmatter via `StripFrontmatter()` (`adf.go:29`)
     - Writes stripped markdown to a temp file
     - Shells out to `node md2adf.bundle.mjs <tmpfile>` (embedded marklassian bundle)
     - Parses stdout as `json.RawMessage` (ADF document)
   - Wraps result in an acli-edit JSON payload: `{"issues": [id], "description": <ADF>}`
   - If `summary` is non-empty, adds `"summary"` field to payload
4. **`Pipeline.Push(compiled)`** (`acli.go:48`)
   - Writes compiled JSON to a temp file
   - Runs `acli jira workitem edit --from-json <tmpfile> --yes`

For `--plain-text` fallback, `engine.BuildPlainTextPayload()` wraps raw text in a minimal ADF paragraph node.

All commands route through the engine Read-Plan-Execute pipeline. Push, pull, and sync
use `engine.Read()` for parallel state fetching (with mutability computation),
`engine.Plan()` for safety rule evaluation, and `engine.Execute()` for state-tracked mutations.

### Pull Pipeline

The pull pipeline fetches a Jira description and converts to local markdown:

1. **`Pipeline.View(id, "description", true)`** (`acli.go:103`)
   - Runs `acli jira workitem view <id> --fields description --json`
   - Returns raw JSON output
2. **`ExtractDescriptionADF(viewJSON)`** (`adf2md.go:77`)
   - Unmarshals the view JSON and extracts `.fields.description` as `json.RawMessage`
   - Returns `nil` if description is absent or `null`
3. **`ConvertADF(adfJSON)`** (`adf2md.go:46`)
   - Writes ADF JSON to a temp file
   - Shells out to `node adf2md.bundle.mjs <tmpfile>` (embedded extended-markdown-adf-parser bundle)
   - Returns markdown bytes from stdout
4. **`mergeWithFrontmatter(filePath, pulled)`** (`cmd_pull.go:205`)
   - Reads existing file, extracts frontmatter via `extractExistingFrontmatter()`
   - Combines: existing frontmatter block + `---\n` + pulled markdown content

Forest-mode pull (`pullForest`) collects `sync:pull` nodes, records state via
`State.RecordPull(key, localHash, remoteHash)`. The `skipState` parameter (used by clone)
prevents recording state so the first sync can detect conflicts cleanly.

### Sync Pipeline

`runSync()` (`cmd_sync.go:12`) orchestrates bidirectional sync:

1. Loads forest and state
2. For each `sync:both` node, fetches remote ADF and calls `State.DetectConflict()`
3. If `ConflictBoth` is detected and no `--resolve` flag, reports conflict and skips
4. Delegates to `pushForest()` then `pullForest()`

`runSync()` loads the forest once and passes it to `pushForest()` and `pullForest()`.
Both functions accept optional pre-loaded `*Forest`/`[]*Node` to avoid redundant loads.

### Runner Interface

From `internal/pipeline/acli.go`:

```go
type Runner func(name string, args ...string) ([]byte, error)
```

`DefaultRunner` shells out via `os/exec`. The `Runner` type is injected into `Pipeline`
for testability — tests can substitute a fake runner that returns canned responses.

Similarly, `internal/setup/check.go` defines:

```go
type Checker func(name string, args ...string) (string, error)
```

`DefaultChecker` shells out via `os/exec`. Both follow the same injectable-function pattern.

## Command Overlap & Consolidation

### (a) ~~`tree` and `search` lack `--json`~~ — resolved

Both `tree` and `search` now support `--json`. `tree --json` emits `[]NodeInfo` (absorbing
the former `discover --json` output). `search --json` passes through to `Pipeline.Search()`.

### (b) ~~`sync` triple-loads the forest~~ — resolved

`pushForest` and `pullForest` now accept optional `*Forest` and `[]*Node` parameters.
`runSync` loads the forest once and passes it through. Standalone `runPush`/`runPull`
pass `nil, nil` and load internally.

### (c) ~~Tree-drawing connector logic duplicated~~ — resolved

`discover` has been consolidated into `tree`. A single `printTree()` with a `verbose bool`
parameter handles both the clean view and the detailed view (sync icons + file paths).

### (d) ~~No `--dry-run` on `push`, `pull`, or `sync`~~ — resolved

All three commands now accept `--dry-run`. Push shows `[dry-run] would push KEY (FILE, N bytes)`.
Pull shows `[dry-run] would pull KEY -> FILE`. Sync runs conflict pre-scan (read-only)
then emits dry-run previews for both phases.

### (e) ~~`clone` hardcodes `sync: both` for all scaffolded nodes~~ — resolved

`clone` now omits `sync:` from scaffolded frontmatter and forest.yml by default (derives
from mutability at runtime). Explicit `--sync push|pull|both` overrides emit the field.
Clone also records a state baseline (skipState=false) so the first sync has real content
hashes.

### (f) ~~`search` outputs raw acli text with no structured output~~ — resolved

`runSearch()` now accepts `--json` and passes it through to `Pipeline.Search()`.

### (g) ~~50-line frontmatter parsing limit is a coupling point~~ — resolved

All four sites (`extractFrontmatterBytes`, `StripFrontmatter`, `extractExistingFrontmatter`,
`rewriteTBDLine`) now use `forest.MaxFrontmatterLines` (defined in `internal/forest/schema.go`).

## Extension Points

### Adding a Command

1. Create `cmd_<name>.go` with a `run<Name>(args []string) int` function
2. Register in `main.go` in the appropriate switch block:
   - Tier 1 (first switch) — no external dependencies
   - Tier 2 (second switch) — needs forest, no Jira
   - Tier 3 (third switch, after `QuickCheck`) — touches Jira
3. Use `parseFlags()` for flag parsing (supports flags before or after positional arguments)
4. Use `loadForestOrFail()` if the command needs a forest
5. Use `output.Result()` for `--json` output, `output.Error()` for JSON errors

### Runner Interface for Testing

Inject a custom `Runner` into `Pipeline` to test without acli:

```go
p := &pipeline.Pipeline{
    Run: func(name string, args ...string) ([]byte, error) {
        return []byte(`{"key":"TEST-1"}`), nil
    },
}
```

Similarly, inject a custom `Checker` into `setup.CheckAll()` for setup tests.

### Output Package

All JSON output goes through `output.Result()` (stdout) and `output.Error()` (stderr).
Commands that support `--json` should use the types from `internal/output/json.go`:
`NodeInfo`, `StatusResult`, `ValidateResult`, `ErrorResult`, `NodeResult`.

