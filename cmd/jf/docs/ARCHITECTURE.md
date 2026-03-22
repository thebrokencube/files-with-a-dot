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
| `cmd_push.go` | `runPush()` — Level 0 single-file push (`pushSingle`) and forest-mode push (`pushForest`). Includes `buildPlainTextPayload()` for `--force` fallback. |
| `cmd_pull.go` | `runPull()` — Level 0 single-file pull (`pullSingle`) and forest-mode pull (`pullForest`). Includes `richPull()` for ADF→markdown conversion, `mergeWithFrontmatter()` to preserve YAML frontmatter on pull, `extractExistingFrontmatter()`. |
| `cmd_sync.go` | `runSync()` — Pre-scans `sync:both` nodes for conflicts via `DetectConflict()`, then delegates to `pushForest()` and `pullForest()`. |
| `cmd_discover.go` | `runDiscover()` — Discovers forest nodes and prints tree or JSON. Includes `printDiscoverTree()` for ASCII tree output with sync-direction icons. |
| `cmd_tree.go` | `runTree()` — Minimal hierarchy view. Includes `printTree()` for ASCII tree output (key + label only). |
| `cmd_list.go` | `runList()` — Flat list of all nodes with optional JSON. Includes `nodeToInfo()` to convert `Node` to `output.NodeInfo`. |
| `cmd_show.go` | `runShow()` — Single-node detail view with state. Includes `nodeStatus()` for stale/clean/unknown and `syncDisplay()`. |
| `cmd_status.go` | `runStatus()` — Forest-wide summary with push/pull staleness counts. |
| `cmd_validate.go` | `runValidate()` — Runs `forest.Validate()` and reports issues as text or JSON. |
| `cmd_clone.go` | `runClone()` — Scaffolds a local forest from a Jira hierarchy. Includes `fetchIssue()`, `fetchTree()`, `parseSearchResults()`, `scaffoldTree()`, `generateForestYAML()`, `slugify()`, `countNodes()`. |
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
| `discover.go` | Filesystem walk and tree building. `FindForest()` walks up directories for `forest.yml`. `Discover()` scans `.md` files, builds parent-child hierarchy using README.md-as-parent convention. Includes `findAncestorParent()`, `sortChildren()`. |
| `traverse.go` | DFS traversal utilities. `PostOrder()` (children before parents — used for push). `PreOrder()` (parents before children — used for create-missing). `Subtree()` and `Resolve()` for target lookup by key, filename stem, or file path. `Flatten()` for pre-order flat list. |
| `validate.go` | `Validate()` orchestrates four checks: `checkKeyUniqueness()`, `checkTBDNodes()`, `checkFieldValues()`, `checkStemUniqueness()`. Defines `ValidationIssue` struct. |
| `state.go` | State tracking in `.jf/state.json`. Types: `State`, `NodeState`, `ConflictStatus`. Methods: `LoadState()`, `SaveState()` (atomic via tempfile+rename), `IsStale()`, `RecordPush()`, `RecordPull()`, `ComputeHash()`, `DetectConflict()`, `IsPullStale()`. |

### `internal/pipeline/`

| File | Responsibility |
|------|---------------|
| `acli.go` | Transport layer. `Pipeline` struct with injectable `Runner`. Methods: `Compile()`, `Push()`, `View()`, `Search()`, `Create()`. Helper: `ExtractJiraKey()`. Defines `Runner` type and `DefaultRunner`. |
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
discover → runDiscover()
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
    Sync    string `yaml:"sync"`    // push | pull
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
    Sync     string  // from sync: field or defaults (push | pull)
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
    PushTotal int    `json:"push_total"`
    PushStale int    `json:"push_stale"`
    PullTotal int    `json:"pull_total"`
    PullStale int    `json:"pull_stale"`
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

The push pipeline converts local markdown to a Jira description:

1. **Read source** — `os.ReadFile()` reads the `.md` file
2. **`Pipeline.Compile(id, source, summary)`** (`acli.go:30`)
   - Calls `CompileMarkdown(source)` which:
     - Strips frontmatter via `StripFrontmatter()` (`adf.go:29`)
     - Writes stripped markdown to a temp file
     - Shells out to `node md2adf.bundle.mjs <tmpfile>` (embedded marklassian bundle)
     - Parses stdout as `json.RawMessage` (ADF document)
   - Wraps result in an acli-edit JSON payload: `{"issues": [id], "description": <ADF>}`
   - If `summary` is non-empty, adds `"summary"` field to payload
3. **`Pipeline.Push(compiled)`** (`acli.go:48`)
   - Writes compiled JSON to a temp file
   - Runs `acli jira workitem edit --from-json <tmpfile> --yes`

For `--force` fallback, `buildPlainTextPayload()` wraps raw text in a minimal ADF paragraph node.

Forest-mode push (`pushForest`) uses `forest.PostOrder()` to push children before parents,
filters to `sync:push` or `sync:both` nodes (skipping TBD and `sync:pull`), tracks state
via `State.RecordPush()`, and supports `--subtree` to scope the push.

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

Note: `runSync()` calls `loadForest()` for conflict pre-scan, then `pushForest()` and
`pullForest()` each call `loadForest()` again internally — three total forest loads per sync.

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

### (a) `tree` and `search` lack `--json`

**Where:** `cmd_tree.go` — `runTree()` only has `--dir` flag. `cmd_search.go` — `runSearch()`
has `--project`, `--type`, `--limit` but no `--json`.

**Context:** `discover`, `list`, `show`, `status`, `validate`, and `setup --check` all
support `--json` output using the `output` package. `tree` and `search` are the only
inspection/query commands without structured output.

**Fix:** Add `--json` flag to both. For `tree`, emit `[]NodeInfo` (same as `discover --json`).
For `search`, pass `jsonOut: true` to `Pipeline.Search()` (it already accepts a `jsonOut`
parameter — `cmd_search.go:31` hardcodes `false`).

### (b) `sync` triple-loads the forest

**Where:** `cmd_sync.go:29` — `loadForest(*dir)` for conflict pre-scan. Then
`cmd_sync.go:91` calls `pushForest()` which calls `loadForest()` at `cmd_push.go:64`.
Then `cmd_sync.go:95` calls `pullForest()` which calls `loadForest()` at `cmd_pull.go:84`.

**Context:** Each `loadForest()` call does `FindForest()` (directory walk) + `Discover()`
(filesystem scan + tree build). Three full scans per sync.

**Fix:** Load forest once in `runSync()`, pass the `*Forest` and `[]*Node` to push/pull
functions instead of having them re-discover independently.

### (c) Tree-drawing connector logic duplicated

**Where:** `cmd_discover.go:44-64` (`printDiscoverTree`) and `cmd_tree.go:35-51`
(`printTree`) both implement the same `├─`/`└─`/`│ ` connector pattern independently.

**Context:** Both functions walk `[]*Node` children, compute last-child branching,
and print indented trees. The only difference is what columns they print per node.

**Fix:** Extract a shared tree-walker that takes a per-node format function, used by both
`discover` and `tree`.

### (d) No `--dry-run` on `push`, `pull`, or `sync`

**Where:** `cmd_push.go` — no dry-run flag. `cmd_pull.go` — no dry-run flag.
`cmd_sync.go` — no dry-run flag. Only `cmd_create.go` has `--dry-run` (`dryRunCreate()`).

**Context:** Push and pull directly modify Jira descriptions or local files with no
preview option. A `--dry-run` would show what would be pushed/pulled without side effects.

**Fix:** Add `--dry-run` to push (list nodes that would be pushed with byte counts),
pull (list nodes that would be pulled), and sync (combine both).

### (e) `clone` hardcodes `sync: both` for all scaffolded nodes

**Where:** `cmd_clone.go:172` — `scaffoldTree()` writes `sync: both` in every scaffolded
file's frontmatter. `cmd_clone.go:199` — `generateForestYAML()` writes `sync: both` as
the forest default.

**Context:** All cloned nodes get bidirectional sync regardless of intent. Users who only
want to push descriptions must manually change every file's frontmatter or the forest default.

**Fix:** Add a `--sync` flag to `clone` (default: `both`) that sets the forest default and
scaffolded node frontmatter. Could also omit per-node `sync:` and let the forest default
apply (less repetitive frontmatter).

### (f) `search` outputs raw acli text with no structured output

**Where:** `cmd_search.go:30` — `p.Search(jql, "summary,issuetype,status", *limit, false)`.
The `false` parameter means acli returns human-readable text. `cmd_search.go:36` —
`fmt.Print(string(out))` passes raw output through.

**Context:** The `Pipeline.Search()` method already accepts a `jsonOut bool` parameter and
passes `--json` to acli when true. The `search` command simply doesn't expose this.

**Fix:** Add `--json` flag to `runSearch()` and pass it through to `Pipeline.Search()`.

### (g) 50-line frontmatter parsing limit is a coupling point

**Where:** `internal/forest/schema.go:131` — `extractFrontmatterBytes()` limits scan to
50 lines. `internal/pipeline/adf.go:36` — `StripFrontmatter()` independently limits to
50 lines. Both use the same hardcoded `50` constant.

**Context:** If either limit changes without the other, frontmatter that parses during
discovery might not be stripped before compilation (or vice versa). The two limits must stay
in sync.

**Fix:** Extract a shared constant (e.g., `MaxFrontmatterLines = 50`) or a shared
frontmatter-extraction function used by both packages.

## Extension Points

### Adding a Command

1. Create `cmd_<name>.go` with a `run<Name>(args []string) int` function
2. Register in `main.go` in the appropriate switch block:
   - Tier 1 (first switch) — no external dependencies
   - Tier 2 (second switch) — needs forest, no Jira
   - Tier 3 (third switch, after `QuickCheck`) — touches Jira
3. Use `parseFlags()` for flag parsing (enforces flags-before-arguments)
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

## Known Gaps

Formatted as future observations:

- `gap(jf): tree command lacks --json output unlike other inspection commands`
- `gap(jf): search command lacks --json output — returns raw acli text`
- `debt(jf): sync re-loads forest then delegates to pushForest/pullForest which each load again`
- `idea(jf): tree-drawing connector logic shared between discover and tree — extract to helper`
- `idea(jf): add --dry-run to push/pull/sync for safe previewing`
- `debt(jf): clone hardcodes sync:both for all scaffolded nodes — no --sync flag`
- `debt(jf): 50-line frontmatter parsing limit is a coupling point across forest and pipeline`
- `gap(jf): search passes jsonOut=false to Pipeline.Search which already supports --json passthrough`
- `debt(jf): clone skips state recording (skipState=true) as a workaround for conflict detection on first sync`
