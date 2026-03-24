# jf — Jira Forest CLI

Standalone CLI for managing Jira ticket hierarchies as local markdown forests.

## Build

```bash
cd cmd/jf
go build -o jf .
```

## Test

```bash
cd cmd/jf
go test ./...
```

## Binary Distribution

Built binary at `cmd/jf/jf`, symlinked to `~/.local/bin/jf` via `symlink_map.txt`.
After code changes: rebuild the binary and commit it.

## Code Conventions

### Adding a Command

1. Create `cmd_<name>.go` with `run<Name>(args []string) int`
2. Register in `main.go` switch — pick the prerequisite tier:
   - Tier 1 (first switch): no external deps (setup, init, schema, version)
   - Tier 2 (second switch): needs forest, no Jira (tree, list, etc.)
   - Tier 3 (third switch, after `QuickCheck`): touches Jira (push, pull, sync, etc.)
3. Use `parseFlags()` for all flag parsing — enforces flags-before-arguments ordering

### Flag Parsing

All commands must use `parseFlags()` instead of raw `fs.Parse()`. It detects trailing
flags after positional arguments and returns a descriptive error.

### Exit Codes

- `0` — success
- `1` — user/validation error (bad input, forest issues, missing args)
- `2` — external tool error (acli failure, Node.js failure)

### Embedded Bundles

`md2adf.bundle.mjs` and `adf2md.bundle.mjs` are checked-in esbuild artifacts:
- `md2adf.bundle.mjs` — marklassian (markdown → ADF)
- `adf2md.bundle.mjs` — extended-markdown-adf-parser (ADF → markdown)

Rebuild when upstream marklassian or extended-markdown-adf-parser change.
These are Go-embedded via `//go:embed` in `internal/pipeline/adf.go` and `adf2md.go`.

### JSON Output

Commands supporting `--json` use the `internal/output` package:
- `output.Result(data)` — marshals to stdout
- `output.Error(msg, detail)` — structured error to stderr

### Testing

`pipeline.Runner` and `setup.Checker` are injectable function types for testability.
Inject via `Pipeline{Run: yourFunc}` and `CheckAll(yourChecker)`.
Substitute with fake runners in tests to avoid shelling out to acli/node.

## Safe Sync Conventions

- Never try to bypass blocked operations — they are blocked for safety reasons
- When jf shows BLOCKED: report to user, don't retry with different flags
- Conflicts require `--resolve local|remote` — pick the direction explicitly
- Empty content never pushes — add substantive content first
- Use `--dry-run` to preview what sync/push/pull will do before executing
- Non-TTY batch operations are blocked — use `jf snapshot --json` + `jf push/pull --token`

## Snapshot-First Development

- `jf snapshot` writes plan-level JSON to `.jf/snapshots/latest.json`
- Auto-snapshot from `engine.Read()` writes raw readings format to the same path
- `--token` validation expects plan-level format — fails gracefully on raw format
- `ComputeToken` lives in `engine/` (engine concern); snapshot types live in `forest/` (avoids circular import)
- Token derivation: SHA256(snapshotID + key + localHash + remoteHash), truncated to 16 hex
- Two JSON fields (`token` for safe ops, `approve_token` for Tier 2), one CLI flag (`--token`)
- All `--token` responses go to stdout as structured JSON (via `dendrik.WriteResult`)
- Token path records state after success — next snapshot sees the node as synced
- Error codes: TOKEN_INVALID, SNAPSHOT_EXPIRED — both exit 1

## Deep Context

@docs/ARCHITECTURE.md
@docs/ROADMAP.md
