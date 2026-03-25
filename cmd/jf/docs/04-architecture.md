# Architecture

## Overview

jf is a CLI that maps a local filesystem of markdown files to Jira ticket descriptions,
using a tree-shaped "forest" model to represent hierarchy. The data flows through a pipeline:

```
filesystem (.md files with YAML frontmatter)
     |
     v
forest model (discover -> tree of Nodes)
     |
     v
pipeline (compile markdown -> ADF JSON, or pull ADF -> markdown)
     |
     v
acli (Atlassian CLI -- transport layer to Jira REST API)
     |
     v
Jira (ticket descriptions, summaries, hierarchy)
```

## Module Structure

### `cmd/jf/` (root package)

| File | Responsibility |
|------|---------------|
| `main.go` | Command router with three-tier dispatch (no-prereq, local-only, Jira-touching). Defines `version` constant and `printUsage()`. |
| `helpers.go` | Shared utilities: `parseFlags()` for trailing-flag detection, `loadForest()` and `loadForestOrFail()` for forest discovery with error handling. |
| `helpers_test.go` | Tests for `parseFlags()` -- trailing flag detection, normal ordering, positional-only. |
| `cmd_push.go` | `runPush()` -- Level 0 single-file push (`pushSingle`) and forest-mode push (`pushForest`). Routes through engine Read-Plan-Execute pipeline. Supports `--dry-run`, `--yes`, `--json`, `--plain-text`. |
| `cmd_pull.go` | `runPull()` -- Level 0 single-file pull (`pullSingle`) and forest-mode pull (`pullForest`). Routes through engine Read-Plan-Execute pipeline. Supports `--dry-run`, `--yes`, `--json`. |
| `cmd_sync.go` | `runSync()` -- Loads forest, routes through engine Read-Plan-Execute pipeline. Supports `--dry-run`, `--yes`, `--json`, `--resolve local\|remote`. |
| `cmd_tree.go` | `runTree()` -- Forest hierarchy view with `--json` and `--verbose` flags. Includes `printTree()` for ASCII tree output with optional sync-direction icons and file paths. |
| `cmd_list.go` | `runList()` -- Flat list of all nodes with optional JSON. Includes `nodeToInfo()` to convert `Node` to `output.NodeInfo`. |
| `cmd_show.go` | `runShow()` -- Single-node detail view with state. Includes `nodeStatus()` for stale/clean/unknown and `syncDisplay()`. |
| `cmd_status.go` | `runStatus()` -- Forest-wide summary with effective derived direction (push+pull, pull-only, push-only, empty) and staleness counts. |
| `cmd_validate.go` | `runValidate()` -- Runs `forest.Validate()` and reports issues as text or JSON. |
| `cmd_clone.go` | `runClone()` -- Scaffolds a local forest from a Jira hierarchy into `.jf/` subdirectory of `--dir`. `--sync` omitted by default (derives from mutability); explicit `--sync push\|pull\|both` overrides. Records state baseline after pulling. Refuses to overwrite existing `.jf/forest.yml`. |
| `cmd_search.go` | `runSearch()` -- Thin JQL wrapper. Includes `buildSearchJQL()` to construct JQL from text/project/type filters. |
| `cmd_create.go` | `runCreateMissing()` -- Creates Jira tickets for TBD nodes. Includes `dryRunCreate()`, `executeCreate()`, `dedupCheck()`, `buildCreatePayload()`, `rewriteFrontmatterKey()`. |
| `cmd_init.go` | `runInit()` -- Creates `.jf/forest.yml` with defaults. Creates `.jf/` directory if needed. Uses `defaultForestYml` template. |
| `cmd_setup.go` | `runSetup()` -- Prerequisite checker. Includes `setupCheck()` (non-interactive) and `setupInteractive()`. |
| `cmd_schema.go` | `runSchema()` -- Emits JSON Schema for forest.yml and frontmatter (input) or command output types (output). |
| `cmd_rm.go` | `runRm()` -- Removes node files from the forest. Refuses if the node has children. |
| `cmd_view.go` | `runView()` -- Fetches remote issue details from Jira. Supports `--fields` and `--json`. |

### `internal/forest/`

| File | Responsibility |
|------|---------------|
| `schema.go` | Core types (`Forest`, `ForestDefaults`, `Node`, `Frontmatter`). Parsing: `ParseForestFile()`, `ParseFrontmatter()`, `ParseFrontmatterFile()`. Label derivation: `DeriveLabel()`, `firstHeading()`. |
| `discover.go` | Filesystem walk and tree building. `FindForest()` walks up directories for `.jf/forest.yml`. `Forest.Dir` is set to the `.jf/` directory. `Discover()` scans `.md` files within `.jf/`, builds parent-child hierarchy using README.md-as-parent convention. |
| `traverse.go` | DFS traversal utilities. `PostOrder()` (children before parents -- used for push). `PreOrder()` (parents before children -- used for create-missing). `Subtree()` and `Resolve()` for target lookup by key, filename stem, or file path. |
| `validate.go` | `Validate()` orchestrates four checks: `checkKeyUniqueness()`, `checkTBDNodes()`, `checkFieldValues()`, `checkStemUniqueness()`. Defines `ValidationIssue` struct. |
| `state.go` | State tracking in `state.json` (directly in the `.jf/` forest dir). Types: `State`, `NodeState`. Methods: `LoadState()`, `SaveState()` (atomic via tempfile+rename), `IsStale()`, `RecordPush()`, `RecordPull()`, `RecordSync()`, `ComputeHash()`, `IsPullStale()`, `MutabilityCache()`, `SetMutability()`. |

### `internal/pipeline/`

| File | Responsibility |
|------|---------------|
| `acli.go` | Transport layer. `Pipeline` struct with injectable `Runner`. Methods: `Compile()`, `Push()`, `View()`, `Search()`, `Create()`. Helper: `ExtractJiraKey()`. Defines `Runner` type and `DefaultRunner`. |
| `lint.go` | Validate restricted markdown subset. `Lint()` returns `[]LintIssue` for unsupported constructs. `FormatLintIssues()` for display. |
| `normalize.go` | Canonicalize markdown for comparison. `NormalizeMarkdown()` collapses blank lines, trims trailing whitespace. `ComputeLocalHash()` for content-addressed caching. |
| `roundtrip.go` | Verify md->ADF->md fidelity. `CheckRoundtrip()` compiles then converts back, comparing normalized output. |
| `adf.go` | Markdown->ADF conversion. `CompileMarkdown()` strips frontmatter and shells out to embedded `md2adf.bundle.mjs` via Node.js. |
| `adf2md.go` | ADF->markdown conversion. `ConvertADF()` shells out to embedded `adf2md.bundle.mjs` via Node.js. `ExtractDescriptionADF()` extracts `.fields.description` from acli JSON. |

### `internal/setup/`

| File | Responsibility |
|------|---------------|
| `check.go` | Prerequisite checking. `CheckAll()` runs `checkNode()`, `checkAcli()`, `checkJiraAuth()`. `QuickCheck()` returns one-line error for the first failure. Defines `CheckResult` struct and `Checker` type. |

### `internal/output/`

| File | Responsibility |
|------|---------------|
| `json.go` | JSON output types. Structs: `NodeResult`, `NodeInfo`, `StatusResult`, `ValidateResult`, `ValidateIssue`. Commands marshal these directly via `json.Marshal`. |

## Command Routing

Commands are dispatched in `main.go` through three switch blocks, forming prerequisite tiers:

### Tier 1: No Prerequisites

No external tools needed. Safe to run immediately.

```
setup    -> runSetup()
init     -> runInit()
schema   -> runSchema()
version  -> prints version constant
```

### Tier 2: Local-Only

Requires a `.jf/forest.yml` (found via `FindForest()`), but no Jira connectivity.

```
tree     -> runTree()
list     -> runList()
validate -> runValidate()
status   -> runStatus()
show     -> runShow()
rm       -> runRm()
```

### Tier 3: Jira-Touching

Guarded by `setup.QuickCheck(setup.DefaultChecker)` -- checks Node.js, acli, and Jira auth before dispatching.

```
push           -> runPush()
pull           -> runPull()
sync           -> runSync()
create-missing -> runCreateMissing()
search         -> runSearch()
clone          -> runClone()
view           -> runView()
```

The `QuickCheck` guard pattern: `QuickCheck()` calls `CheckAll()` which runs `checkNode()`, `checkAcli()`, and `checkJiraAuth()`. If any check fails, it returns a one-line error message. If all pass, it returns an empty string.

## Directory Model (Internals)

jf uses a `.jf/` dot-folder convention (analogous to `.git/`). All forest content lives inside `.jf/`:

```
working-directory/           <-- --dir target, working directory
  .jf/                       <-- forest root (Forest.Dir)
    forest.yml               <-- forest configuration
    state.json               <-- sync state (last push/pull timestamps, hashes)
    README.md                <-- root node
    planning/                <-- child directory
      README.md              <-- directory node
      PROJ-123.md            <-- leaf node
  reference/                 <-- non-forest files coexist cleanly
```

`FindForest()` walks up from the start directory looking for `.jf/forest.yml`.
`Forest.Dir` is set to the `.jf/` directory. `Discover()` walks within `.jf/`.
File paths in command output are prefixed with `.jf/` (working-dir-relative) for navigation.

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
    LastSync   time.Time `json:"last_sync,omitempty"`
    Direction  string    `json:"direction,omitempty"`
    LastPush   time.Time `json:"last_push,omitempty"`   // legacy, migrated to LastSync
    LastPull   time.Time `json:"last_pull,omitempty"`   // legacy, migrated to LastSync
    LocalHash  string    `json:"local_hash,omitempty"`  // sha256 of content below frontmatter
    RemoteHash string    `json:"remote_hash,omitempty"` // sha256 of ADF JSON from Jira
    MutableClean bool   `json:"mutable_clean,omitempty"`
    MutableHash  string `json:"mutable_hash,omitempty"`
}
```

Conflict detection compares content hashes. `IsStale()` checks `LastSync` first (engine pipeline), falling back to `LastPush` for pre-migration nodes. `MutabilityCache` caches lint+roundtrip results keyed by content hash.
```

### Output Types

From `internal/output/json.go`:

```go
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

### Validation and Setup Types

From `internal/forest/validate.go` and `internal/setup/check.go`:

```go
type ValidationIssue struct {
    Level   string // "error" | "warning"
    File    string
    Message string
}

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

1. **Read source** -- `os.ReadFile()` reads the `.md` file
2. **Lint + Roundtrip** -- `Lint()` validates the restricted subset, `CheckRoundtrip()` verifies md->ADF->md fidelity. Nodes that fail are read-only (skip push, demote to pull-only).
3. **Compile** -- `Pipeline.Compile(id, source, summary)` strips frontmatter, shells out to `node md2adf.bundle.mjs`, wraps result in an acli-edit JSON payload.
4. **Push** -- `Pipeline.Push(compiled)` writes compiled JSON to a temp file and runs `acli jira workitem edit --from-json`.

For `--plain-text` fallback, `engine.BuildPlainTextPayload()` wraps raw text in a minimal ADF paragraph node.

All commands route through the engine Read-Plan-Execute pipeline for state-tracked mutations.

### Pull Pipeline

The pull pipeline fetches a Jira description and converts to local markdown:

1. **View** -- `Pipeline.View(id, "description", true)` runs `acli jira workitem view` with `--json`
2. **Extract ADF** -- `ExtractDescriptionADF(viewJSON)` extracts `.fields.description` as JSON
3. **Convert** -- `ConvertADF(adfJSON)` shells out to `node adf2md.bundle.mjs`
4. **Merge frontmatter** -- `mergeWithFrontmatter(filePath, pulled)` preserves existing frontmatter

### Sync Pipeline

`runSync()` orchestrates bidirectional sync:

1. Loads forest and state
2. For each bidirectional node, compares local and remote content hashes to detect conflicts
3. If both sides changed and no `--resolve` flag, reports conflict and skips
4. Delegates to `pushForest()` then `pullForest()`

## Runner Interface

From `internal/pipeline/acli.go`:

```go
type Runner func(name string, args ...string) ([]byte, error)
```

`DefaultRunner` shells out via `os/exec`. The `Runner` type is injected into `Pipeline` for testability -- tests substitute a fake runner that returns canned responses.

Similarly, `internal/setup/check.go` defines `Checker func(name string, args ...string) (string, error)`. Both follow the same injectable-function pattern.

The `Runner` interface is the abstraction point for potential future transport changes (e.g., MCP instead of acli).
